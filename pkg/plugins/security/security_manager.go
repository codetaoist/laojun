package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnablePermissionCheck bool          `json:"enable_permission_check"`
	EnableResourceLimit   bool          `json:"enable_resource_limit"`
	EnableAuditLog        bool          `json:"enable_audit_log"`
	DefaultPermissions    []string      `json:"default_permissions"`
	MaxCPUUsage           float64       `json:"max_cpu_usage"`
	MaxMemoryUsage        int64         `json:"max_memory_usage"`
	MaxDiskUsage          int64         `json:"max_disk_usage"`
	MaxNetworkBandwidth   int64         `json:"max_network_bandwidth"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	TokenExpiry           time.Duration `json:"token_expiry"`
	AuditLogRetention     time.Duration `json:"audit_log_retention"`
}

// Permission 权限定义
type Permission struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Conditions  map[string]string `json:"conditions,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Role 角色定义
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// User 用户定义
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Roles       []Role            `json:"roles"`
	Permissions []Permission      `json:"permissions"` // 直接分配的权限
	Metadata    map[string]string `json:"metadata,omitempty"`
	Status      string            `json:"status"` // active, suspended, banned
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastLoginAt time.Time         `json:"last_login_at"`
}

// SecurityToken 安全令牌
type SecurityToken struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Token     string            `json:"token"`
	Type      string            `json:"type"` // access, refresh, api_key
	Scope     []string          `json:"scope"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	ExpiresAt time.Time         `json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
	RevokedAt *time.Time        `json:"revoked_at,omitempty"`
}

