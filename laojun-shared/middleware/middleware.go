package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// RequestID 中间件 - 为每个请求生成唯一 ID
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

// CORS 中间件配置
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" env:"CORS_ALLOW_ORIGINS" config:"cors.allow_origins"`
	AllowMethods     []string `yaml:"allow_methods" env:"CORS_ALLOW_METHODS" config:"cors.allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers" env:"CORS_ALLOW_HEADERS" config:"cors.allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers" env:"CORS_EXPOSE_HEADERS" config:"cors.expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" env:"CORS_ALLOW_CREDENTIALS" config:"cors.allow_credentials" default:"false"`
	MaxAge           int      `yaml:"max_age" env:"CORS_MAX_AGE" config:"cors.max_age" default:"86400"`
}

// CORS 中间件 - 处理跨域请求
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该来源
		if len(config.AllowOrigins) > 0 {
			allowed := false
			for _, allowedOrigin := range config.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			if !allowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// 设置 CORS 响应头
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 限流中间件配置
type RateLimitConfig struct {
	Enabled bool    `yaml:"enabled" env:"RATE_LIMIT_ENABLED" config:"rate_limit.enabled" default:"true"`
	RPS     float64 `yaml:"rps" env:"RATE_LIMIT_RPS" config:"rate_limit.rps" default:"100"`
	Burst   int     `yaml:"burst" env:"RATE_LIMIT_BURST" config:"rate_limit.burst" default:"200"`
}

// RateLimit 限流中间件 - 限制每个 IP 的请求速率
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := rate.NewLimiter(rate.Limit(config.RPS), config.Burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// 恢复中间件 - 处理 panic
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 获取堆栈信息
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)

		// 记录错误
		fmt.Printf("[PANIC] %v\n%s\n", recovered, stack[:length])

		// 返回错误响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"message":    "An unexpected error occurred",
			"request_id": c.GetString("request_id"),
		})
	})
}

// 超时中间件 - 限制请求处理时间
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":      "Request timeout",
				"message":    "Request took too long to process",
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
		}
	}
}

// API 密钥认证中间件 - 验证请求中的 API 密钥
func APIKeyAuth(validKeys []string) gin.HandlerFunc {
	keySet := make(map[string]bool)
	for _, key := range validKeys {
		keySet[key] = true
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" || !keySet[apiKey] {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Set("api_key", apiKey)
		c.Next()
	}
}

// IP 白名单中间件
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	ipSet := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipSet[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if len(ipSet) > 0 && !ipSet[clientIP] {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "IP address not allowed",
			})
			c.Abort()
			return
		}

		c.Set("client_ip", clientIP)
		c.Next()
	}
}

// 请求日志中间件配置
type LoggerConfig struct {
	SkipPaths []string `yaml:"skip_paths" env:"LOGGER_SKIP_PATHS" config:"logger.skip_paths"`
	LogBody   bool     `yaml:"log_body" env:"LOGGER_LOG_BODY" config:"logger.log_body" default:"false"`
}

// Logger 日志中间件 - 记录请求日志
func Logger(config LoggerConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 跳过指定路径
		if skipPaths[param.Path] {
			return ""
		}

		// 构建日志格式
		return fmt.Sprintf("[%s] %s %s %d %s %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.ErrorMessage,
		)
	})
}

// 响应体包装器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// 请求响应记录中间件 - 记录请求和响应信息
func RequestResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器以记录响应体
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = w

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// 记录请求响应信息
		c.Set("request_body", string(requestBody))
		c.Set("response_body", w.body.String())
		c.Set("duration", duration)
	}
}

// 健康检查跳过中间件
func SkipHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}
		c.Next()
	}
}

// 版本中间件 - 添加服务版本头
func Version(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Service-Version", version)
		c.Set("service_version", version)
		c.Next()
	}
}

// 内容类型验证中间件 - 验证请求内容类型
func ContentType(allowedTypes ...string) gin.HandlerFunc {
	typeSet := make(map[string]bool)
	for _, t := range allowedTypes {
		typeSet[t] = true
	}

	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Bad Request",
					"message": "Content-Type header is required",
				})
				c.Abort()
				return
			}

			// 提取主要内容类型（忽略参数）
			mainType := strings.Split(contentType, ";")[0]
			mainType = strings.TrimSpace(mainType)

			if len(typeSet) > 0 && !typeSet[mainType] {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":   "Unsupported Media Type",
					"message": fmt.Sprintf("Content-Type %s is not supported", mainType),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// 请求大小限制中间件 - 限制请求体大小
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request Entity Too Large",
				"message": fmt.Sprintf("Request body too large, max size is %d bytes", maxSize),
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// 缓存控制中间件 - 添加缓存头
func CacheControl(maxAge int) gin.HandlerFunc {
	cacheHeader := fmt.Sprintf("public, max-age=%d", maxAge)

	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Header("Cache-Control", cacheHeader)
		}
		c.Next()
	}
}

// 压缩中间件（简单实现）
func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持压缩
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 设置压缩响应头
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

// 中间件链构建器 - 用于组合多个中间件
type MiddlewareChain struct {
	middlewares []gin.HandlerFunc
}

// NewMiddlewareChain 创建中间件链
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]gin.HandlerFunc, 0),
	}
}

// Use 添加中间件到链中
func (mc *MiddlewareChain) Use(middleware gin.HandlerFunc) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, middleware)
	return mc
}

// Build 构建中间件链
func (mc *MiddlewareChain) Build() []gin.HandlerFunc {
	return mc.middlewares
}

// Apply 应用到路由器
func (mc *MiddlewareChain) Apply(router *gin.Engine) {
	for _, middleware := range mc.middlewares {
		router.Use(middleware)
	}
}
