package gateway

import (
	"math/rand"
	"sync"
	"time"
)

// NewLoadBalancer 创建新的负载均衡器
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		strategy: &RoundRobinStrategy{},
	}
}

// SetStrategy 设置负载均衡策略
func (lb *LoadBalancer) SetStrategy(strategy LoadBalanceStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.strategy = strategy
}

// SelectInstance 选择实例
func (lb *LoadBalancer) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	// 过滤健康的实例
	healthyInstances := make([]*PluginInstance, 0)
	for _, instance := range instances {
		if instance.Status == "running" &&
			instance.Health != nil &&
			instance.Health.Status == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil
	}

	lb.mu.RLock()
	strategy := lb.strategy
	lb.mu.RUnlock()

	return strategy.SelectInstance(healthyInstances)
}

// RoundRobinStrategy 轮询策略实现
func (rr *RoundRobinStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	rr.mu.Lock()
	defer rr.mu.Unlock()

	instance := instances[rr.current%len(instances)]
	rr.current++
	return instance
}

// WeightedRoundRobinStrategy 加权轮询策略实现
func NewWeightedRoundRobinStrategy(weights map[string]int) *WeightedRoundRobinStrategy {
	return &WeightedRoundRobinStrategy{
		weights: weights,
		current: make(map[string]int),
	}
}

func (wrr *WeightedRoundRobinStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	// 如果没有权重配置，使用默认权重
	totalWeight := 0
	for _, instance := range instances {
		weight := wrr.weights[instance.ID]
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return instances[0]
	}

	// 计算当前权重
	var selectedInstance *PluginInstance
	maxCurrentWeight := -1

	for _, instance := range instances {
		weight := wrr.weights[instance.ID]
		if weight <= 0 {
			weight = 1
		}

		wrr.current[instance.ID] += weight

		if wrr.current[instance.ID] > maxCurrentWeight {
			maxCurrentWeight = wrr.current[instance.ID]
			selectedInstance = instance
		}
	}

	if selectedInstance != nil {
		wrr.current[selectedInstance.ID] -= totalWeight
	}

	return selectedInstance
}

// LeastConnectionsStrategy 最少连接策略实现
func (lc *LeastConnectionsStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	var selectedInstance *PluginInstance
	minConnections := int64(-1)

	for _, instance := range instances {
		connections := instance.Metrics.ActiveRequests
		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			selectedInstance = instance
		}
	}

	return selectedInstance
}

// RandomStrategy 随机策略
type RandomStrategy struct {
	rand *rand.Rand
	mu   sync.Mutex
}

func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (rs *RandomStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	index := rs.rand.Intn(len(instances))
	return instances[index]
}

// WeightedRandomStrategy 加权随机策略
type WeightedRandomStrategy struct {
	weights map[string]int
	rand    *rand.Rand
	mu      sync.Mutex
}

