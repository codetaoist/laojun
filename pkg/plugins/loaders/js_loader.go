// Package loaders 实现 JavaScript 插件加载中心
package loaders

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"../core"
)

// JSPluginLoader JavaScript 插件加载中心
type JSPluginLoader struct {
	loadedPlugins map[string]*JSPluginInstance
	mutex         sync.RWMutex
}

// NewJSPluginLoader 创建新的 JavaScript 插件加载中心
func NewJSPluginLoader() *JSPluginLoader {
	return &JSPluginLoader{
		loadedPlugins: make(map[string]*JSPluginInstance),
	}
}

// LoadPlugin 加载 JavaScript 插件
func (loader *JSPluginLoader) LoadPlugin(ctx context.Context, path string, metadata *core.PluginMetadata) (core.Plugin, error) {
	// 1. 读取 JavaScript 文件
	jsPath := filepath.Join(path, metadata.EntryPoint)
	if filepath.Ext(jsPath) != ".js" {
		jsPath += ".js"
	}

	jsCode, err := os.ReadFile(jsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JavaScript file %s: %w", jsPath, err)
	}

	// 2. 创建 JavaScript 插件实例
	instance := &JSPluginInstance{
		metadata: metadata,
		jsCode:   string(jsCode),
		loader:   loader,
		status:   core.StatusUnloaded,
	}

	// 3. 保存加载的插件实例
	loader.mutex.Lock()
	loader.loadedPlugins[metadata.ID] = instance
	loader.mutex.Unlock()

	return instance, nil
}

// UnloadPlugin 卸载 JavaScript 插件
func (loader *JSPluginLoader) UnloadPlugin(ctx context.Context, pluginID string) error {
	loader.mutex.Lock()
	defer loader.mutex.Unlock()

	if instance, exists := loader.loadedPlugins[pluginID]; exists {
		// 清理 V8 上下文
		if instance.v8Context != nil {
			instance.v8Context.Close()
		}
		delete(loader.loadedPlugins, pluginID)
	}

	return nil
}

// GetSupportedRuntimes 获取支持的运行时类型
func (loader *JSPluginLoader) GetSupportedRuntimes() []core.PluginRuntime {
	return []core.PluginRuntime{core.RuntimeJS}
}

// ValidatePlugin 验证 JavaScript 插件
func (loader *JSPluginLoader) ValidatePlugin(path string, metadata *core.PluginMetadata) error {
	// 1. 检查运行时类型
	if metadata.Runtime != core.RuntimeJS {
		return fmt.Errorf("unsupported runtime: %s", metadata.Runtime)
	}

	// 2. 检查入口点
	if metadata.EntryPoint == "" {
		return fmt.Errorf("entry point is required for JavaScript plugins")
	}

	// 3. 检查 JavaScript 文件是否存在
	jsPath := filepath.Join(path, metadata.EntryPoint)
	if filepath.Ext(jsPath) != ".js" {
		jsPath += ".js"
	}

	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		return fmt.Errorf("JavaScript file not found: %s", jsPath)
	}

	return nil
}

// JSPluginInstance JavaScript 插件实例
type JSPluginInstance struct {
	metadata  *core.PluginMetadata
	jsCode    string
	loader    *JSPluginLoader
	status    core.PluginStatus
	v8Context *V8Context // 简化的 V8 上下文接口（实际实现需要使用 V8 绑定库）
	sandbox   core.Sandbox
	mutex     sync.RWMutex
}

// V8Context 简化的 V8 上下文接口（实际实现需要使用 V8 绑定库）
type V8Context struct {
	// 这里是简化版本，实际需要使用如 v8go 等库
	isolate   interface{}            // V8 Isolate 实例
	context   interface{}            // V8 Context 实例
	functions map[string]interface{} // 存储 JavaScript 函数的映射
}

// NewV8Context 创建新的 V8 上下文实例
func NewV8Context() *V8Context {
	return &V8Context{
		functions: make(map[string]interface{}),
	}
}

// ExecuteScript 执行 JavaScript 代码
func (v8 *V8Context) ExecuteScript(script string) (interface{}, error) {
	// 这里是简化实现，实际需要使用 V8 引擎执行 JavaScript 代码
	// 在真实实现中，这里会调用 V8 引擎执行 JavaScript 代码
	return nil, fmt.Errorf("V8 execution not implemented in this demo")
}

// CallFunction 调用 JavaScript 函数
func (v8 *V8Context) CallFunction(name string, args ...interface{}) (interface{}, error) {
	// 这里是简化实现，实际需要使用 V8 引擎调用 JavaScript 函数
	return nil, fmt.Errorf("V8 function call not implemented in this demo")
}

