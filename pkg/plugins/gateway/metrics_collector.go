package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestMetrics:  make(map[string]*RequestMetrics),
		pluginMetrics:   make(map[string]*PluginMetrics),
		systemMetrics:   &SystemMetrics{},
		errorMetrics:    make(map[string]*ErrorMetrics),
		customMetrics:   make(map[string]interface{}),
		histograms:      make(map[string]*Histogram),
		counters:        make(map[string]*Counter),
		gauges:          make(map[string]*Gauge),
		timers:          make(map[string]*Timer),
		startTime:       time.Now(),
	}
}

// RecordRequest 记录请求指标
func (mc *MetricsCollector) RecordRequest(pluginID, method, path string, statusCode int, duration time.Duration, size int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", pluginID, method, path)
	
	if _, exists := mc.requestMetrics[key]; !exists {
		mc.requestMetrics[key] = &RequestMetrics{
			PluginID:     pluginID,
			Method:       method,
			Path:         path,
			Count:        0,
			TotalTime:    0,
			MinTime:      duration,
			MaxTime:      duration,
			TotalSize:    0,
			StatusCodes:  make(map[int]int),
			LastRequest:  time.Now(),
		}
	}

	metrics := mc.requestMetrics[key]
	metrics.Count++
	metrics.TotalTime += duration
	metrics.TotalSize += size
	metrics.LastRequest = time.Now()

	if duration < metrics.MinTime {
		metrics.MinTime = duration
	}
	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}

	metrics.StatusCodes[statusCode]++

	// 更新直方图指标
	mc.recordHistogram(fmt.Sprintf("request_duration_%s", pluginID), float64(duration.Milliseconds()))
	mc.recordHistogram("request_duration_total", float64(duration.Milliseconds()))
	
	// 更新计数指标
	mc.incrementCounter(fmt.Sprintf("requests_total_%s", pluginID))
	mc.incrementCounter("requests_total")
	mc.incrementCounter(fmt.Sprintf("requests_%d", statusCode))
}

// RecordPlugin 记录插件指标
func (mc *MetricsCollector) RecordPlugin(pluginID string, status string, cpuUsage, memoryUsage float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.pluginMetrics[pluginID]; !exists {
		mc.pluginMetrics[pluginID] = &PluginMetrics{
			PluginID:      pluginID,
			Status:        status,
			StartTime:     time.Now(),
			RequestCount:  0,
			ErrorCount:    0,
			CPUUsage:      cpuUsage,
			MemoryUsage:   memoryUsage,
			LastUpdate:    time.Now(),
		}
	}

	metrics := mc.pluginMetrics[pluginID]
	metrics.Status = status
	metrics.CPUUsage = cpuUsage
	metrics.MemoryUsage = memoryUsage
	metrics.LastUpdate = time.Now()

	// 更新仪表指标
	mc.setGauge(fmt.Sprintf("plugin_cpu_%s", pluginID), cpuUsage)
	mc.setGauge(fmt.Sprintf("plugin_memory_%s", pluginID), memoryUsage)
}

// RecordError 记录错误指标
func (mc *MetricsCollector) RecordError(pluginID, errorType, errorMessage string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := fmt.Sprintf("%s:%s", pluginID, errorType)
	
	if _, exists := mc.errorMetrics[key]; !exists {
		mc.errorMetrics[key] = &ErrorMetrics{
			PluginID:     pluginID,
			ErrorType:    errorType,
			Count:        0,
			LastError:    time.Now(),
			LastMessage:  errorMessage,
		}
	}

	metrics := mc.errorMetrics[key]
	metrics.Count++
	metrics.LastError = time.Now()
	metrics.LastMessage = errorMessage

	// 更新插件错误计数
	if pluginMetrics, exists := mc.pluginMetrics[pluginID]; exists {
		pluginMetrics.ErrorCount++
	}

	// 更新计数指标
	mc.incrementCounter(fmt.Sprintf("errors_total_%s", pluginID))
	mc.incrementCounter(fmt.Sprintf("errors_%s", errorType))
	mc.incrementCounter("errors_total")
}

// RecordSystemMetrics 记录系统指标
func (mc *MetricsCollector) RecordSystemMetrics(cpuUsage, memoryUsage, diskUsage float64, goroutines int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.CPUUsage = cpuUsage
	mc.systemMetrics.MemoryUsage = memoryUsage
	mc.systemMetrics.DiskUsage = diskUsage
	mc.systemMetrics.Goroutines = goroutines
	mc.systemMetrics.LastUpdate = time.Now()

	// 更新仪表指标
	mc.setGauge("system_cpu", cpuUsage)
	mc.setGauge("system_memory", memoryUsage)
	mc.setGauge("system_disk", diskUsage)
	mc.setGauge("system_goroutines", float64(goroutines))
}

