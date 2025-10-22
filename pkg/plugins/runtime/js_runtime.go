package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/taishanglaojun/plugins/core"
)

// JSRuntime JavaScript运行时环境
type JSRuntime struct {
	instances   map[string]*JSPluginInstance
	mutex       sync.RWMutex
	config      *JSRuntimeConfig
	monitor     *JSResourceMonitor
	securityMgr core.SecurityManager
	v8Pool      *V8ContextPool
}

// JSPluginInstance JavaScript插件实例
type JSPluginInstance struct {
	ID         string
	Code       string
	Metadata   *core.PluginMetadata
	Status     core.PluginStatus
	LoadTime   time.Time
	LastAccess time.Time
	Config     map[string]any
	Context    *V8Context
	Sandbox    core.Sandbox
	Stats      *JSPluginStats
	mutex      sync.RWMutex
}

// JSRuntimeConfig JavaScript运行时配置
type JSRuntimeConfig struct {
	MaxInstances     int           `json:"max_instances"`
	ContextPoolSize  int           `json:"context_pool_size"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
	MemoryLimit      int64         `json:"memory_limit"`
	EnableDebug      bool          `json:"enable_debug"`
	AllowedAPIs      []string      `json:"allowed_apis"`
	StatsInterval    time.Duration `json:"stats_interval"`
}

// JSPluginStats JavaScript插件统计信息
type JSPluginStats struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastError       string        `json:"last_error"`
	LastErrorTime   time.Time     `json:"last_error_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	HeapUsage       int64         `json:"heap_usage"`
	CompilationTime time.Duration `json:"compilation_time"`
	GCCount         int64         `json:"gc_count"`
}

// V8Context V8上下文封装
type V8Context struct {
	ID          string
	CreatedAt   time.Time
	LastUsed    time.Time
	MemoryUsage int64
	IsActive    bool
	mutex       sync.Mutex
	// 这里应该包含实际的V8上下文，为了简化暂时用interface{}
	context      interface{}
	globalObject map[string]interface{}
}

// V8ContextPool V8上下文池
type V8ContextPool struct {
	contexts  []*V8Context
	available chan *V8Context
	maxSize   int
	mutex     sync.Mutex
	config    *JSRuntimeConfig
}

// JSResourceMonitor JavaScript资源监控
type JSResourceMonitor struct {
	enabled     bool
	interval    time.Duration
	stopCh      chan struct{}
	instances   map[string]*JSPluginInstance
	mutex       sync.RWMutex
	memoryLimit int64
}

// NewJSRuntime 创建JavaScript运行时环境
func NewJSRuntime(config *JSRuntimeConfig, securityMgr core.SecurityManager) *JSRuntime {
	runtime := &JSRuntime{
		instances:   make(map[string]*JSPluginInstance),
		config:      config,
		securityMgr: securityMgr,
		v8Pool:      NewV8ContextPool(config),
		monitor: &JSResourceMonitor{
			enabled:     true,
			interval:    config.StatsInterval,
			stopCh:      make(chan struct{}),
			instances:   make(map[string]*JSPluginInstance),
			memoryLimit: config.MemoryLimit,
		},
	}

	go runtime.monitor.start()
	return runtime
}

// NewV8ContextPool 创建V8上下文池
func NewV8ContextPool(config *JSRuntimeConfig) *V8ContextPool {
	pool := &V8ContextPool{
		contexts:  make([]*V8Context, 0, config.ContextPoolSize),
		available: make(chan *V8Context, config.ContextPoolSize),
		maxSize:   config.ContextPoolSize,
		config:    config,
	}

	// 预创建上下文
	for i := 0; i < config.ContextPoolSize; i++ {
		ctx := pool.createContext()
		pool.contexts = append(pool.contexts, ctx)
		pool.available <- ctx
	}

	return pool
}

