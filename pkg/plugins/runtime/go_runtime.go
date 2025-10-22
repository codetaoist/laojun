package runtime

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
	"sync"
	"time"

	"github.com/taishanglaojun/plugins/core"
)

// GoPluginRuntime Go插件运行时环境
type GoPluginRuntime struct {
	plugins     map[string]*GoPluginInstance
	mutex       sync.RWMutex
	config      *GoRuntimeConfig
	monitor     *ResourceMonitor
	securityMgr core.SecurityManager
}

// GoPluginInstance Go插件实例
type GoPluginInstance struct {
	ID         string
	Plugin     *plugin.Plugin
	Instance   core.Plugin
	Metadata   *core.PluginMetadata
	Status     core.PluginStatus
	LoadTime   time.Time
	LastAccess time.Time
	Config     map[string]any
	Sandbox    core.Sandbox
	Stats      *PluginStats
	mutex      sync.RWMutex
}

// GoRuntimeConfig Go运行时配置
type GoRuntimeConfig struct {
	PluginDir       string        `json:"plugin_dir"`
	MaxPlugins      int           `json:"max_plugins"`
	LoadTimeout     time.Duration `json:"load_timeout"`
	EnableHotReload bool          `json:"enable_hot_reload"`
	WatchInterval   time.Duration `json:"watch_interval"`
	EnableStats     bool          `json:"enable_stats"`
	StatsInterval   time.Duration `json:"stats_interval"`
}

