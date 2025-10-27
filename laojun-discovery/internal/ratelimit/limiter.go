package ratelimit

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrInvalidConfig     = errors.New("invalid rate limit config")
)

// Algorithm 限流算法类型
type Algorithm string

const (
	TokenBucket   Algorithm = "token_bucket"
	LeakyBucket   Algorithm = "leaky_bucket"
	FixedWindow   Algorithm = "fixed_window"
	SlidingWindow Algorithm = "sliding_window"
)

// Config 限流器配置
type Config struct {
	Algorithm Algorithm     `yaml:"algorithm"`
	Rate      float64       `yaml:"rate"`      // 每秒允许的请求数
	Burst     int           `yaml:"burst"`     // 突发请求数
	Window    time.Duration `yaml:"window"`    // 时间窗口
	Capacity  int           `yaml:"capacity"`  // 容量
}

// RateLimiter 限流器接口
type RateLimiter interface {
	// Allow 检查是否允许请求
	Allow() bool
	// AllowN 检查是否允许N个请求
	AllowN(n int) bool
	// Wait 等待直到允许请求
	Wait(ctx context.Context) error
	// WaitN 等待直到允许N个请求
	WaitN(ctx context.Context, n int) error
	// Reserve 预留令牌
	Reserve() Reservation
	// ReserveN 预留N个令牌
	ReserveN(n int) Reservation
	// GetStats 获取统计信息
	GetStats() map[string]interface{}
	// Reset 重置限流器
	Reset()
}

// Reservation 预留信息
type Reservation struct {
	OK        bool
	Delay     time.Duration
	TimeToAct time.Time
	Limit     RateLimiter
	Tokens    int
}

// Cancel 取消预留
func (r *Reservation) Cancel() {
	if !r.OK {
		return
	}
	// 实现取消逻辑
}

// Delay 获取延迟时间
func (r *Reservation) Delay() time.Duration {
	return r.Delay
}

// Manager 限流器管理器
type Manager struct {
	limiters map[string]RateLimiter
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewManager 创建限流器管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		limiters: make(map[string]RateLimiter),
		logger:   logger,
	}
}

// GetLimiter 获取限流器
func (m *Manager) GetLimiter(key string, config *Config) (RateLimiter, error) {
	m.mu.RLock()
	limiter, exists := m.limiters[key]
	m.mu.RUnlock()

	if exists {
		return limiter, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if limiter, exists := m.limiters[key]; exists {
		return limiter, nil
	}

	// 创建新的限流器
	limiter, err := m.createLimiter(config)
	if err != nil {
		return nil, err
	}

	m.limiters[key] = limiter
	return limiter, nil
}

// createLimiter 创建限流器
func (m *Manager) createLimiter(config *Config) (RateLimiter, error) {
	switch config.Algorithm {
	case TokenBucket:
		return NewTokenBucketLimiter(config), nil
	case LeakyBucket:
		return NewLeakyBucketLimiter(config), nil
	case FixedWindow:
		return NewFixedWindowLimiter(config), nil
	case SlidingWindow:
		return NewSlidingWindowLimiter(config), nil
	default:
		return nil, ErrInvalidConfig
	}
}

// RemoveLimiter 移除限流器
func (m *Manager) RemoveLimiter(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.limiters, key)
}

// GetStats 获取所有限流器统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	for key, limiter := range m.limiters {
		stats[key] = limiter.GetStats()
	}
	return stats
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	rate     float64
	burst    int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
	
	// 统计信息
	totalRequests   uint64
	allowedRequests uint64
	rejectedRequests uint64
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(config *Config) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:     config.Rate,
		burst:    config.Burst,
		tokens:   float64(config.Burst),
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucketLimiter) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (tb *TokenBucketLimiter) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	atomic.AddUint64(&tb.totalRequests, 1)

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// 添加令牌
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}

	// 检查是否有足够的令牌
	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		atomic.AddUint64(&tb.allowedRequests, 1)
		return true
	}

	atomic.AddUint64(&tb.rejectedRequests, 1)
	return false
}

// Wait 等待直到允许请求
func (tb *TokenBucketLimiter) Wait(ctx context.Context) error {
	return tb.WaitN(ctx, 1)
}

