package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultServiceRegistry 默认服务注册实现
type DefaultServiceRegistry struct {
	config    *RegistryConfig
	logger    *zap.Logger
	mu        sync.RWMutex
	services  map[string]*ServiceInfo
	watchers  map[string][]chan *ServiceEvent
	heartbeats map[string]time.Time
	stopCh    chan struct{}
	started   bool
}

// NewDefaultServiceRegistry 创建默认服务注册实例
func NewDefaultServiceRegistry(config *RegistryConfig, logger *zap.Logger) *DefaultServiceRegistry {
	return &DefaultServiceRegistry{
		config:     config,
		logger:     logger,
		services:   make(map[string]*ServiceInfo),
		watchers:   make(map[string][]chan *ServiceEvent),
		heartbeats: make(map[string]time.Time),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动服务注册中心
func (r *DefaultServiceRegistry) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.started {
		return fmt.Errorf("registry already started")
	}

	r.started = true
	
	// 启动心跳检查协程
	go r.heartbeatChecker()
	
	r.logger.Info("Service registry started")
	return nil
}

// Stop 停止服务注册中心
func (r *DefaultServiceRegistry) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return nil
	}

	close(r.stopCh)
	r.started = false
	
	// 关闭所有监听器
	for _, watchers := range r.watchers {
		for _, watcher := range watchers {
			close(watcher)
		}
	}
	r.watchers = make(map[string][]chan *ServiceEvent)
	
	r.logger.Info("Service registry stopped")
	return nil
}

// RegisterService 注册服务
func (r *DefaultServiceRegistry) RegisterService(ctx context.Context, service *ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if service.ID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}

	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if service.Address == "" {
		return fmt.Errorf("service address cannot be empty")
	}

	if service.Port <= 0 {
		return fmt.Errorf("service port must be positive")
	}

	// 设置默认值
	if service.RegisteredAt.IsZero() {
		service.RegisteredAt = time.Now()
	}
	service.LastSeen = time.Now()
	
	if service.Status == "" {
		service.Status = ServiceStatusActive
	}

	if service.TTL <= 0 {
		service.TTL = int(r.config.TTL.Seconds())
	}

	if service.Weight <= 0 {
		service.Weight = 1
	}

	// 存储服务信息
	r.services[service.ID] = service
	r.heartbeats[service.ID] = time.Now()

	// 通知监听器
	r.notifyWatchers(service.Name, &ServiceEvent{
		Type:      EventTypeRegister,
		Service:   service,
		Timestamp: time.Now(),
	})

	r.logger.Info("Service registered",
		zap.String("service_id", service.ID),
		zap.String("service_name", service.Name),
		zap.String("address", fmt.Sprintf("%s:%d", service.Address, service.Port)))

	return nil
}

// DeregisterService 注销服务
func (r *DefaultServiceRegistry) DeregisterService(ctx context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	delete(r.services, serviceID)
	delete(r.heartbeats, serviceID)

	// 通知监听器
	r.notifyWatchers(service.Name, &ServiceEvent{
		Type:      EventTypeDeregister,
		Service:   service,
		Timestamp: time.Now(),
	})

	r.logger.Info("Service deregistered",
		zap.String("service_id", serviceID),
		zap.String("service_name", service.Name))

	return nil
}

// UpdateService 更新服务信息
func (r *DefaultServiceRegistry) UpdateService(ctx context.Context, service *ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[service.ID]; !exists {
		return fmt.Errorf("service not found: %s", service.ID)
	}

	service.LastSeen = time.Now()
	r.services[service.ID] = service
	r.heartbeats[service.ID] = time.Now()

	// 通知监听器
	r.notifyWatchers(service.Name, &ServiceEvent{
		Type:      EventTypeUpdate,
		Service:   service,
		Timestamp: time.Now(),
	})

	r.logger.Info("Service updated",
		zap.String("service_id", service.ID),
		zap.String("service_name", service.Name))

	return nil
}

// GetService 获取服务信息
func (r *DefaultServiceRegistry) GetService(ctx context.Context, serviceID string) (*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	// 返回副本
	serviceCopy := *service
	return &serviceCopy, nil
}

// ListServices 列出指定名称的服务
func (r *DefaultServiceRegistry) ListServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []*ServiceInfo
	for _, service := range r.services {
		if service.Name == serviceName {
			serviceCopy := *service
			services = append(services, &serviceCopy)
		}
	}

	return services, nil
}

