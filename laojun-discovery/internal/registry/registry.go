package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/storage"
	"github.com/codetaoist/laojun-shared/registry"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DiscoveryServiceRegistry 发现服务注册管理器
type DiscoveryServiceRegistry struct {
	store  storage.Storage
	logger *zap.Logger
	mu     sync.RWMutex
	
	// 内存缓存
	cache map[string]*storage.ServiceInstance
	
	// 监听器
	listeners map[string][]ChangeListener
	
	// 统一注册接口
	registryManager registry.ServiceRegistry
}

// NewDiscoveryServiceRegistry 创建发现服务注册管理器
func NewDiscoveryServiceRegistry(store storage.Storage, logger *zap.Logger) *DiscoveryServiceRegistry {
	registryManager := registry.NewDefaultServiceRegistry(&registry.RegistryConfig{
		Type: "memory",
		TTL:  30 * time.Second,
	})
	
	return &DiscoveryServiceRegistry{
		store:           store,
		logger:          logger,
		cache:           make(map[string]*storage.ServiceInstance),
		listeners:       make(map[string][]ChangeListener),
		registryManager: registryManager,
	}
}

// RegisterService 使用统一接口注册服务
func (r *DiscoveryServiceRegistry) RegisterService(ctx context.Context, serviceInfo *registry.ServiceInfo) error {
	return r.registryManager.Register(ctx, serviceInfo)
}

// DeregisterService 使用统一接口注销服务
func (r *DiscoveryServiceRegistry) DeregisterService(ctx context.Context, serviceID string) error {
	return r.registryManager.Deregister(ctx, serviceID)
}

// GetServiceByID 使用统一接口获取服务
func (r *DiscoveryServiceRegistry) GetServiceByID(ctx context.Context, serviceID string) (*registry.ServiceInfo, error) {
	return r.registryManager.GetService(ctx, serviceID)
}

// DiscoverServices 使用统一接口发现服务
func (r *DiscoveryServiceRegistry) DiscoverServices(ctx context.Context, serviceName string) ([]*registry.ServiceInfo, error) {
	return r.registryManager.ListServices(ctx, serviceName)
}

// UpdateServiceHealth 使用统一接口更新服务健康状态
func (r *DiscoveryServiceRegistry) UpdateServiceHealth(ctx context.Context, serviceID string, health *registry.HealthCheck) error {
	return r.registryManager.UpdateHealth(ctx, serviceID, health)
}

// StartHeartbeat 使用统一接口启动心跳
func (r *DiscoveryServiceRegistry) StartHeartbeat(ctx context.Context, serviceID string) error {
	return r.registryManager.StartHeartbeat(ctx, serviceID)
}

// StopHeartbeat 使用统一接口停止心跳
func (r *DiscoveryServiceRegistry) StopHeartbeat(ctx context.Context, serviceID string) error {
	return r.registryManager.StopHeartbeat(ctx, serviceID)
}

// WatchServices 使用统一接口监听服务变化
func (r *DiscoveryServiceRegistry) WatchServices(ctx context.Context, serviceName string) (<-chan *registry.ServiceEvent, error) {
	return r.registryManager.Watch(ctx, serviceName)
}

// GetRegistryHealth 使用统一接口获取注册中心健康状态
func (r *DiscoveryServiceRegistry) GetRegistryHealth(ctx context.Context) (*registry.RegistryHealth, error) {
	return r.registryManager.Health(ctx)
}

// ServiceRegistry 服务注册表 (已弃用，建议使用 DiscoveryServiceRegistry)
type ServiceRegistry struct {
	store  storage.Storage
	logger *zap.Logger
	mu     sync.RWMutex
	
	// 内存缓存
	cache map[string]*storage.ServiceInstance
	
	// 监听器
	listeners map[string][]ChangeListener
}

// ChangeListener 变更监听器
type ChangeListener func(event ChangeEvent)

// ChangeEvent 变更事件
type ChangeEvent struct {
	Type     string                    `json:"type"`     // register, deregister, health_change
	Service  string                    `json:"service"`
	Instance *storage.ServiceInstance `json:"instance"`
	Time     time.Time                 `json:"time"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name    string            `json:"name" binding:"required"`
	Address string            `json:"address" binding:"required"`
	Port    int               `json:"port" binding:"required"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
	TTL     int               `json:"ttl"`
	Health  *HealthCheck      `json:"health"`
}

