package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// MetricDefinition 指标定义
type MetricDefinition struct {
	Name        string            `json:"name"`
	Help        string            `json:"help"`
	Type        MetricType        `json:"type"`
	Labels      []string          `json:"labels"`
	Buckets     []float64         `json:"buckets,omitempty"`     // 用于Histogram
	Objectives  map[float64]float64 `json:"objectives,omitempty"` // 用于Summary
	Namespace   string            `json:"namespace"`
	Subsystem   string            `json:"subsystem"`
}

// MetricRegistry 指标注册表
type MetricRegistry struct {
	registry    *prometheus.Registry
	metrics     map[string]prometheus.Collector
	definitions map[string]*MetricDefinition
	logger      *zap.Logger
	mu          sync.RWMutex
}

// NewMetricRegistry 创建指标注册表
func NewMetricRegistry(logger *zap.Logger) *MetricRegistry {
	return &MetricRegistry{
		registry:    prometheus.NewRegistry(),
		metrics:     make(map[string]prometheus.Collector),
		definitions: make(map[string]*MetricDefinition),
		logger:      logger,
	}
}

// RegisterMetric 注册指标
func (r *MetricRegistry) RegisterMetric(def *MetricDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.definitions[def.Name]; exists {
		return ErrMetricAlreadyExists
	}

	var collector prometheus.Collector
	var err error

	switch def.Type {
	case MetricTypeCounter:
		collector = r.createCounter(def)
	case MetricTypeGauge:
		collector = r.createGauge(def)
	case MetricTypeHistogram:
		collector = r.createHistogram(def)
	case MetricTypeSummary:
		collector = r.createSummary(def)
	default:
		return ErrUnsupportedMetricType
	}

	if err = r.registry.Register(collector); err != nil {
		return err
	}

	r.metrics[def.Name] = collector
	r.definitions[def.Name] = def

	r.logger.Info("Metric registered successfully",
		zap.String("name", def.Name),
		zap.String("type", string(def.Type)))

	return nil
}

// UnregisterMetric 注销指标
func (r *MetricRegistry) UnregisterMetric(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	collector, exists := r.metrics[name]
	if !exists {
		return ErrMetricNotFound
	}

	if !r.registry.Unregister(collector) {
		return ErrMetricUnregisterFailed
	}

	delete(r.metrics, name)
	delete(r.definitions, name)

	r.logger.Info("Metric unregistered successfully", zap.String("name", name))
	return nil
}

// GetMetric 获取指标
func (r *MetricRegistry) GetMetric(name string) (prometheus.Collector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collector, exists := r.metrics[name]
	if !exists {
		return nil, ErrMetricNotFound
	}

	return collector, nil
}

// GetCounter 获取计数器指标
func (r *MetricRegistry) GetCounter(name string) (*prometheus.CounterVec, error) {
	collector, err := r.GetMetric(name)
	if err != nil {
		return nil, err
	}

	counter, ok := collector.(*prometheus.CounterVec)
	if !ok {
		return nil, ErrMetricTypeMismatch
	}

	return counter, nil
}

// GetGauge 获取仪表盘指标
func (r *MetricRegistry) GetGauge(name string) (*prometheus.GaugeVec, error) {
	collector, err := r.GetMetric(name)
	if err != nil {
		return nil, err
	}

	gauge, ok := collector.(*prometheus.GaugeVec)
	if !ok {
		return nil, ErrMetricTypeMismatch
	}

	return gauge, nil
}

// GetHistogram 获取直方图指标
func (r *MetricRegistry) GetHistogram(name string) (*prometheus.HistogramVec, error) {
	collector, err := r.GetMetric(name)
	if err != nil {
		return nil, err
	}

	histogram, ok := collector.(*prometheus.HistogramVec)
	if !ok {
		return nil, ErrMetricTypeMismatch
	}

	return histogram, nil
}

// GetSummary 获取摘要指标
func (r *MetricRegistry) GetSummary(name string) (*prometheus.SummaryVec, error) {
	collector, err := r.GetMetric(name)
	if err != nil {
		return nil, err
	}

	summary, ok := collector.(*prometheus.SummaryVec)
	if !ok {
		return nil, ErrMetricTypeMismatch
	}

	return summary, nil
}

// ListMetrics 列出所有指标
func (r *MetricRegistry) ListMetrics() []*MetricDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]*MetricDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		metrics = append(metrics, def)
	}

	return metrics
}

// GetRegistry 获取Prometheus注册表
func (r *MetricRegistry) GetRegistry() *prometheus.Registry {
	return r.registry
}

// createCounter 创建计数器
func (r *MetricRegistry) createCounter(def *MetricDefinition) prometheus.Collector {
	opts := prometheus.CounterOpts{
		Namespace: def.Namespace,
		Subsystem: def.Subsystem,
		Name:      def.Name,
		Help:      def.Help,
	}

	if len(def.Labels) > 0 {
		return prometheus.NewCounterVec(opts, def.Labels)
	}
	return prometheus.NewCounter(opts)
}

// createGauge 创建仪表盘
func (r *MetricRegistry) createGauge(def *MetricDefinition) prometheus.Collector {
	opts := prometheus.GaugeOpts{
		Namespace: def.Namespace,
		Subsystem: def.Subsystem,
		Name:      def.Name,
		Help:      def.Help,
	}

	if len(def.Labels) > 0 {
		return prometheus.NewGaugeVec(opts, def.Labels)
	}
	return prometheus.NewGauge(opts)
}