// WaitN 等待直到允许N个请求
func (tb *TokenBucketLimiter) WaitN(ctx context.Context, n int) error {
	for {
		if tb.AllowN(n) {
			return nil
		}

		// 计算等待时间
		waitTime := time.Duration(float64(n)/tb.rate) * time.Second
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

// Reserve 预留令牌
func (tb *TokenBucketLimiter) Reserve() Reservation {
	return tb.ReserveN(1)
}

// ReserveN 预留N个令牌
func (tb *TokenBucketLimiter) ReserveN(n int) Reservation {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// 添加令牌
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}

	// 检查是否有足够的令牌
	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: now,
			Limit:     tb,
			Tokens:    n,
		}
	}

	// 计算需要等待的时间
	needed := float64(n) - tb.tokens
	delay := time.Duration(needed/tb.rate) * time.Second
	
	return Reservation{
		OK:        true,
		Delay:     delay,
		TimeToAct: now.Add(delay),
		Limit:     tb,
		Tokens:    n,
	}
}

// GetStats 获取统计信息
func (tb *TokenBucketLimiter) GetStats() map[string]interface{} {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	return map[string]interface{}{
		"algorithm":         "token_bucket",
		"rate":              tb.rate,
		"burst":             tb.burst,
		"current_tokens":    tb.tokens,
		"total_requests":    atomic.LoadUint64(&tb.totalRequests),
		"allowed_requests":  atomic.LoadUint64(&tb.allowedRequests),
		"rejected_requests": atomic.LoadUint64(&tb.rejectedRequests),
	}
}

// Reset 重置限流器
func (tb *TokenBucketLimiter) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.tokens = float64(tb.burst)
	tb.lastTime = time.Now()
	atomic.StoreUint64(&tb.totalRequests, 0)
	atomic.StoreUint64(&tb.allowedRequests, 0)
	atomic.StoreUint64(&tb.rejectedRequests, 0)
}

// LeakyBucketLimiter 漏桶限流器
type LeakyBucketLimiter struct {
	rate     float64
	capacity int
	volume   float64
	lastTime time.Time
	mu       sync.Mutex
	
	// 统计信息
	totalRequests    uint64
	allowedRequests  uint64
	rejectedRequests uint64
}

// NewLeakyBucketLimiter 创建漏桶限流器
func NewLeakyBucketLimiter(config *Config) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		rate:     config.Rate,
		capacity: config.Capacity,
		volume:   0,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (lb *LeakyBucketLimiter) Allow() bool {
	return lb.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (lb *LeakyBucketLimiter) AllowN(n int) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	atomic.AddUint64(&lb.totalRequests, 1)

	now := time.Now()
	elapsed := now.Sub(lb.lastTime).Seconds()
	lb.lastTime = now

	// 漏水
	lb.volume -= elapsed * lb.rate
	if lb.volume < 0 {
		lb.volume = 0
	}

	// 检查是否可以添加水
	if lb.volume+float64(n) <= float64(lb.capacity) {
		lb.volume += float64(n)
		atomic.AddUint64(&lb.allowedRequests, 1)
		return true
	}

	atomic.AddUint64(&lb.rejectedRequests, 1)
	return false
}

// Wait 等待直到允许请求
func (lb *LeakyBucketLimiter) Wait(ctx context.Context) error {
	return lb.WaitN(ctx, 1)
}

// WaitN 等待直到允许N个请求
func (lb *LeakyBucketLimiter) WaitN(ctx context.Context, n int) error {
	for {
		if lb.AllowN(n) {
			return nil
		}

		// 计算等待时间
		lb.mu.Lock()
		overflow := lb.volume + float64(n) - float64(lb.capacity)
		waitTime := time.Duration(overflow/lb.rate) * time.Second
		lb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

// Reserve 预留令牌
func (lb *LeakyBucketLimiter) Reserve() Reservation {
	return lb.ReserveN(1)
}

// ReserveN 预留N个令牌
func (lb *LeakyBucketLimiter) ReserveN(n int) Reservation {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastTime).Seconds()
	lb.lastTime = now

	// 漏水
	lb.volume -= elapsed * lb.rate
	if lb.volume < 0 {
		lb.volume = 0
	}

	// 检查是否可以添加水
	if lb.volume+float64(n) <= float64(lb.capacity) {
		lb.volume += float64(n)
		return Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: now,
			Limit:     lb,
			Tokens:    n,
		}
	}

	// 计算需要等待的时间
	overflow := lb.volume + float64(n) - float64(lb.capacity)
	delay := time.Duration(overflow/lb.rate) * time.Second

	return Reservation{
		OK:        true,
		Delay:     delay,
		TimeToAct: now.Add(delay),
		Limit:     lb,
		Tokens:    n,
	}
}

