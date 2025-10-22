// Package security 实现插件安全管理功能
package security

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"../core"
)

// DefaultSecurityManager 默认安全管理器实现
type DefaultSecurityManager struct {
	sandboxes map[string]core.Sandbox
	policies  map[string]*core.SecurityPolicy
	mutex     sync.RWMutex
	config    *core.SecurityConfig
}

// NewDefaultSecurityManager 创建新的安全管理器实例
func NewDefaultSecurityManager(config *core.SecurityConfig) *DefaultSecurityManager {
	if config == nil {
		config = &core.SecurityConfig{
			EnableSandbox:          true,
			MaxMemoryPerPlugin:     128 * 1024 * 1024, // 128MB
			MaxCPUTimePerPlugin:    60 * time.Second,
			MaxGoroutinesPerPlugin: 20,
			AllowedNetworkHosts:    []string{},
			BlockedAPIs:            []string{"os.Exit", "os.Kill", "syscall"},
		}
	}

	return &DefaultSecurityManager{
		sandboxes: make(map[string]core.Sandbox),
		policies:  make(map[string]*core.SecurityPolicy),
		config:    config,
	}
}

// CreateSandbox 为插件创建沙箱环境
func (sm *DefaultSecurityManager) CreateSandbox(pluginID string, config *core.SandboxConfig) (core.Sandbox, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 检查是否已存在沙箱
	if _, exists := sm.sandboxes[pluginID]; exists {
		return nil, fmt.Errorf("sandbox already exists for plugin %s", pluginID)
	}

	// 创建沙箱实例
	sandbox := &DefaultSandbox{
		pluginID:      pluginID,
		config:        config,
		securityMgr:   sm,
		resourceUsage: &ResourceUsage{},
		startTime:     time.Now(),
		active:        true,
	}

	// 初始化资源监控
	if err := sandbox.initializeMonitoring(); err != nil {
		return nil, fmt.Errorf("failed to initialize sandbox monitoring: %w", err)
	}

	sm.sandboxes[pluginID] = sandbox
	return sandbox, nil
}

// DestroySandbox 销毁插件沙箱环境
func (sm *DefaultSecurityManager) DestroySandbox(pluginID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sandbox, exists := sm.sandboxes[pluginID]
	if !exists {
		return fmt.Errorf("sandbox not found for plugin %s", pluginID)
	}

	// 停止沙箱
	if err := sandbox.Stop(); err != nil {
		return fmt.Errorf("failed to stop sandbox: %w", err)
	}

	delete(sm.sandboxes, pluginID)
	return nil
}

// ValidatePlugin 验证插件安全配置
func (sm *DefaultSecurityManager) ValidatePlugin(metadata *core.PluginMetadata, manifest []byte) error {
	// 1. 检查插件权限
	if err := sm.validatePermissions(metadata); err != nil {
		return fmt.Errorf("permission validation failed: %w", err)
	}

	// 2. 检查插件签名（如果启用）
	if sm.config.RequireSignature {
		if err := sm.validateSignature(metadata, manifest); err != nil {
			return fmt.Errorf("signature validation failed: %w", err)
		}
	}

	// 3. 检查插件来源（如果启用）
	if sm.config.RequireSourceVerification {
		if err := sm.validateSource(metadata); err != nil {
			return fmt.Errorf("source validation failed: %w", err)
		}
	}

	return nil
}

// SetPolicy 设置插件安全策略
func (sm *DefaultSecurityManager) SetPolicy(pluginID string, policy *core.SecurityPolicy) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.policies[pluginID] = policy
	return nil
}

// GetPolicy 获取插件安全策略
func (sm *DefaultSecurityManager) GetPolicy(pluginID string) (*core.SecurityPolicy, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	policy, exists := sm.policies[pluginID]
	if !exists {
		return nil, fmt.Errorf("policy not found for plugin %s", pluginID)
	}

	return policy, nil
}

// MonitorResources 监控插件资源使用
func (sm *DefaultSecurityManager) MonitorResources(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sm.checkResourceUsage()
		}
	}
}

// checkResourceUsage 检查所有沙箱的资源使用情况
func (sm *DefaultSecurityManager) checkResourceUsage() {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	for pluginID, sandbox := range sm.sandboxes {
		usage := sandbox.GetResourceUsage()

		// 检查内存使用是否超过限制
		if usage.MemoryUsage > sm.config.MaxMemoryPerPlugin {
			// 记录警告或采取措施（如终止插件）
			fmt.Printf("Plugin %s exceeded memory limit: %d bytes\n", pluginID, usage.MemoryUsage)
		}

		// 检查 CPU 时间是否超过限制
		if usage.CPUTime > sm.config.MaxCPUTimePerPlugin {
			fmt.Printf("Plugin %s exceeded CPU time limit: %v\n", pluginID, usage.CPUTime)
		}

		// 检查 Goroutine 数量是否超过限制
		if usage.GoroutineCount > sm.config.MaxGoroutinesPerPlugin {
			fmt.Printf("Plugin %s exceeded goroutine limit: %d\n", pluginID, usage.GoroutineCount)
		}
	}
}

