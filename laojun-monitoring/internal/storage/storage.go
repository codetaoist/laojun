package storage

import (
	"context"
	"time"
)

// MetricSample 指标样本
type MetricSample struct {
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	MetricName string           `json:"metric_name"`
}

// QueryRange 查询范围
type QueryRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Step  time.Duration `json:"step"`
}

// QueryResult 查询结果
type QueryResult struct {
	MetricName string          `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Samples    []MetricSample   `json:"samples"`
}

// Storage 存储接口
type Storage interface {
	// 写入指标数据
	Write(ctx context.Context, samples []MetricSample) error
	
	// 查询指标数据
	Query(ctx context.Context, query string, timestamp time.Time) ([]QueryResult, error)
	
	// 范围查询
	QueryRange(ctx context.Context, query string, queryRange QueryRange) ([]QueryResult, error)
	
	// 获取标签值
	LabelValues(ctx context.Context, labelName string) ([]string, error)
	
	// 获取标签名称
	LabelNames(ctx context.Context) ([]string, error)
	
	// 获取指标名称
	MetricNames(ctx context.Context) ([]string, error)
	
	// 删除过期数据
	DeleteExpiredData(ctx context.Context, before time.Time) error
	
	// 健康检查
	Health(ctx context.Context) error
	
	// 关闭存储
	Close() error
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string                 `mapstructure:"type"`
	Config     map[string]interface{} `mapstructure:"config"`
	Retention  time.Duration          `mapstructure:"retention"`
	BatchSize  int                    `mapstructure:"batch_size"`
	FlushInterval time.Duration       `mapstructure:"flush_interval"`
}

// StorageFactory 存储工厂
type StorageFactory interface {
	Create(config StorageConfig) (Storage, error)
	SupportedTypes() []string
}

// Registry 存储注册表
type Registry struct {
	factories map[string]StorageFactory
}

// NewRegistry 创建存储注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]StorageFactory),
	}
}

// Register 注册存储工厂
func (r *Registry) Register(storageType string, factory StorageFactory) {
	r.factories[storageType] = factory
}

// Create 创建存储实例
func (r *Registry) Create(config StorageConfig) (Storage, error) {
	factory, exists := r.factories[config.Type]
	if !exists {
		return nil, ErrUnsupportedStorageType
	}
	
	return factory.Create(config)
}

// SupportedTypes 获取支持的存储类型
func (r *Registry) SupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for storageType := range r.factories {
		types = append(types, storageType)
	}
	return types
}

// MetricWriter 指标写入器
type MetricWriter struct {
	storage   Storage
	batchSize int
	buffer    []MetricSample
	flushInterval time.Duration
	stopCh    chan struct{}
}

// NewMetricWriter 创建指标写入器
func NewMetricWriter(storage Storage, batchSize int, flushInterval time.Duration) *MetricWriter {
	return &MetricWriter{
		storage:       storage,
		batchSize:     batchSize,
		buffer:        make([]MetricSample, 0, batchSize),
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
	}
}

// Start 启动写入器
func (w *MetricWriter) Start(ctx context.Context) {
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			w.flush(ctx)
			return
		case <-ticker.C:
			w.flush(ctx)
		case <-w.stopCh:
			w.flush(ctx)
			return
		}
	}
}

// Write 写入指标样本
func (w *MetricWriter) Write(sample MetricSample) {
	w.buffer = append(w.buffer, sample)
	
	if len(w.buffer) >= w.batchSize {
		w.flush(context.Background())
	}
}

// flush 刷新缓冲区
func (w *MetricWriter) flush(ctx context.Context) {
	if len(w.buffer) == 0 {
		return
	}
	
	if err := w.storage.Write(ctx, w.buffer); err != nil {
		// 记录错误，但不阻塞
		// TODO: 添加错误处理逻辑
	}
	
	w.buffer = w.buffer[:0]
}

// Stop 停止写入器
func (w *MetricWriter) Stop() {
	close(w.stopCh)
}

// MetricReader 指标读取器
type MetricReader struct {
	storage Storage
}

// NewMetricReader 创建指标读取器
func NewMetricReader(storage Storage) *MetricReader {
	return &MetricReader{
		storage: storage,
	}
}

// Query 查询指标
func (r *MetricReader) Query(ctx context.Context, query string, timestamp time.Time) ([]QueryResult, error) {
	return r.storage.Query(ctx, query, timestamp)
}

// QueryRange 范围查询
func (r *MetricReader) QueryRange(ctx context.Context, query string, queryRange QueryRange) ([]QueryResult, error) {
	return r.storage.QueryRange(ctx, query, queryRange)
}

// GetLabelValues 获取标签值
func (r *MetricReader) GetLabelValues(ctx context.Context, labelName string) ([]string, error) {
	return r.storage.LabelValues(ctx, labelName)
}

// GetLabelNames 获取标签名称
func (r *MetricReader) GetLabelNames(ctx context.Context) ([]string, error) {
	return r.storage.LabelNames(ctx)
}

// GetMetricNames 获取指标名称
func (r *MetricReader) GetMetricNames(ctx context.Context) ([]string, error) {
	return r.storage.MetricNames(ctx)
}

// StorageManager 存储管理器
type StorageManager struct {
	storage Storage
	writer  *MetricWriter
	reader  *MetricReader
	config  StorageConfig
}

// NewStorageManager 创建存储管理器
func NewStorageManager(storage Storage, config StorageConfig) *StorageManager {
	writer := NewMetricWriter(storage, config.BatchSize, config.FlushInterval)
	reader := NewMetricReader(storage)
	
	return &StorageManager{
		storage: storage,
		writer:  writer,
		reader:  reader,
		config:  config,
	}
}

// Start 启动存储管理器
func (m *StorageManager) Start(ctx context.Context) {
	go m.writer.Start(ctx)
	
	// 启动数据清理任务
	if m.config.Retention > 0 {
		go m.startCleanupTask(ctx)
	}
}

// startCleanupTask 启动清理任务
func (m *StorageManager) startCleanupTask(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // 每天清理一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-m.config.Retention)
			if err := m.storage.DeleteExpiredData(ctx, cutoff); err != nil {
				// 记录错误但继续运行
				// TODO: 添加错误处理逻辑
			}
		}
	}
}

// Write 写入指标
func (m *StorageManager) Write(sample MetricSample) {
	m.writer.Write(sample)
}

// Query 查询指标
func (m *StorageManager) Query(ctx context.Context, query string, timestamp time.Time) ([]QueryResult, error) {
	return m.reader.Query(ctx, query, timestamp)
}

// QueryRange 范围查询
func (m *StorageManager) QueryRange(ctx context.Context, query string, queryRange QueryRange) ([]QueryResult, error) {
	return m.reader.QueryRange(ctx, query, queryRange)
}

// GetStorage 获取存储实例
func (m *StorageManager) GetStorage() Storage {
	return m.storage
}

// Stop 停止存储管理器
func (m *StorageManager) Stop() {
	m.writer.Stop()
	m.storage.Close()
}

// Health 健康检查
func (m *StorageManager) Health(ctx context.Context) error {
	return m.storage.Health(ctx)
}

// IsHealthy 检查存储是否健康
func (m *StorageManager) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Health(ctx) == nil
}

// StorageStats 存储统计信息
type StorageStats struct {
	MetricsWritten int64 `json:"metrics_written"`
	QueriesExecuted int64 `json:"queries_executed"`
	ErrorCount     int64 `json:"error_count"`
	LastWrite      time.Time `json:"last_write"`
	LastQuery      time.Time `json:"last_query"`
}

// GetStats 获取存储统计信息
func (m *StorageManager) GetStats() StorageStats {
	// 这里返回基本的统计信息，实际实现中可以从writer和reader获取更详细的统计
	return StorageStats{
		MetricsWritten:  0, // TODO: 从writer获取
		QueriesExecuted: 0, // TODO: 从reader获取
		ErrorCount:      0, // TODO: 统计错误数量
		LastWrite:       time.Time{}, // TODO: 记录最后写入时间
		LastQuery:       time.Time{}, // TODO: 记录最后查询时间
	}
}