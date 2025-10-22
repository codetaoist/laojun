package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NewAuthManager 创建新的认证管理
func NewAuthManager(config AuthConfig) *AuthManager {
	return &AuthManager{
		config:    config,
		apiKeys:   make(map[string]*APIKey),
		sessions:  make(map[string]*Session),
		blacklist: make(map[string]time.Time),
	}
}

// Authenticate 认证请求
func (am *AuthManager) Authenticate(r *http.Request) (*AuthContext, error) {
	// 检查是否需要认证
	if !am.requiresAuth(r) {
		return &AuthContext{
			Authenticated: false,
			Anonymous:     true,
		}, nil
	}

	// 尝试不同的认证方法
	authMethods := []func(*http.Request) (*AuthContext, error){
		am.authenticateJWT,
		am.authenticateAPIKey,
		am.authenticateBasic,
		am.authenticateSession,
		am.authenticateOAuth,
	}

	var lastErr error
	for _, method := range authMethods {
		if ctx, err := method(r); err == nil && ctx != nil {
			// 检查权限
			if err := am.authorize(ctx, r); err != nil {
				return nil, err
			}
			return ctx, nil
		} else if err != nil {
			lastErr = err
		}
	}

	return nil, fmt.Errorf("authentication failed: %v", lastErr)
}

// authenticateJWT JWT认证
func (am *AuthManager) authenticateJWT(r *http.Request) (*AuthContext, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("no bearer token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 检查黑名单
	if _, blacklisted := am.blacklist[tokenString]; blacklisted {
		return nil, errors.New("token is blacklisted")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// 检查过期时间
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("token expired")
		}
	}

	userID, _ := claims["user_id"].(string)
	roles, _ := claims["roles"].([]interface{})
	permissions, _ := claims["permissions"].([]interface{})

	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i], _ = role.(string)
	}

	permissionStrings := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionStrings[i], _ = perm.(string)
	}

	return &AuthContext{
		Authenticated: true,
		UserID:        userID,
		Roles:         roleStrings,
		Permissions:   permissionStrings,
		TokenType:     "jwt",
		Token:         tokenString,
		Claims:        claims,
	}, nil
}

// authenticateAPIKey API Key认证
func (am *AuthManager) authenticateAPIKey(r *http.Request) (*AuthContext, error) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}

	if apiKey == "" {
		return nil, errors.New("no api key")
	}

	am.mu.RLock()
	key, exists := am.apiKeys[apiKey]
	am.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid api key")
	}

	if !key.Active {
		return nil, errors.New("api key is disabled")
	}

	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, errors.New("api key expired")
	}

	// 更新使用统计
	am.mu.Lock()
	key.LastUsed = time.Now()
	key.UsageCount++
	am.mu.Unlock()

	return &AuthContext{
		Authenticated: true,
		UserID:        key.UserID,
		Roles:         key.Roles,
		Permissions:   key.Permissions,
		TokenType:     "api_key",
		Token:         apiKey,
		APIKey:        key,
	}, nil
}

// authenticateBasic Basic认证
func (am *AuthManager) authenticateBasic(r *http.Request) (*AuthContext, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("no basic auth")
	}

	// 这里应该连接到用户数据库验证
	// 为了示例，我们使用简单的硬编码验证
	if am.validateCredentials(username, password) {
		return &AuthContext{
			Authenticated: true,
			UserID:        username,
			TokenType:     "basic",
		}, nil
	}

	return nil, errors.New("invalid credentials")
}

// authenticateSession Session认证
func (am *AuthManager) authenticateSession(r *http.Request) (*AuthContext, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, errors.New("no session cookie")
	}

	am.mu.RLock()
	session, exists := am.sessions[cookie.Value]
	am.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		am.mu.Lock()
		delete(am.sessions, cookie.Value)
		am.mu.Unlock()
		return nil, errors.New("session expired")
	}

	// 更新会话
	am.mu.Lock()
	session.LastAccess = time.Now()
	am.mu.Unlock()

	return &AuthContext{
		Authenticated: true,
		UserID:        session.UserID,
		Roles:         session.Roles,
		Permissions:   session.Permissions,
		TokenType:     "session",
		Token:         cookie.Value,
		Session:       session,
	}, nil
}

