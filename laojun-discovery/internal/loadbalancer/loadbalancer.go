package loadbalancer

import (
	"context"
	"errors"
	"hash/crc32"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codetaoist/laojun-discovery/internal/storage"
)

var (
	ErrNoHealthyInstances = errors.New("no healthy instances available")
	ErrInvalidAlgorithm   = errors.New("invalid load balancing algorithm")
)

// Algorithm 负载均衡算法类型
type Algorithm string

const (
	RoundRobin       Algorithm = "round_robin"
	WeightedRoundRobin Algorithm = "weighted_round_robin"
	LeastConnections Algorithm = "least_connections"
	Random           Algorithm = "random"
	WeightedRandom   Algorithm = "weighted_random"
	ConsistentHash   Algorithm = "consistent_hash"
	IPHash           Algorithm = "ip_hash"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// Select 选择一个服务实例
	Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error)
	// UpdateStats 更新实例统计信息
	UpdateStats(instanceID string, stats *InstanceStats)
	// GetStats 获取负载均衡统计信息
	GetStats() map[string]interface{}
}

// InstanceStats 实例统计信息
type InstanceStats struct {
	ActiveConnections int64
	TotalRequests     int64
	FailedRequests    int64
	ResponseTime      time.Duration
	LastUsed          time.Time
}

// Config 负载均衡器配置
type Config struct {
	Algorithm          Algorithm     `yaml:"algorithm"`
	HealthCheckEnabled bool          `yaml:"health_check_enabled"`
	StatsEnabled       bool          `yaml:"stats_enabled"`
	HashKey            string        `yaml:"hash_key"`
	Weights            map[string]int `yaml:"weights"`
}

// Manager 负载均衡管理器
type Manager struct {
	config      *Config
	balancers   map[Algorithm]LoadBalancer
	mu          sync.RWMutex
	stats       map[string]*InstanceStats
	statsEnabled bool
}

// NewManager 创建负载均衡管理器
func NewManager(config *Config) *Manager {
	manager := &Manager{
		config:       config,
		balancers:    make(map[Algorithm]LoadBalancer),
		stats:        make(map[string]*InstanceStats),
		statsEnabled: config.StatsEnabled,
	}

	// 注册所有负载均衡算法
	manager.registerBalancers()
	
	return manager
}

// registerBalancers 注册所有负载均衡器
func (m *Manager) registerBalancers() {
	m.balancers[RoundRobin] = NewRoundRobinBalancer()
	m.balancers[WeightedRoundRobin] = NewWeightedRoundRobinBalancer(m.config.Weights)
	m.balancers[LeastConnections] = NewLeastConnectionsBalancer(m)
	m.balancers[Random] = NewRandomBalancer()
	m.balancers[WeightedRandom] = NewWeightedRandomBalancer(m.config.Weights)
	m.balancers[ConsistentHash] = NewConsistentHashBalancer()
	m.balancers[IPHash] = NewIPHashBalancer()
}

// Select 选择服务实例
func (m *Manager) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// 过滤健康的实例
	healthyInstances := m.filterHealthyInstances(instances)
	if len(healthyInstances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// 获取负载均衡器
	balancer, exists := m.balancers[m.config.Algorithm]
	if !exists {
		return nil, ErrInvalidAlgorithm
	}

	// 选择实例
	selected, err := balancer.Select(ctx, healthyInstances, key)
	if err != nil {
		return nil, err
	}

	// 更新统计信息
	if m.statsEnabled {
		m.updateInstanceStats(selected.ID)
	}

	return selected, nil
}

// filterHealthyInstances 过滤健康的实例
func (m *Manager) filterHealthyInstances(instances []*storage.ServiceInstance) []*storage.ServiceInstance {
	if !m.config.HealthCheckEnabled {
		return instances
	}

	var healthy []*storage.ServiceInstance
	for _, instance := range instances {
		if instance.Health.Status == "passing" {
			healthy = append(healthy, instance)
		}
	}
	return healthy
}

// updateInstanceStats 更新实例统计信息
func (m *Manager) updateInstanceStats(instanceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.stats[instanceID]
	if !exists {
		stats = &InstanceStats{}
		m.stats[instanceID] = stats
	}

	atomic.AddInt64(&stats.TotalRequests, 1)
	stats.LastUsed = time.Now()
}

