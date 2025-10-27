package observability

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// OptimizedBatchProcessor 优化的批量处理器
type OptimizedBatchProcessor struct {
	config       *Config
	exportConfig *ExportConfig
	
	// 缓冲区和对象池
	buffer       []BatchItem
	bufferMu     sync.RWMutex
	itemPool     *sync.Pool
	bufferPool   *sync.Pool
	
	// 指标缓存
	metricsCache *MetricsCache
	
	// 统计信息
	stats OptimizedBatchStats
	
	// 控制通道
	flushCh   chan struct{}
	stopCh    chan struct{}
	doneCh    chan struct{}
	
	// 导出器
	exporters map[string]OptimizedExporter
	
	// 状态
	running int32
	
	// 性能监控
	perfMonitor *PerformanceMonitor
}

// OptimizedBatchStats 优化的批量处理统计信息
type OptimizedBatchStats struct {
	TotalItems         int64         `json:"total_items"`
	ProcessedItems     int64         `json:"processed_items"`
	DroppedItems       int64         `json:"dropped_items"`
	ExportedBatches    int64         `json:"exported_batches"`
	FailedExports      int64         `json:"failed_exports"`
	BufferSize         int64         `json:"buffer_size"`
	LastFlushTime      time.Time     `json:"last_flush_time"`
	LastExportTime     time.Time     `json:"last_export_time"`
	AverageFlushTime   time.Duration `json:"average_flush_time"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
	MemoryUsage        int64         `json:"memory_usage_bytes"`
	GCCount            int64         `json:"gc_count"`
	ConnectionPoolSize int           `json:"connection_pool_size"`
}

// MetricsCache 指标缓存
type MetricsCache struct {
	cache    map[string]*CachedMetric
	mu       sync.RWMutex
	ttl      time.Duration
	maxSize  int
	hits     int64
	misses   int64
}

// CachedMetric 缓存的指标
type CachedMetric struct {
	Data      []byte
	Hash      uint64
	Timestamp time.Time
	AccessCount int64
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	mu              sync.RWMutex
	startTime       time.Time
	lastGCTime      time.Time
	gcCount         int64
	memoryUsage     int64
	cpuUsage        float64
	goroutineCount  int
}

// OptimizedExporter 优化的导出器接口
type OptimizedExporter interface {
	Export(ctx context.Context, items []BatchItem) error
	ExportCached(ctx context.Context, cachedData []byte) error
	Name() string
	Close() error
	GetConnectionPoolStats() ConnectionPoolStats
}

// ConnectionPoolStats 连接池统计信息
type ConnectionPoolStats struct {
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	TotalConnections  int `json:"total_connections"`
	MaxConnections    int `json:"max_connections"`
}

// NewOptimizedBatchProcessor 创建新的优化批量处理器
func NewOptimizedBatchProcessor(config *Config) *OptimizedBatchProcessor {
	if config.Export == nil {
		config.Export = &ExportConfig{}
		config.Export.ApplyDefaults()
	}

	bp := &OptimizedBatchProcessor{
		config:       config,
		exportConfig: config.Export,
		buffer:       make([]BatchItem, 0, config.BufferSize),
		flushCh:      make(chan struct{}, 1),
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
		exporters:    make(map[string]OptimizedExporter),
		perfMonitor:  NewPerformanceMonitor(),
	}

	// 初始化对象池
	bp.initObjectPools()
	
	// 初始化指标缓存
	bp.initMetricsCache()
	
	// 初始化导出器
	bp.initializeOptimizedExporters()

	return bp
}

// initObjectPools 初始化对象池
func (bp *OptimizedBatchProcessor) initObjectPools() {
	// BatchItem对象池
	bp.itemPool = &sync.Pool{
		New: func() interface{} {
			return &BatchItem{
				Labels:   make(map[string]string),
				Metadata: make(map[string]interface{}),
			}
		},
	}
	
	// 缓冲区对象池
	bp.bufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]BatchItem, 0, bp.exportConfig.BatchSize)
		},
	}
}

// initMetricsCache 初始化指标缓存
func (bp *OptimizedBatchProcessor) initMetricsCache() {
	bp.metricsCache = &MetricsCache{
		cache:   make(map[string]*CachedMetric),
		ttl:     5 * time.Minute, // 缓存5分钟
		maxSize: 1000,            // 最大缓存1000个指标
	}
	
	// 启动缓存清理goroutine
	go bp.cacheCleanupLoop()
}

// initializeOptimizedExporters 初始化优化的导出器
func (bp *OptimizedBatchProcessor) initializeOptimizedExporters() {
	factory := &OptimizedExporterFactory{}
	
	for name, endpoint := range bp.exportConfig.Endpoints {
		exporterType := bp.determineExporterType(endpoint)
		
		exporter, err := factory.CreateOptimizedExporter(exporterType, name, endpoint, bp.exportConfig)
		if err != nil {
			fmt.Printf("Failed to create optimized exporter %s: %v\n", name, err)
			continue
		}
		
		bp.exporters[name] = exporter
	}
}

// determineExporterType 确定导出器类型
func (bp *OptimizedBatchProcessor) determineExporterType(endpoint string) string {
	if len(endpoint) >= 7 && endpoint[:7] == "http://" {
		return "http"
	} else if len(endpoint) >= 8 && endpoint[:8] == "https://" {
		return "https"
	} else if len(endpoint) >= 10 && endpoint[:10] == "console://" {
		return "console"
	} else if len(endpoint) >= 7 && endpoint[:7] == "file://" {
		return "file"
	}
	return "http" // 默认类型
}

// Start 启动优化的批量处理器
func (bp *OptimizedBatchProcessor) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&bp.running, 0, 1) {
		return fmt.Errorf("optimized batch processor is already running")
	}

	bp.perfMonitor.Start()
	go bp.processLoop(ctx)
	go bp.performanceMonitorLoop()
	
	return nil
}

// Stop 停止优化的批量处理器
func (bp *OptimizedBatchProcessor) Stop() error {
	if !atomic.CompareAndSwapInt32(&bp.running, 1, 0) {
		return fmt.Errorf("optimized batch processor is not running")
	}

	close(bp.stopCh)
	<-bp.doneCh

	// 关闭所有导出器
	for _, exporter := range bp.exporters {
		if err := exporter.Close(); err != nil {
			fmt.Printf("Error closing optimized exporter %s: %v\n", exporter.Name(), err)
		}
	}

	bp.perfMonitor.Stop()
	return nil
}

// AddItemOptimized 优化的添加项目方法
func (bp *OptimizedBatchProcessor) AddItemOptimized(item BatchItem) error {
	if atomic.LoadInt32(&bp.running) == 0 {
		return fmt.Errorf("optimized batch processor is not running")
	}

	// 从对象池获取item（如果需要复制）
	pooledItem := bp.itemPool.Get().(*BatchItem)
	defer bp.itemPool.Put(pooledItem)

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

// triggerFlush 触发刷新
func (bp *OptimizedBatchProcessor) triggerFlush() {
	select {
	case bp.flushCh <- struct{}{}:
	default:
		// 通道已满，刷新已在进行中
	}
}

// processLoop 处理循环
func (bp *OptimizedBatchProcessor) processLoop(ctx context.Context) {
	defer close(bp.doneCh)

	flushTicker := time.NewTicker(bp.config.FlushPeriod)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			bp.flushBufferOptimized(ctx)
			return
		case <-bp.stopCh:
			bp.flushBufferOptimized(context.Background())
			return
		case <-bp.flushCh:
			bp.flushBufferOptimized(ctx)
		case <-flushTicker.C:
			bp.flushBufferOptimized(ctx)
		}
	}
}

// flushBufferOptimized 优化的刷新缓冲区
func (bp *OptimizedBatchProcessor) flushBufferOptimized(ctx context.Context) {
	startTime := time.Now()
	
	bp.bufferMu.Lock()
	if len(bp.buffer) == 0 {
		bp.bufferMu.Unlock()
		return
	}

	// 从对象池获取缓冲区
	items := bp.bufferPool.Get().([]BatchItem)
	items = items[:0] // 重置长度但保留容量
	
	// 复制缓冲区数据
	items = append(items, bp.buffer...)
	
	// 清空缓冲区
	bp.buffer = bp.buffer[:0]
	atomic.StoreInt64(&bp.stats.BufferSize, 0)
	bp.bufferMu.Unlock()

	// 导出数据
	bp.exportItemsOptimized(ctx, items)
	
	// 将缓冲区返回对象池
	bp.bufferPool.Put(items)

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

// exportItemsOptimized 优化的导出项目
func (bp *OptimizedBatchProcessor) exportItemsOptimized(ctx context.Context, items []BatchItem) {
	if len(items) == 0 {
		return
	}

	// 创建带超时的上下文
	exportCtx, cancel := context.WithTimeout(ctx, bp.exportConfig.Timeout)
	defer cancel()

	// 尝试从缓存获取序列化数据
	cacheKey := bp.generateCacheKey(items)
	cachedData := bp.getCachedData(cacheKey)
	
	// 并发导出到所有配置的端点
	var wg sync.WaitGroup
	for _, exporter := range bp.exporters {
		wg.Add(1)
		go func(exp OptimizedExporter) {
			defer wg.Done()
			
			if cachedData != nil {
				// 使用缓存数据
				bp.exportCachedToExporter(exportCtx, exp, cachedData)
			} else {
				// 正常导出并缓存结果
				bp.exportToOptimizedExporter(exportCtx, exp, items, cacheKey)
			}
		}(exporter)
	}

	wg.Wait()
	atomic.AddInt64(&bp.stats.ExportedBatches, 1)
	bp.stats.LastExportTime = time.Now()
}

// cacheCleanupLoop 缓存清理循环
func (bp *OptimizedBatchProcessor) cacheCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			bp.cleanupExpiredCache()
		case <-bp.stopCh:
			return
		}
	}
}

// performanceMonitorLoop 性能监控循环
func (bp *OptimizedBatchProcessor) performanceMonitorLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			bp.updatePerformanceStats()
		case <-bp.stopCh:
			return
		}
	}
}

// GetOptimizedStats 获取优化的统计信息
func (bp *OptimizedBatchProcessor) GetOptimizedStats() OptimizedBatchStats {
	bp.bufferMu.RLock()
	defer bp.bufferMu.RUnlock()
	
	stats := bp.stats
	stats.BufferSize = int64(len(bp.buffer))
	
	// 更新缓存命中率
	bp.metricsCache.mu.RLock()
	total := bp.metricsCache.hits + bp.metricsCache.misses
	if total > 0 {
		stats.CacheHitRate = float64(bp.metricsCache.hits) / float64(total)
	}
	bp.metricsCache.mu.RUnlock()
	
	// 更新性能统计
	bp.perfMonitor.mu.RLock()
	stats.MemoryUsage = bp.perfMonitor.memoryUsage
	stats.GCCount = bp.perfMonitor.gcCount
	bp.perfMonitor.mu.RUnlock()
	
	return stats
}

// 其他辅助方法将在后续实现...