// GetStats 获取统计信息
func (lb *LeakyBucketLimiter) GetStats() map[string]interface{} {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	return map[string]interface{}{
		"algorithm":         "leaky_bucket",
		"rate":              lb.rate,
		"capacity":          lb.capacity,
		"current_volume":    lb.volume,
		"total_requests":    atomic.LoadUint64(&lb.totalRequests),
		"allowed_requests":  atomic.LoadUint64(&lb.allowedRequests),
		"rejected_requests": atomic.LoadUint64(&lb.rejectedRequests),
	}
}

// Reset 重置限流器
func (lb *LeakyBucketLimiter) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.volume = 0
	lb.lastTime = time.Now()
	atomic.StoreUint64(&lb.totalRequests, 0)
	atomic.StoreUint64(&lb.allowedRequests, 0)
	atomic.StoreUint64(&lb.rejectedRequests, 0)
}

// FixedWindowLimiter 固定窗口限流器
type FixedWindowLimiter struct {
	rate       int
	window     time.Duration
	counter    int
	windowStart time.Time
	mu         sync.Mutex
	
	// 统计信息
	totalRequests    uint64
	allowedRequests  uint64
	rejectedRequests uint64
}

// NewFixedWindowLimiter 创建固定窗口限流器
func NewFixedWindowLimiter(config *Config) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		rate:        int(config.Rate),
		window:      config.Window,
		counter:     0,
		windowStart: time.Now(),
	}
}

// Allow 检查是否允许请求
func (fw *FixedWindowLimiter) Allow() bool {
	return fw.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (fw *FixedWindowLimiter) AllowN(n int) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	atomic.AddUint64(&fw.totalRequests, 1)

	now := time.Now()
	
	// 检查是否需要重置窗口
	if now.Sub(fw.windowStart) >= fw.window {
		fw.counter = 0
		fw.windowStart = now
	}

	// 检查是否超过限制
	if fw.counter+n <= fw.rate {
		fw.counter += n
		atomic.AddUint64(&fw.allowedRequests, 1)
		return true
	}

	atomic.AddUint64(&fw.rejectedRequests, 1)
	return false
}

// Wait 等待直到允许请求
func (fw *FixedWindowLimiter) Wait(ctx context.Context) error {
	return fw.WaitN(ctx, 1)
}

// WaitN 等待直到允许N个请求
func (fw *FixedWindowLimiter) WaitN(ctx context.Context, n int) error {
	for {
		if fw.AllowN(n) {
			return nil
		}

		// 等待到下一个窗口
		fw.mu.Lock()
		nextWindow := fw.windowStart.Add(fw.window)
		fw.mu.Unlock()

		waitTime := time.Until(nextWindow)
		if waitTime <= 0 {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

// Reserve 预留令牌
func (fw *FixedWindowLimiter) Reserve() Reservation {
	return fw.ReserveN(1)
}

// ReserveN 预留N个令牌
func (fw *FixedWindowLimiter) ReserveN(n int) Reservation {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	
	// 检查是否需要重置窗口
	if now.Sub(fw.windowStart) >= fw.window {
		fw.counter = 0
		fw.windowStart = now
	}

	// 检查是否超过限制
	if fw.counter+n <= fw.rate {
		fw.counter += n
		return Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: now,
			Limit:     fw,
			Tokens:    n,
		}
	}

	// 需要等待到下一个窗口
	nextWindow := fw.windowStart.Add(fw.window)
	delay := nextWindow.Sub(now)

	return Reservation{
		OK:        true,
		Delay:     delay,
		TimeToAct: nextWindow,
		Limit:     fw,
		Tokens:    n,
	}
}

// GetStats 获取统计信息
func (fw *FixedWindowLimiter) GetStats() map[string]interface{} {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	return map[string]interface{}{
		"algorithm":         "fixed_window",
		"rate":              fw.rate,
		"window":            fw.window,
		"current_counter":   fw.counter,
		"window_start":      fw.windowStart,
		"total_requests":    atomic.LoadUint64(&fw.totalRequests),
		"allowed_requests":  atomic.LoadUint64(&fw.allowedRequests),
		"rejected_requests": atomic.LoadUint64(&fw.rejectedRequests),
	}
}

// Reset 重置限流器
func (fw *FixedWindowLimiter) Reset() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.counter = 0
	fw.windowStart = time.Now()
	atomic.StoreUint64(&fw.totalRequests, 0)
	atomic.StoreUint64(&fw.allowedRequests, 0)
	atomic.StoreUint64(&fw.rejectedRequests, 0)
}

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	rate     int
	window   time.Duration
	requests []time.Time
	mu       sync.Mutex
	
	// 统计信息
	totalRequests    uint64
	allowedRequests  uint64
	rejectedRequests uint64
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(config *Config) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		rate:     int(config.Rate),
		window:   config.Window,
		requests: make([]time.Time, 0),
	}
}

