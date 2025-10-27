package errors

import (
	"fmt"
	"time"
)

// ErrorType 定义错误类型
type ErrorType string

const (
	// 配置错误
	ErrorTypeConfig ErrorType = "CONFIG"
	// 网络错误
	ErrorTypeNetwork ErrorType = "NETWORK"
	// 存储错误
	ErrorTypeStorage ErrorType = "STORAGE"
	// 认证错误
	ErrorTypeAuth ErrorType = "AUTH"
	// 验证错误
	ErrorTypeValidation ErrorType = "VALIDATION"
	// 资源错误
	ErrorTypeResource ErrorType = "RESOURCE"
	// 超时错误
	ErrorTypeTimeout ErrorType = "TIMEOUT"
	// 系统错误
	ErrorTypeSystem ErrorType = "SYSTEM"
	// 业务逻辑错误
	ErrorTypeBusiness ErrorType = "BUSINESS"
)

// ErrorSeverity 定义错误严重程度
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "CRITICAL" // 严重错误，需要立即处理
	SeverityHigh     ErrorSeverity = "HIGH"     // 高优先级错误
	SeverityMedium   ErrorSeverity = "MEDIUM"   // 中等优先级错误
	SeverityLow      ErrorSeverity = "LOW"      // 低优先级错误
	SeverityInfo     ErrorSeverity = "INFO"     // 信息性错误
)

// MonitoringError 监控系统专用错误类型
type MonitoringError struct {
	Type        ErrorType     `json:"type"`
	Severity    ErrorSeverity `json:"severity"`
	Code        string        `json:"code"`
	Message     string        `json:"message"`
	Details     string        `json:"details,omitempty"`
	Component   string        `json:"component"`
	Operation   string        `json:"operation,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
	Cause       error         `json:"-"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Recoverable bool          `json:"recoverable"`
	RetryAfter  *time.Duration `json:"retry_after,omitempty"`
}

// Error 实现error接口
func (e *MonitoringError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *MonitoringError) Unwrap() error {
	return e.Cause
}

// IsCritical 检查是否为严重错误
func (e *MonitoringError) IsCritical() bool {
	return e.Severity == SeverityCritical
}

// IsRecoverable 检查是否可恢复
func (e *MonitoringError) IsRecoverable() bool {
	return e.Recoverable
}

// ShouldRetry 检查是否应该重试
func (e *MonitoringError) ShouldRetry() bool {
	return e.Recoverable && e.RetryAfter != nil
}

// GetRetryDelay 获取重试延迟
func (e *MonitoringError) GetRetryDelay() time.Duration {
	if e.RetryAfter != nil {
		return *e.RetryAfter
	}
	return 0
}

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	err *MonitoringError
}

// NewError 创建新的错误构建器
func NewError(errorType ErrorType, code string, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &MonitoringError{
			Type:      errorType,
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
		},
	}
}

// WithSeverity 设置错误严重程度
func (b *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Severity = severity
	return b
}

// WithDetails 设置错误详情
func (b *ErrorBuilder) WithDetails(details string) *ErrorBuilder {
	b.err.Details = details
	return b
}

// WithComponent 设置组件名称
func (b *ErrorBuilder) WithComponent(component string) *ErrorBuilder {
	b.err.Component = component
	return b
}

// WithOperation 设置操作名称
func (b *ErrorBuilder) WithOperation(operation string) *ErrorBuilder {
	b.err.Operation = operation
	return b
}

// WithCause 设置原始错误
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

// WithContext 添加上下文信息
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.err.Context[key] = value
	return b
}

// WithRecoverable 设置是否可恢复
func (b *ErrorBuilder) WithRecoverable(recoverable bool) *ErrorBuilder {
	b.err.Recoverable = recoverable
	return b
}

// WithRetryAfter 设置重试延迟
func (b *ErrorBuilder) WithRetryAfter(delay time.Duration) *ErrorBuilder {
	b.err.RetryAfter = &delay
	return b
}

// Build 构建错误
func (b *ErrorBuilder) Build() *MonitoringError {
	// 设置默认值
	if b.err.Severity == "" {
		b.err.Severity = SeverityMedium
	}
	return b.err
}

// 预定义的常见错误
var (
	// 配置错误
	ErrConfigNotFound = NewError(ErrorTypeConfig, "CONFIG_001", "Configuration file not found").
				WithSeverity(SeverityCritical).
				WithRecoverable(false).
				Build()

	ErrConfigInvalid = NewError(ErrorTypeConfig, "CONFIG_002", "Invalid configuration").
				WithSeverity(SeverityHigh).
				WithRecoverable(false).
				Build()

	// 网络错误
	ErrNetworkTimeout = NewError(ErrorTypeNetwork, "NETWORK_001", "Network operation timeout").
				WithSeverity(SeverityMedium).
				WithRecoverable(true).
				WithRetryAfter(5 * time.Second).
				Build()

	ErrNetworkUnavailable = NewError(ErrorTypeNetwork, "NETWORK_002", "Network service unavailable").
					WithSeverity(SeverityHigh).
					WithRecoverable(true).
					WithRetryAfter(10 * time.Second).
					Build()

	// 存储错误
	ErrStorageNotFound = NewError(ErrorTypeStorage, "STORAGE_001", "Storage resource not found").
				WithSeverity(SeverityMedium).
				WithRecoverable(false).
				Build()

	ErrStorageFull = NewError(ErrorTypeStorage, "STORAGE_002", "Storage space exhausted").
			WithSeverity(SeverityCritical).
			WithRecoverable(false).
			Build()

	// 认证错误
	ErrAuthInvalidCredentials = NewError(ErrorTypeAuth, "AUTH_001", "Invalid credentials").
					WithSeverity(SeverityHigh).
					WithRecoverable(false).
					Build()

	ErrAuthTokenExpired = NewError(ErrorTypeAuth, "AUTH_002", "Authentication token expired").
				WithSeverity(SeverityMedium).
				WithRecoverable(true).
				Build()

	// 验证错误
	ErrValidationFailed = NewError(ErrorTypeValidation, "VALIDATION_001", "Data validation failed").
				WithSeverity(SeverityMedium).
				WithRecoverable(false).
				Build()

	// 资源错误
	ErrResourceExhausted = NewError(ErrorTypeResource, "RESOURCE_001", "System resources exhausted").
				WithSeverity(SeverityCritical).
				WithRecoverable(true).
				WithRetryAfter(30 * time.Second).
				Build()

	ErrResourceLocked = NewError(ErrorTypeResource, "RESOURCE_002", "Resource is locked").
				WithSeverity(SeverityMedium).
				WithRecoverable(true).
				WithRetryAfter(1 * time.Second).
				Build()
)

// WrapError 包装现有错误为MonitoringError
func WrapError(err error, errorType ErrorType, code string, component string) *MonitoringError {
	return NewError(errorType, code, err.Error()).
		WithCause(err).
		WithComponent(component).
		Build()
}

// IsMonitoringError 检查是否为MonitoringError类型
func IsMonitoringError(err error) bool {
	_, ok := err.(*MonitoringError)
	return ok
}

// AsMonitoringError 转换为MonitoringError类型
func AsMonitoringError(err error) (*MonitoringError, bool) {
	if monErr, ok := err.(*MonitoringError); ok {
		return monErr, true
	}
	return nil, false
}