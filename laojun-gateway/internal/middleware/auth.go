package middleware

import (
	"net/http"
	"strings"

	"github.com/codetaoist/laojun-gateway/internal/auth"
	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(cfg config.AuthConfig, logger *zap.Logger) gin.HandlerFunc {
	authService := auth.NewService(cfg, logger)

	return func(c *gin.Context) {
		// 检查是否在白名单中
		if isWhitelisted(c.Request.URL.Path, cfg.WhiteList) {
			c.Next()
			return
		}

		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header", 
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()))
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// 检查Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("Invalid authorization header format", 
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()))
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			logger.Warn("Token validation failed", 
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.Error(err))
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		logger.Debug("User authenticated", 
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}

// isWhitelisted 检查路径是否在白名单中
func isWhitelisted(path string, whitelist []string) bool {
	for _, pattern := range whitelist {
		if matchPattern(path, pattern) {
			return true
		}
	}
	return false
}

// matchPattern 匹配路径模式
func matchPattern(path, pattern string) bool {
	// 简单的通配符匹配
	if pattern == "*" {
		return true
	}
	
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	
	return path == pattern
}