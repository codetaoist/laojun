package exporters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"github.com/codetaoist/laojun-monitoring/internal/config"
)

// BatchProcessorExporter 批量处理器导出器
type BatchProcessorExporter struct {
	name           string
	config         *config.BatchProcessorConfig
	logger         *zap.Logger
	processor      *BatchProcessor
	running        bool
	healthy        bool
	ready          bool
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	stats          ExporterStats
	startTime      time.Time
}



// BatchProcessor 批量处理器
type BatchProcessor struct {
	config    *config.BatchProcessorConfig
	logger    *zap.Logger
	buffer    []interface{}
	exporters []Exporter
	mu        sync.RWMutex
	ticker    *time.Ticker
	ctx       context.Context
	cancel    context.CancelFunc
	stats     *ProcessorStats
}

// ProcessorStats 处理器统计信息
type ProcessorStats struct {
	ProcessedItems   int64     `json:"processed_items"`
	DroppedItems     int64     `json:"dropped_items"`
	ExportBatches    int64     `json:"export_batches"`
	FailedExports    int64     `json:"failed_exports"`
	LastProcessTime  time.Time `json:"last_process_time,omitempty"`
	AverageBatchSize float64   `json:"average_batch_size"`
	TotalLatency     int64     `json:"total_latency_ms"`
}

// NewBatchProcessorExporter 创建新的批量处理器导出器
func NewBatchProcessorExporter(config *config.BatchProcessorConfig, logger *zap.Logger) (*BatchProcessorExporter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	if logger == nil {
		logger = zap.NewNop()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	exporter := &BatchProcessorExporter{
		name:      config.Name,
		config:    config,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
		stats: ExporterStats{
			StartTime: time.Now(),
		},
	}
	
	// 创建批量处理器
	processor, err := NewBatchProcessor(config, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create batch processor: %w", err)
	}
	
	exporter.processor = processor
	exporter.ready = true
	
	return exporter, nil
}

// NewBatchProcessor 创建新的批量处理器
func NewBatchProcessor(config *config.BatchProcessorConfig, logger *zap.Logger) (*BatchProcessor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	processor := &BatchProcessor{
		config: config,
		logger: logger,
		buffer: make([]interface{}, 0, config.BufferSize),
		ctx:    ctx,
		cancel: cancel,
		stats: &ProcessorStats{
			AverageBatchSize: 0,
		},
	}
	
	// 初始化导出器
	if err := processor.initExporters(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init exporters: %w", err)
	}
	
	return processor, nil
}

// Name 返回导出器名称
func (e *BatchProcessorExporter) Name() string {
	return e.name
}

// Start 启动导出器
func (e *BatchProcessorExporter) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		return nil
	}
	
	if !e.config.IsEnabled() {
		e.logger.Info("Batch processor exporter is disabled", zap.String("name", e.name))
		return nil
	}
	
	e.logger.Info("Starting batch processor exporter", zap.String("name", e.name))
	
	// 启动批量处理器
	if err := e.processor.Start(); err != nil {
		return fmt.Errorf("failed to start batch processor: %w", err)
	}
	
	e.running = true
	e.healthy = true
	
	e.logger.Info("Batch processor exporter started successfully", zap.String("name", e.name))
	
	return nil
}

// Stop 停止导出器
func (e *BatchProcessorExporter) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.running {
		return nil
	}
	
	e.logger.Info("Stopping batch processor exporter", zap.String("name", e.name))
	
	// 停止批量处理器
	if e.processor != nil {
		if err := e.processor.Stop(); err != nil {
			e.logger.Error("Failed to stop batch processor", zap.Error(err))
		}
	}
	
	e.cancel()
	e.running = false
	e.healthy = false
	
	e.logger.Info("Batch processor exporter stopped", zap.String("name", e.name))
	
	return nil
}

// IsHealthy 检查导出器健康状态
func (e *BatchProcessorExporter) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.healthy && e.running
}

// IsReady 检查导出器就绪状态
func (e *BatchProcessorExporter) IsReady() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ready
}

// IsRunning 检查导出器运行状态
func (e *BatchProcessorExporter) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Export 导出数据
func (e *BatchProcessorExporter) Export(data interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("exporter is not healthy")
	}
	
	start := time.Now()
	
	// 添加到批量处理器
	if err := e.processor.Add(data); err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to add data to processor: %w", err)
	}
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// ExportBatch 批量导出数据
func (e *BatchProcessorExporter) ExportBatch(data []interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("exporter is not healthy")
	}
	
	start := time.Now()
	
	// 批量添加到处理器
	for _, item := range data {
		if err := e.processor.Add(item); err != nil {
			e.updateStats(false, time.Since(start), err)
			return fmt.Errorf("failed to add batch data to processor: %w", err)
		}
	}
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// GetBatchSize 获取批量大小
func (e *BatchProcessorExporter) GetBatchSize() int {
	return e.config.BatchSize
}

// SetBatchSize 设置批量大小
func (e *BatchProcessorExporter) SetBatchSize(size int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.BatchSize = size
}

// GetFlushInterval 获取刷新间隔
func (e *BatchProcessorExporter) GetFlushInterval() time.Duration {
	return e.config.FlushInterval
}

