// Package config 实现插件系统配置管理
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SystemConfig 系统配置
type SystemConfig struct {
	// 插件系统配置
	PluginSystem PluginSystemConfig `json:"plugin_system"`

	// 安全配置
	Security SecurityConfig `json:"security"`

	// 事件系统配置
	EventSystem EventSystemConfig `json:"event_system"`

	// 注册表配置
	Registry RegistryConfig `json:"registry"`

	// 日志配置
	Logging LoggingConfig `json:"logging"`

	// 性能配置
	Performance PerformanceConfig `json:"performance"`
}

// PluginSystemConfig 插件系统配置
type PluginSystemConfig struct {
	// 插件目录
	PluginDirs []string `json:"plugin_dirs"`

	// 启用的加载器
	EnabledLoaders []string `json:"enabled_loaders"`

	// 自动加载
	AutoLoad bool `json:"auto_load"`

	// 热重加载
	HotReload bool `json:"hot_reload"`

	// 最大插件数
	MaxPlugins int `json:"max_plugins"`

	// 插件启动超时
	StartupTimeout time.Duration `json:"startup_timeout"`

	// 插件停止超时
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 启用沙箱
	EnableSandbox bool `json:"enable_sandbox"`

	// 需要签名
	RequireSignature bool `json:"require_signature"`

	// 可信来源
	TrustedSources []string `json:"trusted_sources"`

	// 允许的权限
	AllowedPermissions []string `json:"allowed_permissions"`

	// 禁止的权限
	BlockedPermissions []string `json:"blocked_permissions"`

	// 每个插件最大内存占用
	MaxMemoryPerPlugin uint64 `json:"max_memory_per_plugin"`

	// 每个插件最大CPU时间
	MaxCPUTimePerPlugin time.Duration `json:"max_cpu_time_per_plugin"`

	// 每个插件最大Goroutine数
	MaxGoroutinesPerPlugin int `json:"max_goroutines_per_plugin"`

	// 允许的网络主机
	AllowedNetworkHosts []string `json:"allowed_network_hosts"`

	// 禁止的API
	BlockedAPIs []string `json:"blocked_apis"`
}

// EventSystemConfig 事件系统配置
type EventSystemConfig struct {
	// 工作协程数
	WorkerCount int `json:"worker_count"`

	// 事件队列大小
	QueueSize int `json:"queue_size"`

	// 事件处理超时
	HandlerTimeout time.Duration `json:"handler_timeout"`

	// 启用事件持久化
	EnablePersistence bool `json:"enable_persistence"`

	// 事件存储路径
	StoragePath string `json:"storage_path"`

	// 最大事件历史记录数
	MaxEventHistory int `json:"max_event_history"`
}

// RegistryConfig 注册表配置
type RegistryConfig struct {
	// 自动扫描
	AutoScan bool `json:"auto_scan"`

	// 扫描间隔
	ScanInterval time.Duration `json:"scan_interval"`

	// 插件路径
	PluginPaths []string `json:"plugin_paths"`

	// 启用观察插件目录
	EnableWatcher bool `json:"enable_watcher"`

	// 启用缓存
	CacheEnabled bool `json:"cache_enabled"`

	// 缓存文件
	CacheFile string `json:"cache_file"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	// 日志级别
	Level string `json:"level"`

	// 日志格式
	Format string `json:"format"`

	// 日志输出
	Output []string `json:"output"`

	// 日志文件路径
	FilePath string `json:"file_path"`

	// 最大文件大小
	MaxFileSize int64 `json:"max_file_size"`

	// 最大文件数
	MaxFiles int `json:"max_files"`

	// 启用插件日志隔离
	EnablePluginIsolation bool `json:"enable_plugin_isolation"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	// 启用性能监控
	EnableMonitoring bool `json:"enable_monitoring"`

	// 监控间隔
	MonitoringInterval time.Duration `json:"monitoring_interval"`

	// 启用指标收集
	EnableMetrics bool `json:"enable_metrics"`

	// 指标存储路径
	MetricsStoragePath string `json:"metrics_storage_path"`

	// 启用性能分析
	EnableProfiling bool `json:"enable_profiling"`

	// 分析数据路径
	ProfilingDataPath string `json:"profiling_data_path"`
}

// ConfigManager 配置管理
type ConfigManager struct {
	config     *SystemConfig
	configPath string
	mutex      sync.RWMutex
	watchers   []ConfigWatcher
}

