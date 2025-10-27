package exporters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HTTPExporter HTTP导出器
type HTTPExporter struct {
	name      string
	config    ExporterConfig
	logger    *zap.Logger
	running   bool
	healthy   bool
	ready     bool
	mu        sync.RWMutex
	stats     ExporterStats
	startTime time.Time
	client    *http.Client
	endpoint  string
	headers   map[string]string
	timeout   time.Duration
	retries   int
}

// NewHTTPExporter 创建HTTP导出器
func NewHTTPExporter(config ExporterConfig, logger *zap.Logger) (*HTTPExporter, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	// 获取配置
	configMap := config.GetConfig()
	endpoint, ok := configMap["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, fmt.Errorf("endpoint is required for HTTP exporter")
	}
	
	timeout := 30 * time.Second
	if t, ok := configMap["timeout"].(time.Duration); ok {
		timeout = t
	} else if t, ok := configMap["timeout"].(string); ok {
		if parsed, err := time.ParseDuration(t); err == nil {
			timeout = parsed
		}
	}
	
	retries := 3
	if r, ok := configMap["retries"].(int); ok {
		retries = r
	}
	
	headers := make(map[string]string)
	if h, ok := configMap["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}
	
	// 设置默认头部
	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json"
	}
	
	exporter := &HTTPExporter{
		name:      config.GetName(),
		config:    config,
		logger:    logger,
		endpoint:  endpoint,
		headers:   headers,
		timeout:   timeout,
		retries:   retries,
		ready:     true,
		startTime: time.Now(),
		stats: ExporterStats{
			StartTime: time.Now(),
		},
		client: &http.Client{
			Timeout: timeout,
		},
	}
	
	return exporter, nil
}

// Name 返回导出器名称
func (e *HTTPExporter) Name() string {
	return e.name
}

// Start 启动导出器
func (e *HTTPExporter) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		return nil
	}
	
	if !e.config.IsEnabled() {
		e.logger.Info("HTTP exporter is disabled", zap.String("name", e.name))
		return nil
	}
	
	e.logger.Info("Starting HTTP exporter", 
		zap.String("name", e.name),
		zap.String("endpoint", e.endpoint))
	
	// 测试连接
	if err := e.testConnection(); err != nil {
		e.logger.Warn("Failed to test HTTP connection", 
			zap.String("endpoint", e.endpoint),
			zap.Error(err))
		// 不阻止启动，可能是临时网络问题
	}
	
	e.running = true
	e.healthy = true
	
	e.logger.Info("HTTP exporter started successfully", 
		zap.String("name", e.name),
		zap.String("endpoint", e.endpoint))
	
	return nil
}

// Stop 停止导出器
func (e *HTTPExporter) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.running {
		return nil
	}
	
	e.logger.Info("Stopping HTTP exporter", zap.String("name", e.name))
	
	e.running = false
	e.healthy = false
	
	e.logger.Info("HTTP exporter stopped", zap.String("name", e.name))
	
	return nil
}

// IsHealthy 检查导出器健康状态
func (e *HTTPExporter) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.healthy && e.running
}

// IsReady 检查导出器就绪状态
func (e *HTTPExporter) IsReady() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ready
}

// Export 导出数据
func (e *HTTPExporter) Export(data interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("HTTP exporter is not healthy")
	}
	
	start := time.Now()
	
	// 序列化数据
	payload, err := e.serializeData(data)
	if err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to serialize data: %w", err)
	}
	
	// 发送HTTP请求
	if err := e.sendRequest(payload); err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// ExportBatch 批量导出数据
func (e *HTTPExporter) ExportBatch(data []interface{}) error {
	if !e.IsHealthy() {
		return fmt.Errorf("HTTP exporter is not healthy")
	}
	
	start := time.Now()
	
	// 构建批量数据
	batchData := map[string]interface{}{
		"exporter":  e.name,
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"batch_size": len(data),
		"items":     data,
	}
	
	// 序列化批量数据
	payload, err := json.Marshal(batchData)
	if err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to serialize batch data: %w", err)
	}
	
	// 发送HTTP请求
	if err := e.sendRequest(payload); err != nil {
		e.updateStats(false, time.Since(start), err)
		return fmt.Errorf("failed to send batch HTTP request: %w", err)
	}
	
	e.logger.Debug("Batch exported via HTTP",
		zap.String("endpoint", e.endpoint),
		zap.Int("batch_size", len(data)),
		zap.Duration("latency", time.Since(start)))
	
	e.updateStats(true, time.Since(start), nil)
	return nil
}

// GetBatchSize 获取批量大小
func (e *HTTPExporter) GetBatchSize() int {
	config := e.config.GetConfig()
	if batchSize, ok := config["batch_size"].(int); ok {
		return batchSize
	}
	return 50 // 默认批量大小
}

// SetBatchSize 设置批量大小
func (e *HTTPExporter) SetBatchSize(size int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.SetConfig("batch_size", size)
}

// GetFlushInterval 获取刷新间隔
func (e *HTTPExporter) GetFlushInterval() time.Duration {
	config := e.config.GetConfig()
	if interval, ok := config["flush_interval"].(time.Duration); ok {
		return interval
	}
	return 10 * time.Second // 默认刷新间隔
}

// SetFlushInterval 设置刷新间隔
func (e *HTTPExporter) SetFlushInterval(interval time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.SetConfig("flush_interval", interval)
}

// GetStats 获取导出器统计信息
func (e *HTTPExporter) GetStats() ExporterStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

// testConnection 测试HTTP连接
func (e *HTTPExporter) testConnection() error {
	req, err := http.NewRequest("HEAD", e.endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}
	
	// 设置头部
	for k, v := range e.headers {
		req.Header.Set(k, v)
	}
	
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("test request failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// sendRequest 发送HTTP请求
func (e *HTTPExporter) sendRequest(payload []byte) error {
	var lastErr error
	
	for attempt := 0; attempt <= e.retries; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
			
			e.logger.Debug("Retrying HTTP request",
				zap.String("endpoint", e.endpoint),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff))
		}
		
		req, err := http.NewRequest("POST", e.endpoint, bytes.NewBuffer(payload))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		
		// 设置头部
		for k, v := range e.headers {
			req.Header.Set(k, v)
		}
		
		resp, err := e.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send request: %w", err)
			continue
		}
		
		resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil // 成功
		}
		
		lastErr = fmt.Errorf("request failed with status: %d", resp.StatusCode)
		
		// 如果是客户端错误（4xx），不重试
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
	}
	
	return lastErr
}

// serializeData 序列化数据
func (e *HTTPExporter) serializeData(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case *MetricData:
		return json.Marshal(map[string]interface{}{
			"type":      "metric",
			"exporter":  e.name,
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"data":      v,
		})
		
	case *EventData:
		return json.Marshal(map[string]interface{}{
			"type":      "event",
			"exporter":  e.name,
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"data":      v,
		})
		
	case *TraceData:
		return json.Marshal(map[string]interface{}{
			"type":      "trace",
			"exporter":  e.name,
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"data":      v,
		})
		
	default:
		return json.Marshal(map[string]interface{}{
			"type":      "generic",
			"exporter":  e.name,
			"timestamp": time.Now().Format(time.RFC3339Nano),
			"data":      data,
		})
	}
}

// updateStats 更新统计信息
func (e *HTTPExporter) updateStats(success bool, latency time.Duration, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.stats.RequestCount++
	e.stats.LastRequestTime = time.Now()
	
	if !success {
		e.stats.ErrorCount++
		if err != nil {
			e.stats.LastError = err.Error()
		}
	}
}