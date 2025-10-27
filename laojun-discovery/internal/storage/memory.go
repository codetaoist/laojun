package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	services   map[string]*ServiceInstance // serviceID -> ServiceInstance
	indexes    map[string][]string         // serviceName -> []serviceID
	tagIndexes map[string][]string         // tag -> []serviceID (新增标签索引)
	mutex      sync.RWMutex
	logger     *zap.Logger
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage(logger *zap.Logger) *MemoryStorage {
	return &MemoryStorage{
		services:   make(map[string]*ServiceInstance),
		indexes:    make(map[string][]string),
		tagIndexes: make(map[string][]string),
		logger:     logger,
	}
}

// RegisterService 注册服务实例
func (m *MemoryStorage) RegisterService(ctx context.Context, instance *ServiceInstance) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 更新最后见到时间
	instance.LastSeen = time.Now()

	// 如果是更新现有服务，先清理旧的标签索引
	if oldInstance, exists := m.services[instance.ID]; exists {
		m.removeFromTagIndexes(instance.ID, oldInstance.Tags)
	}

	// 存储服务实例
	m.services[instance.ID] = instance

	// 更新服务名索引
	serviceIDs := m.indexes[instance.Name]
	found := false
	for _, id := range serviceIDs {
		if id == instance.ID {
			found = true
			break
		}
	}
	if !found {
		m.indexes[instance.Name] = append(serviceIDs, instance.ID)
	}

	// 更新标签索引
	m.addToTagIndexes(instance.ID, instance.Tags)

	m.logger.Debug("Service registered", 
		zap.String("service_id", instance.ID),
		zap.String("service_name", instance.Name),
		zap.Strings("tags", instance.Tags))

	return nil
}

// addToTagIndexes 添加服务到标签索引
func (m *MemoryStorage) addToTagIndexes(serviceID string, tags []string) {
	for _, tag := range tags {
		tagServiceIDs := m.tagIndexes[tag]
		found := false
		for _, id := range tagServiceIDs {
			if id == serviceID {
				found = true
				break
			}
		}
		if !found {
			m.tagIndexes[tag] = append(tagServiceIDs, serviceID)
		}
	}
}

// removeFromTagIndexes 从标签索引中移除服务
func (m *MemoryStorage) removeFromTagIndexes(serviceID string, tags []string) {
	for _, tag := range tags {
		tagServiceIDs := m.tagIndexes[tag]
		for i, id := range tagServiceIDs {
			if id == serviceID {
				// 移除该服务ID
				m.tagIndexes[tag] = append(tagServiceIDs[:i], tagServiceIDs[i+1:]...)
				// 如果标签下没有服务了，删除该标签索引
				if len(m.tagIndexes[tag]) == 0 {
					delete(m.tagIndexes, tag)
				}
				break
			}
		}
	}
}

// DeregisterService 注销服务实例
func (m *MemoryStorage) DeregisterService(ctx context.Context, serviceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	instance, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	// 从服务映射中删除
	delete(m.services, serviceID)

	// 从服务名索引中删除
	serviceIDs := m.indexes[instance.Name]
	for i, id := range serviceIDs {
		if id == serviceID {
			m.indexes[instance.Name] = append(serviceIDs[:i], serviceIDs[i+1:]...)
			break
		}
	}

	// 如果该服务名下没有实例了，删除索引
	if len(m.indexes[instance.Name]) == 0 {
		delete(m.indexes, instance.Name)
	}

	// 从标签索引中删除
	m.removeFromTagIndexes(serviceID, instance.Tags)

	m.logger.Info("Service deregistered", 
		zap.String("service_id", serviceID),
		zap.String("service_name", instance.Name))

	return nil
}

// GetService 获取服务实例
func (m *MemoryStorage) GetService(ctx context.Context, serviceID string) (*ServiceInstance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	instance, exists := m.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	// 返回副本以避免并发修改
	instanceCopy := *instance
	return &instanceCopy, nil
}

// ListServices 列出指定服务名的所有实例
func (m *MemoryStorage) ListServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	serviceIDs, exists := m.indexes[serviceName]
	if !exists {
		return []*ServiceInstance{}, nil
	}

	instances := make([]*ServiceInstance, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		if instance, exists := m.services[serviceID]; exists {
			// 返回副本以避免并发修改
			instanceCopy := *instance
			instances = append(instances, &instanceCopy)
		}
	}

	return instances, nil
}

// ListAllServices 列出所有服务
func (m *MemoryStorage) ListAllServices(ctx context.Context) (map[string][]*ServiceInstance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string][]*ServiceInstance)
	
	for serviceName, serviceIDs := range m.indexes {
		instances := make([]*ServiceInstance, 0, len(serviceIDs))
		for _, serviceID := range serviceIDs {
			if instance, exists := m.services[serviceID]; exists {
				// 返回副本以避免并发修改
				instanceCopy := *instance
				instances = append(instances, &instanceCopy)
			}
		}
		result[serviceName] = instances
	}

	return result, nil
}