// PluginStats 插件统计信息
type PluginStats struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastError       string        `json:"last_error"`
	LastErrorTime   time.Time     `json:"last_error_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	GoroutineCount  int           `json:"goroutine_count"`
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	enabled        bool
	interval       time.Duration
	stopCh         chan struct{}
	plugins        map[string]*GoPluginInstance
	mutex          sync.RWMutex
	memoryLimit    int64
	goroutineLimit int
}

// NewGoPluginRuntime 创建Go插件运行时环境
func NewGoPluginRuntime(config *GoRuntimeConfig, securityMgr core.SecurityManager) *GoPluginRuntime {
	runtime := &GoPluginRuntime{
		plugins:     make(map[string]*GoPluginInstance),
		config:      config,
		securityMgr: securityMgr,
		monitor: &ResourceMonitor{
			enabled:        config.EnableStats,
			interval:       config.StatsInterval,
			stopCh:         make(chan struct{}),
			plugins:        make(map[string]*GoPluginInstance),
			memoryLimit:    100 * 1024 * 1024, // 100MB
			goroutineLimit: 1000,
		},
	}

	if config.EnableStats {
		go runtime.monitor.start()
	}

	if config.EnableHotReload {
		go runtime.watchPluginDir()
	}

	return runtime
}

// LoadPlugin 加载插件
func (r *GoPluginRuntime) LoadPlugin(ctx context.Context, pluginPath string, config map[string]any) (*GoPluginInstance, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查插件数量限制
	if len(r.plugins) >= r.config.MaxPlugins {
		return nil, fmt.Errorf("maximum plugin limit reached: %d", r.config.MaxPlugins)
	}

	// 加载插件文件
	loadCtx, cancel := context.WithTimeout(ctx, r.config.LoadTimeout)
	defer cancel()

	pluginFile, err := r.loadPluginFile(loadCtx, pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin file: %w", err)
	}

	// 查找插件符号
	pluginSymbol, err := pluginFile.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("plugin symbol not found: %w", err)
	}

	// 类型断言
	pluginInstance, ok := pluginSymbol.(core.Plugin)
	if !ok {
		return nil, fmt.Errorf("invalid plugin type")
	}

	// 获取插件元数据
	metadata := pluginInstance.GetMetadata()
	if metadata == nil {
		return nil, fmt.Errorf("plugin metadata is nil")
	}

	// 创建沙箱
	sandbox, err := r.securityMgr.CreateSandbox(ctx, metadata.ID, &core.SecurityPolicy{
		AllowedAPIs:    []string{"http", "json"},
		MemoryLimit:    r.monitor.memoryLimit,
		GoroutineLimit: r.monitor.goroutineLimit,
		TimeLimit:      30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	// 创建插件实例
	instance := &GoPluginInstance{
		ID:         metadata.ID,
		Plugin:     pluginFile,
		Instance:   pluginInstance,
		Metadata:   metadata,
		Status:     core.StatusLoaded,
		LoadTime:   time.Now(),
		LastAccess: time.Now(),
		Config:     config,
		Sandbox:    sandbox,
		Stats:      &PluginStats{},
	}

	// 初始化插件
	if err := instance.Instance.Initialize(ctx, config); err != nil {
		sandbox.Destroy(ctx)
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// 注册插件
	r.plugins[metadata.ID] = instance
	r.monitor.addPlugin(instance)

	return instance, nil
}

// UnloadPlugin 卸载插件
func (r *GoPluginRuntime) UnloadPlugin(ctx context.Context, pluginID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 停止插件
	if instance.Status == core.StatusRunning {
		if err := instance.Instance.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop plugin: %w", err)
		}
	}

	// 销毁沙箱
	if err := instance.Sandbox.Destroy(ctx); err != nil {
		return fmt.Errorf("failed to destroy sandbox: %w", err)
	}

	// 移除插件
	delete(r.plugins, pluginID)
	r.monitor.removePlugin(pluginID)

	instance.Status = core.StatusUnloaded
	return nil
}

// GetPlugin 获取插件实例
func (r *GoPluginRuntime) GetPlugin(pluginID string) (*GoPluginInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	instance.mutex.Lock()
	instance.LastAccess = time.Now()
	instance.mutex.Unlock()

	return instance, nil
}

// ListPlugins 列出所有插件实例
func (r *GoPluginRuntime) ListPlugins() []*GoPluginInstance {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugins := make([]*GoPluginInstance, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// ExecutePlugin 执行插件请求
func (r *GoPluginRuntime) ExecutePlugin(ctx context.Context, pluginID string, req *core.PluginRequest) (*core.PluginResponse, error) {
	instance, err := r.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	// 记录开始时间
	startTime := time.Now()

	// 在沙箱中执行
	var response *core.PluginResponse
	err = instance.Sandbox.Execute(ctx, func() error {
		var execErr error
		response, execErr = instance.Instance.HandleRequest(ctx, req)
		return execErr
	})

	// 更新统计信息
	duration := time.Since(startTime)
	instance.updateStats(duration, err)

	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	return response, nil
}

// loadPluginFile 加载插件文件
func (r *GoPluginRuntime) loadPluginFile(ctx context.Context, pluginPath string) (*plugin.Plugin, error) {
	// 检查文件是否存在
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// 加载插件
	pluginFile, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	return pluginFile, nil
}

// watchPluginDir 监控插件目录变化
func (r *GoPluginRuntime) watchPluginDir() {
	ticker := time.NewTicker(r.config.WatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.scanPluginDir()
		case <-r.monitor.stopCh:
			return
		}
	}
}

// scanPluginDir 扫描插件目录
func (r *GoPluginRuntime) scanPluginDir() {
	if r.config.PluginDir == "" {
		return
	}

	err := filepath.Walk(r.config.PluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".so" {
			// 检查是否为新插件或已更新的插件
			// 这里可以实现热重载逻辑
		}

		return nil
	})

	if err != nil {
		// 记录错误日志
	}
}

// updateStats 更新插件统计信息
func (instance *GoPluginInstance) updateStats(duration time.Duration, err error) {
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
}

// Start 启动插件
func (instance *GoPluginInstance) Start(ctx context.Context) error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	if instance.Status == core.StatusRunning {
		return nil
	}

	if err := instance.Instance.Start(ctx); err != nil {
		return err
	}

	instance.Status = core.StatusRunning
	return nil
}

// Stop 停止插件
func (instance *GoPluginInstance) Stop(ctx context.Context) error {
	instance.mutex.Lock()
	defer instance.mutex.Unlock()

	if instance.Status != core.StatusRunning {
		return nil
	}

	if err := instance.Instance.Stop(ctx); err != nil {
		return err
	}

	instance.Status = core.StatusStopped
	return nil
}

// GetStats 获取插件统计信息
func (instance *GoPluginInstance) GetStats() *PluginStats {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()

	// 更新内存和协程统计信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	instance.Stats.MemoryUsage = int64(m.Alloc)
	instance.Stats.GoroutineCount = runtime.NumGoroutine()

	return instance.Stats
}

// ResourceMonitor 方法

// start 启动资源监控
func (m *ResourceMonitor) start() {
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

// addPlugin 添加插件到监控
func (m *ResourceMonitor) addPlugin(instance *GoPluginInstance) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.plugins[instance.ID] = instance
}

// removePlugin 从监控中移除插件
func (m *ResourceMonitor) removePlugin(pluginID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.plugins, pluginID)
}

// collectStats 收集统计信息
func (m *ResourceMonitor) collectStats() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, instance := range m.plugins {
		stats := instance.GetStats()

		// 检查资源限制
		if stats.MemoryUsage > m.memoryLimit {
			// 触发内存限制警告
		}

		if stats.GoroutineCount > m.goroutineLimit {
			// 触发协程限制警告
		}
	}
}

// Stop 停止运行时
func (r *GoPluginRuntime) Stop(ctx context.Context) error {
	close(r.monitor.stopCh)

	// 停止所有插件实例
	for pluginID := range r.plugins {
		if err := r.UnloadPlugin(ctx, pluginID); err != nil {
			// 记录错误但继续停止其他插件实例
			log.Printf("Error unloading plugin %s: %v", pluginID, err)
		}
	}

	return nil
}
