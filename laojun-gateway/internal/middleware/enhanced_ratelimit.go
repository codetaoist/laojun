package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnhancedRateLimitMiddleware 增强限流中间件
type EnhancedRateLimitMiddleware struct {
	config      *config.RateLimitConfig
	logger      *zap.Logger
	limiters    map[string]*RateLimiter
	mutex       sync.RWMutex
	stats       *RateLimitStats
}

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
	GetRemaining(key string) int
	GetResetTime(key string) time.Time
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	capacity   int
	tokens     int
	rate       int
	lastRefill time.Time
	mutex      sync.Mutex
}

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	windowSize time.Duration
	limit      int
	requests   map[string][]time.Time
	mutex      sync.RWMutex
}

// RateLimitStats 限流统计
type RateLimitStats struct {
	TotalRequests   int64
	BlockedRequests int64
	AllowedRequests int64
	mutex           sync.RWMutex
}

// NewEnhancedRateLimitMiddleware 创建增强限流中间件
func NewEnhancedRateLimitMiddleware(cfg *config.RateLimitConfig, logger *zap.Logger) *EnhancedRateLimitMiddleware {
	return &EnhancedRateLimitMiddleware{
		config:   cfg,
		logger:   logger,
		limiters: make(map[string]*RateLimiter),
		stats:    &RateLimitStats{},
	}
}

// GlobalRateLimitMiddleware 全局限流中间件
func (erlm *EnhancedRateLimitMiddleware) GlobalRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !erlm.config.Enabled {
			c.Next()
			return
		}

		erlm.incrementTotalRequests()

		// 检查全局限流
		if !erlm.checkGlobalRateLimit() {
			erlm.incrementBlockedRequests()
			erlm.handleRateLimitExceeded(c, "GLOBAL_RATE_LIMIT_EXCEEDED", "Global rate limit exceeded")
			return
		}

		erlm.incrementAllowedRequests()
		c.Next()
	}
}

// IPRateLimitMiddleware IP限流中间件
func (erlm *EnhancedRateLimitMiddleware) IPRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !erlm.config.Enabled {
			c.Next()
			return
		}

		clientIP := erlm.getClientIP(c)
		
		// 检查IP白名单
		if erlm.isIPWhitelisted(clientIP) {
			c.Next()
			return
		}

		// 检查IP限流
		if !erlm.checkIPRateLimit(clientIP) {
			erlm.incrementBlockedRequests()
			erlm.handleRateLimitExceeded(c, "IP_RATE_LIMIT_EXCEEDED", 
				fmt.Sprintf("IP rate limit exceeded for %s", clientIP))
			return
		}

		c.Next()
	}
}

// PathRateLimitMiddleware 路径限流中间件
func (erlm *EnhancedRateLimitMiddleware) PathRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !erlm.config.Enabled {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		// 查找匹配的路径规则
		rule := erlm.findPathRule(path, method)
		if rule == nil {
			c.Next()
			return
		}

		// 检查路径限流
		key := fmt.Sprintf("path:%s:%s", method, path)
		if !erlm.checkRateLimit(key, rule.Rate, rule.Burst) {
			erlm.incrementBlockedRequests()
			erlm.handleRateLimitExceeded(c, "PATH_RATE_LIMIT_EXCEEDED", 
				fmt.Sprintf("Path rate limit exceeded for %s %s", method, path))
			return
		}

		c.Next()
	}
}

// UserRateLimitMiddleware 用户限流中间件
func (erlm *EnhancedRateLimitMiddleware) UserRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !erlm.config.Enabled {
			c.Next()
			return
		}

		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		// 检查用户限流
		if !erlm.checkUserRateLimit(userIDStr) {
			erlm.incrementBlockedRequests()
			erlm.handleRateLimitExceeded(c, "USER_RATE_LIMIT_EXCEEDED", 
				fmt.Sprintf("User rate limit exceeded for user %s", userIDStr))
			return
		}

		c.Next()
	}
}

// APIKeyRateLimitMiddleware API密钥限流中间件
func (erlm *EnhancedRateLimitMiddleware) APIKeyRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !erlm.config.Enabled {
			c.Next()
			return
		}

		// 获取API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.Next()
			return
		}

		// 检查API Key限流
		if !erlm.checkAPIKeyRateLimit(apiKey) {
			erlm.incrementBlockedRequests()
			erlm.handleRateLimitExceeded(c, "API_KEY_RATE_LIMIT_EXCEEDED", 
				"API key rate limit exceeded")
			return
		}

		c.Next()
	}
}

