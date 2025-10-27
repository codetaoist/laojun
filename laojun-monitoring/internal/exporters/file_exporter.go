package exporters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FileExporter 文件导出器
type FileExporter struct {
	name      string
	config    ExporterConfig
	logger    *zap.Logger
	running   bool
	healthy   bool
	ready     bool
	mu        sync.RWMutex
	stats     ExporterStats
	startTime time.Time
	file      *os.File
	filePath  string
}

// NewFileExporter 创建文件导出器
func NewFileExporter(config ExporterConfig, logger *zap.Logger) (*FileExporter, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	// 获取文件路径配置
	filePath, ok := config.GetConfig()["file_path"].(string)
	if !ok || filePath == "" {
		filePath = fmt.Sprintf("./logs/%s_%s.log", 
			config.GetName(), 
			time.Now().Format("20060102_150405"))
	}
	
	exporter := &FileExporter{
		name:      config.GetName(),
		config:    config,
		logger:    logger,
		filePath:  filePath,
		ready:     true,
		startTime: time.Now(),
		stats: ExporterStats{
			StartTime: time.Now(),
		},
	}
	
	return exporter, nil
}

// Name 返回导出器名称
func (e *FileExporter) Name() string {
	return e.name
}

// Start 启动导出器
func (e *FileExporter) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		return nil
	}
	
	if !e.config.IsEnabled() {
		e.logger.Info("File exporter is disabled", zap.String("name", e.name))
		return nil
	}
	
	e.logger.Info("Starting file exporter", 
		zap.String("name", e.name),
		zap.String("file_path", e.filePath))
	
	// 创建目录
	dir := filepath.Dir(e.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// 打开文件
	file, err := os.OpenFile(e.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", e.filePath, err)
	}
	
	e.file = file
	e.running = true
	e.healthy = true
	
	e.logger.Info("File exporter started successfully", 
		zap.String("name", e.name),
		zap.String("file_path", e.filePath))
	
	return nil
}

// Stop 停止导出器
func (e *FileExporter) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.running {
		return nil
	}
	
	e.logger.Info("Stopping file exporter", zap.String("name", e.name))
	
	if e.file != nil {
		if err := e.file.Close(); err != nil {
			e.logger.Error("Failed to close file", zap.Error(err))
		}
		e.file = nil
	}
	
	e.running = false
	e.healthy = false
	
	e.logger.Info("File exporter stopped", zap.String("name", e.name))
	
	return nil
}

// IsHealthy 检查导出器健康状态
func (e *FileExporter) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.healthy && e.running && e.file != nil
}

// IsReady 检查导出器就绪状态
func (e *FileExporter) IsReady() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ready
}

// Export 导出数据
func (e *FileExporter) Export(data interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("file exporter is not healthy")
	}
	
	start := time.Now()
	
	// 格式化数据
	output, err := e.formatData(data)
	if err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to format data: %w", err)
	}
	
	// 写入文件
	e.mu.Lock()
	if e.file != nil {
		_, err = fmt.Fprintf(e.file, "[%s] %s\n", 
			time.Now().Format("2006-01-02 15:04:05.000"), 
			output)
		if err == nil {
			e.file.Sync() // 强制刷新到磁盘
		}
	}
	e.mu.Unlock()
	
	if err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to write to file: %w", err)
	}
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// ExportBatch 批量导出数据
func (e *FileExporter) ExportBatch(data []interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("file exporter is not healthy")
	}
	
	start := time.Now()
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.file == nil {
		err := fmt.Errorf("file is not open")
		e.updateStats(false, time.Since(start), err)
		return err
	}
	
	// 写入批量开始标记
	fmt.Fprintf(e.file, "[%s] %s: Start Export [file] - Batch size: %d\n",
		time.Now().Format("2006-01-02 15:04:05.000"),
		e.name,
		len(data))
	
	// 写入每个数据项
	for i, item := range data {
		output, err := e.formatData(item)
		if err != nil {
			e.logger.Error("Failed to format data item",
				zap.Int("index", i),
				zap.Error(err))
			continue
		}
		
		fmt.Fprintf(e.file, "  [%d] %s\n", i+1, output)
	}
	
	// 写入批量结束标记
	fmt.Fprintf(e.file, "[%s] %s: End Export [file] - Exported %d items in %v\n",
		time.Now().Format("2006-01-02 15:04:05.000"),
		e.name,
		len(data),
		time.Since(start))
	
	// 强制刷新到磁盘
	if err := e.file.Sync(); err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to sync file: %w", err)
	}
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// GetBatchSize 获取批量大小
func (e *FileExporter) GetBatchSize() int {
	config := e.config.GetConfig()
	if batchSize, ok := config["batch_size"].(int); ok {
		return batchSize
	}
	return 100 // 默认批量大小
}

// SetBatchSize 设置批量大小
func (e *FileExporter) SetBatchSize(size int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.SetConfig("batch_size", size)
}

// GetFlushInterval 获取刷新间隔
func (e *FileExporter) GetFlushInterval() time.Duration {
	config := e.config.GetConfig()
	if interval, ok := config["flush_interval"].(time.Duration); ok {
		return interval
	}
	return 5 * time.Second // 默认刷新间隔
}

// SetFlushInterval 设置刷新间隔
func (e *FileExporter) SetFlushInterval(interval time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.SetConfig("flush_interval", interval)
}

// GetStats 获取导出器统计信息
func (e *FileExporter) GetStats() ExporterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

// formatData 格式化数据
func (e *FileExporter) formatData(data interface{}) (string, error) {
	switch v := data.(type) {
	case *MetricData:
		jsonData, err := json.Marshal(map[string]interface{}{
			"type":      "metric",
			"name":      v.Name,
			"value":     v.Value,
			"labels":    v.Labels,
			"timestamp": v.Timestamp.Format(time.RFC3339Nano),
			"metadata":  v.Metadata,
		})
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
		
	case *EventData:
		jsonData, err := json.Marshal(map[string]interface{}{
			"type":      "event",
			"name":      v.Name,
			"message":   v.Message,
			"level":     v.Level,
			"labels":    v.Labels,
			"timestamp": v.Timestamp.Format(time.RFC3339Nano),
			"metadata":  v.Metadata,
		})
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
		
	case *TraceData:
		jsonData, err := json.Marshal(map[string]interface{}{
			"type":      "trace",
			"trace_id":  v.TraceID,
			"span_id":   v.SpanID,
			"operation": v.Operation,
			"duration":  v.Duration.Nanoseconds(),
			"labels":    v.Labels,
			"timestamp": v.Timestamp.Format(time.RFC3339Nano),
			"metadata":  v.Metadata,
		})
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
		
	default:
		// 通用 JSON 序列化
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Sprintf("%+v", data), nil
		}
		return string(jsonData), nil
	}
}

// updateStats 更新统计信息
func (e *FileExporter) updateStats(success bool, latency time.Duration, err error) {
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