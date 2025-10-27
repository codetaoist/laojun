package errors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RecoveryStrategy 恢复策略接口
type RecoveryStrategy interface {
	// ShouldRecover 判断是否应该尝试恢复
	ShouldRecover(err *MonitoringError) bool
	// Recover 执行恢复操作
	Recover(ctx context.Context, err *MonitoringError) error
	// GetMaxRetries 获取最大重试次数
	GetMaxRetries() int
	// GetBackoffStrategy 获取退避策略
	GetBackoffStrategy() BackoffStrategy
}

// BackoffStrategy 退避策略
type BackoffStrategy interface {
	// NextDelay 计算下次重试的延迟时间
	NextDelay(attempt int) time.Duration
	// Reset 重置退避状态
	Reset()
}

// ExponentialBackoff 指数退避策略
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// NewExponentialBackoff 创建指数退避策略
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, multiplier float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   multiplier,
	}
}

// NextDelay 计算下次重试的延迟时间
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return e.InitialDelay
	}
	
	delay := float64(e.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= e.Multiplier
	}
	
	if time.Duration(delay) > e.MaxDelay {
		return e.MaxDelay
	}
	
	return time.Duration(delay)
}

// Reset 重置退避状态
func (e *ExponentialBackoff) Reset() {
	// 指数退避策略无需重置状态
}

// LinearBackoff 线性退避策略
type LinearBackoff struct {
	InitialDelay time.Duration
	Increment    time.Duration
	MaxDelay     time.Duration
}

// NewLinearBackoff 创建线性退避策略
func NewLinearBackoff(initialDelay, increment time.Duration) *LinearBackoff {
	return &LinearBackoff{
		InitialDelay: initialDelay,
		Increment:    increment,
		MaxDelay:     30 * time.Second, // 默认最大延迟
	}
}

// NextDelay 计算下次重试的延迟时间
func (l *LinearBackoff) NextDelay(attempt int) time.Duration {
	delay := l.InitialDelay + time.Duration(attempt)*l.Increment
	if delay > l.MaxDelay {
		return l.MaxDelay
	}
	return delay
}

// Reset 重置退避状态
func (l *LinearBackoff) Reset() {
	// 线性退避策略无需重置状态
}

// FixedBackoff 固定延迟退避策略
type FixedBackoff struct {
	Delay time.Duration
}

// NewFixedBackoff 创建固定延迟退避策略
func NewFixedBackoff(delay time.Duration) *FixedBackoff {
	return &FixedBackoff{
		Delay: delay,
	}
}

// NextDelay 计算下次重试的延迟时间
func (f *FixedBackoff) NextDelay(attempt int) time.Duration {
	return f.Delay
}

// Reset 重置退避状态
func (f *FixedBackoff) Reset() {
	// 固定延迟策略无需重置状态
}

// DefaultRecoveryStrategy 默认恢复策略
type DefaultRecoveryStrategy struct {
	MaxRetries      int
	BackoffStrategy BackoffStrategy
	Logger          *zap.Logger
}

// NewDefaultRecoveryStrategy 创建默认恢复策略
func NewDefaultRecoveryStrategy(backoffStrategy BackoffStrategy, maxRetries int) *DefaultRecoveryStrategy {
	return &DefaultRecoveryStrategy{
		MaxRetries:      maxRetries,
		BackoffStrategy: backoffStrategy,
		Logger:          nil, // 将在使用时设置
	}
}

// NewDefaultRecoveryStrategyWithLogger 创建带日志的默认恢复策略
func NewDefaultRecoveryStrategyWithLogger(logger *zap.Logger) *DefaultRecoveryStrategy {
	return &DefaultRecoveryStrategy{
		MaxRetries: 3,
		BackoffStrategy: &ExponentialBackoff{
			InitialDelay: 1 * time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
		Logger: logger,
	}
}

// ShouldRecover 判断是否应该尝试恢复
func (d *DefaultRecoveryStrategy) ShouldRecover(err *MonitoringError) bool {
	// 只有可恢复的错误才尝试恢复
	return err.IsRecoverable() && !err.IsCritical()
}

// Recover 执行恢复操作
func (d *DefaultRecoveryStrategy) Recover(ctx context.Context, err *MonitoringError) error {
	if d.Logger != nil {
		d.Logger.Info("Attempting to recover from error",
			zap.String("error_type", string(err.Type)),
			zap.String("error_code", err.Code),
			zap.String("component", err.Component))
	}
	
	// 基本的恢复逻辑：等待一段时间
	if err.RetryAfter != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*err.RetryAfter):
			return nil
		}
	}
	
	return nil
}

// GetMaxRetries 获取最大重试次数
func (d *DefaultRecoveryStrategy) GetMaxRetries() int {
	return d.MaxRetries
}

// GetBackoffStrategy 获取退避策略
func (d *DefaultRecoveryStrategy) GetBackoffStrategy() BackoffStrategy {
	return d.BackoffStrategy
}

