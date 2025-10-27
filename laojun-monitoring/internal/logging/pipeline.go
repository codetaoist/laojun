package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Pipeline 日志处理管道
type Pipeline struct {
	name        string
	config      PipelineConfig
	logger      *zap.Logger
	
	// 组件
	collectors  []LogCollector
	processors  []Processor
	outputs     []LogOutput
	
	// 状态
	mu          sync.RWMutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	
	// 通道
	inputChan   chan *LogEntry
	outputChan  chan *LogEntry
	
	// 统计
	stats       *PipelineStats
	
	// 缓冲区
	buffer      *LogBuffer
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	Name        string                    `mapstructure:"name"`
	Description string                    `mapstructure:"description"`
	Enabled     bool                      `mapstructure:"enabled"`
	
	// 缓冲配置
	BufferSize     int           `mapstructure:"buffer_size"`
	FlushInterval  time.Duration `mapstructure:"flush_interval"`
	FlushThreshold int           `mapstructure:"flush_threshold"`
	
	// 处理配置
	Workers        int           `mapstructure:"workers"`
	Timeout        time.Duration `mapstructure:"timeout"`
	RetryAttempts  int           `mapstructure:"retry_attempts"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
	
	// 组件配置
	Collectors []CollectorConfig `mapstructure:"collectors"`
	Processors []ProcessorConfig `mapstructure:"processors"`
	Outputs    []OutputConfig    `mapstructure:"outputs"`
}

// CollectorConfig 收集器配置
type CollectorConfig struct {
	Type    string                 `mapstructure:"type"`
	Name    string                 `mapstructure:"name"`
	Enabled bool                   `mapstructure:"enabled"`
	Config  map[string]interface{} `mapstructure:"config"`
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	Type    string                 `mapstructure:"type"`
	Name    string                 `mapstructure:"name"`
	Enabled bool                   `mapstructure:"enabled"`
	Config  map[string]interface{} `mapstructure:"config"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Type    string                 `mapstructure:"type"`
	Name    string                 `mapstructure:"name"`
	Enabled bool                   `mapstructure:"enabled"`
	Config  map[string]interface{} `mapstructure:"config"`
}

// PipelineStats 管道统计
type PipelineStats struct {
	mu                sync.RWMutex
	StartTime         time.Time `json:"start_time"`
	ProcessedCount    int64     `json:"processed_count"`
	DroppedCount      int64     `json:"dropped_count"`
	ErrorCount        int64     `json:"error_count"`
	LastProcessedTime time.Time `json:"last_processed_time"`
	LastError         string    `json:"last_error"`
	
	// 性能统计
	ProcessingRate    float64       `json:"processing_rate"`    // 每秒处理数量
	AverageLatency    time.Duration `json:"average_latency"`    // 平均延迟
	BufferUtilization float64       `json:"buffer_utilization"` // 缓冲区利用率
}

// LogBuffer 日志缓冲区
type LogBuffer struct {
	mu       sync.Mutex
	entries  []*LogEntry
	maxSize  int
	flushCh  chan struct{}
}

// NewPipeline 创建新的管道
func NewPipeline(name string, config PipelineConfig, logger *zap.Logger) *Pipeline {
	// 设置默认值
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.FlushThreshold == 0 {
		config.FlushThreshold = 100
	}
	if config.Workers == 0 {
		config.Workers = 2
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Pipeline{
		name:       name,
		config:     config,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
		inputChan:  make(chan *LogEntry, config.BufferSize),
		outputChan: make(chan *LogEntry, config.BufferSize),
		stats:      &PipelineStats{StartTime: time.Now()},
		buffer: &LogBuffer{
			maxSize: config.FlushThreshold,
			flushCh: make(chan struct{}, 1),
		},
	}
}

// Start 启动管道
func (p *Pipeline) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.running {
		return fmt.Errorf("pipeline %s is already running", p.name)
	}
	
	if !p.config.Enabled {
		return fmt.Errorf("pipeline %s is disabled", p.name)
	}
	
	p.logger.Info("Starting pipeline", zap.String("name", p.name))
	
	// 初始化组件
	if err := p.initializeComponents(); err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}
	
	// 启动收集器
	for _, collector := range p.collectors {
		if err := collector.Start(p.ctx); err != nil {
			p.logger.Error("Failed to start collector", 
				zap.String("collector", collector.Name()), 
				zap.Error(err))
			continue
		}
		
		// 设置输出通道
		collector.SetOutput(p.inputChan)
	}
	
	// 启动处理工作器
	for i := 0; i < p.config.Workers; i++ {
		go p.processWorker(i)
	}
	
	// 启动输出工作器
	go p.outputWorker()
	
	// 启动缓冲区刷新器
	go p.bufferFlusher()
	
	// 启动统计更新器
	go p.statsUpdater()
	
	p.running = true
	p.stats.StartTime = time.Now()
	
	p.logger.Info("Pipeline started successfully", zap.String("name", p.name))
	
	return nil
}