// AdaptiveRateLimitMiddleware 自适应限流中间件
func (erlm *EnhancedRateLimitMiddleware) AdaptiveRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !erlm.config.Enabled {
			c.Next()
			return
		}

		// 根据系统负载动态调整限流
		loadFactor := erlm.calculateSystemLoad()
		adjustedRate := erlm.adjustRateByLoad(erlm.config.GlobalRate, loadFactor)

		if !erlm.checkRateLimit("adaptive", adjustedRate, erlm.config.GlobalBurst) {
			erlm.incrementBlockedRequests()
			erlm.handleRateLimitExceeded(c, "ADAPTIVE_RATE_LIMIT_EXCEEDED", 
				"Adaptive rate limit exceeded due to high system load")
			return
		}

		c.Next()
	}
}

// checkGlobalRateLimit 检查全局限流
func (erlm *EnhancedRateLimitMiddleware) checkGlobalRateLimit() bool {
	return erlm.checkRateLimit("global", erlm.config.GlobalRate, erlm.config.GlobalBurst)
}

// checkIPRateLimit 检查IP限流
func (erlm *EnhancedRateLimitMiddleware) checkIPRateLimit(ip string) bool {
	key := fmt.Sprintf("ip:%s", ip)
	return erlm.checkRateLimit(key, erlm.config.IPRate, erlm.config.IPBurst)
}

// checkUserRateLimit 检查用户限流
func (erlm *EnhancedRateLimitMiddleware) checkUserRateLimit(userID string) bool {
	key := fmt.Sprintf("user:%s", userID)
	return erlm.checkRateLimit(key, erlm.config.UserRate, erlm.config.UserBurst)
}

// checkAPIKeyRateLimit 检查API Key限流
func (erlm *EnhancedRateLimitMiddleware) checkAPIKeyRateLimit(apiKey string) bool {
	key := fmt.Sprintf("apikey:%s", apiKey)
	return erlm.checkRateLimit(key, erlm.config.APIKeyRate, erlm.config.APIKeyBurst)
}

// checkRateLimit 通用限流检查
func (erlm *EnhancedRateLimitMiddleware) checkRateLimit(key string, rate, burst int) bool {
	erlm.mutex.Lock()
	defer erlm.mutex.Unlock()

	limiter, exists := erlm.limiters[key]
	if !exists {
		// 根据配置创建限流器
		switch erlm.config.Algorithm {
		case "token_bucket":
			limiter = erlm.createTokenBucketLimiter(rate, burst)
		case "sliding_window":
			limiter = erlm.createSlidingWindowLimiter(rate, time.Minute)
		default:
			limiter = erlm.createTokenBucketLimiter(rate, burst)
		}
		erlm.limiters[key] = &limiter
	}

	return (*limiter).Allow(key)
}

// createTokenBucketLimiter 创建令牌桶限流器
func (erlm *EnhancedRateLimitMiddleware) createTokenBucketLimiter(rate, capacity int) RateLimiter {
	return &TokenBucketLimiter{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// createSlidingWindowLimiter 创建滑动窗口限流器
func (erlm *EnhancedRateLimitMiddleware) createSlidingWindowLimiter(limit int, window time.Duration) RateLimiter {
	return &SlidingWindowLimiter{
		windowSize: window,
		limit:      limit,
		requests:   make(map[string][]time.Time),
	}
}

// Allow 令牌桶允许检查
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tbl.lastRefill)
	
	// 添加令牌
	tokensToAdd := int(elapsed.Seconds()) * tbl.rate
	tbl.tokens = min(tbl.capacity, tbl.tokens+tokensToAdd)
	tbl.lastRefill = now

	if tbl.tokens > 0 {
		tbl.tokens--
		return true
	}

	return false
}

// GetRemaining 获取剩余令牌
func (tbl *TokenBucketLimiter) GetRemaining(key string) int {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	return tbl.tokens
}

// GetResetTime 获取重置时间
func (tbl *TokenBucketLimiter) GetResetTime(key string) time.Time {
	return tbl.lastRefill.Add(time.Second)
}

// Allow 滑动窗口允许检查
func (swl *SlidingWindowLimiter) Allow(key string) bool {
	swl.mutex.Lock()
	defer swl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-swl.windowSize)

	// 获取或创建请求记录
	requests, exists := swl.requests[key]
	if !exists {
		requests = make([]time.Time, 0)
	}

	// 清理过期请求
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= swl.limit {
		swl.requests[key] = validRequests
		return false
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	swl.requests[key] = validRequests
	return true
}