// createContext 创建V8上下文封装
func (pool *V8ContextPool) createContext() *V8Context {
	ctx := &V8Context{
		ID:           fmt.Sprintf("ctx_%d", time.Now().UnixNano()),
		CreatedAt:    time.Now(),
		LastUsed:     time.Now(),
		IsActive:     false,
		globalObject: make(map[string]interface{}),
	}

	// 初始化V8上下文封装
	ctx.initializeV8Context(pool.config)
	return ctx
}

// initializeV8Context 初始化V8上下文封装
func (ctx *V8Context) initializeV8Context(config *JSRuntimeConfig) {
	// 这里应该初始化实际的V8上下文封装
	// 为了简化，我们使用模拟实现

	// 设置全局对象
	ctx.globalObject["console"] = map[string]interface{}{
		"log":   ctx.createConsoleLog(),
		"error": ctx.createConsoleError(),
		"warn":  ctx.createConsoleWarn(),
	}

	// 根据配置添加允许的API
	for _, api := range config.AllowedAPIs {
		switch api {
		case "http":
			ctx.globalObject["http"] = ctx.createHttpAPI()
		case "json":
			ctx.globalObject["JSON"] = ctx.createJsonAPI()
		case "crypto":
			ctx.globalObject["crypto"] = ctx.createCryptoAPI()
		}
	}
}

// GetContext 获取可用上下文封装
func (pool *V8ContextPool) GetContext() (*V8Context, error) {
	select {
	case ctx := <-pool.available:
		ctx.mutex.Lock()
		ctx.IsActive = true
		ctx.LastUsed = time.Now()
		ctx.mutex.Unlock()
		return ctx, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for available context")
	}
}

// ReturnContext 归还上下文封装
func (pool *V8ContextPool) ReturnContext(ctx *V8Context) {
	ctx.mutex.Lock()
	ctx.IsActive = false
	ctx.LastUsed = time.Now()
	ctx.mutex.Unlock()

	// 清理上下文状态
	ctx.cleanup()

	select {
	case pool.available <- ctx:
	default:
		// 池已满，丢弃上下文封装
	}
}

