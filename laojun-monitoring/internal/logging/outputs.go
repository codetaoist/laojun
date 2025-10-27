package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LogOutput 日志输出接口
type LogOutput interface {
	// Write 写入日志条目
	Write(ctx context.Context, entries []*LogEntry) error
	
	// Name 返回输出名称
	Name() string
	
	// IsEnabled 返回是否启用
	IsEnabled() bool
	
	// SetEnabled 设置启用状态
	SetEnabled(enabled bool)
	
	// Close 关闭输出
	Close() error
	
	// GetStats 获取统计信息
	GetStats() OutputStats
}

// OutputStats 输出统计信息
type OutputStats struct {
	Name         string    `json:"name"`
	WrittenCount int64     `json:"written_count"`
	ErrorCount   int64     `json:"error_count"`
	LastWrite    time.Time `json:"last_write"`
	LastError    string    `json:"last_error,omitempty"`
	LastErrorTime time.Time `json:"last_error_time,omitempty"`
	IsEnabled    bool      `json:"is_enabled"`
}

// FileOutput 文件输出
type FileOutput struct {
	mu       sync.RWMutex
	name     string
	config   FileOutputConfig
	logger   *zap.Logger
	file     *os.File
	enabled  bool
	stats    OutputStats
}

// FileOutputConfig 文件输出配置
type FileOutputConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    int64  `mapstructure:"max_size"`    // 最大文件大小(字节)
	MaxFiles   int    `mapstructure:"max_files"`   // 最大文件数量
	Compress   bool   `mapstructure:"compress"`    // 是否压缩旧文件
	Format     string `mapstructure:"format"`      // 输出格式: json, text
	TimeFormat string `mapstructure:"time_format"` // 时间格式
}

// NewFileOutput 创建文件输出
func NewFileOutput(name string, config FileOutputConfig, logger *zap.Logger) *FileOutput {
	if config.MaxSize == 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 10
	}
	if config.Format == "" {
		config.Format = "json"
	}
	if config.TimeFormat == "" {
		config.TimeFormat = time.RFC3339
	}
	
	return &FileOutput{
		name:    name,
		config:  config,
		logger:  logger,
		enabled: true,
		stats: OutputStats{
			Name:      name,
			IsEnabled: true,
		},
	}
}

// Write 写入日志条目
func (fo *FileOutput) Write(ctx context.Context, entries []*LogEntry) error {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	
	if !fo.enabled {
		return nil
	}
	
	// 确保文件已打开
	if err := fo.ensureFile(); err != nil {
		fo.updateError(err)
		return err
	}
	
	// 检查文件大小，必要时轮转
	if err := fo.checkRotation(); err != nil {
		fo.updateError(err)
		return err
	}
	
	// 写入条目
	for _, entry := range entries {
		line, err := fo.formatEntry(entry)
		if err != nil {
			fo.updateError(err)
			continue
		}
		
		if _, err := fo.file.WriteString(line + "\n"); err != nil {
			fo.updateError(err)
			return err
		}
		
		fo.stats.WrittenCount++
	}
	
	// 刷新到磁盘
	if err := fo.file.Sync(); err != nil {
		fo.updateError(err)
		return err
	}
	
	fo.stats.LastWrite = time.Now()
	
	return nil
}

// Name 返回输出名称
func (fo *FileOutput) Name() string {
	return fo.name
}

// IsEnabled 返回是否启用
func (fo *FileOutput) IsEnabled() bool {
	fo.mu.RLock()
	defer fo.mu.RUnlock()
	return fo.enabled
}

// SetEnabled 设置启用状态
func (fo *FileOutput) SetEnabled(enabled bool) {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	fo.enabled = enabled
	fo.stats.IsEnabled = enabled
}

// Close 关闭输出
func (fo *FileOutput) Close() error {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	
	if fo.file != nil {
		err := fo.file.Close()
		fo.file = nil
		return err
	}
	
	return nil
}

