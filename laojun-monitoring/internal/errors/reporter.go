package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorReport 错误报告
type ErrorReport struct {
	ID          string                 `json:"id"`
	Error       *MonitoringError       `json:"error"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Version     string                 `json:"version,omitempty"`
}

// ErrorStatistics 错误统计信息
type ErrorStatistics struct {
	TotalErrors     int64                    `json:"total_errors"`
	ErrorsByType    map[ErrorType]int64      `json:"errors_by_type"`
	ErrorsBySeverity map[ErrorSeverity]int64 `json:"errors_by_severity"`
	ErrorsByComponent map[string]int64       `json:"errors_by_component"`
	ErrorsByCode    map[string]int64         `json:"errors_by_code"`
	LastUpdated     time.Time                `json:"last_updated"`
	TimeWindow      time.Duration            `json:"time_window"`
}

// ErrorReporter 错误报告器接口
type ErrorReporter interface {
	// ReportError 报告错误
	ReportError(ctx context.Context, report *ErrorReport) error
	// GetStatistics 获取错误统计信息
	GetStatistics(ctx context.Context, timeWindow time.Duration) (*ErrorStatistics, error)
	// Close 关闭报告器
	Close() error
}

// InMemoryErrorReporter 内存错误报告器
type InMemoryErrorReporter struct {
	reports []ErrorReport
	mu      sync.RWMutex
	logger  *zap.Logger
	maxSize int
}

// NewInMemoryErrorReporter 创建内存错误报告器
func NewInMemoryErrorReporter(logger *zap.Logger, maxSize int) *InMemoryErrorReporter {
	if maxSize <= 0 {
		maxSize = 1000 // 默认最大存储1000个错误报告
	}
	
	return &InMemoryErrorReporter{
		reports: make([]ErrorReport, 0, maxSize),
		logger:  logger,
		maxSize: maxSize,
	}
}

// ReportError 报告错误
func (r *InMemoryErrorReporter) ReportError(ctx context.Context, report *ErrorReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 如果达到最大容量，移除最旧的报告
	if len(r.reports) >= r.maxSize {
		r.reports = r.reports[1:]
	}
	
	// 添加新报告
	r.reports = append(r.reports, *report)
	
	r.logger.Debug("Error reported",
		zap.String("error_id", report.ID),
		zap.String("error_type", string(report.Error.Type)),
		zap.String("error_code", report.Error.Code),
		zap.String("component", report.Error.Component))
	
	return nil
}

// GetStatistics 获取错误统计信息
func (r *InMemoryErrorReporter) GetStatistics(ctx context.Context, timeWindow time.Duration) (*ErrorStatistics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	now := time.Now()
	cutoff := now.Add(-timeWindow)
	
	stats := &ErrorStatistics{
		ErrorsByType:      make(map[ErrorType]int64),
		ErrorsBySeverity:  make(map[ErrorSeverity]int64),
		ErrorsByComponent: make(map[string]int64),
		ErrorsByCode:      make(map[string]int64),
		LastUpdated:       now,
		TimeWindow:        timeWindow,
	}
	
	// 统计指定时间窗口内的错误
	for _, report := range r.reports {
		if report.Timestamp.After(cutoff) {
			stats.TotalErrors++
			stats.ErrorsByType[report.Error.Type]++
			stats.ErrorsBySeverity[report.Error.Severity]++
			stats.ErrorsByComponent[report.Error.Component]++
			stats.ErrorsByCode[report.Error.Code]++
		}
	}
	
	return stats, nil
}

// GetReports 获取错误报告列表
func (r *InMemoryErrorReporter) GetReports(ctx context.Context, timeWindow time.Duration) ([]ErrorReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	now := time.Now()
	cutoff := now.Add(-timeWindow)
	
	var filteredReports []ErrorReport
	for _, report := range r.reports {
		if report.Timestamp.After(cutoff) {
			filteredReports = append(filteredReports, report)
		}
	}
	
	return filteredReports, nil
}

// Close 关闭报告器
func (r *InMemoryErrorReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.reports = nil
	return nil
}

// FileErrorReporter 文件错误报告器
type FileErrorReporter struct {
	filePath string
	logger   *zap.Logger
	mu       sync.Mutex
}

// NewFileErrorReporter 创建文件错误报告器
func NewFileErrorReporter(filePath string, logger *zap.Logger) *FileErrorReporter {
	return &FileErrorReporter{
		filePath: filePath,
		logger:   logger,
	}
}

// ReportError 报告错误到文件
func (r *FileErrorReporter) ReportError(ctx context.Context, report *ErrorReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 将错误报告序列化为JSON
	data, err := json.Marshal(report)
	if err != nil {
		return NewError(ErrorTypeSystem, "REPORTER_001", "Failed to marshal error report").
			WithCause(err).
			WithComponent("FileErrorReporter").
			Build()
	}
	
	// 写入文件
	file, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return NewError(ErrorTypeSystem, "REPORTER_002", "Failed to open error log file").
			WithCause(err).
			WithComponent("FileErrorReporter").
			Build()
	}
	defer file.Close()
	
	// 写入JSON数据和换行符
	if _, err := file.Write(append(data, '\n')); err != nil {
		return NewError(ErrorTypeSystem, "REPORTER_003", "Failed to write error report to file").
			WithCause(err).
			WithComponent("FileErrorReporter").
			Build()
	}
	
	r.logger.Info("Error report written to file",
		zap.String("file_path", r.filePath),
		zap.String("error_id", report.ID))
	
	return nil
}

// GetStatistics 获取错误统计信息（文件版本需要读取文件）
func (r *FileErrorReporter) GetStatistics(ctx context.Context, timeWindow time.Duration) (*ErrorStatistics, error) {
	// 这里应该读取文件并统计，为了简化示例，返回空统计
	return &ErrorStatistics{
		ErrorsByType:      make(map[ErrorType]int64),
		ErrorsBySeverity:  make(map[ErrorSeverity]int64),
		ErrorsByComponent: make(map[string]int64),
		ErrorsByCode:      make(map[string]int64),
		LastUpdated:       time.Now(),
		TimeWindow:        timeWindow,
	}, nil
}

// Close 关闭文件报告器
func (r *FileErrorReporter) Close() error {
	return nil
}

// ErrorMonitor 错误监控器
type ErrorMonitor struct {
	reporter        ErrorReporter
	recoveryManager *RecoveryManager
	logger          *zap.Logger
	
	// 错误阈值配置
	thresholds map[ErrorType]ErrorThreshold
	mu         sync.RWMutex
	
	// 监控状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// ErrorThreshold 错误阈值配置
type ErrorThreshold struct {
	MaxErrorsPerMinute int           `json:"max_errors_per_minute"`
	MaxErrorsPerHour   int           `json:"max_errors_per_hour"`
	AlertSeverity      ErrorSeverity `json:"alert_severity"`
}

// NewErrorMonitor 创建错误监控器
func NewErrorMonitor(reporter ErrorReporter, recoveryManager *RecoveryManager, logger *zap.Logger) *ErrorMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ErrorMonitor{
		reporter:        reporter,
		recoveryManager: recoveryManager,
		logger:          logger,
		thresholds:      make(map[ErrorType]ErrorThreshold),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// SetThreshold 设置错误阈值
func (m *ErrorMonitor) SetThreshold(errorType ErrorType, threshold ErrorThreshold) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thresholds[errorType] = threshold
}

// HandleError 处理错误（包括报告和恢复）
func (m *ErrorMonitor) HandleError(ctx context.Context, err *MonitoringError) error {
	// 生成错误报告
	report := &ErrorReport{
		ID:        generateErrorID(),
		Error:     err,
		Timestamp: time.Now(),
		Context:   err.Context,
	}
	
	// 报告错误
	if reportErr := m.reporter.ReportError(ctx, report); reportErr != nil {
		m.logger.Error("Failed to report error",
			zap.String("error_code", err.Code),
			zap.Error(reportErr))
	}
	
	// 检查错误阈值
	if m.checkThresholds(ctx, err) {
		m.logger.Warn("Error threshold exceeded",
			zap.String("error_type", string(err.Type)),
			zap.String("error_code", err.Code))
	}
	
	// 尝试恢复
	if m.recoveryManager != nil {
		return m.recoveryManager.HandleError(ctx, err)
	}
	
	return err
}

// checkThresholds 检查错误阈值
func (m *ErrorMonitor) checkThresholds(ctx context.Context, err *MonitoringError) bool {
	m.mu.RLock()
	threshold, exists := m.thresholds[err.Type]
	m.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// 获取最近1分钟和1小时的错误统计
	minuteStats, _ := m.reporter.GetStatistics(ctx, time.Minute)
	hourStats, _ := m.reporter.GetStatistics(ctx, time.Hour)
	
	if minuteStats != nil && minuteStats.ErrorsByType[err.Type] > int64(threshold.MaxErrorsPerMinute) {
		return true
	}
	
	if hourStats != nil && hourStats.ErrorsByType[err.Type] > int64(threshold.MaxErrorsPerHour) {
		return true
	}
	
	return false
}

// Start 启动错误监控
func (m *ErrorMonitor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.running {
		return NewError(ErrorTypeSystem, "MONITOR_001", "Error monitor is already running").
			WithComponent("ErrorMonitor").
			Build()
	}
	
	m.running = true
	m.logger.Info("Error monitor started")
	
	// 启动监控goroutine
	go m.monitorLoop()
	
	return nil
}

// Stop 停止错误监控
func (m *ErrorMonitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return nil
	}
	
	m.cancel()
	m.running = false
	m.logger.Info("Error monitor stopped")
	
	return nil
}

// monitorLoop 监控循环
func (m *ErrorMonitor) monitorLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performPeriodicChecks()
		}
	}
}

// performPeriodicChecks 执行周期性检查
func (m *ErrorMonitor) performPeriodicChecks() {
	// 获取错误统计信息
	stats, err := m.reporter.GetStatistics(m.ctx, time.Hour)
	if err != nil {
		m.logger.Error("Failed to get error statistics", zap.Error(err))
		return
	}
	
	// 记录统计信息
	m.logger.Debug("Error statistics",
		zap.Int64("total_errors", stats.TotalErrors),
		zap.Any("errors_by_type", stats.ErrorsByType),
		zap.Any("errors_by_severity", stats.ErrorsBySeverity))
}

// GetStatistics 获取错误统计信息
func (m *ErrorMonitor) GetStatistics(ctx context.Context, timeWindow time.Duration) (*ErrorStatistics, error) {
	return m.reporter.GetStatistics(ctx, timeWindow)
}

// Close 关闭错误监控器
func (m *ErrorMonitor) Close() error {
	if err := m.Stop(); err != nil {
		return err
	}
	
	if err := m.reporter.Close(); err != nil {
		return err
	}
	
	return nil
}

// generateErrorID 生成错误ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}