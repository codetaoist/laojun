package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTPExporter HTTP导出器
type HTTPExporter struct {
	name     string
	endpoint string
	config   *ExportConfig
	client   *http.Client
	mu       sync.RWMutex
	closed   bool
}

// ExportPayload 导出负载
type ExportPayload struct {
	ServiceName    string      `json:"service_name"`
	ServiceVersion string      `json:"service_version"`
	Environment    string      `json:"environment"`
	Timestamp      time.Time   `json:"timestamp"`
	Items          []BatchItem `json:"items"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NewHTTPExporter 创建新的HTTP导出器
func NewHTTPExporter(name, endpoint string, config *ExportConfig) *HTTPExporter {
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HTTPExporter{
		name:     name,
		endpoint: endpoint,
		config:   config,
		client:   client,
	}
}

// Name 返回导出器名称
func (e *HTTPExporter) Name() string {
	return e.name
}

// Export 导出数据
func (e *HTTPExporter) Export(ctx context.Context, items []BatchItem) error {
	e.mu.RLock()
	if e.closed {
		e.mu.RUnlock()
		return fmt.Errorf("exporter %s is closed", e.name)
	}
	e.mu.RUnlock()

	if len(items) == 0 {
		return nil
	}

	// 创建导出负载
	payload := ExportPayload{
		Timestamp: time.Now(),
		Items:     items,
		Metadata: map[string]interface{}{
			"exporter":    e.name,
			"batch_size":  len(items),
			"endpoint":    e.endpoint,
		},
	}

	// 序列化为JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal export payload: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", e.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("laojun-observability-exporter/%s", e.name))
	
	// 添加自定义头部
	for key, value := range e.config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close 关闭导出器
func (e *HTTPExporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.closed {
		return nil
	}
	
	e.closed = true
	
	// 关闭HTTP客户端的空闲连接
	if transport, ok := e.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	
	return nil
}

// ConsoleExporter 控制台导出器（用于调试）
type ConsoleExporter struct {
	name   string
	mu     sync.RWMutex
	closed bool
}

// NewConsoleExporter 创建新的控制台导出器
func NewConsoleExporter(name string) *ConsoleExporter {
	return &ConsoleExporter{
		name: name,
	}
}

// Name 返回导出器名称
func (e *ConsoleExporter) Name() string {
	return e.name
}

// Export 导出数据到控制台
func (e *ConsoleExporter) Export(ctx context.Context, items []BatchItem) error {
	e.mu.RLock()
	if e.closed {
		e.mu.RUnlock()
		return fmt.Errorf("console exporter %s is closed", e.name)
	}
	e.mu.RUnlock()

	if len(items) == 0 {
		return nil
	}

	fmt.Printf("=== Console Export [%s] ===\n", e.name)
	fmt.Printf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Batch Size: %d\n", len(items))
	
	for i, item := range items {
		fmt.Printf("Item %d:\n", i+1)
		fmt.Printf("  Type: %s\n", item.Type)
		fmt.Printf("  Name: %s\n", item.Name)
		fmt.Printf("  Value: %v\n", item.Value)
		fmt.Printf("  Labels: %v\n", item.Labels)
		fmt.Printf("  Timestamp: %s\n", item.Timestamp.Format(time.RFC3339))
		if len(item.Metadata) > 0 {
			fmt.Printf("  Metadata: %v\n", item.Metadata)
		}
		fmt.Println()
	}
	
	fmt.Printf("=== End Export [%s] ===\n\n", e.name)
	return nil
}

// Close 关闭控制台导出器
func (e *ConsoleExporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.closed {
		return nil
	}
	
	e.closed = true
	fmt.Printf("Console exporter %s closed\n", e.name)
	return nil
}

// FileExporter 文件导出器
type FileExporter struct {
	name     string
	filename string
	mu       sync.RWMutex
	closed   bool
}

// NewFileExporter 创建新的文件导出器
func NewFileExporter(name, filename string) *FileExporter {
	return &FileExporter{
		name:     name,
		filename: filename,
	}
}

// Name 返回导出器名称
func (e *FileExporter) Name() string {
	return e.name
}

// Export 导出数据到文件
func (e *FileExporter) Export(ctx context.Context, items []BatchItem) error {
	e.mu.RLock()
	if e.closed {
		e.mu.RUnlock()
		return fmt.Errorf("file exporter %s is closed", e.name)
	}
	e.mu.RUnlock()

	if len(items) == 0 {
		return nil
	}

	// 创建导出负载
	payload := ExportPayload{
		Timestamp: time.Now(),
		Items:     items,
		Metadata: map[string]interface{}{
			"exporter":   e.name,
			"batch_size": len(items),
			"filename":   e.filename,
		},
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export payload: %w", err)
	}

	// 这里简化实现，实际应该写入文件
	fmt.Printf("=== File Export [%s] to %s ===\n", e.name, e.filename)
	fmt.Printf("%s\n", string(data))
	fmt.Printf("=== End File Export [%s] ===\n\n", e.name)

	return nil
}

// Close 关闭文件导出器
func (e *FileExporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.closed {
		return nil
	}
	
	e.closed = true
	fmt.Printf("File exporter %s closed\n", e.name)
	return nil
}

// ExporterFactory 导出器工厂
type ExporterFactory struct{}

// CreateExporter 创建导出器
func (f *ExporterFactory) CreateExporter(exporterType, name, endpoint string, config *ExportConfig) (Exporter, error) {
	switch exporterType {
	case "http", "https":
		return NewHTTPExporter(name, endpoint, config), nil
	case "console":
		return NewConsoleExporter(name), nil
	case "file":
		return NewFileExporter(name, endpoint), nil
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", exporterType)
	}
}