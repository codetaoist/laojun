package health

import (
	"context"
	"time"
)

// HealthChecker 统一的健康检查器接口
type HealthChecker interface {
	// Check 执行健康检查
	Check(ctx context.Context) CheckResult
	// Name 返回检查器名称
	Name() string
	// Type 返回检查器类型
	Type() CheckerType
	// Priority 返回检查器优先级
	Priority() Priority
}

// HealthManager 健康检查管理器接口
type HealthManager interface {
	// AddChecker 添加检查器
	AddChecker(checker HealthChecker) error
	// RemoveChecker 移除检查器
	RemoveChecker(name string) error
	// Check 执行所有健康检查
	Check(ctx context.Context) HealthReport
	// CheckByType 按类型执行健康检查
	CheckByType(ctx context.Context, checkerType CheckerType) HealthReport
	// CheckByPriority 按优先级执行健康检查
	CheckByPriority(ctx context.Context, priority Priority) HealthReport
	// GetChecker 获取指定检查器
	GetChecker(name string) (HealthChecker, bool)
	// ListCheckers 列出所有检查器
	ListCheckers() []HealthChecker
}

// CheckerType 检查器类型
type CheckerType string

const (
	CheckerTypeDatabase    CheckerType = "database"
	CheckerTypeCache       CheckerType = "cache"
	CheckerTypeHTTP        CheckerType = "http"
	CheckerTypeGRPC        CheckerType = "grpc"
	CheckerTypeMessage     CheckerType = "message"
	CheckerTypeStorage     CheckerType = "storage"
	CheckerTypeExternal    CheckerType = "external"
	CheckerTypeCustom      CheckerType = "custom"
	CheckerTypeSystem      CheckerType = "system"
	CheckerTypeApplication CheckerType = "application"
)

// Priority 检查器优先级
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
	PriorityCritical Priority = 4
)

// CheckerConfig 检查器配置
type CheckerConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Type        CheckerType       `yaml:"type" json:"type"`
	Priority    Priority          `yaml:"priority" json:"priority"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Timeout     time.Duration     `yaml:"timeout" json:"timeout"`
	Interval    time.Duration     `yaml:"interval" json:"interval"`
	Retries     int               `yaml:"retries" json:"retries"`
	Metadata    map[string]string `yaml:"metadata" json:"metadata"`
	Tags        []string          `yaml:"tags" json:"tags"`
	Description string            `yaml:"description" json:"description"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled         bool                     `yaml:"enabled" json:"enabled"`
	Path            string                   `yaml:"path" json:"path"`
	Timeout         time.Duration            `yaml:"timeout" json:"timeout"`
	Service         ServiceConfig            `yaml:"service" json:"service"`
	Checkers        []CheckerConfig          `yaml:"checkers" json:"checkers"`
	Notifications   NotificationConfig       `yaml:"notifications" json:"notifications"`
	Metrics         MetricsConfig            `yaml:"metrics" json:"metrics"`
	Cache           CacheConfig              `yaml:"cache" json:"cache"`
	Thresholds      ThresholdConfig          `yaml:"thresholds" json:"thresholds"`
	ResponseFormats []ResponseFormatConfig   `yaml:"response_formats" json:"response_formats"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
	Region      string `yaml:"region" json:"region"`
	Zone        string `yaml:"zone" json:"zone"`
	Instance    string `yaml:"instance" json:"instance"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Enabled   bool                   `yaml:"enabled" json:"enabled"`
	Channels  []NotificationChannel  `yaml:"channels" json:"channels"`
	Rules     []NotificationRule     `yaml:"rules" json:"rules"`
	Templates map[string]string      `yaml:"templates" json:"templates"`
}

// NotificationChannel 通知渠道
type NotificationChannel struct {
	Name    string            `yaml:"name" json:"name"`
	Type    string            `yaml:"type" json:"type"`
	Config  map[string]string `yaml:"config" json:"config"`
	Enabled bool              `yaml:"enabled" json:"enabled"`
}

// NotificationRule 通知规则
type NotificationRule struct {
	Name      string      `yaml:"name" json:"name"`
	Condition string      `yaml:"condition" json:"condition"`
	Channels  []string    `yaml:"channels" json:"channels"`
	Severity  string      `yaml:"severity" json:"severity"`
	Cooldown  time.Duration `yaml:"cooldown" json:"cooldown"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Namespace  string            `yaml:"namespace" json:"namespace"`
	Subsystem  string            `yaml:"subsystem" json:"subsystem"`
	Labels     map[string]string `yaml:"labels" json:"labels"`
	Collectors []string          `yaml:"collectors" json:"collectors"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled bool          `yaml:"enabled" json:"enabled"`
	TTL     time.Duration `yaml:"ttl" json:"ttl"`
	MaxSize int           `yaml:"max_size" json:"max_size"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	ResponseTime ResponseTimeThreshold `yaml:"response_time" json:"response_time"`
	ErrorRate    ErrorRateThreshold    `yaml:"error_rate" json:"error_rate"`
	Availability AvailabilityThreshold `yaml:"availability" json:"availability"`
}

// ResponseTimeThreshold 响应时间阈值
type ResponseTimeThreshold struct {
	Warning  time.Duration `yaml:"warning" json:"warning"`
	Critical time.Duration `yaml:"critical" json:"critical"`
}

// ErrorRateThreshold 错误率阈值
type ErrorRateThreshold struct {
	Warning  float64 `yaml:"warning" json:"warning"`
	Critical float64 `yaml:"critical" json:"critical"`
}

// AvailabilityThreshold 可用性阈值
type AvailabilityThreshold struct {
	Warning  float64 `yaml:"warning" json:"warning"`
	Critical float64 `yaml:"critical" json:"critical"`
}

// ResponseFormatConfig 响应格式配置
type ResponseFormatConfig struct {
	Name        string `yaml:"name" json:"name"`
	ContentType string `yaml:"content_type" json:"content_type"`
	Template    string `yaml:"template" json:"template"`
	Default     bool   `yaml:"default" json:"default"`
}

// HealthEventListener 健康事件监听器
type HealthEventListener interface {
	OnHealthChanged(event HealthEvent)
	OnCheckerAdded(checker HealthChecker)
	OnCheckerRemoved(name string)
	OnCheckCompleted(result CheckResult)
	OnCheckFailed(name string, err error)
}

// HealthEvent 健康事件
type HealthEvent struct {
	Type      EventType     `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Service   ServiceInfo   `json:"service"`
	Previous  Status        `json:"previous"`
	Current   Status        `json:"current"`
	Details   string        `json:"details"`
	Metadata  map[string]string `json:"metadata"`
}

// EventType 事件类型
type EventType string

const (
	EventTypeHealthy   EventType = "healthy"
	EventTypeUnhealthy EventType = "unhealthy"
	EventTypeDegraded  EventType = "degraded"
	EventTypeRecovered EventType = "recovered"
)