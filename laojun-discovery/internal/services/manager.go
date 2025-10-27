package services

import (
	"context"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/config"
	"github.com/codetaoist/laojun-discovery/internal/health"
	"github.com/codetaoist/laojun-discovery/internal/registry"
	"go.uber.org/zap"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	config         *config.Config
	registry       *registry.ServiceRegistry
	healthChecker  *health.Checker
	logger         *zap.Logger
	
	// 控制通道
	stopCh         chan struct{}
	cleanupCtx     context.Context
	cleanupCancel  context.CancelFunc
}

// NewServiceManager 创建服务管理器
func NewServiceManager(config *config.Config, registry *registry.ServiceRegistry, logger *zap.Logger) *ServiceManager {
	healthChecker := health.NewChecker(&config.Health, registry, logger)
	
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	
	return &ServiceManager{
		config:        config,
		registry:      registry,
		healthChecker: healthChecker,
		logger:        logger,
		stopCh:        make(chan struct{}),
		cleanupCtx:    cleanupCtx,
		cleanupCancel: cleanupCancel,
	}
}

// Start 启动服务管理器
func (sm *ServiceManager) Start() error {
	sm.logger.Info("Starting service manager")

	// 启动健康检查器
	if err := sm.healthChecker.Start(sm.cleanupCtx); err != nil {
		return err
	}

	// 启动清理任务
	if sm.config.Registry.EnableAutoCleanup {
		go sm.registry.StartCleanup(
			sm.cleanupCtx,
			time.Duration(sm.config.Registry.CleanupInterval)*time.Second,
		)
	}

	sm.logger.Info("Service manager started successfully")
	return nil
}

// Stop 停止服务管理器
func (sm *ServiceManager) Stop() {
	sm.logger.Info("Stopping service manager")

	// 停止清理任务
	sm.cleanupCancel()

	// 停止健康检查器
	sm.healthChecker.Stop()

	// 关闭控制通道
	close(sm.stopCh)

	sm.logger.Info("Service manager stopped")
}

// GetRegistry 获取服务注册表
func (sm *ServiceManager) GetRegistry() *registry.ServiceRegistry {
	return sm.registry
}

// GetHealthChecker 获取健康检查器
func (sm *ServiceManager) GetHealthChecker() *health.Checker {
	return sm.healthChecker
}

// GetStats 获取统计信息
func (sm *ServiceManager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	registryStats, err := sm.registry.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	healthStats := sm.healthChecker.GetStats()

	return map[string]interface{}{
		"registry": registryStats,
		"health":   healthStats,
		"config": map[string]interface{}{
			"registry_ttl":           sm.config.Registry.TTL,
			"cleanup_interval":       sm.config.Registry.CleanupInterval,
			"auto_cleanup_enabled":   sm.config.Registry.EnableAutoCleanup,
			"health_check_enabled":   sm.config.Health.Enabled,
			"health_check_interval":  sm.config.Health.CheckInterval,
		},
	}, nil
}

// IsHealthy 检查服务管理器是否健康
func (sm *ServiceManager) IsHealthy(ctx context.Context) bool {
	// 检查注册表是否可用
	_, err := sm.registry.ListAllServices(ctx)
	if err != nil {
		sm.logger.Error("Registry health check failed", zap.Error(err))
		return false
	}

	return true
}

// IsReady 检查服务管理器是否就绪
func (sm *ServiceManager) IsReady(ctx context.Context) bool {
	return sm.IsHealthy(ctx)
}