// LoadPlugin 加载JavaScript插件
func (r *JSRuntime) LoadPlugin(ctx context.Context, pluginPath string, config map[string]any) (*JSPluginInstance, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查实例数量限制
	if len(r.instances) >= r.config.MaxInstances {
		return nil, fmt.Errorf("maximum instance limit reached: %d", r.config.MaxInstances)
	}

	// 读取JavaScript文件
	code, err := ioutil.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin file: %w", err)
	}

	// 获取V8上下文封装
	v8Context, err := r.v8Pool.GetContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get V8 context: %w", err)
	}

	// 编译和执行插件代码
	startTime := time.Now()
	metadata, err := r.compileAndExecute(v8Context, string(code))
	compilationTime := time.Since(startTime)

	if err != nil {
		r.v8Pool.ReturnContext(v8Context)
		return nil, fmt.Errorf("failed to compile plugin: %w", err)
	}

	// 创建沙箱
	sandbox, err := r.securityMgr.CreateSandbox(ctx, metadata.ID, &core.SecurityPolicy{
		AllowedAPIs: r.config.AllowedAPIs,
		MemoryLimit: r.config.MemoryLimit,
		TimeLimit:   r.config.ExecutionTimeout,
	})
	if err != nil {
		r.v8Pool.ReturnContext(v8Context)
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	// 创建插件实例
	instance := &JSPluginInstance{
		ID:         metadata.ID,
		Code:       string(code),
		Metadata:   metadata,
		Status:     core.StatusLoaded,
		LoadTime:   time.Now(),
		LastAccess: time.Now(),
		Config:     config,
		Context:    v8Context,
		Sandbox:    sandbox,
		Stats: &JSPluginStats{
			CompilationTime: compilationTime,
		},
	}

	// 初始化插件
	if err := r.initializePlugin(ctx, instance, config); err != nil {
		r.v8Pool.ReturnContext(v8Context)
		sandbox.Destroy(ctx)
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 注册插件
	r.instances[metadata.ID] = instance
	r.monitor.addInstance(instance)

	return instance, nil
}

// compileAndExecute 编译并执行JavaScript代码
func (r *JSRuntime) compileAndExecute(ctx *V8Context, code string) (*core.PluginMetadata, error) {
	// 这里应该使用实际的V8引擎编译和执行代码
	// 为了简化，我们使用模拟实现

	// 执行代码获取插件元数据
	result, err := ctx.executeScript(code)
	if err != nil {
		return nil, err
	}

	// 解析插件元数据
	metadataMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid plugin metadata format")
	}

	metadata := &core.PluginMetadata{}
	if err := r.parseMetadata(metadataMap, metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

// executeScript 执行JavaScript脚本
func (ctx *V8Context) executeScript(code string) (interface{}, error) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	// 这里应该使用实际的V8引擎执行代码
	// 为了简化，我们返回模拟的插件元数据
	return map[string]interface{}{
		"id":          "js-plugin-" + ctx.ID,
		"name":        "JavaScript Plugin",
		"version":     "1.0.0",
		"description": "A JavaScript plugin",
		"type":        "filter",
		"runtime":     "js",
		"author":      "Developer",
	}, nil
}

// parseMetadata 解析插件元数据
func (r *JSRuntime) parseMetadata(data map[string]interface{}, metadata *core.PluginMetadata) error {
	if id, ok := data["id"].(string); ok {
		metadata.ID = id
	}
	if name, ok := data["name"].(string); ok {
		metadata.Name = name
	}
	if version, ok := data["version"].(string); ok {
		metadata.Version = version
	}
	if description, ok := data["description"].(string); ok {
		metadata.Description = description
	}
	if author, ok := data["author"].(string); ok {
		metadata.Author = author
	}

	metadata.Runtime = core.RuntimeJS
	metadata.Type = core.TypeFilter // 默认类型

	return nil
}

// initializePlugin 初始化插件
func (r *JSRuntime) initializePlugin(ctx context.Context, instance *JSPluginInstance, config map[string]any) error {
	// 调用插件的初始化函数
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 在沙箱中执行初始化函数
	return instance.Sandbox.Execute(ctx, func() error {
		return instance.Context.callFunction("initialize", string(configJSON))
	})
}

// ExecutePlugin 执行插件请求
func (r *JSRuntime) ExecutePlugin(ctx context.Context, pluginID string, req *core.PluginRequest) (*core.PluginResponse, error) {
	instance, err := r.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	// 记录开始时间
	startTime := time.Now()

	// 序列化请求
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 在沙箱中执行
	var responseJSON string
	err = instance.Sandbox.Execute(ctx, func() error {
		result, execErr := instance.Context.callFunction("handleRequest", string(reqJSON))
		if execErr != nil {
			return execErr
		}
		responseJSON = result
		return nil
	})

	// 更新统计信息
	duration := time.Since(startTime)
	instance.updateStats(duration, err)

	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// 反序列化响应
	var response core.PluginResponse
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetPlugin 获取插件实例
func (r *JSRuntime) GetPlugin(pluginID string) (*JSPluginInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.instances[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	instance.mutex.Lock()
	instance.LastAccess = time.Now()
	instance.mutex.Unlock()

	return instance, nil
}

// UnloadPlugin 卸载插件
func (r *JSRuntime) UnloadPlugin(ctx context.Context, pluginID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.instances[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 停止插件
	if instance.Status == core.StatusRunning {
		instance.Status = core.StatusStopped
	}

	// 销毁沙箱
	if err := instance.Sandbox.Destroy(ctx); err != nil {
		return fmt.Errorf("failed to destroy sandbox: %w", err)
	}

	// 归还V8上下文
	r.v8Pool.ReturnContext(instance.Context)

	// 移除插件
	delete(r.instances, pluginID)
	r.monitor.removeInstance(pluginID)

	instance.Status = core.StatusUnloaded
	return nil
}

// V8Context 方法

// callFunction 调用JavaScript函数
func (ctx *V8Context) callFunction(funcName string, args ...string) (string, error) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	// 这里应该使用实际的V8引擎调用函数
	// 为了简化，我们返回模拟结果
	switch funcName {
	case "initialize":
		return "", nil
	case "handleRequest":
		return `{"success": true, "data": {"message": "Hello from JS plugin"}}`, nil
	default:
		return "", fmt.Errorf("function not found: %s", funcName)
	}
}

// cleanup 清理上下文状态
func (ctx *V8Context) cleanup() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	// 清理全局变量
	// 重置内存使用
	ctx.MemoryUsage = 0
}

// 创建控制台API
func (ctx *V8Context) createConsoleLog() interface{} {
	return func(args ...interface{}) {
		// 实现console.log
	}
}

func (ctx *V8Context) createConsoleError() interface{} {
	return func(args ...interface{}) {
		// 实现console.error
	}
}

func (ctx *V8Context) createConsoleWarn() interface{} {
	return func(args ...interface{}) {
		// 实现console.warn
	}
}

// 创建HTTP API
func (ctx *V8Context) createHttpAPI() interface{} {
	return map[string]interface{}{
		"get":  func(url string) interface{} { return nil },
		"post": func(url string, data interface{}) interface{} { return nil },
	}
}

// 创建JSON API
func (ctx *V8Context) createJsonAPI() interface{} {
	return map[string]interface{}{
		"parse":     func(str string) interface{} { return nil },
		"stringify": func(obj interface{}) string { return "" },
	}
}

// 创建Crypto API
func (ctx *V8Context) createCryptoAPI() interface{} {
	return map[string]interface{}{
		"randomUUID": func() string { return "" },
		"hash":       func(data string) string { return "" },
	}
}

// updateStats 更新插件统计信息
func (instance *JSPluginInstance) updateStats(duration time.Duration, err error) {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	instance.Stats.RequestCount++
	instance.Stats.TotalDuration += duration

	if instance.Stats.RequestCount > 0 {
		instance.Stats.AverageDuration = instance.Stats.TotalDuration / time.Duration(instance.Stats.RequestCount)
	}

	if err != nil {
		instance.Stats.ErrorCount++
		instance.Stats.LastError = err.Error()
		instance.Stats.LastErrorTime = time.Now()
	}

	// 更新内存使用情况
	instance.Stats.MemoryUsage = instance.Context.MemoryUsage
}

// JSResourceMonitor 方法

// start 启动资源监控
func (m *JSResourceMonitor) start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectStats()
		case <-m.stopCh:
			return
		}
	}
}

// addInstance 添加实例到监控
func (m *JSResourceMonitor) addInstance(instance *JSPluginInstance) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.instances[instance.ID] = instance
}

// removeInstance 从监控中移除实例
func (m *JSResourceMonitor) removeInstance(pluginID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.instances, pluginID)
}

// collectStats 收集统计信息
func (m *JSResourceMonitor) collectStats() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, instance := range m.instances {
		stats := instance.GetStats()

		// 检查内存限制
		if stats.MemoryUsage > m.memoryLimit {
			// 触发内存限制警告
		}
	}
}

// GetStats 获取插件统计信息
func (instance *JSPluginInstance) GetStats() *JSPluginStats {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	// 更新内存使用情况
	instance.Stats.MemoryUsage = instance.Context.MemoryUsage

	return instance.Stats
}

// Stop 停止JavaScript运行时
func (r *JSRuntime) Stop(ctx context.Context) error {
	close(r.monitor.stopCh)

	// 停止所有插件实例
	for pluginID := range r.instances {
		if err := r.UnloadPlugin(ctx, pluginID); err != nil {
			// 记录错误但继续停止其他实例
		}
	}

	return nil
}
