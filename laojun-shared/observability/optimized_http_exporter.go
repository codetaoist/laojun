package observability

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// OptimizedHTTPExporter 优化的HTTP导出器
type OptimizedHTTPExporter struct {
	name     string
	endpoint string
	client   *http.Client
	config   *ExportConfig
	
	// 连接池统计
	poolStats ConnectionPoolStats
	statsMu   sync.RWMutex
	
	// 性能统计
	requestCount    int64
	successCount    int64
	errorCount      int64
	totalLatency    int64
	lastRequestTime time.Time
	
	// 连接管理
	transport *OptimizedTransport
}

// OptimizedTransport 优化的传输层
type OptimizedTransport struct {
	*http.Transport
	activeConnections int32
	totalConnections  int32
	maxConnections    int32
	connMu           sync.RWMutex
	connStats        map[string]*ConnectionInfo
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	CreatedAt    time.Time
	LastUsed     time.Time
	RequestCount int64
	IsIdle       bool
}

// OptimizedExporterFactory 优化导出器工厂
type OptimizedExporterFactory struct{}

// CreateOptimizedExporter 创建优化导出器
func (f *OptimizedExporterFactory) CreateOptimizedExporter(exporterType, name, endpoint string, config *ExportConfig) (OptimizedExporter, error) {
	switch exporterType {
	case "http", "https":
		return NewOptimizedHTTPExporter(name, endpoint, config)
	case "console":
		return NewOptimizedConsoleExporter(name, config)
	case "file":
		return NewOptimizedFileExporter(name, endpoint, config)
	default:
		return nil, fmt.Errorf("unsupported optimized exporter type: %s", exporterType)
	}
}

// NewOptimizedHTTPExporter 创建新的优化HTTP导出器
func NewOptimizedHTTPExporter(name, endpoint string, config *ExportConfig) (*OptimizedHTTPExporter, error) {
	transport := &OptimizedTransport{
		Transport: &http.Transport{
			// 连接池优化配置
			MaxIdleConns:        100,              // 最大空闲连接数
			MaxIdleConnsPerHost: 20,               // 每个主机最大空闲连接数
			MaxConnsPerHost:     50,               // 每个主机最大连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时
			
			// TCP连接优化
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second, // 连接超时
				KeepAlive: 30 * time.Second, // TCP Keep-Alive
			}).DialContext,
			
			// TLS优化
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12,
			},
			
			// HTTP/2支持
			ForceAttemptHTTP2: true,
			
			// 响应头超时
			ResponseHeaderTimeout: 30 * time.Second,
			
			// 期望继续超时
			ExpectContinueTimeout: 1 * time.Second,
		},
		maxConnections: 50,
		connStats:      make(map[string]*ConnectionInfo),
	}
	
	// 设置连接状态跟踪
	transport.setupConnectionTracking()

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 限制重定向次数
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	exporter := &OptimizedHTTPExporter{
		name:      name,
		endpoint:  endpoint,
		client:    client,
		config:    config,
		transport: transport,
	}

	return exporter, nil
}

// setupConnectionTracking 设置连接跟踪
func (t *OptimizedTransport) setupConnectionTracking() {
	originalDial := t.Transport.DialContext
	
	t.Transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := originalDial(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		
		// 跟踪新连接
		atomic.AddInt32(&t.activeConnections, 1)
		atomic.AddInt32(&t.totalConnections, 1)
		
		t.connMu.Lock()
		t.connStats[addr] = &ConnectionInfo{
			CreatedAt:    time.Now(),
			LastUsed:     time.Now(),
			RequestCount: 0,
			IsIdle:       false,
		}
		t.connMu.Unlock()
		
		// 包装连接以跟踪关闭
		return &trackedConnection{
			Conn:      conn,
			transport: t,
			addr:      addr,
		}, nil
	}
}

// trackedConnection 被跟踪的连接
type trackedConnection struct {
	net.Conn
	transport *OptimizedTransport
	addr      string
	closed    int32
}

// Close 关闭连接
func (tc *trackedConnection) Close() error {
	if atomic.CompareAndSwapInt32(&tc.closed, 0, 1) {
		atomic.AddInt32(&tc.transport.activeConnections, -1)
		
		tc.transport.connMu.Lock()
		delete(tc.transport.connStats, tc.addr)
		tc.transport.connMu.Unlock()
	}
	return tc.Conn.Close()
}

