// Package registry 实现插件注册中心
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"../core"
)

// DefaultPluginRegistry 默认插件注册表实现
type DefaultPluginRegistry struct {
	plugins     map[string]*RegistryEntry
	pluginPaths []string
	mutex       sync.RWMutex
	watchers    []RegistryWatcher
	config      *RegistryConfig
}

// RegistryEntry 注册表条目
type RegistryEntry struct {
	Metadata     *core.PluginMetadata
	Path         string
	RegisterTime time.Time
	LastUpdate   time.Time
	Status       RegistryStatus
	Dependencies []string
	Dependents   []string
}

// RegistryStatus 注册状态
type RegistryStatus string

const (
	StatusRegistered   RegistryStatus = "registered"
	StatusUnregistered RegistryStatus = "unregistered"
	StatusError        RegistryStatus = "error"
	StatusDeprecated   RegistryStatus = "deprecated"
)

// RegistryConfig 注册表配置
type RegistryConfig struct {
	AutoScan      bool
	ScanInterval  time.Duration
	PluginPaths   []string
	EnableWatcher bool
	CacheEnabled  bool
	CacheFile     string
}

// RegistryWatcher 注册表观察者
type RegistryWatcher interface {
	OnPluginRegistered(metadata *core.PluginMetadata)
	OnPluginUnregistered(pluginID string)
	OnPluginUpdated(metadata *core.PluginMetadata)
}

// NewDefaultPluginRegistry 创建新的插件注册中心
func NewDefaultPluginRegistry(config *RegistryConfig) *DefaultPluginRegistry {
	if config == nil {
		config = &RegistryConfig{
			AutoScan:      true,
			ScanInterval:  30 * time.Second,
			PluginPaths:   []string{"./plugins"},
			EnableWatcher: true,
			CacheEnabled:  true,
			CacheFile:     "plugin_registry.json",
		}
	}

	registry := &DefaultPluginRegistry{
		plugins:     make(map[string]*RegistryEntry),
		pluginPaths: config.PluginPaths,
		config:      config,
	}

	// 加载缓存
	if config.CacheEnabled {
		registry.loadCache()
	}

	// 启动自动扫描
	if config.AutoScan {
		go registry.autoScan()
	}

	return registry
}

// Register 注册插件
func (r *DefaultPluginRegistry) Register(metadata *core.PluginMetadata, path string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查插件是否已注册
	if existing, exists := r.plugins[metadata.ID]; exists {
		// 检查版本是否相同
		if existing.Metadata.Version == metadata.Version {
			return fmt.Errorf("plugin %s version %s already registered", metadata.ID, metadata.Version)
		}

		// 更新现有插件
		existing.Metadata = metadata
		existing.Path = path
		existing.LastUpdate = time.Now()
		existing.Status = StatusRegistered

		// 通知观察者插件更新
		r.notifyWatchers(func(w RegistryWatcher) {
			w.OnPluginUpdated(metadata)
		})

		return nil
	}

	// 创建新的注册条目
	entry := &RegistryEntry{
		Metadata:     metadata,
		Path:         path,
		RegisterTime: time.Now(),
		LastUpdate:   time.Now(),
		Status:       StatusRegistered,
		Dependencies: metadata.Dependencies,
	}

	// 更新依赖关系
	r.updateDependencies(metadata.ID, metadata.Dependencies)

	r.plugins[metadata.ID] = entry

	// 通知观察者插件注册
	r.notifyWatchers(func(w RegistryWatcher) {
		w.OnPluginRegistered(metadata)
	})

	// 保存缓存
	if r.config.CacheEnabled {
		go r.saveCache()
	}

	return nil
}