// Stop 停止管道
func (p *Pipeline) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.running {
		return nil
	}
	
	p.logger.Info("Stopping pipeline", zap.String("name", p.name))
	
	// 取消上下文
	p.cancel()
	
	// 停止收集器
	for _, collector := range p.collectors {
		if err := collector.Stop(); err != nil {
			p.logger.Error("Failed to stop collector", 
				zap.String("collector", collector.Name()), 
				zap.Error(err))
		}
	}
	
	// 等待处理完成
	time.Sleep(100 * time.Millisecond)
	
	// 刷新缓冲区
	p.flushBuffer()
	
	// 输出不需要显式停止
	
	p.running = false
	
	p.logger.Info("Pipeline stopped", zap.String("name", p.name))
	
	return nil
}

// IsRunning 检查是否运行中
func (p *Pipeline) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// GetStats 获取统计信息
func (p *Pipeline) GetStats() *PipelineStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()
	
	// 创建副本
	stats := *p.stats
	return &stats
}

// GetName 获取管道名称
func (p *Pipeline) GetName() string {
	return p.name
}

// GetConfig 获取管道配置
func (p *Pipeline) GetConfig() PipelineConfig {
	return p.config
}

// initializeComponents 初始化组件
func (p *Pipeline) initializeComponents() error {
	// 初始化收集器
	for _, config := range p.config.Collectors {
		if !config.Enabled {
			continue
		}
		
		collector, err := p.createCollector(config)
		if err != nil {
			p.logger.Error("Failed to create collector", 
				zap.String("type", config.Type), 
				zap.String("name", config.Name), 
				zap.Error(err))
			continue
		}
		
		p.collectors = append(p.collectors, collector)
	}
	
	// 初始化处理器
	for _, config := range p.config.Processors {
		if !config.Enabled {
			continue
		}
		
		processor, err := p.createProcessor(config)
		if err != nil {
			p.logger.Error("Failed to create processor", 
				zap.String("type", config.Type), 
				zap.String("name", config.Name), 
				zap.Error(err))
			continue
		}
		
		p.processors = append(p.processors, processor)
	}
	
	// 初始化输出
	for _, config := range p.config.Outputs {
		if !config.Enabled {
			continue
		}
		
		output, err := p.createOutput(config)
		if err != nil {
			p.logger.Error("Failed to create output", 
				zap.String("type", config.Type), 
				zap.String("name", config.Name), 
				zap.Error(err))
			continue
		}
		
		p.outputs = append(p.outputs, output)
	}
	
	return nil
}

// createCollector 创建收集器
func (p *Pipeline) createCollector(config CollectorConfig) (LogCollector, error) {
	switch config.Type {
	case "file":
		// 这里需要根据配置创建文件收集器
		return nil, fmt.Errorf("file collector not implemented")
	case "systemd":
		// 这里需要根据配置创建systemd收集器
		return nil, fmt.Errorf("systemd collector not implemented")
	default:
		return nil, fmt.Errorf("unknown collector type: %s", config.Type)
	}
}

