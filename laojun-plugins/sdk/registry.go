package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginRegistry 插件注册表接口
type PluginRegistry interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(plugin Plugin) error
	
	// UnregisterPlugin 注销插件
	UnregisterPlugin(pluginID string) error
	
	// GetPlugin 获取插件
	GetPlugin(pluginID string) (Plugin, error)
	
	// ListPlugins 列出所有插件
	ListPlugins() []Plugin
	
	// FindPlugins 查找插件
	FindPlugins(filter *PluginFilter) []Plugin
	
	// GetPluginInfo 获取插件信息
	GetPluginInfo(pluginID string) (*PluginInfo, error)
	
	// GetPluginState 获取插件状态
	GetPluginState(pluginID string) (PluginState, error)
	
	// SaveRegistry 保存注册表
	SaveRegistry(path string) error
	
	// LoadRegistry 加载注册表
	LoadRegistry(path string) error
	
	// WatchPlugins 监听插件变化
	WatchPlugins(callback func(event *PluginEvent)) error
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	Name         string              `json:"name,omitempty"`
	Version      string              `json:"version,omitempty"`
	Author       string              `json:"author,omitempty"`
	Category     string              `json:"category,omitempty"`
	Tags         []string            `json:"tags,omitempty"`
	Capabilities []PluginCapability  `json:"capabilities,omitempty"`
	State        PluginState         `json:"state,omitempty"`
	MinVersion   string              `json:"minVersion,omitempty"`
	MaxVersion   string              `json:"maxVersion,omitempty"`
}

