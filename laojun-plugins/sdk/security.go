package sdk

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SecurityManager 安全管理器接口
type SecurityManager interface {
	// ValidatePlugin 验证插件
	ValidatePlugin(plugin Plugin) error
	
	// CheckPermission 检查权限
	CheckPermission(pluginID string, permission Permission) error
	
	// GrantPermission 授予权限
	GrantPermission(pluginID string, permission Permission) error
	
	// RevokePermission 撤销权限
	RevokePermission(pluginID string, permission Permission) error
	
	// GetPermissions 获取插件权限
	GetPermissions(pluginID string) []Permission
	
	// CreateSecurityContext 创建安全上下文
	CreateSecurityContext(pluginID string) (*SecurityContext, error)
	
	// ValidateSecurityPolicy 验证安全策略
	ValidateSecurityPolicy(policy *SecurityPolicy) error
	
	// ApplySecurityPolicy 应用安全策略
	ApplySecurityPolicy(pluginID string, policy *SecurityPolicy) error
}

// Permission 权限定义
type Permission struct {
	Type        PermissionType `json:"type"`
	Resource    string         `json:"resource"`
	Action      string         `json:"action"`
	Description string         `json:"description,omitempty"`
	Level       PermissionLevel `json:"level"`
}

// PermissionType 权限类型
type PermissionType string

const (
	PermissionTypeFile    PermissionType = "file"
	PermissionTypeNetwork PermissionType = "network"
	PermissionTypeSystem  PermissionType = "system"
	PermissionTypeAPI     PermissionType = "api"
	PermissionTypeData    PermissionType = "data"
	PermissionTypeConfig  PermissionType = "config"
)

// PermissionLevel 权限级别
type PermissionLevel string

const (
	PermissionLevelNone   PermissionLevel = "none"
	PermissionLevelRead   PermissionLevel = "read"
	PermissionLevelWrite  PermissionLevel = "write"
	PermissionLevelAdmin  PermissionLevel = "admin"
)

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Version           string                 `json:"version"`
	Permissions       []Permission           `json:"permissions"`
	Restrictions      *SecurityRestrictions  `json:"restrictions,omitempty"`
	Sandbox           *SandboxConfig         `json:"sandbox,omitempty"`
	Validation        *ValidationConfig      `json:"validation,omitempty"`
	Monitoring        *MonitoringConfig      `json:"monitoring,omitempty"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
}

// SecurityRestrictions 安全限制
type SecurityRestrictions struct {
	AllowedPaths      []string `json:"allowedPaths,omitempty"`
	DeniedPaths       []string `json:"deniedPaths,omitempty"`
	AllowedHosts      []string `json:"allowedHosts,omitempty"`
	DeniedHosts       []string `json:"deniedHosts,omitempty"`
	AllowedPorts      []int    `json:"allowedPorts,omitempty"`
	DeniedPorts       []int    `json:"deniedPorts,omitempty"`
	MaxFileSize       int64    `json:"maxFileSize,omitempty"`
	MaxNetworkConns   int      `json:"maxNetworkConns,omitempty"`
	MaxExecutionTime  int      `json:"maxExecutionTime,omitempty"`
	AllowShellAccess  bool     `json:"allowShellAccess"`
	AllowNetworkAccess bool    `json:"allowNetworkAccess"`
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	Enabled         bool              `json:"enabled"`
	Type            string            `json:"type"`
	RootPath        string            `json:"rootPath,omitempty"`
	ReadOnlyPaths   []string          `json:"readOnlyPaths,omitempty"`
	TempPath        string            `json:"tempPath,omitempty"`
	Environment     map[string]string `json:"environment,omitempty"`
	ResourceLimits  *ResourceLimits   `json:"resourceLimits,omitempty"`
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	RequireSignature    bool     `json:"requireSignature"`
	TrustedSigners      []string `json:"trustedSigners,omitempty"`
	RequireChecksum     bool     `json:"requireChecksum"`
	AllowedExtensions   []string `json:"allowedExtensions,omitempty"`
	DeniedExtensions    []string `json:"deniedExtensions,omitempty"`
	MaxPluginSize       int64    `json:"maxPluginSize,omitempty"`
	ScanForMalware      bool     `json:"scanForMalware"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	LogLevel            string `json:"logLevel"`
	LogSensitiveData    bool   `json:"logSensitiveData"`
	MonitorFileAccess   bool   `json:"monitorFileAccess"`
	MonitorNetworkAccess bool  `json:"monitorNetworkAccess"`
	MonitorSystemCalls  bool   `json:"monitorSystemCalls"`
	AlertOnViolation    bool   `json:"alertOnViolation"`
}

// DefaultSecurityManager 默认安全管理器实现
type DefaultSecurityManager struct {
	permissions map[string][]Permission
	policies    map[string]*SecurityPolicy
	contexts    map[string]*SecurityContext
	logger      *logrus.Logger
	mu          sync.RWMutex
}

