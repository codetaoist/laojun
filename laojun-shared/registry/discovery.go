package registry

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultServiceDiscovery 默认服务发现实现
type DefaultServiceDiscovery struct {
	registry ServiceRegistry
	logger   *zap.Logger
	mu       sync.RWMutex
	
	// 负载均衡状态
	roundRobinCounters map[string]int
	consistentHashRing map[string]*ConsistentHashRing
}

// NewDefaultServiceDiscovery 创建默认服务发现实例
func NewDefaultServiceDiscovery(registry ServiceRegistry, logger *zap.Logger) *DefaultServiceDiscovery {
	return &DefaultServiceDiscovery{
		registry:           registry,
		logger:             logger,
		roundRobinCounters: make(map[string]int),
		consistentHashRing: make(map[string]*ConsistentHashRing),
	}
}

// DiscoverServices 发现服务
func (d *DefaultServiceDiscovery) DiscoverServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	services, err := d.registry.GetHealthyServices(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	d.logger.Debug("Services discovered",
		zap.String("service_name", serviceName),
		zap.Int("count", len(services)))

	return services, nil
}

// DiscoverServicesByTag 根据标签发现服务
func (d *DefaultServiceDiscovery) DiscoverServicesByTag(ctx context.Context, tag string) ([]*ServiceInfo, error) {
	allServices, err := d.registry.ListAllServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all services: %w", err)
	}

	var matchedServices []*ServiceInfo
	for _, services := range allServices {
		for _, service := range services {
			if d.hasTag(service, tag) && d.isServiceHealthy(service) {
				matchedServices = append(matchedServices, service)
			}
		}
	}

	d.logger.Debug("Services discovered by tag",
		zap.String("tag", tag),
		zap.Int("count", len(matchedServices)))

	return matchedServices, nil
}

// DiscoverServicesByMeta 根据元数据发现服务
func (d *DefaultServiceDiscovery) DiscoverServicesByMeta(ctx context.Context, meta map[string]string) ([]*ServiceInfo, error) {
	allServices, err := d.registry.ListAllServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all services: %w", err)
	}

	var matchedServices []*ServiceInfo
	for _, services := range allServices {
		for _, service := range services {
			if d.hasMetadata(service, meta) && d.isServiceHealthy(service) {
				matchedServices = append(matchedServices, service)
			}
		}
	}

	d.logger.Debug("Services discovered by metadata",
		zap.Any("meta", meta),
		zap.Int("count", len(matchedServices)))

	return matchedServices, nil
}

// GetServiceEndpoint 获取服务端点
func (d *DefaultServiceDiscovery) GetServiceEndpoint(ctx context.Context, serviceName string, strategy LoadBalanceStrategy) (*ServiceInfo, error) {
	services, err := d.DiscoverServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy services found for: %s", serviceName)
	}

	service, err := d.selectService(services, serviceName, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to select service: %w", err)
	}

	d.logger.Debug("Service endpoint selected",
		zap.String("service_name", serviceName),
		zap.String("service_id", service.ID),
		zap.String("strategy", string(strategy)))

	return service, nil
}

// Subscribe 订阅服务变化事件
func (d *DefaultServiceDiscovery) Subscribe(ctx context.Context, serviceName string) (<-chan *ServiceEvent, error) {
	return d.registry.WatchServices(ctx, serviceName)
}

// Unsubscribe 取消订阅
func (d *DefaultServiceDiscovery) Unsubscribe(ctx context.Context, subscription <-chan *ServiceEvent) error {
	// 由于使用了context取消机制，这里不需要额外操作
	return nil
}

// 私有方法

// selectService 根据负载均衡策略选择服务
func (d *DefaultServiceDiscovery) selectService(services []*ServiceInfo, serviceName string, strategy LoadBalanceStrategy) (*ServiceInfo, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no services available")
	}

	switch strategy {
	case LoadBalanceRoundRobin:
		return d.selectRoundRobin(services, serviceName), nil
	case LoadBalanceRandom:
		return d.selectRandom(services), nil
	case LoadBalanceWeighted:
		return d.selectWeighted(services), nil
	case LoadBalanceLeastConn:
		return d.selectLeastConn(services), nil
	case LoadBalanceConsistentHash:
		return d.selectConsistentHash(services, serviceName), nil
	default:
		return d.selectRandom(services), nil
	}
}

