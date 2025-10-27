package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	internalmodels "github.com/codetaoist/laojun-admin-api/internal/models"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// MarketplaceUser 扩展的用户模型（包含marketplace特有字段）
type MarketplaceUser struct {
	models.User
	FullName               *string    `json:"full_name"`
	IsEmailVerified        bool       `json:"is_email_verified"`
	EmailVerificationToken *string    `json:"-"`
	PasswordResetToken     *string    `json:"-"`
	PasswordResetExpiresAt *time.Time `json:"-"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	FullName        string `json:"full_name,omitempty"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *MarketplaceUser `json:"user"`
	Token        string           `json:"token"`
	RefreshToken string           `json:"refresh_token,omitempty"`
	ExpiresAt    int64            `json:"expires_at"`
}

// AuthService 认证服务
type AuthService struct {
	db *shareddb.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *shareddb.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*MarketplaceUser, error) {
	// 验证密码确认
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("密码确认不匹配")
	}

	// 检查用户名是否已存在
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM mp_users WHERE username = $1)", req.Username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM mp_users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 生成邮箱验证token
	verificationToken, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("生成验证token失败: %w", err)
	}

	// 创建用户
	userID := uuid.New()
	now := time.Now()

	// 插入数据
	query := `
		INSERT INTO mp_users (
			id, username, email, password_hash, full_name, is_active, 
			is_email_verified, email_verification_token, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var fullName *string
	if req.FullName != "" {
		fullName = &req.FullName
	}

	_, err = s.db.Exec(query, userID, req.Username, req.Email, string(hashedPassword),
		fullName, true, false, verificationToken, now, now)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建返回的用户对象
	user := &MarketplaceUser{
		User: models.User{
			ID:          userID,
			Username:    req.Username,
			Email:       req.Email,
			Avatar:      nil,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
			LastLoginAt: nil,
		},
		FullName:        fullName,
		IsEmailVerified: false,
	}

	return user, nil
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*MarketplaceUser, error) {
	// 根据用户名或邮箱查找用户
	query := `
		SELECT id, username, email, password_hash, full_name, avatar, avatar_url,
			   is_active, is_email_verified, created_at, updated_at, last_login_at
		FROM mp_users 
		WHERE (username = $1 OR email = $1) AND is_active = true`

	var (
		id              uuid.UUID
		username        string
		email           string
		passwordHash    string
		fullName        sql.NullString
		avatar          sql.NullString
		avatarURL       sql.NullString
		isActive        bool
		isEmailVerified bool
		createdAt       time.Time
		updatedAt       time.Time
		lastLoginAt     sql.NullTime
	)

	err := s.db.QueryRow(query, req.Username).Scan(
		&id, &username, &email, &passwordHash, &fullName, &avatar, &avatarURL,
		&isActive, &isEmailVerified, &createdAt, &updatedAt, &lastLoginAt,
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
	_, err = s.db.Exec("UPDATE mp_users SET last_login_at = $1, updated_at = $2 WHERE id = $3",
		now, now, id)
	if err != nil {
		// 记录错误但不影响登录
		fmt.Printf("更新最后登录时间失败: %v\n", err)
	}

	// 构建用户对象
	user := &MarketplaceUser{
		User: models.User{
			ID:          id,
			Username:    username,
			Email:       email,
			IsActive:    isActive,
			CreatedAt:   createdAt,
			UpdatedAt:   now,
			LastLoginAt: &now,
		},
		IsEmailVerified: isEmailVerified,
	}

	if fullName.Valid {
		user.FullName = &fullName.String
	}
	if avatar.Valid {
		user.User.Avatar = &avatar.String
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID uuid.UUID) (*MarketplaceUser, error) {
	query := `
		SELECT id, username, email, full_name, avatar, avatar_url,
			   is_active, is_email_verified, created_at, updated_at, last_login_at
		FROM mp_users 
		WHERE id = $1 AND is_active = true`

	var (
		id              uuid.UUID
		username        string
		email           string
		fullName        sql.NullString
		avatar          sql.NullString
		avatarURL       sql.NullString
		isActive        bool
		isEmailVerified bool
		createdAt       time.Time
		updatedAt       time.Time
		lastLoginAt     sql.NullTime
	)

	err := s.db.QueryRow(query, userID).Scan(
		&id, &username, &email, &fullName, &avatar, &avatarURL,
		&isActive, &isEmailVerified, &createdAt, &updatedAt, &lastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	user := &MarketplaceUser{
		User: models.User{
			ID:        id,
			Username:  username,
			Email:     email,
			IsActive:  isActive,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		IsEmailVerified: isEmailVerified,
	}

	if fullName.Valid {
		user.FullName = &fullName.String
	}
	if avatar.Valid {
		user.User.Avatar = &avatar.String
	}
	if lastLoginAt.Valid {
		user.User.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(userID uuid.UUID, updates map[string]interface{}) (*MarketplaceUser, error) {
	// 构建更新查询
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "full_name", "avatar", "avatar_url", "bio":
			argIndex++
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
		}
	}

	if len(setParts) == 1 {
		return nil, errors.New("没有可更新的字段")
	}

	query := fmt.Sprintf("UPDATE mp_users SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex+1)
	args = append(args, userID)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("更新用户资料失败: %w", err)
	}

	return s.GetUserByID(userID)
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	// 获取当前密码哈希
	var currentHash string
	err := s.db.DB.QueryRow("SELECT password_hash FROM mp_users WHERE id = $1", userID).Scan(&currentHash)
	if err != nil {
		return fmt.Errorf("获取用户密码失败: %w", err)
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword))
	if err != nil {
		return errors.New("当前密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	_, err = s.db.Exec("UPDATE mp_users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// ValidateToken 验证JWT令牌并提取声明
func (s *AuthService) ValidateToken(tokenString string) (*internalmodels.JWTClaims, error) {
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

	// 支持 roles、单角色 role 以及 is_admin 三种形式
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

// generateRandomToken 生成随机token
func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
