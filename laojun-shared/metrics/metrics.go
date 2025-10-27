package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics 监控指标接口
type Metrics interface {
	// 计数器
	IncCounter(name string, labels map[string]string)
	AddCounter(name string, value float64, labels map[string]string)

	// 直方图
	ObserveHistogram(name string, value float64, labels map[string]string)

	// 仪表盘
	SetGauge(name string, value float64, labels map[string]string)
	IncGauge(name string, labels map[string]string)
	DecGauge(name string, labels map[string]string)
	AddGauge(name string, value float64, labels map[string]string)

	// 摘要
	ObserveSummary(name string, value float64, labels map[string]string)

	// HTTP 指标
	RecordHTTPRequest(method, path, status string, duration time.Duration)
	RecordHTTPRequestSize(method, path string, size float64)
	RecordHTTPResponseSize(method, path string, size float64)

	// 数据库指标
	RecordDBQuery(operation, table string, duration time.Duration, success bool)
	RecordDBConnection(database string, active, idle, total int)

	// 缓存指标
	RecordCacheOperation(operation string, hit bool, duration time.Duration)

	// 业务指标
	RecordUserAction(action string, userID string)
	RecordPluginOperation(operation, pluginID string, duration time.Duration, success bool)

	// 系统指标
	RecordMemoryUsage(usage float64)
	RecordCPUUsage(usage float64)
	RecordGoroutineCount(count int)

	// 获取 HTTP 处理函数
	Handler() http.Handler
}

// Config 监控配置
type Config struct {
	Enabled     bool   `yaml:"enabled" env:"METRICS_ENABLED" config:"metrics.enabled" default:"true"`
	Path        string `yaml:"path" env:"METRICS_PATH" config:"metrics.path" default:"/metrics"`
	Namespace   string `yaml:"namespace" env:"METRICS_NAMESPACE" config:"metrics.namespace" default:"laojun"`
	Subsystem   string `yaml:"subsystem" env:"METRICS_SUBSYSTEM" config:"metrics.subsystem" default:""`
	Service     string `yaml:"service" env:"SERVICE_NAME" config:"metrics.service" default:"unknown"`
	Version     string `yaml:"version" env:"SERVICE_VERSION" config:"metrics.version" default:"unknown"`
	Environment string `yaml:"environment" env:"ENVIRONMENT" config:"metrics.environment" default:"development"`
}

// prometheusMetrics Prometheus 实现
type prometheusMetrics struct {
	config   Config
	registry *prometheus.Registry

	// HTTP 指标
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// 数据库指标
	dbQueryDuration *prometheus.HistogramVec
	dbQueryTotal    *prometheus.CounterVec
	dbConnections   *prometheus.GaugeVec

	// 缓存指标
	cacheOperations *prometheus.CounterVec
	cacheHitRatio   *prometheus.GaugeVec
	cacheDuration   *prometheus.HistogramVec

	// 业务指标
	userActions      *prometheus.CounterVec
	pluginOperations *prometheus.CounterVec
	pluginDuration   *prometheus.HistogramVec

	// 系统指标
	memoryUsage    prometheus.Gauge
	cpuUsage       prometheus.Gauge
	goroutineCount prometheus.Gauge

	// 通用指标
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	gauges     map[string]*prometheus.GaugeVec
	summaries  map[string]*prometheus.SummaryVec
}

// New 创建新的监控实例
func New(config Config) Metrics {
	registry := prometheus.NewRegistry()

	// 添加默认收集器
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	m := &prometheusMetrics{
		config:     config,
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}

	m.initMetrics()
	return m
}

// initMetrics 初始化指标
func (m *prometheusMetrics) initMetrics() {
	// HTTP 指标
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "service", "version"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "service", "version"},
	)

	m.httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{"method", "path", "service", "version"},
	)

	m.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{"method", "path", "service", "version"},
	)

	// 数据库指标
	m.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},
		[]string{"operation", "table", "service", "version"},
	)

	m.dbQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "db_queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"operation", "table", "status", "service", "version"},
	)

	m.dbConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "db_connections",
			Help:      "Number of database connections",
		},
		[]string{"database", "state", "service", "version"},
	)

	// 缓存指标
	m.cacheOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "cache_operations_total",
			Help:      "Total number of cache operations",
		},
		[]string{"operation", "result", "service", "version"},
	)

	m.cacheHitRatio = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "cache_hit_ratio",
			Help:      "Cache hit ratio",
		},
		[]string{"service", "version"},
	)

	m.cacheDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "cache_operation_duration_seconds",
			Help:      "Cache operation duration in seconds",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{"operation", "service", "version"},
	)

	// 业务指标
	m.userActions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "user_actions_total",
			Help:      "Total number of user actions",
		},
		[]string{"action", "service", "version"},
	)

	m.pluginOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "plugin_operations_total",
			Help:      "Total number of plugin operations",
		},
		[]string{"operation", "plugin_id", "status", "service", "version"},
	)

	m.pluginDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "plugin_operation_duration_seconds",
			Help:      "Plugin operation duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation", "plugin_id", "service", "version"},
	)

	// 系统指标
	m.memoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes",
		},
	)

	m.cpuUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "cpu_usage_percent",
			Help:      "CPU usage percentage",
		},
	)

	m.goroutineCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      "goroutines_count",
			Help:      "Number of goroutines",
		},
	)

	// 注册所有指标
	m.registry.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestSize,
		m.httpResponseSize,
		m.dbQueryDuration,
		m.dbQueryTotal,
		m.dbConnections,
		m.cacheOperations,
		m.cacheHitRatio,
		m.cacheDuration,
		m.userActions,
		m.pluginOperations,
		m.pluginDuration,
		m.memoryUsage,
		m.cpuUsage,
		m.goroutineCount,
	)
}

