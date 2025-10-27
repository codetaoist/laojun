package sdk

import (
	"context"
	"time"
)

// PluginCapability 插件能力枚举
type PluginCapability string

const (
	CapabilityHTTP      PluginCapability = "http"      // HTTP服务能力
	CapabilityEvent     PluginCapability = "event"     // 事件处理能力
	CapabilityScheduled PluginCapability = "scheduled" // 定时任务能力
	CapabilityData      PluginCapability = "data"      // 数据处理能力
	CapabilityConfig    PluginCapability = "config"    // 配置管理能力
	CapabilityStorage   PluginCapability = "storage"   // 存储能力
	CapabilityAuth      PluginCapability = "auth"      // 认证能力
	CapabilityUI        PluginCapability = "ui"        // UI组件能力
)

// PluginState 插件状态枚举
type PluginState string

const (
	StateUnloaded    PluginState = "unloaded"    // 未加载
	StateLoaded      PluginState = "loaded"      // 已加载
	StateInitialized PluginState = "initialized" // 已初始化
	StateStarted     PluginState = "started"     // 已启动
	StateStopped     PluginState = "stopped"     // 已停止
	StateError       PluginState = "error"       // 错误状态
)

// PluginPriority 插件优先级
type PluginPriority int

const (
	PriorityLow    PluginPriority = 1
	PriorityNormal PluginPriority = 5
	PriorityHigh   PluginPriority = 10
)

// PluginManifest 插件清单
type PluginManifest struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Version      string                 `json:"version" yaml:"version"`
	Description  string                 `json:"description" yaml:"description"`
	Author       string                 `json:"author" yaml:"author"`
	Email        string                 `json:"email,omitempty" yaml:"email,omitempty"`
	Homepage     string                 `json:"homepage,omitempty" yaml:"homepage,omitempty"`
	Repository   string                 `json:"repository,omitempty" yaml:"repository,omitempty"`
	License      string                 `json:"license,omitempty" yaml:"license,omitempty"`
	Category     string                 `json:"category" yaml:"category"`
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Keywords     []string               `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	Icon         string                 `json:"icon,omitempty" yaml:"icon,omitempty"`
	Screenshots  []string               `json:"screenshots,omitempty" yaml:"screenshots,omitempty"`
	
	// 技术规格
	Capabilities []PluginCapability     `json:"capabilities" yaml:"capabilities"`
	Priority     PluginPriority         `json:"priority,omitempty" yaml:"priority,omitempty"`
	Permissions  []string               `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	Dependencies []PluginDependency     `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	
	// 运行时配置
	Runtime      RuntimeConfig          `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	
	// 元数据
	CreatedAt    time.Time              `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt    time.Time              `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	MinVersion   string                 `json:"min_version,omitempty" yaml:"min_version,omitempty"`
	MaxVersion   string                 `json:"max_version,omitempty" yaml:"max_version,omitempty"`
}

// PluginDependency 插件依赖
type PluginDependency struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Type    string `json:"type,omitempty" yaml:"type,omitempty"` // required, optional
}

// RuntimeConfig 运行时配置
type RuntimeConfig struct {
	MaxMemoryMB    int           `json:"max_memory_mb,omitempty" yaml:"max_memory_mb,omitempty"`
	MaxCPUPercent  float64       `json:"max_cpu_percent,omitempty" yaml:"max_cpu_percent,omitempty"`
	MaxGoroutines  int           `json:"max_goroutines,omitempty" yaml:"max_goroutines,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RestartPolicy  string        `json:"restart_policy,omitempty" yaml:"restart_policy,omitempty"`
	HealthCheck    HealthConfig  `json:"health_check,omitempty" yaml:"health_check,omitempty"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retries  int           `json:"retries,omitempty" yaml:"retries,omitempty"`
}

// PluginRegistry 插件注册中心接口
type PluginRegistry interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(ctx context.Context, manifest *PluginManifest) error
	
	// UnregisterPlugin 注销插件
	UnregisterPlugin(ctx context.Context, pluginID string) error
	
	// GetPlugin 获取插件信息
	GetPlugin(ctx context.Context, pluginID string) (*PluginManifest, error)
	
	// ListPlugins 列出插件
	ListPlugins(ctx context.Context, filter *PluginFilter) ([]*PluginManifest, error)
	
	// SearchPlugins 搜索插件
	SearchPlugins(ctx context.Context, query string) ([]*PluginManifest, error)
	
	// UpdatePlugin 更新插件信息
	UpdatePlugin(ctx context.Context, manifest *PluginManifest) error
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	Category     string             `json:"category,omitempty"`
	Tags         []string           `json:"tags,omitempty"`
	Capabilities []PluginCapability `json:"capabilities,omitempty"`
	Author       string             `json:"author,omitempty"`
	State        PluginState        `json:"state,omitempty"`
	Limit        int                `json:"limit,omitempty"`
	Offset       int                `json:"offset,omitempty"`
}

