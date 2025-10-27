package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/config"
	"github.com/codetaoist/laojun-discovery/internal/registry"
	"github.com/codetaoist/laojun-discovery/internal/storage"
	"go.uber.org/zap"
	
	// 使用统一的健康检查接口
	"github.com/taishanglaojun/laojun/laojun-shared/health"
)

// DiscoveryHealthManager 服务发现健康管理器
type DiscoveryHealthManager struct {
	manager  health.HealthManager
	config   *config.HealthConfig
	registry *registry.ServiceRegistry
	logger   *zap.Logger
	client   *http.Client
}

// NewDiscoveryHealthManager 创建服务发现健康管理器
func NewDiscoveryHealthManager(config *config.HealthConfig, registry *registry.ServiceRegistry, logger *zap.Logger) *DiscoveryHealthManager {
	healthConfig := &health.HealthConfig{
		EnableCache:    true,
		CacheTTL:       30 * time.Second,
		EnableMetrics:  true,
		DefaultTimeout: time.Duration(config.Timeout) * time.Second,
	}
	
	manager := health.NewDefaultHealthManager(healthConfig, logger)
	
	return &DiscoveryHealthManager{
		manager:  manager,
		config:   config,
		registry: registry,
		logger:   logger,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// RegisterSystemCheck 注册系统健康检查
func (dhm *DiscoveryHealthManager) RegisterSystemCheck() {
	config := &health.CheckerConfig{
		Name:        "discovery_system",
		Type:        health.CheckerTypeSystem,
		Priority:    health.PriorityMedium,
		Timeout:     5 * time.Second,
		Interval:    60 * time.Second,
		Enabled:     true,
		Retries:     2,
		RetryDelay:  time.Second,
	}
	
	systemConfig := &health.SystemConfig{
		CPUThreshold:    80.0,
		MemoryThreshold: 85.0,
		DiskThreshold:   90.0,
	}
	
	checker := health.NewSystemChecker(config, systemConfig)
	dhm.manager.AddChecker(checker)
}

// RegisterApplicationCheck 注册应用程序健康检查
func (dhm *DiscoveryHealthManager) RegisterApplicationCheck() {
	config := &health.CheckerConfig{
		Name:        "discovery_application",
		Type:        health.CheckerTypeApplication,
		Priority:    health.PriorityHigh,
		Timeout:     3 * time.Second,
		Interval:    30 * time.Second,
		Enabled:     true,
		Retries:     2,
		RetryDelay:  time.Second,
	}
	
	// 自定义应用程序检查逻辑
	checkFunc := func(ctx context.Context) *health.CheckResult {
		// 检查服务注册表状态
		stats := dhm.GetRegistryStats(ctx)
		
		return &health.CheckResult{
			Status:    health.StatusHealthy,
			Message:   "Discovery service is running normally",
			Timestamp: time.Now(),
			Details:   stats,
		}
	}
	
	checker := health.NewApplicationChecker(config, checkFunc)
	dhm.manager.AddChecker(checker)
}

// RegisterServiceCheck 注册服务健康检查
func (dhm *DiscoveryHealthManager) RegisterServiceCheck(serviceID, checkType, target string, interval, timeout time.Duration) {
	config := &health.CheckerConfig{
		Name:        fmt.Sprintf("service_%s", serviceID),
		Type:        health.CheckerTypeHTTP, // 默认使用HTTP类型
		Priority:    health.PriorityMedium,
		Timeout:     timeout,
		Interval:    interval,
		Enabled:     true,
		Retries:     3,
		RetryDelay:  time.Second,
	}
	
	var checker health.HealthChecker
	
	switch checkType {
	case "http":
		httpConfig := &health.HTTPConfig{
			URL:    target,
			Method: "GET",
			ExpectedStatusCodes: []int{200, 201, 204},
			Headers: map[string]string{
				"User-Agent": "discovery-health-check",
			},
		}
		checker = health.NewEnhancedHTTPChecker(config, httpConfig)
		
	case "tcp":
		// 为TCP检查创建自定义检查器
		checkFunc := func(ctx context.Context) *health.CheckResult {
			return dhm.checkTCP(ctx, target)
		}
		checker = health.NewApplicationChecker(config, checkFunc)
		
	default:
		dhm.logger.Warn("Unknown check type, using HTTP as default",
			zap.String("service_id", serviceID),
			zap.String("check_type", checkType))
		
		httpConfig := &health.HTTPConfig{
			URL:    target,
			Method: "GET",
			ExpectedStatusCodes: []int{200, 201, 204},
		}
		checker = health.NewEnhancedHTTPChecker(config, httpConfig)
	}
	
	dhm.manager.AddChecker(checker)
	
	dhm.logger.Info("Service health check registered",
		zap.String("service_id", serviceID),
		zap.String("type", checkType),
		zap.String("target", target))
}

// RemoveServiceCheck 移除服务健康检查
func (dhm *DiscoveryHealthManager) RemoveServiceCheck(serviceID string) {
	checkerName := fmt.Sprintf("service_%s", serviceID)
	dhm.manager.RemoveChecker(checkerName)
	
	dhm.logger.Info("Service health check removed",
		zap.String("service_id", serviceID))
}

// checkTCP 执行TCP健康检查
func (dhm *DiscoveryHealthManager) checkTCP(ctx context.Context, address string) *health.CheckResult {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return &health.CheckResult{
			Status:    health.StatusUnhealthy,
			Message:   fmt.Sprintf("TCP connection failed: %v", err),
			Timestamp: time.Now(),
		}
	}
	defer conn.Close()

	return &health.CheckResult{
		Status:    health.StatusHealthy,
		Message:   "TCP connection successful",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"address": address,
			"type":    "tcp",
		},
	}
}

