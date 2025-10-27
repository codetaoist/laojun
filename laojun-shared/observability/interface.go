package observability

import (
	"context"
	"time"

	"github.com/codetaoist/laojun-shared/monitoring"
	"github.com/codetaoist/laojun-shared/tracing"
)

// Observability 统一的可观测性接口，整合监控、追踪和日志功能
type Observability interface {
	// 操作管理
	StartOperation(ctx context.Context, name string, opts ...OperationOption) (Operation, context.Context)
	
	// 监控功能
	Monitor() monitoring.Monitor
	
	// 追踪功能
	Tracer() tracing.Tracer
	
	// 健康检查
	HealthCheck() HealthStatus
	
	// 导出数据
	Export(ctx context.Context, format string) error
	
	// 关闭资源
	Close() error
}

// Operation 表示一个可观测的操作
type Operation interface {
	// 基本信息
	Name() string
	ID() string
	StartTime() time.Time
	Duration() time.Duration
	
	// 状态管理
	SetStatus(status OperationStatus)
	GetStatus() OperationStatus
	
	// 属性管理
	SetAttribute(key string, value interface{})
	GetAttribute(key string) interface{}
	SetAttributes(attrs map[string]interface{})
	GetAttributes() map[string]interface{}
	
	// 事件记录
	AddEvent(name string, attrs ...EventAttribute)
	
	// 错误处理
	SetError(err error)
	GetError() error
	
	// 监控指标方法
	IncrementCounter(name string, value float64, labels ...map[string]string)
	SetGauge(name string, value float64, labels ...map[string]string)
	AddCounter(name string, value float64, labels ...map[string]string)
	RecordHistogram(name string, value float64, labels ...map[string]string)
	RecordSummary(name string, value float64, labels ...map[string]string)
	
	// 子操作
	StartChild(name string, opts ...OperationOption) Operation
	
	// 完成操作
	Finish()
	FinishWithOptions(opts FinishOptions)
}

// OperationOption 操作选项
type OperationOption func(*OperationConfig)

// OperationConfig 操作配置
type OperationConfig struct {
	// 操作类型
	Type OperationType
	
	// 初始属性
	Attributes map[string]interface{}
	
	// 初始标签
	Labels map[string]string
	
	// 采样率
	SampleRate float64
	
	// 超时时间
	Timeout time.Duration
	
	// 是否记录详细信息
	Detailed bool
	
	// 自定义字段
	CustomFields map[string]interface{}
}

// OperationType 操作类型
type OperationType string

const (
	// 通用操作类型
	OperationTypeGeneric   OperationType = "generic"
	OperationTypeHTTP      OperationType = "http"
	OperationTypeDatabase  OperationType = "database"
	OperationTypeCache     OperationType = "cache"
	OperationTypeMessage   OperationType = "message"
	OperationTypeExternal  OperationType = "external"
	OperationTypeInternal  OperationType = "internal"
	OperationTypeBatch     OperationType = "batch"
	OperationTypeScheduled OperationType = "scheduled"
)

// OperationStatus 操作状态
type OperationStatus string

const (
	// 操作状态常量
	OperationStatusUnknown   OperationStatus = "unknown"
	OperationStatusStarted   OperationStatus = "started"
	OperationStatusRunning   OperationStatus = "running"
	OperationStatusSuccess   OperationStatus = "success"
	OperationStatusError     OperationStatus = "error"
	OperationStatusTimeout   OperationStatus = "timeout"
	OperationStatusCancelled OperationStatus = "cancelled"
	OperationStatusRetrying  OperationStatus = "retrying"
)

// EventAttribute 事件属性
type EventAttribute struct {
	Key   string
	Value interface{}
}

// FinishOptions 完成选项
type FinishOptions struct {
	// 完成时间
	FinishTime *time.Time
	
	// 最终状态
	FinalStatus OperationStatus
	
	// 最终错误
	FinalError error
	
	// 额外属性
	ExtraAttributes map[string]interface{}
	
	// 是否强制导出
	ForceExport bool
}

