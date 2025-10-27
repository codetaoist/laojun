package services

import (
	"context"
	"fmt"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/codetaoist/laojun-gateway/internal/services/discovery"
	"github.com/codetaoist/laojun-gateway/internal/services/ratelimit"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	config    *config.Config
	logger    *zap.Logger
	redis     *redis.Client
	discovery discovery.Service
	rateLimit ratelimit.Service
}

// NewServiceManager 创建服务管理器
func NewServiceManager(cfg *config.Config, logger *zap.Logger) *ServiceManager {
	return &ServiceManager{
		config: cfg,
		logger: logger,
	}
}

// Initialize 初始化所有服务
func (sm *ServiceManager) Initialize() error {
	// 初始化服务发现
	if err := sm.initDiscovery(); err != nil {
		return fmt.Errorf("failed to initialize discovery: %w", err)
	}

	// 如果启用了限流，则初始化Redis和限流服务
	if sm.config.RateLimit.Enabled {
		if err := sm.initRedis(); err != nil {
			return fmt.Errorf("failed to initialize redis: %w", err)
		}

		if err := sm.initRateLimit(); err != nil {
			return fmt.Errorf("failed to initialize rate limit: %w", err)
		}
	} else {
		sm.logger.Info("Rate limiting disabled, skipping Redis initialization")
	}

	sm.logger.Info("All services initialized successfully")
	return nil
}

// initRedis 初始化Redis连接
func (sm *ServiceManager) initRedis() error {
	sm.redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", sm.config.Redis.Host, sm.config.Redis.Port),
		Password: sm.config.Redis.Password,
		DB:       sm.config.Redis.DB,
	})

	// 测试连接
	ctx := context.Background()
	if err := sm.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	sm.logger.Info("Redis connected successfully")
	return nil
}

// initDiscovery 初始化服务发现
func (sm *ServiceManager) initDiscovery() error {
	var err error
	sm.discovery, err = discovery.NewService(sm.config.Discovery, sm.logger)
	if err != nil {
		return err
	}

	sm.logger.Info("Service discovery initialized", 
		zap.String("type", sm.config.Discovery.Type))
	return nil
}

// initRateLimit 初始化限流服务
func (sm *ServiceManager) initRateLimit() error {
	sm.rateLimit = ratelimit.NewService(sm.config.RateLimit, sm.redis, sm.logger)
	sm.logger.Info("Rate limit service initialized")
	return nil
}

// GetRedis 获取Redis客户端
func (sm *ServiceManager) GetRedis() *redis.Client {
	return sm.redis
}

// GetDiscovery 获取服务发现服务
func (sm *ServiceManager) GetDiscovery() discovery.Service {
	return sm.discovery
}

// GetRateLimit 获取限流服务
func (sm *ServiceManager) GetRateLimit() ratelimit.Service {
	return sm.rateLimit
}

// Cleanup 清理资源
func (sm *ServiceManager) Cleanup() {
	if sm.redis != nil {
		sm.redis.Close()
		sm.logger.Info("Redis connection closed")
	}

	if sm.discovery != nil {
		sm.discovery.Close()
		sm.logger.Info("Service discovery closed")
	}

	sm.logger.Info("All services cleaned up")
}