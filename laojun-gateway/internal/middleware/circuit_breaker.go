package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name            string
	maxRequests     uint32
	interval        time.Duration
	timeout         time.Duration
	failureThreshold uint32
	successThreshold uint32
	
	mutex       sync.Mutex
	state       CircuitBreakerState
	generation  uint64
	counts      *Counts
	expiry      time.Time
	
	onStateChange func(name string, from CircuitBreakerState, to CircuitBreakerState)
}

// Counts 计数器
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreakerMiddleware 熔断器中间件
type CircuitBreakerMiddleware struct {
	breakers map[string]*CircuitBreaker
	config   *config.CircuitBreakerConfig
	logger   *zap.Logger
	mutex    sync.RWMutex
}

// NewCircuitBreakerMiddleware 创建熔断器中间件
func NewCircuitBreakerMiddleware(cfg *config.CircuitBreakerConfig, logger *zap.Logger) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		breakers: make(map[string]*CircuitBreaker),
		config:   cfg,
		logger:   logger,
	}
}

// CircuitBreakerMiddleware 熔断器中间件函数
func (cbm *CircuitBreakerMiddleware) CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cbm.config.Enabled {
			c.Next()
			return
		}

		// 获取服务名称
		serviceName := cbm.getServiceName(c)
		if serviceName == "" {
			c.Next()
			return
		}

		// 获取或创建熔断器
		breaker := cbm.getOrCreateBreaker(serviceName)

		// 执行请求
		err := breaker.Execute(func() error {
			c.Next()
			
			// 检查响应状态码
			if c.Writer.Status() >= 500 {
				return fmt.Errorf("server error: %d", c.Writer.Status())
			}
			
			return nil
		})

		if err != nil {
			cbm.handleCircuitBreakerError(c, serviceName, breaker.state, err)
		}
	}
}

// ServiceCircuitBreakerMiddleware 服务级熔断器中间件
func (cbm *CircuitBreakerMiddleware) ServiceCircuitBreakerMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cbm.config.Enabled {
			c.Next()
			return
		}

		breaker := cbm.getOrCreateBreaker(serviceName)

		err := breaker.Execute(func() error {
			c.Next()
			
			if c.Writer.Status() >= 500 {
				return fmt.Errorf("server error: %d", c.Writer.Status())
			}
			
			return nil
		})

		if err != nil {
			cbm.handleCircuitBreakerError(c, serviceName, breaker.state, err)
		}
	}
}

// PathCircuitBreakerMiddleware 路径级熔断器中间件
func (cbm *CircuitBreakerMiddleware) PathCircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cbm.config.Enabled {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method
		breakerName := fmt.Sprintf("%s:%s", method, path)

		breaker := cbm.getOrCreateBreaker(breakerName)

		err := breaker.Execute(func() error {
			c.Next()
			
			if c.Writer.Status() >= 500 {
				return fmt.Errorf("server error: %d", c.Writer.Status())
			}
			
			return nil
		})

		if err != nil {
			cbm.handleCircuitBreakerError(c, breakerName, breaker.state, err)
		}
	}
}

// getOrCreateBreaker 获取或创建熔断器
func (cbm *CircuitBreakerMiddleware) getOrCreateBreaker(name string) *CircuitBreaker {
	cbm.mutex.RLock()
	breaker, exists := cbm.breakers[name]
	cbm.mutex.RUnlock()

	if exists {
		return breaker
	}

	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	// 双重检查
	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}

	// 创建新的熔断器
	breaker = &CircuitBreaker{
		name:             name,
		maxRequests:      cbm.config.MaxRequests,
		interval:         time.Duration(cbm.config.Interval) * time.Second,
		timeout:          time.Duration(cbm.config.Timeout) * time.Second,
		failureThreshold: cbm.config.FailureThreshold,
		successThreshold: cbm.config.SuccessThreshold,
		state:            StateClosed,
		counts:           &Counts{},
		onStateChange:    cbm.onStateChange,
	}

	cbm.breakers[name] = breaker
	return breaker
}

// Execute 执行请求
func (cb *CircuitBreaker) Execute(req func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	err = req()
	cb.afterRequest(generation, err == nil)
	return err
}

// beforeRequest 请求前检查
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, fmt.Errorf("circuit breaker is open")
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, fmt.Errorf("circuit breaker is half-open and max requests exceeded")
	}

	cb.counts.Requests++
	return generation, nil
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

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

