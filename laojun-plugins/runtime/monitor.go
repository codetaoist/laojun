package runtime

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Monitor 插件监控器接口
type Monitor interface {
	// StartMonitoring 开始监控插件
	StartMonitoring(pluginID string) error

	// StopMonitoring 停止监控插件
	StopMonitoring(pluginID string) error

	// GetPluginMetrics 获取插件指标
	GetPluginMetrics(pluginID string) (*PluginMetrics, error)

	// GetSystemMetrics 获取系统指标
	GetSystemMetrics() (*SystemMetrics, error)

	// GetAllMetrics 获取所有指标
	GetAllMetrics() (*MonitoringReport, error)

	// SetAlertThreshold 设置告警阈值
	SetAlertThreshold(pluginID string, threshold *AlertThreshold) error

	// GetAlerts 获取告警信息
	GetAlerts() ([]*Alert, error)

	// ClearAlerts 清除告警
	ClearAlerts(pluginID string) error

	// Start 启动监控器
	Start(ctx context.Context) error

	// Stop 停止监控器
	Stop(ctx context.Context) error
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	PluginID        string            `json:"plugin_id"`
	CPUUsage        float64           `json:"cpu_usage"`        // CPU使用率 (%)
	MemoryUsage     int64             `json:"memory_usage"`     // 内存使用量 (bytes)
	MemoryPercent   float64           `json:"memory_percent"`   // 内存使用率 (%)
	GoroutineCount  int               `json:"goroutine_count"`  // Goroutine数量
	RequestCount    int64             `json:"request_count"`    // 请求总数
	ErrorCount      int64             `json:"error_count"`      // 错误总数
	ResponseTime    time.Duration     `json:"response_time"`    // 平均响应时间
	Uptime          time.Duration     `json:"uptime"`           // 运行时间
	LastUpdated     time.Time         `json:"last_updated"`     // 最后更新时间
	CustomMetrics   map[string]interface{} `json:"custom_metrics,omitempty"` // 自定义指标
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUCount        int               `json:"cpu_count"`
	CPUUsage        float64           `json:"cpu_usage"`
	MemoryTotal     int64             `json:"memory_total"`
	MemoryUsed      int64             `json:"memory_used"`
	MemoryPercent   float64           `json:"memory_percent"`
	GoroutineCount  int               `json:"goroutine_count"`
	PluginCount     int               `json:"plugin_count"`
	ActivePlugins   int               `json:"active_plugins"`
	Timestamp       time.Time         `json:"timestamp"`
}

// MonitoringReport 监控报告
type MonitoringReport struct {
	SystemMetrics   *SystemMetrics             `json:"system_metrics"`
	PluginMetrics   map[string]*PluginMetrics `json:"plugin_metrics"`
	Alerts          []*Alert                   `json:"alerts"`
	GeneratedAt     time.Time                  `json:"generated_at"`
}

// AlertThreshold 告警阈值
type AlertThreshold struct {
	PluginID        string  `json:"plugin_id"`
	CPUThreshold    float64 `json:"cpu_threshold"`     // CPU使用率阈值 (%)
	MemoryThreshold int64   `json:"memory_threshold"`  // 内存使用量阈值 (bytes)
	ErrorThreshold  int64   `json:"error_threshold"`   // 错误数量阈值
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"` // 响应时间阈值
	Enabled         bool    `json:"enabled"`
}