// ConfigWatcher 配置观察器
type ConfigWatcher interface {
	OnConfigChanged(config *SystemConfig)
}

// NewConfigManager 创建配置管理
func NewConfigManager(configPath string) (*ConfigManager, error) {
	manager := &ConfigManager{
		configPath: configPath,
	}

	// 加载配置
	if err := manager.Load(); err != nil {
		// 如果加载失败，使用默认配置
		manager.config = DefaultSystemConfig()
	}

	return manager, nil
}

// DefaultSystemConfig 默认系统配置
func DefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		PluginSystem: PluginSystemConfig{
			PluginDirs:      []string{"./plugins", "/usr/lib/laojun/plugins"},
			EnabledLoaders:  []string{"go", "js"},
			AutoLoad:        true,
			HotReload:       false,
			MaxPlugins:      100,
			StartupTimeout:  30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Security: SecurityConfig{
			EnableSandbox:          true,
			RequireSignature:       false,
			TrustedSources:         []string{},
			AllowedPermissions:     []string{},
			BlockedPermissions:     []string{"os.Exit", "os.Kill", "syscall"},
			MaxMemoryPerPlugin:     128 * 1024 * 1024, // 128MB
			MaxCPUTimePerPlugin:    60 * time.Second,
			MaxGoroutinesPerPlugin: 20,
			AllowedNetworkHosts:    []string{},
			BlockedAPIs:            []string{"os.Exit", "os.Kill", "syscall"},
		},
		EventSystem: EventSystemConfig{
			WorkerCount:       5,
			QueueSize:         1000,
			HandlerTimeout:    30 * time.Second,
			EnablePersistence: false,
			StoragePath:       "./events",
			MaxEventHistory:   10000,
		},
		Registry: RegistryConfig{
			AutoScan:      true,
			ScanInterval:  30 * time.Second,
			PluginPaths:   []string{"./plugins"},
			EnableWatcher: true,
			CacheEnabled:  true,
			CacheFile:     "plugin_registry.json",
		},
		Logging: LoggingConfig{
			Level:                 "info",
			Format:                "json",
			Output:                []string{"stdout", "file"},
			FilePath:              "./logs/plugin_system.log",
			MaxFileSize:           100 * 1024 * 1024, // 100MB
			MaxFiles:              10,
			EnablePluginIsolation: true,
		},
		Performance: PerformanceConfig{
			EnableMonitoring:   true,
			MonitoringInterval: 10 * time.Second,
			EnableMetrics:      true,
			MetricsStoragePath: "./metrics",
			EnableProfiling:    false,
			ProfilingDataPath:  "./profiling",
		},
	}
}

// Load 加载配置
func (cm *ConfigManager) Load() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config SystemConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.config = &config

	// 通知观察器配置已更改
	cm.notifyWatchers()

	return nil
}

// Save 保存配置
func (cm *ConfigManager) Save() error {
	cm.mutex.RLock()
	config := cm.config
	cm.mutex.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Get 获取配置
func (cm *ConfigManager) Get() *SystemConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 返回配置的副本，防止外部修改
	configCopy := *cm.config
	return &configCopy
}

// Update 更新配置
func (cm *ConfigManager) Update(config *SystemConfig) error {
	cm.mutex.Lock()
	cm.config = config
	cm.mutex.Unlock()

	// 保存到文件
	if err := cm.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// 通知观察器配置已更改
	cm.notifyWatchers()

	return nil
}

// GetPluginSystemConfig 获取插件系统配置
func (cm *ConfigManager) GetPluginSystemConfig() PluginSystemConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.config.PluginSystem
}

// GetSecurityConfig 获取安全配置
func (cm *ConfigManager) GetSecurityConfig() SecurityConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.config.Security
}

// GetEventSystemConfig 获取事件系统配置
func (cm *ConfigManager) GetEventSystemConfig() EventSystemConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.config.EventSystem
}

// GetRegistryConfig 获取注册表配置
func (cm *ConfigManager) GetRegistryConfig() RegistryConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.config.Registry
}

// AddWatcher 添加配置观察器
func (cm *ConfigManager) AddWatcher(watcher ConfigWatcher) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.watchers = append(cm.watchers, watcher)
}

// RemoveWatcher 移除配置观察器
func (cm *ConfigManager) RemoveWatcher(watcher ConfigWatcher) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for i, w := range cm.watchers {
		if w == watcher {
			cm.watchers = append(cm.watchers[:i], cm.watchers[i+1:]...)
			break
		}
	}
}

