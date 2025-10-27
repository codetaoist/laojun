package collectors

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
)

// SystemCollector 系统指标收集器
type SystemCollector struct {
	mu       sync.RWMutex
	running  bool
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *logrus.Logger

	// Prometheus metrics
	cpuUsage      prometheus.Gauge
	memoryUsage   prometheus.Gauge
	diskUsage     prometheus.Gauge
	networkIO     *prometheus.CounterVec
	loadAverage   prometheus.Gauge
	uptime        prometheus.Gauge
	processCount  prometheus.Gauge
	
	// Statistics
	stats CollectorStats
}



// NewSystemCollector 创建新的系统收集器
func NewSystemCollector(interval time.Duration, logger *logrus.Logger) *SystemCollector {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SystemCollector{
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		cpuUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		}),
		memoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_memory_usage_percent",
			Help: "Current memory usage percentage",
		}),
		diskUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_disk_usage_percent",
			Help: "Current disk usage percentage",
		}),
		networkIO: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "system_network_io_bytes_total",
			Help: "Total network I/O bytes",
		}, []string{"direction"}),
		loadAverage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_load_average",
			Help: "System load average (1 minute)",
		}),
		uptime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_uptime_seconds",
			Help: "System uptime in seconds",
		}),
		processCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_process_count",
			Help: "Number of running processes",
		}),
	}
}

// Start 启动收集器
func (sc *SystemCollector) Start() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if sc.running {
		return nil
	}
	
	sc.running = true
	sc.logger.Info("Starting system collector")
	
	// 注册 Prometheus 指标
	prometheus.MustRegister(sc.cpuUsage)
	prometheus.MustRegister(sc.memoryUsage)
	prometheus.MustRegister(sc.diskUsage)
	prometheus.MustRegister(sc.networkIO)
	prometheus.MustRegister(sc.loadAverage)
	prometheus.MustRegister(sc.uptime)
	prometheus.MustRegister(sc.processCount)
	
	go sc.collectLoop()
	
	return nil
}

// Stop 停止收集器
func (sc *SystemCollector) Stop() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if !sc.running {
		return nil
	}
	
	sc.running = false
	sc.cancel()
	sc.logger.Info("Stopping system collector")
	
	// 注销 Prometheus 指标
	prometheus.Unregister(sc.cpuUsage)
	prometheus.Unregister(sc.memoryUsage)
	prometheus.Unregister(sc.diskUsage)
	prometheus.Unregister(sc.networkIO)
	prometheus.Unregister(sc.loadAverage)
	prometheus.Unregister(sc.uptime)
	prometheus.Unregister(sc.processCount)
	
	return nil
}

// IsRunning 检查收集器是否运行中
func (sc *SystemCollector) IsRunning() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.running
}

// GetStats 获取收集器统计信息
func (sc *SystemCollector) GetStats() CollectorStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.stats
}

// collectLoop 收集循环
func (sc *SystemCollector) collectLoop() {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sc.ctx.Done():
			return
		case <-ticker.C:
			sc.collect()
		}
	}
}

// collect 执行一次收集
func (sc *SystemCollector) collect() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.stats.CollectCount++
	sc.stats.LastCollectTime = time.Now()
	
	// 收集 CPU 使用率
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		sc.cpuUsage.Set(cpuPercent[0])
	} else if err != nil {
		sc.handleError("CPU", err)
	}
	
	// 收集内存使用率
	if memInfo, err := mem.VirtualMemory(); err == nil {
		sc.memoryUsage.Set(memInfo.UsedPercent)
	} else {
		sc.handleError("Memory", err)
	}
	
	// 收集磁盘使用率
	if diskInfo, err := disk.Usage("/"); err == nil {
		sc.diskUsage.Set(diskInfo.UsedPercent)
	} else {
		sc.handleError("Disk", err)
	}
	
	// 收集网络 I/O
	if netIO, err := net.IOCounters(false); err == nil && len(netIO) > 0 {
		sc.networkIO.WithLabelValues("rx").Add(float64(netIO[0].BytesRecv))
		sc.networkIO.WithLabelValues("tx").Add(float64(netIO[0].BytesSent))
	} else if err != nil {
		sc.handleError("Network", err)
	}
	
	// 收集系统负载
	if loadInfo, err := load.Avg(); err == nil {
		sc.loadAverage.Set(loadInfo.Load1)
	} else {
		sc.handleError("Load", err)
	}
	
	// 收集系统运行时间
	if hostInfo, err := host.Info(); err == nil {
		sc.uptime.Set(float64(hostInfo.Uptime))
	} else {
		sc.handleError("Host", err)
	}
	
	// 收集进程数量
	sc.processCount.Set(float64(runtime.NumGoroutine()))
	
	sc.logger.WithFields(logrus.Fields{
		"collect_count": sc.stats.CollectCount,
		"timestamp":     sc.stats.LastCollectTime,
	}).Debug("System metrics collected")
}

// handleError 处理错误
func (sc *SystemCollector) handleError(component string, err error) {
	sc.stats.ErrorCount++
	sc.stats.LastError = err.Error()
	sc.logger.WithFields(logrus.Fields{
		"component": component,
		"error":     err,
	}).Error("Failed to collect system metric")
}

// Health 检查收集器健康状态
func (sc *SystemCollector) Health() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	status := "healthy"
	if !sc.running {
		status = "stopped"
	} else if time.Since(sc.stats.LastCollectTime) > sc.interval*2 {
		status = "unhealthy"
	}
	
	return map[string]interface{}{
		"status":            status,
		"running":           sc.running,
		"collect_count":     sc.stats.CollectCount,
		"error_count":       sc.stats.ErrorCount,
		"last_collect_time": sc.stats.LastCollectTime,
		"last_error":        sc.stats.LastError,
		"interval":          sc.interval.String(),
	}
}

// Name 获取收集器名称
func (sc *SystemCollector) Name() string {
	return "system"
}

// IsHealthy 检查收集器是否健康
func (sc *SystemCollector) IsHealthy() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	if !sc.running {
		return false
	}
	
	// 如果超过2个采集间隔没有采集，认为不健康
	return time.Since(sc.stats.LastCollectTime) <= sc.interval*2
}

// IsReady 检查收集器是否准备就绪
func (sc *SystemCollector) IsReady() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	// 收集器已初始化且运行即为准备就绪
	return sc.running
}