// createProcessor 创建处理器
func (p *Pipeline) createProcessor(config ProcessorConfig) (Processor, error) {
	switch config.Type {
	case "filter":
		// 这里需要根据配置创建过滤处理器
		return nil, fmt.Errorf("filter processor not implemented")
	case "enrich":
		// 这里需要根据配置创建丰富处理器
		return nil, fmt.Errorf("enrich processor not implemented")
	case "parse":
		// 这里需要根据配置创建解析处理器
		return nil, fmt.Errorf("parse processor not implemented")
	case "transform":
		// 这里需要根据配置创建转换处理器
		return nil, fmt.Errorf("transform processor not implemented")
	case "rate_limit":
		// 这里需要根据配置创建限流处理器
		return nil, fmt.Errorf("rate_limit processor not implemented")
	default:
		return nil, fmt.Errorf("unknown processor type: %s", config.Type)
	}
}

// createOutput 创建输出
func (p *Pipeline) createOutput(config OutputConfig) (LogOutput, error) {
	switch config.Type {
	case "file":
		// 这里需要根据配置创建文件输出
		return nil, fmt.Errorf("file output not implemented")
	case "console":
		// 这里需要根据配置创建控制台输出
		return nil, fmt.Errorf("console output not implemented")
	case "elasticsearch":
		// 这里需要根据配置创建Elasticsearch输出
		return nil, fmt.Errorf("elasticsearch output not implemented")
	default:
		return nil, fmt.Errorf("unknown output type: %s", config.Type)
	}
}

// processWorker 处理工作器
func (p *Pipeline) processWorker(id int) {
	p.logger.Debug("Starting process worker", zap.Int("worker_id", id))
	
	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("Process worker stopping", zap.Int("worker_id", id))
			return
		case entry := <-p.inputChan:
			if entry == nil {
				continue
			}
			
			// 处理日志条目
			processed := p.processEntry(entry)
			if processed != nil {
				// 发送到输出通道
				select {
				case p.outputChan <- processed:
					p.updateStats(true, false, nil)
				case <-p.ctx.Done():
					return
				default:
					// 输出通道满了，丢弃日志
					p.updateStats(false, true, fmt.Errorf("output channel full"))
				}
			} else {
				// 日志被丢弃
				p.updateStats(false, true, nil)
			}
		}
	}
}

// processEntry 处理日志条目
func (p *Pipeline) processEntry(entry *LogEntry) *LogEntry {
	result := entry
	
	// 应用所有处理器
	for _, processor := range p.processors {
		result = processor.Process(result)
		if result == nil {
			// 日志被处理器丢弃
			break
		}
	}
	
	return result
}

// outputWorker 输出工作器
func (p *Pipeline) outputWorker() {
	p.logger.Debug("Starting output worker")
	
	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("Output worker stopping")
			return
		case entry := <-p.outputChan:
			if entry == nil {
				continue
			}
			
			// 添加到缓冲区
			p.addToBuffer(entry)
		case <-p.buffer.flushCh:
			// 刷新缓冲区
			p.flushBuffer()
		}
	}
}

// addToBuffer 添加到缓冲区
func (p *Pipeline) addToBuffer(entry *LogEntry) {
	p.buffer.mu.Lock()
	defer p.buffer.mu.Unlock()
	
	p.buffer.entries = append(p.buffer.entries, entry)
	
	// 检查是否需要刷新
	if len(p.buffer.entries) >= p.buffer.maxSize {
		select {
		case p.buffer.flushCh <- struct{}{}:
		default:
		}
	}
}

// flushBuffer 刷新缓冲区
func (p *Pipeline) flushBuffer() {
	p.buffer.mu.Lock()
	entries := make([]*LogEntry, len(p.buffer.entries))
	copy(entries, p.buffer.entries)
	p.buffer.entries = p.buffer.entries[:0]
	p.buffer.mu.Unlock()
	
	if len(entries) == 0 {
		return
	}
	
	// 发送到所有输出
	for _, output := range p.outputs {
		if err := output.Write(p.ctx, entries); err != nil {
			p.logger.Error("Failed to write to output", 
				zap.String("output", output.Name()), 
				zap.Error(err))
			p.updateStats(false, false, err)
		}
	}
}

// bufferFlusher 缓冲区刷新器
func (p *Pipeline) bufferFlusher() {
	ticker := time.NewTicker(p.config.FlushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			select {
			case p.buffer.flushCh <- struct{}{}:
			default:
			}
		}
	}
}