// Unregister 注销插件
func (r *DefaultPluginRegistry) Unregister(pluginID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 检查是否有其他插件依赖此插件
	if len(entry.Dependents) > 0 {
		return fmt.Errorf("plugin %s has dependents: %v", pluginID, entry.Dependents)
	}

	// 清理依赖关系
	r.cleanupDependencies(pluginID)

	delete(r.plugins, pluginID)

	// 通知观察者插件注销
	r.notifyWatchers(func(w RegistryWatcher) {
		w.OnPluginUnregistered(pluginID)
	})

	// 保存缓存
	if r.config.CacheEnabled {
		go r.saveCache()
	}

	return nil
}

// Get 获取插件信息
func (r *DefaultPluginRegistry) Get(pluginID string) (*core.PluginMetadata, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return entry.Metadata, nil
}

// List 列出所有已注册插件
func (r *DefaultPluginRegistry) List() []*core.PluginMetadata {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var plugins []*core.PluginMetadata
	for _, entry := range r.plugins {
		if entry.Status == StatusRegistered {
			plugins = append(plugins, entry.Metadata)
		}
	}

	// 按名称排序
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	return plugins
}

// Find 查找插件
func (r *DefaultPluginRegistry) Find(filter *PluginFilter) []*core.PluginMetadata {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var results []*core.PluginMetadata

	for _, entry := range r.plugins {
		if entry.Status != StatusRegistered {
			continue
		}

		if filter.Match(entry.Metadata) {
			results = append(results, entry.Metadata)
		}
	}

	return results
}

// GetDependencies 获取插件依赖
func (r *DefaultPluginRegistry) GetDependencies(pluginID string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return entry.Dependencies, nil
}

// GetDependents 获取依赖此插件的其他插件
func (r *DefaultPluginRegistry) GetDependents(pluginID string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return entry.Dependents, nil
}

// Scan 扫描插件目录
func (r *DefaultPluginRegistry) Scan(ctx context.Context) error {
	for _, pluginPath := range r.pluginPaths {
		if err := r.scanDirectory(ctx, pluginPath); err != nil {
			return fmt.Errorf("failed to scan directory %s: %w", pluginPath, err)
		}
	}
	return nil
}

// scanDirectory 扫描指定目录
func (r *DefaultPluginRegistry) scanDirectory(ctx context.Context, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 查找插件清单文件
		if info.Name() == "plugin.json" || info.Name() == "manifest.json" {
			if err := r.loadPluginFromManifest(path); err != nil {
				// 记录错误但继续扫描其他文件				fmt.Printf("Failed to load plugin from %s: %v\n", path, err)
			}
		}

		return nil
	})
}

// loadPluginFromManifest 从清单文件加载插件
func (r *DefaultPluginRegistry) loadPluginFromManifest(manifestPath string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	var metadata core.PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// 验证必要字段
	if metadata.ID == "" || metadata.Name == "" || metadata.Version == "" {
		return fmt.Errorf("invalid plugin metadata: missing required fields")
	}

	pluginDir := filepath.Dir(manifestPath)
	return r.Register(&metadata, pluginDir)
}

// updateDependencies 更新依赖关系
func (r *DefaultPluginRegistry) updateDependencies(pluginID string, dependencies []string) {
	// 为每个依赖项添加此插件为依赖者
	for _, depID := range dependencies {
		if depEntry, exists := r.plugins[depID]; exists {
			// 检查是否已存在
			found := false
			for _, dependent := range depEntry.Dependents {
				if dependent == pluginID {
					found = true
					break
				}
			}
			if !found {
				depEntry.Dependents = append(depEntry.Dependents, pluginID)
			}
		}
	}
}

// cleanupDependencies 清理依赖关系
func (r *DefaultPluginRegistry) cleanupDependencies(pluginID string) {
	entry := r.plugins[pluginID]

	// 从依赖项的依赖者列表中移除此插件
	for _, depID := range entry.Dependencies {
		if depEntry, exists := r.plugins[depID]; exists {
			for i, dependent := range depEntry.Dependents {
				if dependent == pluginID {
					depEntry.Dependents = append(depEntry.Dependents[:i], depEntry.Dependents[i+1:]...)
					break
				}
			}
		}
	}
}

