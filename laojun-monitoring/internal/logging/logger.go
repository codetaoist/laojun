package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     LogLevel          `json:"level"`
	Message   string            `json:"message"`
	Source    string            `json:"source"`
	Service   string            `json:"service"`
	TraceID   string            `json:"trace_id,omitempty"`
	SpanID    string            `json:"span_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// LogCollector 日志收集器接口
type LogCollector interface {
	// Start 启动收集器
	Start(ctx context.Context) error
	
	// Stop 停止收集器
	Stop() error
	
	// Name 返回收集器名称
	Name() string
	
	// IsRunning 返回是否正在运行
	IsRunning() bool
	
	// GetStats 获取统计信息
	GetStats() CollectorStats
	
	// SetOutput 设置输出通道
	SetOutput(output chan<- *LogEntry)
}

// CollectorStats 收集器统计信息
type CollectorStats struct {
	Name           string    `json:"name"`
	StartTime      time.Time `json:"start_time"`
	CollectedCount int64     `json:"collected_count"`
	ErrorCount     int64     `json:"error_count"`
	LastError      string    `json:"last_error,omitempty"`
	LastErrorTime  time.Time `json:"last_error_time,omitempty"`
	IsRunning      bool      `json:"is_running"`
}

// FileCollector 文件日志收集器
type FileCollector struct {
	mu       sync.RWMutex
	name     string
	config   FileCollectorConfig
	logger   *zap.Logger
	output   chan<- *LogEntry
	running  bool
	cancel   context.CancelFunc
	stats    CollectorStats
	watchers map[string]*FileWatcher
}

// FileCollectorConfig 文件收集器配置
type FileCollectorConfig struct {
	Paths       []string          `mapstructure:"paths"`
	Patterns    []string          `mapstructure:"patterns"`
	Exclude     []string          `mapstructure:"exclude"`
	Recursive   bool              `mapstructure:"recursive"`
	TailMode    bool              `mapstructure:"tail_mode"`
	BufferSize  int               `mapstructure:"buffer_size"`
	Parser      string            `mapstructure:"parser"`
	Fields      map[string]string `mapstructure:"fields"`
	Tags        map[string]string `mapstructure:"tags"`
	Service     string            `mapstructure:"service"`
}

// NewFileCollector 创建文件收集器
func NewFileCollector(name string, config FileCollectorConfig, logger *zap.Logger) *FileCollector {
	if config.BufferSize == 0 {
		config.BufferSize = 1024
	}
	
	return &FileCollector{
		name:     name,
		config:   config,
		logger:   logger,
		watchers: make(map[string]*FileWatcher),
		stats: CollectorStats{
			Name: name,
		},
	}
}

// Start 启动文件收集器
func (fc *FileCollector) Start(ctx context.Context) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	if fc.running {
		return fmt.Errorf("collector already running")
	}
	
	ctx, cancel := context.WithCancel(ctx)
	fc.cancel = cancel
	fc.running = true
	fc.stats.StartTime = time.Now()
	fc.stats.IsRunning = true
	
	// 启动文件监控
	for _, path := range fc.config.Paths {
		if err := fc.watchPath(ctx, path); err != nil {
			fc.logger.Error("Failed to watch path", 
				zap.String("path", path), 
				zap.Error(err))
			fc.updateError(err)
		}
	}
	
	fc.logger.Info("File collector started", 
		zap.String("name", fc.name),
		zap.Strings("paths", fc.config.Paths))
	
	return nil
}

// Stop 停止文件收集器
func (fc *FileCollector) Stop() error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	
	if !fc.running {
		return nil
	}
	
	if fc.cancel != nil {
		fc.cancel()
	}
	
	// 停止所有文件监控器
	for _, watcher := range fc.watchers {
		watcher.Stop()
	}
	
	fc.running = false
	fc.stats.IsRunning = false
	
	fc.logger.Info("File collector stopped", zap.String("name", fc.name))
	
	return nil
}

// Name 返回收集器名称
func (fc *FileCollector) Name() string {
	return fc.name
}

// IsRunning 返回是否正在运行
func (fc *FileCollector) IsRunning() bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.running
}

// GetStats 获取统计信息
func (fc *FileCollector) GetStats() CollectorStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.stats
}

// SetOutput 设置输出通道
func (fc *FileCollector) SetOutput(output chan<- *LogEntry) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.output = output
}

// watchPath 监控路径
func (fc *FileCollector) watchPath(ctx context.Context, path string) error {
	// 检查路径是否存在
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %s", path)
	}
	
	if info.IsDir() {
		return fc.watchDirectory(ctx, path)
	} else {
		return fc.watchFile(ctx, path)
	}
}

// watchDirectory 监控目录
func (fc *FileCollector) watchDirectory(ctx context.Context, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// 检查文件是否匹配模式
		if fc.shouldWatchFile(path) {
			return fc.watchFile(ctx, path)
		}
		
		return nil
	})
}

// watchFile 监控单个文件
func (fc *FileCollector) watchFile(ctx context.Context, filePath string) error {
	watcher := NewFileWatcher(filePath, fc.config.TailMode, fc.logger)
	
	// 设置处理函数
	watcher.SetHandler(func(line string) {
		entry := fc.parseLogLine(line, filePath)
		if entry != nil && fc.output != nil {
			select {
			case fc.output <- entry:
				fc.updateStats()
			case <-ctx.Done():
				return
			}
		}
	})
	
	// 启动监控
	if err := watcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}
	
	fc.watchers[filePath] = watcher
	
	return nil
}

// shouldWatchFile 检查是否应该监控文件
func (fc *FileCollector) shouldWatchFile(filePath string) bool {
	// 检查排除模式
	for _, exclude := range fc.config.Exclude {
		if matched, _ := filepath.Match(exclude, filepath.Base(filePath)); matched {
			return false
		}
	}
	
	// 检查包含模式
	if len(fc.config.Patterns) == 0 {
		return true
	}
	
	for _, pattern := range fc.config.Patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
	}
	
	return false
}

// parseLogLine 解析日志行
func (fc *FileCollector) parseLogLine(line, source string) *LogEntry {
	if strings.TrimSpace(line) == "" {
		return nil
	}
	
	entry := &LogEntry{
		Timestamp: time.Now(),
		Message:   line,
		Source:    source,
		Service:   fc.config.Service,
		Fields:    make(map[string]interface{}),
		Tags:      make(map[string]string),
	}
	
	// 添加配置的字段和标签
	for k, v := range fc.config.Fields {
		entry.Fields[k] = v
	}
	
	for k, v := range fc.config.Tags {
		entry.Tags[k] = v
	}
	
	// 根据解析器类型解析
	switch fc.config.Parser {
	case "json":
		fc.parseJSONLog(line, entry)
	case "nginx":
		fc.parseNginxLog(line, entry)
	case "apache":
		fc.parseApacheLog(line, entry)
	default:
		fc.parseTextLog(line, entry)
	}
	
	return entry
}

// parseJSONLog 解析JSON格式日志
func (fc *FileCollector) parseJSONLog(line string, entry *LogEntry) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		entry.Level = LogLevelInfo
		return
	}
	
	// 提取标准字段
	if timestamp, ok := data["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			entry.Timestamp = t
		}
	}
	
	if level, ok := data["level"].(string); ok {
		entry.Level = LogLevel(level)
	} else {
		entry.Level = LogLevelInfo
	}
	
	if message, ok := data["message"].(string); ok {
		entry.Message = message
	}
	
	if traceID, ok := data["trace_id"].(string); ok {
		entry.TraceID = traceID
	}
	
	if spanID, ok := data["span_id"].(string); ok {
		entry.SpanID = spanID
	}
	
	// 其他字段作为额外字段
	for k, v := range data {
		if k != "timestamp" && k != "level" && k != "message" && k != "trace_id" && k != "span_id" {
			entry.Fields[k] = v
		}
	}
}

// parseTextLog 解析纯文本日志
func (fc *FileCollector) parseTextLog(line string, entry *LogEntry) {
	// 简单的文本解析，尝试提取日志级别
	lowerLine := strings.ToLower(line)
	
	switch {
	case strings.Contains(lowerLine, "error"):
		entry.Level = LogLevelError
	case strings.Contains(lowerLine, "warn"):
		entry.Level = LogLevelWarn
	case strings.Contains(lowerLine, "debug"):
		entry.Level = LogLevelDebug
	case strings.Contains(lowerLine, "fatal"):
		entry.Level = LogLevelFatal
	default:
		entry.Level = LogLevelInfo
	}
}

// parseNginxLog 解析Nginx日志
func (fc *FileCollector) parseNginxLog(line string, entry *LogEntry) {
	// Nginx访问日志格式解析
	// 这里实现简化版本，实际应该根据具体格式解析
	entry.Level = LogLevelInfo
	entry.Fields["log_type"] = "nginx_access"
}

// parseApacheLog 解析Apache日志
func (fc *FileCollector) parseApacheLog(line string, entry *LogEntry) {
	// Apache访问日志格式解析
	// 这里实现简化版本，实际应该根据具体格式解析
	entry.Level = LogLevelInfo
	entry.Fields["log_type"] = "apache_access"
}

// updateStats 更新统计信息
func (fc *FileCollector) updateStats() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.stats.CollectedCount++
}

// updateError 更新错误信息
func (fc *FileCollector) updateError(err error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.stats.ErrorCount++
	fc.stats.LastError = err.Error()
	fc.stats.LastErrorTime = time.Now()
}

// FileWatcher 文件监控器
type FileWatcher struct {
	filePath string
	tailMode bool
	logger   *zap.Logger
	handler  func(string)
	running  bool
	cancel   context.CancelFunc
	mu       sync.RWMutex
}

// NewFileWatcher 创建文件监控器
func NewFileWatcher(filePath string, tailMode bool, logger *zap.Logger) *FileWatcher {
	return &FileWatcher{
		filePath: filePath,
		tailMode: tailMode,
		logger:   logger,
	}
}

// SetHandler 设置处理函数
func (fw *FileWatcher) SetHandler(handler func(string)) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.handler = handler
}

// Start 启动文件监控
func (fw *FileWatcher) Start(ctx context.Context) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	
	if fw.running {
		return fmt.Errorf("file watcher already running")
	}
	
	ctx, cancel := context.WithCancel(ctx)
	fw.cancel = cancel
	fw.running = true
	
	go fw.watchLoop(ctx)
	
	return nil
}

// Stop 停止文件监控
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	
	if !fw.running {
		return
	}
	
	if fw.cancel != nil {
		fw.cancel()
	}
	
	fw.running = false
}

// watchLoop 监控循环
func (fw *FileWatcher) watchLoop(ctx context.Context) {
	defer func() {
		fw.mu.Lock()
		fw.running = false
		fw.mu.Unlock()
	}()
	
	var offset int64
	
	// 如果是tail模式，从文件末尾开始
	if fw.tailMode {
		if info, err := os.Stat(fw.filePath); err == nil {
			offset = info.Size()
		}
	}
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newOffset, err := fw.readNewLines(offset)
			if err != nil {
				fw.logger.Error("Failed to read file", 
					zap.String("file", fw.filePath), 
					zap.Error(err))
				continue
			}
			offset = newOffset
		}
	}
}

// readNewLines 读取新行
func (fw *FileWatcher) readNewLines(offset int64) (int64, error) {
	file, err := os.Open(fw.filePath)
	if err != nil {
		return offset, err
	}
	defer file.Close()
	
	// 获取当前文件大小
	info, err := file.Stat()
	if err != nil {
		return offset, err
	}
	
	// 如果文件被截断，从头开始读
	if info.Size() < offset {
		offset = 0
	}
	
	// 如果没有新内容，返回
	if info.Size() == offset {
		return offset, nil
	}
	
	// 定位到偏移位置
	if _, err := file.Seek(offset, 0); err != nil {
		return offset, err
	}
	
	// 读取新行
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if fw.handler != nil {
			fw.handler(line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return offset, err
	}
	
	return info.Size(), nil
}

// SystemdCollector systemd日志收集器
type SystemdCollector struct {
	mu      sync.RWMutex
	name    string
	config  SystemdCollectorConfig
	logger  *zap.Logger
	output  chan<- *LogEntry
	running bool
	cancel  context.CancelFunc
	stats   CollectorStats
}

// SystemdCollectorConfig systemd收集器配置
type SystemdCollectorConfig struct {
	Units    []string          `mapstructure:"units"`
	Since    string            `mapstructure:"since"`
	Follow   bool              `mapstructure:"follow"`
	Fields   map[string]string `mapstructure:"fields"`
	Tags     map[string]string `mapstructure:"tags"`
	Service  string            `mapstructure:"service"`
}

// NewSystemdCollector 创建systemd收集器
func NewSystemdCollector(name string, config SystemdCollectorConfig, logger *zap.Logger) *SystemdCollector {
	return &SystemdCollector{
		name:   name,
		config: config,
		logger: logger,
		stats: CollectorStats{
			Name: name,
		},
	}
}

// Start 启动systemd收集器
func (sc *SystemdCollector) Start(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if sc.running {
		return fmt.Errorf("collector already running")
	}
	
	ctx, cancel := context.WithCancel(ctx)
	sc.cancel = cancel
	sc.running = true
	sc.stats.StartTime = time.Now()
	sc.stats.IsRunning = true
	
	// 启动journalctl监控
	go sc.collectLoop(ctx)
	
	sc.logger.Info("Systemd collector started", 
		zap.String("name", sc.name),
		zap.Strings("units", sc.config.Units))
	
	return nil
}

// Stop 停止systemd收集器
func (sc *SystemdCollector) Stop() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if !sc.running {
		return nil
	}
	
	if sc.cancel != nil {
		sc.cancel()
	}
	
	sc.running = false
	sc.stats.IsRunning = false
	
	sc.logger.Info("Systemd collector stopped", zap.String("name", sc.name))
	
	return nil
}

// Name 返回收集器名称
func (sc *SystemdCollector) Name() string {
	return sc.name
}

// IsRunning 返回是否正在运行
func (sc *SystemdCollector) IsRunning() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.running
}

// GetStats 获取统计信息
func (sc *SystemdCollector) GetStats() CollectorStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.stats
}

// SetOutput 设置输出通道
func (sc *SystemdCollector) SetOutput(output chan<- *LogEntry) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.output = output
}

// collectLoop 收集循环
func (sc *SystemdCollector) collectLoop(ctx context.Context) {
	defer func() {
		sc.mu.Lock()
		sc.running = false
		sc.mu.Unlock()
	}()
	
	// 这里应该实现journalctl的调用和解析
	// 由于复杂性，这里只是示例框架
	sc.logger.Info("Systemd collection loop started")
	
	<-ctx.Done()
}

// LogProcessor 日志处理器
type LogProcessor struct {
	mu         sync.RWMutex
	processors []Processor
	logger     *zap.Logger
}

// Processor 处理器接口
type Processor interface {
	Process(entry *LogEntry) *LogEntry
	Name() string
}

// NewLogProcessor 创建日志处理器
func NewLogProcessor(logger *zap.Logger) *LogProcessor {
	return &LogProcessor{
		processors: make([]Processor, 0),
		logger:     logger,
	}
}

// AddProcessor 添加处理器
func (lp *LogProcessor) AddProcessor(processor Processor) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	lp.processors = append(lp.processors, processor)
	lp.logger.Info("Added log processor", zap.String("name", processor.Name()))
}

// Process 处理日志条目
func (lp *LogProcessor) Process(entry *LogEntry) *LogEntry {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	result := entry
	
	for _, processor := range lp.processors {
		if result == nil {
			break
		}
		result = processor.Process(result)
	}
	
	return result
}