package errors

import (
	"testing"
	"time"
)

// simpleError 用于测试的简单错误类型
type simpleError struct {
	msg string
}

func (e simpleError) Error() string {
	return e.msg
}

func TestMonitoringError(t *testing.T) {
	t.Run("创建基本错误", func(t *testing.T) {
		err := NewError(ErrorTypeConfig, "CONFIG_001", "Configuration error").
			WithSeverity(SeverityHigh).
			WithComponent("TestComponent").
			WithOperation("TestOperation").
			Build()

		if err.Type != ErrorTypeConfig {
			t.Errorf("Expected error type %s, got %s", ErrorTypeConfig, err.Type)
		}
		if err.Code != "CONFIG_001" {
			t.Errorf("Expected error code CONFIG_001, got %s", err.Code)
		}
		if err.Message != "Configuration error" {
			t.Errorf("Expected message 'Configuration error', got %s", err.Message)
		}
		if err.Severity != SeverityHigh {
			t.Errorf("Expected severity %s, got %s", SeverityHigh, err.Severity)
		}
		if err.Component != "TestComponent" {
			t.Errorf("Expected component 'TestComponent', got %s", err.Component)
		}
		if err.Operation != "TestOperation" {
			t.Errorf("Expected operation 'TestOperation', got %s", err.Operation)
		}
	})

	t.Run("错误方法测试", func(t *testing.T) {
		err := NewError(ErrorTypeSystem, "SYS_001", "System error").
			WithSeverity(SeverityCritical).
			WithRecoverable(true).
			WithRetryAfter(5 * time.Second).
			Build()

		if !err.IsCritical() {
			t.Error("Expected error to be critical")
		}
		if !err.IsRecoverable() {
			t.Error("Expected error to be recoverable")
		}
		if !err.ShouldRetry() {
			t.Error("Expected error to allow retry")
		}
		if err.GetRetryDelay() != 5*time.Second {
			t.Errorf("Expected retry delay 5s, got %v", err.GetRetryDelay())
		}
	})

	t.Run("错误字符串表示", func(t *testing.T) {
		err := NewError(ErrorTypeNetwork, "NET_001", "Network timeout").
			WithComponent("HTTPClient").
			Build()

		errorStr := err.Error()
		if errorStr == "" {
			t.Error("Error string should not be empty")
		}
		t.Logf("Error string: %s", errorStr)
	})
}

func TestErrorBuilder(t *testing.T) {
	t.Run("构建器模式", func(t *testing.T) {
		err := NewError(ErrorTypeValidation, "VAL_001", "Validation failed").
			WithDetails("Field 'name' is required").
			WithContext("field", "name").
			WithContext("value", "").
			WithRecoverable(false).
			Build()

		if err.Details != "Field 'name' is required" {
			t.Errorf("Expected details 'Field 'name' is required', got %s", err.Details)
		}
		if err.Context["field"] != "name" {
			t.Errorf("Expected context field 'name', got %v", err.Context["field"])
		}
		if err.Context["value"] != "" {
			t.Errorf("Expected context value '', got %v", err.Context["value"])
		}
		if err.Recoverable {
			t.Error("Expected error to be non-recoverable")
		}
	})

	t.Run("默认值设置", func(t *testing.T) {
		err := NewError(ErrorTypeAuth, "AUTH_001", "Authentication failed").Build()

		if err.Severity != SeverityMedium {
			t.Errorf("Expected default severity %s, got %s", SeverityMedium, err.Severity)
		}
		if err.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	})
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *MonitoringError
		expected struct {
			errorType   ErrorType
			code        string
			severity    ErrorSeverity
			recoverable bool
		}
	}{
		{
			name: "配置文件未找到",
			err:  ErrConfigNotFound,
			expected: struct {
				errorType   ErrorType
				code        string
				severity    ErrorSeverity
				recoverable bool
			}{ErrorTypeConfig, "CONFIG_001", SeverityCritical, false},
		},
		{
			name: "网络超时",
			err:  ErrNetworkTimeout,
			expected: struct {
				errorType   ErrorType
				code        string
				severity    ErrorSeverity
				recoverable bool
			}{ErrorTypeNetwork, "NETWORK_001", SeverityMedium, true},
		},
		{
			name: "资源耗尽",
			err:  ErrResourceExhausted,
			expected: struct {
				errorType   ErrorType
				code        string
				severity    ErrorSeverity
				recoverable bool
			}{ErrorTypeResource, "RESOURCE_001", SeverityCritical, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Type != tt.expected.errorType {
				t.Errorf("Expected type %s, got %s", tt.expected.errorType, tt.err.Type)
			}
			if tt.err.Code != tt.expected.code {
				t.Errorf("Expected code %s, got %s", tt.expected.code, tt.err.Code)
			}
			if tt.err.Severity != tt.expected.severity {
				t.Errorf("Expected severity %s, got %s", tt.expected.severity, tt.err.Severity)
			}
			if tt.err.Recoverable != tt.expected.recoverable {
				t.Errorf("Expected recoverable %v, got %v", tt.expected.recoverable, tt.err.Recoverable)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := NewError(ErrorTypeSystem, "SYS_001", "Original error").Build()
	
	wrappedErr := WrapError(originalErr, ErrorTypeNetwork, "NET_001", "NetworkComponent")
	
	if wrappedErr.Type != ErrorTypeNetwork {
		t.Errorf("Expected wrapped error type %s, got %s", ErrorTypeNetwork, wrappedErr.Type)
	}
	if wrappedErr.Code != "NET_001" {
		t.Errorf("Expected wrapped error code NET_001, got %s", wrappedErr.Code)
	}
	if wrappedErr.Component != "NetworkComponent" {
		t.Errorf("Expected component 'NetworkComponent', got %s", wrappedErr.Component)
	}
	if wrappedErr.Cause != originalErr {
		t.Error("Expected wrapped error to contain original error as cause")
	}
}

func TestIsMonitoringError(t *testing.T) {
	t.Run("MonitoringError类型", func(t *testing.T) {
		err := NewError(ErrorTypeSystem, "SYS_001", "System error").Build()
		if !IsMonitoringError(err) {
			t.Error("Expected IsMonitoringError to return true for MonitoringError")
		}
	})

	t.Run("普通error类型", func(t *testing.T) {
		err := NewError(ErrorTypeSystem, "SYS_001", "System error").Build()
		if !IsMonitoringError(err) {
			t.Error("Expected IsMonitoringError to return true for MonitoringError")
		}
	})
}

func TestAsMonitoringError(t *testing.T) {
	t.Run("转换MonitoringError", func(t *testing.T) {
		originalErr := NewError(ErrorTypeSystem, "SYS_001", "System error").Build()
		
		monErr, ok := AsMonitoringError(originalErr)
		if !ok {
			t.Error("Expected AsMonitoringError to return true for MonitoringError")
		}
		if monErr != originalErr {
			t.Error("Expected AsMonitoringError to return the same error")
		}
	})

	t.Run("转换普通error", func(t *testing.T) {
		// 创建一个普通的error来测试
		err := simpleError{msg: "simple error"}
		
		_, ok := AsMonitoringError(err)
		if ok {
			t.Error("Expected AsMonitoringError to return false for non-MonitoringError")
		}
	})
}