// autoScan 自动扫描
func (r *DefaultPluginRegistry) autoScan() {
	ticker := time.NewTicker(r.config.ScanInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := r.Scan(ctx); err != nil {
			fmt.Printf("Auto scan error: %v\n", err)
		}
		cancel()
	}
}

// AddWatcher 添加观察器
func (r *DefaultPluginRegistry) AddWatcher(watcher RegistryWatcher) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.watchers = append(r.watchers, watcher)
}

// RemoveWatcher 移除观察器
func (r *DefaultPluginRegistry) RemoveWatcher(watcher RegistryWatcher) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, w := range r.watchers {
		if w == watcher {
			r.watchers = append(r.watchers[:i], r.watchers[i+1:]...)
			break
		}
	}
}

// notifyWatchers 通知观察器
func (r *DefaultPluginRegistry) notifyWatchers(notify func(RegistryWatcher)) {
	for _, watcher := range r.watchers {
		go notify(watcher)
	}
}

// loadCache 加载缓存
func (r *DefaultPluginRegistry) loadCache() error {
	if r.config.CacheFile == "" {
		return nil
	}

	data, err := os.ReadFile(r.config.CacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 缓存文件不存在，忽略
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache map[string]*RegistryEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("failed to parse cache file: %w", err)
	}

	r.mutex.Lock()
	r.plugins = cache
	r.mutex.Unlock()

	return nil
}

// saveCache 保存缓存
func (r *DefaultPluginRegistry) saveCache() error {
	if r.config.CacheFile == "" {
		return nil
	}

	r.mutex.RLock()
	cache := make(map[string]*RegistryEntry)
	for k, v := range r.plugins {
		cache[k] = v
	}
	r.mutex.RUnlock()

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	return os.WriteFile(r.config.CacheFile, data, 0644)
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	Name       string
	Type       core.PluginType
	Runtime    core.PluginRuntime
	Version    string
	Author     string
	Tags       []string
	MinVersion string
	MaxVersion string
	FilterFunc func(*core.PluginMetadata) bool
}

// Match 检查插件是否匹配过滤器
func (f *PluginFilter) Match(metadata *core.PluginMetadata) bool {
	// 检查名称
	if f.Name != "" && metadata.Name != f.Name {
		return false
	}

	// 检查类型
	if f.Type != "" && metadata.Type != f.Type {
		return false
	}

	// 检查运行时
	if f.Runtime != "" && metadata.Runtime != f.Runtime {
		return false
	}

	// 检查版本
	if f.Version != "" && metadata.Version != f.Version {
		return false
	}

	// 检查作者
	if f.Author != "" && metadata.Author != f.Author {
		return false
	}

	// 检查标签
	if len(f.Tags) > 0 {
		for _, tag := range f.Tags {
			found := false
			for _, metaTag := range metadata.Tags {
				if tag == metaTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// 自定义过滤函数
	if f.FilterFunc != nil {
		return f.FilterFunc(metadata)
	}

	return true
}

// GetStats 获取注册表统计信息
func (r *DefaultPluginRegistry) GetStats() *RegistryStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &RegistryStats{
		TotalPlugins: len(r.plugins),
		ByType:       make(map[core.PluginType]int),
		ByRuntime:    make(map[core.PluginRuntime]int),
		ByStatus:     make(map[RegistryStatus]int),
	}

	for _, entry := range r.plugins {
		stats.ByType[entry.Metadata.Type]++
		stats.ByRuntime[entry.Metadata.Runtime]++
		stats.ByStatus[entry.Status]++
	}

	return stats
}

// RegistryStats 注册表统计信息
type RegistryStats struct {
	TotalPlugins int
	ByType       map[core.PluginType]int
	ByRuntime    map[core.PluginRuntime]int
	ByStatus     map[RegistryStatus]int
}