// IncCounter 增加计数器
func (m *prometheusMetrics) IncCounter(name string, labels map[string]string) {
	m.AddCounter(name, 1, labels)
}

// AddCounter 添加计数器
func (m *prometheusMetrics) AddCounter(name string, value float64, labels map[string]string) {
	counter := m.getOrCreateCounter(name, getKeys(labels))
	counter.With(m.addServiceLabels(labels)).Add(value)
}

// ObserveHistogram 观察直方图
func (m *prometheusMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {
	histogram := m.getOrCreateHistogram(name, getKeys(labels))
	histogram.With(m.addServiceLabels(labels)).Observe(value)
}

// SetGauge 设置仪表盘
func (m *prometheusMetrics) SetGauge(name string, value float64, labels map[string]string) {
	gauge := m.getOrCreateGauge(name, getKeys(labels))
	gauge.With(m.addServiceLabels(labels)).Set(value)
}

// IncGauge 增加仪表盘
func (m *prometheusMetrics) IncGauge(name string, labels map[string]string) {
	gauge := m.getOrCreateGauge(name, getKeys(labels))
	gauge.With(m.addServiceLabels(labels)).Inc()
}

// DecGauge 减少仪表盘
func (m *prometheusMetrics) DecGauge(name string, labels map[string]string) {
	gauge := m.getOrCreateGauge(name, getKeys(labels))
	gauge.With(m.addServiceLabels(labels)).Dec()
}

// AddGauge 添加仪表盘
func (m *prometheusMetrics) AddGauge(name string, value float64, labels map[string]string) {
	gauge := m.getOrCreateGauge(name, getKeys(labels))
	gauge.With(m.addServiceLabels(labels)).Add(value)
}

// ObserveSummary 观察摘要
func (m *prometheusMetrics) ObserveSummary(name string, value float64, labels map[string]string) {
	summary := m.getOrCreateSummary(name, getKeys(labels))
	summary.With(m.addServiceLabels(labels)).Observe(value)
}

// RecordHTTPRequest 记录 HTTP 请求
func (m *prometheusMetrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	labels := prometheus.Labels{
		"method":  method,
		"path":    path,
		"status":  status,
		"service": m.config.Service,
		"version": m.config.Version,
	}

	m.httpRequestsTotal.With(labels).Inc()

	durationLabels := prometheus.Labels{
		"method":  method,
		"path":    path,
		"service": m.config.Service,
		"version": m.config.Version,
	}
	m.httpRequestDuration.With(durationLabels).Observe(duration.Seconds())
}

// RecordHTTPRequestSize 记录 HTTP 请求大小
func (m *prometheusMetrics) RecordHTTPRequestSize(method, path string, size float64) {
	labels := prometheus.Labels{
		"method":  method,
		"path":    path,
		"service": m.config.Service,
		"version": m.config.Version,
	}
	m.httpRequestSize.With(labels).Observe(size)
}

// RecordHTTPResponseSize 记录 HTTP 响应大小
func (m *prometheusMetrics) RecordHTTPResponseSize(method, path string, size float64) {
	labels := prometheus.Labels{
		"method":  method,
		"path":    path,
		"service": m.config.Service,
		"version": m.config.Version,
	}
	m.httpResponseSize.With(labels).Observe(size)
}

// RecordDBQuery 记录数据库查询
func (m *prometheusMetrics) RecordDBQuery(operation, table string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	queryLabels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"status":    status,
		"service":   m.config.Service,
		"version":   m.config.Version,
	}
	m.dbQueryTotal.With(queryLabels).Inc()

	durationLabels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"service":   m.config.Service,
		"version":   m.config.Version,
	}
	m.dbQueryDuration.With(durationLabels).Observe(duration.Seconds())
}

// RecordDBConnection 记录数据库连接
func (m *prometheusMetrics) RecordDBConnection(database string, active, idle, total int) {
	baseLabels := prometheus.Labels{
		"database": database,
		"service":  m.config.Service,
		"version":  m.config.Version,
	}

	activeLabels := prometheus.Labels{}
	for k, v := range baseLabels {
		activeLabels[k] = v
	}
	activeLabels["state"] = "active"
	m.dbConnections.With(activeLabels).Set(float64(active))

	idleLabels := prometheus.Labels{}
	for k, v := range baseLabels {
		idleLabels[k] = v
	}
	idleLabels["state"] = "idle"
	m.dbConnections.With(idleLabels).Set(float64(idle))

	totalLabels := prometheus.Labels{}
	for k, v := range baseLabels {
		totalLabels[k] = v
	}
	totalLabels["state"] = "total"
	m.dbConnections.With(totalLabels).Set(float64(total))
}

