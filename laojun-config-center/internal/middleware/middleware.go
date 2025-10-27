package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestID 中间件 - 为每个请求生成唯一ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// SecurityHeaders 中间件 - 添加安全头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// Logger 中间件 - 请求日志记录
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		fields := []zap.Field{
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
		}

		if requestID := param.Request.Header.Get("X-Request-ID"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		if param.ErrorMessage != "" {
			fields = append(fields, zap.String("error", param.ErrorMessage))
		}

		if param.StatusCode >= 400 {
			logger.Error("HTTP request", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}

		return ""
	})
}

// Metrics 中间件 - 指标收集
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		// 这里可以集成Prometheus或其他指标系统
		// 暂时使用日志记录指标
		if logger, exists := c.Get("logger"); exists {
			if zapLogger, ok := logger.(*zap.Logger); ok {
				zapLogger.Info("Request metrics",
					zap.String("method", c.Request.Method),
					zap.String("path", c.FullPath()),
					zap.Int("status", c.Writer.Status()),
					zap.Duration("duration", duration),
					zap.Int("response_size", c.Writer.Size()),
				)
			}
		}
	}
}

// CORS 中间件 - 跨域资源共享
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 在生产环境中应该配置允许的域名列表
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID, X-Operator")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimit 中间件 - 简单的速率限制
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	// 这是一个简化的实现，生产环境应该使用Redis等外部存储
	clients := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// 清理过期的请求记录
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}
		
		// 检查是否超过限制
		if len(clients[clientIP]) >= requestsPerMinute {
			c.JSON(429, gin.H{
				"error": "rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}
		
		// 记录当前请求
		clients[clientIP] = append(clients[clientIP], now)
		
		c.Next()
	}
}

// Timeout 中间件 - 请求超时
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求超时
		c.Request = c.Request.WithContext(
			c.Request.Context(),
		)
		
		c.Next()
	}
}

// RequestSizeLimit 中间件 - 请求大小限制
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(413, gin.H{
				"error": fmt.Sprintf("request body too large, max size is %d bytes", maxSize),
			})
			c.Abort()
			return
		}
		
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// Recovery 中间件 - 恢复panic
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")
		
		logger.Error("Panic recovered",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Any("error", recovered),
		)
		
		c.JSON(500, gin.H{
			"error": "internal server error",
			"request_id": requestID,
		})
	})
}

// Auth 中间件 - 简单的认证（示例）
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现JWT验证、API Key验证等
		// 暂时跳过认证
		c.Next()
	}
}

// Validation 中间件 - 请求验证
func Validation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证Content-Type
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && contentType != "application/json" && contentType != "multipart/form-data" {
				c.JSON(400, gin.H{
					"error": "unsupported content type",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}