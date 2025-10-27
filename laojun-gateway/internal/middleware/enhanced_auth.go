package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/auth"
	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnhancedAuthMiddleware 增强认证中间件
type EnhancedAuthMiddleware struct {
	authService    *auth.Service
	config         config.AuthConfig
	logger         *zap.Logger
	rateLimiter    map[string]*UserRateLimit
	sessionStore   map[string]*SessionInfo
}

// UserRateLimit 用户限流信息
type UserRateLimit struct {
	Count     int
	ResetTime time.Time
}

// SessionInfo 会话信息
type SessionInfo struct {
	UserID    string
	Username  string
	Roles     []string
	LoginTime time.Time
	LastSeen  time.Time
	IP        string
}

// NewEnhancedAuthMiddleware 创建增强认证中间件
func NewEnhancedAuthMiddleware(cfg config.AuthConfig, logger *zap.Logger) *EnhancedAuthMiddleware {
	authService := auth.NewService(cfg, logger)
	
	return &EnhancedAuthMiddleware{
		authService:  authService,
		config:       cfg,
		logger:       logger,
		rateLimiter:  make(map[string]*UserRateLimit),
		sessionStore: make(map[string]*SessionInfo),
	}
}

// AuthMiddleware 认证中间件
func (eam *EnhancedAuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否在白名单中
		if eam.isWhitelisted(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 获取认证信息
		authInfo, err := eam.extractAuthInfo(c)
		if err != nil {
			eam.handleAuthError(c, err, "AUTH_EXTRACTION_FAILED")
			return
		}

		// 验证认证信息
		claims, err := eam.validateAuth(authInfo)
		if err != nil {
			eam.handleAuthError(c, err, "AUTH_VALIDATION_FAILED")
			return
		}

		// 检查用户限流
		if !eam.checkUserRateLimit(claims.UserID) {
			eam.handleAuthError(c, fmt.Errorf("user rate limit exceeded"), "USER_RATE_LIMIT_EXCEEDED")
			return
		}

		// 检查权限
		if !eam.checkPermissions(c.Request.URL.Path, c.Request.Method, claims.Roles) {
			eam.handleAuthError(c, fmt.Errorf("insufficient permissions"), "INSUFFICIENT_PERMISSIONS")
			return
		}

		// 更新会话信息
		eam.updateSession(claims, c.ClientIP())

		// 设置用户上下文
		eam.setUserContext(c, claims)

		eam.logger.Debug("User authenticated successfully",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method))

		c.Next()
	}
}

// RoleBasedAuthMiddleware 基于角色的认证中间件
func (eam *EnhancedAuthMiddleware) RoleBasedAuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行基础认证
		eam.AuthMiddleware()(c)
		if c.IsAborted() {
			return
		}

		// 检查角色权限
		userRoles, exists := c.Get("roles")
		if !exists {
			eam.handleAuthError(c, fmt.Errorf("user roles not found"), "ROLES_NOT_FOUND")
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			eam.handleAuthError(c, fmt.Errorf("invalid roles format"), "INVALID_ROLES_FORMAT")
			return
		}

		if !eam.hasRequiredRole(roles, requiredRoles) {
			eam.handleAuthError(c, fmt.Errorf("insufficient role permissions"), "INSUFFICIENT_ROLE_PERMISSIONS")
			return
		}

		c.Next()
	}
}

// APIKeyAuthMiddleware API密钥认证中间件
func (eam *EnhancedAuthMiddleware) APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header或Query参数获取API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			eam.handleAuthError(c, fmt.Errorf("missing API key"), "MISSING_API_KEY")
			return
		}

		// 验证API Key
		keyInfo, err := eam.authService.ValidateAPIKey(apiKey)
		if err != nil {
			eam.handleAuthError(c, err, "INVALID_API_KEY")
			return
		}

		// 检查API Key限流
		if !eam.checkAPIKeyRateLimit(apiKey) {
			eam.handleAuthError(c, fmt.Errorf("API key rate limit exceeded"), "API_KEY_RATE_LIMIT_EXCEEDED")
			return
		}

		// 设置API Key上下文
		c.Set("api_key", apiKey)
		c.Set("api_key_info", keyInfo)

		eam.logger.Debug("API key authenticated",
			zap.String("api_key", apiKey[:8]+"..."),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// JWTRefreshMiddleware JWT刷新中间件
func (eam *EnhancedAuthMiddleware) JWTRefreshMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken := c.GetHeader("X-Refresh-Token")
		if refreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing refresh token",
				"code":  "MISSING_REFRESH_TOKEN",
			})
			return
		}

		// 验证刷新令牌
		claims, err := eam.authService.ValidateRefreshToken(refreshToken)
		if err != nil {
			eam.handleAuthError(c, err, "INVALID_REFRESH_TOKEN")
			return
		}

		// 生成新的访问令牌
		newToken, err := eam.authService.GenerateToken(claims.UserID, claims.Username, claims.Roles)
		if err != nil {
			eam.logger.Error("Failed to generate new token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate new token",
				"code":  "TOKEN_GENERATION_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": newToken,
			"token_type":   "Bearer",
			"expires_in":   eam.config.TokenExpiry,
		})
	}
}

