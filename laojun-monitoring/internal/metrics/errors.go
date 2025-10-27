package metrics

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	ErrMetricAlreadyExists      = errors.New("metric already exists")
	ErrMetricNotFound          = errors.New("metric not found")
	ErrMetricTypeMismatch      = errors.New("metric type mismatch")
	ErrUnsupportedMetricType   = errors.New("unsupported metric type")
	ErrMetricUnregisterFailed  = errors.New("failed to unregister metric")
	ErrInvalidMetricName       = errors.New("invalid metric name")
	ErrInvalidLabelName        = errors.New("invalid label name")
	ErrInvalidLabelValue       = errors.New("invalid label value")
	ErrMetricValueOutOfRange   = errors.New("metric value out of range")
)

// MetricError 指标错误
type MetricError struct {
	Op     string // 操作名称
	Metric string // 指标名称
	Err    error  // 原始错误
}

func (e *MetricError) Error() string {
	if e.Metric != "" {
		return fmt.Sprintf("metric %s: %s: %v", e.Metric, e.Op, e.Err)
	}
	return fmt.Sprintf("metric operation %s: %v", e.Op, e.Err)
}

func (e *MetricError) Unwrap() error {
	return e.Err
}

// NewMetricError 创建指标错误
func NewMetricError(op, metric string, err error) *MetricError {
	return &MetricError{
		Op:     op,
		Metric: metric,
		Err:    err,
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %s (value: %v): %s", e.Field, e.Value, e.Message)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}