// SetFlushInterval 设置刷新间隔
func (e *BatchProcessorExporter) SetFlushInterval(interval time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.FlushInterval = interval
}

// GetStats 获取导出器统计信息
func (e *BatchProcessorExporter) GetStats() ExporterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

// updateStats 更新统计信息
func (e *BatchProcessorExporter) updateStats(success bool, latency time.Duration, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.stats.RequestCount++
	e.stats.LastRequestTime = time.Now()
	
	if success {
		e.stats.ProcessedItems++
	} else {
		e.stats.ErrorCount++
		if err != nil {
			e.stats.LastError = err.Error()
		}
	}
}

// Start 启动批量处理器
func (p *BatchProcessor) Start() error {
	p.logger.Info("Starting batch processor",
		zap.Int("batch_size", p.config.BatchSize),
		zap.Duration("flush_interval", p.config.FlushInterval))
	
	// 启动所有导出器
	for _, exporter := range p.exporters {
		if err := exporter.Start(p.ctx); err != nil {
			return fmt.Errorf("failed to start exporter %s: %w", exporter.Name(), err)
		}
	}
	
	// 启动定时刷新
	p.ticker = time.NewTicker(p.config.FlushInterval)
	go p.flushLoop()
	
	return nil
}

// Stop 停止批量处理器
func (p *BatchProcessor) Stop() error {
	p.logger.Info("Stopping batch processor")
	
	p.cancel()
	
	if p.ticker != nil {
		p.ticker.Stop()
	}
	
	// 最后一次刷新
	p.flush()
	
	// 停止所有导出器
	for _, exporter := range p.exporters {
		if err := exporter.Stop(); err != nil {
			p.logger.Error("Failed to stop exporter",
				zap.String("exporter", exporter.Name()),
				zap.Error(err))
		}
	}
	
	return nil
}

// Add 添加数据到缓冲区
func (p *BatchProcessor) Add(data interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.buffer) >= p.config.BufferSize {
		p.stats.DroppedItems++
		return fmt.Errorf("buffer is full, dropping item")
	}
	
	p.buffer = append(p.buffer, data)
	p.stats.ProcessedItems++
	
	// 如果达到批量大小，立即刷新
	if len(p.buffer) >= p.config.BatchSize {
		go p.flush()
	}
	
	return nil
}

// flush 刷新缓冲区
func (p *BatchProcessor) flush() {
	p.mu.Lock()
	if len(p.buffer) == 0 {
		p.mu.Unlock()
		return
	}
	
	batch := make([]interface{}, len(p.buffer))
	copy(batch, p.buffer)
	p.buffer = p.buffer[:0]
	p.mu.Unlock()
	
	start := time.Now()
	
	// 导出到所有导出器
	for _, exporter := range p.exporters {
		if batchExporter, ok := exporter.(BatchExporter); ok {
			if err := batchExporter.ExportBatch(batch); err != nil {
				p.logger.Error("Failed to export batch",
					zap.String("exporter", exporter.Name()),
					zap.Error(err))
				p.stats.FailedExports++
			}
		} else {
			// 逐个导出
			for _, item := range batch {
				if err := exporter.Export(item); err != nil {
					p.logger.Error("Failed to export item",
						zap.String("exporter", exporter.Name()),
						zap.Error(err))
					p.stats.FailedExports++
				}
			}
		}
	}
	
	p.stats.ExportBatches++
	p.stats.LastProcessTime = time.Now()
	p.stats.TotalLatency += time.Since(start).Milliseconds()
	
	// 更新平均批量大小
	if p.stats.ExportBatches > 0 {
		p.stats.AverageBatchSize = float64(p.stats.ProcessedItems) / float64(p.stats.ExportBatches)
	}
	
	p.logger.Debug("Batch exported",
		zap.Int("batch_size", len(batch)),
		zap.Duration("latency", time.Since(start)))
}

// flushLoop 定时刷新循环
func (p *BatchProcessor) flushLoop() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.ticker.C:
			p.flush()
		}
	}
}

// initExporters 初始化导出器
func (p *BatchProcessor) initExporters() error {
	p.exporters = make([]Exporter, 0, len(p.config.Exporters))
	
	for _, exporterConfig := range p.config.Exporters {
		if !exporterConfig.Enabled {
			continue
		}
		
		var exporter Exporter
		var err error
		
		switch exporterConfig.Type {
		case "console":
			exporter, err = NewConsoleExporter(&exporterConfig, p.logger)
		case "file":
			exporter, err = NewFileExporter(&exporterConfig, p.logger)
		case "http":
			exporter, err = NewHTTPExporter(&exporterConfig, p.logger)
		default:
			return fmt.Errorf("unsupported exporter type: %s", exporterConfig.Type)
		}
		
		if err != nil {
			return fmt.Errorf("failed to create exporter %s: %w", exporterConfig.Name, err)
		}
		
		p.exporters = append(p.exporters, exporter)
	}
	
	return nil
}

// GetProcessorStats 获取处理器统计信息
func (p *BatchProcessor) GetProcessorStats() *ProcessorStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	stats := *p.stats
	return &stats
}