// notifyWatchers 通知观察器配置已更改
func (cm *ConfigManager) notifyWatchers() {
	for _, watcher := range cm.watchers {
		go watcher.OnConfigChanged(cm.config)
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *SystemConfig) error {
	// 验证插件系统配置
	if config.PluginSystem.MaxPlugins <= 0 {
		return fmt.Errorf("max_plugins must be greater than 0")
	}

	if config.PluginSystem.StartupTimeout <= 0 {
		return fmt.Errorf("startup_timeout must be greater than 0")
	}

	// 验证安全配置
	if config.Security.MaxMemoryPerPlugin <= 0 {
		return fmt.Errorf("max_memory_per_plugin must be greater than 0")
	}

	if config.Security.MaxCPUTimePerPlugin <= 0 {
		return fmt.Errorf("max_cpu_time_per_plugin must be greater than 0")
	}

	// 验证事件系统配置
	if config.EventSystem.WorkerCount <= 0 {
		return fmt.Errorf("worker_count must be greater than 0")
	}

	if config.EventSystem.QueueSize <= 0 {
		return fmt.Errorf("queue_size must be greater than 0")
	}

	// 验证注册表配置
	if config.Registry.ScanInterval <= 0 {
		return fmt.Errorf("scan_interval must be greater than 0")
	}

	return nil
}

// MergeConfigs 合并配置
func MergeConfigs(base, override *SystemConfig) *SystemConfig {
	// 这里实现配置合并逻辑
	// 简化实现，实际可能需要更复杂的合并策略
	result := *base

	if override != nil {
		// 合并插件系统配置
		if len(override.PluginSystem.PluginDirs) > 0 {
			result.PluginSystem.PluginDirs = override.PluginSystem.PluginDirs
		}

		// 合并其他配置...
		// 这里可以根据需要实现更详细的合并逻辑
	}

	return &result
}

// PluginConfig 单个插件配置
type PluginConfig struct {
	ID       string                 `json:"id"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"`
}

// PluginConfigManager 插件配置管理
type PluginConfigManager struct {
	configs    map[string]*PluginConfig
	configPath string
	mutex      sync.RWMutex
}

// NewPluginConfigManager 创建插件配置管理
func NewPluginConfigManager(configPath string) *PluginConfigManager {
	manager := &PluginConfigManager{
		configs:    make(map[string]*PluginConfig),
		configPath: configPath,
	}

	// 加载配置
	manager.Load()

	return manager
}

// Load 加载插件配置
func (pcm *PluginConfigManager) Load() error {
	pcm.mutex.Lock()
	defer pcm.mutex.Unlock()

	data, err := os.ReadFile(pcm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 配置文件不存在，使用空配置
		}
		return fmt.Errorf("failed to read plugin config file: %w", err)
	}

	var configs map[string]*PluginConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse plugin config file: %w", err)
	}

	pcm.configs = configs
	return nil
}

// Save 保存插件配置
func (pcm *PluginConfigManager) Save() error {
	pcm.mutex.RLock()
	configs := make(map[string]*PluginConfig)
	for k, v := range pcm.configs {
		configs[k] = v
	}
	pcm.mutex.RUnlock()

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin configs: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(pcm.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create plugin config directory: %w", err)
	}

	return os.WriteFile(pcm.configPath, data, 0644)
}

// GetPluginConfig 获取插件配置
func (pcm *PluginConfigManager) GetPluginConfig(pluginID string) (*PluginConfig, error) {
	pcm.mutex.RLock()
	defer pcm.mutex.RUnlock()

	config, exists := pcm.configs[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin config not found: %s", pluginID)
	}

	// 返回副本
	configCopy := *config
	return &configCopy, nil
}

// SetPluginConfig 设置插件配置
func (pcm *PluginConfigManager) SetPluginConfig(pluginID string, config *PluginConfig) error {
	pcm.mutex.Lock()
	pcm.configs[pluginID] = config
	pcm.mutex.Unlock()

	return pcm.Save()
}

// IsPluginEnabled 检查插件是否启用
func (pcm *PluginConfigManager) IsPluginEnabled(pluginID string) bool {
	pcm.mutex.RLock()
	defer pcm.mutex.RUnlock()

	config, exists := pcm.configs[pluginID]
	if !exists {
		return true // 默认启用
	}

	return config.Enabled
}