// authenticateOAuth OAuth认证
func (am *AuthManager) authenticateOAuth(r *http.Request) (*AuthContext, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("no oauth token")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// 验证OAuth token（这里应该调用OAuth服务器）
	userInfo, err := am.validateOAuthToken(token)
	if err != nil {
		return nil, err
	}

	return &AuthContext{
		Authenticated: true,
		UserID:        userInfo.UserID,
		Roles:         userInfo.Roles,
		Permissions:   userInfo.Permissions,
		TokenType:     "oauth",
		Token:         token,
		OAuthInfo:     userInfo,
	}, nil
}

// authorize 权限检查
func (am *AuthManager) authorize(ctx *AuthContext, r *http.Request) error {
	if !ctx.Authenticated {
		return errors.New("not authenticated")
	}

	// 检查路径权限
	path := r.URL.Path
	method := r.Method

	// 检查是否有访问权限
	if !am.hasPermission(ctx, method, path) {
		return errors.New("insufficient permissions")
	}

	return nil
}

// hasPermission 检查是否有权限
func (am *AuthManager) hasPermission(ctx *AuthContext, method, path string) bool {
	// 超级管理员有所有权
	for _, role := range ctx.Roles {
		if role == "admin" || role == "super_admin" {
			return true
		}
	}

	// 检查具体权限
	requiredPermission := fmt.Sprintf("%s:%s", method, path)
	for _, perm := range ctx.Permissions {
		if perm == requiredPermission || perm == "*" {
			return true
		}
		// 支持通配符权限
		if strings.HasSuffix(perm, "*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(requiredPermission, prefix) {
				return true
			}
		}
	}

	return false
}

// requiresAuth 检查是否需要认证
func (am *AuthManager) requiresAuth(r *http.Request) bool {
	path := r.URL.Path

	// 公开路径不需要认证
	for _, publicPath := range am.config.PublicPaths {
		if strings.HasPrefix(path, publicPath) {
			return false
		}
	}

	return true
}

// validateCredentials 验证用户凭据
func (am *AuthManager) validateCredentials(username, password string) bool {
	// 这里应该连接到用户数据库
	// 为了示例，使用简单的硬编码验证
	return username == "admin" && password == "password"
}

// validateOAuthToken 验证OAuth token
func (am *AuthManager) validateOAuthToken(token string) (*OAuthUserInfo, error) {
	// 这里应该调用OAuth服务器验证token
	// 为了示例，返回模拟数据
	return &OAuthUserInfo{
		UserID:      "oauth_user",
		Email:       "user@example.com",
		Name:        "OAuth User",
		Roles:       []string{"user"},
		Permissions: []string{"read:*"},
	}, nil
}

// CreateAPIKey 创建API Key
func (am *AuthManager) CreateAPIKey(userID string, name string, roles []string, permissions []string, expiresAt *time.Time) (*APIKey, error) {
	apiKey := &APIKey{
		Key:         am.generateAPIKey(),
		UserID:      userID,
		Name:        name,
		Roles:       roles,
		Permissions: permissions,
		Active:      true,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		UsageCount:  0,
	}

	am.mu.Lock()
	am.apiKeys[apiKey.Key] = apiKey
	am.mu.Unlock()

	return apiKey, nil
}

// RevokeAPIKey 撤销API Key
func (am *AuthManager) RevokeAPIKey(key string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if apiKey, exists := am.apiKeys[key]; exists {
		apiKey.Active = false
		return nil
	}

	return errors.New("api key not found")
}

// CreateSession 创建会话
func (am *AuthManager) CreateSession(userID string, roles []string, permissions []string, duration time.Duration) (*Session, error) {
	sessionID := am.generateSessionID()
	session := &Session{
		ID:          sessionID,
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		ExpiresAt:   time.Now().Add(duration),
	}

	am.mu.Lock()
	am.sessions[sessionID] = session
	am.mu.Unlock()

	return session, nil
}

// RevokeSession 撤销会话
func (am *AuthManager) RevokeSession(sessionID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.sessions[sessionID]; exists {
		delete(am.sessions, sessionID)
		return nil
	}

	return errors.New("session not found")
}

