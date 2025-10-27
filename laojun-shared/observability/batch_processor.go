package observability

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// BatchItem 表示批量处理的数据项
type BatchItem struct {
	Type      string                 `json:"type"`      // metric, event, trace
	Name      string                 `json:"name"`      // 名称
	Value     interface{}            `json:"value"`     // 值
	Labels    map[string]string      `json:"labels"`    // 标签
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// BatchProcessor 批量处理器
type BatchProcessor struct {
	config       *Config
	exportConfig *ExportConfig
	
	// 缓冲区
	buffer    []BatchItem
	bufferMu  sync.RWMutex
	
	// 统计信息
	stats BatchStats
	
	// 控制通道
	flushCh   chan struct{}
	stopCh    chan struct{}
	doneCh    chan struct{}
	
	// 导出器
	exporters map[string]Exporter
	
	// 状态
	running int32
}

// BatchStats 批量处理统计信息
type BatchStats struct {
	TotalItems       int64 `json:"total_items"`
	ProcessedItems   int64 `json:"processed_items"`
	DroppedItems     int64 `json:"dropped_items"`
	ExportedBatches  int64 `json:"exported_batches"`
	FailedExports    int64 `json:"failed_exports"`
	BufferSize       int64 `json:"buffer_size"`
	LastFlushTime    time.Time `json:"last_flush_time"`
	LastExportTime   time.Time `json:"last_export_time"`
	AverageFlushTime time.Duration `json:"average_flush_time"`
}

// Exporter 导出器接口
type Exporter interface {
	Export(ctx context.Context, items []BatchItem) error
	Name() string
	Close() error
}

// NewBatchProcessor 创建新的批量处理器
func NewBatchProcessor(config *Config) *BatchProcessor {
	if config.Export == nil {
		config.Export = &ExportConfig{}
		config.Export.ApplyDefaults()
	}

	bp := &BatchProcessor{
		config:       config,
		exportConfig: config.Export,
		buffer:       make([]BatchItem, 0, config.BufferSize),
		flushCh:      make(chan struct{}, 1),
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
		exporters:    make(map[string]Exporter),
	}

	// 初始化导出器
	bp.initializeExporters()

	return bp
}

// initializeExporters 初始化导出器
func (bp *BatchProcessor) initializeExporters() {
	factory := &ExporterFactory{}
	
	for name, endpoint := range bp.exportConfig.Endpoints {
		// 根据endpoint确定导出器类型
		exporterType := "http" // 默认类型
		if len(endpoint) > 0 {
			if len(endpoint) >= 7 && endpoint[:7] == "http://" {
				exporterType = "http"
			} else if len(endpoint) >= 8 && endpoint[:8] == "https://" {
				exporterType = "https"
			} else if len(endpoint) >= 10 && endpoint[:10] == "console://" {
				exporterType = "console"
			} else if len(endpoint) >= 7 && endpoint[:7] == "file://" {
				exporterType = "file"
				endpoint = endpoint[7:] // 移除file://前缀
			}
		}
		
		exporter, err := factory.CreateExporter(exporterType, name, endpoint, bp.exportConfig)
		if err != nil {
			fmt.Printf("Failed to create exporter %s: %v\n", name, err)
			continue
		}
		
		bp.exporters[name] = exporter
	}
}

// Start 启动批量处理器
func (bp *BatchProcessor) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&bp.running, 0, 1) {
		return fmt.Errorf("batch processor is already running")
	}

	go bp.processLoop(ctx)
	return nil
}

// Stop 停止批量处理器
func (bp *BatchProcessor) Stop() error {
	if !atomic.CompareAndSwapInt32(&bp.running, 1, 0) {
		return fmt.Errorf("batch processor is not running")
	}

	close(bp.stopCh)
	<-bp.doneCh

	// 关闭所有导出器
	for _, exporter := range bp.exporters {
		if err := exporter.Close(); err != nil {
			// 记录错误但继续关闭其他导出器
			fmt.Printf("Error closing exporter %s: %v\n", exporter.Name(), err)
		}
	}

	return nil
}

// AddItem 添加项目到缓冲区
func (bp *BatchProcessor) AddItem(item BatchItem) error {
	if atomic.LoadInt32(&bp.running) == 0 {
		return fmt.Errorf("batch processor is not running")
	}

	bp.bufferMu.Lock()
	defer bp.bufferMu.Unlock()

	// 检查缓冲区是否已满
	if len(bp.buffer) >= bp.config.BufferSize {
		if bp.exportConfig.DropOnFailure {
			atomic.AddInt64(&bp.stats.DroppedItems, 1)
			return fmt.Errorf("buffer is full, item dropped")
		} else {
			// 强制刷新缓冲区
			bp.triggerFlush()
		}
	}

	// 添加时间戳
	if item.Timestamp.IsZero() {
		item.Timestamp = time.Now()
	}

	bp.buffer = append(bp.buffer, item)
	atomic.AddInt64(&bp.stats.TotalItems, 1)
	atomic.StoreInt64(&bp.stats.BufferSize, int64(len(bp.buffer)))

	// 检查是否需要触发刷新
	if len(bp.buffer) >= bp.exportConfig.BatchSize {
		bp.triggerFlush()
	}

	return nil
}

// AddMetric 添加指标
func (bp *BatchProcessor) AddMetric(name string, value float64, labels map[string]string) error {
	item := BatchItem{
		Type:   "metric",
		Name:   name,
		Value:  value,
		Labels: labels,
	}
	return bp.AddItem(item)
}