// ResourceLimit 资源限制
type ResourceLimit struct {
	PluginID         string    `json:"plugin_id"`
	MaxCPUUsage      float64   `json:"max_cpu_usage"`      // CPU使用率百分比
	MaxMemoryUsage   int64     `json:"max_memory_usage"`   // 内存使用量（字节）
	MaxDiskUsage     int64     `json:"max_disk_usage"`     // 磁盘使用量（字节）
	MaxNetworkIO     int64     `json:"max_network_io"`     // 网络IO（字节/秒）
	MaxFileHandles   int       `json:"max_file_handles"`   // 最大文件句柄数
	MaxConnections   int       `json:"max_connections"`    // 最大连接数
	MaxExecutionTime int64     `json:"max_execution_time"` // 最大执行时间（秒）
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id,omitempty"`
	PluginID  string                 `json:"plugin_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"` // success, failure, denied
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SecurityContext 安全上下文
type SecurityContext struct {
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	TokenID     string            `json:"token_id"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// SecurityManager 安全管理器
type SecurityManager struct {
	config         SecurityConfig
	users          map[string]*User
	roles          map[string]*Role
	permissions    map[string]*Permission
	tokens         map[string]*SecurityToken
	resourceLimits map[string]*ResourceLimit
	auditLogs      []*AuditLog
	mu             sync.RWMutex

	// 资源监控
	resourceMonitor *ResourceMonitor

	// 权限缓存
	permissionCache map[string]map[string]bool
	cacheMu         sync.RWMutex
}

// NewSecurityManager 创建新的安全管理器
func NewSecurityManager(config SecurityConfig) *SecurityManager {
	if config.SessionTimeout <= 0 {
		config.SessionTimeout = 24 * time.Hour
	}
	if config.TokenExpiry <= 0 {
		config.TokenExpiry = 1 * time.Hour
	}
	if config.AuditLogRetention <= 0 {
		config.AuditLogRetention = 30 * 24 * time.Hour // 30天
	}

	sm := &SecurityManager{
		config:          config,
		users:           make(map[string]*User),
		roles:           make(map[string]*Role),
		permissions:     make(map[string]*Permission),
		tokens:          make(map[string]*SecurityToken),
		resourceLimits:  make(map[string]*ResourceLimit),
		auditLogs:       make([]*AuditLog, 0),
		permissionCache: make(map[string]map[string]bool),
		resourceMonitor: NewResourceMonitor(),
	}

	// 初始化默认权限和角色
	sm.initializeDefaults()

	// 启动清理任务
	go sm.startCleanupTasks()

	return sm
}

// 用户管理

// CreateUser 创建用户
func (sm *SecurityManager) CreateUser(user *User) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if user.ID == "" {
		user.ID = sm.generateID()
	}

	if _, exists := sm.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Status = "active"

	sm.users[user.ID] = user

	sm.logAudit("", "", "create_user", "user", "success", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	}, "", "")

	return nil
}

// GetUser 获取用户
func (sm *SecurityManager) GetUser(userID string) (*User, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	user, exists := sm.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUser 更新用户
func (sm *SecurityManager) UpdateUser(user *User) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, exists := sm.users[user.ID]
	if !exists {
		return errors.New("user not found")
	}

	user.CreatedAt = existing.CreatedAt
	user.UpdatedAt = time.Now()

	sm.users[user.ID] = user

	// 清除权限缓存
	sm.clearPermissionCache(user.ID)

	sm.logAudit("", "", "update_user", "user", "success", map[string]interface{}{
		"user_id": user.ID,
	}, "", "")

	return nil
}

// DeleteUser 删除用户
func (sm *SecurityManager) DeleteUser(userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.users[userID]; !exists {
		return errors.New("user not found")
	}

	delete(sm.users, userID)

	// 撤销所有关联令牌
	for tokenID, token := range sm.tokens {
		if token.UserID == userID {
			now := time.Now()
			token.RevokedAt = &now
		}
	}

	// 清除权限缓存
	sm.clearPermissionCache(userID)

	sm.logAudit("", "", "delete_user", "user", "success", map[string]interface{}{
		"user_id": userID,
	}, "", "")

	return nil
}

// 角色管理

// CreateRole 创建角色
func (sm *SecurityManager) CreateRole(role *Role) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if role.ID == "" {
		role.ID = sm.generateID()
	}

	if _, exists := sm.roles[role.ID]; exists {
		return errors.New("role already exists")
	}

	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	sm.roles[role.ID] = role

	sm.logAudit("", "", "create_role", "role", "success", map[string]interface{}{
		"role_id":   role.ID,
		"role_name": role.Name,
	}, "", "")

	return nil
}

// GetRole 获取角色
func (sm *SecurityManager) GetRole(roleID string) (*Role, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	role, exists := sm.roles[roleID]
	if !exists {
		return nil, errors.New("role not found")
	}

	return role, nil
}

// AssignRole 分配角色给用户
func (sm *SecurityManager) AssignRole(userID, roleID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	user, exists := sm.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	role, exists := sm.roles[roleID]
	if !exists {
		return errors.New("role not found")
	}

	// 检查是否已分配
	for _, userRole := range user.Roles {
		if userRole.ID == roleID {
			return errors.New("role already assigned")
		}
	}

	user.Roles = append(user.Roles, *role)
	user.UpdatedAt = time.Now()

	// 清除权限缓存
	sm.clearPermissionCache(userID)

	sm.logAudit(userID, "", "assign_role", "user_role", "success", map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	}, "", "")

	return nil
}

// RevokeRole 撤销用户角色
func (sm *SecurityManager) RevokeRole(userID, roleID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	user, exists := sm.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	for i, role := range user.Roles {
		if role.ID == roleID {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			user.UpdatedAt = time.Now()

			// 清除权限缓存
			sm.clearPermissionCache(userID)

			sm.logAudit(userID, "", "revoke_role", "user_role", "success", map[string]interface{}{
				"user_id": userID,
				"role_id": roleID,
			}, "", "")

			return nil
		}
	}

	return errors.New("role not assigned to user")
}

// 权限管理

// CreatePermission 创建权限
func (sm *SecurityManager) CreatePermission(permission *Permission) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if permission.ID == "" {
		permission.ID = sm.generateID()
	}

	if _, exists := sm.permissions[permission.ID]; exists {
		return errors.New("permission already exists")
	}

	permission.CreatedAt = time.Now()

	sm.permissions[permission.ID] = permission

	sm.logAudit("", "", "create_permission", "permission", "success", map[string]interface{}{
		"permission_id":   permission.ID,
		"permission_name": permission.Name,
	}, "", "")

	return nil
}

// CheckPermission 检查权限
func (sm *SecurityManager) CheckPermission(userID, resource, action string) bool {
	if !sm.config.EnablePermissionCheck {
		return true
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 检查缓存
	cacheKey := fmt.Sprintf("%s:%s:%s", userID, resource, action)
	sm.cacheMu.RLock()
	if userCache, exists := sm.permissionCache[userID]; exists {
		if result, cached := userCache[cacheKey]; cached {
			sm.cacheMu.RUnlock()
			return result
		}
	}
	sm.cacheMu.RUnlock()

	user, exists := sm.users[userID]
	if !exists {
		sm.cachePermissionResult(userID, cacheKey, false)
		return false
	}

	// 检查用户状态
	if user.Status != "active" {
		sm.cachePermissionResult(userID, cacheKey, false)
		return false
	}

	// 检查直接权限
	for _, permission := range user.Permissions {
		if sm.matchPermission(permission, resource, action) {
			sm.cachePermissionResult(userID, cacheKey, true)
			return true
		}
	}

	// 检查角色权限
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if sm.matchPermission(permission, resource, action) {
				sm.cachePermissionResult(userID, cacheKey, true)
				return true
			}
		}
	}

	sm.cachePermissionResult(userID, cacheKey, false)
	return false
}

// 令牌管理

// CreateToken 创建令牌
func (sm *SecurityManager) CreateToken(userID, tokenType string, scope []string) (*SecurityToken, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	user, exists := sm.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	if user.Status != "active" {
		return nil, errors.New("user is not active")
	}

	token := &SecurityToken{
		ID:        sm.generateID(),
		UserID:    userID,
		Token:     sm.generateToken(),
		Type:      tokenType,
		Scope:     scope,
		ExpiresAt: time.Now().Add(sm.config.TokenExpiry),
		CreatedAt: time.Now(),
	}

	sm.tokens[token.ID] = token

	sm.logAudit(userID, "", "create_token", "token", "success", map[string]interface{}{
		"token_id":   token.ID,
		"token_type": tokenType,
	}, "", "")

	return token, nil
}

// ValidateToken 验证令牌
func (sm *SecurityManager) ValidateToken(tokenString string) (*SecurityContext, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, token := range sm.tokens {
		if token.Token == tokenString {
			// 检查是否已撤销
			if token.RevokedAt != nil {
				return nil, errors.New("token has been revoked")
			}

			// 检查是否过期
			if time.Now().After(token.ExpiresAt) {
				return nil, errors.New("token has expired")
			}

			// 获取用户信息
			user, exists := sm.users[token.UserID]
			if !exists {
				return nil, errors.New("user not found")
			}

			if user.Status != "active" {
				return nil, errors.New("user is not active")
			}

			// 构建安全上下文
			context := &SecurityContext{
				UserID:    user.ID,
				Username:  user.Username,
				TokenID:   token.ID,
				CreatedAt: time.Now(),
			}

			// 收集角色
			for _, role := range user.Roles {
				context.Roles = append(context.Roles, role.Name)
			}

			// 收集权限
			permissionSet := make(map[string]bool)
			for _, permission := range user.Permissions {
				permissionSet[permission.Name] = true
			}
			for _, role := range user.Roles {
				for _, permission := range role.Permissions {
					permissionSet[permission.Name] = true
				}
			}

			for permission := range permissionSet {
				context.Permissions = append(context.Permissions, permission)
			}

			return context, nil
		}
	}

	return nil, errors.New("invalid token")
}

// RevokeToken 撤销令牌
func (sm *SecurityManager) RevokeToken(tokenID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token, exists := sm.tokens[tokenID]
	if !exists {
		return errors.New("token not found")
	}

	now := time.Now()
	token.RevokedAt = &now

	sm.logAudit(token.UserID, "", "revoke_token", "token", "success", map[string]interface{}{
		"token_id": tokenID,
	}, "", "")

	return nil
}

// 资源限制管理

// SetResourceLimit 设置资源限制
func (sm *SecurityManager) SetResourceLimit(pluginID string, limit *ResourceLimit) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	limit.PluginID = pluginID
	limit.UpdatedAt = time.Now()
	if limit.CreatedAt.IsZero() {
		limit.CreatedAt = time.Now()
	}

	sm.resourceLimits[pluginID] = limit

	sm.logAudit("", pluginID, "set_resource_limit", "resource_limit", "success", map[string]interface{}{
		"plugin_id": pluginID,
		"limits":    limit,
	}, "", "")

	return nil
}

// GetResourceLimit 获取资源限制
func (sm *SecurityManager) GetResourceLimit(pluginID string) (*ResourceLimit, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	limit, exists := sm.resourceLimits[pluginID]
	if !exists {
		// 返回默认限制
		return &ResourceLimit{
			PluginID:         pluginID,
			MaxCPUUsage:      sm.config.MaxCPUUsage,
			MaxMemoryUsage:   sm.config.MaxMemoryUsage,
			MaxDiskUsage:     sm.config.MaxDiskUsage,
			MaxNetworkIO:     sm.config.MaxNetworkBandwidth,
			MaxFileHandles:   1000,
			MaxConnections:   100,
			MaxExecutionTime: 3600, // 1小时
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}, nil
	}

	return limit, nil
}

// CheckResourceUsage 检查资源使用情况
func (sm *SecurityManager) CheckResourceUsage(pluginID string) (*ResourceUsage, error) {
	if !sm.config.EnableResourceLimit {
		return nil, nil
	}

	return sm.resourceMonitor.GetResourceUsage(pluginID)
}

// 审计日志

// LogAudit 记录审计日志
func (sm *SecurityManager) LogAudit(userID, pluginID, action, resource, result string, details map[string]interface{}, ipAddress, userAgent string) {
	if !sm.config.EnableAuditLog {
		return
	}

	sm.logAudit(userID, pluginID, action, resource, result, details, ipAddress, userAgent)
}

// GetAuditLogs 获取审计日志
func (sm *SecurityManager) GetAuditLogs(filter map[string]interface{}, limit int, offset int) ([]*AuditLog, int, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var filtered []*AuditLog

	for _, log := range sm.auditLogs {
		if sm.matchAuditFilter(log, filter) {
			filtered = append(filtered, log)
		}
	}

	total := len(filtered)

	// 分页
	if offset >= total {
		return []*AuditLog{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

// 私有方法

// initializeDefaults 初始化默认权限和角色
func (sm *SecurityManager) initializeDefaults() {
	// 创建默认权限
	defaultPermissions := []Permission{
		{ID: "read_plugins", Name: "Read Plugins", Description: "Read plugin information", Resource: "plugin", Action: "read"},
		{ID: "write_plugins", Name: "Write Plugins", Description: "Create and update plugins", Resource: "plugin", Action: "write"},
		{ID: "delete_plugins", Name: "Delete Plugins", Description: "Delete plugins", Resource: "plugin", Action: "delete"},
		{ID: "manage_users", Name: "Manage Users", Description: "Manage user accounts", Resource: "user", Action: "*"},
		{ID: "view_audit_logs", Name: "View Audit Logs", Description: "View audit logs", Resource: "audit_log", Action: "read"},
	}

	for _, permission := range defaultPermissions {
		permission.CreatedAt = time.Now()
		sm.permissions[permission.ID] = &permission
	}

	// 创建默认角色
	adminRole := &Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full system access",
		Permissions: defaultPermissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	userRole := &Role{
		ID:          "user",
		Name:        "User",
		Description: "Basic user access",
		Permissions: []Permission{defaultPermissions[0]}, // 只有读权限
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	sm.roles[adminRole.ID] = adminRole
	sm.roles[userRole.ID] = userRole
}

// generateID 生成唯一ID
func (sm *SecurityManager) generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateToken 生成令牌
func (sm *SecurityManager) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// matchPermission 匹配权限
func (sm *SecurityManager) matchPermission(permission Permission, resource, action string) bool {
	if permission.Resource != "*" && permission.Resource != resource {
		return false
	}

	if permission.Action != "*" && permission.Action != action {
		return false
	}

	return true
}

// cachePermissionResult 缓存权限检查结果
func (sm *SecurityManager) cachePermissionResult(userID, cacheKey string, result bool) {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()

	if _, exists := sm.permissionCache[userID]; !exists {
		sm.permissionCache[userID] = make(map[string]bool)
	}

	sm.permissionCache[userID][cacheKey] = result
}

// clearPermissionCache 清除权限缓存
func (sm *SecurityManager) clearPermissionCache(userID string) {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()

	delete(sm.permissionCache, userID)
}

// logAudit 记录审计日志
func (sm *SecurityManager) logAudit(userID, pluginID, action, resource, result string, details map[string]interface{}, ipAddress, userAgent string) {
	log := &AuditLog{
		ID:        sm.generateID(),
		UserID:    userID,
		PluginID:  pluginID,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	sm.auditLogs = append(sm.auditLogs, log)
}

// matchAuditFilter 匹配审计日志过滤条件
func (sm *SecurityManager) matchAuditFilter(log *AuditLog, filter map[string]interface{}) bool {
	if userID, ok := filter["user_id"].(string); ok && userID != "" && log.UserID != userID {
		return false
	}

	if pluginID, ok := filter["plugin_id"].(string); ok && pluginID != "" && log.PluginID != pluginID {
		return false
	}

	if action, ok := filter["action"].(string); ok && action != "" && log.Action != action {
		return false
	}

	if resource, ok := filter["resource"].(string); ok && resource != "" && log.Resource != resource {
		return false
	}

	if result, ok := filter["result"].(string); ok && result != "" && log.Result != result {
		return false
	}

	if startTime, ok := filter["start_time"].(time.Time); ok && log.Timestamp.Before(startTime) {
		return false
	}

	if endTime, ok := filter["end_time"].(time.Time); ok && log.Timestamp.After(endTime) {
		return false
	}

	return true
}

// startCleanupTasks 启动清理任务
func (sm *SecurityManager) startCleanupTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanupExpiredTokens()
		sm.cleanupOldAuditLogs()
	}
}

// cleanupExpiredTokens 清理过期令牌
func (sm *SecurityManager) cleanupExpiredTokens() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for tokenID, token := range sm.tokens {
		if now.After(token.ExpiresAt) || token.RevokedAt != nil {
			delete(sm.tokens, tokenID)
		}
	}
}

// cleanupOldAuditLogs 清理旧审计日志
func (sm *SecurityManager) cleanupOldAuditLogs() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-sm.config.AuditLogRetention)
	var filtered []*AuditLog

	for _, log := range sm.auditLogs {
		if log.Timestamp.After(cutoff) {
			filtered = append(filtered, log)
		}
	}

	sm.auditLogs = filtered
}

// GetStats 获取安全统计信息
func (sm *SecurityManager) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	activeTokens := 0
	now := time.Now()
	for _, token := range sm.tokens {
		if token.RevokedAt == nil && now.Before(token.ExpiresAt) {
			activeTokens++
		}
	}

	activeUsers := 0
	for _, user := range sm.users {
		if user.Status == "active" {
			activeUsers++
		}
	}

	return map[string]interface{}{
		"total_users":       len(sm.users),
		"active_users":      activeUsers,
		"total_roles":       len(sm.roles),
		"total_permissions": len(sm.permissions),
		"active_tokens":     activeTokens,
		"total_audit_logs":  len(sm.auditLogs),
		"resource_limits":   len(sm.resourceLimits),
	}
}
