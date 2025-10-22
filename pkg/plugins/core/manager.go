// Package core 实现插件管理接口
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DefaultPluginManager 默认插件管理器实现
type DefaultPluginManager struct {
	loaders  map[PluginRuntime]PluginLoader
	plugins  map[string]Plugin
	metadata map[string]*PluginMetadata
	registry PluginRegistry
	security SecurityManager
	eventBus EventBus
	config   *PluginConfig
	mutex    sync.RWMutex
	started  bool
}

// NewPluginManager 创建新的插件管理器实例
func NewPluginManager(config *PluginConfig, registry PluginRegistry, security SecurityManager, eventBus EventBus) *DefaultPluginManager {
	return &DefaultPluginManager{
		loaders:  make(map[PluginRuntime]PluginLoader),
		plugins:  make(map[string]Plugin),
		metadata: make(map[string]*PluginMetadata),
		registry: registry,
		security: security,
		eventBus: eventBus,
		config:   config,
	}
}

// RegisterLoader 注册插件加载器
func (pm *DefaultPluginManager) RegisterLoader(runtime PluginRuntime, loader PluginLoader) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if loader == nil {
		return fmt.Errorf("loader cannot be nil")
	}

	pm.loaders[runtime] = loader
	return nil
}

// LoadPlugin 加载插件
func (pm *DefaultPluginManager) LoadPlugin(ctx context.Context, path string) error {
	// 1. 读取插件元数据
	metadata, err := pm.loadPluginMetadata(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin metadata: %w", err)
	}

	// 2. 检查插件是否已存在
	pm.mutex.RLock()
	if _, exists := pm.plugins[metadata.ID]; exists {
		pm.mutex.RUnlock()
		return fmt.Errorf("plugin %s already loaded", metadata.ID)
	}
	pm.mutex.RUnlock()

	// 3. 安全检查
	if pm.config.EnableSecurity {
		if err := pm.security.ValidatePermissions(metadata.ID, metadata.Permissions); err != nil {
			return fmt.Errorf("permission validation failed: %w", err)
		}

		report, err := pm.security.ScanCode(path)
		if err != nil {
			return fmt.Errorf("security scan failed: %w", err)
		}

		if !report.Passed {
			return fmt.Errorf("security scan failed: plugin contains security issues")
		}
	}

	// 4. 获取对应的加载器
	pm.mutex.RLock()
	loader, exists := pm.loaders[metadata.Runtime]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no loader found for runtime %s", metadata.Runtime)
	}

	// 5. 验证插件
	if err := loader.ValidatePlugin(path, metadata); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// 6. 加载插件
	plugin, err := loader.LoadPlugin(ctx, path, metadata)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// 7. 初始化插件
	if err := plugin.Initialize(ctx, metadata.Config); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 8. 注册插件
	pm.mutex.Lock()
	pm.plugins[metadata.ID] = plugin
	pm.metadata[metadata.ID] = metadata
	pm.mutex.Unlock()

	// 9. 注册到注册表
	if err := pm.registry.Register(metadata); err != nil {
		// 回滚
		pm.mutex.Lock()
		delete(pm.plugins, metadata.ID)
		delete(pm.metadata, metadata.ID)
		pm.mutex.Unlock()
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	// 10. 发布加载事件
	event := &PluginEvent{
		Type:      "plugin.loaded",
		Source:    "plugin_manager",
		Target:    metadata.ID,
		Data:      metadata,
		Timestamp: time.Now(),
	}

	if err := pm.eventBus.PublishAsync(ctx, event); err != nil {
		// 记录日志但不失败
		fmt.Printf("Failed to publish plugin loaded event: %v\n", err)
	}

	return nil
}

// UnloadPlugin 卸载插件
func (pm *DefaultPluginManager) UnloadPlugin(ctx context.Context, pluginID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 1. 检查插件是否存在
	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	metadata := pm.metadata[pluginID]

	// 2. 停止插件（如果正在运行）
	if plugin.GetStatus() == StatusStarted {
		if err := plugin.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop plugin before unloading: %w", err)
		}
	}

	// 3. 获取加载器并卸载
	loader, exists := pm.loaders[metadata.Runtime]
	if exists {
		if err := loader.UnloadPlugin(ctx, pluginID); err != nil {
			return fmt.Errorf("failed to unload plugin: %w", err)
		}
	}

	// 4. 从注册表中移除
	if err := pm.registry.Unregister(pluginID); err != nil {
		return fmt.Errorf("failed to unregister plugin: %w", err)
	}

	// 5. 从内存中移除
	delete(pm.plugins, pluginID)
	delete(pm.metadata, pluginID)

	// 6. 发布卸载事件
	event := &PluginEvent{
		Type:      "plugin.unloaded",
		Source:    "plugin_manager",
		Target:    pluginID,
		Data:      metadata,
		Timestamp: time.Now(),
	}

	if err := pm.eventBus.PublishAsync(ctx, event); err != nil {
		fmt.Printf("Failed to publish plugin unloaded event: %v\n", err)
	}

	return nil
}

