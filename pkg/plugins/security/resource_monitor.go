package security

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	PluginID       string    `json:"plugin_id"`
	ProcessID      int32     `json:"process_id"`
	CPUUsage       float64   `json:"cpu_usage"`        // CPU使用率百分比
	MemoryUsage    int64     `json:"memory_usage"`     // 内存使用量（字节）
	DiskUsage      int64     `json:"disk_usage"`       // 磁盘使用量（字节）
	NetworkRxBytes int64     `json:"network_rx_bytes"` // 网络接收字节数
	NetworkTxBytes int64     `json:"network_tx_bytes"` // 网络发送字节数
	FileHandles    int       `json:"file_handles"`     // 文件句柄数
	Connections    int       `json:"connections"`      // 连接数
	ExecutionTime  int64     `json:"execution_time"`   // 执行时间（秒）
	ThreadCount    int32     `json:"thread_count"`     // 线程数
	LastUpdated    time.Time `json:"last_updated"`
}

// ResourceAlert 资源告警
type ResourceAlert struct {
	ID           string                 `json:"id"`
	PluginID     string                 `json:"plugin_id"`
	Type         string                 `json:"type"`          // cpu, memory, disk, network, file_handles, connections
	Threshold    float64                `json:"threshold"`     // 阈值（百分比或字节数）
	CurrentValue float64                `json:"current_value"` // 当前值（百分比或字节数）
	Severity     string                 `json:"severity"`      // low, medium, high, critical
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Resolved     bool                   `json:"resolved"`
	ResolvedAt   *time.Time             `json:"resolved_at,omitempty"`
}