// currentState 获取当前状态
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
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

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
	case StateHalfOpen:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
		if cb.counts.ConsecutiveSuccesses >= cb.successThreshold {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
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

// readyToTrip 检查是否准备跳闸
func (cb *CircuitBreaker) readyToTrip(counts *Counts) bool {
	return counts.ConsecutiveFailures >= cb.failureThreshold
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// toNewGeneration 转到新一代
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = &Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// getServiceName 获取服务名称
func (cbm *CircuitBreakerMiddleware) getServiceName(c *gin.Context) string {
	// 从路由配置中获取服务名称
	if service, exists := c.Get("service_name"); exists {
		if serviceName, ok := service.(string); ok {
			return serviceName
		}
	}

	// 从路径中提取服务名称
	path := c.Request.URL.Path
	if len(path) > 1 {
		parts := strings.Split(path[1:], "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	return "default"
}

// onStateChange 状态变化回调
func (cbm *CircuitBreakerMiddleware) onStateChange(name string, from, to CircuitBreakerState) {
	cbm.logger.Info("Circuit breaker state changed",
		zap.String("name", name),
		zap.String("from", from.String()),
		zap.String("to", to.String()))
}

// handleCircuitBreakerError 处理熔断器错误
func (cbm *CircuitBreakerMiddleware) handleCircuitBreakerError(c *gin.Context, serviceName string, state CircuitBreakerState, err error) {
	cbm.logger.Warn("Circuit breaker triggered",
		zap.String("service", serviceName),
		zap.String("state", state.String()),
		zap.String("path", c.Request.URL.Path),
		zap.Error(err))

	statusCode := http.StatusServiceUnavailable
	message := "Service temporarily unavailable"
	code := "CIRCUIT_BREAKER_OPEN"

	if state == StateHalfOpen {
		message = "Service is recovering, please try again later"
		code = "CIRCUIT_BREAKER_HALF_OPEN"
	}

	c.JSON(statusCode, gin.H{
		"error":     message,
		"code":      code,
		"service":   serviceName,
		"state":     state.String(),
		"timestamp": time.Now().Unix(),
		"retry_after": cbm.config.Timeout,
	})
	c.Abort()
}

// GetBreakerStatus 获取熔断器状态
func (cbm *CircuitBreakerMiddleware) GetBreakerStatus(name string) map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	breaker, exists := cbm.breakers[name]
	if !exists {
		return nil
	}

	breaker.mutex.Lock()
	defer breaker.mutex.Unlock()

	return map[string]interface{}{
		"name":                   breaker.name,
		"state":                  breaker.state.String(),
		"generation":             breaker.generation,
		"requests":               breaker.counts.Requests,
		"total_successes":        breaker.counts.TotalSuccesses,
		"total_failures":         breaker.counts.TotalFailures,
		"consecutive_successes":  breaker.counts.ConsecutiveSuccesses,
		"consecutive_failures":   breaker.counts.ConsecutiveFailures,
		"failure_threshold":      breaker.failureThreshold,
		"success_threshold":      breaker.successThreshold,
		"max_requests":           breaker.maxRequests,
		"interval":               breaker.interval.Seconds(),
		"timeout":                breaker.timeout.Seconds(),
		"expiry":                 breaker.expiry.Unix(),
	}
}

// GetAllBreakerStatus 获取所有熔断器状态
func (cbm *CircuitBreakerMiddleware) GetAllBreakerStatus() map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	result := make(map[string]interface{})
	for name := range cbm.breakers {
		result[name] = cbm.GetBreakerStatus(name)
	}

	return result
}

// ResetBreaker 重置熔断器
func (cbm *CircuitBreakerMiddleware) ResetBreaker(name string) error {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	breaker, exists := cbm.breakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	breaker.mutex.Lock()
	defer breaker.mutex.Unlock()

	breaker.state = StateClosed
	breaker.counts = &Counts{}
	breaker.generation++
	breaker.expiry = time.Time{}

	cbm.logger.Info("Circuit breaker reset", zap.String("name", name))
	return nil
}

// ForceOpen 强制打开熔断器
func (cbm *CircuitBreakerMiddleware) ForceOpen(name string) error {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	breaker, exists := cbm.breakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	breaker.mutex.Lock()
	defer breaker.mutex.Unlock()

	breaker.setState(StateOpen, time.Now())

	cbm.logger.Info("Circuit breaker forced open", zap.String("name", name))
	return nil
}

// ForceClose 强制关闭熔断器
func (cbm *CircuitBreakerMiddleware) ForceClose(name string) error {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	breaker, exists := cbm.breakers[name]
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}

	breaker.mutex.Lock()
	defer breaker.mutex.Unlock()

	breaker.setState(StateClosed, time.Now())

	cbm.logger.Info("Circuit breaker forced close", zap.String("name", name))
	return nil
}