// UpdateStats 更新实例统计信息
func (m *Manager) UpdateStats(instanceID string, stats *InstanceStats) {
	if !m.statsEnabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.stats[instanceID]
	if !exists {
		existing = &InstanceStats{}
		m.stats[instanceID] = existing
	}

	existing.ActiveConnections = stats.ActiveConnections
	existing.FailedRequests = stats.FailedRequests
	existing.ResponseTime = stats.ResponseTime
}

// GetStats 获取统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["algorithm"] = string(m.config.Algorithm)
	stats["instances"] = make(map[string]*InstanceStats)

	for id, instanceStats := range m.stats {
		stats["instances"].(map[string]*InstanceStats)[id] = instanceStats
	}

	return stats
}

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	counter uint64
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Select 轮询选择实例
func (rb *RoundRobinBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	index := atomic.AddUint64(&rb.counter, 1) % uint64(len(instances))
	return instances[index], nil
}

// UpdateStats 更新统计信息
func (rb *RoundRobinBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	// 轮询算法不需要统计信息
}

// GetStats 获取统计信息
func (rb *RoundRobinBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"algorithm": "round_robin",
		"counter":   atomic.LoadUint64(&rb.counter),
	}
}

// WeightedRoundRobinBalancer 加权轮询负载均衡器
type WeightedRoundRobinBalancer struct {
	weights map[string]int
	mu      sync.RWMutex
}

// NewWeightedRoundRobinBalancer 创建加权轮询负载均衡器
func NewWeightedRoundRobinBalancer(weights map[string]int) *WeightedRoundRobinBalancer {
	if weights == nil {
		weights = make(map[string]int)
	}
	return &WeightedRoundRobinBalancer{
		weights: weights,
	}
}

// Select 加权轮询选择实例
func (wrb *WeightedRoundRobinBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	wrb.mu.RLock()
	defer wrb.mu.RUnlock()

	// 构建加权列表
	var weightedInstances []*storage.ServiceInstance
	for _, instance := range instances {
		weight := wrb.weights[instance.ID]
		if weight <= 0 {
			weight = 1 // 默认权重
		}
		
		for i := 0; i < weight; i++ {
			weightedInstances = append(weightedInstances, instance)
		}
	}

	if len(weightedInstances) == 0 {
		return instances[0], nil
	}

	// 使用时间戳作为随机种子
	index := time.Now().UnixNano() % int64(len(weightedInstances))
	return weightedInstances[index], nil
}

// UpdateStats 更新统计信息
func (wrb *WeightedRoundRobinBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	// 加权轮询算法不需要统计信息
}

// GetStats 获取统计信息
func (wrb *WeightedRoundRobinBalancer) GetStats() map[string]interface{} {
	wrb.mu.RLock()
	defer wrb.mu.RUnlock()

	return map[string]interface{}{
		"algorithm": "weighted_round_robin",
		"weights":   wrb.weights,
	}
}

// LeastConnectionsBalancer 最少连接负载均衡器
type LeastConnectionsBalancer struct {
	manager *Manager
}

// NewLeastConnectionsBalancer 创建最少连接负载均衡器
func NewLeastConnectionsBalancer(manager *Manager) *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{
		manager: manager,
	}
}

// Select 选择连接数最少的实例
func (lcb *LeastConnectionsBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	lcb.manager.mu.RLock()
	defer lcb.manager.mu.RUnlock()

	var selected *storage.ServiceInstance
	minConnections := int64(-1)

	for _, instance := range instances {
		stats, exists := lcb.manager.stats[instance.ID]
		connections := int64(0)
		if exists {
			connections = atomic.LoadInt64(&stats.ActiveConnections)
		}

		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			selected = instance
		}
	}

	if selected == nil {
		return instances[0], nil
	}

	return selected, nil
}

// UpdateStats 更新统计信息
func (lcb *LeastConnectionsBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	lcb.manager.UpdateStats(instanceID, stats)
}

// GetStats 获取统计信息
func (lcb *LeastConnectionsBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"algorithm": "least_connections",
	}
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewRandomBalancer 创建随机负载均衡器
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 随机选择实例
func (rb *RandomBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	index := rb.rand.Intn(len(instances))
	return instances[index], nil
}

// UpdateStats 更新统计信息
func (rb *RandomBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	// 随机算法不需要统计信息
}

// GetStats 获取统计信息
func (rb *RandomBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"algorithm": "random",
	}
}

// WeightedRandomBalancer 加权随机负载均衡器
type WeightedRandomBalancer struct {
	weights map[string]int
	mu      sync.RWMutex
	rand    *rand.Rand
}