// ResourceThreshold 资源阈值配置
type ResourceThreshold struct {
	Type     string  `json:"type"`
	Warning  float64 `json:"warning"`  // 警告阈值（百分比或字节数）
	Critical float64 `json:"critical"` // 严重阈值（百分比或字节数）
	Enabled  bool    `json:"enabled"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID         int32                   `json:"pid"`
	Name        string                  `json:"name"`
	Status      string                  `json:"status"`
	CreateTime  time.Time               `json:"create_time"`
	CPUPercent  float64                 `json:"cpu_percent"`
	MemoryInfo  *process.MemoryInfoStat `json:"memory_info"`
	IOCounters  *process.IOCountersStat `json:"io_counters"`
	NumThreads  int32                   `json:"num_threads"`
	NumFDs      int32                   `json:"num_fds"`
	Connections []net.ConnectionStat    `json:"connections"`
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	pluginProcesses map[string]*ProcessInfo       // 插件进程映射
	resourceUsage   map[string]*ResourceUsage     // 资源使用情况
	alerts          []*ResourceAlert              // 告警列表
	thresholds      map[string]*ResourceThreshold // 阈值配置
	mu              sync.RWMutex

	// 监控配置
	monitorInterval time.Duration
	alertRetention  time.Duration

	// 监控状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewResourceMonitor 创建新的资源监控器
func NewResourceMonitor() *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &ResourceMonitor{
		pluginProcesses: make(map[string]*ProcessInfo),
		resourceUsage:   make(map[string]*ResourceUsage),
		alerts:          make([]*ResourceAlert, 0),
		thresholds:      make(map[string]*ResourceThreshold),
		monitorInterval: 30 * time.Second,
		alertRetention:  24 * time.Hour,
		ctx:             ctx,
		cancel:          cancel,
	}

	// 初始化默认阈值配置
	rm.initializeDefaultThresholds()

	return rm
}

// Start 启动资源监控
func (rm *ResourceMonitor) Start() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.running {
		return fmt.Errorf("resource monitor is already running")
	}

	rm.running = true

	// 启动监控协程
	go rm.monitorLoop()
	go rm.alertCleanupLoop()

	return nil
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.running {
		return fmt.Errorf("resource monitor is not running")
	}

	rm.running = false
	rm.cancel()

	return nil
}

// RegisterPlugin 注册插件进程
func (rm *ResourceMonitor) RegisterPlugin(pluginID string, pid int32) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to get process info: %v", err)
	}

	name, _ := proc.Name()
	status, _ := proc.Status()
	createTime, _ := proc.CreateTime()

	processInfo := &ProcessInfo{
		PID:        pid,
		Name:       name,
		Status:     status,
		CreateTime: time.Unix(createTime/1000, 0),
	}

	rm.pluginProcesses[pluginID] = processInfo

	// 初始化资源使用情况
	rm.resourceUsage[pluginID] = &ResourceUsage{
		PluginID:    pluginID,
		ProcessID:   pid,
		LastUpdated: time.Now(),
	}

	return nil
}

// UnregisterPlugin 注销插件进程
func (rm *ResourceMonitor) UnregisterPlugin(pluginID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.pluginProcesses, pluginID)
	delete(rm.resourceUsage, pluginID)

	// 解决相关告警
	for _, alert := range rm.alerts {
		if alert.PluginID == pluginID && !alert.Resolved {
			alert.Resolved = true
			now := time.Now()
			alert.ResolvedAt = &now
		}
	}
}

// GetResourceUsage 获取插件资源使用情况
func (rm *ResourceMonitor) GetResourceUsage(pluginID string) (*ResourceUsage, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	usage, exists := rm.resourceUsage[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return usage, nil
}

// GetAllResourceUsage 获取所有插件资源使用情况
func (rm *ResourceMonitor) GetAllResourceUsage() map[string]*ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*ResourceUsage)
	for pluginID, usage := range rm.resourceUsage {
		result[pluginID] = usage
	}

	return result
}

// GetAlerts 获取告警列表
func (rm *ResourceMonitor) GetAlerts(pluginID string, resolved bool) []*ResourceAlert {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var result []*ResourceAlert
	for _, alert := range rm.alerts {
		if pluginID != "" && alert.PluginID != pluginID {
			continue
		}
		if alert.Resolved != resolved {
			continue
		}
		result = append(result, alert)
	}

	return result
}

// SetThreshold 设置资源阈值配置
func (rm *ResourceMonitor) SetThreshold(thresholdType string, warning, critical float64, enabled bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.thresholds[thresholdType] = &ResourceThreshold{
		Type:     thresholdType,
		Warning:  warning,
		Critical: critical,
		Enabled:  enabled,
	}
}

// GetThresholds 获取所有阈值配置
func (rm *ResourceMonitor) GetThresholds() map[string]*ResourceThreshold {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*ResourceThreshold)
	for thresholdType, threshold := range rm.thresholds {
		result[thresholdType] = threshold
	}

	return result
}

// GetSystemStats 获取系统统计信息
func (rm *ResourceMonitor) GetSystemStats() (map[string]interface{}, error) {
	// CPU信息
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}

	// 磁盘信息
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %v", err)
	}

	// 网络信息
	netIO, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %v", err)
	}

	stats := map[string]interface{}{
		"cpu": map[string]interface{}{
			"usage_percent": cpuPercent[0],
			"num_cores":     runtime.NumCPU(),
		},
		"memory": map[string]interface{}{
			"total":        memInfo.Total,
			"available":    memInfo.Available,
			"used":         memInfo.Used,
			"used_percent": memInfo.UsedPercent,
		},
		"disk": map[string]interface{}{
			"total":        diskInfo.Total,
			"free":         diskInfo.Free,
			"used":         diskInfo.Used,
			"used_percent": diskInfo.UsedPercent,
		},
		"goroutines": runtime.NumGoroutine(),
	}

	if len(netIO) > 0 {
		stats["network"] = map[string]interface{}{
			"bytes_sent": netIO[0].BytesSent,
			"bytes_recv": netIO[0].BytesRecv,
		}
	}

	return stats, nil
}

// 私有方法

// monitorLoop 监控循环
func (rm *ResourceMonitor) monitorLoop() {
	ticker := time.NewTicker(rm.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateResourceUsage()
			rm.checkThresholds()
		}
	}
}

// updateResourceUsage 更新资源使用情况
func (rm *ResourceMonitor) updateResourceUsage() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for pluginID, processInfo := range rm.pluginProcesses {
		usage, exists := rm.resourceUsage[pluginID]
		if !exists {
			continue
		}

		proc, err := process.NewProcess(processInfo.PID)
		if err != nil {
			continue
		}

		// 更新CPU使用情况
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			usage.CPUUsage = cpuPercent
		}

		// 更新内存使用情况
		if memInfo, err := proc.MemoryInfo(); err == nil {
			usage.MemoryUsage = int64(memInfo.RSS)
		}

		// 更新IO统计
		if ioCounters, err := proc.IOCounters(); err == nil {
			usage.NetworkRxBytes = int64(ioCounters.ReadBytes)
			usage.NetworkTxBytes = int64(ioCounters.WriteBytes)
		}

		// 更新文件句柄数
		if numFDs, err := proc.NumFDs(); err == nil {
			usage.FileHandles = int(numFDs)
		}

		// 更新连接数
		if connections, err := proc.Connections(); err == nil {
			usage.Connections = len(connections)
		}

		// 更新线程数
		if numThreads, err := proc.NumThreads(); err == nil {
			usage.ThreadCount = numThreads
		}

		// 更新执行时间
		if createTime, err := proc.CreateTime(); err == nil {
			usage.ExecutionTime = int64(time.Since(time.Unix(createTime/1000, 0)).Seconds())
		}

		usage.LastUpdated = time.Now()
	}
}

// checkThresholds 检查阈值
func (rm *ResourceMonitor) checkThresholds() {
	for pluginID, usage := range rm.resourceUsage {
		rm.checkCPUThreshold(pluginID, usage)
		rm.checkMemoryThreshold(pluginID, usage)
		rm.checkFileHandlesThreshold(pluginID, usage)
		rm.checkConnectionsThreshold(pluginID, usage)
	}
}

// checkCPUThreshold 检查CPU阈值
func (rm *ResourceMonitor) checkCPUThreshold(pluginID string, usage *ResourceUsage) {
	threshold, exists := rm.thresholds["cpu"]
	if !exists || !threshold.Enabled {
		return
	}

	var severity string
	var message string

	if usage.CPUUsage >= threshold.Critical {
		severity = "critical"
		message = fmt.Sprintf("CPU usage is critically high: %.2f%%", usage.CPUUsage)
	} else if usage.CPUUsage >= threshold.Warning {
		severity = "warning"
		message = fmt.Sprintf("CPU usage is high: %.2f%%", usage.CPUUsage)
	} else {
		// 检查是否需要解决现有警告
		rm.resolveAlert(pluginID, "cpu")
		return
	}

	rm.createAlert(pluginID, "cpu", threshold.Critical, usage.CPUUsage, severity, message, nil)
}

// checkMemoryThreshold 检查内存阈值
func (rm *ResourceMonitor) checkMemoryThreshold(pluginID string, usage *ResourceUsage) {
	threshold, exists := rm.thresholds["memory"]
	if !exists || !threshold.Enabled {
		return
	}

	memoryMB := float64(usage.MemoryUsage) / 1024 / 1024
	var severity string
	var message string

	if memoryMB >= threshold.Critical {
		severity = "critical"
		message = fmt.Sprintf("Memory usage is critically high: %.2f MB", memoryMB)
	} else if memoryMB >= threshold.Warning {
		severity = "warning"
		message = fmt.Sprintf("Memory usage is high: %.2f MB", memoryMB)
	} else {
		rm.resolveAlert(pluginID, "memory")
		return
	}

	rm.createAlert(pluginID, "memory", threshold.Critical, memoryMB, severity, message, nil)
}

// checkFileHandlesThreshold 检查文件句柄阈值
func (rm *ResourceMonitor) checkFileHandlesThreshold(pluginID string, usage *ResourceUsage) {
	threshold, exists := rm.thresholds["file_handles"]
	if !exists || !threshold.Enabled {
		return
	}

	fileHandles := float64(usage.FileHandles)
	var severity string
	var message string

	if fileHandles >= threshold.Critical {
		severity = "critical"
		message = fmt.Sprintf("File handles usage is critically high: %d", usage.FileHandles)
	} else if fileHandles >= threshold.Warning {
		severity = "warning"
		message = fmt.Sprintf("File handles usage is high: %d", usage.FileHandles)
	} else {
		rm.resolveAlert(pluginID, "file_handles")
		return
	}

	rm.createAlert(pluginID, "file_handles", threshold.Critical, fileHandles, severity, message, nil)
}

// checkConnectionsThreshold 检查连接数阈值
func (rm *ResourceMonitor) checkConnectionsThreshold(pluginID string, usage *ResourceUsage) {
	threshold, exists := rm.thresholds["connections"]
	if !exists || !threshold.Enabled {
		return
	}

	connections := float64(usage.Connections)
	var severity string
	var message string

	if connections >= threshold.Critical {
		severity = "critical"
		message = fmt.Sprintf("Connections count is critically high: %d", usage.Connections)
	} else if connections >= threshold.Warning {
		severity = "warning"
		message = fmt.Sprintf("Connections count is high: %d", usage.Connections)
	} else {
		rm.resolveAlert(pluginID, "connections")
		return
	}

	rm.createAlert(pluginID, "connections", threshold.Critical, connections, severity, message, nil)
}

// createAlert 创建告警
func (rm *ResourceMonitor) createAlert(pluginID, alertType string, threshold, currentValue float64, severity, message string, details map[string]interface{}) {
	// 检查是否已存在相同类型的未解决告警
	for _, alert := range rm.alerts {
		if alert.PluginID == pluginID && alert.Type == alertType && !alert.Resolved {
			// 更新现有告警
			alert.CurrentValue = currentValue
			alert.Severity = severity
			alert.Message = message
			alert.Timestamp = time.Now()
			return
		}
	}

	// 创建新告警
	alert := &ResourceAlert{
		ID:           rm.generateAlertID(),
		PluginID:     pluginID,
		Type:         alertType,
		Threshold:    threshold,
		CurrentValue: currentValue,
		Severity:     severity,
		Message:      message,
		Details:      details,
		Timestamp:    time.Now(),
		Resolved:     false,
	}

	rm.alerts = append(rm.alerts, alert)
}

// resolveAlert 解决告警
func (rm *ResourceMonitor) resolveAlert(pluginID, alertType string) {
	for _, alert := range rm.alerts {
		if alert.PluginID == pluginID && alert.Type == alertType && !alert.Resolved {
			alert.Resolved = true
			now := time.Now()
			alert.ResolvedAt = &now
		}
	}
}

// generateAlertID 生成告警ID
func (rm *ResourceMonitor) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// alertCleanupLoop 告警清理循环
func (rm *ResourceMonitor) alertCleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.cleanupOldAlerts()
		}
	}
}

// cleanupOldAlerts 清理旧告警
func (rm *ResourceMonitor) cleanupOldAlerts() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	cutoff := time.Now().Add(-rm.alertRetention)
	var filtered []*ResourceAlert

	for _, alert := range rm.alerts {
		if alert.Timestamp.After(cutoff) {
			filtered = append(filtered, alert)
		}
	}

	rm.alerts = filtered
}

// initializeDefaultThresholds 初始化默认阈值
func (rm *ResourceMonitor) initializeDefaultThresholds() {
	rm.thresholds["cpu"] = &ResourceThreshold{
		Type:     "cpu",
		Warning:  70.0,
		Critical: 90.0,
		Enabled:  true,
	}

	rm.thresholds["memory"] = &ResourceThreshold{
		Type:     "memory",
		Warning:  512.0,  // 512MB
		Critical: 1024.0, // 1GB
		Enabled:  true,
	}

	rm.thresholds["file_handles"] = &ResourceThreshold{
		Type:     "file_handles",
		Warning:  800.0,
		Critical: 950.0,
		Enabled:  true,
	}

	rm.thresholds["connections"] = &ResourceThreshold{
		Type:     "connections",
		Warning:  80.0,
		Critical: 95.0,
		Enabled:  true,
	}
}

// GetProcessInfo 获取进程信息
func (rm *ResourceMonitor) GetProcessInfo(pluginID string) (*ProcessInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	processInfo, exists := rm.pluginProcesses[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin process not found: %s", pluginID)
	}

	return processInfo, nil
}

// KillProcess 终止进程
func (rm *ResourceMonitor) KillProcess(pluginID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	processInfo, exists := rm.pluginProcesses[pluginID]
	if !exists {
		return fmt.Errorf("plugin process not found: %s", pluginID)
	}

	proc, err := process.NewProcess(processInfo.PID)
	if err != nil {
		return fmt.Errorf("failed to get process: %v", err)
	}

	return proc.Kill()
}

// ExportMetrics 导出指标数据
func (rm *ResourceMonitor) ExportMetrics() ([]byte, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	metrics := map[string]interface{}{
		"resource_usage": rm.resourceUsage,
		"alerts":         rm.alerts,
		"thresholds":     rm.thresholds,
		"timestamp":      time.Now(),
	}

	return json.Marshal(metrics)
}
