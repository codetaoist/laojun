package health

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// DefaultHealthManager 默认健康检查管理器实现
type DefaultHealthManager struct {
	config    HealthConfig
	checkers  map[string]HealthChecker
	listeners []HealthEventListener
	startTime time.Time
	mu        sync.RWMutex
	cache     *healthCache
	metrics   *healthMetrics
}

// NewHealthManager 创建健康检查管理器
func NewHealthManager(config HealthConfig) HealthManager {
	manager := &DefaultHealthManager{
		config:    config,
		checkers:  make(map[string]HealthChecker),
		listeners: make([]HealthEventListener, 0),
		startTime: time.Now(),
	}

	// 初始化缓存
	if config.Cache.Enabled {
		manager.cache = newHealthCache(config.Cache)
	}

	// 初始化指标
	if config.Metrics.Enabled {
		manager.metrics = newHealthMetrics(config.Metrics)
	}

	return manager
}

// AddChecker 添加检查器
func (m *DefaultHealthManager) AddChecker(checker HealthChecker) error {
	if checker == nil {
		return fmt.Errorf("checker cannot be nil")
	}

	name := checker.Name()
	if name == "" {
		return fmt.Errorf("checker name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if _, exists := m.checkers[name]; exists {
		return fmt.Errorf("checker with name '%s' already exists", name)
	}

	m.checkers[name] = checker

	// 通知监听器
	for _, listener := range m.listeners {
		listener.OnCheckerAdded(checker)
	}

	return nil
}

// RemoveChecker 移除检查器
func (m *DefaultHealthManager) RemoveChecker(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.checkers[name]; !exists {
		return fmt.Errorf("checker with name '%s' not found", name)
	}

	delete(m.checkers, name)

	// 通知监听器
	for _, listener := range m.listeners {
		listener.OnCheckerRemoved(name)
	}

	return nil
}

// Check 执行所有健康检查
func (m *DefaultHealthManager) Check(ctx context.Context) HealthReport {
	return m.checkWithFilter(ctx, func(checker HealthChecker) bool {
		return true
	})
}

// CheckByType 按类型执行健康检查
func (m *DefaultHealthManager) CheckByType(ctx context.Context, checkerType CheckerType) HealthReport {
	return m.checkWithFilter(ctx, func(checker HealthChecker) bool {
		return checker.Type() == checkerType
	})
}

// CheckByPriority 按优先级执行健康检查
func (m *DefaultHealthManager) CheckByPriority(ctx context.Context, priority Priority) HealthReport {
	return m.checkWithFilter(ctx, func(checker HealthChecker) bool {
		return checker.Priority() >= priority
	})
}

// checkWithFilter 使用过滤器执行健康检查
func (m *DefaultHealthManager) checkWithFilter(ctx context.Context, filter func(HealthChecker) bool) HealthReport {
	start := time.Now()

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	// 获取符合条件的检查器
	m.mu.RLock()
	var checkers []HealthChecker
	for _, checker := range m.checkers {
		if filter(checker) {
			checkers = append(checkers, checker)
		}
	}
	m.mu.RUnlock()

	// 按优先级排序
	sort.Slice(checkers, func(i, j int) bool {
		return checkers[i].Priority() > checkers[j].Priority()
	})

	// 并发执行检查
	results := make(map[string]CheckResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, checker := range checkers {
		wg.Add(1)
		go func(checker HealthChecker) {
			defer wg.Done()

			// 检查缓存
			if m.cache != nil {
				if cached, found := m.cache.Get(checker.Name()); found {
					mu.Lock()
					results[checker.Name()] = cached
					mu.Unlock()
					return
				}
			}

			// 执行检查
			result := m.executeCheck(ctx, checker)

			// 缓存结果
			if m.cache != nil {
				m.cache.Set(checker.Name(), result)
			}

			mu.Lock()
			results[checker.Name()] = result
			mu.Unlock()

			// 通知监听器
			for _, listener := range m.listeners {
				listener.OnCheckCompleted(result)
			}

			// 记录指标
			if m.metrics != nil {
				m.metrics.RecordCheck(result)
			}
		}(checker)
	}

	wg.Wait()

	// 计算总体状态
	overallStatus := m.calculateOverallStatus(results)

	// 计算摘要
	summary := m.calculateSummary(results)

	report := HealthReport{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Service: ServiceInfo{
			Name:        m.config.Service.Name,
			Version:     m.config.Service.Version,
			Environment: m.config.Service.Environment,
			StartTime:   m.startTime,
			Uptime:      time.Since(m.startTime),
		},
		Checks:  results,
		Summary: summary,
	}

	// 检查状态变化并发送事件
	m.checkStatusChange(report)

	return report
}

// executeCheck 执行单个检查
func (m *DefaultHealthManager) executeCheck(ctx context.Context, checker HealthChecker) CheckResult {
	defer func() {
		if r := recover(); r != nil {
			// 处理panic
			for _, listener := range m.listeners {
				listener.OnCheckFailed(checker.Name(), fmt.Errorf("panic: %v", r))
			}
		}
	}()

	return checker.Check(ctx)
}

// GetChecker 获取指定检查器
func (m *DefaultHealthManager) GetChecker(name string) (HealthChecker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	checker, exists := m.checkers[name]
	return checker, exists
}

// ListCheckers 列出所有检查器
func (m *DefaultHealthManager) ListCheckers() []HealthChecker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	checkers := make([]HealthChecker, 0, len(m.checkers))
	for _, checker := range m.checkers {
		checkers = append(checkers, checker)
	}
	
	// 按优先级排序
	sort.Slice(checkers, func(i, j int) bool {
		return checkers[i].Priority() > checkers[j].Priority()
	})
	
	return checkers
}

// AddListener 添加事件监听器
func (m *DefaultHealthManager) AddListener(listener HealthEventListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// RemoveListener 移除事件监听器
func (m *DefaultHealthManager) RemoveListener(listener HealthEventListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i, l := range m.listeners {
		if l == listener {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			break
		}
	}
}

// calculateOverallStatus 计算总体状态
func (m *DefaultHealthManager) calculateOverallStatus(results map[string]CheckResult) Status {
	if len(results) == 0 {
		return StatusHealthy
	}

	criticalUnhealthy := 0
	highUnhealthy := 0
	degraded := 0
	total := 0

	for _, result := range results {
		total++
		
		// 根据检查器获取优先级
		if checker, exists := m.checkers[result.Name]; exists {
			priority := checker.Priority()
			
			switch result.Status {
			case StatusUnhealthy:
				if priority == PriorityCritical {
					criticalUnhealthy++
				} else if priority == PriorityHigh {
					highUnhealthy++
				}
			case StatusDegraded, StatusUnknown:
				degraded++
			}
		}
	}

	// 如果有关键服务不健康，整体状态为不健康
	if criticalUnhealthy > 0 {
		return StatusUnhealthy
	}

	// 如果高优先级服务不健康超过阈值，整体状态为不健康
	if highUnhealthy > 0 && float64(highUnhealthy)/float64(total) > 0.5 {
		return StatusUnhealthy
	}

	// 如果有降级服务，整体状态为降级
	if degraded > 0 || highUnhealthy > 0 {
		return StatusDegraded
	}

	return StatusHealthy
}

// calculateSummary 计算摘要
func (m *DefaultHealthManager) calculateSummary(results map[string]CheckResult) Summary {
	summary := Summary{
		Total: len(results),
	}

	for _, result := range results {
		switch result.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusUnhealthy:
			summary.Unhealthy++
		case StatusDegraded:
			summary.Degraded++
		case StatusUnknown:
			summary.Unknown++
		}
	}

	return summary
}

// checkStatusChange 检查状态变化
func (m *DefaultHealthManager) checkStatusChange(report HealthReport) {
	// 这里可以实现状态变化检测和事件发送逻辑
	// 为了简化，暂时跳过实现
}

// healthCache 健康检查缓存
type healthCache struct {
	cache   map[string]cacheEntry
	ttl     time.Duration
	maxSize int
	mu      sync.RWMutex
}

type cacheEntry struct {
	result    CheckResult
	timestamp time.Time
}

func newHealthCache(config CacheConfig) *healthCache {
	return &healthCache{
		cache:   make(map[string]cacheEntry),
		ttl:     config.TTL,
		maxSize: config.MaxSize,
	}
}

func (c *healthCache) Get(key string) (CheckResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, exists := c.cache[key]
	if !exists {
		return CheckResult{}, false
	}
	
	// 检查是否过期
	if time.Since(entry.timestamp) > c.ttl {
		return CheckResult{}, false
	}
	
	return entry.result, true
}

func (c *healthCache) Set(key string, result CheckResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 检查缓存大小
	if len(c.cache) >= c.maxSize {
		// 简单的LRU实现：删除最旧的条目
		var oldestKey string
		var oldestTime time.Time
		
		for k, v := range c.cache {
			if oldestKey == "" || v.timestamp.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.timestamp
			}
		}
		
		if oldestKey != "" {
			delete(c.cache, oldestKey)
		}
	}
	
	c.cache[key] = cacheEntry{
		result:    result,
		timestamp: time.Now(),
	}
}

// healthMetrics 健康检查指标
type healthMetrics struct {
	config MetricsConfig
	// 这里可以添加具体的指标收集实现
}

func newHealthMetrics(config MetricsConfig) *healthMetrics {
	return &healthMetrics{
		config: config,
	}
}

func (m *healthMetrics) RecordCheck(result CheckResult) {
	// 这里可以实现指标记录逻辑
	// 例如记录检查耗时、成功率等
}