// selectRoundRobin 轮询选择
func (d *DefaultServiceDiscovery) selectRoundRobin(services []*ServiceInfo, serviceName string) *ServiceInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	counter := d.roundRobinCounters[serviceName]
	service := services[counter%len(services)]
	d.roundRobinCounters[serviceName] = counter + 1

	return service
}

// selectRandom 随机选择
func (d *DefaultServiceDiscovery) selectRandom(services []*ServiceInfo) *ServiceInfo {
	return services[rand.Intn(len(services))]
}

// selectWeighted 加权选择
func (d *DefaultServiceDiscovery) selectWeighted(services []*ServiceInfo) *ServiceInfo {
	totalWeight := 0
	for _, service := range services {
		totalWeight += service.Weight
	}

	if totalWeight == 0 {
		return d.selectRandom(services)
	}

	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, service := range services {
		currentWeight += service.Weight
		if currentWeight > randomWeight {
			return service
		}
	}

	return services[len(services)-1]
}

// selectLeastConn 最少连接选择（简化实现，基于权重）
func (d *DefaultServiceDiscovery) selectLeastConn(services []*ServiceInfo) *ServiceInfo {
	// 简化实现：选择权重最高的服务（假设权重反映连接数）
	sort.Slice(services, func(i, j int) bool {
		return services[i].Weight > services[j].Weight
	})
	return services[0]
}

// selectConsistentHash 一致性哈希选择
func (d *DefaultServiceDiscovery) selectConsistentHash(services []*ServiceInfo, serviceName string) *ServiceInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	ring, exists := d.consistentHashRing[serviceName]
	if !exists || ring.Size() != len(services) {
		ring = NewConsistentHashRing()
		for _, service := range services {
			ring.Add(service.ID, service)
		}
		d.consistentHashRing[serviceName] = ring
	}

	service := ring.Get(serviceName)
	if service == nil {
		return d.selectRandom(services)
	}

	return service.(*ServiceInfo)
}

// hasTag 检查服务是否包含指定标签
func (d *DefaultServiceDiscovery) hasTag(service *ServiceInfo, tag string) bool {
	for _, t := range service.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// hasMetadata 检查服务是否包含指定元数据
func (d *DefaultServiceDiscovery) hasMetadata(service *ServiceInfo, meta map[string]string) bool {
	for key, value := range meta {
		if serviceValue, exists := service.Meta[key]; !exists || serviceValue != value {
			return false
		}
	}
	return true
}

// isServiceHealthy 检查服务是否健康（简化实现）
func (d *DefaultServiceDiscovery) isServiceHealthy(service *ServiceInfo) bool {
	return service.Status == ServiceStatusActive
}

// ConsistentHashRing 一致性哈希环
type ConsistentHashRing struct {
	nodes    map[uint32]*HashNode
	sortedKeys []uint32
	mu       sync.RWMutex
}

// HashNode 哈希节点
type HashNode struct {
	Key   string
	Value interface{}
}

// NewConsistentHashRing 创建一致性哈希环
func NewConsistentHashRing() *ConsistentHashRing {
	return &ConsistentHashRing{
		nodes: make(map[uint32]*HashNode),
	}
}

// Add 添加节点
func (r *ConsistentHashRing) Add(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hash := r.hash(key)
	r.nodes[hash] = &HashNode{Key: key, Value: value}
	r.sortedKeys = append(r.sortedKeys, hash)
	sort.Slice(r.sortedKeys, func(i, j int) bool {
		return r.sortedKeys[i] < r.sortedKeys[j]
	})
}

// Get 获取节点
func (r *ConsistentHashRing) Get(key string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.sortedKeys) == 0 {
		return nil
	}

	hash := r.hash(key)
	idx := sort.Search(len(r.sortedKeys), func(i int) bool {
		return r.sortedKeys[i] >= hash
	})

	if idx == len(r.sortedKeys) {
		idx = 0
	}

	return r.nodes[r.sortedKeys[idx]].Value
}

// Size 获取节点数量
func (r *ConsistentHashRing) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}

// hash 简单哈希函数
func (r *ConsistentHashRing) hash(key string) uint32 {
	h := uint32(0)
	for _, c := range key {
		h = h*31 + uint32(c)
	}
	return h
}