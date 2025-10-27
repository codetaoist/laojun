package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter 限流中间件 - 基于 Redis 的限流实现
type RateLimiter struct {
	redisClient *redis.Client
	prefix      string
}

// RedisRateLimitConfig Redis限流配置
type RedisRateLimitConfig struct {
	Prefix      string
	DefaultRate int           // 默认每分钟请求数
	DefaultTTL  time.Duration // 默认过期时间
}

// RateLimitRule 限流规则
type RateLimitRule struct {
	Key      string        // 限流键值
	Rate     int           // 每分钟请求数
	Duration time.Duration // 时间窗口
	Burst    int           // 突发请求数
}

// NewRateLimiter 创建新的限流中间件
func NewRateLimiter(redisClient *redis.Client, config RedisRateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		prefix:      config.Prefix,
	}
}

// IPRateLimit IP 限流中间件 - 基于 IP 地址的限流
func (r *RateLimiter) IPRateLimit(rate int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:ip:%s", r.prefix, clientIP)

		allowed, remaining, resetTime, err := r.checkRateLimit(key, rate, duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests",
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimit 用户限流中间件 - 基于用户 ID 的限流
func (r *RateLimiter) UserRateLimit(rate int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// 如果没有用户ID，使用IP限流
			clientIP := c.ClientIP()
			key := fmt.Sprintf("%s:anonymous:%s", r.prefix, clientIP)

			allowed, remaining, resetTime, err := r.checkRateLimit(key, rate/2, duration) // 匿名用户限制更严格
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_server_error",
					"message": "Rate limit check failed",
				})
				c.Abort()
				return
			}

			c.Header("X-RateLimit-Limit", strconv.Itoa(rate/2))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate_limit_exceeded",
					"message":     "Too many requests",
					"retry_after": resetTime.Unix(),
				})
				c.Abort()
				return
			}
		} else {
			userIDStr := userID.(string)
			key := fmt.Sprintf("%s:user:%s", r.prefix, userIDStr)

			allowed, remaining, resetTime, err := r.checkRateLimit(key, rate, duration)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_server_error",
					"message": "Rate limit check failed",
				})
				c.Abort()
				return
			}

			c.Header("X-RateLimit-Limit", strconv.Itoa(rate))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "rate_limit_exceeded",
					"message":     "Too many requests",
					"retry_after": resetTime.Unix(),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// APIKeyRateLimit API Key 限流中间件 - 基于 API Key 的限流
func (r *RateLimiter) APIKeyRateLimit(rate int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "API key required",
			})
			c.Abort()
			return
		}

		key := fmt.Sprintf("%s:apikey:%s", r.prefix, apiKey)

		allowed, remaining, resetTime, err := r.checkRateLimit(key, rate, duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests",
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointRateLimit 端点限流中间件 - 基于请求方法和路径的限流
func (r *RateLimiter) EndpointRateLimit(rate int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		endpoint := c.Request.Method + ":" + c.FullPath()
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:endpoint:%s:%s", r.prefix, endpoint, clientIP)

		allowed, remaining, resetTime, err := r.checkRateLimit(key, rate, duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests for this endpoint",
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CustomRateLimit 自定义限流中间件
func (r *RateLimiter) CustomRateLimit(keyFunc func(*gin.Context) string, rate int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		customKey := keyFunc(c)
		if customKey == "" {
			c.Next()
			return
		}

		key := fmt.Sprintf("%s:custom:%s", r.prefix, customKey)

		allowed, remaining, resetTime, err := r.checkRateLimit(key, rate, duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests",
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查限流状态
func (r *RateLimiter) checkRateLimit(key string, rate int, duration time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
	ctx := context.Background()
	now := time.Now()
	window := now.Truncate(duration)
	resetTime = window.Add(duration)

	// 使用 Redis 管道提高性能
	pipe := r.redisClient.Pipeline()

	// 获取当前计数
	countCmd := pipe.Get(ctx, key)

	// 设置过期时间
	pipe.Expire(ctx, key, duration)

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, 0, resetTime, err
	}

	var currentCount int
	if err == redis.Nil {
		currentCount = 0
	} else {
		currentCount, err = strconv.Atoi(countCmd.Val())
		if err != nil {
			currentCount = 0
		}
	}

	if currentCount >= rate {
		return false, 0, resetTime, nil
	}

	// 增加计数
	newCount, err := r.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, 0, resetTime, err
	}

	// 如果是第一次设置，设置过期时间
	if newCount == 1 {
		r.redisClient.Expire(ctx, key, duration)
	}

	remaining = rate - int(newCount)
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime, nil
}

// GetRateLimitStatus 获取限流状态
func (r *RateLimiter) GetRateLimitStatus(key string) (count int, ttl time.Duration, err error) {
	ctx := context.Background()

	// 获取当前计数
	countStr, err := r.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}

	count, err = strconv.Atoi(countStr)
	if err != nil {
		return 0, 0, err
	}

	// 获取 TTL
	ttl, err = r.redisClient.TTL(ctx, key).Result()
	if err != nil {
		return count, 0, err
	}

	return count, ttl, nil
}

// ResetRateLimit 重置限流计数
func (r *RateLimiter) ResetRateLimit(key string) error {
	ctx := context.Background()
	return r.redisClient.Del(ctx, key).Err()
}

// BulkResetRateLimit 批量重置限流计数
func (r *RateLimiter) BulkResetRateLimit(pattern string) error {
	ctx := context.Background()

	keys, err := r.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.redisClient.Del(ctx, keys...).Err()
}

// GetRateLimitStats 获取限流统计信息
func (r *RateLimiter) GetRateLimitStats(pattern string) (map[string]int, error) {
	ctx := context.Background()

	keys, err := r.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)

	for _, key := range keys {
		countStr, err := r.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			stats[key] = 0
			continue
		}
		if err != nil {
			continue
		}

		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}

		stats[key] = count
	}

	return stats, nil
}

// CleanupExpiredKeys 清理过期的限流键
func (r *RateLimiter) CleanupExpiredKeys(pattern string) error {
	ctx := context.Background()

	keys, err := r.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	var expiredKeys []string
	for _, key := range keys {
		ttl, err := r.redisClient.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// 如果 TTL 为 -1（没有过期时间）或 -2（键不存在），则删除
		if ttl == -1 || ttl == -2 {
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		return r.redisClient.Del(ctx, expiredKeys...).Err()
	}

	return nil
}