// HealthCheck 健康检查配置
type HealthCheck struct {
	HTTP     string `json:"http"`
	TCP      string `json:"tcp"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

// NewServiceRegistry 创建服务注册表 (已弃用，建议使用 NewDiscoveryServiceRegistry)
func NewServiceRegistry(store storage.Storage, logger *zap.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		store:     store,
		logger:    logger,
		cache:     make(map[string]*storage.ServiceInstance),
		listeners: make(map[string][]ChangeListener),
	}
}

// Register 注册服务
func (r *ServiceRegistry) Register(ctx context.Context, req *RegisterRequest) (*storage.ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 生成服务ID
	serviceID := uuid.New().String()

	// 设置默认TTL
	if req.TTL <= 0 {
		req.TTL = 30
	}

	// 创建服务实例
	instance := &storage.ServiceInstance{
		ID:      serviceID,
		Name:    req.Name,
		Address: req.Address,
		Port:    req.Port,
		Tags:    req.Tags,
		Meta:    req.Meta,
		TTL:     req.TTL,
		Health: storage.HealthStatus{
			Status:      "passing",
			Output:      "Initial registration",
			LastChecked: time.Now(),
		},
		LastSeen: time.Now(),
	}

	// 存储到持久化存储
	if err := r.store.RegisterService(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	// 更新内存缓存
	r.cache[serviceID] = instance

	// 触发变更事件
	r.notifyListeners(ChangeEvent{
		Type:     "register",
		Service:  req.Name,
		Instance: instance,
		Time:     time.Now(),
	})

	r.logger.Info("Service registered",
		zap.String("service", req.Name),
		zap.String("id", serviceID),
		zap.String("address", fmt.Sprintf("%s:%d", req.Address, req.Port)))

	return instance, nil
}

// Deregister 注销服务
func (r *ServiceRegistry) Deregister(ctx context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 从缓存获取服务信息
	instance, exists := r.cache[serviceID]
	if !exists {
		// 尝试从存储获取
		var err error
		instance, err = r.store.GetService(ctx, serviceID)
		if err != nil {
			return fmt.Errorf("service not found: %s", serviceID)
		}
	}

	// 从持久化存储删除
	if err := r.store.DeregisterService(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// 从内存缓存删除
	delete(r.cache, serviceID)

	// 触发变更事件
	r.notifyListeners(ChangeEvent{
		Type:     "deregister",
		Service:  instance.Name,
		Instance: instance,
		Time:     time.Now(),
	})

	r.logger.Info("Service deregistered",
		zap.String("service", instance.Name),
		zap.String("id", serviceID))

	return nil
}

// GetService 获取服务实例
func (r *ServiceRegistry) GetService(ctx context.Context, serviceID string) (*storage.ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 先从缓存查找
	if instance, exists := r.cache[serviceID]; exists {
		return instance, nil
	}

	// 从存储查找
	instance, err := r.store.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	r.cache[serviceID] = instance

	return instance, nil
}

// ListServices 列出服务实例
func (r *ServiceRegistry) ListServices(ctx context.Context, serviceName string) ([]*storage.ServiceInstance, error) {
	instances, err := r.store.ListServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	r.mu.Lock()
	for _, instance := range instances {
		r.cache[instance.ID] = instance
	}
	r.mu.Unlock()

	return instances, nil
}

// ListAllServices 列出所有服务
func (r *ServiceRegistry) ListAllServices(ctx context.Context) (map[string][]*storage.ServiceInstance, error) {
	services, err := r.store.ListAllServices(ctx)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	r.mu.Lock()
	for _, instances := range services {
		for _, instance := range instances {
			r.cache[instance.ID] = instance
		}
	}
	r.mu.Unlock()

	return services, nil
}

// UpdateHealth 更新健康状态
func (r *ServiceRegistry) UpdateHealth(ctx context.Context, serviceID string, health storage.HealthStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新存储
	if err := r.store.UpdateHealth(ctx, serviceID, health); err != nil {
		return fmt.Errorf("failed to update health: %w", err)
	}

	// 更新缓存
	if instance, exists := r.cache[serviceID]; exists {
		oldStatus := instance.Health.Status
		instance.Health = health
		
		// 如果健康状态发生变化，触发事件
		if oldStatus != health.Status {
			r.notifyListeners(ChangeEvent{
				Type:     "health_change",
				Service:  instance.Name,
				Instance: instance,
				Time:     time.Now(),
			})
		}
	}

	return nil
}

// GetHealthyServices 获取健康的服务实例
func (r *ServiceRegistry) GetHealthyServices(ctx context.Context, serviceName string) ([]*storage.ServiceInstance, error) {
	return r.store.GetHealthyServices(ctx, serviceName)
}

// RefreshTTL 刷新TTL
func (r *ServiceRegistry) RefreshTTL(ctx context.Context, serviceID string, ttl int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新存储
	if err := r.store.RefreshTTL(ctx, serviceID, ttl); err != nil {
		return fmt.Errorf("failed to refresh TTL: %w", err)
	}

	// 更新缓存
	if instance, exists := r.cache[serviceID]; exists {
		instance.TTL = ttl
		instance.LastSeen = time.Now()
	}

	return nil
}

// AddChangeListener 添加变更监听器
func (r *ServiceRegistry) AddChangeListener(serviceName string, listener ChangeListener) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.listeners[serviceName] == nil {
		r.listeners[serviceName] = make([]ChangeListener, 0)
	}
	r.listeners[serviceName] = append(r.listeners[serviceName], listener)
}

// RemoveChangeListener 移除变更监听器
func (r *ServiceRegistry) RemoveChangeListener(serviceName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.listeners, serviceName)
}

// notifyListeners 通知监听器
func (r *ServiceRegistry) notifyListeners(event ChangeEvent) {
	// 通知特定服务的监听器
	if listeners, exists := r.listeners[event.Service]; exists {
		for _, listener := range listeners {
			go listener(event)
		}
	}

	// 通知全局监听器
	if listeners, exists := r.listeners["*"]; exists {
		for _, listener := range listeners {
			go listener(event)
		}
	}
}

// StartCleanup 启动清理任务
func (r *ServiceRegistry) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.store.CleanupExpiredServices(ctx); err != nil {
				r.logger.Error("Failed to cleanup expired services", zap.Error(err))
			}
		}
	}
}

// GetStats 获取统计信息
func (r *ServiceRegistry) GetStats(ctx context.Context) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services, err := r.store.ListAllServices(ctx)
	if err != nil {
		return nil, err
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
		"cached_instances":   len(r.cache),
		"active_listeners":   len(r.listeners),
	}, nil
}