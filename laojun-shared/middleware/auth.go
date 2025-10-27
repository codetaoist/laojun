package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/codetaoist/laojun-shared/models"
)

// AuthMiddleware 增强的认证中间件
type AuthMiddleware struct {
	db          *gorm.DB
	redisClient *redis.Client
	jwtSecret   string
	tokenExpiry time.Duration
}

// Claims JWT 声明结构
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret   string
	TokenExpiry time.Duration
	RedisPrefix string
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(db *gorm.DB, redisClient *redis.Client, config AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		db:          db,
		redisClient: redisClient,
		jwtSecret:   config.JWTSecret,
		tokenExpiry: config.TokenExpiry,
	}
}

// GenerateToken 生成 JWT token
func (a *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
	// 获取用户的主要角色
	primaryRole := a.getUserPrimaryRole(user.ID)
	
	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     primaryRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "laojun",
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", err
	}

	// 存储 token 到 Redis
	ctx := context.Background()
	key := fmt.Sprintf("auth:token:%s", claims.ID)
	err = a.redisClient.Set(ctx, key, tokenString, a.tokenExpiry).Err()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证 JWT token
func (a *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查 token 是否在 Redis 中（未被撤销）
		ctx := context.Background()
		key := fmt.Sprintf("auth:token:%s", claims.ID)
		exists := a.redisClient.Exists(ctx, key).Val()
		if exists == 0 {
			return nil, fmt.Errorf("token has been revoked")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RevokeToken 撤销 token
func (a *AuthMiddleware) RevokeToken(tokenID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("auth:token:%s", tokenID)
	return a.redisClient.Del(ctx, key).Err()
}

// RevokeUserTokens 撤销用户的所有 token
func (a *AuthMiddleware) RevokeUserTokens(userID string) error {
	ctx := context.Background()
	pattern := fmt.Sprintf("auth:token:*")
	keys, err := a.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		tokenString, err := a.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			continue
		}

		if claims.UserID == userID {
			a.redisClient.Del(ctx, key)
		}
	}

	return nil
}

// RequireAuth 需要认证的中间件
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing or invalid authorization token",
			})
			c.Abort()
			return
		}

		claims, err := a.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (a *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "No authentication information found",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authentication information",
			})
			c.Abort()
			return
		}

		if !a.hasPermission(userClaims.Role, roles) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 需要特定权限的中间件
func (a *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "No authentication information found",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authentication information",
			})
			c.Abort()
			return
		}

		// 检查用户权限
		var user models.User
		if err := a.db.Where("id = ?", userClaims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User not found",
			})
			c.Abort()
			return
		}

		// 获取用户的主要角色进行权限检查
		userID, _ := uuid.Parse(userClaims.UserID)
		primaryRole := a.getUserPrimaryRole(userID)
		
		// 这里可以实现更复杂的权限检查逻辑
		// 目前简化为基于角色的权限检查
		if !a.hasRolePermission(primaryRole, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractToken(c)
		if token != "" {
			claims, err := a.ValidateToken(token)
			if err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("email", claims.Email)
				c.Set("role", claims.Role)
				c.Set("claims", claims)
			}
		}
		c.Next()
	}
}

// extractToken 从请求中提取 token
func (a *AuthMiddleware) extractToken(c *gin.Context) string {
	// 从 Authorization header 中提取 token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从查询参数中提取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从 cookie 中提取 token
	cookie, err := c.Cookie("auth_token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// hasPermission 检查角色是否有权限
func (a *AuthMiddleware) hasPermission(userRole string, allowedRoles []string) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// hasRolePermission 检查角色是否有特定权限
func (a *AuthMiddleware) hasRolePermission(role, permission string) bool {
	// 定义角色权限映射
	rolePermissions := map[string][]string{
		"admin": {
			"user:create", "user:read", "user:update", "user:delete",
			"config:create", "config:read", "config:update", "config:delete",
			"plugin:create", "plugin:read", "plugin:update", "plugin:delete",
			"system:read", "system:update", "audit:read",
		},
		"moderator": {
			"user:read", "user:update",
			"config:read", "config:update",
			"plugin:read", "plugin:update",
			"system:read",
		},
		"developer": {
			"config:read", "config:create", "config:update",
			"plugin:create", "plugin:read", "plugin:update",
		},
		"user": {
			"config:read", "plugin:read",
		},
	}

	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// RefreshToken 刷新 token
func (a *AuthMiddleware) RefreshToken(oldToken string) (string, error) {
	claims, err := a.ValidateToken(oldToken)
	if err != nil {
		return "", err
	}

	// 撤销旧 token
	a.RevokeToken(claims.ID)

	// 获取用户信息
	var user models.User
	if err := a.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return "", err
	}

	// 生成新 token
	return a.GenerateToken(&user)
}

// GetCurrentUser 获取当前用户
func (a *AuthMiddleware) GetCurrentUser(c *gin.Context) (*models.User, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	var user models.User
	if err := a.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetCurrentUserID 获取当前用户 ID
func (a *AuthMiddleware) GetCurrentUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", fmt.Errorf("user not authenticated")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID format")
	}

	return userIDStr, nil
}

// IsAuthenticated 检查是否已认证
func (a *AuthMiddleware) IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// HasRole 检查是否有指定角色
func (a *AuthMiddleware) HasRole(c *gin.Context, role string) bool {
	userRole, exists := c.Get("role")
	if !exists {
		return false
	}

	userRoleStr, ok := userRole.(string)
	if !ok {
		return false
	}

	return userRoleStr == role
}

// HasAnyRole 检查是否有任意指定角色
func (a *AuthMiddleware) HasAnyRole(c *gin.Context, roles ...string) bool {
	userRole, exists := c.Get("role")
	if !exists {
		return false
	}

	userRoleStr, ok := userRole.(string)
	if !ok {
		return false
	}

	for _, role := range roles {
		if userRoleStr == role {
			return true
		}
	}

	return false
}

// getUserPrimaryRole 获取用户的主要角色
func (a *AuthMiddleware) getUserPrimaryRole(userID uuid.UUID) string {
	var roleName string
	query := `
		SELECT r.name 
		FROM az_roles r
		INNER JOIN az_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.is_system DESC, r.created_at ASC
		LIMIT 1
	`
	
	err := a.db.Raw(query, userID).Scan(&roleName).Error
	if err != nil {
		// 如果查询失败或没有角色，返回默认角色
		return "user"
	}
	
	if roleName == "" {
		return "user"
	}
	
	return roleName
}
