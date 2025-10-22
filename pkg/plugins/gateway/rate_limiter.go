package gateway

import (
	"sync"
	"time"
)

// NewRateLimiter 创建新的限流中心
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*TokenBucket),
	}
}

// Allow 检查是否允许一个请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	bucket, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		// 创建新的令牌桶（默认配置：容量100，每秒补0个令牌）
		rl.mu.Lock()
		bucket = NewTokenBucket(100, 10) // 容量100，每秒补0个令牌
		rl.limiters[key] = bucket
		rl.mu.Unlock()
	}

	return bucket.Allow()
}

// AllowN 检查是否允许N个请求
func (rl *RateLimiter) AllowN(key string, n int) bool {
	rl.mu.RLock()
	bucket, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		bucket = NewTokenBucket(100, 10)
		rl.limiters[key] = bucket
		rl.mu.Unlock()
	}

	return bucket.AllowN(n)
}

// SetLimit 设置限流配置
func (rl *RateLimiter) SetLimit(key string, capacity, rate int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiters[key] = NewTokenBucket(capacity, rate)
}

// RemoveLimit 移除限流配置
func (rl *RateLimiter) RemoveLimit(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.limiters, key)
}

// GetStats 获取限流统计
func (rl *RateLimiter) GetStats(key string) map[string]interface{} {
	rl.mu.RLock()
	bucket, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if !exists {
		return nil
	}

	return bucket.GetStats()
}

// Cleanup 清理过期的限流器
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.limiters {
		if now.Sub(bucket.lastRefill) > maxAge {
			delete(rl.limiters, key)
		}
	}
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许一个请求
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}

	return false
}

// refill 补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	if elapsed <= 0 {
		return
	}

	// 计算应该补充的令牌数
	tokensToAdd := int(elapsed.Seconds()) * tb.rate
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// GetStats 获取令牌桶统计信息
func (tb *TokenBucket) GetStats() map[string]interface{} {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	return map[string]interface{}{
		"capacity":    tb.capacity,
		"tokens":      tb.tokens,
		"rate":        tb.rate,
		"last_refill": tb.lastRefill,
		"utilization": float64(tb.capacity-tb.tokens) / float64(tb.capacity),
	}
}

// SetCapacity 设置容量
func (tb *TokenBucket) SetCapacity(capacity int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.capacity = capacity
	if tb.tokens > capacity {
		tb.tokens = capacity
	}
}

// SetRate 设置补充速率
func (tb *TokenBucket) SetRate(rate int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.rate = rate
}

// SlidingWindowRateLimiter 滑动窗口限流中心
type SlidingWindowRateLimiter struct {
	windows map[string]*SlidingWindow
	mu      sync.RWMutex
}

type SlidingWindow struct {
	requests []time.Time
	limit    int
	window   time.Duration
	mu       sync.Mutex
}

func NewSlidingWindowRateLimiter() *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		windows: make(map[string]*SlidingWindow),
	}
}

func (sw *SlidingWindowRateLimiter) Allow(key string, limit int, window time.Duration) bool {
	sw.mu.RLock()
	slidingWindow, exists := sw.windows[key]
	sw.mu.RUnlock()

	if !exists {
		sw.mu.Lock()
		slidingWindow = &SlidingWindow{
			requests: make([]time.Time, 0),
			limit:    limit,
			window:   window,
		}
		sw.windows[key] = slidingWindow
		sw.mu.Unlock()
	}

	return slidingWindow.Allow()
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// 移除过期的请求
	validRequests := make([]time.Time, 0)
	for _, req := range sw.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	sw.requests = validRequests

	// 检查是否超过限流阈值
	if len(sw.requests) >= sw.limit {
		return false
	}

	// 添加当前请求
	sw.requests = append(sw.requests, now)
	return true
}

// FixedWindowRateLimiter 固定窗口限流中心
type FixedWindowRateLimiter struct {
	windows map[string]*FixedWindow
	mu      sync.RWMutex
}

type FixedWindow struct {
	count       int
	limit       int
	window      time.Duration
	windowStart time.Time
	mu          sync.Mutex
}

func NewFixedWindowRateLimiter() *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		windows: make(map[string]*FixedWindow),
	}
}