// RecordCacheOperation 记录缓存操作
func (m *prometheusMetrics) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	result := "miss"
	if hit {
		result = "hit"
	}

	labels := prometheus.Labels{
		"operation": operation,
		"result":    result,
		"service":   m.config.Service,
		"version":   m.config.Version,
	}
	m.cacheOperations.With(labels).Inc()

	durationLabels := prometheus.Labels{
		"operation": operation,
		"service":   m.config.Service,
		"version":   m.config.Version,
	}
	m.cacheDuration.With(durationLabels).Observe(duration.Seconds())
}

// RecordUserAction 记录用户行为
func (m *prometheusMetrics) RecordUserAction(action string, userID string) {
	labels := prometheus.Labels{
		"action":  action,
		"service": m.config.Service,
		"version": m.config.Version,
	}
	m.userActions.With(labels).Inc()
}

// RecordPluginOperation 记录插件操作
func (m *prometheusMetrics) RecordPluginOperation(operation, pluginID string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	labels := prometheus.Labels{
		"operation": operation,
		"plugin_id": pluginID,
		"status":    status,
		"service":   m.config.Service,
		"version":   m.config.Version,
	}
	m.pluginOperations.With(labels).Inc()

	durationLabels := prometheus.Labels{
		"operation": operation,
		"plugin_id": pluginID,
		"service":   m.config.Service,
		"version":   m.config.Version,
	}
	m.pluginDuration.With(durationLabels).Observe(duration.Seconds())
}

// RecordMemoryUsage 记录内存使用
func (m *prometheusMetrics) RecordMemoryUsage(usage float64) {
	m.memoryUsage.Set(usage)
}

// RecordCPUUsage 记录 CPU 使用
func (m *prometheusMetrics) RecordCPUUsage(usage float64) {
	m.cpuUsage.Set(usage)
}

// RecordGoroutineCount 记录协程数量
func (m *prometheusMetrics) RecordGoroutineCount(count int) {
	m.goroutineCount.Set(float64(count))
}

// Handler 获取 HTTP 处理程序
func (m *prometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// 辅助方法
func (m *prometheusMetrics) getOrCreateCounter(name string, labelNames []string) *prometheus.CounterVec {
	if counter, exists := m.counters[name]; exists {
		return counter
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      name,
			Help:      "Custom counter metric",
		},
		append(labelNames, "service", "version"),
	)

	m.registry.MustRegister(counter)
	m.counters[name] = counter
	return counter
}

func (m *prometheusMetrics) getOrCreateHistogram(name string, labelNames []string) *prometheus.HistogramVec {
	if histogram, exists := m.histograms[name]; exists {
		return histogram
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      name,
			Help:      "Custom histogram metric",
			Buckets:   prometheus.DefBuckets,
		},
		append(labelNames, "service", "version"),
	)

	m.registry.MustRegister(histogram)
	m.histograms[name] = histogram
	return histogram
}

func (m *prometheusMetrics) getOrCreateGauge(name string, labelNames []string) *prometheus.GaugeVec {
	if gauge, exists := m.gauges[name]; exists {
		return gauge
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      name,
			Help:      "Custom gauge metric",
		},
		append(labelNames, "service", "version"),
	)

	m.registry.MustRegister(gauge)
	m.gauges[name] = gauge
	return gauge
}

func (m *prometheusMetrics) getOrCreateSummary(name string, labelNames []string) *prometheus.SummaryVec {
	if summary, exists := m.summaries[name]; exists {
		return summary
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: m.config.Namespace,
			Subsystem: m.config.Subsystem,
			Name:      name,
			Help:      "Custom summary metric",
		},
		append(labelNames, "service", "version"),
	)

	m.registry.MustRegister(summary)
	m.summaries[name] = summary
	return summary
}

func (m *prometheusMetrics) addServiceLabels(labels map[string]string) prometheus.Labels {
	result := prometheus.Labels{
		"service": m.config.Service,
		"version": m.config.Version,
	}

	for k, v := range labels {
		result[k] = v
	}

	return result
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Gin 中间件
func GinMiddleware(metrics Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			status,
			duration,
		)

		if c.Request.ContentLength > 0 {
			metrics.RecordHTTPRequestSize(
				c.Request.Method,
				c.FullPath(),
				float64(c.Request.ContentLength),
			)
		}

		metrics.RecordHTTPResponseSize(
			c.Request.Method,
			c.FullPath(),
			float64(c.Writer.Size()),
		)
	}
}

// DefaultMetrics 默认监控实例
var DefaultMetrics Metrics

// init 初始化默认监控实例
func init() {
	DefaultMetrics = New(Config{
		Enabled:     true,
		Path:        "/metrics",
		Namespace:   "laojun",
		Service:     "unknown",
		Version:     "unknown",
		Environment: "development",
	})
}