// NewWeightedRandomBalancer 创建加权随机负载均衡器
func NewWeightedRandomBalancer(weights map[string]int) *WeightedRandomBalancer {
	if weights == nil {
		weights = make(map[string]int)
	}
	return &WeightedRandomBalancer{
		weights: weights,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 加权随机选择实例
func (wrb *WeightedRandomBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	// 计算总权重
	totalWeight := 0
	for _, instance := range instances {
		weight := wrb.weights[instance.ID]
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return instances[0], nil
	}

	// 随机选择
	randomWeight := wrb.rand.Intn(totalWeight)
	currentWeight := 0

	for _, instance := range instances {
		weight := wrb.weights[instance.ID]
		if weight <= 0 {
			weight = 1
		}
		currentWeight += weight
		if randomWeight < currentWeight {
			return instance, nil
		}
	}

	return instances[len(instances)-1], nil
}

// UpdateStats 更新统计信息
func (wrb *WeightedRandomBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	// 加权随机算法不需要统计信息
}

// GetStats 获取统计信息
func (wrb *WeightedRandomBalancer) GetStats() map[string]interface{} {
	wrb.mu.RLock()
	defer wrb.mu.RUnlock()

	return map[string]interface{}{
		"algorithm": "weighted_random",
		"weights":   wrb.weights,
	}
}

// ConsistentHashBalancer 一致性哈希负载均衡器
type ConsistentHashBalancer struct {
	mu    sync.RWMutex
	ring  map[uint32]*storage.ServiceInstance
	keys  []uint32
}

// NewConsistentHashBalancer 创建一致性哈希负载均衡器
func NewConsistentHashBalancer() *ConsistentHashBalancer {
	return &ConsistentHashBalancer{
		ring: make(map[uint32]*storage.ServiceInstance),
		keys: make([]uint32, 0),
	}
}

// Select 一致性哈希选择实例
func (chb *ConsistentHashBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	chb.mu.Lock()
	defer chb.mu.Unlock()

	// 重建哈希环
	chb.rebuildRing(instances)

	if len(chb.keys) == 0 {
		return instances[0], nil
	}

	// 计算key的哈希值
	hash := crc32.ChecksumIEEE([]byte(key))

	// 在环上查找最近的节点
	idx := sort.Search(len(chb.keys), func(i int) bool {
		return chb.keys[i] >= hash
	})

	if idx == len(chb.keys) {
		idx = 0
	}

	return chb.ring[chb.keys[idx]], nil
}

// rebuildRing 重建哈希环
func (chb *ConsistentHashBalancer) rebuildRing(instances []*storage.ServiceInstance) {
	chb.ring = make(map[uint32]*storage.ServiceInstance)
	chb.keys = make([]uint32, 0)

	for _, instance := range instances {
		// 为每个实例创建多个虚拟节点
		for i := 0; i < 100; i++ {
			virtualKey := fmt.Sprintf("%s:%d", instance.ID, i)
			hash := crc32.ChecksumIEEE([]byte(virtualKey))
			chb.ring[hash] = instance
			chb.keys = append(chb.keys, hash)
		}
	}

	sort.Slice(chb.keys, func(i, j int) bool {
		return chb.keys[i] < chb.keys[j]
	})
}

// UpdateStats 更新统计信息
func (chb *ConsistentHashBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	// 一致性哈希算法不需要统计信息
}

// GetStats 获取统计信息
func (chb *ConsistentHashBalancer) GetStats() map[string]interface{} {
	chb.mu.RLock()
	defer chb.mu.RUnlock()

	return map[string]interface{}{
		"algorithm":    "consistent_hash",
		"virtual_nodes": len(chb.keys),
		"instances":    len(chb.ring),
	}
}

// IPHashBalancer IP哈希负载均衡器
type IPHashBalancer struct{}

// NewIPHashBalancer 创建IP哈希负载均衡器
func NewIPHashBalancer() *IPHashBalancer {
	return &IPHashBalancer{}
}

// Select IP哈希选择实例
func (ihb *IPHashBalancer) Select(ctx context.Context, instances []*storage.ServiceInstance, key string) (*storage.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// 使用IP地址作为哈希key
	hash := crc32.ChecksumIEEE([]byte(key))
	index := hash % uint32(len(instances))
	
	return instances[index], nil
}

// UpdateStats 更新统计信息
func (ihb *IPHashBalancer) UpdateStats(instanceID string, stats *InstanceStats) {
	// IP哈希算法不需要统计信息
}

// GetStats 获取统计信息
func (ihb *IPHashBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"algorithm": "ip_hash",
	}
}