// UpdateHealth 更新健康状态
func (m *MemoryStorage) UpdateHealth(ctx context.Context, serviceID string, health HealthStatus) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	instance, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	instance.Health = health
	m.logger.Debug("Health updated", 
		zap.String("service_id", serviceID),
		zap.String("status", health.Status))

	return nil
}

// GetHealthyServices 获取健康的服务实例
func (m *MemoryStorage) GetHealthyServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	instances, err := m.ListServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	healthyInstances := make([]*ServiceInstance, 0)
	for _, instance := range instances {
		if instance.Health.Status == "passing" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances, nil
}

// ListServicesByTags 根据标签高效查询服务实例
func (m *MemoryStorage) ListServicesByTags(ctx context.Context, serviceName string, tags []string) ([]*ServiceInstance, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(tags) == 0 {
		// 如果没有标签过滤，直接返回所有服务实例
		return m.ListServices(ctx, serviceName)
	}

	// 使用标签索引找到候选服务ID集合
	var candidateIDs []string
	
	// 获取第一个标签对应的服务ID列表作为初始候选集
	if len(tags) > 0 {
		if serviceIDs, exists := m.tagIndexes[tags[0]]; exists {
			candidateIDs = make([]string, len(serviceIDs))
			copy(candidateIDs, serviceIDs)
		} else {
			// 如果第一个标签没有对应的服务，直接返回空结果
			return []*ServiceInstance{}, nil
		}
	}

	// 对于剩余的标签，取交集
	for i := 1; i < len(tags); i++ {
		tag := tags[i]
		tagServiceIDs, exists := m.tagIndexes[tag]
		if !exists {
			// 如果任何一个标签没有对应的服务，返回空结果
			return []*ServiceInstance{}, nil
		}

		// 计算交集
		intersection := make([]string, 0)
		tagIDMap := make(map[string]bool)
		for _, id := range tagServiceIDs {
			tagIDMap[id] = true
		}

		for _, id := range candidateIDs {
			if tagIDMap[id] {
				intersection = append(intersection, id)
			}
		}
		candidateIDs = intersection

		// 如果交集为空，直接返回
		if len(candidateIDs) == 0 {
			return []*ServiceInstance{}, nil
		}
	}

	// 过滤出指定服务名的实例
	instances := make([]*ServiceInstance, 0)
	for _, serviceID := range candidateIDs {
		if instance, exists := m.services[serviceID]; exists {
			if serviceName == "" || instance.Name == serviceName {
				// 返回副本以避免并发修改
				instanceCopy := *instance
				instances = append(instances, &instanceCopy)
			}
		}
	}

	return instances, nil
}

// GetHealthyServicesByTags 根据标签获取健康的服务实例
func (m *MemoryStorage) GetHealthyServicesByTags(ctx context.Context, serviceName string, tags []string) ([]*ServiceInstance, error) {
	instances, err := m.ListServicesByTags(ctx, serviceName, tags)
	if err != nil {
		return nil, err
	}

	healthyInstances := make([]*ServiceInstance, 0)
	for _, instance := range instances {
		if instance.Health.Status == "passing" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances, nil
}

// RefreshTTL 刷新TTL
func (m *MemoryStorage) RefreshTTL(ctx context.Context, serviceID string, ttl int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	instance, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceID)
	}

	instance.TTL = ttl
	instance.LastSeen = time.Now()

	return nil
}

// CleanupExpiredServices 清理过期服务
func (m *MemoryStorage) CleanupExpiredServices(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	expiredServices := make([]string, 0)

	for serviceID, instance := range m.services {
		if instance.TTL > 0 {
			expiredTime := instance.LastSeen.Add(time.Duration(instance.TTL) * time.Second)
			if now.After(expiredTime) {
				expiredServices = append(expiredServices, serviceID)
			}
		}
	}

	// 删除过期服务
	for _, serviceID := range expiredServices {
		instance := m.services[serviceID]
		delete(m.services, serviceID)

		// 从服务名索引中删除
		serviceIDs := m.indexes[instance.Name]
		for i, id := range serviceIDs {
			if id == serviceID {
				m.indexes[instance.Name] = append(serviceIDs[:i], serviceIDs[i+1:]...)
				break
			}
		}

		// 如果该服务名下没有实例了，删除索引
		if len(m.indexes[instance.Name]) == 0 {
			delete(m.indexes, instance.Name)
		}

		// 从标签索引中删除
		m.removeFromTagIndexes(serviceID, instance.Tags)

		m.logger.Info("Expired service cleaned up", 
			zap.String("service_id", serviceID),
			zap.String("service_name", instance.Name))
	}

	if len(expiredServices) > 0 {
		m.logger.Info("Cleanup completed", zap.Int("expired_count", len(expiredServices)))
	}

	return nil
}

// Close 关闭存储
func (m *MemoryStorage) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 清空所有数据
	m.services = make(map[string]*ServiceInstance)
	m.indexes = make(map[string][]string)
	m.tagIndexes = make(map[string][]string)

	m.logger.Info("Memory storage closed")
	return nil
}