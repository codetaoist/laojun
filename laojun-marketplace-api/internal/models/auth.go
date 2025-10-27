package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/laojun-shared/models"
)

// 使用shared模块的登录相关模型
type LoginRequest = models.LoginRequest
type LoginResponse = models.LoginResponse

// UserSession 用户会话
type UserSession struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash  string     `json:"-" db:"token_hash"`
	IPAddress  *string    `json:"ip_address" db:"ip_address"`
	UserAgent  *string    `json:"user_agent" db:"user_agent"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt time.Time  `json:"last_used_at" db:"last_used_at"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Roles    []string  `json:"roles"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}