// StartPlugin 启动插件
func (pm *DefaultPluginManager) StartPlugin(ctx context.Context, pluginID string) error {
	pm.mutex.RLock()
	plugin, exists := pm.plugins[pluginID]
	metadata := pm.metadata[pluginID]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 检查插件状态
	if plugin.GetStatus() == StatusStarted {
		return fmt.Errorf("plugin %s is already started", pluginID)
	}

	// 启动插件
	if err := plugin.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// 发布启动事件
	event := &PluginEvent{
		Type:      "plugin.started",
		Source:    "plugin_manager",
		Target:    pluginID,
		Data:      metadata,
		Timestamp: time.Now(),
	}

	if err := pm.eventBus.PublishAsync(ctx, event); err != nil {
		fmt.Printf("Failed to publish plugin started event: %v\n", err)
	}

	return nil
}

// StopPlugin 停止插件
func (pm *DefaultPluginManager) StopPlugin(ctx context.Context, pluginID string) error {
	pm.mutex.RLock()
	plugin, exists := pm.plugins[pluginID]
	metadata := pm.metadata[pluginID]
	pm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 检查插件状态
	if plugin.GetStatus() != StatusStarted {
		return fmt.Errorf("plugin %s is not started", pluginID)
	}

	// 停止插件
	if err := plugin.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// 发布停止事件
	event := &PluginEvent{
		Type:      "plugin.stopped",
		Source:    "plugin_manager",
		Target:    pluginID,
		Data:      metadata,
		Timestamp: time.Now(),
	}

	if err := pm.eventBus.PublishAsync(ctx, event); err != nil {
		fmt.Printf("Failed to publish plugin stopped event: %v\n", err)
	}

	return nil
}

// GetPlugin 获取插件实例
func (pm *DefaultPluginManager) GetPlugin(pluginID string) (Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (pm *DefaultPluginManager) ListPlugins() []*PluginMetadata {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make([]*PluginMetadata, 0, len(pm.metadata))
	for _, metadata := range pm.metadata {
		result = append(result, metadata)
	}

	return result
}

// CallPlugin 调用插件方法
func (pm *DefaultPluginManager) CallPlugin(ctx context.Context, pluginID string, method string, params map[string]any) (*PluginResponse, error) {
	plugin, err := pm.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	// 检查插件状态
	if plugin.GetStatus() != StatusStarted {
		return nil, fmt.Errorf("plugin %s is not started", pluginID)
	}

	// 创建请求
	req := &PluginRequest{
		ID:      fmt.Sprintf("%s_%d", pluginID, time.Now().UnixNano()),
		Method:  method,
		Params:  params,
		Context: ctx,
		Metadata: map[string]any{
			"plugin_id": pluginID,
			"timestamp": time.Now(),
		},
	}

	// 调用插件
	response, err := plugin.HandleRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("plugin call failed: %w", err)
	}

	return response, nil
}

// BroadcastEvent 广播事件
func (pm *DefaultPluginManager) BroadcastEvent(ctx context.Context, event *PluginEvent) error {
	return pm.eventBus.Publish(ctx, event)
}

// loadPluginMetadata 加载插件元数据
func (pm *DefaultPluginManager) loadPluginMetadata(path string) (*PluginMetadata, error) {
	// 查找 manifest.json 文件
	manifestPath := filepath.Join(path, "manifest.json")

	// 检查文件是否存在
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest.json not found in %s", path)
	}

	// 读取文件内容
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %w", err)
	}

	// 解析 JSON
	var metadata PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}

	// 验证必要字段
	if metadata.ID == "" {
		return nil, fmt.Errorf("plugin ID is required")
	}

	if metadata.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}

	if metadata.Version == "" {
		return nil, fmt.Errorf("plugin version is required")
	}

	if metadata.Runtime == "" {
		return nil, fmt.Errorf("plugin runtime is required")
	}

	if metadata.Type == "" {
		metadata.Type = PluginTypeInProcess // 默认为进程内插件
	}

	// 设置时间戳
	now := time.Now()
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = now
	}
	metadata.UpdatedAt = now

	return &metadata, nil
}

// Start 启动插件管理
func (pm *DefaultPluginManager) Start(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.started {
		return fmt.Errorf("plugin manager already started")
	}

	// 自动加载插件目录中的插件
	for _, pluginPath := range pm.config.PluginPaths {
		if err := pm.loadPluginsFromPath(ctx, pluginPath); err != nil {
			fmt.Printf("Failed to load plugins from %s: %v\n", pluginPath, err)
		}
	}

	pm.started = true
	return nil
}

// Stop 停止插件管理
func (pm *DefaultPluginManager) Stop(ctx context.Context) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.started {
		return nil
	}

	// 停止所有插件
	for pluginID := range pm.plugins {
		if err := pm.StopPlugin(ctx, pluginID); err != nil {
			fmt.Printf("Failed to stop plugin %s: %v\n", pluginID, err)
		}
	}

	pm.started = false
	return nil
}

// loadPluginsFromPath 从指定路径加载插件
func (pm *DefaultPluginManager) loadPluginsFromPath(ctx context.Context, basePath string) error {
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过文件，只处理目录
		if !info.IsDir() {
			return nil
		}

		// 检查是否包含 manifest.json
		manifestPath := filepath.Join(path, "manifest.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return nil // 跳过没有 manifest.json 的目录
		}

		// 尝试加载插件
		if err := pm.LoadPlugin(ctx, path); err != nil {
			fmt.Printf("Failed to load plugin from %s: %v\n", path, err)
		}

		return nil
	})
}
