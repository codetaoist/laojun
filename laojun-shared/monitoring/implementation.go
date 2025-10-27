package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MonitorImpl implements the Monitor interface.
type MonitorImpl struct {
	config  Config
	mu      sync.RWMutex
	metrics map[string]*metricImpl
	timers  map[string]*timerImpl
	started bool
}

// metricImpl implements the Metric interface.
type metricImpl struct {
	name      string
	metricType MetricType
	value     interface{}
	labels    map[string]string
	timestamp time.Time
	mu        sync.RWMutex
}

// timerImpl implements the Timer interface.
type timerImpl struct {
	name      string
	labels    map[string]string
	startTime time.Time
	endTime   *time.Time
	mu        sync.RWMutex
}

// NewMonitor creates a new monitor instance.
func NewMonitor(config Config) (*MonitorImpl, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	impl := &MonitorImpl{
		config:  config,
		metrics: make(map[string]*metricImpl),
		timers:  make(map[string]*timerImpl),
		started: true,
	}
	
	return impl, nil
}

// IncrementCounter increments a counter metric by 1.
func (m *MonitorImpl) IncrementCounter(ctx context.Context, name string, labels map[string]string) error {
	return m.AddCounter(ctx, name, 1, labels)
}

// AddCounter adds a value to a counter metric.
func (m *MonitorImpl) AddCounter(ctx context.Context, name string, value float64, labels map[string]string) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	if name == "" {
		return ErrInvalidMetricName
	}
	
	if value < 0 {
		return ErrInvalidMetricValue
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	metric, exists := m.metrics[key]
	
	if !exists {
		metric = &metricImpl{
			name:       name,
			metricType: MetricTypeCounter,
			value:      value,
			labels:     m.mergeLabels(labels),
			timestamp:  time.Now(),
		}
		m.metrics[key] = metric
	} else {
		metric.mu.Lock()
		if currentValue, ok := metric.value.(float64); ok {
			metric.value = currentValue + value
		} else {
			metric.value = value
		}
		metric.timestamp = time.Now()
		metric.mu.Unlock()
	}
	
	return nil
}

// SetGauge sets a gauge metric value.
func (m *MonitorImpl) SetGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	if name == "" {
		return ErrInvalidMetricName
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	metric, exists := m.metrics[key]
	
	if !exists {
		metric = &metricImpl{
			name:       name,
			metricType: MetricTypeGauge,
			value:      value,
			labels:     m.mergeLabels(labels),
			timestamp:  time.Now(),
		}
		m.metrics[key] = metric
	} else {
		metric.mu.Lock()
		metric.value = value
		metric.timestamp = time.Now()
		metric.mu.Unlock()
	}
	
	return nil
}

// AddGauge adds a value to a gauge metric.
func (m *MonitorImpl) AddGauge(ctx context.Context, name string, value float64, labels map[string]string) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	if name == "" {
		return ErrInvalidMetricName
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	metric, exists := m.metrics[key]
	
	if !exists {
		metric = &metricImpl{
			name:       name,
			metricType: MetricTypeGauge,
			value:      value,
			labels:     m.mergeLabels(labels),
			timestamp:  time.Now(),
		}
		m.metrics[key] = metric
	} else {
		metric.mu.Lock()
		if currentValue, ok := metric.value.(float64); ok {
			metric.value = currentValue + value
		} else {
			metric.value = value
		}
		metric.timestamp = time.Now()
		metric.mu.Unlock()
	}
	
	return nil
}

// RecordHistogram records a histogram metric value.
func (m *MonitorImpl) RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	if name == "" {
		return ErrInvalidMetricName
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	metric := &metricImpl{
		name:       name,
		metricType: MetricTypeHistogram,
		value:      value,
		labels:     m.mergeLabels(labels),
		timestamp:  time.Now(),
	}
	m.metrics[key+"_"+fmt.Sprintf("%d", time.Now().UnixNano())] = metric
	
	return nil
}

// RecordSummary records a summary metric value.
func (m *MonitorImpl) RecordSummary(ctx context.Context, name string, value float64, labels map[string]string) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	if name == "" {
		return ErrInvalidMetricName
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(name, labels)
	metric := &metricImpl{
		name:       name,
		metricType: MetricTypeSummary,
		value:      value,
		labels:     m.mergeLabels(labels),
		timestamp:  time.Now(),
	}
	m.metrics[key+"_"+fmt.Sprintf("%d", time.Now().UnixNano())] = metric
	
	return nil
}

// StartTimer starts a timer and returns a Timer instance.
func (m *MonitorImpl) StartTimer(ctx context.Context, name string, labels map[string]string) Timer {
	timer := &timerImpl{
		name:      name,
		labels:    m.mergeLabels(labels),
		startTime: time.Now(),
	}
	
	m.mu.Lock()
	key := m.buildMetricKey(name, labels)
	m.timers[key] = timer
	m.mu.Unlock()
	
	return timer
}

// RecordDuration records a duration metric.
func (m *MonitorImpl) RecordDuration(ctx context.Context, name string, duration time.Duration, labels map[string]string) error {
	return m.RecordHistogram(ctx, name+"_duration_seconds", duration.Seconds(), labels)
}

// RegisterMetric registers a custom metric.
func (m *MonitorImpl) RegisterMetric(ctx context.Context, metric Metric) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	if metric.Name() == "" {
		return ErrInvalidMetricName
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := m.buildMetricKey(metric.Name(), metric.Labels())
	if _, exists := m.metrics[key]; exists {
		return ErrMetricAlreadyExists
	}
	
	impl := &metricImpl{
		name:       metric.Name(),
		metricType: metric.Type(),
		value:      metric.Value(),
		labels:     metric.Labels(),
		timestamp:  metric.Timestamp(),
	}
	m.metrics[key] = impl
	
	return nil
}

