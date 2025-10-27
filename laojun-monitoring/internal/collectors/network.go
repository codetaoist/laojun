package collectors

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
)

// NetworkCollector 网络指标收集器
type NetworkCollector struct {
	mu       sync.RWMutex
	running  bool
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *logrus.Logger

	// Prometheus metrics
	bytesReceived    prometheus.Counter
	bytesSent        prometheus.Counter
	packetsReceived  prometheus.Counter
	packetsSent      prometheus.Counter
	errorsReceived   prometheus.Counter
	errorsSent       prometheus.Counter
	dropsReceived    prometheus.Counter
	dropsSent        prometheus.Counter
	
	// Statistics
	stats CollectorStats
	
	// Last values for calculating deltas
	lastBytesRecv    uint64
	lastBytesSent    uint64
	lastPacketsRecv  uint64
	lastPacketsSent  uint64
	lastErrsRecv     uint64
	lastErrsSent     uint64
	lastDropsRecv    uint64
	lastDropsSent    uint64
}

// NewNetworkCollector 创建新的网络收集器
func NewNetworkCollector(interval time.Duration, logger *logrus.Logger) *NetworkCollector {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &NetworkCollector{
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		logger:   logger,
		bytesReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_bytes_received_total",
			Help: "Total bytes received",
		}),
		bytesSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_bytes_sent_total",
			Help: "Total bytes sent",
		}),
		packetsReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_packets_received_total",
			Help: "Total packets received",
		}),
		packetsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_packets_sent_total",
			Help: "Total packets sent",
		}),
		errorsReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_errors_received_total",
			Help: "Total receive errors",
		}),
		errorsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_errors_sent_total",
			Help: "Total send errors",
		}),
		dropsReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_drops_received_total",
			Help: "Total receive drops",
		}),
		dropsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "network_drops_sent_total",
			Help: "Total send drops",
		}),
	}
}

// Start 启动收集器
func (nc *NetworkCollector) Start() error {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	
	if nc.running {
		return nil
	}
	
	nc.running = true
	nc.logger.Info("Starting network collector")
	
	// 注册 Prometheus 指标
	prometheus.MustRegister(nc.bytesReceived)
	prometheus.MustRegister(nc.bytesSent)
	prometheus.MustRegister(nc.packetsReceived)
	prometheus.MustRegister(nc.packetsSent)
	prometheus.MustRegister(nc.errorsReceived)
	prometheus.MustRegister(nc.errorsSent)
	prometheus.MustRegister(nc.dropsReceived)
	prometheus.MustRegister(nc.dropsSent)
	
	// 初始化基线数据
	nc.initializeBaseline()
	
	go nc.collectLoop()
	
	return nil
}

// Stop 停止收集器
func (nc *NetworkCollector) Stop() error {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	
	if !nc.running {
		return nil
	}
	
	nc.running = false
	nc.cancel()
	nc.logger.Info("Stopping network collector")
	
	// 注销 Prometheus 指标
	prometheus.Unregister(nc.bytesReceived)
	prometheus.Unregister(nc.bytesSent)
	prometheus.Unregister(nc.packetsReceived)
	prometheus.Unregister(nc.packetsSent)
	prometheus.Unregister(nc.errorsReceived)
	prometheus.Unregister(nc.errorsSent)
	prometheus.Unregister(nc.dropsReceived)
	prometheus.Unregister(nc.dropsSent)
	
	return nil
}

// IsRunning 检查收集器是否运行中
func (nc *NetworkCollector) IsRunning() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.running
}

// GetStats 获取收集器统计信息
func (nc *NetworkCollector) GetStats() CollectorStats {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.stats
}

// initializeBaseline 初始化基线数据
func (nc *NetworkCollector) initializeBaseline() {
	if netIO, err := net.IOCounters(false); err == nil && len(netIO) > 0 {
		nc.lastBytesRecv = netIO[0].BytesRecv
		nc.lastBytesSent = netIO[0].BytesSent
		nc.lastPacketsRecv = netIO[0].PacketsRecv
		nc.lastPacketsSent = netIO[0].PacketsSent
		nc.lastErrsRecv = netIO[0].Errin
		nc.lastErrsSent = netIO[0].Errout
		nc.lastDropsRecv = netIO[0].Dropin
		nc.lastDropsSent = netIO[0].Dropout
	}
}

// collectLoop 收集循环
func (nc *NetworkCollector) collectLoop() {
	ticker := time.NewTicker(nc.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-nc.ctx.Done():
			return
		case <-ticker.C:
			nc.collect()
		}
	}
}

