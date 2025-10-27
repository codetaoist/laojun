package errors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestExponentialBackoff(t *testing.T) {
	backoff := NewExponentialBackoff(1*time.Second, 10*time.Second, 2.0)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 10 * time.Second}, // 达到最大延迟
		{5, 10 * time.Second}, // 保持最大延迟
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := backoff.NextDelay(tt.attempt)
			if delay != tt.expected {
				t.Errorf("Expected delay %v for attempt %d, got %v", tt.expected, tt.attempt, delay)
			}
		})
	}

	// 测试重置
	backoff.Reset()
	delay := backoff.NextDelay(0)
	if delay != 1*time.Second {
		t.Errorf("Expected delay 1s after reset, got %v", delay)
	}
}

func TestLinearBackoff(t *testing.T) {
	backoff := NewLinearBackoff(1*time.Second, 500*time.Millisecond)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 1500 * time.Millisecond},
		{2, 2 * time.Second},
		{3, 2500 * time.Millisecond},
		{4, 3 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := backoff.NextDelay(tt.attempt)
			if delay != tt.expected {
				t.Errorf("Expected delay %v for attempt %d, got %v", tt.expected, tt.attempt, delay)
			}
		})
	}

	// 测试最大延迟限制
	delay := backoff.NextDelay(100) // 很大的尝试次数
	if delay > 30*time.Second {
		t.Errorf("Expected delay to not exceed max delay 30s, got %v", delay)
	}
}

func TestFixedBackoff(t *testing.T) {
	backoff := NewFixedBackoff(2 * time.Second)

	for i := 0; i < 5; i++ {
		delay := backoff.NextDelay(i)
		if delay != 2*time.Second {
			t.Errorf("Expected fixed delay 2s for attempt %d, got %v", i, delay)
		}
	}

	backoff.Reset()
	delay := backoff.NextDelay(0)
	if delay != 2*time.Second {
		t.Errorf("Expected delay 2s after reset, got %v", delay)
	}
}

func TestDefaultRecoveryStrategy(t *testing.T) {
	backoff := NewExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0)
	strategy := NewDefaultRecoveryStrategy(backoff, 3)

	t.Run("可恢复错误", func(t *testing.T) {
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithRecoverable(true).
			Build()

		if !strategy.ShouldRecover(err) {
			t.Error("Expected strategy to allow recovery for recoverable error")
		}
	})

	t.Run("不可恢复错误", func(t *testing.T) {
		err := NewError(ErrorTypeConfig, "CONFIG_001", "Config error").
			WithRecoverable(false).
			Build()

		if strategy.ShouldRecover(err) {
			t.Error("Expected strategy to not allow recovery for non-recoverable error")
		}
	})

	t.Run("恢复尝试", func(t *testing.T) {
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithRecoverable(true).
			Build()

		ctx := context.Background()
		recoveryErr := strategy.Recover(ctx, err)
		
		// 默认策略应该成功恢复（返回nil）
		if recoveryErr != nil {
			t.Errorf("Expected default strategy to recover successfully, got error: %v", recoveryErr)
		}
	})

	t.Run("获取配置", func(t *testing.T) {
		if strategy.GetMaxRetries() != 3 {
			t.Errorf("Expected max retries 3, got %d", strategy.GetMaxRetries())
		}
		if strategy.GetBackoffStrategy() != backoff {
			t.Error("Expected backoff strategy to match")
		}
	})
}

func TestDefaultRecoveryStrategyWithLogger(t *testing.T) {
	logger := zap.NewNop()
	strategy := NewDefaultRecoveryStrategyWithLogger(logger)

	if strategy.Logger != logger {
		t.Error("Expected logger to be set")
	}
	if strategy.GetMaxRetries() != 3 {
		t.Errorf("Expected default max retries 3, got %d", strategy.GetMaxRetries())
	}
}

func TestRecoveryManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRecoveryManager(logger)

	t.Run("注册策略", func(t *testing.T) {
		strategy := NewDefaultRecoveryStrategy(
			NewFixedBackoff(1*time.Second),
			2,
		)
		
		manager.RegisterStrategy(ErrorTypeNetwork, strategy)
		
		// 验证策略已注册（通过处理错误来间接验证）
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithRecoverable(true).
			Build()
		
		ctx := context.Background()
		result := manager.HandleError(ctx, err)
		
		// 默认策略应该成功恢复（返回nil）
		if result != nil {
			t.Errorf("Expected manager to recover successfully, got error: %v", result)
		}
	})

	t.Run("处理未注册类型的错误", func(t *testing.T) {
		err := NewError(ErrorTypeStorage, "STORAGE_001", "Storage error").
			WithRecoverable(true).
			Build()
		
		ctx := context.Background()
		result := manager.HandleError(ctx, err)
		
		// 应该使用默认策略并成功恢复
		if result != nil {
			t.Errorf("Expected manager to recover successfully with default strategy, got error: %v", result)
		}
	})

	t.Run("重试状态管理", func(t *testing.T) {
		// 初始状态应该为空
		state := manager.GetRetryState("TestComponent", "TestOperation", "TEST_001")
		if state != nil {
			t.Error("Expected retry state to be nil initially")
		}

		// 注册一个会失败的策略来保持重试状态
		failingStrategy := &TestRecoveryStrategy{
			maxRetries:    3,
			backoff:       NewFixedBackoff(100 * time.Millisecond),
			shouldRecover: true,
			recoverResult: fmt.Errorf("recovery failed"), // 恢复失败
		}
		manager.RegisterStrategy(ErrorTypeNetwork, failingStrategy)

		// 创建一个错误来触发状态创建
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithComponent("TestComponent").
			WithOperation("TestOperation").
			WithRecoverable(true).
			Build()

		ctx := context.Background()
		result := manager.HandleError(ctx, err)

		// 恢复应该失败
		if result == nil {
			t.Error("Expected recovery to fail")
		}

		// 现在应该有状态了
		state = manager.GetRetryState("TestComponent", "TestOperation", "NET_001")
		if state == nil {
			t.Error("Expected retry state to be created after handling error")
		}
		if state.Attempts != 1 {
			t.Errorf("Expected attempts 1, got %d", state.Attempts)
		}
		if state.Component != "TestComponent" {
			t.Errorf("Expected component 'TestComponent', got %s", state.Component)
		}

		// 清除重试状态
		manager.ClearRetryState("TestComponent", "TestOperation", "NET_001")
		
		// 验证状态已清除
		state = manager.GetRetryState("TestComponent", "TestOperation", "NET_001")
		if state != nil {
			t.Error("Expected retry state to be cleared")
		}
	})

	t.Run("重试状态更新", func(t *testing.T) {
		// 创建一个错误并多次处理以测试重试状态
		err := NewError(ErrorTypeNetwork, "NET_001", "Network error").
			WithComponent("TestComponent").
			WithOperation("TestOperation").
			WithRecoverable(true).
			Build()

		ctx := context.Background()
		
		// 第一次处理
		manager.HandleError(ctx, err)
		state1 := manager.GetRetryState("TestComponent", "TestOperation", "NET_001")
		if state1.Attempts != 1 {
			t.Errorf("Expected attempts 1 after first error, got %d", state1.Attempts)
		}

		// 第二次处理
		manager.HandleError(ctx, err)
		state2 := manager.GetRetryState("TestComponent", "TestOperation", "NET_001")
		if state2.Attempts != 2 {
			t.Errorf("Expected attempts 2 after second error, got %d", state2.Attempts)
		}

		// 验证时间戳更新
		if !state2.LastAttempt.After(state1.LastAttempt) {
			t.Error("Expected last attempt time to be updated")
		}
	})
}

func TestRecoveryManagerWithCustomStrategy(t *testing.T) {
	logger := zap.NewNop()
	manager := NewRecoveryManager(logger)

	// 创建自定义策略
	customStrategy := &TestRecoveryStrategy{
		maxRetries: 5,
		backoff:    NewFixedBackoff(500 * time.Millisecond),
		shouldRecover: true,
	}

	manager.RegisterStrategy(ErrorTypeSystem, customStrategy)

	err := NewError(ErrorTypeSystem, "SYS_001", "System error").
		WithComponent("TestComponent").
		WithOperation("TestOperation").
		WithRecoverable(true).
		Build()

	ctx := context.Background()
	result := manager.HandleError(ctx, err)

	if !customStrategy.recoverCalled {
		t.Error("Expected custom strategy Recover method to be called")
	}

	// 验证返回的是自定义策略的结果
	if result != customStrategy.recoverResult {
		t.Error("Expected result from custom strategy")
	}
}

// TestRecoveryStrategy 用于测试的自定义恢复策略
type TestRecoveryStrategy struct {
	maxRetries     int
	backoff        BackoffStrategy
	shouldRecover  bool
	recoverCalled  bool
	recoverResult  error
}

func (t *TestRecoveryStrategy) ShouldRecover(err *MonitoringError) bool {
	return t.shouldRecover
}

func (t *TestRecoveryStrategy) Recover(ctx context.Context, err *MonitoringError) error {
	t.recoverCalled = true
	t.recoverResult = err // 返回原始错误作为测试
	return t.recoverResult
}

func (t *TestRecoveryStrategy) GetMaxRetries() int {
	return t.maxRetries
}

func (t *TestRecoveryStrategy) GetBackoffStrategy() BackoffStrategy {
	return t.backoff
}