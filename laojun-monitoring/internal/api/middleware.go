package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Middleware 中间件接口
type Middleware interface {
	Handler(next http.Handler) http.Handler
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Handler 处理请求
func (lm *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 创建响应写入器包装器
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// 处理请求
		next.ServeHTTP(wrapped, r)
		
		// 记录日志
		duration := time.Since(start)
		lm.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Int("status_code", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.Int64("content_length", r.ContentLength),
		)
	})
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 写入状态码
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware CORS中间件
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
	maxAge         int
}

// NewCORSMiddleware 创建CORS中间件
func NewCORSMiddleware(origins, methods, headers []string, maxAge int) *CORSMiddleware {
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(headers) == 0 {
		headers = []string{"Content-Type", "Authorization", "X-Requested-With"}
	}
	if maxAge == 0 {
		maxAge = 86400 // 24小时
	}
	
	return &CORSMiddleware{
		allowedOrigins: origins,
		allowedMethods: methods,
		allowedHeaders: headers,
		maxAge:         maxAge,
	}
}

// Handler 处理请求
func (cm *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// 检查是否允许该来源
		if cm.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(cm.allowedOrigins) == 1 && cm.allowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(cm.allowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(cm.allowedHeaders, ", "))
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cm.maxAge))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed 检查来源是否被允许
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range cm.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	logger    *zap.Logger
	apiKeys   map[string]string // API密钥映射
	jwtSecret string            // JWT密钥
	enabled   bool              // 是否启用认证
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(logger *zap.Logger, apiKeys map[string]string, jwtSecret string, enabled bool) *AuthMiddleware {
	if apiKeys == nil {
		apiKeys = make(map[string]string)
	}
	
	return &AuthMiddleware{
		logger:    logger,
		apiKeys:   apiKeys,
		jwtSecret: jwtSecret,
		enabled:   enabled,
	}
}

// Handler 处理请求
func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果认证未启用，直接通过
		if !am.enabled {
			next.ServeHTTP(w, r)
			return
		}
		
		// 检查是否为公开端点
		if am.isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		
		// 验证认证
		if !am.authenticate(r) {
			am.writeUnauthorizedResponse(w)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// authenticate 验证认证
func (am *AuthMiddleware) authenticate(r *http.Request) bool {
	// 检查API密钥
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return am.validateAPIKey(apiKey)
	}
	
	// 检查Bearer Token
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			return am.validateJWT(token)
		}
	}
	
	// 检查基本认证
	if username, password, ok := r.BasicAuth(); ok {
		return am.validateBasicAuth(username, password)
	}
	
	return false
}

// validateAPIKey 验证API密钥
func (am *AuthMiddleware) validateAPIKey(apiKey string) bool {
	_, exists := am.apiKeys[apiKey]
	return exists
}

// validateJWT 验证JWT令牌
func (am *AuthMiddleware) validateJWT(token string) bool {
	// 这里应该实现JWT验证逻辑
	// 为了简化，这里只做基本检查
	return token != "" && am.jwtSecret != ""
}

// validateBasicAuth 验证基本认证
func (am *AuthMiddleware) validateBasicAuth(username, password string) bool {
	// 这里应该实现基本认证逻辑
	// 为了简化，这里只做基本检查
	return username != "" && password != ""
}

// isPublicEndpoint 检查是否为公开端点
func (am *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/ready",
		"/metrics", // Prometheus指标端点通常是公开的
	}
	
	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}
	
	return false
}

// writeUnauthorizedResponse 写入未授权响应
func (am *AuthMiddleware) writeUnauthorizedResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	response := `{"error": "Unauthorized", "timestamp": ` + fmt.Sprintf("%d", time.Now().Unix()) + `}`
	w.Write([]byte(response))
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	logger    *zap.Logger
	limiter   *RateLimiter
	enabled   bool
}

// RateLimiter 限流器
type RateLimiter struct {
	requests map[string]*ClientLimit
	maxRate  int           // 每秒最大请求数
	window   time.Duration // 时间窗口
}