// statsUpdater 统计更新器
func (p *Pipeline) statsUpdater() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	var lastProcessed int64
	var lastTime time.Time = time.Now()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.stats.mu.Lock()
			
			// 计算处理速率
			now := time.Now()
			duration := now.Sub(lastTime).Seconds()
			if duration > 0 {
				processed := p.stats.ProcessedCount - lastProcessed
				p.stats.ProcessingRate = float64(processed) / duration
				lastProcessed = p.stats.ProcessedCount
				lastTime = now
			}
			
			// 计算缓冲区利用率
			p.buffer.mu.Lock()
			bufferSize := len(p.buffer.entries)
			p.buffer.mu.Unlock()
			
			p.stats.BufferUtilization = float64(bufferSize) / float64(p.config.BufferSize) * 100
			
			p.stats.mu.Unlock()
		}
	}
}

// updateStats 更新统计信息
func (p *Pipeline) updateStats(processed, dropped bool, err error) {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()
	
	if processed {
		p.stats.ProcessedCount++
		p.stats.LastProcessedTime = time.Now()
	}
	
	if dropped {
		p.stats.DroppedCount++
	}
	
	if err != nil {
		p.stats.ErrorCount++
		p.stats.LastError = err.Error()
	}
}

// PipelineManager 管道管理器
type PipelineManager struct {
	mu        sync.RWMutex
	pipelines map[string]*Pipeline
	logger    *zap.Logger
}

// NewPipelineManager 创建管道管理器
func NewPipelineManager(logger *zap.Logger) *PipelineManager {
	return &PipelineManager{
		pipelines: make(map[string]*Pipeline),
		logger:    logger,
	}
}

// AddPipeline 添加管道
func (pm *PipelineManager) AddPipeline(pipeline *Pipeline) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	name := pipeline.GetName()
	if _, exists := pm.pipelines[name]; exists {
		return fmt.Errorf("pipeline %s already exists", name)
	}
	
	pm.pipelines[name] = pipeline
	pm.logger.Info("Pipeline added", zap.String("name", name))
	
	return nil
}

// RemovePipeline 移除管道
func (pm *PipelineManager) RemovePipeline(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pipeline, exists := pm.pipelines[name]
	if !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}
	
	// 停止管道
	if pipeline.IsRunning() {
		if err := pipeline.Stop(); err != nil {
			return fmt.Errorf("failed to stop pipeline %s: %w", name, err)
		}
	}
	
	delete(pm.pipelines, name)
	pm.logger.Info("Pipeline removed", zap.String("name", name))
	
	return nil
}

// GetPipeline 获取管道
func (pm *PipelineManager) GetPipeline(name string) (*Pipeline, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	pipeline, exists := pm.pipelines[name]
	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", name)
	}
	
	return pipeline, nil
}

// ListPipelines 列出所有管道
func (pm *PipelineManager) ListPipelines() []*Pipeline {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	pipelines := make([]*Pipeline, 0, len(pm.pipelines))
	for _, pipeline := range pm.pipelines {
		pipelines = append(pipelines, pipeline)
	}
	
	return pipelines
}

// StartAll 启动所有管道
func (pm *PipelineManager) StartAll() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	for name, pipeline := range pm.pipelines {
		if !pipeline.IsRunning() {
			if err := pipeline.Start(); err != nil {
				pm.logger.Error("Failed to start pipeline", 
					zap.String("name", name), 
					zap.Error(err))
			}
		}
	}
	
	return nil
}

// StopAll 停止所有管道
func (pm *PipelineManager) StopAll() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	for name, pipeline := range pm.pipelines {
		if pipeline.IsRunning() {
			if err := pipeline.Stop(); err != nil {
				pm.logger.Error("Failed to stop pipeline", 
					zap.String("name", name), 
					zap.Error(err))
			}
		}
	}
	
	return nil
}

// GetStats 获取所有管道统计
func (pm *PipelineManager) GetStats() map[string]*PipelineStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	stats := make(map[string]*PipelineStats)
	for name, pipeline := range pm.pipelines {
		stats[name] = pipeline.GetStats()
	}
	
	return stats
}