// AddEvent 添加事件
func (bp *BatchProcessor) AddEvent(name string, attributes map[string]interface{}) error {
	item := BatchItem{
		Type:     "event",
		Name:     name,
		Metadata: attributes,
	}
	return bp.AddItem(item)
}

// AddTrace 添加跟踪数据
func (bp *BatchProcessor) AddTrace(name string, metadata map[string]interface{}) error {
	item := BatchItem{
		Type:     "trace",
		Name:     name,
		Metadata: metadata,
	}
	return bp.AddItem(item)
}

// triggerFlush 触发刷新
func (bp *BatchProcessor) triggerFlush() {
	select {
	case bp.flushCh <- struct{}{}:
	default:
		// 通道已满，刷新已在进行中
	}
}

// processLoop 处理循环
func (bp *BatchProcessor) processLoop(ctx context.Context) {
	defer close(bp.doneCh)

	flushTicker := time.NewTicker(bp.config.FlushPeriod)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			bp.flushBuffer(ctx)
			return
		case <-bp.stopCh:
			bp.flushBuffer(context.Background())
			return
		case <-bp.flushCh:
			bp.flushBuffer(ctx)
		case <-flushTicker.C:
			bp.flushBuffer(ctx)
		}
	}
}

// flushBuffer 刷新缓冲区
func (bp *BatchProcessor) flushBuffer(ctx context.Context) {
	startTime := time.Now()
	
	bp.bufferMu.Lock()
	if len(bp.buffer) == 0 {
		bp.bufferMu.Unlock()
		return
	}

	// 复制缓冲区数据
	items := make([]BatchItem, len(bp.buffer))
	copy(items, bp.buffer)
	
	// 清空缓冲区
	bp.buffer = bp.buffer[:0]
	atomic.StoreInt64(&bp.stats.BufferSize, 0)
	bp.bufferMu.Unlock()

	// 导出数据
	bp.exportItems(ctx, items)

	// 更新统计信息
	flushDuration := time.Since(startTime)
	bp.stats.LastFlushTime = time.Now()
	
	// 计算平均刷新时间
	if bp.stats.AverageFlushTime == 0 {
		bp.stats.AverageFlushTime = flushDuration
	} else {
		bp.stats.AverageFlushTime = (bp.stats.AverageFlushTime + flushDuration) / 2
	}
}

// exportItems 导出项目
func (bp *BatchProcessor) exportItems(ctx context.Context, items []BatchItem) {
	if len(items) == 0 {
		return
	}

	// 创建带超时的上下文
	exportCtx, cancel := context.WithTimeout(ctx, bp.exportConfig.Timeout)
	defer cancel()

	// 并发导出到所有配置的端点
	var wg sync.WaitGroup
	for _, exporter := range bp.exporters {
		wg.Add(1)
		go func(exp Exporter) {
			defer wg.Done()
			bp.exportToExporter(exportCtx, exp, items)
		}(exporter)
	}

	wg.Wait()
	atomic.AddInt64(&bp.stats.ExportedBatches, 1)
	bp.stats.LastExportTime = time.Now()
}

// exportToExporter 导出到特定导出器
func (bp *BatchProcessor) exportToExporter(ctx context.Context, exporter Exporter, items []BatchItem) {
	var lastErr error
	
	for retry := 0; retry <= bp.exportConfig.MaxRetries; retry++ {
		if retry > 0 {
			// 等待重试延迟
			select {
			case <-ctx.Done():
				return
			case <-time.After(bp.exportConfig.RetryDelay):
			}
		}

		err := exporter.Export(ctx, items)
		if err == nil {
			atomic.AddInt64(&bp.stats.ProcessedItems, int64(len(items)))
			return
		}

		lastErr = err
		fmt.Printf("Export to %s failed (attempt %d/%d): %v\n", 
			exporter.Name(), retry+1, bp.exportConfig.MaxRetries+1, err)
	}

	// 所有重试都失败了
	atomic.AddInt64(&bp.stats.FailedExports, 1)
	if bp.exportConfig.DropOnFailure {
		atomic.AddInt64(&bp.stats.DroppedItems, int64(len(items)))
	}
	
	fmt.Printf("Failed to export to %s after %d retries: %v\n", 
		exporter.Name(), bp.exportConfig.MaxRetries+1, lastErr)
}

// GetStats 获取统计信息
func (bp *BatchProcessor) GetStats() BatchStats {
	bp.bufferMu.RLock()
	defer bp.bufferMu.RUnlock()
	
	stats := bp.stats
	stats.BufferSize = int64(len(bp.buffer))
	return stats
}

// IsRunning 检查是否正在运行
func (bp *BatchProcessor) IsRunning() bool {
	return atomic.LoadInt32(&bp.running) == 1
}

// ForceFlush 强制刷新缓冲区
func (bp *BatchProcessor) ForceFlush() {
	if bp.IsRunning() {
		bp.triggerFlush()
	}
}

// GetBufferSize 获取当前缓冲区大小
func (bp *BatchProcessor) GetBufferSize() int {
	bp.bufferMu.RLock()
	defer bp.bufferMu.RUnlock()
	return len(bp.buffer)
}

// GetExporters 获取导出器列表
func (bp *BatchProcessor) GetExporters() []string {
	names := make([]string, 0, len(bp.exporters))
	for name := range bp.exporters {
		names = append(names, name)
	}
	return names
}