// GetRemaining 获取剩余请求数
func (swl *SlidingWindowLimiter) GetRemaining(key string) int {
	swl.mutex.RLock()
	defer swl.mutex.RUnlock()

	requests, exists := swl.requests[key]
	if !exists {
		return swl.limit
	}

	now := time.Now()
	windowStart := now.Add(-swl.windowSize)
	
	validCount := 0
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validCount++
		}
	}

	return swl.limit - validCount
}

// GetResetTime 获取重置时间
func (swl *SlidingWindowLimiter) GetResetTime(key string) time.Time {
	swl.mutex.RLock()
	defer swl.mutex.RUnlock()

	requests, exists := swl.requests[key]
	if !exists || len(requests) == 0 {
		return time.Now()
	}

	// 返回最早请求的过期时间
	earliest := requests[0]
	for _, reqTime := range requests {
		if reqTime.Before(earliest) {
			earliest = reqTime
		}
	}

	return earliest.Add(swl.windowSize)
}

// findPathRule 查找路径规则
func (erlm *EnhancedRateLimitMiddleware) findPathRule(path, method string) *config.RateLimitRule {
	for _, rule := range erlm.config.Rules {
		if erlm.matchPath(path, rule.Path) && erlm.matchMethod(method, rule.Method) {
			return &rule
		}
	}
	return nil
}

// matchPath 匹配路径
func (erlm *EnhancedRateLimitMiddleware) matchPath(path, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

// matchMethod 匹配方法
func (erlm *EnhancedRateLimitMiddleware) matchMethod(method, pattern string) bool {
	return pattern == "*" || pattern == method
}

// isIPWhitelisted 检查IP是否在白名单中
func (erlm *EnhancedRateLimitMiddleware) isIPWhitelisted(ip string) bool {
	for _, whiteIP := range erlm.config.WhiteList {
		if ip == whiteIP {
			return true
		}
	}
	return false
}

// getClientIP 获取客户端IP
func (erlm *EnhancedRateLimitMiddleware) getClientIP(c *gin.Context) string {
	// 优先从X-Forwarded-For获取
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// 从X-Real-IP获取
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 使用默认IP
	return c.ClientIP()
}

// calculateSystemLoad 计算系统负载
func (erlm *EnhancedRateLimitMiddleware) calculateSystemLoad() float64 {
	// 这里可以实现真实的系统负载计算
	// 例如：CPU使用率、内存使用率、连接数等
	return 1.0 // 默认负载因子
}

// adjustRateByLoad 根据负载调整速率
func (erlm *EnhancedRateLimitMiddleware) adjustRateByLoad(baseRate int, loadFactor float64) int {
	adjustedRate := float64(baseRate) / loadFactor
	return int(adjustedRate)
}

// handleRateLimitExceeded 处理限流超出
func (erlm *EnhancedRateLimitMiddleware) handleRateLimitExceeded(c *gin.Context, code, message string) {
	erlm.logger.Warn("Rate limit exceeded",
		zap.String("ip", c.ClientIP()),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("code", code))

	// 设置限流相关的响应头
	c.Header("X-RateLimit-Limit", strconv.Itoa(erlm.config.GlobalRate))
	c.Header("X-RateLimit-Remaining", "0")
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":     message,
		"code":      code,
		"timestamp": time.Now().Unix(),
		"retry_after": 60,
	})
	c.Abort()
}

// 统计方法
func (erlm *EnhancedRateLimitMiddleware) incrementTotalRequests() {
	erlm.stats.mutex.Lock()
	defer erlm.stats.mutex.Unlock()
	erlm.stats.TotalRequests++
}

func (erlm *EnhancedRateLimitMiddleware) incrementBlockedRequests() {
	erlm.stats.mutex.Lock()
	defer erlm.stats.mutex.Unlock()
	erlm.stats.BlockedRequests++
}

func (erlm *EnhancedRateLimitMiddleware) incrementAllowedRequests() {
	erlm.stats.mutex.Lock()
	defer erlm.stats.mutex.Unlock()
	erlm.stats.AllowedRequests++
}

// GetStats 获取统计信息
func (erlm *EnhancedRateLimitMiddleware) GetStats() *RateLimitStats {
	erlm.stats.mutex.RLock()
	defer erlm.stats.mutex.RUnlock()
	
	return &RateLimitStats{
		TotalRequests:   erlm.stats.TotalRequests,
		BlockedRequests: erlm.stats.BlockedRequests,
		AllowedRequests: erlm.stats.AllowedRequests,
	}
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}