// BlacklistToken 将token加入黑名单
func (am *AuthManager) BlacklistToken(token string, expiry time.Time) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.blacklist[token] = expiry
}

// CleanupExpired 清理过期的会话和黑名单
func (am *AuthManager) CleanupExpired() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()

	// 清理过期会话
	for id, session := range am.sessions {
		if now.After(session.ExpiresAt) {
			delete(am.sessions, id)
		}
	}

	// 清理过期黑名单
	for token, expiry := range am.blacklist {
		if now.After(expiry) {
			delete(am.blacklist, token)
		}
	}
}

// generateAPIKey 生成API Key
func (am *AuthManager) generateAPIKey() string {
	// 生成32字节随机数据
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(time.Now().UnixNano() % 256)
	}

	// 使用HMAC-SHA256生成key
	h := hmac.New(sha256.New, []byte(am.config.APIKeySecret))
	h.Write(data)
	hash := h.Sum(nil)

	return hex.EncodeToString(hash)
}

// generateSessionID 生成会话ID
func (am *AuthManager) generateSessionID() string {
	data := make([]byte, 24)
	for i := range data {
		data[i] = byte(time.Now().UnixNano() % 256)
	}
	return base64.URLEncoding.EncodeToString(data)
}

// GetStats 获取认证统计信息
func (am *AuthManager) GetStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	activeAPIKeys := 0
	for _, key := range am.apiKeys {
		if key.Active {
			activeAPIKeys++
		}
	}

	activeSessions := 0
	now := time.Now()
	for _, session := range am.sessions {
		if now.Before(session.ExpiresAt) {
			activeSessions++
		}
	}

	return map[string]interface{}{
		"api_keys": map[string]interface{}{
			"total":  len(am.apiKeys),
			"active": activeAPIKeys,
		},
		"sessions": map[string]interface{}{
			"total":  len(am.sessions),
			"active": activeSessions,
		},
		"blacklist_size": len(am.blacklist),
	}
}

// ValidateJWTClaims 验证JWT声明
func (am *AuthManager) ValidateJWTClaims(claims jwt.MapClaims) error {
	// 检查必需的声明
	if _, ok := claims["user_id"]; !ok {
		return errors.New("missing user_id claim")
	}

	if _, ok := claims["exp"]; !ok {
		return errors.New("missing exp claim")
	}

	if _, ok := claims["iat"]; !ok {
		return errors.New("missing iat claim")
	}

	// 检查发行者
	if iss, ok := claims["iss"].(string); ok {
		if iss != am.config.JWTIssuer {
			return errors.New("invalid issuer")
		}
	}

	// 检查受众
	if aud, ok := claims["aud"].(string); ok {
		if aud != am.config.JWTAudience {
			return errors.New("invalid audience")
		}
	}

	return nil
}

// RefreshToken 刷新JWT token
func (am *AuthManager) RefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	// 创建新的token
	newClaims := jwt.MapClaims{
		"user_id":     claims["user_id"],
		"roles":       claims["roles"],
		"permissions": claims["permissions"],
		"iss":         am.config.JWTIssuer,
		"aud":         am.config.JWTAudience,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return newToken.SignedString([]byte(am.config.JWTSecret))
}

// EncryptSensitiveData 加密敏感数据
func (am *AuthManager) EncryptSensitiveData(data string) (string, error) {
	// 这里应该使用适当的加密算法，例如AES-GCM
	// 为了示例，使用简单的base64编码
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

// DecryptSensitiveData 解密敏感数据
func (am *AuthManager) DecryptSensitiveData(encryptedData string) (string, error) {
	// 这里应该使用适当的解密算法，例如AES-GCM
	// 为了示例，使用简单的base64解码
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LogAuthEvent 记录认证事件
func (am *AuthManager) LogAuthEvent(event AuthEvent) {
	// 这里应该记录到审计日志或数据库
	eventJSON, _ := json.Marshal(event)
	fmt.Printf("Auth Event: %s\n", string(eventJSON))
}

// AuthEvent 认证事件
type AuthEvent struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id,omitempty"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
	Path      string                 `json:"path"`
	Method    string                 `json:"method"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
