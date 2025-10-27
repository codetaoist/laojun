package config

import (
	"context"
	"time"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// 获取配置
	Get(ctx context.Context, key string) (interface{}, error)
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetBool(ctx context.Context, key string) (bool, error)
	GetFloat64(ctx context.Context, key string) (float64, error)
	GetDuration(ctx context.Context, key string) (time.Duration, error)
	GetStringSlice(ctx context.Context, key string) ([]string, error)
	GetStringMap(ctx context.Context, key string) (map[string]interface{}, error)

	// 设置配置
	Set(ctx context.Context, key string, value interface{}) error
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// 删除配置
	Delete(ctx context.Context, key string) error

	// 检查配置是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// 获取所有配置键
	Keys(ctx context.Context, pattern string) ([]string, error)

	// 监听配置变化
	Watch(ctx context.Context, key string, callback ConfigChangeCallback) error
	Unwatch(ctx context.Context, key string) error

	// 批量操作
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMultiple(ctx context.Context, configs map[string]interface{}) error

	// 配置版本管理
	GetVersion(ctx context.Context, key string) (int64, error)
	GetHistory(ctx context.Context, key string, limit int) ([]ConfigHistory, error)

	// 健康检查
	Health(ctx context.Context) (*ConfigHealth, error)

	// 关闭连接
	Close() error
}

// ConfigClient 配置中心客户端接口
type ConfigClient interface {
	// 连接配置中心
	Connect(ctx context.Context) error

	// 获取配置
	GetConfig(ctx context.Context, service, environment, key string) (*ConfigItem, error)
	GetConfigs(ctx context.Context, service, environment string) (map[string]*ConfigItem, error)

	// 设置配置
	SetConfig(ctx context.Context, item *ConfigItem) error
	UpdateConfig(ctx context.Context, service, environment, key string, value interface{}) error

	// 删除配置
	DeleteConfig(ctx context.Context, service, environment, key string) error

	// 监听配置变化
	Subscribe(ctx context.Context, service, environment string, callback ConfigChangeCallback) error
	Unsubscribe(ctx context.Context, service, environment string) error

	// 配置版本管理
	GetConfigVersion(ctx context.Context, service, environment, key string) (int64, error)
	GetConfigHistory(ctx context.Context, service, environment, key string, limit int) ([]ConfigHistory, error)
	RollbackConfig(ctx context.Context, service, environment, key string, version int64) error

	// 配置发布
	PublishConfig(ctx context.Context, service, environment string, configs map[string]interface{}) error

	// 健康检查
	Ping(ctx context.Context) error

	// 关闭连接
	Close() error
}

// ConfigStorage 配置存储接口
type ConfigStorage interface {
	// 基本操作
	Get(ctx context.Context, key string) (*ConfigItem, error)
	Set(ctx context.Context, item *ConfigItem) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	GetMultiple(ctx context.Context, keys []string) (map[string]*ConfigItem, error)
	SetMultiple(ctx context.Context, items []*ConfigItem) error
	DeleteMultiple(ctx context.Context, keys []string) error

	// 查询操作
	List(ctx context.Context, prefix string) ([]*ConfigItem, error)
	Search(ctx context.Context, query *ConfigQuery) ([]*ConfigItem, error)

	// 版本管理
	GetVersions(ctx context.Context, key string, limit int) ([]ConfigHistory, error)
	SaveVersion(ctx context.Context, item *ConfigItem) error

	// 事务操作
	Transaction(ctx context.Context, fn func(tx ConfigTransaction) error) error

	// 健康检查
	Health(ctx context.Context) error

	// 关闭连接
	Close() error
}

