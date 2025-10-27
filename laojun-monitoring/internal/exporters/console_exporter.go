package exporters

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConsoleExporter 控制台导出器
type ConsoleExporter struct {
	name      string
	config    ExporterConfig
	logger    *zap.Logger
	running   bool
	healthy   bool
	ready     bool
	mu        sync.RWMutex
	stats     ExporterStats
	startTime time.Time
}

// NewConsoleExporter 创建控制台导出器
func NewConsoleExporter(config ExporterConfig, logger *zap.Logger) (*ConsoleExporter, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	exporter := &ConsoleExporter{
		name:      config.GetName(),
		config:    config,
		logger:    logger,
		ready:     true,
		startTime: time.Now(),
		stats: ExporterStats{
			StartTime: time.Now(),
		},
	}
	
	return exporter, nil
}

// Name 返回导出器名称
func (e *ConsoleExporter) Name() string {
	return e.name
}

// Start 启动导出器
func (e *ConsoleExporter) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		return nil
	}
	
	if !e.config.IsEnabled() {
		e.logger.Info("Console exporter is disabled", zap.String("name", e.name))
		return nil
	}
	
	e.logger.Info("Starting console exporter", zap.String("name", e.name))
	
	e.running = true
	e.healthy = true
	
	e.logger.Info("Console exporter started successfully", zap.String("name", e.name))
	
	return nil
}

// Stop 停止导出器
func (e *ConsoleExporter) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.running {
		return nil
	}
	
	e.logger.Info("Stopping console exporter", zap.String("name", e.name))
	
	e.running = false
	e.healthy = false
	
	e.logger.Info("Console exporter stopped", zap.String("name", e.name))
	
	return nil
}

// IsHealthy 检查导出器健康状态
func (e *ConsoleExporter) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.healthy && e.running
}

// IsReady 检查导出器就绪状态
func (e *ConsoleExporter) IsReady() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ready
}

// Export 导出数据
func (e *ConsoleExporter) Export(data interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("console exporter is not healthy")
	}
	
	start := time.Now()
	
	// 格式化输出
	output, err := e.formatData(data)
	if err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to format data: %w", err)
	}
	
	// 输出到控制台
	fmt.Printf("[%s] %s: %s\n", 
		time.Now().Format("2006-01-02 15:04:05.000"), 
		e.name, 
		output)
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// ExportBatch 批量导出数据
func (e *ConsoleExporter) ExportBatch(data []interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("console exporter is not healthy")
	}
	
	start := time.Now()
	
	fmt.Printf("[%s] %s: Start Export [console]\n", 
		time.Now().Format("2006-01-02 15:04:05.000"), 
		e.name)
	
	for i, item := range data {
		output, err := e.formatData(item)
		if err != nil {
			e.logger.Error("Failed to format data item",
				zap.Int("index", i),
				zap.Error(err))
			continue
		}
		
		fmt.Printf("  [%d] %s\n", i+1, output)
	}
	
	fmt.Printf("[%s] %s: End Export [console] - Exported %d items in %v\n",
		time.Now().Format("2006-01-02 15:04:05.000"),
		e.name,
		len(data),
		time.Since(start))
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// GetBatchSize 获取批量大小
func (e *ConsoleExporter) GetBatchSize() int {
	return 100 // 默认批量大小
}

// SetBatchSize 设置批量大小
func (e *ConsoleExporter) SetBatchSize(size int) {
	// Console exporter 不需要特殊的批量大小设置
}

// GetFlushInterval 获取刷新间隔
func (e *ConsoleExporter) GetFlushInterval() time.Duration {
	return time.Second // 默认刷新间隔
}

// SetFlushInterval 设置刷新间隔
func (e *ConsoleExporter) SetFlushInterval(interval time.Duration) {
	// Console exporter 不需要特殊的刷新间隔设置
}

// GetStats 获取导出器统计信息
func (e *ConsoleExporter) GetStats() ExporterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

// formatData 格式化数据
func (e *ConsoleExporter) formatData(data interface{}) (string, error) {
	switch v := data.(type) {
	case *MetricData:
		return fmt.Sprintf("Metric{name=%s, value=%.2f, labels=%v, timestamp=%s}",
			v.Name, v.Value, v.Labels, v.Timestamp.Format("15:04:05.000")), nil
	case *EventData:
		return fmt.Sprintf("Event{name=%s, message=%s, level=%s, labels=%v, timestamp=%s}",
			v.Name, v.Message, v.Level, v.Labels, v.Timestamp.Format("15:04:05.000")), nil
	case *TraceData:
		return fmt.Sprintf("Trace{trace_id=%s, span_id=%s, operation=%s, duration=%v, labels=%v, timestamp=%s}",
			v.TraceID, v.SpanID, v.Operation, v.Duration, v.Labels, v.Timestamp.Format("15:04:05.000")), nil
	default:
		// 尝试 JSON 序列化
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Sprintf("%+v", data), nil
		}
		return string(jsonData), nil
	}
}

// updateStats 更新统计信息
func (e *ConsoleExporter) updateStats(success bool, latency time.Duration, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.stats.RequestCount++
	e.stats.LastRequestTime = time.Now()
	
	if !success {
		e.stats.ErrorCount++
		if err != nil {
			e.stats.LastError = err.Error()
		}
	}
}