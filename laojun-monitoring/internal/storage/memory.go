package storage

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	mu      sync.RWMutex
	data    map[string][]MetricSample // key: metric_name + labels hash
	labels  map[string]map[string]bool // label_name -> {label_value: true}
	metrics map[string]bool            // metric_name -> true
	config  MemoryStorageConfig
}

// MemoryStorageConfig 内存存储配置
type MemoryStorageConfig struct {
	MaxSamples   int           `mapstructure:"max_samples"`
	MaxSeries    int           `mapstructure:"max_series"`
	Retention    time.Duration `mapstructure:"retention"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// MemoryStorageFactory 内存存储工厂
type MemoryStorageFactory struct{}

// Create 创建内存存储实例
func (f *MemoryStorageFactory) Create(config StorageConfig) (Storage, error) {
	memConfig := MemoryStorageConfig{
		MaxSamples:      10000,
		MaxSeries:       1000,
		Retention:       24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	// 解析配置
	if maxSamples, ok := config.Config["max_samples"].(int); ok {
		memConfig.MaxSamples = maxSamples
	}
	if maxSeries, ok := config.Config["max_series"].(int); ok {
		memConfig.MaxSeries = maxSeries
	}
	if retention, ok := config.Config["retention"].(string); ok {
		if d, err := time.ParseDuration(retention); err == nil {
			memConfig.Retention = d
		}
	}

	return NewMemoryStorage(memConfig), nil
}

// SupportedTypes 支持的存储类型
func (f *MemoryStorageFactory) SupportedTypes() []string {
	return []string{"memory"}
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage(config MemoryStorageConfig) *MemoryStorage {
	storage := &MemoryStorage{
		data:    make(map[string][]MetricSample),
		labels:  make(map[string]map[string]bool),
		metrics: make(map[string]bool),
		config:  config,
	}

	// 启动清理任务
	go storage.startCleanupTask()

	return storage
}

// Write 写入指标数据
func (s *MemoryStorage) Write(ctx context.Context, samples []MetricSample) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sample := range samples {
		key := s.generateKey(sample.MetricName, sample.Labels)
		
		// 检查系列数量限制
		if len(s.data) >= s.config.MaxSeries && s.data[key] == nil {
			continue // 跳过新系列
		}

		// 添加样本
		if s.data[key] == nil {
			s.data[key] = make([]MetricSample, 0)
		}
		
		s.data[key] = append(s.data[key], sample)
		
		// 检查样本数量限制
		if len(s.data[key]) > s.config.MaxSamples {
			// 删除最旧的样本
			s.data[key] = s.data[key][1:]
		}

		// 更新标签和指标索引
		s.updateIndexes(sample)
	}

	return nil
}

// Query 查询指标数据
func (s *MemoryStorage) Query(ctx context.Context, query string, timestamp time.Time) ([]QueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 简单的查询实现，支持基本的指标名称匹配
	metricName := strings.TrimSpace(query)
	results := make([]QueryResult, 0)

	for key, samples := range s.data {
		if !strings.Contains(key, metricName) {
			continue
		}

		// 查找最接近时间戳的样本
		var closestSample *MetricSample
		minDiff := time.Duration(1<<63 - 1) // 最大时间差

		for _, sample := range samples {
			diff := timestamp.Sub(sample.Timestamp)
			if diff < 0 {
				diff = -diff
			}
			if diff < minDiff {
				minDiff = diff
				closestSample = &sample
			}
		}

		if closestSample != nil {
			results = append(results, QueryResult{
				MetricName: closestSample.MetricName,
				Labels:     closestSample.Labels,
				Samples:    []MetricSample{*closestSample},
			})
		}
	}

	return results, nil
}

// QueryRange 范围查询
func (s *MemoryStorage) QueryRange(ctx context.Context, query string, queryRange QueryRange) ([]QueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metricName := strings.TrimSpace(query)
	results := make([]QueryResult, 0)

	for key, samples := range s.data {
		if !strings.Contains(key, metricName) {
			continue
		}

		// 过滤时间范围内的样本
		var filteredSamples []MetricSample
		for _, sample := range samples {
			if sample.Timestamp.After(queryRange.Start) && sample.Timestamp.Before(queryRange.End) {
				filteredSamples = append(filteredSamples, sample)
			}
		}

		if len(filteredSamples) > 0 {
			// 按时间戳排序
			sort.Slice(filteredSamples, func(i, j int) bool {
				return filteredSamples[i].Timestamp.Before(filteredSamples[j].Timestamp)
			})

			results = append(results, QueryResult{
				MetricName: filteredSamples[0].MetricName,
				Labels:     filteredSamples[0].Labels,
				Samples:    filteredSamples,
			})
		}
	}

	return results, nil
}

// LabelValues 获取标签值
func (s *MemoryStorage) LabelValues(ctx context.Context, labelName string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values, exists := s.labels[labelName]
	if !exists {
		return []string{}, nil
	}

	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}

	sort.Strings(result)
	return result, nil
}

// LabelNames 获取标签名称
func (s *MemoryStorage) LabelNames(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.labels))
	for name := range s.labels {
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

// MetricNames 获取指标名称
func (s *MemoryStorage) MetricNames(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.metrics))
	for name := range s.metrics {
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

// DeleteExpiredData 删除过期数据
func (s *MemoryStorage) DeleteExpiredData(ctx context.Context, before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, samples := range s.data {
		var validSamples []MetricSample
		for _, sample := range samples {
			if sample.Timestamp.After(before) {
				validSamples = append(validSamples, sample)
			}
		}

		if len(validSamples) == 0 {
			delete(s.data, key)
		} else {
			s.data[key] = validSamples
		}
	}

	// 重建索引
	s.rebuildIndexes()

	return nil
}

// Health 健康检查
func (s *MemoryStorage) Health(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查内存使用情况
	totalSamples := 0
	for _, samples := range s.data {
		totalSamples += len(samples)
	}

	if totalSamples > s.config.MaxSamples*len(s.data) {
		return NewStorageError("health_check", "memory", 
			errors.New("memory usage too high"))
	}

	return nil
}

// Close 关闭存储
func (s *MemoryStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空数据
	s.data = make(map[string][]MetricSample)
	s.labels = make(map[string]map[string]bool)
	s.metrics = make(map[string]bool)

	return nil
}

// generateKey 生成存储键
func (s *MemoryStorage) generateKey(metricName string, labels map[string]string) string {
	if len(labels) == 0 {
		return metricName
	}

	// 按标签名排序以确保一致性
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	parts = append(parts, metricName)
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}

	return strings.Join(parts, "|")
}

// updateIndexes 更新索引
func (s *MemoryStorage) updateIndexes(sample MetricSample) {
	// 更新指标索引
	s.metrics[sample.MetricName] = true

	// 更新标签索引
	for labelName, labelValue := range sample.Labels {
		if s.labels[labelName] == nil {
			s.labels[labelName] = make(map[string]bool)
		}
		s.labels[labelName][labelValue] = true
	}
}

// rebuildIndexes 重建索引
func (s *MemoryStorage) rebuildIndexes() {
	s.labels = make(map[string]map[string]bool)
	s.metrics = make(map[string]bool)

	for _, samples := range s.data {
		for _, sample := range samples {
			s.updateIndexes(sample)
		}
	}
}

// startCleanupTask 启动清理任务
func (s *MemoryStorage) startCleanupTask() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-s.config.Retention)
		s.DeleteExpiredData(context.Background(), cutoff)
	}
}

// GetStats 获取存储统计信息
func (s *MemoryStorage) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalSamples := 0
	for _, samples := range s.data {
		totalSamples += len(samples)
	}

	return map[string]interface{}{
		"type":          "memory",
		"series_count":  len(s.data),
		"samples_count": totalSamples,
		"labels_count":  len(s.labels),
		"metrics_count": len(s.metrics),
		"max_samples":   s.config.MaxSamples,
		"max_series":    s.config.MaxSeries,
	}
}