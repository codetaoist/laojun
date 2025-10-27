package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// OptimizedConsoleExporter 优化的控制台导出器
type OptimizedConsoleExporter struct {
	name         string
	config       *ExportConfig
	mu           sync.Mutex
	requestCount int64
	errorCount   int64
}

// OptimizedFileExporter 优化的文件导出器
type OptimizedFileExporter struct {
	name         string
	filePath     string
	config       *ExportConfig
	file         *os.File
	mu           sync.Mutex
	requestCount int64
	errorCount   int64
	bytesWritten int64
}

// NewOptimizedConsoleExporter 创建新的优化控制台导出器
func NewOptimizedConsoleExporter(name string, config *ExportConfig) (*OptimizedConsoleExporter, error) {
	return &OptimizedConsoleExporter{
		name:   name,
		config: config,
	}, nil
}

// Export 导出数据到控制台
func (e *OptimizedConsoleExporter) Export(ctx context.Context, items []BatchItem) error {
	atomic.AddInt64(&e.requestCount, 1)
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to marshal items: %w", err)
	}
	
	fmt.Printf("[%s] Optimized Console Exporter - %s:\n%s\n", 
		time.Now().Format("2006-01-02 15:04:05"), e.name, string(data))
	
	return nil
}

// ExportCached 导出缓存数据到控制台
func (e *OptimizedConsoleExporter) ExportCached(ctx context.Context, cachedData []byte) error {
	atomic.AddInt64(&e.requestCount, 1)
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	fmt.Printf("[%s] Optimized Console Exporter - %s (Cached):\n%s\n", 
		time.Now().Format("2006-01-02 15:04:05"), e.name, string(cachedData))
	
	return nil
}

// Name 返回导出器名称
func (e *OptimizedConsoleExporter) Name() string {
	return e.name
}

// Close 关闭导出器
func (e *OptimizedConsoleExporter) Close() error {
	return nil
}

// GetConnectionPoolStats 获取连接池统计（控制台导出器不需要连接池）
func (e *OptimizedConsoleExporter) GetConnectionPoolStats() ConnectionPoolStats {
	return ConnectionPoolStats{
		ActiveConnections: 0,
		IdleConnections:   0,
		TotalConnections:  0,
		MaxConnections:    0,
	}
}

// NewOptimizedFileExporter 创建新的优化文件导出器
func NewOptimizedFileExporter(name, endpoint string, config *ExportConfig) (*OptimizedFileExporter, error) {
	// 从endpoint中提取文件路径（去掉file://前缀）
	filePath := endpoint
	if len(endpoint) >= 7 && endpoint[:7] == "file://" {
		filePath = endpoint[7:]
	}
	
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// 打开文件（追加模式）
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	
	return &OptimizedFileExporter{
		name:     name,
		filePath: filePath,
		config:   config,
		file:     file,
	}, nil
}

// Export 导出数据到文件
func (e *OptimizedFileExporter) Export(ctx context.Context, items []BatchItem) error {
	atomic.AddInt64(&e.requestCount, 1)
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// 检查文件是否仍然有效
	if e.file == nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("file is closed")
	}
	
	// 序列化数据
	data, err := json.Marshal(items)
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to marshal items: %w", err)
	}
	
	// 添加时间戳和换行
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	line := fmt.Sprintf("[%s] %s\n", timestamp, string(data))
	
	// 写入文件
	n, err := e.file.WriteString(line)
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to write to file: %w", err)
	}
	
	atomic.AddInt64(&e.bytesWritten, int64(n))
	
	// 强制刷新到磁盘
	if err := e.file.Sync(); err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to sync file: %w", err)
	}
	
	return nil
}

// ExportCached 导出缓存数据到文件
func (e *OptimizedFileExporter) ExportCached(ctx context.Context, cachedData []byte) error {
	atomic.AddInt64(&e.requestCount, 1)
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// 检查文件是否仍然有效
	if e.file == nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("file is closed")
	}
	
	// 添加时间戳和换行
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	line := fmt.Sprintf("[%s] (Cached) %s\n", timestamp, string(cachedData))
	
	// 写入文件
	n, err := e.file.WriteString(line)
	if err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to write to file: %w", err)
	}
	
	atomic.AddInt64(&e.bytesWritten, int64(n))
	
	// 强制刷新到磁盘
	if err := e.file.Sync(); err != nil {
		atomic.AddInt64(&e.errorCount, 1)
		return fmt.Errorf("failed to sync file: %w", err)
	}
	
	return nil
}

// Name 返回导出器名称
func (e *OptimizedFileExporter) Name() string {
	return e.name
}

// Close 关闭导出器
func (e *OptimizedFileExporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.file != nil {
		err := e.file.Close()
		e.file = nil
		return err
	}
	
	return nil
}

// GetConnectionPoolStats 获取连接池统计（文件导出器不需要连接池）
func (e *OptimizedFileExporter) GetConnectionPoolStats() ConnectionPoolStats {
	return ConnectionPoolStats{
		ActiveConnections: 0,
		IdleConnections:   0,
		TotalConnections:  0,
		MaxConnections:    0,
	}
}

// GetFileStats 获取文件统计信息
func (e *OptimizedFileExporter) GetFileStats() map[string]interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	stats := map[string]interface{}{
		"file_path":      e.filePath,
		"request_count":  atomic.LoadInt64(&e.requestCount),
		"error_count":    atomic.LoadInt64(&e.errorCount),
		"bytes_written":  atomic.LoadInt64(&e.bytesWritten),
	}
	
	// 获取文件信息
	if e.file != nil {
		if fileInfo, err := e.file.Stat(); err == nil {
			stats["file_size"] = fileInfo.Size()
			stats["file_mode"] = fileInfo.Mode().String()
			stats["modified_time"] = fileInfo.ModTime()
		}
	}
	
	return stats
}

// RotateFile 轮转文件
func (e *OptimizedFileExporter) RotateFile() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.file == nil {
		return fmt.Errorf("file is not open")
	}
	
	// 关闭当前文件
	if err := e.file.Close(); err != nil {
		return fmt.Errorf("failed to close current file: %w", err)
	}
	
	// 重命名当前文件
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := fmt.Sprintf("%s.%s", e.filePath, timestamp)
	
	if err := os.Rename(e.filePath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate file: %w", err)
	}
	
	// 创建新文件
	file, err := os.OpenFile(e.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}
	
	e.file = file
	return nil
}

// GetConsoleStats 获取控制台导出器统计信息
func (e *OptimizedConsoleExporter) GetConsoleStats() map[string]interface{} {
	return map[string]interface{}{
		"request_count": atomic.LoadInt64(&e.requestCount),
		"error_count":   atomic.LoadInt64(&e.errorCount),
	}
}