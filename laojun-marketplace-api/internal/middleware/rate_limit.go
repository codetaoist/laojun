package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	Requests int                       // 允许的请求数
	Window   time.Duration             // 时间窗口
	KeyFunc  func(*gin.Context) string // 生成限制键的函数
}

// RateLimitMiddleware 频率限制中间件
type RateLimitMiddleware struct {
	redisClient  *redis.Client
	config       RateLimitConfig
	rateLimitCfg *config.RateLimitConfig
}

// NewRateLimitMiddleware 创建频率限制中间件
func NewRateLimitMiddleware(redisClient *redis.Client, config RateLimitConfig) *RateLimitMiddleware {
	if config.KeyFunc == nil {
		config.KeyFunc = defaultKeyFunc
	}
	return &RateLimitMiddleware{
		redisClient: redisClient,
		config:      config,
	}
}

// NewRateLimitMiddlewareWithConfig 创建带配置的频率限制中间件
func NewRateLimitMiddlewareWithConfig(redisClient *redis.Client, config RateLimitConfig, rateLimitCfg *config.RateLimitConfig) *RateLimitMiddleware {
	if config.KeyFunc == nil {
		config.KeyFunc = defaultKeyFunc
	}
	return &RateLimitMiddleware{
		redisClient:  redisClient,
		config:       config,
		rateLimitCfg: rateLimitCfg,
	}
}

// defaultKeyFunc 默认的键生成函数（基于IP地址）
func defaultKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
}

// UserBasedKeyFunc 基于用户的键生成函数
func UserBasedKeyFunc(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
	}

	if uid, ok := userID.(uuid.UUID); ok {
		return fmt.Sprintf("rate_limit:user:%s", uid.String())
	}

	return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
}

// EndpointBasedKeyFunc 基于端点的键生成函数
func EndpointBasedKeyFunc(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	endpoint := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

	if !exists {
		return fmt.Sprintf("rate_limit:endpoint:%s:ip:%s", endpoint, c.ClientIP())
	}

	if uid, ok := userID.(uuid.UUID); ok {
		return fmt.Sprintf("rate_limit:endpoint:%s:user:%s", endpoint, uid.String())
	}

	return fmt.Sprintf("rate_limit:endpoint:%s:ip:%s", endpoint, c.ClientIP())
}

// Handler 频率限制处理中间件
func (rl *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.config.KeyFunc(c)

		// 使用Redis的滑动窗口算法检查频率限制
		allowed, remaining, resetTime, err := rl.checkRateLimit(key)
		if err != nil {
			// Redis错误时，记录日志但不阻止请求
			c.Header("X-RateLimit-Error", err.Error())
			c.Next()
			return
		}

		// 设置响应头信息
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			errorMsg := "请求过于频繁"
			detailMsg := "您的请求频率超过了限制，请稍后再试"
			retryAfter := resetTime - time.Now().Unix()

			// 如果有配置，使用配置中的信息
			if rl.rateLimitCfg != nil {
				if rl.rateLimitCfg.ErrorMessage != "" {
					errorMsg = rl.rateLimitCfg.ErrorMessage
				}
				if rl.rateLimitCfg.ErrorDetail != "" {
					detailMsg = rl.rateLimitCfg.ErrorDetail
				}
				if rl.rateLimitCfg.RetryAfterSeconds > 0 {
					retryAfter = int64(rl.rateLimitCfg.RetryAfterSeconds)
				}
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       errorMsg,
				"message":     detailMsg,
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 检查频率限制
func (rl *RateLimitMiddleware) checkRateLimit(key string) (allowed bool, remaining int, resetTime int64, err error) {
	ctx := rl.redisClient.Context()
	now := time.Now()
	windowStart := now.Add(-rl.config.Window)

	// 使用Redis的有序集合实现滑动窗口算法
	pipe := rl.redisClient.Pipeline()

	// 移除过期的请求记录
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// 添加当前请求
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// 计算当前窗口内的请求数
	pipe.ZCard(ctx, key)

	// 设置键的过期时间
	pipe.Expire(ctx, key, rl.config.Window+time.Minute)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	// 获取当前窗口内的请求数
	count := results[2].(*redis.IntCmd).Val()

	allowed = count <= int64(rl.config.Requests)
	remaining = rl.config.Requests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(rl.config.Window).Unix()

	return allowed, remaining, resetTime, nil
}

// GlobalRateLimit 全局频率限制
func GlobalRateLimit(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc:  defaultKeyFunc,
	}
	middleware := NewRateLimitMiddleware(redisClient, config)
	return middleware.Handler()
}

// UserRateLimit 用户频率限制
func UserRateLimit(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc:  UserBasedKeyFunc,
	}
	middleware := NewRateLimitMiddleware(redisClient, config)
	return middleware.Handler()
}

// EndpointRateLimit 端点频率限制
func EndpointRateLimit(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyFunc:  EndpointBasedKeyFunc,
	}
	middleware := NewRateLimitMiddleware(redisClient, config)
	return middleware.Handler()
}

// LoginRateLimit 登录频率限制
func LoginRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 5,                // 5次尝试登录失败后锁定
		Window:   15 * time.Minute, // 15分钟窗口
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("rate_limit:login:ip:%s", c.ClientIP())
		},
	}
	middleware := NewRateLimitMiddleware(redisClient, config)
	return middleware.Handler()
}

// LoginRateLimitWithConfig 使用配置的登录频率限制
func LoginRateLimitWithConfig(redisClient *redis.Client, rateLimitCfg *config.RateLimitConfig) gin.HandlerFunc {
	if !rateLimitCfg.Enabled {
		// 如果禁用频率限制，返回空中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}

	config := RateLimitConfig{
		Requests: rateLimitCfg.LoginRequests,
		Window:   time.Duration(rateLimitCfg.LoginWindowMinutes) * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return fmt.Sprintf("rate_limit:login:ip:%s", c.ClientIP())
		},
	}
	middleware := NewRateLimitMiddlewareWithConfig(redisClient, config, rateLimitCfg)
	return middleware.Handler()
}

// APIRateLimit API频率限制
func APIRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 100,       // 100次请求 per hour
		Window:   time.Hour, // 1小时窗口
		KeyFunc:  UserBasedKeyFunc,
	}
	middleware := NewRateLimitMiddleware(redisClient, config)
	return middleware.Handler()
}

// APIRateLimitWithConfig 使用配置的API频率限制
func APIRateLimitWithConfig(redisClient *redis.Client, rateLimitCfg *config.RateLimitConfig) gin.HandlerFunc {
	if !rateLimitCfg.Enabled {
		// 如果禁用频率限制，返回空中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}

	config := RateLimitConfig{
		Requests: rateLimitCfg.APIRequests,
		Window:   time.Duration(rateLimitCfg.APIWindowHours) * time.Hour,
		KeyFunc:  UserBasedKeyFunc,
	}
	middleware := NewRateLimitMiddlewareWithConfig(redisClient, config, rateLimitCfg)
	return middleware.Handler()
}