// UnregisterMetric unregisters a metric.
func (m *MonitorImpl) UnregisterMetric(ctx context.Context, name string) error {
	if !m.started {
		return ErrMonitorNotInitialized
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	found := false
	for key := range m.metrics {
		if strings.HasPrefix(key, name+"_") || key == name {
			delete(m.metrics, key)
			found = true
		}
	}
	
	if !found {
		return ErrMetricNotFound
	}
	
	return nil
}

// GetMetrics returns all registered metrics.
func (m *MonitorImpl) GetMetrics(ctx context.Context) ([]Metric, error) {
	if !m.started {
		return nil, ErrMonitorNotInitialized
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := make([]Metric, 0, len(m.metrics))
	for _, metric := range m.metrics {
		metrics = append(metrics, metric)
	}
	
	return metrics, nil
}

// Export exports metrics in the specified format.
func (m *MonitorImpl) Export(ctx context.Context, format ExportFormat) ([]byte, error) {
	if !m.started {
		return nil, ErrMonitorNotInitialized
	}
	
	metrics, err := m.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}
	
	switch format {
	case ExportFormatJSON:
		return m.exportJSON(metrics)
	case ExportFormatPrometheus:
		return m.exportPrometheus(metrics)
	case ExportFormatInfluxDB:
		return m.exportInfluxDB(metrics)
	default:
		return nil, ErrExportFormatNotSupported
	}
}

// IsHealthy returns the health status of the monitor.
func (m *MonitorImpl) IsHealthy(ctx context.Context) bool {
	return m.started
}

// Close releases resources and stops the monitor.
func (m *MonitorImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.started = false
	m.metrics = make(map[string]*metricImpl)
	m.timers = make(map[string]*timerImpl)
	
	return nil
}

// Helper methods

func (m *MonitorImpl) buildMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(parts)
	
	return fmt.Sprintf("%s{%s}", name, strings.Join(parts, ","))
}

func (m *MonitorImpl) mergeLabels(labels map[string]string) map[string]string {
	merged := make(map[string]string)
	
	// Add default labels
	for k, v := range m.config.DefaultLabels {
		merged[k] = v
	}
	
	// Add service labels
	merged["service"] = m.config.ServiceName
	merged["version"] = m.config.ServiceVersion
	
	// Add provided labels (override defaults)
	for k, v := range labels {
		merged[k] = v
	}
	
	return merged
}

func (m *MonitorImpl) exportJSON(metrics []Metric) ([]byte, error) {
	data := make([]map[string]interface{}, len(metrics))
	for i, metric := range metrics {
		data[i] = map[string]interface{}{
			"name":      metric.Name(),
			"type":      metric.Type(),
			"value":     metric.Value(),
			"labels":    metric.Labels(),
			"timestamp": metric.Timestamp(),
		}
	}
	return json.Marshal(data)
}

func (m *MonitorImpl) exportPrometheus(metrics []Metric) ([]byte, error) {
	var lines []string
	
	for _, metric := range metrics {
		var metricType string
		switch metric.Type() {
		case MetricTypeCounter:
			metricType = "counter"
		case MetricTypeGauge:
			metricType = "gauge"
		case MetricTypeHistogram:
			metricType = "histogram"
		case MetricTypeSummary:
			metricType = "summary"
		}
		
		lines = append(lines, fmt.Sprintf("# TYPE %s %s", metric.Name(), metricType))
		
		labelStr := ""
		if len(metric.Labels()) > 0 {
			var labelParts []string
			for k, v := range metric.Labels() {
				labelParts = append(labelParts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(labelParts)
			labelStr = fmt.Sprintf("{%s}", strings.Join(labelParts, ","))
		}
		
		lines = append(lines, fmt.Sprintf("%s%s %v %d", 
			metric.Name(), labelStr, metric.Value(), metric.Timestamp().Unix()))
	}
	
	return []byte(strings.Join(lines, "\n")), nil
}

func (m *MonitorImpl) exportInfluxDB(metrics []Metric) ([]byte, error) {
	var lines []string
	
	for _, metric := range metrics {
		var tagParts []string
		for k, v := range metric.Labels() {
			tagParts = append(tagParts, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(tagParts)
		
		tagStr := ""
		if len(tagParts) > 0 {
			tagStr = "," + strings.Join(tagParts, ",")
		}
		
		lines = append(lines, fmt.Sprintf("%s%s value=%v %d", 
			metric.Name(), tagStr, metric.Value(), metric.Timestamp().UnixNano()))
	}
	
	return []byte(strings.Join(lines, "\n")), nil
}

// Metric implementation

func (m *metricImpl) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

func (m *metricImpl) Type() MetricType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metricType
}

func (m *metricImpl) Value() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

func (m *metricImpl) Labels() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	labels := make(map[string]string)
	for k, v := range m.labels {
		labels[k] = v
	}
	return labels
}

func (m *metricImpl) Timestamp() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.timestamp
}

// Timer implementation

func (t *timerImpl) Stop() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.endTime != nil {
		return t.endTime.Sub(t.startTime)
	}
	
	now := time.Now()
	t.endTime = &now
	return now.Sub(t.startTime)
}

func (t *timerImpl) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.endTime != nil {
		return t.endTime.Sub(t.startTime)
	}
	
	return time.Since(t.startTime)
}
