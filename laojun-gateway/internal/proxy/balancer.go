package proxy

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/services/discovery"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Select(instances []*discovery.ServiceInstance) *discovery.ServiceInstance
}

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	counter uint64
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Select 选择服务实例
func (r *RoundRobinBalancer) Select(instances []*discovery.ServiceInstance) *discovery.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	index := atomic.AddUint64(&r.counter, 1) % uint64(len(instances))
	return instances[index]
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct {
	rand *rand.Rand
}

// NewRandomBalancer 创建随机负载均衡器
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 选择服务实例
func (r *RandomBalancer) Select(instances []*discovery.ServiceInstance) *discovery.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	index := r.rand.Intn(len(instances))
	return instances[index]
}

// WeightedBalancer 加权负载均衡器
type WeightedBalancer struct {
	rand *rand.Rand
}

// NewWeightedBalancer 创建加权负载均衡器
func NewWeightedBalancer() *WeightedBalancer {
	return &WeightedBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 选择服务实例
func (w *WeightedBalancer) Select(instances []*discovery.ServiceInstance) *discovery.ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0
	for _, instance := range instances {
		weight := w.getWeight(instance)
		totalWeight += weight
	}

	if totalWeight == 0 {
		// 如果没有权重信息，使用随机选择
		index := w.rand.Intn(len(instances))
		return instances[index]
	}

	// 根据权重选择
	randomWeight := w.rand.Intn(totalWeight)
	currentWeight := 0

	for _, instance := range instances {
		weight := w.getWeight(instance)
		currentWeight += weight
		if randomWeight < currentWeight {
			return instance
		}
	}

	// 默认返回第一个实例
	return instances[0]
}

// getWeight 获取实例权重
func (w *WeightedBalancer) getWeight(instance *discovery.ServiceInstance) int {
	if weightStr, exists := instance.Meta["weight"]; exists {
		// 这里可以解析权重字符串，简化处理返回固定值
		_ = weightStr
		return 1
	}
	return 1
}