package exporters

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// PrometheusExporter Prometheus 导出器
type PrometheusExporter struct {
	mu       sync.RWMutex
	running  bool
	server   *http.Server
	registry *prometheus.Registry
	logger   *logrus.Logger
	config   PrometheusConfig

	// Statistics
	stats ExporterStats
}

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// ExporterStats 导出器统计信息
type ExporterStats struct {
	RequestCount    int64     `json:"request_count"`
	LastRequestTime time.Time `json:"last_request_time"`
	ErrorCount      int64     `json:"error_count"`
	LastError       string    `json:"last_error,omitempty"`
	StartTime       time.Time `json:"start_time"`
	ProcessedItems  int64     `json:"processed_items"`
}

// NewPrometheusExporter 创建新的 Prometheus 导出器
func NewPrometheusExporter(config PrometheusConfig, logger *logrus.Logger) *PrometheusExporter {
	registry := prometheus.NewRegistry()

	return &PrometheusExporter{
		registry: registry,
		logger:   logger,
		config:   config,
		stats: ExporterStats{
			StartTime: time.Now(),
		},
	}
}

// Start 启动导出器
func (pe *PrometheusExporter) Start(ctx context.Context) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if pe.running {
		return nil
	}

	if !pe.config.Enabled {
		pe.logger.Info("Prometheus exporter is disabled")
		return nil
	}

	pe.running = true
	pe.logger.Info("Starting Prometheus exporter")

	// 创建 HTTP 服务器
	mux := http.NewServeMux()

	// 添加指标处理器
	handler := promhttp.HandlerFor(pe.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Registry:          pe.registry,
	})

	// 包装处理器以收集统计信息
	wrappedHandler := pe.wrapHandler(handler)
	mux.Handle(pe.config.Path, wrappedHandler)

	// 添加健康检查端点
	mux.HandleFunc("/health", pe.healthHandler)
	mux.HandleFunc("/ready", pe.readyHandler)

	pe.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", pe.config.Host, pe.config.Port),
		Handler: mux,
	}

	// 启动服务器
	go func() {
		if err := pe.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pe.mu.Lock()
			pe.stats.ErrorCount++
			pe.stats.LastError = err.Error()
			pe.mu.Unlock()
			pe.logger.WithError(err).Error("Prometheus exporter server failed")
		}
	}()

	pe.logger.WithFields(logrus.Fields{
		"address": pe.server.Addr,
		"path":    pe.config.Path,
	}).Info("Prometheus exporter started")

	return nil
}

// Stop 停止导出器
func (pe *PrometheusExporter) Stop() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if !pe.running {
		return nil
	}

	pe.running = false
	pe.logger.Info("Stopping Prometheus exporter")

	if pe.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := pe.server.Shutdown(ctx); err != nil {
			pe.logger.WithError(err).Error("Failed to shutdown Prometheus exporter gracefully")
			return err
		}
	}

	return nil
}

// IsRunning 检查导出器是否运行中
func (pe *PrometheusExporter) IsRunning() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.running
}

// GetStats 获取导出器统计信息
func (pe *PrometheusExporter) GetStats() ExporterStats {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.stats
}

// GetRegistry 获取 Prometheus 注册表
func (pe *PrometheusExporter) GetRegistry() *prometheus.Registry {
	return pe.registry
}

// RegisterCollector 注册收集器
func (pe *PrometheusExporter) RegisterCollector(collector prometheus.Collector) error {
	return pe.registry.Register(collector)
}

// UnregisterCollector 注销收集器
func (pe *PrometheusExporter) UnregisterCollector(collector prometheus.Collector) bool {
	return pe.registry.Unregister(collector)
}

// wrapHandler 包装处理器以收集统计信息
func (pe *PrometheusExporter) wrapHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		pe.mu.Lock()
		pe.stats.RequestCount++
		pe.stats.LastRequestTime = start
		pe.mu.Unlock()

		// 记录请求
		pe.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.UserAgent(),
			"remote_ip":  r.RemoteAddr,
		}).Debug("Prometheus metrics request")

		handler.ServeHTTP(w, r)

		duration := time.Since(start)
		pe.logger.WithFields(logrus.Fields{
			"duration": duration,
			"path":     r.URL.Path,
		}).Debug("Prometheus metrics request completed")
	})
}

// healthHandler 健康检查处理器
func (pe *PrometheusExporter) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := pe.Health()

	w.Header().Set("Content-Type", "application/json")

	if health["status"] == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, `{
		"status": "%s",
		"timestamp": "%s",
		"uptime": "%s",
		"request_count": %d,
		"error_count": %d
	}`,
		health["status"],
		time.Now().Format(time.RFC3339),
		health["uptime"],
		pe.stats.RequestCount,
		pe.stats.ErrorCount,
	)
}

// readyHandler 就绪检查处理器
func (pe *PrometheusExporter) readyHandler(w http.ResponseWriter, r *http.Request) {
	if pe.IsRunning() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ready"}`)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"status": "not ready"}`)
	}
}

// Health 检查导出器健康状态
func (pe *PrometheusExporter) Health() map[string]interface{} {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	status := "healthy"
	if !pe.running {
		status = "stopped"
	} else if pe.stats.ErrorCount > 0 && time.Since(pe.stats.LastRequestTime) < time.Minute {
		status = "unhealthy"
	}

	uptime := time.Since(pe.stats.StartTime)

	return map[string]interface{}{
		"status":            status,
		"running":           pe.running,
		"uptime":            uptime.String(),
		"request_count":     pe.stats.RequestCount,
		"error_count":       pe.stats.ErrorCount,
		"last_request_time": pe.stats.LastRequestTime,
		"last_error":        pe.stats.LastError,
		"config": map[string]interface{}{
			"enabled": pe.config.Enabled,
			"address": fmt.Sprintf("%s:%d", pe.config.Host, pe.config.Port),
			"path":    pe.config.Path,
		},
	}
}

// GetMetrics 获取当前指标
func (pe *PrometheusExporter) GetMetrics() (map[string]interface{}, error) {
	metricFamilies, err := pe.registry.Gather()
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]interface{})
	for _, mf := range metricFamilies {
		metrics[mf.GetName()] = map[string]interface{}{
			"type":   mf.GetType().String(),
			"help":   mf.GetHelp(),
			"values": len(mf.GetMetric()),
		}
	}

	return map[string]interface{}{
		"metrics":      metrics,
		"total_count":  len(metricFamilies),
		"collected_at": time.Now(),
	}, nil
}

// Name 返回导出器名称
func (pe *PrometheusExporter) Name() string {
	return "prometheus"
}

// IsHealthy 检查导出器健康状态
func (pe *PrometheusExporter) IsHealthy() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	return pe.running && pe.stats.ErrorCount == 0
}

// IsReady 检查导出器就绪状态
func (pe *PrometheusExporter) IsReady() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	return pe.config.Enabled
}

// Export 导出数据（Prometheus通过HTTP端点暴露指标，此方法为接口兼容性）
func (pe *PrometheusExporter) Export(data interface{}) error {
	// Prometheus导出器通过HTTP端点暴露指标，不需要主动导出
	// 这里只是为了实现Exporter接口
	return nil
}