// collect 执行一次收集
func (nc *NetworkCollector) collect() {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	
	nc.stats.CollectCount++
	nc.stats.LastCollectTime = time.Now()
	
	// 收集网络 I/O 统计
	if netIO, err := net.IOCounters(false); err == nil && len(netIO) > 0 {
		current := netIO[0]
		
		// 计算增量并更新计数器
		if current.BytesRecv >= nc.lastBytesRecv {
			nc.bytesReceived.Add(float64(current.BytesRecv - nc.lastBytesRecv))
		}
		if current.BytesSent >= nc.lastBytesSent {
			nc.bytesSent.Add(float64(current.BytesSent - nc.lastBytesSent))
		}
		if current.PacketsRecv >= nc.lastPacketsRecv {
			nc.packetsReceived.Add(float64(current.PacketsRecv - nc.lastPacketsRecv))
		}
		if current.PacketsSent >= nc.lastPacketsSent {
			nc.packetsSent.Add(float64(current.PacketsSent - nc.lastPacketsSent))
		}
		if current.Errin >= nc.lastErrsRecv {
			nc.errorsReceived.Add(float64(current.Errin - nc.lastErrsRecv))
		}
		if current.Errout >= nc.lastErrsSent {
			nc.errorsSent.Add(float64(current.Errout - nc.lastErrsSent))
		}
		if current.Dropin >= nc.lastDropsRecv {
			nc.dropsReceived.Add(float64(current.Dropin - nc.lastDropsRecv))
		}
		if current.Dropout >= nc.lastDropsSent {
			nc.dropsSent.Add(float64(current.Dropout - nc.lastDropsSent))
		}
		
		// 更新基线值
		nc.lastBytesRecv = current.BytesRecv
		nc.lastBytesSent = current.BytesSent
		nc.lastPacketsRecv = current.PacketsRecv
		nc.lastPacketsSent = current.PacketsSent
		nc.lastErrsRecv = current.Errin
		nc.lastErrsSent = current.Errout
		nc.lastDropsRecv = current.Dropin
		nc.lastDropsSent = current.Dropout
		
		nc.logger.WithFields(logrus.Fields{
			"collect_count":    nc.stats.CollectCount,
			"bytes_received":   current.BytesRecv,
			"bytes_sent":       current.BytesSent,
			"packets_received": current.PacketsRecv,
			"packets_sent":     current.PacketsSent,
			"timestamp":        nc.stats.LastCollectTime,
		}).Debug("Network metrics collected")
	} else if err != nil {
		nc.handleError(err)
	}
}

// handleError 处理错误
func (nc *NetworkCollector) handleError(err error) {
	nc.stats.ErrorCount++
	nc.stats.LastError = err.Error()
	nc.logger.WithFields(logrus.Fields{
		"error": err,
	}).Error("Failed to collect network metrics")
}

// Health 检查收集器健康状态
func (nc *NetworkCollector) Health() map[string]interface{} {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	
	status := "healthy"
	if !nc.running {
		status = "stopped"
	} else if time.Since(nc.stats.LastCollectTime) > nc.interval*2 {
		status = "unhealthy"
	}
	
	return map[string]interface{}{
		"status":            status,
		"running":           nc.running,
		"collect_count":     nc.stats.CollectCount,
		"error_count":       nc.stats.ErrorCount,
		"last_collect_time": nc.stats.LastCollectTime,
		"last_error":        nc.stats.LastError,
		"interval":          nc.interval.String(),
	}
}

// GetCurrentMetrics 获取当前网络指标
func (nc *NetworkCollector) GetCurrentMetrics() map[string]interface{} {
	if netIO, err := net.IOCounters(false); err == nil && len(netIO) > 0 {
		current := netIO[0]
		return map[string]interface{}{
			"bytes_received":   current.BytesRecv,
			"bytes_sent":       current.BytesSent,
			"packets_received": current.PacketsRecv,
			"packets_sent":     current.PacketsSent,
			"errors_received":  current.Errin,
			"errors_sent":      current.Errout,
			"drops_received":   current.Dropin,
			"drops_sent":       current.Dropout,
		}
	}
	return map[string]interface{}{}
}

// Name 获取收集器名称
func (nc *NetworkCollector) Name() string {
	return "network"
}

// IsHealthy 检查收集器是否健康
func (nc *NetworkCollector) IsHealthy() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	
	// 检查是否运行且最近有数据收集
	if !nc.running {
		return false
	}
	
	// 检查最近是否有数据收集（两个间隔时间内）
	return time.Since(nc.stats.LastCollectTime) < 2*nc.interval
}

// IsReady 检查收集器是否准备就绪
func (nc *NetworkCollector) IsReady() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	
	// 收集器已初始化且运行即为准备就绪
	return nc.running
}