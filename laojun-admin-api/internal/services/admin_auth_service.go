package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	internalmodels "github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminUser 后台管理用户模型
type AdminUser struct {
	models.User
}

// AdminLoginRequest 后台登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

// AdminAuthResponse 后台认证响应
type AdminAuthResponse struct {
	User      *AdminUser `json:"user"`
	Token     string     `json:"token"`
	ExpiresAt int64      `json:"expires_at"`
}

// AdminAuthService 后台认证服务
type AdminAuthService struct {
	db *shareddb.DB
}

// NewAdminAuthService 创建后台认证服务
func NewAdminAuthService(db *shareddb.DB) *AdminAuthService {
	return &AdminAuthService{
		db: db,
	}
}

// Login 后台用户登录
func (s *AdminAuthService) Login(req *AdminLoginRequest) (*AdminUser, error) {
	// 根据用户名或邮箱查找用户
	query := `
		SELECT id, username, email, password_hash, avatar, 
			   is_active, created_at, updated_at, last_login_at
		FROM ua_admin 
		WHERE (username = $1 OR email = $1) AND is_active = true`

	var (
		id           uuid.UUID
		username     string
		email        string
		passwordHash string
		avatar       sql.NullString
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
		lastLoginAt  sql.NullTime
	)

	err := s.db.QueryRow(query, req.Username).Scan(
		&id, &username, &email, &passwordHash, &avatar,
		&isActive, &createdAt, &updatedAt, &lastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	_, err = s.db.Exec("UPDATE ua_admin SET last_login_at = $1, updated_at = $2 WHERE id = $3",
		now, now, id)
	if err != nil {
		// 记录错误但不影响登录
		fmt.Printf("更新最后登录时间失败: %v\n", err)
	}

	// 构建用户对象
	user := &AdminUser{
		User: models.User{
			ID:          id,
			Username:    username,
			Email:       email,
			IsActive:    isActive,
			CreatedAt:   createdAt,
			UpdatedAt:   now,
			LastLoginAt: &now,
		},
	}

	if avatar.Valid {
		user.User.Avatar = &avatar.String
	}

	return user, nil
}

// GetUserByID 根据ID获取后台用户
func (s *AdminAuthService) GetUserByID(userID uuid.UUID) (*AdminUser, error) {
	query := `
		SELECT id, username, email, avatar,
			   is_active, created_at, updated_at, last_login_at
		FROM ua_admin 
		WHERE id = $1 AND is_active = true`

	var (
		id          uuid.UUID
		username    string
		email       string
		avatar      sql.NullString
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
		lastLoginAt sql.NullTime
	)

	err := s.db.QueryRow(query, userID).Scan(
		&id, &username, &email, &avatar,
		&isActive, &createdAt, &updatedAt, &lastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	user := &AdminUser{
		User: models.User{
			ID:        id,
			Username:  username,
			Email:     email,
			IsActive:  isActive,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
	}

	if avatar.Valid {
		user.User.Avatar = &avatar.String
	}
	if lastLoginAt.Valid {
		user.User.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// CreateAdminUser 创建后台管理员用户
func (s *AdminAuthService) CreateAdminUser(username, email, password string) (*AdminUser, error) {
	// 检查用户名和邮箱是否已存在
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM ua_admin WHERE username = $1 OR email = $2", username, email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("检查用户存在性失败: %w", err)
	}
	if count > 0 {
		return nil, errors.New("用户名或邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 插入用户
	userID := uuid.New()
	now := time.Now()
	query := `
		INSERT INTO ua_admin (id, username, email, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.Exec(query, userID, username, email, string(hashedPassword), true, now, now)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &AdminUser{
		User: models.User{
			ID:        userID,
			Username:  username,
			Email:     email,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

// ValidateToken 验证JWT令牌并提取声明
func (s *AdminAuthService) ValidateToken(tokenString string) (*internalmodels.JWTClaims, error) {
	cfg, err := sharedconfig.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	// 过期检查
	if exp, ok := claimsMap["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	var userID uuid.UUID
	if uidRaw, ok := claimsMap["user_id"]; ok {
		switch v := uidRaw.(type) {
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				userID = parsed
			}
		}
	}
	username, _ := claimsMap["username"].(string)
	email, _ := claimsMap["email"].(string)

	// 支持 roles、单个role 以及 is_admin 三种形式
	var roles []string
	if r, ok := claimsMap["roles"]; ok {
		if arr, ok := r.([]interface{}); ok {
			for _, it := range arr {
				if s, ok := it.(string); ok {
					roles = append(roles, s)
				}
			}
		} else if arrStr, ok := r.([]string); ok {
			roles = arrStr
		}
	}
	if roleSingle, ok := claimsMap["role"].(string); ok && roleSingle != "" {
		roles = append(roles, roleSingle)
	}
	if isAdmin, ok := claimsMap["is_admin"].(bool); ok && isAdmin {
		roles = append(roles, "admin")
	}

	return &internalmodels.JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
	}, nil
}