// ConfigTransaction 配置事务接口
type ConfigTransaction interface {
	Get(ctx context.Context, key string) (*ConfigItem, error)
	Set(ctx context.Context, item *ConfigItem) error
	Delete(ctx context.Context, key string) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ConfigWatcher 配置监听器接口
type ConfigWatcher interface {
	// 开始监听
	Start(ctx context.Context) error

	// 停止监听
	Stop() error

	// 添加监听键
	AddKey(key string, callback ConfigChangeCallback) error

	// 移除监听键
	RemoveKey(key string) error

	// 获取监听状态
	IsWatching(key string) bool

	// 获取所有监听的键
	GetWatchedKeys() []string
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	// 验证配置项
	Validate(ctx context.Context, item *ConfigItem) error

	// 验证配置值
	ValidateValue(ctx context.Context, key string, value interface{}) error

	// 添加验证规则
	AddRule(key string, rule ValidationRule) error

	// 移除验证规则
	RemoveRule(key string) error

	// 获取验证规则
	GetRules() map[string]ValidationRule
}

// ConfigItem 配置项
type ConfigItem struct {
	// 基本信息
	Service     string      `json:"service" yaml:"service"`
	Environment string      `json:"environment" yaml:"environment"`
	Key         string      `json:"key" yaml:"key"`
	Value       interface{} `json:"value" yaml:"value"`
	Type        ConfigType  `json:"type" yaml:"type"`

	// 元数据
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// 版本信息
	Version   int64     `json:"version" yaml:"version"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty" yaml:"updated_by,omitempty"`

	// 配置属性
	Encrypted bool          `json:"encrypted,omitempty" yaml:"encrypted,omitempty"`
	TTL       time.Duration `json:"ttl,omitempty" yaml:"ttl,omitempty"`
	ExpiresAt *time.Time    `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
}

// ConfigHistory 配置历史记录
type ConfigHistory struct {
	ID          int64       `json:"id" yaml:"id"`
	Service     string      `json:"service" yaml:"service"`
	Environment string      `json:"environment" yaml:"environment"`
	Key         string      `json:"key" yaml:"key"`
	Value       interface{} `json:"value" yaml:"value"`
	Version     int64       `json:"version" yaml:"version"`
	Operation   string      `json:"operation" yaml:"operation"` // create, update, delete
	CreatedAt   time.Time   `json:"created_at" yaml:"created_at"`
	CreatedBy   string      `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	Comment     string      `json:"comment,omitempty" yaml:"comment,omitempty"`
}

// ConfigQuery 配置查询条件
type ConfigQuery struct {
	Service     string            `json:"service,omitempty" yaml:"service,omitempty"`
	Environment string            `json:"environment,omitempty" yaml:"environment,omitempty"`
	KeyPattern  string            `json:"key_pattern,omitempty" yaml:"key_pattern,omitempty"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Type        ConfigType        `json:"type,omitempty" yaml:"type,omitempty"`
	Limit       int               `json:"limit,omitempty" yaml:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty" yaml:"offset,omitempty"`
	SortBy      string            `json:"sort_by,omitempty" yaml:"sort_by,omitempty"`
	SortOrder   string            `json:"sort_order,omitempty" yaml:"sort_order,omitempty"`
}

// ConfigHealth 配置健康状态
type ConfigHealth struct {
	Status      HealthStatus      `json:"status" yaml:"status"`
	Message     string            `json:"message,omitempty" yaml:"message,omitempty"`
	Timestamp   time.Time         `json:"timestamp" yaml:"timestamp"`
	Details     map[string]string `json:"details,omitempty" yaml:"details,omitempty"`
	Connections int               `json:"connections" yaml:"connections"`
	Latency     time.Duration     `json:"latency" yaml:"latency"`
}

// ConfigChangeEvent 配置变化事件
type ConfigChangeEvent struct {
	Type        EventType   `json:"type" yaml:"type"`
	Service     string      `json:"service" yaml:"service"`
	Environment string      `json:"environment" yaml:"environment"`
	Key         string      `json:"key" yaml:"key"`
	OldValue    interface{} `json:"old_value,omitempty" yaml:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty" yaml:"new_value,omitempty"`
	Version     int64       `json:"version" yaml:"version"`
	Timestamp   time.Time   `json:"timestamp" yaml:"timestamp"`
	Operator    string      `json:"operator,omitempty" yaml:"operator,omitempty"`
}

// ValidationRule 验证规则
type ValidationRule struct {
	Type        ValidationType `json:"type" yaml:"type"`
	Required    bool           `json:"required" yaml:"required"`
	MinLength   int            `json:"min_length,omitempty" yaml:"min_length,omitempty"`
	MaxLength   int            `json:"max_length,omitempty" yaml:"max_length,omitempty"`
	Pattern     string         `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinValue    float64        `json:"min_value,omitempty" yaml:"min_value,omitempty"`
	MaxValue    float64        `json:"max_value,omitempty" yaml:"max_value,omitempty"`
	AllowedValues []interface{} `json:"allowed_values,omitempty" yaml:"allowed_values,omitempty"`
	CustomValidator func(interface{}) error `json:"-" yaml:"-"`
}

// ConfigType 配置类型枚举
type ConfigType string

const (
	ConfigTypeString   ConfigType = "string"
	ConfigTypeInt      ConfigType = "int"
	ConfigTypeBool     ConfigType = "bool"
	ConfigTypeFloat    ConfigType = "float"
	ConfigTypeJSON     ConfigType = "json"
	ConfigTypeYAML     ConfigType = "yaml"
	ConfigTypeSecret   ConfigType = "secret"
	ConfigTypeFile     ConfigType = "file"
	ConfigTypeArray    ConfigType = "array"
	ConfigTypeObject   ConfigType = "object"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// EventType 事件类型枚举
type EventType string

const (
	EventTypeCreate EventType = "create"
	EventTypeUpdate EventType = "update"
	EventTypeDelete EventType = "delete"
)

// ValidationType 验证类型枚举
type ValidationType string

const (
	ValidationTypeString ValidationType = "string"
	ValidationTypeNumber ValidationType = "number"
	ValidationTypeBool   ValidationType = "bool"
	ValidationTypeEmail  ValidationType = "email"
	ValidationTypeURL    ValidationType = "url"
	ValidationTypeRegex  ValidationType = "regex"
	ValidationTypeCustom ValidationType = "custom"
)

// ConfigChangeCallback 配置变化回调函数
type ConfigChangeCallback func(event *ConfigChangeEvent) error

// ConfigOptions 配置选项
type ConfigOptions struct {
	// 连接配置
	Endpoints   []string      `json:"endpoints" yaml:"endpoints"`
	Username    string        `json:"username,omitempty" yaml:"username,omitempty"`
	Password    string        `json:"password,omitempty" yaml:"password,omitempty"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	RetryCount  int           `json:"retry_count" yaml:"retry_count"`
	RetryDelay  time.Duration `json:"retry_delay" yaml:"retry_delay"`

	// TLS配置
	TLS *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`

	// 缓存配置
	CacheEnabled bool          `json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL     time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
	CacheSize    int           `json:"cache_size" yaml:"cache_size"`

	// 监听配置
	WatchEnabled   bool          `json:"watch_enabled" yaml:"watch_enabled"`
	WatchInterval  time.Duration `json:"watch_interval" yaml:"watch_interval"`
	WatchBufferSize int          `json:"watch_buffer_size" yaml:"watch_buffer_size"`

	// 服务信息
	ServiceName string `json:"service_name" yaml:"service_name"`
	Environment string `json:"environment" yaml:"environment"`

	// 其他选项
	Namespace string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	CertFile           string `json:"cert_file,omitempty" yaml:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty" yaml:"key_file,omitempty"`
	CAFile             string `json:"ca_file,omitempty" yaml:"ca_file,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	ServerName         string `json:"server_name,omitempty" yaml:"server_name,omitempty"`
}