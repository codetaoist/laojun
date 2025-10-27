package collectors

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ApplicationCollector 应用程序指标收集器
type ApplicationCollector struct {
	mu       sync.RWMutex
	running  bool
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *logrus.Logger

	// Prometheus metrics
	goroutineCount    prometheus.Gauge
	heapSize          prometheus.Gauge
	heapUsed          prometheus.Gauge
	stackSize         prometheus.Gauge
	gcCount           prometheus.Counter
	gcDuration        prometheus.Histogram
	allocRate         prometheus.Gauge
	
	// Statistics
	stats CollectorStats
	
	// Runtime stats for calculation
	lastGCCount    uint32
	lastGCTime     time.Time
	lastAllocBytes uint64
	lastSampleTime time.Time
}

// NewApplicationCollector 创建新的应用程序收集器
func NewApplicationCollector(interval time.Duration, logger *logrus.Logger) *ApplicationCollector {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ApplicationCollector{
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		goroutineCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "application_goroutine_count",
			Help: "Number of goroutines",
		}),
		heapSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "application_heap_size_bytes",
			Help: "Heap size in bytes",
		}),
		heapUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "application_heap_used_bytes",
			Help: "Heap used in bytes",
		}),
		stackSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "application_stack_size_bytes",
			Help: "Stack size in bytes",
		}),
		gcCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "application_gc_count_total",
			Help: "Total number of garbage collections",
		}),
		gcDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "application_gc_duration_seconds",
			Help:    "Garbage collection duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		allocRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "application_alloc_rate_bytes_per_second",
			Help: "Memory allocation rate in bytes per second",
		}),
		lastSampleTime: time.Now(),
	}
}

// Start 启动收集器
func (ac *ApplicationCollector) Start() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if ac.running {
		return nil
	}
	
	ac.running = true
	ac.logger.Info("Starting application collector")
	
	// 注册 Prometheus 指标
	prometheus.MustRegister(ac.goroutineCount)
	prometheus.MustRegister(ac.heapSize)
	prometheus.MustRegister(ac.heapUsed)
	prometheus.MustRegister(ac.stackSize)
	prometheus.MustRegister(ac.gcCount)
	prometheus.MustRegister(ac.gcDuration)
	prometheus.MustRegister(ac.allocRate)
	
	// 初始化基线数据
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ac.lastGCCount = m.NumGC
	ac.lastAllocBytes = m.TotalAlloc
	ac.lastSampleTime = time.Now()
	
	go ac.collectLoop()
	
	return nil
}

// Stop 停止收集器
func (ac *ApplicationCollector) Stop() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if !ac.running {
		return nil
	}
	
	ac.running = false
	ac.cancel()
	ac.logger.Info("Stopping application collector")
	
	// 注销 Prometheus 指标
	prometheus.Unregister(ac.goroutineCount)
	prometheus.Unregister(ac.heapSize)
	prometheus.Unregister(ac.heapUsed)
	prometheus.Unregister(ac.stackSize)
	prometheus.Unregister(ac.gcCount)
	prometheus.Unregister(ac.gcDuration)
	prometheus.Unregister(ac.allocRate)
	
	return nil
}

// IsRunning 检查收集器是否运行中
func (ac *ApplicationCollector) IsRunning() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.running
}

// GetStats 获取收集器统计信息
func (ac *ApplicationCollector) GetStats() CollectorStats {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.stats
}

// collectLoop 收集循环
func (ac *ApplicationCollector) collectLoop() {
	ticker := time.NewTicker(ac.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.collect()
		}
	}
}

// collect 执行一次收集
func (ac *ApplicationCollector) collect() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	ac.stats.CollectCount++
	ac.stats.LastCollectTime = time.Now()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 收集 Goroutine 数量
	ac.goroutineCount.Set(float64(runtime.NumGoroutine()))
	
	// 收集堆内存信息
	ac.heapSize.Set(float64(m.HeapSys))
	ac.heapUsed.Set(float64(m.HeapInuse))
	
	// 收集栈内存信息
	ac.stackSize.Set(float64(m.StackSys))
	
	// 收集 GC 信息
	if m.NumGC > ac.lastGCCount {
		newGCs := m.NumGC - ac.lastGCCount
		ac.gcCount.Add(float64(newGCs))
		
		// 计算平均 GC 时间
		if newGCs > 0 {
			lastGCTimeNs := ac.lastGCTime.UnixNano()
			avgGCTime := float64(m.PauseTotalNs-uint64(lastGCTimeNs)) / float64(newGCs) / 1e9
			ac.gcDuration.Observe(avgGCTime)
		}
		
		ac.lastGCCount = m.NumGC
		ac.lastGCTime = time.Unix(0, int64(m.PauseTotalNs))
	}
	
	// 计算内存分配速率
	now := time.Now()
	timeDiff := now.Sub(ac.lastSampleTime).Seconds()
	if timeDiff > 0 {
		allocDiff := m.TotalAlloc - ac.lastAllocBytes
		allocRate := float64(allocDiff) / timeDiff
		ac.allocRate.Set(allocRate)
		
		ac.lastAllocBytes = m.TotalAlloc
		ac.lastSampleTime = now
	}
	
	ac.logger.WithFields(logrus.Fields{
		"collect_count": ac.stats.CollectCount,
		"goroutines":    runtime.NumGoroutine(),
		"heap_used":     m.HeapInuse,
		"heap_size":     m.HeapSys,
		"gc_count":      m.NumGC,
		"timestamp":     ac.stats.LastCollectTime,
	}).Debug("Application metrics collected")
}

// Health 检查收集器健康状态
func (ac *ApplicationCollector) Health() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	
	status := "healthy"
	if !ac.running {
		status = "stopped"
	} else if time.Since(ac.stats.LastCollectTime) > ac.interval*2 {
		status = "unhealthy"
	}
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"status":            status,
		"running":           ac.running,
		"collect_count":     ac.stats.CollectCount,
		"error_count":       ac.stats.ErrorCount,
		"last_collect_time": ac.stats.LastCollectTime,
		"last_error":        ac.stats.LastError,
		"interval":          ac.interval.String(),
		"current_metrics": map[string]interface{}{
			"goroutines":  runtime.NumGoroutine(),
			"heap_used":   m.HeapInuse,
			"heap_size":   m.HeapSys,
			"stack_size":  m.StackSys,
			"gc_count":    m.NumGC,
			"alloc_total": m.TotalAlloc,
		},
	}
}

// GetCurrentMetrics 获取当前应用程序指标
func (ac *ApplicationCollector) GetCurrentMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"heap_used":      m.HeapInuse,
		"heap_size":      m.HeapSys,
		"heap_objects":   m.HeapObjects,
		"stack_size":     m.StackSys,
		"gc_count":       m.NumGC,
		"gc_pause_total": m.PauseTotalNs,
		"alloc_total":    m.TotalAlloc,
		"sys_total":      m.Sys,
		"lookups":        m.Lookups,
		"mallocs":        m.Mallocs,
		"frees":          m.Frees,
	}
}

// Name 获取收集器名称
func (ac *ApplicationCollector) Name() string {
	return "application"
}

// IsHealthy 检查收集器是否健康
func (ac *ApplicationCollector) IsHealthy() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	
	// 检查是否运行且最近有数据收集
	if !ac.running {
		return false
	}
	
	// 检查最近是否有数据收集（两个间隔时间内）
	return time.Since(ac.stats.LastCollectTime) < 2*ac.interval
}

// IsReady 检查收集器是否准备就绪
func (ac *ApplicationCollector) IsReady() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	
	// 收集器已初始化且运行即为准备就绪
	return ac.running
}