// GetMetrics 获取所有指标
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	uptime := time.Since(mc.startTime)

	return map[string]interface{}{
		"uptime":         uptime.String(),
		"uptime_seconds": uptime.Seconds(),
		"requests":       mc.requestMetrics,
		"plugins":        mc.pluginMetrics,
		"system":         mc.systemMetrics,
		"errors":         mc.errorMetrics,
		"custom":         mc.customMetrics,
		"histograms":     mc.getHistogramStats(),
		"counters":       mc.getCounterStats(),
		"gauges":         mc.getGaugeStats(),
		"timers":         mc.getTimerStats(),
		"timestamp":      time.Now(),
	}
}

// GetPluginMetrics 获取特定插件的指标
func (mc *MetricsCollector) GetPluginMetrics(pluginID string) map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]interface{})

	// 插件基本指标
	if pluginMetrics, exists := mc.pluginMetrics[pluginID]; exists {
		result["plugin"] = pluginMetrics
	}

	// 请求指标
	requests := make(map[string]*RequestMetrics)
	for key, metrics := range mc.requestMetrics {
		if metrics.PluginID == pluginID {
			requests[key] = metrics
		}
	}
	result["requests"] = requests

	// 错误指标
	errors := make(map[string]*ErrorMetrics)
	for key, metrics := range mc.errorMetrics {
		if metrics.PluginID == pluginID {
			errors[key] = metrics
		}
	}
	result["errors"] = errors

	return result
}

// GetSummary 获取指标摘要
func (mc *MetricsCollector) GetSummary() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	totalRequests := 0
	totalErrors := 0
	avgResponseTime := float64(0)
	totalResponseTime := int64(0)

	for _, metrics := range mc.requestMetrics {
		totalRequests += metrics.Count
		totalResponseTime += int64(metrics.TotalTime)
	}

	for _, metrics := range mc.errorMetrics {
		totalErrors += metrics.Count
	}

	if totalRequests > 0 {
		avgResponseTime = float64(totalResponseTime) / float64(totalRequests) / float64(time.Millisecond)
	}

	activePlugins := 0
	for _, metrics := range mc.pluginMetrics {
		if metrics.Status == "running" {
			activePlugins++
		}
	}

	return map[string]interface{}{
		"total_requests":       totalRequests,
		"total_errors":         totalErrors,
		"error_rate":           float64(totalErrors) / float64(totalRequests) * 100,
		"avg_response_time_ms": avgResponseTime,
		"active_plugins":       activePlugins,
		"total_plugins":        len(mc.pluginMetrics),
		"uptime":               time.Since(mc.startTime).String(),
		"system":               mc.systemMetrics,
	}
}

// recordHistogram 记录直方图指标
func (mc *MetricsCollector) recordHistogram(name string, value float64) {		
	if _, exists := mc.histograms[name]; !exists {
		mc.histograms[name] = NewHistogram()
	}
	mc.histograms[name].Record(value)
}

// incrementCounter 增加计数指标
func (mc *MetricsCollector) incrementCounter(name string) {
	if _, exists := mc.counters[name]; !exists {
		mc.counters[name] = NewCounter()
	}
	mc.counters[name].Increment()
}

// setGauge 设置仪表盘指标
func (mc *MetricsCollector) setGauge(name string, value float64) {
	if _, exists := mc.gauges[name]; !exists {
		mc.gauges[name] = NewGauge()
	}
	mc.gauges[name].Set(value)
}

// startTimer 开始计时器
func (mc *MetricsCollector) StartTimer(name string) *Timer {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.timers[name]; !exists {
		mc.timers[name] = NewTimer()
	}
	mc.timers[name].Start()
	return mc.timers[name]
}

// getHistogramStats 获取直方图统计指标
func (mc *MetricsCollector) getHistogramStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for name, histogram := range mc.histograms {
		stats[name] = histogram.GetStats()
	}
	return stats
}

// getCounterStats 获取计数器统计指标
func (mc *MetricsCollector) getCounterStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for name, counter := range mc.counters {
		stats[name] = counter.GetValue()
	}
	return stats
}

// getGaugeStats 获取仪表盘统计指标
func (mc *MetricsCollector) getGaugeStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for name, gauge := range mc.gauges {
		stats[name] = gauge.GetValue()
	}
	return stats
}

// getTimerStats 获取计时器统计指标
func (mc *MetricsCollector) getTimerStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for name, timer := range mc.timers {
		stats[name] = timer.GetStats()
	}
	return stats
}

