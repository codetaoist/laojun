package circuit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests")
)

// State 熔断器状态
type State int32

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// Config 熔断器配置
type Config struct {
	// 失败阈值
	FailureThreshold uint32 `yaml:"failure_threshold"`
	// 成功阈值（半开状态下）
	SuccessThreshold uint32 `yaml:"success_threshold"`
	// 超时时间
	Timeout time.Duration `yaml:"timeout"`
	// 最大请求数（半开状态下）
	MaxRequests uint32 `yaml:"max_requests"`
	// 统计窗口时间
	Interval time.Duration `yaml:"interval"`
	// 最小请求数
	MinRequests uint32 `yaml:"min_requests"`
	// 失败率阈值
	FailureRatio float64 `yaml:"failure_ratio"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          60 * time.Second,
		MaxRequests:      1,
		Interval:         60 * time.Second,
		MinRequests:      3,
		FailureRatio:     0.6,
	}
}

// Counts 统计信息
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// OnStateChange 状态变更回调
type OnStateChange func(name string, from State, to State)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name           string
	config         *Config
	state          State
	generation     uint64
	counts         Counts
	expiry         time.Time
	onStateChange  OnStateChange
	logger         *zap.Logger
	mu             sync.RWMutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(name string, config *Config, logger *zap.Logger) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig()
	}

	cb := &CircuitBreaker{
		name:   name,
		config: config,
		state:  StateClosed,
		logger: logger,
	}

	cb.toNewGeneration(time.Now())
	return cb
}

// SetOnStateChange 设置状态变更回调
func (cb *CircuitBreaker) SetOnStateChange(fn OnStateChange) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

// Name 获取熔断器名称
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// State 获取当前状态
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Counts 获取统计信息
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// Execute 执行函数
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := recover(); e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result, err := req()
	cb.afterRequest(generation, err == nil)
	return result, err
}

// Call 调用函数（简化版本）
func (cb *CircuitBreaker) Call(fn func() error) error {
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

// beforeRequest 请求前检查
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitOpen
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.config.MaxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.Requests++
	return generation, nil
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
	case StateHalfOpen:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
		if cb.counts.ConsecutiveSuccesses >= cb.config.SuccessThreshold {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalFailures++
		cb.counts.ConsecutiveFailures++
		cb.counts.ConsecutiveSuccesses = 0
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

// currentState 获取当前状态
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}

	cb.logger.Info("Circuit breaker state changed",
		zap.String("name", cb.name),
		zap.String("from", prev.String()),
		zap.String("to", state.String()))
}

// toNewGeneration 开始新的统计周期
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.config.Interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.config.Interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.config.Timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// readyToTrip 检查是否应该跳闸
func (cb *CircuitBreaker) readyToTrip(counts Counts) bool {
	return counts.Requests >= cb.config.MinRequests &&
		counts.ConsecutiveFailures >= cb.config.FailureThreshold ||
		(counts.Requests >= cb.config.MinRequests &&
			float64(counts.TotalFailures)/float64(counts.Requests) >= cb.config.FailureRatio)
}

// GetStats 获取统计信息
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":       cb.name,
		"state":      cb.state.String(),
		"generation": cb.generation,
		"counts":     cb.counts,
		"expiry":     cb.expiry,
		"config":     cb.config,
	}
}

// Manager 熔断器管理器
type Manager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewManager 创建熔断器管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetBreaker 获取熔断器
func (m *Manager) GetBreaker(name string, config *Config) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[name]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if breaker, exists := m.breakers[name]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(name, config, m.logger)
	m.breakers[name] = breaker
	return breaker
}

// RemoveBreaker 移除熔断器
func (m *Manager) RemoveBreaker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.breakers, name)
}

// ListBreakers 列出所有熔断器
func (m *Manager) ListBreakers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.breakers))
	for name := range m.breakers {
		names = append(names, name)
	}
	return names
}

// GetStats 获取所有熔断器统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	for name, breaker := range m.breakers {
		stats[name] = breaker.GetStats()
	}
	return stats
}

// Reset 重置熔断器
func (m *Manager) Reset(name string) error {
	m.mu.RLock()
	breaker, exists := m.breakers[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.state = StateClosed
	breaker.toNewGeneration(time.Now())
	return nil
}

// ResetAll 重置所有熔断器
func (m *Manager) ResetAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, breaker := range m.breakers {
		breaker.mu.Lock()
		breaker.state = StateClosed
		breaker.toNewGeneration(time.Now())
		breaker.mu.Unlock()
	}
}

// TwoStepCircuitBreaker 两步熔断器（支持异步操作）
type TwoStepCircuitBreaker struct {
	*CircuitBreaker
}

// NewTwoStepCircuitBreaker 创建两步熔断器
func NewTwoStepCircuitBreaker(name string, config *Config, logger *zap.Logger) *TwoStepCircuitBreaker {
	return &TwoStepCircuitBreaker{
		CircuitBreaker: NewCircuitBreaker(name, config, logger),
	}
}

// Allow 检查是否允许请求
func (cb *TwoStepCircuitBreaker) Allow() (done func(success bool), err error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	return func(success bool) {
		cb.afterRequest(generation, success)
	}, nil
}

// HTTPCircuitBreaker HTTP熔断器
type HTTPCircuitBreaker struct {
	*CircuitBreaker
	statusCodes map[int]bool // 哪些状态码算作失败
}

// NewHTTPCircuitBreaker 创建HTTP熔断器
func NewHTTPCircuitBreaker(name string, config *Config, logger *zap.Logger) *HTTPCircuitBreaker {
	return &HTTPCircuitBreaker{
		CircuitBreaker: NewCircuitBreaker(name, config, logger),
		statusCodes:    map[int]bool{500: true, 502: true, 503: true, 504: true},
	}
}

// SetFailureStatusCodes 设置失败状态码
func (hcb *HTTPCircuitBreaker) SetFailureStatusCodes(codes []int) {
	hcb.statusCodes = make(map[int]bool)
	for _, code := range codes {
		hcb.statusCodes[code] = true
	}
}

// IsFailureStatusCode 检查是否为失败状态码
func (hcb *HTTPCircuitBreaker) IsFailureStatusCode(statusCode int) bool {
	return hcb.statusCodes[statusCode]
}

// ExecuteHTTP 执行HTTP请求
func (hcb *HTTPCircuitBreaker) ExecuteHTTP(req func() (int, error)) (int, error) {
	result, err := hcb.Execute(func() (interface{}, error) {
		statusCode, err := req()
		if err != nil {
			return statusCode, err
		}
		if hcb.IsFailureStatusCode(statusCode) {
			return statusCode, fmt.Errorf("HTTP error: %d", statusCode)
		}
		return statusCode, nil
	})

	if result != nil {
		return result.(int), err
	}
	return 0, err
}

// BulkheadCircuitBreaker 舱壁模式熔断器
type BulkheadCircuitBreaker struct {
	*CircuitBreaker
	semaphore chan struct{}
}

// NewBulkheadCircuitBreaker 创建舱壁模式熔断器
func NewBulkheadCircuitBreaker(name string, config *Config, maxConcurrent int, logger *zap.Logger) *BulkheadCircuitBreaker {
	return &BulkheadCircuitBreaker{
		CircuitBreaker: NewCircuitBreaker(name, config, logger),
		semaphore:      make(chan struct{}, maxConcurrent),
	}
}

// Execute 执行函数（带并发控制）
func (bcb *BulkheadCircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	// 获取信号量
	select {
	case bcb.semaphore <- struct{}{}:
		defer func() { <-bcb.semaphore }()
	default:
		return nil, ErrTooManyRequests
	}

	return bcb.CircuitBreaker.Execute(req)
}

// GetConcurrentRequests 获取当前并发请求数
func (bcb *BulkheadCircuitBreaker) GetConcurrentRequests() int {
	return len(bcb.semaphore)
}

// GetMaxConcurrentRequests 获取最大并发请求数
func (bcb *BulkheadCircuitBreaker) GetMaxConcurrentRequests() int {
	return cap(bcb.semaphore)
}