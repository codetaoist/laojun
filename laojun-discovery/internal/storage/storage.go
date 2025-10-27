package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Meta     map[string]string `json:"meta"`
	Health   HealthStatus      `json:"health"`
	TTL      int               `json:"ttl"`
	LastSeen time.Time         `json:"last_seen"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status      string    `json:"status"` // passing, warning, critical
	Output      string    `json:"output"`
	LastChecked time.Time `json:"last_checked"`
}

// Storage 存储接口
type Storage interface {
	// 服务实例管理
	RegisterService(ctx context.Context, instance *ServiceInstance) error
	DeregisterService(ctx context.Context, serviceID string) error
	GetService(ctx context.Context, serviceID string) (*ServiceInstance, error)
	ListServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	ListAllServices(ctx context.Context) (map[string][]*ServiceInstance, error)
	
	// 健康状态管理
	UpdateHealth(ctx context.Context, serviceID string, health HealthStatus) error
	GetHealthyServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	
	// TTL管理
	RefreshTTL(ctx context.Context, serviceID string, ttl int) error
	CleanupExpiredServices(ctx context.Context) error
	
	// 关闭连接
	Close() error
}

// RedisStorage Redis存储实现
type RedisStorage struct {
	client *redis.Client
	logger *zap.Logger
}

// NewStorage 创建存储实例
func NewStorage(cfg *config.Config, logger *zap.Logger) (Storage, error) {
	switch cfg.Storage.Type {
	case "redis":
		return NewRedisStorage(cfg.Storage.Redis, logger)
	case "memory":
		return NewMemoryStorage(logger), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}

// NewRedisStorage 创建Redis存储
func NewRedisStorage(cfg config.RedisConfig, logger *zap.Logger) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		zap.Int("db", cfg.DB))

	return &RedisStorage{
		client: client,
		logger: logger,
	}, nil
}

// RegisterService 注册服务
func (r *RedisStorage) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	instance.LastSeen = time.Now()
	
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal service instance: %w", err)
	}

	// 存储服务实例
	serviceKey := fmt.Sprintf("services:%s:%s", instance.Name, instance.ID)
	if err := r.client.Set(ctx, serviceKey, data, time.Duration(instance.TTL)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to store service instance: %w", err)
	}

	// 添加到服务列表
	listKey := fmt.Sprintf("service_list:%s", instance.Name)
	if err := r.client.SAdd(ctx, listKey, instance.ID).Err(); err != nil {
		return fmt.Errorf("failed to add to service list: %w", err)
	}

	// 设置服务列表过期时间
	r.client.Expire(ctx, listKey, time.Duration(instance.TTL*2)*time.Second)

	r.logger.Info("Service registered",
		zap.String("service", instance.Name),
		zap.String("id", instance.ID),
		zap.String("address", fmt.Sprintf("%s:%d", instance.Address, instance.Port)))

	return nil
}

// DeregisterService 注销服务
func (r *RedisStorage) DeregisterService(ctx context.Context, serviceID string) error {
	// 查找服务实例
	pattern := fmt.Sprintf("services:*:%s", serviceID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find service keys: %w", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	for _, key := range keys {
		// 获取服务信息
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var instance ServiceInstance
		if err := json.Unmarshal([]byte(data), &instance); err != nil {
			continue
		}

		// 从服务列表中移除
		listKey := fmt.Sprintf("service_list:%s", instance.Name)
		r.client.SRem(ctx, listKey, serviceID)

		// 删除服务实例
		r.client.Del(ctx, key)

		r.logger.Info("Service deregistered",
			zap.String("service", instance.Name),
			zap.String("id", serviceID))
	}

	return nil
}

// GetService 获取服务实例
func (r *RedisStorage) GetService(ctx context.Context, serviceID string) (*ServiceInstance, error) {
	pattern := fmt.Sprintf("services:*:%s", serviceID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to find service keys: %w", err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	data, err := r.client.Get(ctx, keys[0]).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service data: %w", err)
	}

	var instance ServiceInstance
	if err := json.Unmarshal([]byte(data), &instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service data: %w", err)
	}

	return &instance, nil
}

// ListServices 列出指定服务的所有实例
func (r *RedisStorage) ListServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	listKey := fmt.Sprintf("service_list:%s", serviceName)
	serviceIDs, err := r.client.SMembers(ctx, listKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service list: %w", err)
	}

	var instances []*ServiceInstance
	for _, serviceID := range serviceIDs {
		serviceKey := fmt.Sprintf("services:%s:%s", serviceName, serviceID)
		data, err := r.client.Get(ctx, serviceKey).Result()
		if err != nil {
			// 服务可能已过期，从列表中移除
			r.client.SRem(ctx, listKey, serviceID)
			continue
		}

		var instance ServiceInstance
		if err := json.Unmarshal([]byte(data), &instance); err != nil {
			r.logger.Warn("Failed to unmarshal service data",
				zap.String("service", serviceName),
				zap.String("id", serviceID),
				zap.Error(err))
			continue
		}

		instances = append(instances, &instance)
	}

	return instances, nil
}

// ListAllServices 列出所有服务
func (r *RedisStorage) ListAllServices(ctx context.Context) (map[string][]*ServiceInstance, error) {
	pattern := "service_list:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service list keys: %w", err)
	}

	result := make(map[string][]*ServiceInstance)
	for _, key := range keys {
		serviceName := key[13:] // 移除 "service_list:" 前缀
		instances, err := r.ListServices(ctx, serviceName)
		if err != nil {
			r.logger.Warn("Failed to list services",
				zap.String("service", serviceName),
				zap.Error(err))
			continue
		}
		result[serviceName] = instances
	}

	return result, nil
}

// UpdateHealth 更新健康状态
func (r *RedisStorage) UpdateHealth(ctx context.Context, serviceID string, health HealthStatus) error {
	instance, err := r.GetService(ctx, serviceID)
	if err != nil {
		return err
	}

	instance.Health = health
	instance.LastSeen = time.Now()

	return r.RegisterService(ctx, instance)
}

// GetHealthyServices 获取健康的服务实例
func (r *RedisStorage) GetHealthyServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	instances, err := r.ListServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	var healthy []*ServiceInstance
	for _, instance := range instances {
		if instance.Health.Status == "passing" {
			healthy = append(healthy, instance)
		}
	}

	return healthy, nil
}

// RefreshTTL 刷新TTL
func (r *RedisStorage) RefreshTTL(ctx context.Context, serviceID string, ttl int) error {
	pattern := fmt.Sprintf("services:*:%s", serviceID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find service keys: %w", err)
	}

	for _, key := range keys {
		r.client.Expire(ctx, key, time.Duration(ttl)*time.Second)
	}

	return nil
}

// CleanupExpiredServices 清理过期服务
func (r *RedisStorage) CleanupExpiredServices(ctx context.Context) error {
	pattern := "service_list:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get service list keys: %w", err)
	}

	for _, key := range keys {
		serviceName := key[13:] // 移除 "service_list:" 前缀
		serviceIDs, err := r.client.SMembers(ctx, key).Result()
		if err != nil {
			continue
		}

		for _, serviceID := range serviceIDs {
			serviceKey := fmt.Sprintf("services:%s:%s", serviceName, serviceID)
			exists, err := r.client.Exists(ctx, serviceKey).Result()
			if err != nil {
				continue
			}

			if exists == 0 {
				// 服务已过期，从列表中移除
				r.client.SRem(ctx, key, serviceID)
				r.logger.Info("Cleaned up expired service",
					zap.String("service", serviceName),
					zap.String("id", serviceID))
			}
		}
	}

	return nil
}

// Close 关闭连接
func (r *RedisStorage) Close() error {
	return r.client.Close()
}