// Alert 告警信息
type Alert struct {
	ID          string      `json:"id"`
	PluginID    string      `json:"plugin_id"`
	Type        AlertType   `json:"type"`
	Level       AlertLevel  `json:"level"`
	Message     string      `json:"message"`
	Value       interface{} `json:"value"`
	Threshold   interface{} `json:"threshold"`
	Timestamp   time.Time   `json:"timestamp"`
	Resolved    bool        `json:"resolved"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}

// AlertType 告警类型
type AlertType string

const (
	AlertTypeCPU          AlertType = "cpu"
	AlertTypeMemory       AlertType = "memory"
	AlertTypeError        AlertType = "error"
	AlertTypeResponseTime AlertType = "response_time"
	AlertTypeHealth       AlertType = "health"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// DefaultMonitor 默认监控器实现
type DefaultMonitor struct {
	pluginManager   PluginManager
	pluginMetrics   map[string]*PluginMetrics
	alertThresholds map[string]*AlertThreshold
	alerts          []*Alert
	monitorInterval time.Duration
	logger          *logrus.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex
	running         bool
}

// NewDefaultMonitor 创建默认监控器
func NewDefaultMonitor(
	pluginManager PluginManager,
	monitorInterval time.Duration,
	logger *logrus.Logger,
) *DefaultMonitor {
	if monitorInterval <= 0 {
		monitorInterval = 30 * time.Second
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &DefaultMonitor{
		pluginManager:   pluginManager,
		pluginMetrics:   make(map[string]*PluginMetrics),
		alertThresholds: make(map[string]*AlertThreshold),
		alerts:          make([]*Alert, 0),
		monitorInterval: monitorInterval,
		logger:          logger,
	}
}

// Start 启动监控器
func (m *DefaultMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("monitor already running")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true

	// 启动监控协程
	m.wg.Add(1)
	go m.monitorLoop()

	m.logger.Info("Plugin monitor started")
	return nil
}

// Stop 停止监控器
func (m *DefaultMonitor) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("monitor not running")
	}

	m.cancel()
	m.running = false

	// 等待监控协程结束
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Plugin monitor stopped")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for monitor to stop")
	}
}

// StartMonitoring 开始监控插件
func (m *DefaultMonitor) StartMonitoring(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pluginMetrics[pluginID]; exists {
		return fmt.Errorf("plugin %s already being monitored", pluginID)
	}

	m.pluginMetrics[pluginID] = &PluginMetrics{
		PluginID:    pluginID,
		LastUpdated: time.Now(),
	}

	m.logger.Infof("Started monitoring plugin: %s", pluginID)
	return nil
}

// StopMonitoring 停止监控插件
func (m *DefaultMonitor) StopMonitoring(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.pluginMetrics, pluginID)
	delete(m.alertThresholds, pluginID)

	// 清除相关告警
	var filteredAlerts []*Alert
	for _, alert := range m.alerts {
		if alert.PluginID != pluginID {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}
	m.alerts = filteredAlerts

	m.logger.Infof("Stopped monitoring plugin: %s", pluginID)
	return nil
}

// GetPluginMetrics 获取插件指标
func (m *DefaultMonitor) GetPluginMetrics(pluginID string) (*PluginMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, exists := m.pluginMetrics[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not being monitored", pluginID)
	}

	// 返回副本
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// GetSystemMetrics 获取系统指标
func (m *DefaultMonitor) GetSystemMetrics() (*SystemMetrics, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.RLock()
	pluginCount := len(m.pluginMetrics)
	activePlugins := 0
	for _, metrics := range m.pluginMetrics {
		if time.Since(metrics.LastUpdated) < m.monitorInterval*2 {
			activePlugins++
		}
	}
	m.mu.RUnlock()

	return &SystemMetrics{
		CPUCount:       runtime.NumCPU(),
		MemoryTotal:    int64(memStats.Sys),
		MemoryUsed:     int64(memStats.Alloc),
		MemoryPercent:  float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		GoroutineCount: runtime.NumGoroutine(),
		PluginCount:    pluginCount,
		ActivePlugins:  activePlugins,
		Timestamp:      time.Now(),
	}, nil
}

// GetAllMetrics 获取所有指标
func (m *DefaultMonitor) GetAllMetrics() (*MonitoringReport, error) {
	systemMetrics, err := m.GetSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}

	m.mu.RLock()
	pluginMetrics := make(map[string]*PluginMetrics)
	for id, metrics := range m.pluginMetrics {
		metricsCopy := *metrics
		pluginMetrics[id] = &metricsCopy
	}
	alertsCopy := make([]*Alert, len(m.alerts))
	copy(alertsCopy, m.alerts)
	m.mu.RUnlock()

	return &MonitoringReport{
		SystemMetrics: systemMetrics,
		PluginMetrics: pluginMetrics,
		Alerts:        alertsCopy,
		GeneratedAt:   time.Now(),
	}, nil
}

// SetAlertThreshold 设置告警阈值
func (m *DefaultMonitor) SetAlertThreshold(pluginID string, threshold *AlertThreshold) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	threshold.PluginID = pluginID
	m.alertThresholds[pluginID] = threshold

	m.logger.Infof("Set alert threshold for plugin: %s", pluginID)
	return nil
}

// GetAlerts 获取告警信息
func (m *DefaultMonitor) GetAlerts() ([]*Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alertsCopy := make([]*Alert, len(m.alerts))
	copy(alertsCopy, m.alerts)
	return alertsCopy, nil
}

// ClearAlerts 清除告警
func (m *DefaultMonitor) ClearAlerts(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, alert := range m.alerts {
		if alert.PluginID == pluginID && !alert.Resolved {
			alert.Resolved = true
			alert.ResolvedAt = &now
		}
	}

	m.logger.Infof("Cleared alerts for plugin: %s", pluginID)
	return nil
}

// monitorLoop 监控循环
func (m *DefaultMonitor) monitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
			m.checkAlerts()
		}
	}
}

// collectMetrics 收集指标
func (m *DefaultMonitor) collectMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for pluginID := range m.pluginMetrics {
		metrics := m.collectPluginMetrics(pluginID)
		if metrics != nil {
			m.pluginMetrics[pluginID] = metrics
		}
	}
}

// collectPluginMetrics 收集插件指标
func (m *DefaultMonitor) collectPluginMetrics(pluginID string) *PluginMetrics {
	// 这里应该从插件管理器获取实际的指标数据
	// 为了演示，我们返回模拟数据
	_, err := m.pluginManager.GetPlugin(pluginID)
	if err != nil {
		m.logger.Errorf("Failed to get plugin %s: %v", pluginID, err)
		return nil
	}

	// 简化的指标收集，不依赖未定义的方法
	return &PluginMetrics{
		PluginID:       pluginID,
		CPUUsage:       0.0, // 实际实现中应该测量真实的CPU使用率
		MemoryUsage:    0,   // 实际实现中应该测量真实的内存使用量
		MemoryPercent:  0.0,
		GoroutineCount: runtime.NumGoroutine(),
		RequestCount:   0,
		ErrorCount:     0,
		ResponseTime:   0,
		Uptime:         time.Since(time.Now().Add(-time.Hour)), // 模拟运行时间
		LastUpdated:    time.Now(),
		CustomMetrics:  make(map[string]interface{}),
	}
}

// checkAlerts 检查告警
func (m *DefaultMonitor) checkAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for pluginID, metrics := range m.pluginMetrics {
		threshold, exists := m.alertThresholds[pluginID]
		if !exists || !threshold.Enabled {
			continue
		}

		m.checkPluginAlerts(pluginID, metrics, threshold)
	}
}

// checkPluginAlerts 检查插件告警
func (m *DefaultMonitor) checkPluginAlerts(pluginID string, metrics *PluginMetrics, threshold *AlertThreshold) {
	// 检查CPU使用率
	if threshold.CPUThreshold > 0 && metrics.CPUUsage > threshold.CPUThreshold {
		m.addAlert(&Alert{
			ID:        m.generateAlertID(),
			PluginID:  pluginID,
			Type:      AlertTypeCPU,
			Level:     m.getAlertLevel(metrics.CPUUsage, threshold.CPUThreshold),
			Message:   fmt.Sprintf("High CPU usage: %.2f%%", metrics.CPUUsage),
			Value:     metrics.CPUUsage,
			Threshold: threshold.CPUThreshold,
			Timestamp: time.Now(),
		})
	}

	// 检查内存使用量
	if threshold.MemoryThreshold > 0 && metrics.MemoryUsage > threshold.MemoryThreshold {
		m.addAlert(&Alert{
			ID:        m.generateAlertID(),
			PluginID:  pluginID,
			Type:      AlertTypeMemory,
			Level:     m.getAlertLevel(float64(metrics.MemoryUsage), float64(threshold.MemoryThreshold)),
			Message:   fmt.Sprintf("High memory usage: %d bytes", metrics.MemoryUsage),
			Value:     metrics.MemoryUsage,
			Threshold: threshold.MemoryThreshold,
			Timestamp: time.Now(),
		})
	}

	// 检查错误数量
	if threshold.ErrorThreshold > 0 && metrics.ErrorCount > threshold.ErrorThreshold {
		m.addAlert(&Alert{
			ID:        m.generateAlertID(),
			PluginID:  pluginID,
			Type:      AlertTypeError,
			Level:     AlertLevelCritical,
			Message:   fmt.Sprintf("High error count: %d", metrics.ErrorCount),
			Value:     metrics.ErrorCount,
			Threshold: threshold.ErrorThreshold,
			Timestamp: time.Now(),
		})
	}

	// 检查响应时间
	if threshold.ResponseTimeThreshold > 0 && metrics.ResponseTime > threshold.ResponseTimeThreshold {
		m.addAlert(&Alert{
			ID:        m.generateAlertID(),
			PluginID:  pluginID,
			Type:      AlertTypeResponseTime,
			Level:     m.getAlertLevel(float64(metrics.ResponseTime), float64(threshold.ResponseTimeThreshold)),
			Message:   fmt.Sprintf("High response time: %v", metrics.ResponseTime),
			Value:     metrics.ResponseTime,
			Threshold: threshold.ResponseTimeThreshold,
			Timestamp: time.Now(),
		})
	}
}

// addAlert 添加告警
func (m *DefaultMonitor) addAlert(alert *Alert) {
	// 检查是否已存在相同的告警
	for _, existingAlert := range m.alerts {
		if existingAlert.PluginID == alert.PluginID &&
			existingAlert.Type == alert.Type &&
			!existingAlert.Resolved {
			return // 已存在相同告警，不重复添加
		}
	}

	m.alerts = append(m.alerts, alert)
	m.logger.Warnf("Alert generated: %s - %s", alert.PluginID, alert.Message)
}

// getAlertLevel 获取告警级别
func (m *DefaultMonitor) getAlertLevel(value, threshold float64) AlertLevel {
	ratio := value / threshold
	if ratio >= 2.0 {
		return AlertLevelCritical
	} else if ratio >= 1.5 {
		return AlertLevelWarning
	}
	return AlertLevelInfo
}

// generateAlertID 生成告警ID
func (m *DefaultMonitor) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}