// extractAuthInfo 提取认证信息
func (eam *EnhancedAuthMiddleware) extractAuthInfo(c *gin.Context) (string, error) {
	// 优先从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
		return "", fmt.Errorf("invalid authorization header format")
	}

	// 从Cookie获取
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie, nil
	}

	// 从Query参数获取
	if token := c.Query("token"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("missing authentication token")
}

// validateAuth 验证认证信息
func (eam *EnhancedAuthMiddleware) validateAuth(token string) (*auth.Claims, error) {
	return eam.authService.ValidateToken(token)
}

// checkUserRateLimit 检查用户限流
func (eam *EnhancedAuthMiddleware) checkUserRateLimit(userID string) bool {
	now := time.Now()
	
	// 获取用户限流信息
	userLimit, exists := eam.rateLimiter[userID]
	if !exists {
		eam.rateLimiter[userID] = &UserRateLimit{
			Count:     1,
			ResetTime: now.Add(time.Minute),
		}
		return true
	}

	// 检查是否需要重置
	if now.After(userLimit.ResetTime) {
		userLimit.Count = 1
		userLimit.ResetTime = now.Add(time.Minute)
		return true
	}

	// 检查是否超过限制
	if userLimit.Count >= eam.config.UserRateLimit {
		return false
	}

	userLimit.Count++
	return true
}

// checkAPIKeyRateLimit 检查API Key限流
func (eam *EnhancedAuthMiddleware) checkAPIKeyRateLimit(apiKey string) bool {
	// 实现API Key限流逻辑
	// 这里可以使用Redis或内存存储
	return true
}

// checkPermissions 检查权限
func (eam *EnhancedAuthMiddleware) checkPermissions(path, method string, roles []string) bool {
	// 检查路径权限配置
	for _, permission := range eam.config.Permissions {
		if eam.matchPathPattern(path, permission.Path) && 
		   eam.matchMethod(method, permission.Methods) {
			return eam.hasRequiredRole(roles, permission.Roles)
		}
	}
	
	// 默认允许
	return true
}

// updateSession 更新会话信息
func (eam *EnhancedAuthMiddleware) updateSession(claims *auth.Claims, ip string) {
	sessionKey := fmt.Sprintf("%s:%s", claims.UserID, ip)
	
	session, exists := eam.sessionStore[sessionKey]
	if !exists {
		session = &SessionInfo{
			UserID:    claims.UserID,
			Username:  claims.Username,
			Roles:     claims.Roles,
			LoginTime: time.Now(),
			IP:        ip,
		}
		eam.sessionStore[sessionKey] = session
	}
	
	session.LastSeen = time.Now()
}

// setUserContext 设置用户上下文
func (eam *EnhancedAuthMiddleware) setUserContext(c *gin.Context, claims *auth.Claims) {
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("roles", claims.Roles)
	c.Set("auth_time", time.Now())
}

// handleAuthError 处理认证错误
func (eam *EnhancedAuthMiddleware) handleAuthError(c *gin.Context, err error, code string) {
	eam.logger.Warn("Authentication failed",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
		zap.String("code", code),
		zap.Error(err))

	statusCode := http.StatusUnauthorized
	if code == "INSUFFICIENT_PERMISSIONS" || code == "INSUFFICIENT_ROLE_PERMISSIONS" {
		statusCode = http.StatusForbidden
	}

	c.JSON(statusCode, gin.H{
		"error":     err.Error(),
		"code":      code,
		"timestamp": time.Now().Unix(),
	})
	c.Abort()
}

// isWhitelisted 检查是否在白名单中
func (eam *EnhancedAuthMiddleware) isWhitelisted(path string) bool {
	for _, pattern := range eam.config.WhiteList {
		if eam.matchPathPattern(path, pattern) {
			return true
		}
	}
	return false
}

// matchPathPattern 匹配路径模式
func (eam *EnhancedAuthMiddleware) matchPathPattern(path, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

// matchMethod 匹配HTTP方法
func (eam *EnhancedAuthMiddleware) matchMethod(method string, methods []string) bool {
	if len(methods) == 0 {
		return true
	}
	for _, m := range methods {
		if m == "*" || m == method {
			return true
		}
	}
	return false
}

// hasRequiredRole 检查是否有必需的角色
func (eam *EnhancedAuthMiddleware) hasRequiredRole(userRoles, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}
	
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if userRole == required || userRole == "admin" {
				return true
			}
		}
	}
	return false
}

// GetActiveSessions 获取活跃会话
func (eam *EnhancedAuthMiddleware) GetActiveSessions() map[string]*SessionInfo {
	// 清理过期会话
	now := time.Now()
	for key, session := range eam.sessionStore {
		if now.Sub(session.LastSeen) > time.Hour*24 {
			delete(eam.sessionStore, key)
		}
	}
	
	return eam.sessionStore
}

// RevokeSession 撤销会话
func (eam *EnhancedAuthMiddleware) RevokeSession(userID, ip string) {
	sessionKey := fmt.Sprintf("%s:%s", userID, ip)
	delete(eam.sessionStore, sessionKey)
}