// RecoveryManager 恢复管理器
type RecoveryManager struct {
	strategies map[ErrorType]RecoveryStrategy
	logger     *zap.Logger
	mu         sync.RWMutex
	
	// 重试状态跟踪
	retryStates map[string]*RetryState
}

// RetryState 重试状态
type RetryState struct {
	Attempts    int
	LastAttempt time.Time
	Component   string
	Operation   string
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager(logger *zap.Logger) *RecoveryManager {
	return &RecoveryManager{
		strategies:  make(map[ErrorType]RecoveryStrategy),
		logger:      logger,
		retryStates: make(map[string]*RetryState),
	}
}

// RegisterStrategy 注册恢复策略
func (r *RecoveryManager) RegisterStrategy(errorType ErrorType, strategy RecoveryStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategies[errorType] = strategy
}

// HandleError 处理错误并尝试恢复
func (r *RecoveryManager) HandleError(ctx context.Context, err *MonitoringError) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 记录错误
	r.logger.Error("Error occurred",
		zap.String("type", string(err.Type)),
		zap.String("code", err.Code),
		zap.String("message", err.Message),
		zap.String("component", err.Component),
		zap.String("operation", err.Operation),
		zap.String("severity", string(err.Severity)))
	
	// 检查是否有对应的恢复策略
	strategy, exists := r.strategies[err.Type]
	if !exists {
		// 使用默认策略
		strategy = NewDefaultRecoveryStrategyWithLogger(r.logger)
	}
	
	// 检查是否应该尝试恢复
	if !strategy.ShouldRecover(err) {
		r.logger.Info("Error is not recoverable",
			zap.String("error_code", err.Code),
			zap.String("component", err.Component))
		return err
	}
	
	// 获取重试状态
	stateKey := fmt.Sprintf("%s:%s:%s", err.Component, err.Operation, err.Code)
	state, exists := r.retryStates[stateKey]
	if !exists {
		state = &RetryState{
			Component: err.Component,
			Operation: err.Operation,
		}
		r.retryStates[stateKey] = state
	}
	
	// 检查是否超过最大重试次数
	if state.Attempts >= strategy.GetMaxRetries() {
		r.logger.Error("Max retry attempts exceeded",
			zap.String("error_code", err.Code),
			zap.String("component", err.Component),
			zap.Int("attempts", state.Attempts),
			zap.Int("max_retries", strategy.GetMaxRetries()))
		
		// 清理重试状态
		delete(r.retryStates, stateKey)
		return err
	}
	
	// 计算退避延迟
	backoff := strategy.GetBackoffStrategy()
	delay := backoff.NextDelay(state.Attempts)
	
	r.logger.Info("Attempting recovery",
		zap.String("error_code", err.Code),
		zap.String("component", err.Component),
		zap.Int("attempt", state.Attempts+1),
		zap.Duration("delay", delay))
	
	// 等待退避时间
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
	}
	
	// 更新重试状态
	state.Attempts++
	state.LastAttempt = time.Now()
	
	// 执行恢复操作
	if recoveryErr := strategy.Recover(ctx, err); recoveryErr != nil {
		r.logger.Error("Recovery failed",
			zap.String("error_code", err.Code),
			zap.String("component", err.Component),
			zap.Error(recoveryErr))
		return recoveryErr
	}
	
	// 恢复成功，清理重试状态
	delete(r.retryStates, stateKey)
	backoff.Reset()
	
	r.logger.Info("Recovery successful",
		zap.String("error_code", err.Code),
		zap.String("component", err.Component),
		zap.Int("attempts", state.Attempts))
	
	return nil
}

// GetRetryState 获取重试状态
func (r *RecoveryManager) GetRetryState(component, operation, errorCode string) *RetryState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stateKey := fmt.Sprintf("%s:%s:%s", component, operation, errorCode)
	if state, exists := r.retryStates[stateKey]; exists {
		// 返回副本以避免并发修改
		return &RetryState{
			Attempts:    state.Attempts,
			LastAttempt: state.LastAttempt,
			Component:   state.Component,
			Operation:   state.Operation,
		}
	}
	return nil
}

// ClearRetryState 清理重试状态
func (r *RecoveryManager) ClearRetryState(component, operation, errorCode string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	stateKey := fmt.Sprintf("%s:%s:%s", component, operation, errorCode)
	delete(r.retryStates, stateKey)
}

// GetRetryStates 获取所有重试状态
func (r *RecoveryManager) GetRetryStates() map[string]*RetryState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	states := make(map[string]*RetryState)
	for key, state := range r.retryStates {
		states[key] = &RetryState{
			Attempts:    state.Attempts,
			LastAttempt: state.LastAttempt,
			Component:   state.Component,
			Operation:   state.Operation,
		}
	}
	return states
}