func (fw *FixedWindowRateLimiter) Allow(key string, limit int, window time.Duration) bool {
	fw.mu.RLock()
	fixedWindow, exists := fw.windows[key]
	fw.mu.RUnlock()

	if !exists {
		fw.mu.Lock()
		fixedWindow = &FixedWindow{
			count:       0,
			limit:       limit,
			window:      window,
			windowStart: time.Now(),
		}
		fw.windows[key] = fixedWindow
		fw.mu.Unlock()
	}

	return fixedWindow.Allow()
}

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()

	// 检查是否需要重置窗口
	if now.Sub(fw.windowStart) >= fw.window {
		fw.count = 0
		fw.windowStart = now
	}

	// 检查是否超过限流阈值
	if fw.count >= fw.limit {
		return false
	}

	fw.count++
	return true
}

// LeakyBucketRateLimiter 漏桶限流中心
type LeakyBucketRateLimiter struct {
	buckets map[string]*LeakyBucket
	mu      sync.RWMutex
}

type LeakyBucket struct {
	capacity int
	queue    []time.Time
	leakRate time.Duration
	lastLeak time.Time
	mu       sync.Mutex
}

func NewLeakyBucketRateLimiter() *LeakyBucketRateLimiter {
	return &LeakyBucketRateLimiter{
		buckets: make(map[string]*LeakyBucket),
	}
}

func (lb *LeakyBucketRateLimiter) Allow(key string, capacity int, leakRate time.Duration) bool {
	lb.mu.RLock()
	bucket, exists := lb.buckets[key]
	lb.mu.RUnlock()

	if !exists {
		lb.mu.Lock()
		bucket = &LeakyBucket{
			capacity: capacity,
			queue:    make([]time.Time, 0),
			leakRate: leakRate,
			lastLeak: time.Now(),
		}
		lb.buckets[key] = bucket
		lb.mu.Unlock()
	}

	return bucket.Allow()
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	// 检查桶是否已满
	if len(lb.queue) >= lb.capacity {
		return false
	}

	// 添加请求到队列
	lb.queue = append(lb.queue, time.Now())
	return true
}

func (lb *LeakyBucket) leak() {
	now := time.Now()
	elapsed := now.Sub(lb.lastLeak)

	// 计算应该漏出的请求数
	leakCount := int(elapsed / lb.leakRate)
	if leakCount > 0 {
		if leakCount >= len(lb.queue) {
			lb.queue = make([]time.Time, 0)
		} else {
			lb.queue = lb.queue[leakCount:]
		}
		lb.lastLeak = now
	}
}

// HierarchicalRateLimiter 分层限流中心
type HierarchicalRateLimiter struct {
	globalLimiter *TokenBucket
	userLimiters  map[string]*TokenBucket
	ipLimiters    map[string]*TokenBucket
	mu            sync.RWMutex
}

func NewHierarchicalRateLimiter(globalCapacity, globalRate int) *HierarchicalRateLimiter {
	return &HierarchicalRateLimiter{
		globalLimiter: NewTokenBucket(globalCapacity, globalRate),
		userLimiters:  make(map[string]*TokenBucket),
		ipLimiters:    make(map[string]*TokenBucket),
	}
}

func (hr *HierarchicalRateLimiter) Allow(userID, clientIP string) bool {
	// 首先检查全局限制
	if !hr.globalLimiter.Allow() {
		return false
	}

	// 检查用户限流阈值
	if userID != "" {
		hr.mu.RLock()
		userLimiter, exists := hr.userLimiters[userID]
		hr.mu.RUnlock()

		if !exists {
			hr.mu.Lock()
			userLimiter = NewTokenBucket(50, 5) // 用户默认限制
			hr.userLimiters[userID] = userLimiter
			hr.mu.Unlock()
		}

		if !userLimiter.Allow() {
			return false
		}
	}

	// 检查IP限制
	if clientIP != "" {
		hr.mu.RLock()
		ipLimiter, exists := hr.ipLimiters[clientIP]
		hr.mu.RUnlock()

		if !exists {
			hr.mu.Lock()
			ipLimiter = NewTokenBucket(20, 2) // IP默认限制
			hr.ipLimiters[clientIP] = ipLimiter
			hr.mu.Unlock()
		}

		if !ipLimiter.Allow() {
			return false
		}
	}

	return true
}

func (hr *HierarchicalRateLimiter) SetUserLimit(userID string, capacity, rate int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.userLimiters[userID] = NewTokenBucket(capacity, rate)
}

func (hr *HierarchicalRateLimiter) SetIPLimit(clientIP string, capacity, rate int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.ipLimiters[clientIP] = NewTokenBucket(capacity, rate)
}

func (hr *HierarchicalRateLimiter) GetStats() map[string]interface{} {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	return map[string]interface{}{
		"global":     hr.globalLimiter.GetStats(),
		"user_count": len(hr.userLimiters),
		"ip_count":   len(hr.ipLimiters),
	}
}