// NewDefaultSecurityManager 创建默认安全管理器
func NewDefaultSecurityManager(logger *logrus.Logger) *DefaultSecurityManager {
	return &DefaultSecurityManager{
		permissions: make(map[string][]Permission),
		policies:    make(map[string]*SecurityPolicy),
		contexts:    make(map[string]*SecurityContext),
		logger:      logger,
	}
}

// ValidatePlugin 验证插件
func (sm *DefaultSecurityManager) ValidatePlugin(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	info := plugin.GetInfo()
	if info == nil {
		return fmt.Errorf("plugin info cannot be nil")
	}

	// 验证插件ID
	if err := sm.validatePluginID(info.ID); err != nil {
		return fmt.Errorf("invalid plugin ID: %w", err)
	}

	// 验证插件名称
	if err := sm.validatePluginName(info.Name); err != nil {
		return fmt.Errorf("invalid plugin name: %w", err)
	}

	// 验证版本
	if err := sm.validateVersion(info.Version); err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	// 验证权限
	if info.Manifest != nil && len(info.Manifest.Permissions) > 0 {
		for _, perm := range info.Manifest.Permissions {
			if err := sm.validatePermission(perm); err != nil {
				return fmt.Errorf("invalid permission %s: %w", perm.Type, err)
			}
		}
	}

	sm.logger.WithField("plugin_id", info.ID).Debug("Plugin validation passed")
	return nil
}

// validatePluginID 验证插件ID
func (sm *DefaultSecurityManager) validatePluginID(id string) error {
	if id == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	// 插件ID只能包含字母、数字、连字符和下划线
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, id)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("plugin ID can only contain letters, numbers, hyphens and underscores")
	}

	// 长度限制
	if len(id) > 64 {
		return fmt.Errorf("plugin ID cannot exceed 64 characters")
	}

	return nil
}

// validatePluginName 验证插件名称
func (sm *DefaultSecurityManager) validatePluginName(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	// 长度限制
	if len(name) > 128 {
		return fmt.Errorf("plugin name cannot exceed 128 characters")
	}

	return nil
}

