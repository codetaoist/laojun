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

// AuthService 认证服务
type AuthService struct {
	db          *gorm.DB
	redisClient *redis.Client
	jwtSecret   string
	tokenExpiry time.Duration
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, redisClient *redis.Client, jwtSecret string, tokenExpiry time.Duration) *AuthService {
	return &AuthService{
		db:          db,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// Claims JWT 声明结构
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	// 获取用户的主要角色
	primaryRole := s.getUserPrimaryRole(user.ID)
	
	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     primaryRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "laojun-marketplace",
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
	key := fmt.Sprintf("auth:token:%s", claims.ID)
	err = s.redisClient.Set(ctx, key, tokenString, s.tokenExpiry).Err()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查token是否在Redis中存在
		ctx := context.Background()
		key := fmt.Sprintf("auth:token:%s", claims.ID)
		exists := s.redisClient.Exists(ctx, key).Val()
		if exists == 0 {
			return nil, fmt.Errorf("token has been revoked")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RevokeToken 撤销令牌
func (s *AuthService) RevokeToken(tokenID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("auth:token:%s", tokenID)
	return s.redisClient.Del(ctx, key).Err()
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return "", err
	}

	// 撤销旧token
	s.RevokeToken(claims.ID)

	// 生成新token
	return s.GenerateToken(&user)
}

// getUserPrimaryRole 获取用户的主要角色
func (s *AuthService) getUserPrimaryRole(userID uuid.UUID) string {
	var roleName string
	query := `
		SELECT r.name 
		FROM az_roles r
		INNER JOIN az_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.is_system DESC, r.created_at ASC
		LIMIT 1
	`
	
	err := s.db.Raw(query, userID).Scan(&roleName).Error
	if err != nil {
		// 如果查询失败或没有角色，返回默认角色
		return "user"
	}
	
	if roleName == "" {
		return "user"
	}
	
	return roleName
}

// GetUserByToken 通过token获取用户信息
func (s *AuthService) GetUserByToken(tokenString string) (*models.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}