// PluginEvent 插件事件
type PluginEvent struct {
	Type      string      `json:"type"`
	PluginID  string      `json:"pluginId"`
	Plugin    Plugin      `json:"plugin,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// PluginEventType 插件事件类型
const (
	PluginEventRegistered   = "registered"
	PluginEventUnregistered = "unregistered"
	PluginEventStateChanged = "state_changed"
	PluginEventUpdated      = "updated"
)

// DefaultPluginRegistry 默认插件注册表实现
type DefaultPluginRegistry struct {
	plugins   map[string]Plugin
	states    map[string]PluginState
	watchers  []func(event *PluginEvent)
	logger    *logrus.Logger
	mu        sync.RWMutex
}

// NewDefaultPluginRegistry 创建默认插件注册表
func NewDefaultPluginRegistry(logger *logrus.Logger) *DefaultPluginRegistry {
	return &DefaultPluginRegistry{
		plugins:  make(map[string]Plugin),
		states:   make(map[string]PluginState),
		watchers: []func(event *PluginEvent){},
		logger:   logger,
	}
}

// RegisterPlugin 注册插件
func (r *DefaultPluginRegistry) RegisterPlugin(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	info := plugin.GetInfo()
	if info == nil {
		return fmt.Errorf("plugin info cannot be nil")
	}

	pluginID := info.ID
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查插件是否已存在
	if _, exists := r.plugins[pluginID]; exists {
		return fmt.Errorf("plugin %s already registered", pluginID)
	}

	// 注册插件
	r.plugins[pluginID] = plugin
	r.states[pluginID] = StateLoaded

	r.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"name":      info.Name,
		"version":   info.Version,
	}).Info("Plugin registered")

	// 触发事件
	r.notifyWatchers(&PluginEvent{
		Type:      PluginEventRegistered,
		PluginID:  pluginID,
		Plugin:    plugin,
		Timestamp: time.Now(),
	})

	return nil
}

// UnregisterPlugin 注销插件
func (r *DefaultPluginRegistry) UnregisterPlugin(pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 删除插件
	delete(r.plugins, pluginID)
	delete(r.states, pluginID)

	r.logger.WithField("plugin_id", pluginID).Info("Plugin unregistered")

	// 触发事件
	r.notifyWatchers(&PluginEvent{
		Type:      PluginEventUnregistered,
		PluginID:  pluginID,
		Plugin:    plugin,
		Timestamp: time.Now(),
	})

	return nil
}

// GetPlugin 获取插件
func (r *DefaultPluginRegistry) GetPlugin(pluginID string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (r *DefaultPluginRegistry) ListPlugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	// 按名称排序
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].GetInfo().Name < plugins[j].GetInfo().Name
	})

	return plugins
}

// FindPlugins 查找插件
func (r *DefaultPluginRegistry) FindPlugins(filter *PluginFilter) []Plugin {
	if filter == nil {
		return r.ListPlugins()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Plugin
	for _, plugin := range r.plugins {
		if r.matchesFilter(plugin, filter) {
			result = append(result, plugin)
		}
	}

	// 按名称排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].GetInfo().Name < result[j].GetInfo().Name
	})

	return result
}

// matchesFilter 检查插件是否匹配过滤器
func (r *DefaultPluginRegistry) matchesFilter(plugin Plugin, filter *PluginFilter) bool {
	info := plugin.GetInfo()
	if info == nil {
		return false
	}

	// 检查名称
	if filter.Name != "" && !strings.Contains(strings.ToLower(info.Name), strings.ToLower(filter.Name)) {
		return false
	}

	// 检查版本
	if filter.Version != "" && info.Version != filter.Version {
		return false
	}

	// 检查作者
	if filter.Author != "" && !strings.Contains(strings.ToLower(info.Author), strings.ToLower(filter.Author)) {
		return false
	}

	// 检查分类
	if filter.Category != "" {
		found := false
		for _, category := range info.Categories {
			if strings.EqualFold(category, filter.Category) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查标签
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, pluginTag := range info.Tags {
				if strings.EqualFold(pluginTag, filterTag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// 检查能力
	if len(filter.Capabilities) > 0 {
		for _, filterCap := range filter.Capabilities {
			found := false
			for _, pluginCap := range info.Capabilities {
				if pluginCap == filterCap {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// 检查状态
	if filter.State != "" {
		if state, exists := r.states[info.ID]; !exists || state != filter.State {
			return false
		}
	}

	// 检查版本范围
	if filter.MinVersion != "" && !r.isVersionGreaterOrEqual(info.Version, filter.MinVersion) {
		return false
	}

	if filter.MaxVersion != "" && !r.isVersionLessOrEqual(info.Version, filter.MaxVersion) {
		return false
	}

	return true
}

// isVersionGreaterOrEqual 检查版本是否大于等于指定版本
func (r *DefaultPluginRegistry) isVersionGreaterOrEqual(version, minVersion string) bool {
	// 简化版本比较，实际应该使用语义版本比较
	return version >= minVersion
}

// isVersionLessOrEqual 检查版本是否小于等于指定版本
func (r *DefaultPluginRegistry) isVersionLessOrEqual(version, maxVersion string) bool {
	// 简化版本比较，实际应该使用语义版本比较
	return version <= maxVersion
}

// GetPluginInfo 获取插件信息
func (r *DefaultPluginRegistry) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	plugin, err := r.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	return plugin.GetInfo(), nil
}

// GetPluginState 获取插件状态
func (r *DefaultPluginRegistry) GetPluginState(pluginID string) (PluginState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, exists := r.states[pluginID]
	if !exists {
		return "", fmt.Errorf("plugin %s not found", pluginID)
	}

	return state, nil
}

// SetPluginState 设置插件状态
func (r *DefaultPluginRegistry) SetPluginState(pluginID string, state PluginState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	oldState := r.states[pluginID]
	r.states[pluginID] = state

	r.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"old_state":  oldState,
		"new_state":  state,
	}).Debug("Plugin state changed")

	// 触发事件
	r.notifyWatchers(&PluginEvent{
		Type:      PluginEventStateChanged,
		PluginID:  pluginID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"old_state": oldState,
			"new_state": state,
		},
	})

	return nil
}

// SaveRegistry 保存注册表
func (r *DefaultPluginRegistry) SaveRegistry(path string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 创建注册表数据
	registryData := &RegistryData{
		Plugins:   make(map[string]*PluginInfo),
		States:    r.states,
		Timestamp: time.Now(),
	}

	for id, plugin := range r.plugins {
		registryData.Plugins[id] = plugin.GetInfo()
	}

	// 序列化数据
	data, err := json.MarshalIndent(registryData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	r.logger.WithField("path", path).Info("Registry saved")
	return nil
}

// LoadRegistry 加载注册表
func (r *DefaultPluginRegistry) LoadRegistry(path string) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		r.logger.WithField("path", path).Info("Registry file not found, starting with empty registry")
		return nil
	}

	// 读取文件
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	// 反序列化数据
	var registryData RegistryData
	if err := json.Unmarshal(data, &registryData); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 加载状态
	r.states = registryData.States
	if r.states == nil {
		r.states = make(map[string]PluginState)
	}

	r.logger.WithFields(logrus.Fields{
		"path":    path,
		"plugins": len(registryData.Plugins),
	}).Info("Registry loaded")

	return nil
}

// WatchPlugins 监听插件变化
func (r *DefaultPluginRegistry) WatchPlugins(callback func(event *PluginEvent)) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.watchers = append(r.watchers, callback)
	return nil
}

// notifyWatchers 通知观察者
func (r *DefaultPluginRegistry) notifyWatchers(event *PluginEvent) {
	for _, watcher := range r.watchers {
		go func(w func(event *PluginEvent)) {
			defer func() {
				if err := recover(); err != nil {
					r.logger.WithField("error", err).Error("Plugin watcher panic")
				}
			}()
			w(event)
		}(watcher)
	}
}

// RegistryData 注册表数据
type RegistryData struct {
	Plugins   map[string]*PluginInfo   `json:"plugins"`
	States    map[string]PluginState   `json:"states"`
	Timestamp time.Time                `json:"timestamp"`
}

// PluginDiscovery 插件发现服务
type PluginDiscovery struct {
	registry    PluginRegistry
	searchPaths []string
	logger      *logrus.Logger
}

// NewPluginDiscovery 创建插件发现服务
func NewPluginDiscovery(registry PluginRegistry, logger *logrus.Logger) *PluginDiscovery {
	return &PluginDiscovery{
		registry:    registry,
		searchPaths: []string{},
		logger:      logger,
	}
}

// AddSearchPath 添加搜索路径
func (d *PluginDiscovery) AddSearchPath(path string) {
	d.searchPaths = append(d.searchPaths, path)
}

// DiscoverPlugins 发现插件
func (d *PluginDiscovery) DiscoverPlugins() error {
	for _, searchPath := range d.searchPaths {
		if err := d.discoverInPath(searchPath); err != nil {
			d.logger.WithError(err).WithField("path", searchPath).Error("Failed to discover plugins in path")
		}
	}
	return nil
}

// discoverInPath 在指定路径中发现插件
func (d *PluginDiscovery) discoverInPath(searchPath string) error {
	return filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 查找插件清单文件
		if info.Name() == "plugin.json" || info.Name() == "plugin.yaml" {
			if err := d.loadPluginFromManifest(path); err != nil {
				d.logger.WithError(err).WithField("manifest", path).Error("Failed to load plugin from manifest")
			}
		}

		return nil
	})
}

// loadPluginFromManifest 从清单文件加载插件
func (d *PluginDiscovery) loadPluginFromManifest(manifestPath string) error {
	// 读取清单文件
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	// 解析清单
	var manifest PluginManifest
	if strings.HasSuffix(manifestPath, ".json") {
		if err := json.Unmarshal(data, &manifest); err != nil {
			return fmt.Errorf("failed to parse JSON manifest: %w", err)
		}
	} else {
		// YAML解析逻辑
		return fmt.Errorf("YAML manifest parsing not implemented")
	}

	// 创建插件信息
	info := &PluginInfo{
		ID:           manifest.ID,
		Name:         manifest.Name,
		Version:      manifest.Version,
		Description:  manifest.Description,
		Author:       manifest.Author,
		License:      manifest.License,
		Homepage:     manifest.Homepage,
		Repository:   manifest.Repository,
		Tags:         manifest.Tags,
		Categories:   manifest.Categories,
		Capabilities: manifest.Capabilities,
		Dependencies: manifest.Dependencies,
		Manifest:     &manifest,
	}

	d.logger.WithFields(logrus.Fields{
		"plugin_id": info.ID,
		"name":      info.Name,
		"version":   info.Version,
		"manifest":  manifestPath,
	}).Info("Discovered plugin")

	return nil
}

// PluginQuery 插件查询构建器
type PluginQuery struct {
	registry PluginRegistry
	filter   *PluginFilter
}

// NewPluginQuery 创建插件查询构建器
func NewPluginQuery(registry PluginRegistry) *PluginQuery {
	return &PluginQuery{
		registry: registry,
		filter:   &PluginFilter{},
	}
}

// WithName 按名称过滤
func (q *PluginQuery) WithName(name string) *PluginQuery {
	q.filter.Name = name
	return q
}

// WithVersion 按版本过滤
func (q *PluginQuery) WithVersion(version string) *PluginQuery {
	q.filter.Version = version
	return q
}

// WithAuthor 按作者过滤
func (q *PluginQuery) WithAuthor(author string) *PluginQuery {
	q.filter.Author = author
	return q
}

// WithCategory 按分类过滤
func (q *PluginQuery) WithCategory(category string) *PluginQuery {
	q.filter.Category = category
	return q
}

// WithTags 按标签过滤
func (q *PluginQuery) WithTags(tags ...string) *PluginQuery {
	q.filter.Tags = tags
	return q
}

// WithCapabilities 按能力过滤
func (q *PluginQuery) WithCapabilities(capabilities ...PluginCapability) *PluginQuery {
	q.filter.Capabilities = capabilities
	return q
}

// WithState 按状态过滤
func (q *PluginQuery) WithState(state PluginState) *PluginQuery {
	q.filter.State = state
	return q
}

// WithVersionRange 按版本范围过滤
func (q *PluginQuery) WithVersionRange(minVersion, maxVersion string) *PluginQuery {
	q.filter.MinVersion = minVersion
	q.filter.MaxVersion = maxVersion
	return q
}

// Find 执行查询
func (q *PluginQuery) Find() []Plugin {
	return q.registry.FindPlugins(q.filter)
}

// First 获取第一个匹配的插件
func (q *PluginQuery) First() (Plugin, error) {
	plugins := q.Find()
	if len(plugins) == 0 {
		return nil, fmt.Errorf("no plugin found matching criteria")
	}
	return plugins[0], nil
}

// Count 获取匹配的插件数量
func (q *PluginQuery) Count() int {
	return len(q.Find())
}

// Exists 检查是否存在匹配的插件
func (q *PluginQuery) Exists() bool {
	return q.Count() > 0
}