// ExportPrometheus 导出Prometheus格式指标
func (mc *MetricsCollector) ExportPrometheus() string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var output []string

	// 导出计数器统计指标	for name, counter := range mc.counters {
		output = append(output, fmt.Sprintf("# TYPE %s counter", name))
		output = append(output, fmt.Sprintf("%s %d", name, counter.GetValue()))
	}

	// 导出仪表盘统计指标	for name, gauge := range mc.gauges {
		output = append(output, fmt.Sprintf("# TYPE %s gauge", name))
		output = append(output, fmt.Sprintf("%s %f", name, gauge.GetValue()))
	}

	// 导出直方图统计指标	for name, histogram := range mc.histograms {
		stats := histogram.GetStats()
		output = append(output, fmt.Sprintf("# TYPE %s histogram", name))
		output = append(output, fmt.Sprintf("%s_count %d", name, stats["count"]))
		output = append(output, fmt.Sprintf("%s_sum %f", name, stats["sum"]))
		
		// 导出分位值指标		for _, p := range []float64{0.5, 0.9, 0.95, 0.99} {
			value := histogram.Quantile(p)
			output = append(output, fmt.Sprintf("%s{quantile=\"%g\"} %f", name, p, value))
		}
	}

	return fmt.Sprintf("%s\n", output)
}

// ServeHTTP 提供HTTP指标端点
func (mc *MetricsCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/metrics":
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(mc.ExportPrometheus()))
	case "/metrics/json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mc.GetMetrics())
	case "/metrics/summary":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mc.GetSummary())
	default:
		http.NotFound(w, r)
	}
}

// Histogram 直方图实现
type Histogram struct {
	values []float64
	mu     sync.RWMutex
}

func NewHistogram() *Histogram {
	return &Histogram{
		values: make([]float64, 0),
	}
}

func (h *Histogram) Record(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.values = append(h.values, value)
}

func (h *Histogram) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.values) == 0 {
		return map[string]interface{}{
			"count": 0,
			"sum":   0,
			"min":   0,
			"max":   0,
			"mean":  0,
		}
	}

	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	sum := 0.0
	for _, v := range h.values {
		sum += v
	}

	return map[string]interface{}{
		"count": len(h.values),
		"sum":   sum,
		"min":   sorted[0],
		"max":   sorted[len(sorted)-1],
		"mean":  sum / float64(len(h.values)),
	}
}

func (h *Histogram) Quantile(p float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.values) == 0 {
		return 0
	}

	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}

// Counter 计数器实现
type Counter struct {
	value int64
	mu    sync.RWMutex
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *Counter) Add(delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += delta
}

func (c *Counter) GetValue() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Gauge 仪表盘实现
type Gauge struct {
	value float64
	mu    sync.RWMutex
}

func NewGauge() *Gauge {
	return &Gauge{}
}

func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

func (g *Gauge) Add(delta float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += delta
}

func (g *Gauge) GetValue() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// Timer 计时器实现
type Timer struct {
	startTime time.Time
	durations []time.Duration
	mu        sync.RWMutex
}

func NewTimer() *Timer {
	return &Timer{
		durations: make([]time.Duration, 0),
	}
}

func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.startTime = time.Now()
}

func (t *Timer) Stop() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	duration := time.Since(t.startTime)
	t.durations = append(t.durations, duration)
	return duration
}

func (t *Timer) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.durations) == 0 {
		return map[string]interface{}{
			"count": 0,
			"total": 0,
			"mean":  0,
		}
	}

	total := time.Duration(0)
	for _, d := range t.durations {
		total += d
	}

	return map[string]interface{}{
		"count": len(t.durations),
		"total": total.String(),
		"mean":  (total / time.Duration(len(t.durations))).String(),
	}
}

// SetCustomMetric 设置自定义指标值
func (mc *MetricsCollector) SetCustomMetric(name string, value interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.customMetrics[name] = value
}

// GetCustomMetric 获取自定义指标值
func (mc *MetricsCollector) GetCustomMetric(name string) interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.customMetrics[name]
}

// Reset 重置所有指标值
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requestMetrics = make(map[string]*RequestMetrics)
	mc.pluginMetrics = make(map[string]*PluginMetrics)
	mc.errorMetrics = make(map[string]*ErrorMetrics)
	mc.customMetrics = make(map[string]interface{})
	mc.histograms = make(map[string]*Histogram)
	mc.counters = make(map[string]*Counter)
	mc.gauges = make(map[string]*Gauge)
	mc.timers = make(map[string]*Timer)
	mc.startTime = time.Now()
}
