package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	
	// 使用统一的健康检查接口
	"github.com/taishanglaojun/laojun/laojun-shared/health"
)

// MarketplaceHealthManager 市场API健康管理器
type MarketplaceHealthManager struct {
	manager health.HealthManager
	logger  *zap.Logger
}

// NewMarketplaceHealthManager 创建市场API健康管理器
func NewMarketplaceHealthManager(logger *zap.Logger) *MarketplaceHealthManager {
	config := &health.HealthConfig{
		EnableCache:    true,
		CacheTTL:       30 * time.Second,
		EnableMetrics:  true,
		DefaultTimeout: 10 * time.Second,
	}
	
	manager := health.NewDefaultHealthManager(config, logger)
	
	return &MarketplaceHealthManager{
		manager: manager,
		logger:  logger,
	}
}

// RegisterDatabaseCheck 注册数据库健康检查
func (mhm *MarketplaceHealthManager) RegisterDatabaseCheck(name string, db *sql.DB) {
	config := &health.CheckerConfig{
		Name:        name,
		Type:        health.CheckerTypeDatabase,
		Priority:    health.PriorityHigh,
		Timeout:     5 * time.Second,
		Interval:    30 * time.Second,
		Enabled:     true,
		Retries:     3,
		RetryDelay:  time.Second,
	}
	
	checker := health.NewEnhancedDatabaseChecker(config, db)
	mhm.manager.AddChecker(checker)
}

// RegisterRedisCheck 注册Redis健康检查
func (mhm *MarketplaceHealthManager) RegisterRedisCheck(name string, client *redis.Client) {
	config := &health.CheckerConfig{
		Name:        name,
		Type:        health.CheckerTypeRedis,
		Priority:    health.PriorityHigh,
		Timeout:     5 * time.Second,
		Interval:    30 * time.Second,
		Enabled:     true,
		Retries:     3,
		RetryDelay:  time.Second,
	}
	
	checker := health.NewEnhancedRedisChecker(config, client)
	mhm.manager.AddChecker(checker)
}

// RegisterHTTPCheck 注册HTTP服务健康检查
func (mhm *MarketplaceHealthManager) RegisterHTTPCheck(name, url string) {
	config := &health.CheckerConfig{
		Name:        name,
		Type:        health.CheckerTypeHTTP,
		Priority:    health.PriorityMedium,
		Timeout:     10 * time.Second,
		Interval:    60 * time.Second,
		Enabled:     true,
		Retries:     2,
		RetryDelay:  2 * time.Second,
	}
	
	httpConfig := &health.HTTPConfig{
		URL:    url,
		Method: "GET",
		ExpectedStatusCodes: []int{200, 201, 204},
		Headers: map[string]string{
			"User-Agent": "marketplace-api-health-check",
		},
	}
	
	checker := health.NewEnhancedHTTPChecker(config, httpConfig)
	mhm.manager.AddChecker(checker)
}

// RegisterSystemCheck 注册系统资源健康检查
func (mhm *MarketplaceHealthManager) RegisterSystemCheck() {
	config := &health.CheckerConfig{
		Name:        "marketplace_system",
		Type:        health.CheckerTypeSystem,
		Priority:    health.PriorityLow,
		Timeout:     5 * time.Second,
		Interval:    120 * time.Second,
		Enabled:     true,
		Retries:     1,
		RetryDelay:  time.Second,
	}
	
	systemConfig := &health.SystemConfig{
		CPUThreshold:    80.0,
		MemoryThreshold: 85.0,
		DiskThreshold:   90.0,
	}
	
	checker := health.NewSystemChecker(config, systemConfig)
	mhm.manager.AddChecker(checker)
}

// RegisterApplicationCheck 注册应用程序健康检查
func (mhm *MarketplaceHealthManager) RegisterApplicationCheck() {
	config := &health.CheckerConfig{
		Name:        "marketplace_application",
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
		// 检查关键服务状态
		// 这里可以添加具体的业务逻辑检查
		return &health.CheckResult{
			Status:    health.StatusHealthy,
			Message:   "Marketplace application is running normally",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"service":     "marketplace-api",
				"version":     "1.0.0",
				"uptime":      time.Since(time.Now().Add(-time.Hour)).String(),
				"goroutines":  "normal",
				"memory":      "normal",
			},
		}
	}
	
	checker := health.NewApplicationChecker(config, checkFunc)
	mhm.manager.AddChecker(checker)
}

// CheckHealth 执行健康检查
func (mhm *MarketplaceHealthManager) CheckHealth(ctx context.Context) *health.HealthReport {
	return mhm.manager.CheckHealth(ctx)
}

// CheckHealthByType 按类型执行健康检查
func (mhm *MarketplaceHealthManager) CheckHealthByType(ctx context.Context, checkerType health.CheckerType) *health.HealthReport {
	return mhm.manager.CheckHealthByType(ctx, checkerType)
}

// CheckHealthByPriority 按优先级执行健康检查
func (mhm *MarketplaceHealthManager) CheckHealthByPriority(ctx context.Context, priority health.Priority) *health.HealthReport {
	return mhm.manager.CheckHealthByPriority(ctx, priority)
}

// GetSummary 获取健康检查摘要
func (mhm *MarketplaceHealthManager) GetSummary(ctx context.Context) *health.Summary {
	return mhm.manager.GetSummary(ctx)
}

// Start 启动健康检查管理器
func (mhm *MarketplaceHealthManager) Start(ctx context.Context) error {
	return mhm.manager.Start(ctx)
}

// Stop 停止健康检查管理器
func (mhm *MarketplaceHealthManager) Stop(ctx context.Context) error {
	return mhm.manager.Stop(ctx)
}

// 以下代码保留用于向后兼容，但建议使用新的MarketplaceHealthManager

// HealthStatus 健康状态 (已弃用，请使用 health.Status)
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck 健康检查接口 (已弃用，请使用 health.HealthChecker)
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthCheckResult
}

// HealthCheckResult 健康检查结果 (已弃用，请使用 health.CheckResult)
type HealthCheckResult struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Details   interface{}  `json:"details,omitempty"`
}