// Export 导出数据
func (e *OptimizedHTTPExporter) Export(ctx context.Context, items []BatchItem) error {
	startTime := time.Now()
	atomic.AddInt64(&e.requestCount, 1)
	
	defer func() {
		latency := time.Since(startTime).Nanoseconds()
		atomic.AddInt64(&e.totalLatency, latency)
		e.lastRequestTime = time.Now()
	}()

	// 序列化数据
	data, err := json.Marshal(items)
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	return e.sendRequest(ctx, data)
}

// ExportCached 导出缓存数据
func (e *OptimizedHTTPExporter) ExportCached(ctx context.Context, cachedData []byte) error {
	startTime := time.Now()
	atomic.AddInt64(&e.requestCount, 1)
	
	defer func() {
		latency := time.Since(startTime).Nanoseconds()
		atomic.AddInt64(&e.totalLatency, latency)
		e.lastRequestTime = time.Now()
	}()

	return e.sendRequest(ctx, cachedData)
}

// sendRequest 发送HTTP请求
func (e *OptimizedHTTPExporter) sendRequest(ctx context.Context, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint, bytes.NewReader(data))
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Laojun-Optimized-Exporter/1.0")
	
	// 添加自定义头部
	for key, value := range e.config.Headers {
		req.Header.Set(key, value)
	}
	
	// 启用压缩
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	
	// 连接复用
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := e.client.Do(req)
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		atomic.AddInt64(&e.errorCount, 1)
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// 读取响应体（即使不使用，也要读取以便连接复用）
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	atomic.AddInt64(&e.successCount, 1)
	return nil
}

// Name 返回导出器名称
func (e *OptimizedHTTPExporter) Name() string {
	return e.name
}

// Close 关闭导出器
func (e *OptimizedHTTPExporter) Close() error {
	// 关闭空闲连接
	e.client.CloseIdleConnections()
	
	// 等待活跃连接完成
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout.C:
			// 超时，强制关闭
			return nil
		case <-ticker.C:
			if atomic.LoadInt32(&e.transport.activeConnections) == 0 {
				return nil
			}
		}
	}
}

// GetConnectionPoolStats 获取连接池统计
func (e *OptimizedHTTPExporter) GetConnectionPoolStats() ConnectionPoolStats {
	e.statsMu.RLock()
	defer e.statsMu.RUnlock()
	
	activeConns := atomic.LoadInt32(&e.transport.activeConnections)
	totalConns := atomic.LoadInt32(&e.transport.totalConnections)
	maxConns := atomic.LoadInt32(&e.transport.maxConnections)
	
	// 计算空闲连接数（估算）
	idleConns := int(maxConns - activeConns)
	if idleConns < 0 {
		idleConns = 0
	}
	
	return ConnectionPoolStats{
		ActiveConnections: int(activeConns),
		IdleConnections:   idleConns,
		TotalConnections:  int(totalConns),
		MaxConnections:    int(maxConns),
	}
}

// GetPerformanceStats 获取性能统计
func (e *OptimizedHTTPExporter) GetPerformanceStats() map[string]interface{} {
	requestCount := atomic.LoadInt64(&e.requestCount)
	successCount := atomic.LoadInt64(&e.successCount)
	errorCount := atomic.LoadInt64(&e.errorCount)
	totalLatency := atomic.LoadInt64(&e.totalLatency)
	
	avgLatency := time.Duration(0)
	if requestCount > 0 {
		avgLatency = time.Duration(totalLatency / requestCount)
	}
	
	successRate := 0.0
	if requestCount > 0 {
		successRate = float64(successCount) / float64(requestCount)
	}
	
	return map[string]interface{}{
		"request_count":     requestCount,
		"success_count":     successCount,
		"error_count":       errorCount,
		"success_rate":      successRate,
		"average_latency":   avgLatency,
		"last_request_time": e.lastRequestTime,
	}
}

// HealthCheck 健康检查
func (e *OptimizedHTTPExporter) HealthCheck(ctx context.Context) error {
	// 创建一个简单的健康检查请求
	req, err := http.NewRequestWithContext(ctx, "HEAD", e.endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	
	// 设置较短的超时
	client := &http.Client{
		Transport: e.transport,
		Timeout:   5 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// OptimizeConnectionPool 优化连接池
func (e *OptimizedHTTPExporter) OptimizeConnectionPool() {
	e.transport.connMu.Lock()
	defer e.transport.connMu.Unlock()
	
	now := time.Now()
	
	// 清理长时间未使用的连接信息
	for addr, info := range e.transport.connStats {
		if now.Sub(info.LastUsed) > 5*time.Minute {
			delete(e.transport.connStats, addr)
		}
	}
	
	// 强制关闭空闲连接
	e.client.CloseIdleConnections()
}