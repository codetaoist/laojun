package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EngineStats 引擎统计信息
type EngineStats struct {
	Running       bool          `json:"running"`
	TotalPlugins  int           `json:"total_plugins"`
	StartTime     time.Time     `json:"start_time"`
	RegistryStats RegistryStats `json:"registry_stats"`
}

// PluginEngine 插件运行时引擎
type PluginEngine struct {
	config           *EngineConfig
	manager          PluginManager
	loader           PluginLoader
	sandbox          Sandbox
	registry         PluginRegistry
	lifecycleManager LifecycleManager
	monitor          Monitor
	dependencyManager DependencyManager
	executor         Executor
	logger           *logrus.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	running          bool
	mu               sync.RWMutex
}

// EngineConfig 引擎配置
type EngineConfig struct {
	PluginDir           string        `yaml:"plugin_dir"`
	MaxPlugins          int           `yaml:"max_plugins"`
	LoadTimeout         time.Duration `yaml:"load_timeout"`
	StartTimeout        time.Duration `yaml:"start_timeout"`
	StopTimeout         time.Duration `yaml:"stop_timeout"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	ResourceMonitoring  bool          `yaml:"resource_monitoring"`
	EventBufferSize     int           `yaml:"event_buffer_size"`
	EnableSandbox       bool          `yaml:"enable_sandbox"`
	LogLevel            string        `yaml:"log_level"`
}

// DefaultEngineConfig 默认引擎配置
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		PluginDir:           "./plugins",
		MaxPlugins:          100,
		LoadTimeout:         30 * time.Second,
		StartTimeout:        10 * time.Second,
		StopTimeout:         10 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		ResourceMonitoring:  true,
		EventBufferSize:     1000,
		EnableSandbox:       true,
		LogLevel:            "info",
	}
}

// NewPluginEngine 创建插件运行时引擎
func NewPluginEngine(config *EngineConfig, logger *logrus.Logger) (*PluginEngine, error) {
	if config == nil {
		config = DefaultEngineConfig()
	}

	if logger == nil {
		logger = logrus.New()
	}

	// 创建组件
	loader := NewDefaultPluginLoader(logger)
	sandbox := NewDefaultSandbox(logger)
	manager := NewDefaultPluginManager(loader, sandbox, logger)
	registry := NewDefaultPluginRegistry(logger)
	lifecycleManager := NewDefaultLifecycleManager(manager, registry, logger)
	monitor := NewDefaultMonitor(manager, config.HealthCheckInterval, logger)
	dependencyManager := NewDefaultDependencyManager(logger)
	executor := NewDefaultExecutor(manager, 10, 1000, logger)

	engine := &PluginEngine{
		config:            config,
		manager:           manager,
		loader:            loader,
		sandbox:           sandbox,
		registry:          registry,
		lifecycleManager:  lifecycleManager,
		monitor:           monitor,
		dependencyManager: dependencyManager,
		executor:          executor,
		logger:            logger,
		running:           false,
	}

	logger.Info("Plugin engine created successfully")
	return engine, nil
}

// Start 启动插件引擎
func (e *PluginEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("plugin engine already started")
	}

	e.logger.Info("Starting plugin engine")

	// 设置上下文
	e.ctx, e.cancel = context.WithCancel(ctx)
	e.running = true

	// 启动组件
	if err := e.manager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin manager: %w", err)
	}

	// 自动加载插件
	if err := e.autoLoadPlugins(); err != nil {
		e.logger.Warnf("Failed to auto-load plugins: %v", err)
	}

	// 启动健康检查
	if e.config.HealthCheckInterval > 0 {
		go e.startHealthCheck()
	}

	e.logger.Info("Plugin engine started successfully")
	return nil
}

// Stop 停止插件引擎
func (e *PluginEngine) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	e.logger.Info("Stopping plugin engine")

	// 停止所有插件
	plugins := e.manager.ListPlugins()
	for pluginID := range plugins {
		if err := e.manager.StopPlugin(ctx, pluginID); err != nil {
			e.logger.WithError(err).WithField("plugin_id", pluginID).Error("Failed to stop plugin")
		}
		if err := e.manager.UnloadPlugin(ctx, pluginID); err != nil {
			e.logger.WithError(err).WithField("plugin_id", pluginID).Error("Failed to unload plugin")
		}
	}

	// 停止监控
	if err := e.monitor.Stop(ctx); err != nil {
		e.logger.WithError(err).Error("Failed to stop monitor")
	}

	// 停止执行器
	if err := e.executor.Stop(ctx); err != nil {
		e.logger.WithError(err).Error("Failed to stop executor")
	}

	// 停止管理器
	if err := e.manager.Stop(ctx); err != nil {
		e.logger.WithError(err).Error("Failed to stop manager")
	}

	// 取消上下文
	if e.cancel != nil {
		e.cancel()
	}

	e.running = false
	e.logger.Info("Plugin engine stopped")

	return nil
}

// LoadPlugin 加载插件
func (e *PluginEngine) LoadPlugin(pluginPath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.WithField("path", pluginPath).Info("Loading plugin")

	ctx, cancel := context.WithTimeout(e.ctx, e.config.LoadTimeout)
	defer cancel()

	// 使用加载器加载插件
	plugin, err := e.loader.LoadPlugin(ctx, pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	metadata := plugin.GetMetadata()

	// 注册插件到注册中心
	if err := e.registry.RegisterPlugin(&metadata); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	// 加载插件到管理器
	if err := e.manager.LoadPlugin(ctx, pluginPath); err != nil {
		// 如果加载失败，从注册中心移除
		e.registry.UnregisterPlugin(metadata.ID)
		return fmt.Errorf("failed to load plugin in manager: %w", err)
	}

	e.logger.WithField("plugin_id", metadata.ID).Info("Plugin loaded successfully")
	return nil
}

// StartPlugin 启动插件
func (e *PluginEngine) StartPlugin(ctx context.Context, pluginID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.WithField("plugin_id", pluginID).Info("Starting plugin")

	// 启动插件
	if err := e.manager.StartPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	e.logger.WithField("plugin_id", pluginID).Info("Plugin started successfully")
	return nil
}

// StopPlugin 停止插件
func (e *PluginEngine) StopPlugin(pluginID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.running {
		return fmt.Errorf("plugin engine is not running")
	}

	// 使用生命周期管理器停止插件
	ctx := context.Background()
	if err := e.lifecycleManager.StopPlugin(ctx, pluginID); err != nil {
		e.logger.WithError(err).WithField("plugin_id", pluginID).Error("Failed to stop plugin")
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	e.logger.WithField("plugin_id", pluginID).Info("Plugin stopped successfully")
	return nil
}

// UnloadPlugin 卸载插件
func (e *PluginEngine) UnloadPlugin(pluginID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.WithField("plugin_id", pluginID).Info("Unloading plugin")

	ctx, cancel := context.WithTimeout(e.ctx, e.config.StopTimeout)
	defer cancel()

	// 从管理器卸载插件
	if err := e.manager.UnloadPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// 从注册中心注销插件
	if err := e.registry.UnregisterPlugin(pluginID); err != nil {
		e.logger.WithError(err).WithField("plugin_id", pluginID).Warn("Failed to unregister plugin")
	}

	e.logger.WithField("plugin_id", pluginID).Info("Plugin unloaded successfully")
	return nil
}

// GetPluginInfo 获取插件信息
func (e *PluginEngine) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	// 从注册中心获取插件元数据
	metadata, err := e.registry.GetPlugin(pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin from registry: %w", err)
	}

	// 从生命周期管理器获取状态
	lifecycleState, err := e.lifecycleManager.GetPluginLifecycleState(pluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin lifecycle state: %w", err)
	}

	// 构建插件信息
	info := &PluginInfo{
		Metadata: *metadata,
		State:    lifecycleState.CurrentState,
		LoadedAt: lifecycleState.InitializedAt,
	}

	return info, nil
}

// ListPlugins 列出所有插件
func (e *PluginEngine) ListPlugins() map[string]*PluginInfo {
	plugins, err := e.registry.ListPlugins()
	if err != nil {
		e.logger.WithError(err).Error("Failed to list plugins from registry")
		return make(map[string]*PluginInfo)
	}

	result := make(map[string]*PluginInfo)
	for _, metadata := range plugins {
		// 获取生命周期状态
		lifecycleState, err := e.lifecycleManager.GetPluginLifecycleState(metadata.ID)
		if err != nil {
			e.logger.WithError(err).WithField("plugin_id", metadata.ID).Warn("Failed to get plugin lifecycle state")
			continue
		}

		info := &PluginInfo{
			Metadata: *metadata,
			State:    lifecycleState.CurrentState,
			LoadedAt: lifecycleState.InitializedAt,
		}

		result[metadata.ID] = info
	}

	return result
}

// GetEngineStatus 获取引擎状态
func (e *PluginEngine) GetEngineStatus() *EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	plugins := e.ListPlugins()
	
	// 获取注册中心统计信息
	registryStats := e.registry.GetStats()
	
	status := &EngineStatus{
		Running:     e.running,
		PluginCount: len(plugins),
		MaxPlugins:  e.config.MaxPlugins,
		Config:      e.config,
		Plugins:     plugins,
		RegistryStats: registryStats,
	}

	return status
}

// autoLoadPlugins 自动加载插件目录中的插件
func (e *PluginEngine) autoLoadPlugins() error {
	// 这里可以实现自动扫描插件目录并加载插件的逻辑
	e.logger.WithField("plugin_dir", e.config.PluginDir).Info("Auto-loading plugins from directory")
	
	// TODO: 实现目录扫描和插件自动加载
	
	return nil
}

// startHealthCheck 启动健康检查
func (e *PluginEngine) startHealthCheck() {
	ticker := time.NewTicker(e.config.HealthCheckInterval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				e.performHealthCheck()
			case <-e.ctx.Done():
				return
			}
		}
	}()
}

// performHealthCheck 执行健康检查
func (e *PluginEngine) performHealthCheck() {
	plugins := e.manager.ListPlugins()
	
	for pluginID, info := range plugins {
		if info.State == StateRunning {
			// 检查插件健康状态
			if usage, err := e.manager.GetResourceUsage(pluginID); err == nil {
				// 检查资源使用情况
				if usage.MemoryBytes > 1024*1024*100 { // 100MB
					e.logger.WithFields(logrus.Fields{
						"plugin_id":    pluginID,
						"memory_usage": usage.MemoryBytes,
					}).Warn("Plugin memory usage is high")
				}
			}
		}
	}
}

// EngineStatus 引擎状态
type EngineStatus struct {
	Running       bool                     `json:"running"`
	PluginCount   int                      `json:"plugin_count"`
	MaxPlugins    int                      `json:"max_plugins"`
	Config        *EngineConfig            `json:"config"`
	Plugins       map[string]*PluginInfo   `json:"plugins"`
	RegistryStats *RegistryStats           `json:"registry_stats"`
}