// createHistogram 创建直方图
func (r *MetricRegistry) createHistogram(def *MetricDefinition) prometheus.Collector {
	opts := prometheus.HistogramOpts{
		Namespace: def.Namespace,
		Subsystem: def.Subsystem,
		Name:      def.Name,
		Help:      def.Help,
	}

	if len(def.Buckets) > 0 {
		opts.Buckets = def.Buckets
	}

	if len(def.Labels) > 0 {
		return prometheus.NewHistogramVec(opts, def.Labels)
	}
	return prometheus.NewHistogram(opts)
}

// createSummary 创建摘要
func (r *MetricRegistry) createSummary(def *MetricDefinition) prometheus.Collector {
	opts := prometheus.SummaryOpts{
		Namespace: def.Namespace,
		Subsystem: def.Subsystem,
		Name:      def.Name,
		Help:      def.Help,
	}

	if len(def.Objectives) > 0 {
		opts.Objectives = def.Objectives
	}

	if len(def.Labels) > 0 {
		return prometheus.NewSummaryVec(opts, def.Labels)
	}
	return prometheus.NewSummary(opts)
}

// RegisterBuiltinMetrics 注册内置指标
func (r *MetricRegistry) RegisterBuiltinMetrics() error {
	builtinMetrics := []*MetricDefinition{
		// 系统指标
		{
			Name:      "cpu_usage_percent",
			Help:      "CPU usage percentage",
			Type:      MetricTypeGauge,
			Labels:    []string{"cpu"},
			Namespace: "laojun",
			Subsystem: "system",
		},
		{
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes",
			Type:      MetricTypeGauge,
			Labels:    []string{"type"},
			Namespace: "laojun",
			Subsystem: "system",
		},
		{
			Name:      "disk_usage_bytes",
			Help:      "Disk usage in bytes",
			Type:      MetricTypeGauge,
			Labels:    []string{"device", "mountpoint"},
			Namespace: "laojun",
			Subsystem: "system",
		},
		{
			Name:      "network_bytes_total",
			Help:      "Network bytes transferred",
			Type:      MetricTypeCounter,
			Labels:    []string{"device", "direction"},
			Namespace: "laojun",
			Subsystem: "system",
		},

		// HTTP指标
		{
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
			Type:      MetricTypeCounter,
			Labels:    []string{"method", "path", "status"},
			Namespace: "laojun",
			Subsystem: "http",
		},
		{
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Type:      MetricTypeHistogram,
			Labels:    []string{"method", "path"},
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10},
			Namespace: "laojun",
			Subsystem: "http",
		},

		// 应用指标
		{
			Name:      "application_info",
			Help:      "Application information",
			Type:      MetricTypeGauge,
			Labels:    []string{"version", "service"},
			Namespace: "laojun",
			Subsystem: "app",
		},
		{
			Name:      "application_uptime_seconds",
			Help:      "Application uptime in seconds",
			Type:      MetricTypeCounter,
			Namespace: "laojun",
			Subsystem: "app",
		},

		// 数据库指标
		{
			Name:      "database_connections_active",
			Help:      "Number of active database connections",
			Type:      MetricTypeGauge,
			Labels:    []string{"database"},
			Namespace: "laojun",
			Subsystem: "db",
		},
		{
			Name:      "database_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Type:      MetricTypeHistogram,
			Labels:    []string{"database", "operation"},
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 2.5, 5},
			Namespace: "laojun",
			Subsystem: "db",
		},
	}

	for _, metric := range builtinMetrics {
		if err := r.RegisterMetric(metric); err != nil {
			r.logger.Error("Failed to register builtin metric",
				zap.String("name", metric.Name),
				zap.Error(err))
			return err
		}
	}

	return nil
}

// UpdateMetricValue 更新指标值
func (r *MetricRegistry) UpdateMetricValue(name string, value float64, labels prometheus.Labels) error {
	r.mu.RLock()
	def, exists := r.definitions[name]
	if !exists {
		r.mu.RUnlock()
		return ErrMetricNotFound
	}
	r.mu.RUnlock()

	switch def.Type {
	case MetricTypeCounter:
		counter, err := r.GetCounter(name)
		if err != nil {
			return err
		}
		counter.With(labels).Add(value)

	case MetricTypeGauge:
		gauge, err := r.GetGauge(name)
		if err != nil {
			return err
		}
		gauge.With(labels).Set(value)

	case MetricTypeHistogram:
		histogram, err := r.GetHistogram(name)
		if err != nil {
			return err
		}
		histogram.With(labels).Observe(value)

	case MetricTypeSummary:
		summary, err := r.GetSummary(name)
		if err != nil {
			return err
		}
		summary.With(labels).Observe(value)

	default:
		return ErrUnsupportedMetricType
	}

	return nil
}

// GetMetricValue 获取指标当前值（仅适用于Gauge）
func (r *MetricRegistry) GetMetricValue(name string, labels prometheus.Labels) (float64, error) {
	gauge, err := r.GetGauge(name)
	if err != nil {
		return 0, err
	}

	metric := &dto.Metric{}
	if err := gauge.With(labels).Write(metric); err != nil {
		return 0, err
	}

	return metric.GetGauge().GetValue(), nil
}

// Reset 重置所有指标
func (r *MetricRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, collector := range r.metrics {
		r.registry.Unregister(collector)
		delete(r.metrics, name)
		delete(r.definitions, name)
	}

	r.logger.Info("All metrics have been reset")
}