// GetRegistryStats 获取注册表统计信息
func (dhm *DiscoveryHealthManager) GetRegistryStats(ctx context.Context) map[string]interface{} {
	services, err := dhm.registry.ListAllServices(ctx)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	totalServices := len(services)
	totalInstances := 0
	healthyInstances := 0
	
	for _, instances := range services {
		totalInstances += len(instances)
		for _, instance := range instances {
			if instance.Health.Status == "passing" {
				healthyInstances++
			}
		}
	}
	
	return map[string]interface{}{
		"total_services":     totalServices,
		"total_instances":    totalInstances,
		"healthy_instances":  healthyInstances,
		"unhealthy_instances": totalInstances - healthyInstances,
		"health_check_enabled": dhm.config.Enabled,
	}
}

// CheckHealth 执行健康检查
func (dhm *DiscoveryHealthManager) CheckHealth(ctx context.Context) *health.HealthReport {
	return dhm.manager.CheckHealth(ctx)
}

// CheckHealthByType 按类型执行健康检查
func (dhm *DiscoveryHealthManager) CheckHealthByType(ctx context.Context, checkerType health.CheckerType) *health.HealthReport {
	return dhm.manager.CheckHealthByType(ctx, checkerType)
}

// GetSummary 获取健康检查摘要
func (dhm *DiscoveryHealthManager) GetSummary(ctx context.Context) *health.Summary {
	return dhm.manager.GetSummary(ctx)
}

// Start 启动健康检查管理器
func (dhm *DiscoveryHealthManager) Start(ctx context.Context) error {
	if !dhm.config.Enabled {
		dhm.logger.Info("Health checker is disabled")
		return nil
	}

	dhm.logger.Info("Starting discovery health manager",
		zap.Int("check_interval", dhm.config.CheckInterval),
		zap.Int("timeout", dhm.config.Timeout))

	return dhm.manager.Start(ctx)
}

// Stop 停止健康检查管理器
func (dhm *DiscoveryHealthManager) Stop(ctx context.Context) error {
	return dhm.manager.Stop(ctx)
}

// 以下代码保留用于向后兼容，但建议使用新的DiscoveryHealthManager

// CheckState 检查状态 (已弃用，请使用 health.CheckResult)
type CheckState struct {
	ServiceID        string
	CheckType        string // http, tcp
	Target           string
	Interval         time.Duration
	Timeout          time.Duration
	FailureCount     int
	SuccessCount     int
	LastCheck        time.Time
	LastStatus       string
	LastOutput       string
}

// Checker 健康检查器 (已弃用，请使用 DiscoveryHealthManager)
type Checker struct {
	config   *config.HealthConfig
	registry *registry.ServiceRegistry
	logger   *zap.Logger
	client   *http.Client
}

// NewChecker 创建健康检查器 (已弃用，请使用 NewDiscoveryHealthManager)
func NewChecker(config *config.HealthConfig, registry *registry.ServiceRegistry, logger *zap.Logger) *Checker {
	return &Checker{
		config:   config,
		registry: registry,
		logger:   logger,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}