func NewWeightedRandomStrategy(weights map[string]int) *WeightedRandomStrategy {
	return &WeightedRandomStrategy{
		weights: weights,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (wrs *WeightedRandomStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	wrs.mu.Lock()
	defer wrs.mu.Unlock()

	// 计算总权重
	totalWeight := 0
	for _, instance := range instances {
		weight := wrs.weights[instance.ID]
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return instances[0]
	}

	// 生成随机权重
	randomWeight := wrs.rand.Intn(totalWeight)

	// 选择实例
	currentWeight := 0
	for _, instance := range instances {
		weight := wrs.weights[instance.ID]
		if weight <= 0 {
			weight = 1
		}
		currentWeight += weight

		if randomWeight < currentWeight {
			return instance
		}
	}

	return instances[len(instances)-1]
}

// IPHashStrategy IP哈希策略
type IPHashStrategy struct {
	mu sync.RWMutex
}

func NewIPHashStrategy() *IPHashStrategy {
	return &IPHashStrategy{}
}

func (iph *IPHashStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	// 注意：这里需要客户端IP，但在LoadBalanceStrategy接口中没有传递上下文
	// 在实际使用中，可能需要修改接口或使用其他方式获取IP
	// 这里使用简单的轮询作为fallback
	return instances[0]
}

// ConsistentHashStrategy 一致性哈希策略
type ConsistentHashStrategy struct {
	hashRing map[uint32]*PluginInstance
	keys     []uint32
	replicas int
	mu       sync.RWMutex
}

func NewConsistentHashStrategy(replicas int) *ConsistentHashStrategy {
	return &ConsistentHashStrategy{
		hashRing: make(map[uint32]*PluginInstance),
		replicas: replicas,
	}
}

func (ch *ConsistentHashStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 重建哈希环
	ch.hashRing = make(map[uint32]*PluginInstance)
	ch.keys = make([]uint32, 0)

	for _, instance := range instances {
		for i := 0; i < ch.replicas; i++ {
			key := ch.hash(instance.ID + string(rune(i)))
			ch.hashRing[key] = instance
			ch.keys = append(ch.keys, key)
		}
	}

	// 排序keys
	for i := 0; i < len(ch.keys)-1; i++ {
		for j := i + 1; j < len(ch.keys); j++ {
			if ch.keys[i] > ch.keys[j] {
				ch.keys[i], ch.keys[j] = ch.keys[j], ch.keys[i]
			}
		}
	}

	if len(ch.keys) == 0 {
		return instances[0]
	}

	// 简单选择第一个实例（在实际使用中应该基于请求的某个特征进行哈希）
	return ch.hashRing[ch.keys[0]]
}

func (ch *ConsistentHashStrategy) hash(key string) uint32 {
	// 简单的哈希函数
	var hash uint32
	for _, c := range key {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// HealthAwareStrategy 健康感知策略
type HealthAwareStrategy struct {
	baseStrategy LoadBalanceStrategy
	mu           sync.RWMutex
}

func NewHealthAwareStrategy(baseStrategy LoadBalanceStrategy) *HealthAwareStrategy {
	return &HealthAwareStrategy{
		baseStrategy: baseStrategy,
	}
}

func (ha *HealthAwareStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	// 按健康状态和响应时间排序
	healthyInstances := make([]*PluginInstance, 0)
	for _, instance := range instances {
		if instance.Status == "running" &&
			instance.Health != nil &&
			instance.Health.Status == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil
	}

	// 按响应时间排序（简单的冒泡排序）
	for i := 0; i < len(healthyInstances)-1; i++ {
		for j := i + 1; j < len(healthyInstances); j++ {
			if healthyInstances[i].Health.ResponseTime > healthyInstances[j].Health.ResponseTime {
				healthyInstances[i], healthyInstances[j] = healthyInstances[j], healthyInstances[i]
			}
		}
	}

	// 选择响应时间最短的前50%实例
	topCount := len(healthyInstances) / 2
	if topCount == 0 {
		topCount = 1
	}

	topInstances := healthyInstances[:topCount]

	ha.mu.RLock()
	strategy := ha.baseStrategy
	ha.mu.RUnlock()

	return strategy.SelectInstance(topInstances)
}

// AdaptiveStrategy 自适应策略
type AdaptiveStrategy struct {
	strategies map[string]LoadBalanceStrategy
	current    string
	metrics    map[string]*StrategyMetrics
	mu         sync.RWMutex
}

type StrategyMetrics struct {
	RequestCount   int64
	ErrorCount     int64
	AverageLatency float64
	LastUsed       time.Time
}

func NewAdaptiveStrategy() *AdaptiveStrategy {
	strategies := map[string]LoadBalanceStrategy{
		"round_robin":       &RoundRobinStrategy{},
		"least_connections": &LeastConnectionsStrategy{},
		"random":            NewRandomStrategy(),
	}

	metrics := make(map[string]*StrategyMetrics)
	for name := range strategies {
		metrics[name] = &StrategyMetrics{}
	}

	return &AdaptiveStrategy{
		strategies: strategies,
		current:    "round_robin",
		metrics:    metrics,
	}
}

func (as *AdaptiveStrategy) SelectInstance(instances []*PluginInstance) *PluginInstance {
	if len(instances) == 0 {
		return nil
	}

	as.mu.Lock()
	defer as.mu.Unlock()

	// 每100次请求评估一次策略性能
	totalRequests := int64(0)
	for _, metric := range as.metrics {
		totalRequests += metric.RequestCount
	}

	if totalRequests > 0 && totalRequests%100 == 0 {
		as.evaluateStrategies()
	}

	strategy := as.strategies[as.current]
	as.metrics[as.current].RequestCount++
	as.metrics[as.current].LastUsed = time.Now()

	return strategy.SelectInstance(instances)
}

func (as *AdaptiveStrategy) evaluateStrategies() {
	bestStrategy := as.current
	bestScore := as.calculateScore(as.current)

	for name := range as.strategies {
		score := as.calculateScore(name)
		if score > bestScore {
			bestScore = score
			bestStrategy = name
		}
	}

	as.current = bestStrategy
}

func (as *AdaptiveStrategy) calculateScore(strategyName string) float64 {
	metric := as.metrics[strategyName]
	if metric.RequestCount == 0 {
		return 0
	}

	// 简单的评分算法：成功率 - 平均延迟（归一化）
	successRate := float64(metric.RequestCount-metric.ErrorCount) / float64(metric.RequestCount)
	latencyPenalty := metric.AverageLatency / 1000.0 // 假设延迟以毫秒为单位

	return successRate - latencyPenalty
}

// RecordError 记录错误
func (as *AdaptiveStrategy) RecordError(strategyName string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if metric, exists := as.metrics[strategyName]; exists {
		metric.ErrorCount++
	}
}

// RecordLatency 记录延迟
func (as *AdaptiveStrategy) RecordLatency(strategyName string, latency float64) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if metric, exists := as.metrics[strategyName]; exists {
		// 简单的移动平均
		if metric.AverageLatency == 0 {
			metric.AverageLatency = latency
		} else {
			metric.AverageLatency = (metric.AverageLatency + latency) / 2
		}
	}
}
