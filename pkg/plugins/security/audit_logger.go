package security

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// AuditLogLevel 审计日志级别
type AuditLogLevel int

const (
	AuditLevelInfo AuditLogLevel = iota
	AuditLevelWarn
	AuditLevelError
	AuditLevelCritical
)

func (level AuditLogLevel) String() string {
	switch level {
	case AuditLevelInfo:
		return "INFO"
	case AuditLevelWarn:
		return "WARN"
	case AuditLevelError:
		return "ERROR"
	case AuditLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// AuditEvent 审计事件
type AuditEvent struct {
	ID          string                 `json:"id"`
	Level       AuditLogLevel          `json:"level"`
	Category    string                 `json:"category"`    // authentication, authorization, plugin_management, system
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	UserID      string                 `json:"user_id,omitempty"`
	PluginID    string                 `json:"plugin_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Result      string                 `json:"result"`      // success, failure, denied, error
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source,omitempty"` // 事件来源
	Correlation string                 `json:"correlation,omitempty"` // 关联ID
}

// AuditLoggerConfig 审计日志配置
type AuditLoggerConfig struct {
	LogDir           string        `json:"log_dir"`
	MaxFileSize      int64         `json:"max_file_size"`      // 最大文件大小（字节）
	MaxFiles         int           `json:"max_files"`          // 最大文件数
	RetentionDays    int           `json:"retention_days"`     // 保留天数
	CompressOldFiles bool          `json:"compress_old_files"` // 压缩旧文件
	FlushInterval    time.Duration `json:"flush_interval"`     // 刷新间隔
	BufferSize       int           `json:"buffer_size"`        // 缓冲区大小
	EnableConsole    bool          `json:"enable_console"`     // 启用控制台输出
	MinLevel         AuditLogLevel `json:"min_level"`          // 最小日志级别
	EnableMetrics    bool          `json:"enable_metrics"`     // 启用指标收集
}

// AuditMetrics 审计指标
type AuditMetrics struct {
	TotalEvents      int64                    `json:"total_events"`
	EventsByLevel    map[string]int64         `json:"events_by_level"`
	EventsByCategory map[string]int64         `json:"events_by_category"`
	EventsByResult   map[string]int64         `json:"events_by_result"`
	EventsByHour     map[string]int64         `json:"events_by_hour"`
	TopUsers         map[string]int64         `json:"top_users"`
	TopPlugins       map[string]int64         `json:"top_plugins"`
	TopActions       map[string]int64         `json:"top_actions"`
	FailureRate      float64                  `json:"failure_rate"`
	LastUpdated      time.Time                `json:"last_updated"`
}

// AuditQuery 审计查询
type AuditQuery struct {
	StartTime    *time.Time             `json:"start_time,omitempty"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Level        *AuditLogLevel         `json:"level,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Action       string                 `json:"action,omitempty"`
	Resource     string                 `json:"resource,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	PluginID     string                 `json:"plugin_id,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	Result       string                 `json:"result,omitempty"`
	Source       string                 `json:"source,omitempty"`
	Correlation  string                 `json:"correlation,omitempty"`
	Keywords     []string               `json:"keywords,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	SortBy       string                 `json:"sort_by,omitempty"`
	SortOrder    string                 `json:"sort_order,omitempty"` // asc, desc
}

// AuditLogger 审计日志记录器
type AuditLogger struct {
	config      AuditLoggerConfig
	currentFile *os.File
	buffer      chan *AuditEvent
	metrics     *AuditMetrics
	mu          sync.RWMutex
	
	// 运行状态
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewAuditLogger 创建新的审计日志记录器
func NewAuditLogger(config AuditLoggerConfig) (*AuditLogger, error) {	
	// 设置默认值
	if config.LogDir == "" {
		config.LogDir = "./logs/audit"
	}
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxFiles <= 0 {
		config.MaxFiles = 10
	}
	if config.RetentionDays <= 0 {
		config.RetentionDays = 30
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	// 创建日志目录
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := &AuditLogger{
		config: config,
		buffer: make(chan *AuditEvent, config.BufferSize),
		metrics: &AuditMetrics{
			EventsByLevel:    make(map[string]int64),
			EventsByCategory: make(map[string]int64),
			EventsByResult:   make(map[string]int64),
			EventsByHour:     make(map[string]int64),
			TopUsers:         make(map[string]int64),
			TopPlugins:       make(map[string]int64),
			TopActions:       make(map[string]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return logger, nil
}

// Start 启动审计日志记录器
func (al *AuditLogger) Start() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.running {
		return fmt.Errorf("audit logger is already running")
	}

	// 打开当前日志文件
	if err := al.openCurrentFile(); err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	al.running = true

	// 启动工作协程
	al.wg.Add(3)
	go al.writeLoop()
	go al.flushLoop()
	go al.cleanupLoop()

	return nil
}

// Stop 停止审计日志记录器
func (al *AuditLogger) Stop() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if !al.running {
		return fmt.Errorf("audit logger is not running")
	}

	al.running = false
	al.cancel()

	// 等待所有协程结束
	al.wg.Wait()

	// 关闭当前文件
	if al.currentFile != nil {
		al.currentFile.Close()
		al.currentFile = nil
	}

	return nil
}

// Log 记录审计事件
func (al *AuditLogger) Log(event *AuditEvent) {
	if event.Level < al.config.MinLevel {
		return
	}

	if event.ID == "" {
		event.ID = al.generateEventID()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 非阻塞发送到缓冲区
	select {
	case al.buffer <- event:
		// 成功发送
	default:
		// 缓冲区满，丢弃事件（可以考虑记录到错误日志）
		fmt.Printf("Audit log buffer full, dropping event: %s\n", event.ID)
	}

	// 控制台输出
	if al.config.EnableConsole {
		al.printToConsole(event)
	}
}

// LogInfo 记录信息级别事件
func (al *AuditLogger) LogInfo(category, action, resource string, details map[string]interface{}) {
	event := &AuditEvent{
		Level:     AuditLevelInfo,
		Category:  category,
		Action:    action,
		Resource:  resource,
		Result:    "success",
		Details:   details,
		Timestamp: time.Now(),
	}
	al.Log(event)
}

// LogWarn 记录警告级别事件
func (al *AuditLogger) LogWarn(category, action, resource, message string, details map[string]interface{}) {
	event := &AuditEvent{
		Level:     AuditLevelWarn,
		Category:  category,
		Action:    action,
		Resource:  resource,
		Result:    "warning",
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
	al.Log(event)
}

// LogError 记录错误级别事件
func (al *AuditLogger) LogError(category, action, resource, message string, details map[string]interface{}) {
	event := &AuditEvent{
		Level:     AuditLevelError,
		Category:  category,
		Action:    action,
		Resource:  resource,
		Result:    "error",
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
	al.Log(event)
}

// LogCritical 记录严重级别事件
func (al *AuditLogger) LogCritical(category, action, resource, message string, details map[string]interface{}) {
	event := &AuditEvent{
		Level:     AuditLevelCritical,
		Category:  category,
		Action:    action,
		Resource:  resource,
		Result:    "critical",
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
	al.Log(event)
}

// Query 查询审计日志
func (al *AuditLogger) Query(query *AuditQuery) ([]*AuditEvent, int, error) {
	files, err := al.getLogFiles()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get log files: %v", err)
	}

	var events []*AuditEvent
	for _, file := range files {
		fileEvents, err := al.queryFile(file, query)
		if err != nil {
			continue // 跳过错误文件
		}
		events = append(events, fileEvents...)
	}

	// 排序
	al.sortEvents(events, query.SortBy, query.SortOrder)

	total := len(events)

	// 分页
	if query.Offset > 0 {
		if query.Offset >= len(events) {
			return []*AuditEvent{}, total, nil
		}
		events = events[query.Offset:]
	}

	if query.Limit > 0 && query.Limit < len(events) {
		events = events[:query.Limit]
	}

	return events, total, nil
}

// GetMetrics 获取审计指标
func (al *AuditLogger) GetMetrics() *AuditMetrics {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// 深拷贝指标
	metrics := &AuditMetrics{
		TotalEvents:      al.metrics.TotalEvents,
		EventsByLevel:    make(map[string]int64),
		EventsByCategory: make(map[string]int64),
		EventsByResult:   make(map[string]int64),
		EventsByHour:     make(map[string]int64),
		TopUsers:         make(map[string]int64),
		TopPlugins:       make(map[string]int64),
		TopActions:       make(map[string]int64),
		FailureRate:      al.metrics.FailureRate,
		LastUpdated:      al.metrics.LastUpdated,
	}

	for k, v := range al.metrics.EventsByLevel {
		metrics.EventsByLevel[k] = v
	}
	for k, v := range al.metrics.EventsByCategory {
		metrics.EventsByCategory[k] = v
	}
	for k, v := range al.metrics.EventsByResult {
		metrics.EventsByResult[k] = v
	}
	for k, v := range al.metrics.EventsByHour {
		metrics.EventsByHour[k] = v
	}
	for k, v := range al.metrics.TopUsers {
		metrics.TopUsers[k] = v
	}
	for k, v := range al.metrics.TopPlugins {
		metrics.TopPlugins[k] = v
	}
	for k, v := range al.metrics.TopActions {
		metrics.TopActions[k] = v
	}

	return metrics
}

// ExportLogs 导出日志
func (al *AuditLogger) ExportLogs(startTime, endTime time.Time, format string) ([]byte, error) {
	query := &AuditQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	events, _, err := al.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %v", err)
	}

	switch strings.ToLower(format) {
	case "json":
		return json.Marshal(events)
	case "csv":
		return al.exportToCSV(events)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// 私有方法

// writeLoop 写入循环
func (al *AuditLogger) writeLoop() {
	defer al.wg.Done()

	for {
		select {
		case <-al.ctx.Done():
			// 处理剩余事件
			for {
				select {
				case event := <-al.buffer:
					al.writeEvent(event)
				default:
					return
				}
			}
		case event := <-al.buffer:
			al.writeEvent(event)
		}
	}
}

// flushLoop 刷新循环
func (al *AuditLogger) flushLoop() {
	defer al.wg.Done()

	ticker := time.NewTicker(al.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-al.ctx.Done():
			return
		case <-ticker.C:
			al.flush()
		}
	}
}

// cleanupLoop 清理循环
func (al *AuditLogger) cleanupLoop() {
	defer al.wg.Done()

	ticker := time.NewTicker(24 * time.Hour) // 每天清理一次旧日志文件
	defer ticker.Stop()		

	for {
		select {
		case <-al.ctx.Done():
			return
		case <-ticker.C:
			al.cleanup()
		}
	}
}

// writeEvent 写入事件
func (al *AuditLogger) writeEvent(event *AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// 检查文件大小，必要时轮转文件
	if al.currentFile != nil {
		if stat, err := al.currentFile.Stat(); err == nil {
			if stat.Size() >= al.config.MaxFileSize {
				al.rotateFile()
			}
		}
	}

	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Failed to marshal audit event: %v\n", err)
		return
	}

	// 写入文件
	if al.currentFile != nil {
		al.currentFile.Write(data)
		al.currentFile.Write([]byte("\n"))
	}

	// 更新指标
	if al.config.EnableMetrics {
		al.updateMetrics(event)
	}
}

// flush 刷新缓冲区
func (al *AuditLogger) flush() {		
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.currentFile != nil {
		al.currentFile.Sync()
	}
}

// rotateFile 轮转文件
func (al *AuditLogger) rotateFile() {
	if al.currentFile != nil {
		al.currentFile.Close()
	}

	// 重命名当前文件为带时间戳的文件名
	currentPath := al.getCurrentFilePath()
	timestamp := time.Now().Format("20060102_150405")
	rotatedPath := filepath.Join(al.config.LogDir, fmt.Sprintf("audit_%s.log", timestamp))

	if err := os.Rename(currentPath, rotatedPath); err != nil {
		fmt.Printf("Failed to rotate log file: %v\n", err)
	}

	// 压缩旧文件（如果配置了）
	if al.config.CompressOldFiles {
		go al.compressFile(rotatedPath)
	}

	// 打开新文件
	al.openCurrentFile()

	// 清理旧文件（如果配置了）
	go al.cleanup()
}

// openCurrentFile 打开当前日志文件
func (al *AuditLogger) openCurrentFile() error {
	path := al.getCurrentFilePath()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	al.currentFile = file
	return nil
}

// getCurrentFilePath 获取当前文件路径
func (al *AuditLogger) getCurrentFilePath() string {
	return filepath.Join(al.config.LogDir, "audit_current.log")
}

// getLogFiles 获取所有日志文件（包括当前文件）
func (al *AuditLogger) getLogFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(al.config.LogDir, "audit_*.log*"))
	if err != nil {
		return nil, err
	}

	// 按修改时间排序（最新的在前面）
	sort.Slice(files, func(i, j int) bool {
		statI, _ := os.Stat(files[i])
		statJ, _ := os.Stat(files[j])
		return statI.ModTime().After(statJ.ModTime())
	})

	return files, nil
}

// queryFile 查询单个文件
func (al *AuditLogger) queryFile(filePath string, query *AuditQuery) ([]*AuditEvent, error) {
	var file io.ReadCloser
	var err error

	// 检查是否是压缩文件
	if strings.HasSuffix(filePath, ".gz") {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		file, err = gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.Open(filePath)
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	var events []*AuditEvent
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var event AuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue // 跳过无效行
		}

		if al.matchQuery(&event, query) {
			events = append(events, &event)
		}
	}

	return events, scanner.Err()
}

// matchQuery 匹配查询条件
func (al *AuditLogger) matchQuery(event *AuditEvent, query *AuditQuery) bool {
	if query.StartTime != nil && event.Timestamp.Before(*query.StartTime) {
		return false
	}

	if query.EndTime != nil && event.Timestamp.After(*query.EndTime) {
		return false
	}

	if query.Level != nil && event.Level != *query.Level {
		return false
	}

	if query.Category != "" && event.Category != query.Category {
		return false
	}

	if query.Action != "" && event.Action != query.Action {
		return false
	}

	if query.Resource != "" && event.Resource != query.Resource {
		return false
	}

	if query.UserID != "" && event.UserID != query.UserID {
		return false
	}

	if query.PluginID != "" && event.PluginID != query.PluginID {
		return false
	}

	if query.IPAddress != "" && event.IPAddress != query.IPAddress {
		return false
	}

	if query.Result != "" && event.Result != query.Result {
		return false
	}

	if query.Source != "" && event.Source != query.Source {
		return false
	}

	if query.Correlation != "" && event.Correlation != query.Correlation {
		return false
	}

	// 关键词搜索（对操作、资源、消息和详细信息进行搜索）
	if len(query.Keywords) > 0 {
		eventText := strings.ToLower(fmt.Sprintf("%s %s %s %s", event.Action, event.Resource, event.Message, event.Details))
		for _, keyword := range query.Keywords {
			if !strings.Contains(eventText, strings.ToLower(keyword)) {
				return false
			}
		}
	}

	return true
}

// sortEvents 排序事件
func (al *AuditLogger) sortEvents(events []*AuditEvent, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = "timestamp"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	sort.Slice(events, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "timestamp":
			less = events[i].Timestamp.Before(events[j].Timestamp)
		case "level":
			less = events[i].Level < events[j].Level
		case "category":
			less = events[i].Category < events[j].Category
		case "action":
			less = events[i].Action < events[j].Action
		default:
			less = events[i].Timestamp.Before(events[j].Timestamp)
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// updateMetrics 更新指标
func (al *AuditLogger) updateMetrics(event *AuditEvent) {
	al.metrics.TotalEvents++
	al.metrics.EventsByLevel[event.Level.String()]++
	al.metrics.EventsByCategory[event.Category]++
	al.metrics.EventsByResult[event.Result]++

	hour := event.Timestamp.Format("2006-01-02 15")
	al.metrics.EventsByHour[hour]++

	if event.UserID != "" {
		al.metrics.TopUsers[event.UserID]++
	}

	if event.PluginID != "" {
		al.metrics.TopPlugins[event.PluginID]++
	}

	al.metrics.TopActions[event.Action]++

	// 计算失败率（失败、错误和拒绝）
	failures := al.metrics.EventsByResult["failure"] + al.metrics.EventsByResult["error"] + al.metrics.EventsByResult["denied"]
	if al.metrics.TotalEvents > 0 {
		al.metrics.FailureRate = float64(failures) / float64(al.metrics.TotalEvents) * 100
	}

	al.metrics.LastUpdated = time.Now()
}

// compressFile 压缩文件
func (al *AuditLogger) compressFile(filePath string) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(filePath + ".gz")
	if err != nil {
		return
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, inputFile)
	if err != nil {
		return
	}

	// 删除原始文件	os.Remove(filePath)
}

// cleanup 清理旧文件（按保留天数和文件数量）
func (al *AuditLogger) cleanup() {
	files, err := al.getLogFiles()
	if err != nil {
		return
	}

	// 按保留天数清理旧文件	cutoff := time.Now().AddDate(0, 0, -al.config.RetentionDays)
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoff) {
			os.Remove(file)
		}
	}

	// 按文件数量清理旧文件	if len(files) > al.config.MaxFiles {
		for i := al.config.MaxFiles; i < len(files); i++ {
			os.Remove(files[i])
		}
	}
}

// printToConsole 打印到控制台
func (al *AuditLogger) printToConsole(event *AuditEvent) {
	timestamp := event.Timestamp.Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s [%s] %s.%s -> %s: %s\n",
		timestamp,
		event.Level.String(),
		event.Category,
		event.Resource,
		event.Action,
		event.Result,
		event.Message,
	)
}

// generateEventID 生成事件ID
func (al *AuditLogger) generateEventID() string {
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

// exportToCSV 导出为CSV格式
func (al *AuditLogger) exportToCSV(events []*AuditEvent) ([]byte, error) {
	var lines []string
	
	// CSV头部
	header := "ID,Timestamp,Level,Category,Action,Resource,UserID,PluginID,Result,Message,IPAddress"
	lines = append(lines, header)

	// 数据行
	for _, event := range events {
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s",
			event.ID,
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Level.String(),
			event.Category,
			event.Action,
			event.Resource,
			event.UserID,
			event.PluginID,
			event.Result,
			strings.ReplaceAll(event.Message, ",", ";"),
			event.IPAddress,
		)
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}
