package middleware

import (
	"net/http"

	"github.com/codetaoist/laojun-gateway/internal/services/ratelimit"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rateLimitService ratelimit.Service, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果限流服务未启用，直接跳过
		if rateLimitService == nil {
			c.Next()
			return
		}
		// 检查全局限流
		allowed, err := rateLimitService.AllowGlobal()
		if err != nil {
			logger.Error("Global rate limit check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limit check failed",
				"code":  "RATE_LIMIT_ERROR",
			})
			c.Abort()
			return
		}

		if !allowed {
			logger.Warn("Global rate limit exceeded", 
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Global rate limit exceeded",
				"code":  "GLOBAL_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		// 检查IP限流
		allowed, err = rateLimitService.AllowIP(c.ClientIP())
		if err != nil {
			logger.Error("IP rate limit check failed", 
				zap.String("ip", c.ClientIP()),
				zap.Error(err))
		} else if !allowed {
			logger.Warn("IP rate limit exceeded", 
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "IP rate limit exceeded",
				"code":  "IP_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		// 检查路径限流
		allowed, err = rateLimitService.AllowPath(c.Request.URL.Path, c.Request.Method)
		if err != nil {
			logger.Error("Path rate limit check failed", 
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Error(err))
		} else if !allowed {
			logger.Warn("Path rate limit exceeded", 
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Path rate limit exceeded",
				"code":  "PATH_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		// 如果用户已认证，检查用户限流
		if userID, exists := c.Get("user_id"); exists {
			if userIDStr, ok := userID.(string); ok {
				allowed, err = rateLimitService.AllowUser(userIDStr)
				if err != nil {
					logger.Error("User rate limit check failed", 
						zap.String("user_id", userIDStr),
						zap.Error(err))
				} else if !allowed {
					logger.Warn("User rate limit exceeded", 
						zap.String("user_id", userIDStr),
						zap.String("ip", c.ClientIP()),
						zap.String("path", c.Request.URL.Path))
					
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "User rate limit exceeded",
						"code":  "USER_RATE_LIMIT_EXCEEDED",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}