// GetStats 获取统计信息
func (fo *FileOutput) GetStats() OutputStats {
	fo.mu.RLock()
	defer fo.mu.RUnlock()
	return fo.stats
}

// ensureFile 确保文件已打开
func (fo *FileOutput) ensureFile() error {
	if fo.file != nil {
		return nil
	}
	
	// 创建目录
	dir := filepath.Dir(fo.config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// 打开文件
	file, err := os.OpenFile(fo.config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	
	fo.file = file
	
	return nil
}

// checkRotation 检查文件轮转
func (fo *FileOutput) checkRotation() error {
	if fo.file == nil {
		return nil
	}
	
	// 获取文件信息
	info, err := fo.file.Stat()
	if err != nil {
		return err
	}
	
	// 检查是否需要轮转
	if info.Size() < fo.config.MaxSize {
		return nil
	}
	
	// 关闭当前文件
	if err := fo.file.Close(); err != nil {
		return err
	}
	fo.file = nil
	
	// 轮转文件
	if err := fo.rotateFiles(); err != nil {
		return err
	}
	
	// 重新打开文件
	return fo.ensureFile()
}

// rotateFiles 轮转文件
func (fo *FileOutput) rotateFiles() error {
	// 移动现有文件
	for i := fo.config.MaxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", fo.config.Path, i)
		newPath := fmt.Sprintf("%s.%d", fo.config.Path, i+1)
		
		if i == fo.config.MaxFiles-1 {
			// 删除最老的文件
			os.Remove(newPath)
		}
		
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}
	
	// 移动当前文件
	backupPath := fmt.Sprintf("%s.1", fo.config.Path)
	if err := os.Rename(fo.config.Path, backupPath); err != nil {
		return err
	}
	
	// 压缩文件（如果启用）
	if fo.config.Compress {
		go fo.compressFile(backupPath)
	}
	
	return nil
}

// compressFile 压缩文件
func (fo *FileOutput) compressFile(filePath string) {
	// 这里应该实现文件压缩逻辑
	// 为简化，这里只是记录日志
	fo.logger.Info("File compression not implemented", zap.String("file", filePath))
}

// formatEntry 格式化日志条目
func (fo *FileOutput) formatEntry(entry *LogEntry) (string, error) {
	switch fo.config.Format {
	case "json":
		return fo.formatJSON(entry)
	case "text":
		return fo.formatText(entry)
	default:
		return fo.formatJSON(entry)
	}
}

// formatJSON 格式化为JSON
func (fo *FileOutput) formatJSON(entry *LogEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatText 格式化为文本
func (fo *FileOutput) formatText(entry *LogEntry) (string, error) {
	timestamp := entry.Timestamp.Format(fo.config.TimeFormat)
	return fmt.Sprintf("%s [%s] %s: %s", timestamp, entry.Level, entry.Source, entry.Message), nil
}

// updateError 更新错误信息
func (fo *FileOutput) updateError(err error) {
	fo.stats.ErrorCount++
	fo.stats.LastError = err.Error()
	fo.stats.LastErrorTime = time.Now()
}

// ConsoleOutput 控制台输出
type ConsoleOutput struct {
	mu      sync.RWMutex
	name    string
	config  ConsoleOutputConfig
	logger  *zap.Logger
	writer  io.Writer
	enabled bool
	stats   OutputStats
}

// ConsoleOutputConfig 控制台输出配置
type ConsoleOutputConfig struct {
	Format     string `mapstructure:"format"`      // 输出格式: json, text, colored
	TimeFormat string `mapstructure:"time_format"` // 时间格式
	Colors     bool   `mapstructure:"colors"`      // 是否使用颜色
}

// NewConsoleOutput 创建控制台输出
func NewConsoleOutput(name string, config ConsoleOutputConfig, logger *zap.Logger) *ConsoleOutput {
	if config.Format == "" {
		config.Format = "colored"
	}
	if config.TimeFormat == "" {
		config.TimeFormat = "15:04:05"
	}
	
	return &ConsoleOutput{
		name:    name,
		config:  config,
		logger:  logger,
		writer:  os.Stdout,
		enabled: true,
		stats: OutputStats{
			Name:      name,
			IsEnabled: true,
		},
	}
}

// Write 写入日志条目
func (co *ConsoleOutput) Write(ctx context.Context, entries []*LogEntry) error {
	co.mu.Lock()
	defer co.mu.Unlock()
	
	if !co.enabled {
		return nil
	}
	
	for _, entry := range entries {
		line, err := co.formatEntry(entry)
		if err != nil {
			co.updateError(err)
			continue
		}
		
		if _, err := fmt.Fprintln(co.writer, line); err != nil {
			co.updateError(err)
			return err
		}
		
		co.stats.WrittenCount++
	}
	
	co.stats.LastWrite = time.Now()
	
	return nil
}

// Name 返回输出名称
func (co *ConsoleOutput) Name() string {
	return co.name
}

// IsEnabled 返回是否启用
func (co *ConsoleOutput) IsEnabled() bool {
	co.mu.RLock()
	defer co.mu.RUnlock()
	return co.enabled
}

// SetEnabled 设置启用状态
func (co *ConsoleOutput) SetEnabled(enabled bool) {
	co.mu.Lock()
	defer co.mu.Unlock()
	co.enabled = enabled
	co.stats.IsEnabled = enabled
}

// Close 关闭输出
func (co *ConsoleOutput) Close() error {
	return nil
}

// GetStats 获取统计信息
func (co *ConsoleOutput) GetStats() OutputStats {
	co.mu.RLock()
	defer co.mu.RUnlock()
	return co.stats
}

// formatEntry 格式化日志条目
func (co *ConsoleOutput) formatEntry(entry *LogEntry) (string, error) {
	switch co.config.Format {
	case "json":
		return co.formatJSON(entry)
	case "text":
		return co.formatText(entry)
	case "colored":
		return co.formatColored(entry)
	default:
		return co.formatColored(entry)
	}
}

// formatJSON 格式化为JSON
func (co *ConsoleOutput) formatJSON(entry *LogEntry) (string, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatText 格式化为文本
func (co *ConsoleOutput) formatText(entry *LogEntry) (string, error) {
	timestamp := entry.Timestamp.Format(co.config.TimeFormat)
	return fmt.Sprintf("%s [%s] %s: %s", timestamp, entry.Level, entry.Source, entry.Message), nil
}

// formatColored 格式化为彩色文本
func (co *ConsoleOutput) formatColored(entry *LogEntry) (string, error) {
	if !co.config.Colors {
		return co.formatText(entry)
	}
	
	timestamp := entry.Timestamp.Format(co.config.TimeFormat)
	
	// ANSI颜色代码
	var levelColor string
	switch entry.Level {
	case LogLevelDebug:
		levelColor = "\033[36m" // 青色
	case LogLevelInfo:
		levelColor = "\033[32m" // 绿色
	case LogLevelWarn:
		levelColor = "\033[33m" // 黄色
	case LogLevelError:
		levelColor = "\033[31m" // 红色
	case LogLevelFatal:
		levelColor = "\033[35m" // 紫色
	default:
		levelColor = "\033[0m" // 默认
	}
	
	reset := "\033[0m"
	
	return fmt.Sprintf("%s [%s%s%s] %s: %s", 
		timestamp, levelColor, entry.Level, reset, entry.Source, entry.Message), nil
}

// updateError 更新错误信息
func (co *ConsoleOutput) updateError(err error) {
	co.stats.ErrorCount++
	co.stats.LastError = err.Error()
	co.stats.LastErrorTime = time.Now()
}

// ElasticsearchOutput Elasticsearch输出
type ElasticsearchOutput struct {
	mu         sync.RWMutex
	name       string
	config     ElasticsearchOutputConfig
	logger     *zap.Logger
	httpClient *http.Client
	enabled    bool
	stats      OutputStats
}

// ElasticsearchOutputConfig Elasticsearch输出配置
type ElasticsearchOutputConfig struct {
	Hosts          []string          `mapstructure:"hosts"`
	Index          string            `mapstructure:"index"`
	IndexPattern   string            `mapstructure:"index_pattern"` // 如: logs-2006-01-02
	Username       string            `mapstructure:"username"`
	Password       string            `mapstructure:"password"`
	Timeout        time.Duration     `mapstructure:"timeout"`
	BatchSize      int               `mapstructure:"batch_size"`
	FlushInterval  time.Duration     `mapstructure:"flush_interval"`
	Headers        map[string]string `mapstructure:"headers"`
}

// NewElasticsearchOutput 创建Elasticsearch输出
func NewElasticsearchOutput(name string, config ElasticsearchOutputConfig, logger *zap.Logger) *ElasticsearchOutput {
	if config.Index == "" {
		config.Index = "logs"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	
	return &ElasticsearchOutput{
		name:   name,
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		enabled: true,
		stats: OutputStats{
			Name:      name,
			IsEnabled: true,
		},
	}
}

// Write 写入日志条目
func (eo *ElasticsearchOutput) Write(ctx context.Context, entries []*LogEntry) error {
	eo.mu.Lock()
	defer eo.mu.Unlock()
	
	if !eo.enabled {
		return nil
	}
	
	// 构建批量请求
	var buf bytes.Buffer
	
	for _, entry := range entries {
		// 确定索引名称
		indexName := eo.getIndexName(entry.Timestamp)
		
		// 构建索引操作
		indexOp := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}
		
		indexOpJSON, err := json.Marshal(indexOp)
		if err != nil {
			eo.updateError(err)
			continue
		}
		
		entryJSON, err := json.Marshal(entry)
		if err != nil {
			eo.updateError(err)
			continue
		}
		
		buf.Write(indexOpJSON)
		buf.WriteString("\n")
		buf.Write(entryJSON)
		buf.WriteString("\n")
	}
	
	if buf.Len() == 0 {
		return nil
	}
	
	// 发送批量请求
	if err := eo.sendBulkRequest(ctx, buf.Bytes()); err != nil {
		eo.updateError(err)
		return err
	}
	
	eo.stats.WrittenCount += int64(len(entries))
	eo.stats.LastWrite = time.Now()
	
	return nil
}

// Name 返回输出名称
func (eo *ElasticsearchOutput) Name() string {
	return eo.name
}

// IsEnabled 返回是否启用
func (eo *ElasticsearchOutput) IsEnabled() bool {
	eo.mu.RLock()
	defer eo.mu.RUnlock()
	return eo.enabled
}

// SetEnabled 设置启用状态
func (eo *ElasticsearchOutput) SetEnabled(enabled bool) {
	eo.mu.Lock()
	defer eo.mu.Unlock()
	eo.enabled = enabled
	eo.stats.IsEnabled = enabled
}

// Close 关闭输出
func (eo *ElasticsearchOutput) Close() error {
	return nil
}

// GetStats 获取统计信息
func (eo *ElasticsearchOutput) GetStats() OutputStats {
	eo.mu.RLock()
	defer eo.mu.RUnlock()
	return eo.stats
}

// getIndexName 获取索引名称
func (eo *ElasticsearchOutput) getIndexName(timestamp time.Time) string {
	if eo.config.IndexPattern != "" {
		return timestamp.Format(eo.config.IndexPattern)
	}
	return eo.config.Index
}

// sendBulkRequest 发送批量请求
func (eo *ElasticsearchOutput) sendBulkRequest(ctx context.Context, data []byte) error {
	if len(eo.config.Hosts) == 0 {
		return fmt.Errorf("no elasticsearch hosts configured")
	}
	
	// 使用第一个主机（实际应该实现负载均衡）
	url := fmt.Sprintf("%s/_bulk", strings.TrimSuffix(eo.config.Hosts[0], "/"))
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置头部
	req.Header.Set("Content-Type", "application/x-ndjson")
	
	// 设置认证
	if eo.config.Username != "" && eo.config.Password != "" {
		req.SetBasicAuth(eo.config.Username, eo.config.Password)
	}
	
	// 设置自定义头部
	for k, v := range eo.config.Headers {
		req.Header.Set(k, v)
	}
	
	// 发送请求
	resp, err := eo.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elasticsearch returned status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// updateError 更新错误信息
func (eo *ElasticsearchOutput) updateError(err error) {
	eo.stats.ErrorCount++
	eo.stats.LastError = err.Error()
	eo.stats.LastErrorTime = time.Now()
}

// OutputManager 输出管理器
type OutputManager struct {
	mu      sync.RWMutex
	outputs map[string]LogOutput
	logger  *zap.Logger
}

// NewOutputManager 创建输出管理器
func NewOutputManager(logger *zap.Logger) *OutputManager {
	return &OutputManager{
		outputs: make(map[string]LogOutput),
		logger:  logger,
	}
}

// Register 注册输出
func (om *OutputManager) Register(name string, output LogOutput) {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	om.outputs[name] = output
	om.logger.Info("Registered log output", zap.String("name", name))
}

// Unregister 注销输出
func (om *OutputManager) Unregister(name string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	if output, exists := om.outputs[name]; exists {
		output.Close()
		delete(om.outputs, name)
		om.logger.Info("Unregistered log output", zap.String("name", name))
	}
}

// Get 获取输出
func (om *OutputManager) Get(name string) (LogOutput, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()
	
	output, exists := om.outputs[name]
	return output, exists
}

// List 列出所有输出
func (om *OutputManager) List() []string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	
	names := make([]string, 0, len(om.outputs))
	for name := range om.outputs {
		names = append(names, name)
	}
	
	return names
}

// WriteToAll 写入到所有启用的输出
func (om *OutputManager) WriteToAll(ctx context.Context, entries []*LogEntry) error {
	om.mu.RLock()
	outputs := make([]LogOutput, 0, len(om.outputs))
	for _, output := range om.outputs {
		if output.IsEnabled() {
			outputs = append(outputs, output)
		}
	}
	om.mu.RUnlock()
	
	var lastErr error
	
	for _, output := range outputs {
		if err := output.Write(ctx, entries); err != nil {
			om.logger.Error("Failed to write to output", 
				zap.String("output", output.Name()), 
				zap.Error(err))
			lastErr = err
		}
	}
	
	return lastErr
}

// WriteToSpecific 写入到指定输出
func (om *OutputManager) WriteToSpecific(ctx context.Context, outputNames []string, entries []*LogEntry) error {
	var lastErr error
	
	for _, name := range outputNames {
		om.mu.RLock()
		output, exists := om.outputs[name]
		om.mu.RUnlock()
		
		if !exists {
			om.logger.Warn("Output not found", zap.String("name", name))
			continue
		}
		
		if !output.IsEnabled() {
			continue
		}
		
		if err := output.Write(ctx, entries); err != nil {
			om.logger.Error("Failed to write to output", 
				zap.String("output", name), 
				zap.Error(err))
			lastErr = err
		}
	}
	
	return lastErr
}

// GetAllStats 获取所有输出统计信息
func (om *OutputManager) GetAllStats() map[string]OutputStats {
	om.mu.RLock()
	defer om.mu.RUnlock()
	
	stats := make(map[string]OutputStats)
	for name, output := range om.outputs {
		stats[name] = output.GetStats()
	}
	
	return stats
}

// Close 关闭所有输出
func (om *OutputManager) Close() error {
	om.mu.Lock()
	defer om.mu.Unlock()
	
	var lastErr error
	
	for name, output := range om.outputs {
		if err := output.Close(); err != nil {
			om.logger.Error("Failed to close output", 
				zap.String("name", name), 
				zap.Error(err))
			lastErr = err
		}
	}
	
	return lastErr
}