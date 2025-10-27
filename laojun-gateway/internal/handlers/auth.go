package handlers

import (
	"net/http"

	"github.com/codetaoist/laojun-gateway/internal/auth"
	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *auth.Service
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(cfg config.AuthConfig, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: auth.NewService(cfg, logger),
		logger:      logger,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshRequest 刷新token请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// 这里应该验证用户凭据，简化处理直接生成token
	// 实际实现中应该调用用户服务验证用户名和密码
	if req.Username == "" || req.Password == "" {
		h.logger.Warn("Empty username or password", 
			zap.String("username", req.Username))
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
			"code":  "INVALID_CREDENTIALS",
		})
		return
	}

	// 模拟用户验证成功，生成token
	userID := "user_" + req.Username
	roles := []string{"user"} // 实际应该从用户服务获取

	accessToken, err := h.authService.GenerateToken(userID, req.Username, roles)
	if err != nil {
		h.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
			"code":  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken(userID)
	if err != nil {
		h.logger.Error("Failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
			"code":  "REFRESH_TOKEN_GENERATION_FAILED",
		})
		return
	}

	h.logger.Info("User logged in successfully", 
		zap.String("username", req.Username),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 从配置获取
	})
}

// RefreshToken 刷新访问token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// 验证刷新token
	userID, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
			"code":  "INVALID_REFRESH_TOKEN",
		})
		return
	}

	// 这里应该从用户服务获取用户信息，简化处理
	username := "user" // 实际应该从用户服务获取
	roles := []string{"user"}

	// 生成新的访问token
	accessToken, err := h.authService.GenerateToken(userID, username, roles)
	if err != nil {
		h.logger.Error("Failed to generate new access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
			"code":  "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// 生成新的刷新token
	newRefreshToken, err := h.authService.GenerateRefreshToken(userID)
	if err != nil {
		h.logger.Error("Failed to generate new refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
			"code":  "REFRESH_TOKEN_GENERATION_FAILED",
		})
		return
	}

	h.logger.Info("Token refreshed successfully", zap.String("user_id", userID))

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 实际实现中应该将token加入黑名单
	// 这里简化处理，只返回成功响应
	
	userID, exists := c.Get("user_id")
	if exists {
		h.logger.Info("User logged out", zap.Any("user_id", userID))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
		"code":    "LOGOUT_SUCCESS",
	})
}