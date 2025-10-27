package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/codetaoist/laojun-marketplace-api/internal/services"
	"github.com/codetaoist/laojun-shared/auth"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MarketplaceAuthHandler marketplace认证处理
type MarketplaceAuthHandler struct {
	authService *services.AuthService
	jwtManager  *auth.JWTManager
	cfg         *sharedconfig.Config
}

// NewMarketplaceAuthHandler 创建marketplace认证处理
func NewMarketplaceAuthHandler(authService *services.AuthService, jwtManager *auth.JWTManager, cfg *sharedconfig.Config) *MarketplaceAuthHandler {
	return &MarketplaceAuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
		cfg:         cfg,
	}
}

// Register 用户注册
func (h *MarketplaceAuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效",
		})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "registration_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Login 用户登录
func (h *MarketplaceAuthHandler) Login(c *gin.Context) {
	// 登录请求增加验证码参数
	var req struct {
		Username   string `json:"username" binding:"required"`
		Password   string `json:"password" binding:"required"`
		Remember   bool   `json:"remember"`
		Captcha    string `json:"captcha"`
		CaptchaKey string `json:"captcha_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效",
		})
		return
	}

	// 根据配置决定是否校验验证码
	if h.cfg != nil && h.cfg.Security.MarketplaceCaptchaEnabled {
		if req.Captcha == "" || req.CaptchaKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少验证码参数"})
			return
		}
		stored, ok := getMarketplaceCaptchaCode(req.CaptchaKey, true)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码已过期或不存在"})
			return
		}
		if stored != req.Captcha {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
			return
		}
	}

	// 转换为服务层的登录请求
	loginReq := services.LoginRequest{Username: req.Username, Password: req.Password}

	user, err := h.authService.Login(&loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "login_failed",
			"message": err.Error(),
		})
		return
	}

	// 生成JWT token
	token, _, err := h.jwtManager.GenerateToken(&user.User, false) // marketplace用户不是管理员
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_generation_failed",
			"message": "生成令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// GetProfile 获取用户资料
func (h *MarketplaceAuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未找到用户信息",
		})
		return
	}

	var userUUID uuid.UUID
	var err error

	// 支持uuid.UUID和string两种类型
	switch v := userID.(type) {
	case uuid.UUID:
		userUUID = v
	case string:
		userUUID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_user_id",
				"message": "无效的用户ID",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "用户ID格式错误",
		})
		return
	}

	user, err := h.authService.GetUserByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                user.ID,
			"username":          user.Username,
			"email":             user.Email,
			"full_name":         user.FullName,
			"avatar":            user.Avatar,
			"is_email_verified": user.IsEmailVerified,
			"created_at":        user.CreatedAt,
			"updated_at":        user.UpdatedAt,
		},
	})
}

// Logout 用户登出
func (h *MarketplaceAuthHandler) Logout(c *gin.Context) {
	// 在实际应用中，这里应该将token加入黑名单，防止重复使用
	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// RefreshToken 刷新令牌
func (h *MarketplaceAuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效",
		})
		return
	}

	// 这里应该验证refresh token并生成新的access token
	// 简化实现，实际应用中需要更复杂的逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "令牌刷新成功",
		"token":   "new_access_token",
	})
}

// UpdateProfile 更新用户资料
func (h *MarketplaceAuthHandler) UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "功能暂未实现",
	})
}

// ChangePassword 修改密码
func (h *MarketplaceAuthHandler) ChangePassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "功能暂未实现",
	})
}

// GetUserStats 获取用户统计
func (h *MarketplaceAuthHandler) GetUserStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "功能暂未实现",
	})
}

// GetCaptcha 获取验证码
func (h *MarketplaceAuthHandler) GetCaptcha(c *gin.Context) {
	code, err := randomMarketplaceCaptchaCode(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成验证码失败"})
		return
	}
	key := uuid.New().String()
	setMarketplaceCaptcha(key, code)

	svg := svgMarketplaceCaptcha(code, 120, 40)
	b64 := base64.StdEncoding.EncodeToString([]byte(svg))
	image := "data:image/svg+xml;base64," + b64

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data": gin.H{
			"image": image,
			"key":   key,
		},
	})
}

// GetCaptchaCodeDebug 获取验证码调试信息
func (h *MarketplaceAuthHandler) GetCaptchaCodeDebug(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少key参数"})
		return
	}
	code, ok := getMarketplaceCaptchaCode(key, false)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "验证码不存在或已过期"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data": gin.H{
			"key":  key,
			"code": code,
		},
	})
}

// 内存验证码存储（带过期）- marketplace专用
type marketplaceCaptchaEntry struct {
	code      string
	expiresAt time.Time
}

var marketplaceCaptchaStore struct {
	mu  sync.Mutex
	m   map[string]marketplaceCaptchaEntry
	ttl time.Duration
}

func ensureMarketplaceCaptchaStore() {
	marketplaceCaptchaStore.mu.Lock()
	defer marketplaceCaptchaStore.mu.Unlock()
	if marketplaceCaptchaStore.m == nil {
		marketplaceCaptchaStore.m = make(map[string]marketplaceCaptchaEntry)
		marketplaceCaptchaStore.ttl = 2 * time.Minute
	}
}

func setMarketplaceCaptcha(key, code string) {
	ensureMarketplaceCaptchaStore()
	marketplaceCaptchaStore.mu.Lock()
	marketplaceCaptchaStore.m[key] = marketplaceCaptchaEntry{code: code, expiresAt: time.Now().Add(marketplaceCaptchaStore.ttl)}
	marketplaceCaptchaStore.mu.Unlock()
}

func getMarketplaceCaptchaCode(key string, clear bool) (string, bool) {
	ensureMarketplaceCaptchaStore()
	marketplaceCaptchaStore.mu.Lock()
	defer marketplaceCaptchaStore.mu.Unlock()
	entry, ok := marketplaceCaptchaStore.m[key]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expiresAt) {
		delete(marketplaceCaptchaStore.m, key)
		return "", false
	}
	if clear {
		delete(marketplaceCaptchaStore.m, key)
	}
	return entry.code, true
}

func randomMarketplaceCaptchaCode(n int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除易混淆字符
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		idxBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		res[i] = alphabet[idxBig.Int64()]
	}
	return string(res), nil
}

func svgMarketplaceCaptcha(code string, width, height int) string {
	// 简单 SVG 验证码，避免额外图像依赖
	return "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"" +
		fmtMarketplaceInt(width) + "\" height=\"" + fmtMarketplaceInt(height) + "\">" +
		"<rect width=\"100%\" height=\"100%\" fill=\"#ffffff\"/>" +
		"<text x=\"50%\" y=\"50%\" dominant-baseline=\"middle\" text-anchor=\"middle\" font-family=\"monospace\" font-size=\"22\" fill=\"#333\" letter-spacing=\"4\">" + code + "</text>" +
		// 简单噪点与线条
		"<line x1=\"10\" y1=\"10\" x2=\"110\" y2=\"10\" stroke=\"#ddd\" stroke-width=\"1\"/>" +
		"<circle cx=\"20\" cy=\"30\" r=\"1\" fill=\"#ccc\"/>" +
		"<circle cx=\"60\" cy=\"15\" r=\"1\" fill=\"#ccc\"/>" +
		"<circle cx=\"90\" cy=\"25\" r=\"1\" fill=\"#ccc\"/>" +
		"</svg>"
}

func fmtMarketplaceInt(n int) string { return strconv.FormatInt(int64(n), 10) }

// GetCaptchaConfig 获取验证码配置
func (h *MarketplaceAuthHandler) GetCaptchaConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled": h.cfg.Security.MarketplaceCaptchaEnabled,
			"type":    h.cfg.Security.CaptchaType,
		},
	})
}
