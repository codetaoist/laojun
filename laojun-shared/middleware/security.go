package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// CORS 配置
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration

	// CSRF 配置
	CSRFSecret     string
	CSRFTokenName  string
	CSRFCookieName string

	// 安全头配置
	ContentSecurityPolicy string
	XFrameOptions         string
	XContentTypeOptions   string
	ReferrerPolicy        string
	PermissionsPolicy     string
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,

		CSRFSecret:     "your-csrf-secret-key",
		CSRFTokenName:  "csrf_token",
		CSRFCookieName: "_csrf",

		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'self';",
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "camera=(), microphone=(), geolocation=()",
	}
}

// CSRFProtection CSRF 保护中间件
func CSRFProtection(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于安全 HTTP 方法，不需要 CSRF 保护
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 获取 CSRF token
		token := getCSRFToken(c, config)
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "csrf_token_missing",
				"message": "CSRF token is missing",
			})
			c.Abort()
			return
		}

		// 验证 CSRF token
		if !validateCSRFToken(token, config.CSRFSecret) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "csrf_token_invalid",
				"message": "CSRF token is invalid",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken 生成 CSRF token
func GenerateCSRFToken(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := generateRandomToken(32)

		// 设置 CSRF token 到 Cookie
		c.SetCookie(config.CSRFCookieName, token, 3600, "/", "", false, true)

		// 设置 CSRF token 到响应头
		c.Header("X-CSRF-Token", token)

		// 设置 CSRF token 到上下文
		c.Set("csrf_token", token)

		c.Next()
	}
}

// UserAgent 用户代理检查中间件
func UserAgentFilter(blockedUserAgents []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")

		for _, blocked := range blockedUserAgents {
			if strings.Contains(strings.ToLower(userAgent), strings.ToLower(blocked)) {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "user_agent_blocked",
					"message": "Your user agent is not allowed",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ContentTypeValidation 内容类型验证中间件
func ContentTypeValidation(allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")

			if contentType == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "content_type_missing",
					"message": "Content-Type header is required",
				})
				c.Abort()
				return
			}

			// 检查内容类型是否被允许
			allowed := false
			for _, allowedType := range allowedTypes {
				if strings.HasPrefix(contentType, allowedType) {
					allowed = true
					break
				}
			}

			if !allowed {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":   "content_type_not_allowed",
					"message": "Content-Type not allowed",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// NoCache 禁用缓存中间件
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
		c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Next()
	}
}

// 辅助函数

// isOriginAllowed 检查来源是否被允许
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}

		// 支持通配符匹配
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// getCSRFToken 获取 CSRF token
func getCSRFToken(c *gin.Context, config SecurityConfig) string {
	// 从头部获取 token
	token := c.GetHeader("X-CSRF-Token")
	if token != "" {
		return token
	}

	// 从表单获取 token
	token = c.PostForm(config.CSRFTokenName)
	if token != "" {
		return token
	}

	// 从 cookie 获取 token
	cookie, err := c.Cookie(config.CSRFCookieName)
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// validateCSRFToken 验证 CSRF token
func validateCSRFToken(token, secret string) bool {
	// 这里应该实现更复杂的 CSRF token 验证逻辑
	// 目前简化为检查 token 是否为空且长度足够长
	return token != "" && len(token) >= 16
}

// generateRandomToken 生成随机 token
func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// isIPAllowed 检查 IP 是否被允许
func isIPAllowed(clientIP string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		return true // 如果没有设置白名单，则允许所有 IP
	}

	for _, allowedIP := range allowedIPs {
		if allowedIP == clientIP {
			return true
		}

		// 支持 CIDR 格式的 IP 范围检查
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err == nil && ipNet.Contains(net.ParseIP(clientIP)) {
				return true
			}
		}

		// 这里可以添加更复杂的 IP 范围检查逻辑
	}

	return false
}