// Allow 检查是否允许请求
func (sw *SlidingWindowLimiter) Allow() bool {
	return sw.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (sw *SlidingWindowLimiter) AllowN(n int) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	atomic.AddUint64(&sw.totalRequests, 1)

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// 清理过期的请求
	validRequests := make([]time.Time, 0)
	for _, reqTime := range sw.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	sw.requests = validRequests

	// 检查是否超过限制
	if len(sw.requests)+n <= sw.rate {
		for i := 0; i < n; i++ {
			sw.requests = append(sw.requests, now)
		}
		atomic.AddUint64(&sw.allowedRequests, 1)
		return true
	}

	atomic.AddUint64(&sw.rejectedRequests, 1)
	return false
}

// Wait 等待直到允许请求
func (sw *SlidingWindowLimiter) Wait(ctx context.Context) error {
	return sw.WaitN(ctx, 1)
}

// WaitN 等待直到允许N个请求
func (sw *SlidingWindowLimiter) WaitN(ctx context.Context, n int) error {
	for {
		if sw.AllowN(n) {
			return nil
		}

		// 计算等待时间
		sw.mu.Lock()
		if len(sw.requests) > 0 {
			oldestRequest := sw.requests[0]
			waitTime := sw.window - time.Since(oldestRequest)
			sw.mu.Unlock()

			if waitTime > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(waitTime):
					continue
				}
			}
		} else {
			sw.mu.Unlock()
		}

		// 短暂等待后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

// Reserve 预留令牌
func (sw *SlidingWindowLimiter) Reserve() Reservation {
	return sw.ReserveN(1)
}

// ReserveN 预留N个令牌
func (sw *SlidingWindowLimiter) ReserveN(n int) Reservation {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// 清理过期的请求
	validRequests := make([]time.Time, 0)
	for _, reqTime := range sw.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	sw.requests = validRequests

	// 检查是否超过限制
	if len(sw.requests)+n <= sw.rate {
		for i := 0; i < n; i++ {
			sw.requests = append(sw.requests, now)
		}
		return Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: now,
			Limit:     sw,
			Tokens:    n,
		}
	}

	// 计算需要等待的时间
	var delay time.Duration
	if len(sw.requests) > 0 {
		oldestRequest := sw.requests[0]
		delay = sw.window - time.Since(oldestRequest)
	}

	return Reservation{
		OK:        true,
		Delay:     delay,
		TimeToAct: now.Add(delay),
		Limit:     sw,
		Tokens:    n,
	}
}

// GetStats 获取统计信息
func (sw *SlidingWindowLimiter) GetStats() map[string]interface{} {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	return map[string]interface{}{
		"algorithm":         "sliding_window",
		"rate":              sw.rate,
		"window":            sw.window,
		"current_requests":  len(sw.requests),
		"total_requests":    atomic.LoadUint64(&sw.totalRequests),
		"allowed_requests":  atomic.LoadUint64(&sw.allowedRequests),
		"rejected_requests": atomic.LoadUint64(&sw.rejectedRequests),
	}
}

// Reset 重置限流器
func (sw *SlidingWindowLimiter) Reset() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.requests = make([]time.Time, 0)
	atomic.StoreUint64(&sw.totalRequests, 0)
	atomic.StoreUint64(&sw.allowedRequests, 0)
	atomic.StoreUint64(&sw.rejectedRequests, 0)
}