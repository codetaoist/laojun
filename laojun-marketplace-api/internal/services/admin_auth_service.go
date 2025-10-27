package services

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-shared/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AdminAuthService 管理员认证服务
type AdminAuthService struct {
	db          *gorm.DB
	redisClient *redis.Client
	jwtSecret   string
	tokenExpiry time.Duration
}

// NewAdminAuthService 创建管理员认证服务
func NewAdminAuthService(db *gorm.DB, redisClient *redis.Client, jwtSecret string, tokenExpiry time.Duration) *AdminAuthService {
	return &AdminAuthService{
		db:          db,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// AdminClaims 管理员JWT声明结构
type AdminClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// GenerateAdminToken 生成管理员JWT令牌
func (s *AdminAuthService) GenerateAdminToken(user *models.User) (string, error) {
	// 检查用户是否有管理员权限
	if !s.isAdmin(user.ID) {
		return "", fmt.Errorf("user is not an admin")
	}

	// 获取用户的权限列表
	permissions := s.getUserPermissions(user.ID)
	
	claims := AdminClaims{
		UserID:      user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		Role:        "admin",
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "laojun-marketplace-admin",
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	// 存储token到Redis
	ctx := context.Background()
	key := fmt.Sprintf("admin:token:%s", claims.ID)
	err = s.redisClient.Set(ctx, key, tokenString, s.tokenExpiry).Err()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateAdminToken 验证管理员JWT令牌
func (s *AdminAuthService) ValidateAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		// 检查token是否在Redis中存在
		ctx := context.Background()
		key := fmt.Sprintf("admin:token:%s", claims.ID)
		exists := s.redisClient.Exists(ctx, key).Val()
		if exists == 0 {
			return nil, fmt.Errorf("admin token has been revoked")
		}

		// 再次验证用户是否仍然是管理员
		userID, _ := uuid.Parse(claims.UserID)
		if !s.isAdmin(userID) {
			return nil, fmt.Errorf("user is no longer an admin")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid admin token")
}

// RevokeAdminToken 撤销管理员令牌
func (s *AdminAuthService) RevokeAdminToken(tokenID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("admin:token:%s", tokenID)
	return s.redisClient.Del(ctx, key).Err()
}

// RefreshAdminToken 刷新管理员令牌
func (s *AdminAuthService) RefreshAdminToken(tokenString string) (string, error) {
	claims, err := s.ValidateAdminToken(tokenString)
	if err != nil {
		return "", err
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return "", err
	}

	// 撤销旧token
	s.RevokeAdminToken(claims.ID)

	// 生成新token
	return s.GenerateAdminToken(&user)
}

// isAdmin 检查用户是否是管理员
func (s *AdminAuthService) isAdmin(userID uuid.UUID) bool {
	var count int64
	query := `
		SELECT COUNT(*)
		FROM az_user_roles ur
		INNER JOIN az_roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name IN ('admin', 'super_admin', 'moderator')
	`
	
	err := s.db.Raw(query, userID).Count(&count).Error
	if err != nil {
		return false
	}
	
	return count > 0
}

// getUserPermissions 获取用户权限列表
func (s *AdminAuthService) getUserPermissions(userID uuid.UUID) []string {
	var permissions []string
	query := `
		SELECT DISTINCT p.code
		FROM az_permissions p
		INNER JOIN az_role_permissions rp ON p.id = rp.permission_id
		INNER JOIN az_user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`
	
	err := s.db.Raw(query, userID).Pluck("code", &permissions).Error
	if err != nil {
		return []string{}
	}
	
	return permissions
}

// HasPermission 检查管理员是否有特定权限
func (s *AdminAuthService) HasPermission(claims *AdminClaims, permission string) bool {
	for _, perm := range claims.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetAdminByToken 通过token获取管理员用户信息
func (s *AdminAuthService) GetAdminByToken(tokenString string) (*models.User, error) {
	claims, err := s.ValidateAdminToken(tokenString)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}