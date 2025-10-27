package collectors

import (
	"fmt"
	"sync"
	
	"go.uber.org/zap"
)

// CollectorManager 收集器管理器
type CollectorManager struct {
	mu         sync.RWMutex
	collectors map[string]Collector
	logger     *zap.Logger
	running    bool
}

// NewCollectorManager 创建收集器管理器
func NewCollectorManager(logger *zap.Logger) *CollectorManager {
	return &CollectorManager{
		collectors: make(map[string]Collector),
		logger:     logger,
	}
}

// RegisterCollector 注册收集器
func (cm *CollectorManager) RegisterCollector(name string, collector Collector) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if _, exists := cm.collectors[name]; exists {
		return fmt.Errorf("collector %s already registered", name)
	}
	
	cm.collectors[name] = collector
	cm.logger.Info("Collector registered", zap.String("name", name))
	return nil
}

// UnregisterCollector 注销收集器
func (cm *CollectorManager) UnregisterCollector(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	collector, exists := cm.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// 停止收集器
	if collector.IsRunning() {
		if err := collector.Stop(); err != nil {
			cm.logger.Error("Failed to stop collector during unregister", 
				zap.String("name", name), zap.Error(err))
		}
	}
	
	delete(cm.collectors, name)
	cm.logger.Info("Collector unregistered", zap.String("name", name))
	return nil
}

// StartCollector 启动指定收集器
func (cm *CollectorManager) StartCollector(name string) error {
	cm.mu.RLock()
	collector, exists := cm.collectors[name]
	cm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	if collector.IsRunning() {
		return fmt.Errorf("collector %s is already running", name)
	}
	
	if err := collector.Start(); err != nil {
		return fmt.Errorf("failed to start collector %s: %w", name, err)
	}
	
	cm.logger.Info("Collector started", zap.String("name", name))
	return nil
}

// StopCollector 停止指定收集器
func (cm *CollectorManager) StopCollector(name string) error {
	cm.mu.RLock()
	collector, exists := cm.collectors[name]
	cm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	if !collector.IsRunning() {
		return fmt.Errorf("collector %s is not running", name)
	}
	
	if err := collector.Stop(); err != nil {
		return fmt.Errorf("failed to stop collector %s: %w", name, err)
	}
	
	cm.logger.Info("Collector stopped", zap.String("name", name))
	return nil
}

// GetCollector 获取指定收集器
func (cm *CollectorManager) GetCollector(name string) (Collector, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	collector, exists := cm.collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector %s not found", name)
	}
	
	return collector, nil
}

// ListCollectors 列出所有收集器
func (cm *CollectorManager) ListCollectors() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	names := make([]string, 0, len(cm.collectors))
	for name := range cm.collectors {
		names = append(names, name)
	}
	
	return names
}

// GetCollectorStats 获取收集器统计信息
func (cm *CollectorManager) GetCollectorStats(name string) (CollectorStats, error) {
	cm.mu.RLock()
	collector, exists := cm.collectors[name]
	cm.mu.RUnlock()
	
	if !exists {
		return CollectorStats{}, fmt.Errorf("collector %s not found", name)
	}
	
	return collector.GetStats(), nil
}

// StartAll 启动所有收集器
func (cm *CollectorManager) StartAll() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.running = true
	
	for name, collector := range cm.collectors {
		if !collector.IsRunning() {
			if err := collector.Start(); err != nil {
				cm.logger.Error("Failed to start collector", 
					zap.String("name", name), zap.Error(err))
				continue
			}
			cm.logger.Info("Collector started", zap.String("name", name))
		}
	}
	
	return nil
}

// StopAll 停止所有收集器
func (cm *CollectorManager) StopAll() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.running = false
	
	for name, collector := range cm.collectors {
		if collector.IsRunning() {
			if err := collector.Stop(); err != nil {
				cm.logger.Error("Failed to stop collector", 
					zap.String("name", name), zap.Error(err))
				continue
			}
			cm.logger.Info("Collector stopped", zap.String("name", name))
		}
	}
	
	return nil
}

// IsHealthy 检查所有收集器是否健康
func (cm *CollectorManager) IsHealthy() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	for _, collector := range cm.collectors {
		if !collector.IsHealthy() {
			return false
		}
	}
	
	return true
}

// GetAllStats 获取所有收集器统计信息
func (cm *CollectorManager) GetAllStats() map[string]CollectorStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	stats := make(map[string]CollectorStats)
	for name, collector := range cm.collectors {
		stats[name] = collector.GetStats()
	}
	
	return stats
}