package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig 日志配置
type LoggingConfig struct {
	IncludeRequestBody  bool
	IncludeResponseBody bool
	MaxBodySize         int64
	SkipPaths          []string
}

// DefaultLoggingConfig 默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		IncludeRequestBody:  false,
		IncludeResponseBody: false,
		MaxBodySize:         1024 * 1024, // 1MB
		SkipPaths:          []string{"/health", "/metrics", "/favicon.ico"},
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
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

// StructuredLoggingMiddleware 结构化日志中间件
func StructuredLoggingMiddleware(logger *zap.Logger, config *LoggingConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return func(c *gin.Context) {
		// 检查是否跳过日志记录
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// 获取请求ID
		requestID, _ := c.Get("request_id")
		
		// 读取请求体
		var requestBody []byte
		if config.IncludeRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, config.MaxBodySize))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器包装器
		var responseBody *bytes.Buffer
		if config.IncludeResponseBody {
			responseBody = &bytes.Buffer{}
			c.Writer = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:          responseBody,
			}
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		
		// 构建日志字段
		fields := []zapcore.Field{
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
		}

		// 添加用户信息（如果存在）
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// 添加请求体（如果配置启用）
		if config.IncludeRequestBody && len(requestBody) > 0 {
			if isJSON(requestBody) {
				var jsonBody interface{}
				if err := json.Unmarshal(requestBody, &jsonBody); err == nil {
					fields = append(fields, zap.Any("request_body", jsonBody))
				} else {
					fields = append(fields, zap.String("request_body", string(requestBody)))
				}
			} else {
				fields = append(fields, zap.String("request_body", string(requestBody)))
			}
		}

		// 添加响应体（如果配置启用）
		if config.IncludeResponseBody && responseBody != nil && responseBody.Len() > 0 {
			responseBytes := responseBody.Bytes()
			if isJSON(responseBytes) {
				var jsonBody interface{}
				if err := json.Unmarshal(responseBytes, &jsonBody); err == nil {
					fields = append(fields, zap.Any("response_body", jsonBody))
				} else {
					fields = append(fields, zap.String("response_body", responseBody.String()))
				}
			} else {
				fields = append(fields, zap.String("response_body", responseBody.String()))
			}
		}

		// 添加错误信息（如果存在）
		if len(c.Errors) > 0 {
			fields = append(fields, zap.Any("errors", c.Errors.Errors()))
		}

		// 根据状态码选择日志级别
		switch {
		case c.Writer.Status() >= 500:
			logger.Error("HTTP Request", fields...)
		case c.Writer.Status() >= 400:
			logger.Warn("HTTP Request", fields...)
		default:
			logger.Info("HTTP Request", fields...)
		}
	}
}

// responseBodyWriter 响应体写入器包装器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// isJSON 检查数据是否为JSON格式
func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

// ErrorLoggingMiddleware 错误日志中间件
func ErrorLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 记录错误
		for _, err := range c.Errors {
			requestID, _ := c.Get("request_id")
			userID, _ := c.Get("user_id")

			fields := []zapcore.Field{
				zap.String("request_id", requestID.(string)),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.Error(err.Err),
				zap.String("error_type", err.Type.String()),
			}

			if userID != nil {
				fields = append(fields, zap.Any("user_id", userID))
			}

			switch err.Type {
			case gin.ErrorTypePublic:
				logger.Warn("Public Error", fields...)
			case gin.ErrorTypeBind:
				logger.Warn("Bind Error", fields...)
			case gin.ErrorTypeRender:
				logger.Error("Render Error", fields...)
			default:
				logger.Error("Internal Error", fields...)
			}
		}
	}
}

// SecurityLoggingMiddleware 安全日志中间件
func SecurityLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录可疑活动
		suspicious := false
		reasons := []string{}

		// 检查异常请求头
		if c.GetHeader("X-Forwarded-For") != "" && c.GetHeader("X-Real-IP") != "" {
			suspicious = true
			reasons = append(reasons, "multiple_proxy_headers")
		}

		// 检查异常User-Agent
		userAgent := c.Request.UserAgent()
		if userAgent == "" {
			suspicious = true
			reasons = append(reasons, "empty_user_agent")
		}

		// 检查SQL注入尝试
		query := c.Request.URL.RawQuery
		if containsSQLInjectionPatterns(query) {
			suspicious = true
			reasons = append(reasons, "sql_injection_attempt")
		}

		// 记录可疑活动
		if suspicious {
			requestID, _ := c.Get("request_id")
			logger.Warn("Suspicious Activity Detected",
				zap.String("request_id", requestID.(string)),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", userAgent),
				zap.Strings("reasons", reasons),
			)
		}

		c.Next()
	}
}

// containsSQLInjectionPatterns 检查是否包含SQL注入模式
func containsSQLInjectionPatterns(query string) bool {
	patterns := []string{
		"union", "select", "insert", "update", "delete", "drop",
		"exec", "execute", "sp_", "xp_", "--", "/*", "*/",
		"'", "\"", ";", "||", "&&",
	}

	queryLower := bytes.ToLower([]byte(query))
	for _, pattern := range patterns {
		if bytes.Contains(queryLower, []byte(pattern)) {
			return true
		}
	}
	return false
}

// AuditLoggingMiddleware 审计日志中间件
func AuditLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录需要审计的操作
		if !needsAudit(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		
		requestID, _ := c.Get("request_id")
		userID, _ := c.Get("user_id")

		logger.Info("Audit Log",
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.Any("user_id", userID),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("action", getActionFromPath(c.Request.URL.Path)),
		)
	}
}

// needsAudit 判断是否需要审计
func needsAudit(method, path string) bool {
	// 需要审计的操作
	auditPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
		"/api/v1/users",
		"/api/v1/plugins",
		"/api/v1/payments",
		"/api/v1/admin",
	}

	for _, auditPath := range auditPaths {
		if bytes.HasPrefix([]byte(path), []byte(auditPath)) {
			return true
		}
	}

	return method == "POST" || method == "PUT" || method == "DELETE"
}

// getActionFromPath 从路径获取操作类型
func getActionFromPath(path string) string {
	if bytes.Contains([]byte(path), []byte("/login")) {
		return "login"
	}
	if bytes.Contains([]byte(path), []byte("/logout")) {
		return "logout"
	}
	if bytes.Contains([]byte(path), []byte("/plugins")) {
		return "plugin_operation"
	}
	if bytes.Contains([]byte(path), []byte("/payments")) {
		return "payment_operation"
	}
	if bytes.Contains([]byte(path), []byte("/admin")) {
		return "admin_operation"
	}
	return "unknown"
}