// validateVersion 验证版本
func (sm *DefaultSecurityManager) validateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// 简单的语义版本验证
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`, version)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("version must follow semantic versioning format (x.y.z)")
	}

	return nil
}

// validatePermission 验证权限
func (sm *DefaultSecurityManager) validatePermission(perm Permission) error {
	// 验证权限类型
	switch perm.Type {
	case PermissionTypeFile, PermissionTypeNetwork, PermissionTypeSystem, 
		 PermissionTypeAPI, PermissionTypeData, PermissionTypeConfig:
		// 有效类型
	default:
		return fmt.Errorf("invalid permission type: %s", perm.Type)
	}

	// 验证权限级别
	switch perm.Level {
	case PermissionLevelNone, PermissionLevelRead, PermissionLevelWrite, PermissionLevelAdmin:
		// 有效级别
	default:
		return fmt.Errorf("invalid permission level: %s", perm.Level)
	}

	// 验证资源
	if perm.Resource == "" {
		return fmt.Errorf("permission resource cannot be empty")
	}

	return nil
}

// CheckPermission 检查权限
func (sm *DefaultSecurityManager) CheckPermission(pluginID string, permission Permission) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	permissions, exists := sm.permissions[pluginID]
	if !exists {
		return fmt.Errorf("no permissions found for plugin %s", pluginID)
	}

	// 检查是否有匹配的权限
	for _, perm := range permissions {
		if sm.permissionMatches(perm, permission) {
			return nil
		}
	}

	return fmt.Errorf("permission denied: %s %s on %s", permission.Type, permission.Action, permission.Resource)
}

// permissionMatches 检查权限是否匹配
func (sm *DefaultSecurityManager) permissionMatches(granted, requested Permission) bool {
	// 类型必须匹配
	if granted.Type != requested.Type {
		return false
	}

	// 检查资源匹配
	if !sm.resourceMatches(granted.Resource, requested.Resource) {
		return false
	}

	// 检查动作匹配
	if granted.Action != "*" && granted.Action != requested.Action {
		return false
	}

	// 检查权限级别
	return sm.levelSufficient(granted.Level, requested.Level)
}

// resourceMatches 检查资源是否匹配
func (sm *DefaultSecurityManager) resourceMatches(granted, requested string) bool {
	if granted == "*" {
		return true
	}

	if granted == requested {
		return true
	}

	// 支持通配符匹配
	if strings.HasSuffix(granted, "*") {
		prefix := strings.TrimSuffix(granted, "*")
		return strings.HasPrefix(requested, prefix)
	}

	return false
}

// levelSufficient 检查权限级别是否足够
func (sm *DefaultSecurityManager) levelSufficient(granted, requested PermissionLevel) bool {
	levelOrder := map[PermissionLevel]int{
		PermissionLevelNone:  0,
		PermissionLevelRead:  1,
		PermissionLevelWrite: 2,
		PermissionLevelAdmin: 3,
	}

	grantedLevel, grantedExists := levelOrder[granted]
	requestedLevel, requestedExists := levelOrder[requested]

	if !grantedExists || !requestedExists {
		return false
	}

	return grantedLevel >= requestedLevel
}

// GrantPermission 授予权限
func (sm *DefaultSecurityManager) GrantPermission(pluginID string, permission Permission) error {
	if err := sm.validatePermission(permission); err != nil {
		return fmt.Errorf("invalid permission: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.permissions[pluginID] == nil {
		sm.permissions[pluginID] = []Permission{}
	}

	// 检查权限是否已存在
	for _, perm := range sm.permissions[pluginID] {
		if sm.permissionEquals(perm, permission) {
			return nil // 权限已存在
		}
	}

	sm.permissions[pluginID] = append(sm.permissions[pluginID], permission)

	sm.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"permission": permission,
	}).Info("Permission granted")

	return nil
}

// permissionEquals 检查权限是否相等
func (sm *DefaultSecurityManager) permissionEquals(a, b Permission) bool {
	return a.Type == b.Type && a.Resource == b.Resource && a.Action == b.Action && a.Level == b.Level
}

// RevokePermission 撤销权限
func (sm *DefaultSecurityManager) RevokePermission(pluginID string, permission Permission) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	permissions, exists := sm.permissions[pluginID]
	if !exists {
		return fmt.Errorf("no permissions found for plugin %s", pluginID)
	}

	// 查找并移除权限
	for i, perm := range permissions {
		if sm.permissionEquals(perm, permission) {
			sm.permissions[pluginID] = append(permissions[:i], permissions[i+1:]...)
			
			sm.logger.WithFields(logrus.Fields{
				"plugin_id":  pluginID,
				"permission": permission,
			}).Info("Permission revoked")
			
			return nil
		}
	}

	return fmt.Errorf("permission not found")
}

// GetPermissions 获取插件权限
func (sm *DefaultSecurityManager) GetPermissions(pluginID string) []Permission {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	permissions, exists := sm.permissions[pluginID]
	if !exists {
		return []Permission{}
	}

	// 返回副本
	result := make([]Permission, len(permissions))
	copy(result, permissions)
	return result
}

// CreateSecurityContext 创建安全上下文
func (sm *DefaultSecurityManager) CreateSecurityContext(pluginID string) (*SecurityContext, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查是否已存在
	if ctx, exists := sm.contexts[pluginID]; exists {
		return ctx, nil
	}

	// 创建新的安全上下文
	ctx := &SecurityContext{
		PluginID:    pluginID,
		SessionID:   sm.generateSessionID(),
		Permissions: sm.GetPermissions(pluginID),
		CreatedAt:   time.Now(),
		LastAccess:  time.Now(),
		Active:      true,
	}

	// 应用安全策略
	if policy, exists := sm.policies[pluginID]; exists {
		ctx.Policy = policy
	}

	sm.contexts[pluginID] = ctx

	sm.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"session_id": ctx.SessionID,
	}).Debug("Security context created")

	return ctx, nil
}

// generateSessionID 生成会话ID
func (sm *DefaultSecurityManager) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateSecurityPolicy 验证安全策略
func (sm *DefaultSecurityManager) ValidateSecurityPolicy(policy *SecurityPolicy) error {
	if policy == nil {
		return fmt.Errorf("security policy cannot be nil")
	}

	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	// 验证权限
	for _, perm := range policy.Permissions {
		if err := sm.validatePermission(perm); err != nil {
			return fmt.Errorf("invalid permission in policy: %w", err)
		}
	}

	// 验证限制
	if policy.Restrictions != nil {
		if err := sm.validateRestrictions(policy.Restrictions); err != nil {
			return fmt.Errorf("invalid restrictions in policy: %w", err)
		}
	}

	return nil
}

// validateRestrictions 验证安全限制
func (sm *DefaultSecurityManager) validateRestrictions(restrictions *SecurityRestrictions) error {
	// 验证路径
	for _, path := range restrictions.AllowedPaths {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("allowed path must be absolute: %s", path)
		}
	}

	for _, path := range restrictions.DeniedPaths {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("denied path must be absolute: %s", path)
		}
	}

	// 验证主机
	for _, host := range restrictions.AllowedHosts {
		if net.ParseIP(host) == nil && !sm.isValidHostname(host) {
			return fmt.Errorf("invalid allowed host: %s", host)
		}
	}

	for _, host := range restrictions.DeniedHosts {
		if net.ParseIP(host) == nil && !sm.isValidHostname(host) {
			return fmt.Errorf("invalid denied host: %s", host)
		}
	}

	// 验证端口
	for _, port := range restrictions.AllowedPorts {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid allowed port: %d", port)
		}
	}

	for _, port := range restrictions.DeniedPorts {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid denied port: %d", port)
		}
	}

	return nil
}

// isValidHostname 检查是否为有效的主机名
func (sm *DefaultSecurityManager) isValidHostname(hostname string) bool {
	matched, err := regexp.MatchString(`^[a-zA-Z0-9.-]+$`, hostname)
	return err == nil && matched
}

// ApplySecurityPolicy 应用安全策略
func (sm *DefaultSecurityManager) ApplySecurityPolicy(pluginID string, policy *SecurityPolicy) error {
	if err := sm.ValidateSecurityPolicy(policy); err != nil {
		return fmt.Errorf("invalid security policy: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 保存策略
	sm.policies[pluginID] = policy

	// 更新权限
	sm.permissions[pluginID] = policy.Permissions

	// 更新安全上下文
	if ctx, exists := sm.contexts[pluginID]; exists {
		ctx.Policy = policy
		ctx.Permissions = policy.Permissions
		ctx.LastAccess = time.Now()
	}

	sm.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"policy_id": policy.ID,
	}).Info("Security policy applied")

	return nil
}

// SecurityValidator 安全验证器
type SecurityValidator struct {
	manager *DefaultSecurityManager
	logger  *logrus.Logger
}

// NewSecurityValidator 创建安全验证器
func NewSecurityValidator(manager *DefaultSecurityManager, logger *logrus.Logger) *SecurityValidator {
	return &SecurityValidator{
		manager: manager,
		logger:  logger,
	}
}

// ValidateFileAccess 验证文件访问
func (sv *SecurityValidator) ValidateFileAccess(pluginID, filePath, action string) error {
	permission := Permission{
		Type:     PermissionTypeFile,
		Resource: filePath,
		Action:   action,
		Level:    sv.getRequiredLevelForAction(action),
	}

	return sv.manager.CheckPermission(pluginID, permission)
}

// ValidateNetworkAccess 验证网络访问
func (sv *SecurityValidator) ValidateNetworkAccess(pluginID, host string, port int, action string) error {
	resource := fmt.Sprintf("%s:%d", host, port)
	permission := Permission{
		Type:     PermissionTypeNetwork,
		Resource: resource,
		Action:   action,
		Level:    sv.getRequiredLevelForAction(action),
	}

	return sv.manager.CheckPermission(pluginID, permission)
}

// ValidateAPIAccess 验证API访问
func (sv *SecurityValidator) ValidateAPIAccess(pluginID, apiPath, method string) error {
	permission := Permission{
		Type:     PermissionTypeAPI,
		Resource: apiPath,
		Action:   method,
		Level:    sv.getRequiredLevelForAction(method),
	}

	return sv.manager.CheckPermission(pluginID, permission)
}

// getRequiredLevelForAction 根据动作获取所需权限级别
func (sv *SecurityValidator) getRequiredLevelForAction(action string) PermissionLevel {
	switch strings.ToLower(action) {
	case "read", "get", "list":
		return PermissionLevelRead
	case "write", "create", "update", "post", "put", "patch":
		return PermissionLevelWrite
	case "delete", "admin":
		return PermissionLevelAdmin
	default:
		return PermissionLevelRead
	}
}

// SecurityAuditor 安全审计器
type SecurityAuditor struct {
	logger *logrus.Logger
}

// NewSecurityAuditor 创建安全审计器
func NewSecurityAuditor(logger *logrus.Logger) *SecurityAuditor {
	return &SecurityAuditor{
		logger: logger,
	}
}

// AuditPermissionCheck 审计权限检查
func (sa *SecurityAuditor) AuditPermissionCheck(pluginID string, permission Permission, result error) {
	status := "granted"
	if result != nil {
		status = "denied"
	}

	sa.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"permission": permission,
		"status":     status,
		"error":      result,
		"timestamp":  time.Now(),
	}).Info("Permission check audited")
}

// AuditSecurityViolation 审计安全违规
func (sa *SecurityAuditor) AuditSecurityViolation(pluginID, violation, details string) {
	sa.logger.WithFields(logrus.Fields{
		"plugin_id": pluginID,
		"violation": violation,
		"details":   details,
		"timestamp": time.Now(),
		"severity":  "high",
	}).Warn("Security violation detected")
}

// AuditPolicyChange 审计策略变更
func (sa *SecurityAuditor) AuditPolicyChange(pluginID string, oldPolicy, newPolicy *SecurityPolicy) {
	sa.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"old_policy": oldPolicy,
		"new_policy": newPolicy,
		"timestamp":  time.Now(),
	}).Info("Security policy changed")
}