// ListAllServices 列出所有服务
func (r *DefaultServiceRegistry) ListAllServices(ctx context.Context) (map[string][]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*ServiceInfo)
	for _, service := range r.services {
		serviceCopy := *service
		result[service.Name] = append(result[service.Name], &serviceCopy)
	}

	return result, nil
}

// GetHealthyServices 获取健康的服务实例
func (r *DefaultServiceRegistry) GetHealthyServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	services, err := r.ListServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	var healthyServices []*ServiceInfo
	for _, service := range services {
		if r.isServiceHealthy(service) {
			healthyServices = append(healthyServices, service)
		}
	}

	return healthyServices, nil
}

// WatchServices 监听服务变化
func (r *DefaultServiceRegistry) WatchServices(ctx context.Context, serviceName string) (<-chan *ServiceEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	watcher := make(chan *ServiceEvent, 100)
	r.watchers[serviceName] = append(r.watchers[serviceName], watcher)

	// 在goroutine中处理context取消
	go func() {
		<-ctx.Done()
		r.removeWatcher(serviceName, watcher)
	}()

	return watcher, nil
}

// Heartbeat 发送心跳
func (r *DefaultServiceRegistry) Heartbeat(ctx context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	service.LastSeen = time.Now()
	r.heartbeats[serviceID] = time.Now()

	return nil
}

// GetRegistryHealth 获取注册中心健康状态
func (r *DefaultServiceRegistry) GetRegistryHealth(ctx context.Context) (*RegistryHealth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalServices := len(r.services)
	healthyServices := 0
	unhealthyServices := 0

	for _, service := range r.services {
		if r.isServiceHealthy(service) {
			healthyServices++
		} else {
			unhealthyServices++
		}
	}

	status := HealthStatusHealthy
	if unhealthyServices > 0 {
		if healthyServices == 0 {
			status = HealthStatusUnhealthy
		} else {
			status = HealthStatusDegraded
		}
	}

	return &RegistryHealth{
		Status:            status,
		TotalServices:     totalServices,
		HealthyServices:   healthyServices,
		UnhealthyServices: unhealthyServices,
		LastCheck:         time.Now(),
		Details: map[string]string{
			"registry_type": string(r.config.Type),
			"namespace":     r.config.Namespace,
		},
	}, nil
}

// 私有方法

// notifyWatchers 通知监听器
func (r *DefaultServiceRegistry) notifyWatchers(serviceName string, event *ServiceEvent) {
	watchers := r.watchers[serviceName]
	for _, watcher := range watchers {
		select {
		case watcher <- event:
		default:
			// 如果通道已满，跳过这个监听器
			r.logger.Warn("Watcher channel full, skipping event",
				zap.String("service_name", serviceName),
				zap.String("event_type", string(event.Type)))
		}
	}
}

// removeWatcher 移除监听器
func (r *DefaultServiceRegistry) removeWatcher(serviceName string, watcher chan *ServiceEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	watchers := r.watchers[serviceName]
	for i, w := range watchers {
		if w == watcher {
			close(watcher)
			r.watchers[serviceName] = append(watchers[:i], watchers[i+1:]...)
			break
		}
	}
}

// isServiceHealthy 检查服务是否健康
func (r *DefaultServiceRegistry) isServiceHealthy(service *ServiceInfo) bool {
	// 检查心跳超时
	if lastHeartbeat, exists := r.heartbeats[service.ID]; exists {
		if time.Since(lastHeartbeat) > time.Duration(service.TTL)*time.Second {
			return false
		}
	}

	// 检查服务状态
	return service.Status == ServiceStatusActive
}

// heartbeatChecker 心跳检查器
func (r *DefaultServiceRegistry) heartbeatChecker() {
	ticker := time.NewTicker(r.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.checkExpiredServices()
		case <-r.stopCh:
			return
		}
	}
}

// checkExpiredServices 检查过期服务
func (r *DefaultServiceRegistry) checkExpiredServices() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var expiredServices []string

	for serviceID, lastHeartbeat := range r.heartbeats {
		service := r.services[serviceID]
		if now.Sub(lastHeartbeat) > time.Duration(service.TTL)*time.Second {
			expiredServices = append(expiredServices, serviceID)
		}
	}

	// 移除过期服务
	for _, serviceID := range expiredServices {
		service := r.services[serviceID]
		delete(r.services, serviceID)
		delete(r.heartbeats, serviceID)

		// 通知监听器
		r.notifyWatchers(service.Name, &ServiceEvent{
			Type:      EventTypeDeregister,
			Service:   service,
			Timestamp: now,
		})

		r.logger.Info("Service expired and removed",
			zap.String("service_id", serviceID),
			zap.String("service_name", service.Name))
	}
}