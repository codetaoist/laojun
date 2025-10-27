package runtime

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Sandbox 沙箱接口
type Sandbox interface {
	// InitializePlugin 在沙箱中初始化插件
	InitializePlugin(ctx context.Context, plugin Plugin) error

	// RemovePlugin 从沙箱中移除插件
	RemovePlugin(ctx context.Context, pluginID string) error

	// GetResourceLimits 获取资源限制
	GetResourceLimits(pluginID string) (*ResourceLimits, error)

	// SetResourceLimits 设置资源限制
	SetResourceLimits(pluginID string, limits *ResourceLimits) error

	// CheckPermission 检查权限
	CheckPermission(pluginID string, permission string) bool

	// GetSecurityContext 获取安全上下文
	GetSecurityContext(pluginID string) (*SecurityContext, error)
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxMemoryBytes   uint64        `json:"max_memory_bytes"`
	MaxCPUPercent    float64       `json:"max_cpu_percent"`
	MaxGoroutines    int           `json:"max_goroutines"`
	MaxFileHandles   int           `json:"max_file_handles"`
	MaxNetworkConns  int           `json:"max_network_conns"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
}

// SecurityContext 安全上下文
type SecurityContext struct {
	PluginID     string            `json:"plugin_id"`
	Permissions  []string          `json:"permissions"`
	AllowedPaths []string          `json:"allowed_paths"`
	AllowedHosts []string          `json:"allowed_hosts"`
	Environment  map[string]string `json:"environment"`
	CreatedAt    time.Time         `json:"created_at"`
}

// DefaultSandbox 默认沙箱实现
type DefaultSandbox struct {
	contexts       map[string]*SecurityContext
	resourceLimits map[string]*ResourceLimits
	logger         *logrus.Logger
	mu             sync.RWMutex
}

// NewDefaultSandbox 创建默认沙箱
func NewDefaultSandbox(logger *logrus.Logger) *DefaultSandbox {
	return &DefaultSandbox{
		contexts:       make(map[string]*SecurityContext),
		resourceLimits: make(map[string]*ResourceLimits),
		logger:         logger,
	}
}

// InitializePlugin 在沙箱中初始化插件
func (s *DefaultSandbox) InitializePlugin(ctx context.Context, plugin Plugin) error {
	metadata := plugin.GetMetadata()
	pluginID := metadata.ID

	s.logger.WithField("plugin_id", pluginID).Info("Initializing plugin in sandbox")

	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建安全上下文
	securityContext := &SecurityContext{
		PluginID:    pluginID,
		Permissions: metadata.Permissions,
		AllowedPaths: []string{
			"/tmp/plugins/" + pluginID,
			"/var/log/plugins/" + pluginID,
		},
		AllowedHosts: []string{
			"localhost",
			"127.0.0.1",
		},
		Environment: map[string]string{
			"PLUGIN_ID":   pluginID,
			"PLUGIN_NAME": metadata.Name,
		},
		CreatedAt: time.Now(),
	}

	// 设置默认资源限制
	resourceLimits := &ResourceLimits{
		MaxMemoryBytes:   100 * 1024 * 1024, // 100MB
		MaxCPUPercent:    10.0,               // 10%
		MaxGoroutines:    50,
		MaxFileHandles:   20,
		MaxNetworkConns:  10,
		ExecutionTimeout: 30 * time.Second,
	}

	// 根据插件类别调整资源限制
	switch metadata.Category {
	case "system":
		resourceLimits.MaxMemoryBytes = 200 * 1024 * 1024 // 200MB
		resourceLimits.MaxCPUPercent = 20.0                // 20%
	case "analytics":
		resourceLimits.MaxMemoryBytes = 500 * 1024 * 1024 // 500MB
		resourceLimits.MaxCPUPercent = 30.0                // 30%
	}

	s.contexts[pluginID] = securityContext
	s.resourceLimits[pluginID] = resourceLimits

	// 初始化插件
	config := make(map[string]interface{})
	config["security_context"] = securityContext
	config["resource_limits"] = resourceLimits

	if err := plugin.Initialize(ctx, config); err != nil {
		// 清理已创建的上下文
		delete(s.contexts, pluginID)
		delete(s.resourceLimits, pluginID)
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	s.logger.WithField("plugin_id", pluginID).Info("Plugin initialized in sandbox successfully")
	return nil
}

// RemovePlugin 从沙箱中移除插件
func (s *DefaultSandbox) RemovePlugin(ctx context.Context, pluginID string) error {
	s.logger.WithField("plugin_id", pluginID).Info("Removing plugin from sandbox")

	s.mu.Lock()
	defer s.mu.Unlock()

	// 移除安全上下文和资源限制
	delete(s.contexts, pluginID)
	delete(s.resourceLimits, pluginID)

	s.logger.WithField("plugin_id", pluginID).Info("Plugin removed from sandbox successfully")
	return nil
}

// GetResourceLimits 获取资源限制
func (s *DefaultSandbox) GetResourceLimits(pluginID string) (*ResourceLimits, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limits, exists := s.resourceLimits[pluginID]
	if !exists {
		return nil, fmt.Errorf("resource limits not found for plugin %s", pluginID)
	}

	// 返回副本
	limitsCopy := *limits
	return &limitsCopy, nil
}

// SetResourceLimits 设置资源限制
func (s *DefaultSandbox) SetResourceLimits(pluginID string, limits *ResourceLimits) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.contexts[pluginID]; !exists {
		return fmt.Errorf("plugin %s not found in sandbox", pluginID)
	}

	// 验证资源限制
	if err := s.validateResourceLimits(limits); err != nil {
		return fmt.Errorf("invalid resource limits: %w", err)
	}

	s.resourceLimits[pluginID] = limits
	s.logger.WithField("plugin_id", pluginID).Info("Resource limits updated")

	return nil
}

// validateResourceLimits 验证资源限制
func (s *DefaultSandbox) validateResourceLimits(limits *ResourceLimits) error {
	if limits.MaxMemoryBytes == 0 {
		return fmt.Errorf("max memory bytes must be greater than 0")
	}
	if limits.MaxCPUPercent <= 0 || limits.MaxCPUPercent > 100 {
		return fmt.Errorf("max CPU percent must be between 0 and 100")
	}
	if limits.MaxGoroutines <= 0 {
		return fmt.Errorf("max goroutines must be greater than 0")
	}
	if limits.ExecutionTimeout <= 0 {
		return fmt.Errorf("execution timeout must be greater than 0")
	}

	return nil
}

// CheckPermission 检查权限
func (s *DefaultSandbox) CheckPermission(pluginID string, permission string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	context, exists := s.contexts[pluginID]
	if !exists {
		return false
	}

	// 检查权限列表
	for _, p := range context.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}

	return false
}

// GetSecurityContext 获取安全上下文
func (s *DefaultSandbox) GetSecurityContext(pluginID string) (*SecurityContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	context, exists := s.contexts[pluginID]
	if !exists {
		return nil, fmt.Errorf("security context not found for plugin %s", pluginID)
	}

	// 返回副本
	contextCopy := *context
	return &contextCopy, nil
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	sandbox *DefaultSandbox
	logger  *logrus.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewResourceMonitor 创建资源监控器
func NewResourceMonitor(sandbox *DefaultSandbox, logger *logrus.Logger) *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ResourceMonitor{
		sandbox: sandbox,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动资源监控
func (m *ResourceMonitor) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
				m.checkResourceUsage()
			}
		}
	}()
}

// Stop 停止资源监控
func (m *ResourceMonitor) Stop() {
	m.cancel()
	m.wg.Wait()
}

// checkResourceUsage 检查资源使用情况
func (m *ResourceMonitor) checkResourceUsage() {
	m.sandbox.mu.RLock()
	pluginIDs := make([]string, 0, len(m.sandbox.contexts))
	for pluginID := range m.sandbox.contexts {
		pluginIDs = append(pluginIDs, pluginID)
	}
	m.sandbox.mu.RUnlock()

	for _, pluginID := range pluginIDs {
		limits, err := m.sandbox.GetResourceLimits(pluginID)
		if err != nil {
			continue
		}

		// 检查内存使用
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		
		// 这里应该实现实际的资源使用检查
		// 为了简化，我们只记录日志
		if memStats.Alloc > limits.MaxMemoryBytes {
			m.logger.WithFields(logrus.Fields{
				"plugin_id":    pluginID,
				"memory_used":  memStats.Alloc,
				"memory_limit": limits.MaxMemoryBytes,
			}).Warn("Plugin memory usage exceeds limit")
		}

		// 检查goroutine数量
		goroutineCount := runtime.NumGoroutine()
		if goroutineCount > limits.MaxGoroutines {
			m.logger.WithFields(logrus.Fields{
				"plugin_id":        pluginID,
				"goroutine_count":  goroutineCount,
				"goroutine_limit":  limits.MaxGoroutines,
			}).Warn("Plugin goroutine count exceeds limit")
		}
	}
}