// Close 关闭 V8 上下文实例
func (v8 *V8Context) Close() {
	// 清理 V8 资源
}

// GetMetadata 获取插件元数据
func (instance *JSPluginInstance) GetMetadata() *core.PluginMetadata {
	return instance.metadata
}

// Initialize 初始化插件
func (instance *JSPluginInstance) Initialize(ctx context.Context, config map[string]any) error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	if instance.status != core.StatusUnloaded {
		return fmt.Errorf("plugin already initialized")
	}

	// 1. 创建 V8 上下文实例
	instance.v8Context = NewV8Context()

	// 2. 设置沙箱环境
	sandboxConfig := &core.SandboxConfig{
		MaxMemory:     64 * 1024 * 1024, // 64MB
		MaxCPUTime:    30 * time.Second,
		MaxGoroutines: 10,
		AllowedAPIs:   []string{"console", "setTimeout", "clearTimeout"},
		NetworkAccess: false,
	}

	// 这里需要实际的沙箱实现
	// instance.sandbox = securityManager.CreateSandbox(instance.metadata.ID, sandboxConfig)

	// 3. 执行插件代码
	_, err := instance.v8Context.ExecuteScript(instance.jsCode)
	if err != nil {
		instance.status = core.StatusError
		return fmt.Errorf("failed to execute plugin script: %w", err)
	}

	// 4. 调用插件初始化函数
	configJSON, _ := json.Marshal(config)
	_, err = instance.v8Context.CallFunction("initialize", string(configJSON))
	if err != nil {
		instance.status = core.StatusError
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	instance.status = core.StatusLoaded
	return nil
}

// Start 启动插件
func (instance *JSPluginInstance) Start(ctx context.Context) error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	if instance.status != core.StatusLoaded && instance.status != core.StatusStopped {
		return fmt.Errorf("plugin not in startable state, current status: %s", instance.status)
	}

	// 调用插件启动函数
	_, err := instance.v8Context.CallFunction("start")
	if err != nil {
		instance.status = core.StatusError
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	instance.status = core.StatusStarted
	return nil
}

// Stop 停止插件
func (instance *JSPluginInstance) Stop(ctx context.Context) error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	if instance.status != core.StatusStarted {
		return fmt.Errorf("plugin not started")
	}

	// 调用插件停止函数
	_, err := instance.v8Context.CallFunction("stop")
	if err != nil {
		instance.status = core.StatusError
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	instance.status = core.StatusStopped
	return nil
}

// HandleRequest 处理请求
func (instance *JSPluginInstance) HandleRequest(ctx context.Context, req *core.PluginRequest) (*core.PluginResponse, error) {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	if instance.status != core.StatusStarted {
		return nil, fmt.Errorf("plugin not started")
	}

	// 序列化请求
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// 调用插件处理函数
	result, err := instance.v8Context.CallFunction("handleRequest", string(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("plugin request handling failed: %w", err)
	}

	// 反序列化响应
	var response core.PluginResponse
	if resultStr, ok := result.(string); ok {
		if err := json.Unmarshal([]byte(resultStr), &response); err != nil {
			return nil, fmt.Errorf("failed to deserialize response: %w", err)
		}
	} else {
		return nil, fmt.Errorf("invalid response format from plugin")
	}

	return &response, nil
}

// HandleEvent 处理事件
func (instance *JSPluginInstance) HandleEvent(ctx context.Context, event *core.PluginEvent) error {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	if instance.status != core.StatusStarted {
		return fmt.Errorf("plugin not started")
	}

	// 序列化事件
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// 调用插件事件处理函数
	_, err = instance.v8Context.CallFunction("handleEvent", string(eventJSON))
	if err != nil {
		return fmt.Errorf("plugin event handling failed: %w", err)
	}

	return nil
}

// GetStatus 获取插件状态
func (instance *JSPluginInstance) GetStatus() core.PluginStatus {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	return instance.status
}

// GetHealth 获取健康状态
func (instance *JSPluginInstance) GetHealth(ctx context.Context) (map[string]any, error) {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	health := map[string]any{
		"status":  instance.status,
		"healthy": instance.status == core.StatusStarted,
	}

	if instance.status == core.StatusStarted {
		// 调用插件健康检查函数
		result, err := instance.v8Context.CallFunction("getHealth")
		if err == nil {
			if healthStr, ok := result.(string); ok {
				var pluginHealth map[string]any
				if json.Unmarshal([]byte(healthStr), &pluginHealth) == nil {
					for k, v := range pluginHealth {
						health[k] = v
					}
				}
			}
		}
	}

	return health, nil
}
