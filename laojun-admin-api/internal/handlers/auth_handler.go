package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/codetaoist/laojun-shared/auth"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler 认证处理
type AuthHandler struct {
	adminAuthService *services.AdminAuthService
	jwtManager       *auth.JWTManager
	cfg              *sharedconfig.Config
}

// NewAuthHandler 创建认证处理
func NewAuthHandler(adminAuthService *services.AdminAuthService, jwtManager *auth.JWTManager, cfg *sharedconfig.Config) *AuthHandler {
	return &AuthHandler{
		adminAuthService: adminAuthService,
		jwtManager:       jwtManager,
		cfg:              cfg,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 后台管理系统不支持注册功能
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error": "后台管理系统不支持注册功能",
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
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
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 根据配置决定是否校验验证码
	// 添加调试日志
	fmt.Printf("DEBUG: AdminCaptchaEnabled = %v\n", h.cfg.Security.AdminCaptchaEnabled)
	if h.cfg != nil && h.cfg.Security.AdminCaptchaEnabled {
		if req.Captcha == "" || req.CaptchaKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少验证码参数"})
			return
		}
		stored, ok := getCaptchaCode(req.CaptchaKey, true)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码已过期或不存在"})
			return
		}
		if stored != req.Captcha {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
			return
		}
	}

	loginReq := services.AdminLoginRequest{Username: req.Username, Password: req.Password, Remember: req.Remember}
	user, err := h.adminAuthService.Login(&loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 生成 JWT token
	token, expiresAt, err := h.jwtManager.GenerateToken(&user.User, true) // 后台用户默认为管理员
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user":      user,
			"token":     token,
			"expiresAt": expiresAt,
		},
		"message": "登录成功",
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从header获取token
	token := c.GetHeader("Authorization")
	if token != "" && len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
		// 将token加入黑名单（这里简化处理，实际项目中可能需要Redis等）
		// h.jwtManager.BlacklistToken(token)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	// 刷新token
	newToken, expiresAt, err := h.jwtManager.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "token刷新失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "刷新成功",
		"data": gin.H{
			"token":      newToken,
			"expires_at": expiresAt.Unix(),
		},
	})
}

// GetProfile 获取用户资料
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从认证中间件获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	var userID uuid.UUID
	var err error

	// 支持string和uuid.UUID两种类型
	switch v := userIDInterface.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "用户ID格式错误",
			})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户信息格式错误",
		})
		return
	}

	// 获取完整的后台用户信息
	adminUser, err := h.adminAuthService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    adminUser,
	})
}

// UpdateProfile 更新用户资料
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// 后台管理系统暂不支持更新用户资料功能
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error": "后台管理系统暂不支持更新用户资料功能",
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// 后台管理系统暂不支持修改密码功能
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error": "后台管理系统暂不支持修改密码功能",
	})
}

// GetUserStats 获取用户统计信息
func (h *AuthHandler) GetUserStats(c *gin.Context) {
	// 从认证中间件获取用户ID
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权访问",
		})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "用户信息格式错误",
		})
		return
	}

	// 这里可以添加获取用户统计信息的逻辑
	// 比如：用户发布的插件数量、收藏数量、评论数量等
	stats := gin.H{
		"user_id":         userID,
		"plugins_count":   0, // 待实现
		"favorites_count": 0, // 待实现
		"reviews_count":   0, // 待实现
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data":    stats,
	})
}

// 内存验证码存储（带过期）
type captchaEntry struct {
	code      string
	expiresAt time.Time
}

var captchaStore struct {
	mu  sync.Mutex
	m   map[string]captchaEntry
	ttl time.Duration
}

func ensureCaptchaStore() {
	captchaStore.mu.Lock()
	defer captchaStore.mu.Unlock()
	if captchaStore.m == nil {
		captchaStore.m = make(map[string]captchaEntry)
		captchaStore.ttl = 2 * time.Minute
	}
}

func setCaptcha(key, code string) {
	ensureCaptchaStore()
	captchaStore.mu.Lock()
	captchaStore.m[key] = captchaEntry{code: code, expiresAt: time.Now().Add(captchaStore.ttl)}
	captchaStore.mu.Unlock()
}

func getCaptchaCode(key string, clear bool) (string, bool) {
	ensureCaptchaStore()
	captchaStore.mu.Lock()
	defer captchaStore.mu.Unlock()
	entry, ok := captchaStore.m[key]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expiresAt) {
		delete(captchaStore.m, key)
		return "", false
	}
	if clear {
		delete(captchaStore.m, key)
	}
	return entry.code, true
}

func randomCaptchaCode(n int) (string, error) {
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

func svgCaptcha(code string, width, height int) string {
	// 简单 SVG 验证码，避免额外图像依赖
	return "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"" +
		fmtInt(width) + "\" height=\"" + fmtInt(height) + "\">" +
		"<rect width=\"100%\" height=\"100%\" fill=\"#ffffff\"/>" +
		"<text x=\"50%\" y=\"50%\" dominant-baseline=\"middle\" text-anchor=\"middle\" font-family=\"monospace\" font-size=\"22\" fill=\"#333\" letter-spacing=\"4\">" + code + "</text>" +
		// 简单噪点与线条
		"<line x1=\"10\" y1=\"10\" x2=\"110\" y2=\"10\" stroke=\"#ddd\" stroke-width=\"1\"/>" +
		"<circle cx=\"20\" cy=\"30\" r=\"1\" fill=\"#ccc\"/>" +
		"<circle cx=\"60\" cy=\"15\" r=\"1\" fill=\"#ccc\"/>" +
		"<circle cx=\"90\" cy=\"25\" r=\"1\" fill=\"#ccc\"/>" +
		"</svg>"
}

func fmtInt(n int) string { return strconv.FormatInt(int64(n), 10) }

// 获取验证码
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	code, err := randomCaptchaCode(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成验证码失败"})
		return
	}
	key := uuid.New().String()
	setCaptcha(key, code)

	svg := svgCaptcha(code, 120, 40)
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

// 调试接口：返回验证码明文
func (h *AuthHandler) GetCaptchaCodeDebug(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少key参数"})
		return
	}
	code, ok := getCaptchaCode(key, false)
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

// GetCaptchaConfig 获取验证码配置
func (h *AuthHandler) GetCaptchaConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled": h.cfg.Security.AdminCaptchaEnabled,
			"type":    h.cfg.Security.CaptchaType,
		},
	})
}