// HealthStatus 健康状态
type HealthStatus struct {
	// 整体状态
	Status string `json:"status"`
	
	// 详细信息
	Details map[string]interface{} `json:"details"`
	
	// 检查时间
	Timestamp time.Time `json:"timestamp"`
	
	// 组件状态
	Components map[string]ComponentHealth `json:"components"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	// 状态
	Status string `json:"status"`
	
	// 消息
	Message string `json:"message,omitempty"`
	
	// 详细信息
	Details map[string]interface{} `json:"details,omitempty"`
	
	// 最后检查时间
	LastCheck time.Time `json:"last_check"`
}

// 操作选项构造函数

// WithOperationType 设置操作类型
func WithOperationType(opType OperationType) OperationOption {
	return func(config *OperationConfig) {
		config.Type = opType
	}
}

// WithAttributes 设置初始属性
func WithAttributes(attrs map[string]interface{}) OperationOption {
	return func(config *OperationConfig) {
		if config.Attributes == nil {
			config.Attributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			config.Attributes[k] = v
		}
	}
}

// WithAttribute 设置单个属性
func WithAttribute(key string, value interface{}) OperationOption {
	return func(config *OperationConfig) {
		if config.Attributes == nil {
			config.Attributes = make(map[string]interface{})
		}
		config.Attributes[key] = value
	}
}

// WithLabels 设置初始标签
func WithLabels(labels map[string]string) OperationOption {
	return func(config *OperationConfig) {
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		}
		for k, v := range labels {
			config.Labels[k] = v
		}
	}
}

// WithLabel 设置单个标签
func WithLabel(key, value string) OperationOption {
	return func(config *OperationConfig) {
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		}
		config.Labels[key] = value
	}
}

// WithSampleRate 设置采样率
func WithSampleRate(rate float64) OperationOption {
	return func(config *OperationConfig) {
		config.SampleRate = rate
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) OperationOption {
	return func(config *OperationConfig) {
		config.Timeout = timeout
	}
}

// WithDetailed 设置是否记录详细信息
func WithDetailed(detailed bool) OperationOption {
	return func(config *OperationConfig) {
		config.Detailed = detailed
	}
}

// WithCustomField 设置自定义字段
func WithCustomField(key string, value interface{}) OperationOption {
	return func(config *OperationConfig) {
		if config.CustomFields == nil {
			config.CustomFields = make(map[string]interface{})
		}
		config.CustomFields[key] = value
	}
}

// 事件属性构造函数

// NewEventAttribute 创建事件属性
func NewEventAttribute(key string, value interface{}) EventAttribute {
	return EventAttribute{
		Key:   key,
		Value: value,
	}
}

// StringAttribute 创建字符串属性
func StringAttribute(key, value string) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// IntAttribute 创建整数属性
func IntAttribute(key string, value int) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// Int64Attribute 创建64位整数属性
func Int64Attribute(key string, value int64) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// Float64Attribute 创建浮点数属性
func Float64Attribute(key string, value float64) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// BoolAttribute 创建布尔属性
func BoolAttribute(key string, value bool) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// TimeAttribute 创建时间属性
func TimeAttribute(key string, value time.Time) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// DurationAttribute 创建时长属性
func DurationAttribute(key string, value time.Duration) EventAttribute {
	return EventAttribute{Key: key, Value: value}
}

// ErrorAttribute 创建错误属性
func ErrorAttribute(err error) EventAttribute {
	return EventAttribute{Key: "error", Value: err.Error()}
}

// 完成选项构造函数

// WithFinishTime 设置完成时间
func WithFinishTime(t time.Time) func(*FinishOptions) {
	return func(opts *FinishOptions) {
		opts.FinishTime = &t
	}
}

// WithFinalStatus 设置最终状态
func WithFinalStatus(status OperationStatus) func(*FinishOptions) {
	return func(opts *FinishOptions) {
		opts.FinalStatus = status
	}
}

// WithFinalError 设置最终错误
func WithFinalError(err error) func(*FinishOptions) {
	return func(opts *FinishOptions) {
		opts.FinalError = err
	}
}

// WithExtraAttributes 设置额外属性
func WithExtraAttributes(attrs map[string]interface{}) func(*FinishOptions) {
	return func(opts *FinishOptions) {
		if opts.ExtraAttributes == nil {
			opts.ExtraAttributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			opts.ExtraAttributes[k] = v
		}
	}
}

// WithForceExport 设置强制导出
func WithForceExport(force bool) func(*FinishOptions) {
	return func(opts *FinishOptions) {
		opts.ForceExport = force
	}
}