// ClientLimit 客户端限制
type ClientLimit struct {
	count     int
	lastReset time.Time
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(logger *zap.Logger, maxRate int, window time.Duration, enabled bool) *RateLimitMiddleware {
	if window == 0 {
		window = time.Second
	}
	
	return &RateLimitMiddleware{
		logger:  logger,
		enabled: enabled,
		limiter: &RateLimiter{
			requests: make(map[string]*ClientLimit),
			maxRate:  maxRate,
			window:   window,
		},
	}
}

// Handler 处理请求
func (rlm *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rlm.enabled {
			next.ServeHTTP(w, r)
			return
		}
		
		clientIP := rlm.getClientIP(r)
		
		if !rlm.limiter.Allow(clientIP) {
			rlm.writeTooManyRequestsResponse(w)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientIP string) bool {
	now := time.Now()
	
	limit, exists := rl.requests[clientIP]
	if !exists {
		rl.requests[clientIP] = &ClientLimit{
			count:     1,
			lastReset: now,
		}
		return true
	}
	
	// 检查是否需要重置计数器
	if now.Sub(limit.lastReset) >= rl.window {
		limit.count = 1
		limit.lastReset = now
		return true
	}
	
	// 检查是否超过限制
	if limit.count >= rl.maxRate {
		return false
	}
	
	limit.count++
	return true
}

// getClientIP 获取客户端IP
func (rlm *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 使用RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// writeTooManyRequestsResponse 写入请求过多响应
func (rlm *RateLimitMiddleware) writeTooManyRequestsResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	
	response := `{"error": "Too Many Requests", "timestamp": ` + fmt.Sprintf("%d", time.Now().Unix()) + `}`
	w.Write([]byte(response))
}

// RecoveryMiddleware 恢复中间件
type RecoveryMiddleware struct {
	logger *zap.Logger
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{logger: logger}
}

// Handler 处理请求
func (rm *RecoveryMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rm.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr),
				)
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				
				response := `{"error": "Internal Server Error", "timestamp": ` + fmt.Sprintf("%d", time.Now().Unix()) + `}`
				w.Write([]byte(response))
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware 请求ID中间件
type RequestIDMiddleware struct {
	header string
}

// NewRequestIDMiddleware 创建请求ID中间件
func NewRequestIDMiddleware(header string) *RequestIDMiddleware {
	if header == "" {
		header = "X-Request-ID"
	}
	
	return &RequestIDMiddleware{header: header}
}

// Handler 处理请求
func (rim *RequestIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(rim.header)
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// 设置响应头
		w.Header().Set(rim.header, requestID)
		
		// 将请求ID添加到上下文
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 简单的请求ID生成，实际应用中可以使用UUID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []Middleware
}

// NewMiddlewareChain 创建中间件链
func NewMiddlewareChain(middlewares ...Middleware) *MiddlewareChain {
	return &MiddlewareChain{middlewares: middlewares}
}

// Then 应用中间件链
func (mc *MiddlewareChain) Then(handler http.Handler) http.Handler {
	// 从后往前应用中间件
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		handler = mc.middlewares[i].Handler(handler)
	}
	return handler
}

// SetupMiddlewares 设置中间件
func SetupMiddlewares(router *mux.Router, logger *zap.Logger, config MiddlewareConfig) {
	// 恢复中间件（最外层）
	recovery := NewRecoveryMiddleware(logger)
	router.Use(recovery.Handler)
	
	// 请求ID中间件
	requestID := NewRequestIDMiddleware(config.RequestIDHeader)
	router.Use(requestID.Handler)
	
	// 日志中间件
	logging := NewLoggingMiddleware(logger)
	router.Use(logging.Handler)
	
	// CORS中间件
	if config.CORS.Enabled {
		cors := NewCORSMiddleware(
			config.CORS.AllowedOrigins,
			config.CORS.AllowedMethods,
			config.CORS.AllowedHeaders,
			config.CORS.MaxAge,
		)
		router.Use(cors.Handler)
	}
	
	// 限流中间件
	if config.RateLimit.Enabled {
		rateLimit := NewRateLimitMiddleware(
			logger,
			config.RateLimit.MaxRate,
			config.RateLimit.Window,
			true,
		)
		router.Use(rateLimit.Handler)
	}
	
	// 认证中间件
	if config.Auth.Enabled {
		auth := NewAuthMiddleware(
			logger,
			config.Auth.APIKeys,
			config.Auth.JWTSecret,
			true,
		)
		router.Use(auth.Handler)
	}
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	RequestIDHeader string           `mapstructure:"request_id_header"`
	CORS            CORSConfig       `mapstructure:"cors"`
	RateLimit       RateLimitConfig  `mapstructure:"rate_limit"`
	Auth            AuthConfig       `mapstructure:"auth"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	MaxAge         int      `mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	MaxRate int           `mapstructure:"max_rate"`
	Window  time.Duration `mapstructure:"window"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled   bool              `mapstructure:"enabled"`
	APIKeys   map[string]string `mapstructure:"api_keys"`
	JWTSecret string            `mapstructure:"jwt_secret"`
}