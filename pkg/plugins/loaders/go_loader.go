// Package loaders 实现各种插件加载中心
package loaders

import (
	"context"
	"fmt"
	"path/filepath"
	"plugin"
	"sync"

	"../core"
)

// GoPluginLoader Go 插件加载中心
type GoPluginLoader struct {
	loadedPlugins map[string]*plugin.Plugin
	mutex         sync.RWMutex
}

// NewGoPluginLoader 创建新的 Go 插件加载中心
func NewGoPluginLoader() *GoPluginLoader {
	return &GoPluginLoader{
		loadedPlugins: make(map[string]*plugin.Plugin),
	}
}

// LoadPlugin 加载 Go 插件
func (loader *GoPluginLoader) LoadPlugin(ctx context.Context, path string, metadata *core.PluginMetadata) (core.Plugin, error) {
	// 1. 构建 .so 文件路径
	soPath := filepath.Join(path, metadata.EntryPoint)
	if filepath.Ext(soPath) != ".so" {
		soPath += ".so"
	}

	// 2. 加载 Go 插件
	p, err := plugin.Open(soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", soPath, err)
	}

	// 3. 查找插件工厂函数
	factorySymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("NewPlugin function not found in plugin: %w", err)
	}

	// 4. 类型断言工厂函数
	factory, ok := factorySymbol.(func() core.Plugin)
	if !ok {
		return nil, fmt.Errorf("NewPlugin function has invalid signature")
	}

	// 5. 创建插件实例
	pluginInstance := factory()
	if pluginInstance == nil {
		return nil, fmt.Errorf("plugin factory returned nil")
	}

	// 6. 包装插件实例
	wrappedPlugin := &GoPluginWrapper{
		plugin:   pluginInstance,
		metadata: metadata,
		loader:   loader,
	}

	// 7. 保存加载的插件
	loader.mutex.Lock()
	loader.loadedPlugins[metadata.ID] = p
	loader.mutex.Unlock()

	return wrappedPlugin, nil
}

// UnloadPlugin 卸载 Go 插件
func (loader *GoPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	loader.mutex.Lock()
	defer loader.mutex.Unlock()

	// Go 插件无法真正卸载，只能从映射中移除
	delete(loader.loadedPlugins, pluginID)
	return nil
}

// GetSupportedRuntimes 获取支持的运行时类型
func (loader *GoPluginLoader) GetSupportedRuntimes() []core.PluginRuntime {
	return []core.PluginRuntime{core.RuntimeGo}
}

// ValidatePlugin 验证 Go 插件
func (loader *GoPluginLoader) ValidatePlugin(path string, metadata *core.PluginMetadata) error {
	// 1. 检查运行时类型
	if metadata.Runtime != core.RuntimeGo {
		return fmt.Errorf("unsupported runtime: %s", metadata.Runtime)
	}

	// 2. 检查入口点
	if metadata.EntryPoint == "" {
		return fmt.Errorf("entry point is required for Go plugins")
	}

	// 3. 检查 .so 文件是否存在
	soPath := filepath.Join(path, metadata.EntryPoint)
	if filepath.Ext(soPath) != ".so" {
		soPath += ".so"
	}

	// 这里可以添加更多验证逻辑，比如检查文件签名等

	return nil
}

// GoPluginWrapper Go 插件包装中心
type GoPluginWrapper struct {
	plugin   core.Plugin
	metadata *core.PluginMetadata
	loader   *GoPluginLoader
	status   core.PluginStatus
	mutex    sync.RWMutex
}

// GetMetadata 获取插件元数据
func (wrapper *GoPluginWrapper) GetMetadata() *core.PluginMetadata {
	return wrapper.metadata
}

// Initialize 初始化插件
func (wrapper *GoPluginWrapper) Initialize(ctx context.Context, config map[string]any) error {
	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	if wrapper.status != core.StatusUnloaded {
		return fmt.Errorf("plugin already initialized")
	}

	err := wrapper.plugin.Initialize(ctx, config)
	if err != nil {
		wrapper.status = core.StatusError
		return err
	}

	wrapper.status = core.StatusLoaded
	return nil
}

// Start 启动插件
func (wrapper *GoPluginWrapper) Start(ctx context.Context) error {
	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	if wrapper.status != core.StatusLoaded && wrapper.status != core.StatusStopped {
		return fmt.Errorf("plugin not in loadable state, current status: %s", wrapper.status)
	}

	err := wrapper.plugin.Start(ctx)
	if err != nil {
		wrapper.status = core.StatusError
		return err
	}

	wrapper.status = core.StatusStarted
	return nil
}

// Stop 停止插件
func (wrapper *GoPluginWrapper) Stop(ctx context.Context) error {
	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	if wrapper.status != core.StatusStarted {
		return fmt.Errorf("plugin not started")
	}

	err := wrapper.plugin.Stop(ctx)
	if err != nil {
		wrapper.status = core.StatusError
		return err
	}

	wrapper.status = core.StatusStopped
	return nil
}

// HandleRequest 处理请求
func (wrapper *GoPluginWrapper) HandleRequest(ctx context.Context, req *core.PluginRequest) (*core.PluginResponse, error) {
	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	if wrapper.status != core.StatusStarted {
		return nil, fmt.Errorf("plugin not started")
	}

	return wrapper.plugin.HandleRequest(ctx, req)
}

// HandleEvent 处理事件
func (wrapper *GoPluginWrapper) HandleEvent(ctx context.Context, event *core.PluginEvent) error {
	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	if wrapper.status != core.StatusStarted {
		return fmt.Errorf("plugin not started")
	}

	return wrapper.plugin.HandleEvent(ctx, event)
}

// GetStatus 获取插件状态
func (wrapper *GoPluginWrapper) GetStatus() core.PluginStatus {
	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	return wrapper.status
}

// GetHealth 获取健康状态
func (wrapper *GoPluginWrapper) GetHealth(ctx context.Context) (map[string]any, error) {
	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	if wrapper.status != core.StatusStarted {
		return map[string]any{
			"status":  wrapper.status,
			"healthy": false,
		}, nil
	}

	health, err := wrapper.plugin.GetHealth(ctx)
	if err != nil {
		return map[string]any{
			"status":  wrapper.status,
			"healthy": false,
			"error":   err.Error(),
		}, nil
	}

	health["status"] = wrapper.status
	health["healthy"] = true
	return health, nil
}