// validatePermissions 验证插件权限
func (sm *DefaultSecurityManager) validatePermissions(metadata *core.PluginMetadata) error {
	// 检查请求的权限是否合理
	for _, permission := range metadata.Permissions {
		if !sm.isPermissionAllowed(permission) {
			return fmt.Errorf("permission not allowed: %s", permission)
		}
	}
	return nil
}

// validateSignature 验证插件签名
func (sm *DefaultSecurityManager) validateSignature(metadata *core.PluginMetadata, manifest []byte) error {
	// 这里应该实现实际的签名验证逻辑
	// 简化实现，实际需要使用加密库
	if metadata.Signature == "" {
		return fmt.Errorf("plugin signature is required")
	}
	return nil
}

// validateSource 验证插件来源
func (sm *DefaultSecurityManager) validateSource(metadata *core.PluginMetadata) error {
	// 检查插件是否来自可信源
	if len(sm.config.TrustedSources) > 0 {
		trusted := false
		for _, source := range sm.config.TrustedSources {
			if metadata.Source == source {
				trusted = true
				break
			}
		}
		if !trusted {
			return fmt.Errorf("plugin source not trusted: %s", metadata.Source)
		}
	}
	return nil
}

// isPermissionAllowed 检查权限是否被允许
func (sm *DefaultSecurityManager) isPermissionAllowed(permission string) bool {
	// 检查是否在黑名单中
	for _, blocked := range sm.config.BlockedPermissions {
		if permission == blocked {
			return false
		}
	}

	// 检查是否在白名单中（如果有白名单）
	if len(sm.config.AllowedPermissions) > 0 {
		for _, allowed := range sm.config.AllowedPermissions {
			if permission == allowed {
				return true
			}
		}
		return false
	}

	return true
}

// DefaultSandbox 默认沙箱实现
type DefaultSandbox struct {
	pluginID      string
	config        *core.SandboxConfig
	securityMgr   *DefaultSecurityManager
	resourceUsage *ResourceUsage
	startTime     time.Time
	active        bool
	mutex         sync.RWMutex
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryUsage    uint64
	CPUTime        time.Duration
	GoroutineCount int
	NetworkCalls   int
	FileOperations int
}

// initializeMonitoring 初始化资源监控
func (sb *DefaultSandbox) initializeMonitoring() error {
	// 启动资源监控 goroutine
	go sb.monitorResources()
	return nil
}

// monitorResources 监控资源使用
func (sb *DefaultSandbox) monitorResources() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for sb.active {
		select {
		case <-ticker.C:
			sb.updateResourceUsage()
		}
	}
}

// updateResourceUsage 更新资源使用情况
func (sb *DefaultSandbox) updateResourceUsage() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	// 获取内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	sb.resourceUsage.MemoryUsage = m.Alloc

	// 获取 Goroutine 数量
	sb.resourceUsage.GoroutineCount = runtime.NumGoroutine()

	// 计算 CPU 时间
	sb.resourceUsage.CPUTime = time.Since(sb.startTime)
}

// Execute 在沙箱中执行代码
func (sb *DefaultSandbox) Execute(ctx context.Context, fn func() error) error {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	if !sb.active {
		return fmt.Errorf("sandbox is not active")
	}

	// 检查资源限制是否被超过
	if err := sb.checkResourceLimits(); err != nil {
		return fmt.Errorf("resource limit exceeded: %w", err)
	}

	// 创建带超时的上下文，以便可以在超时时取消执行	execCtx, cancel := context.WithTimeout(ctx, sb.config.MaxCPUTime)
	defer cancel()

	// 在独立的 goroutine 中执行，以便可以监控和控制执行		errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in sandbox: %v", r)
			}
		}()
		errChan <- fn()
	}()

	select {
	case err := <-errChan:
		return err
	case <-execCtx.Done():
		return fmt.Errorf("execution timeout")
	}
}

// checkResourceLimits 检查资源限制是否被超过
func (sb *DefaultSandbox) checkResourceLimits() error {
	if sb.resourceUsage.MemoryUsage > sb.config.MaxMemory {
		return fmt.Errorf("memory limit exceeded: %d > %d", sb.resourceUsage.MemoryUsage, sb.config.MaxMemory)
	}

	if sb.resourceUsage.GoroutineCount > sb.config.MaxGoroutines {
		return fmt.Errorf("goroutine limit exceeded: %d > %d", sb.resourceUsage.GoroutineCount, sb.config.MaxGoroutines)
	}

	return nil
}

// GetResourceUsage 获取资源使用情况
func (sb *DefaultSandbox) GetResourceUsage() *ResourceUsage {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	// 返回副本以避免并发修改导致的问题
	return &ResourceUsage{
		MemoryUsage:    sb.resourceUsage.MemoryUsage,
		CPUTime:        sb.resourceUsage.CPUTime,
		GoroutineCount: sb.resourceUsage.GoroutineCount,
		NetworkCalls:   sb.resourceUsage.NetworkCalls,
		FileOperations: sb.resourceUsage.FileOperations,
	}
}

// Stop 停止沙箱
func (sb *DefaultSandbox) Stop() error {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	sb.active = false
	return nil
}

// IsActive 检查沙箱是否活动
func (sb *DefaultSandbox) IsActive() bool {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	return sb.active
}