// EventBusClient 事件总线客户端接口
type EventBusClient interface {
	// Subscribe 订阅事件
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error
	
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, eventType string) error
	
	// Publish 发布事件
	Publish(ctx context.Context, event *Event) error
	
	// PublishAsync 异步发布事件
	PublishAsync(ctx context.Context, event *Event) error
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// RegistryClient 注册中心客户端接口
type RegistryClient interface {
	// Register 注册服务
	Register(ctx context.Context, service *ServiceInfo) error
	
	// Deregister 注销服务
	Deregister(ctx context.Context, serviceID string) error
	
	// Discover 发现服务
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	
	// Watch 监听服务变化
	Watch(ctx context.Context, serviceName string, callback ServiceChangeCallback) error
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ServiceChangeCallback 服务变化回调
type ServiceChangeCallback func(services []*ServiceInfo)

// StorageClient 存储客户端接口
type StorageClient interface {
	// Get 获取数据
	Get(ctx context.Context, key string) ([]byte, error)
	
	// Set 设置数据
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	
	// Delete 删除数据
	Delete(ctx context.Context, key string) error
	
	// List 列出键
	List(ctx context.Context, prefix string) ([]string, error)
	
	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)
}

// ConfigClient 配置客户端接口
type ConfigClient interface {
	// GetConfig 获取配置
	GetConfig(ctx context.Context, key string) (interface{}, error)
	
	// SetConfig 设置配置
	SetConfig(ctx context.Context, key string, value interface{}) error
	
	// WatchConfig 监听配置变化
	WatchConfig(ctx context.Context, key string, callback ConfigChangeCallback) error
	
	// GetAllConfig 获取所有配置
	GetAllConfig(ctx context.Context) (map[string]interface{}, error)
}

// ConfigChangeCallback 配置变化回调
type ConfigChangeCallback func(key string, oldValue, newValue interface{})

// LoggerClient 日志客户端接口
type LoggerClient interface {
	// Debug 调试日志
	Debug(msg string, fields ...map[string]interface{})
	
	// Info 信息日志
	Info(msg string, fields ...map[string]interface{})
	
	// Warn 警告日志
	Warn(msg string, fields ...map[string]interface{})
	
	// Error 错误日志
	Error(msg string, err error, fields ...map[string]interface{})
	
	// WithField 添加字段
	WithField(key string, value interface{}) LoggerClient
	
	// WithFields 添加多个字段
	WithFields(fields map[string]interface{}) LoggerClient
}

// MetricsClient 指标客户端接口
type MetricsClient interface {
	// Counter 计数器
	Counter(name string, tags map[string]string) CounterMetric
	
	// Gauge 仪表盘
	Gauge(name string, tags map[string]string) GaugeMetric
	
	// Histogram 直方图
	Histogram(name string, tags map[string]string) HistogramMetric
	
	// Timer 计时器
	Timer(name string, tags map[string]string) TimerMetric
}

// CounterMetric 计数器指标
type CounterMetric interface {
	Inc()
	Add(value float64)
}

// GaugeMetric 仪表盘指标
type GaugeMetric interface {
	Set(value float64)
	Inc()
	Dec()
	Add(value float64)
	Sub(value float64)
}

// HistogramMetric 直方图指标
type HistogramMetric interface {
	Observe(value float64)
}

// TimerMetric 计时器指标
type TimerMetric interface {
	Record(duration time.Duration)
	Start() func()
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	PluginID       string        `json:"plugin_id"`
	MemoryBytes    uint64        `json:"memory_bytes"`
	CPUPercent     float64       `json:"cpu_percent"`
	GoroutineCount int           `json:"goroutine_count"`
	FileHandles    int           `json:"file_handles"`
	NetworkConns   int           `json:"network_conns"`
	DiskUsageBytes uint64        `json:"disk_usage_bytes"`
	Timestamp      time.Time     `json:"timestamp"`
}

// PluginError 插件错误
type PluginError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Cause   error  `json:"-"`
}

func (e *PluginError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// 常见错误代码
const (
	ErrCodePluginNotFound     = "PLUGIN_NOT_FOUND"
	ErrCodePluginAlreadyExists = "PLUGIN_ALREADY_EXISTS"
	ErrCodeInvalidManifest    = "INVALID_MANIFEST"
	ErrCodePermissionDenied   = "PERMISSION_DENIED"
	ErrCodeResourceExceeded   = "RESOURCE_EXCEEDED"
	ErrCodeDependencyMissing  = "DEPENDENCY_MISSING"
	ErrCodeInitializationFailed = "INITIALIZATION_FAILED"
	ErrCodeExecutionTimeout   = "EXECUTION_TIMEOUT"
)