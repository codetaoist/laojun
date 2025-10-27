package query

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MemoryQuerier 内存指标查询器
// 用于查询收集器收集的指标数据
type MemoryQuerier struct {
	mu      sync.RWMutex
	metrics map[string]*MetricSeries
	logger  *logrus.Logger
}

// MetricSeries 指标时间序列
type MetricSeries struct {
	Name      string                 `json:"name"`
	Labels    map[string]string      `json:"labels"`
	Values    []MetricValue          `json:"values"`
	LastValue float64                `json:"last_value"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// MetricValue 指标值
type MetricValue struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// NewMemoryQuerier 创建内存查询器
func NewMemoryQuerier(logger *logrus.Logger) *MemoryQuerier {
	return &MemoryQuerier{
		metrics: make(map[string]*MetricSeries),
		logger:  logger,
	}
}

// Query 执行指标查询
func (mq *MemoryQuerier) Query(ctx context.Context, query string) (*QueryResponse, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	// 解析查询语句
	metricName, operator, threshold, err := mq.parseQuery(query)
	if err != nil {
		return &QueryResponse{Error: err.Error()}, err
	}

	var results []QueryResult
	
	// 查找匹配的指标
	for key, series := range mq.metrics {
		if strings.Contains(key, metricName) {
			// 获取最新值
			if len(series.Values) > 0 {
				latestValue := series.Values[len(series.Values)-1]
				
				// 应用条件判断
				if mq.evaluateCondition(latestValue.Value, operator, threshold) {
					results = append(results, QueryResult{
						Value:     latestValue.Value,
						Labels:    series.Labels,
						Timestamp: latestValue.Timestamp,
					})
				}
			}
		}
	}

	return &QueryResponse{Results: results}, nil
}

// QueryRange 执行范围查询
func (mq *MemoryQuerier) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResponse, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	metricName, _, _, err := mq.parseQuery(query)
	if err != nil {
		return &QueryResponse{Error: err.Error()}, err
	}

	var results []QueryResult
	
	// 查找匹配的指标
	for key, series := range mq.metrics {
		if strings.Contains(key, metricName) {
			// 过滤时间范围内的值
			for _, value := range series.Values {
				if value.Timestamp.After(start) && value.Timestamp.Before(end) {
					results = append(results, QueryResult{
						Value:     value.Value,
						Labels:    series.Labels,
						Timestamp: value.Timestamp,
					})
				}
			}
		}
	}

	return &QueryResponse{Results: results}, nil
}

// AddMetric 添加指标数据
func (mq *MemoryQuerier) AddMetric(name string, value float64, labels map[string]string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	key := mq.buildMetricKey(name, labels)
	
	series, exists := mq.metrics[key]
	if !exists {
		series = &MetricSeries{
			Name:   name,
			Labels: labels,
			Values: make([]MetricValue, 0),
		}
		mq.metrics[key] = series
	}

	// 添加新值
	metricValue := MetricValue{
		Value:     value,
		Timestamp: time.Now(),
	}
	
	series.Values = append(series.Values, metricValue)
	series.LastValue = value
	series.UpdatedAt = time.Now()

	// 保持最近1000个值
	if len(series.Values) > 1000 {
		series.Values = series.Values[len(series.Values)-1000:]
	}
}

// parseQuery 解析查询语句
// 支持简单的查询格式: metric_name > threshold, metric_name < threshold, metric_name = threshold
func (mq *MemoryQuerier) parseQuery(query string) (string, string, float64, error) {
	query = strings.TrimSpace(query)
	
	// 支持的操作符
	operators := []string{">=", "<=", "!=", ">", "<", "="}
	
	for _, op := range operators {
		if strings.Contains(query, op) {
			parts := strings.Split(query, op)
			if len(parts) != 2 {
				continue
			}
			
			metricName := strings.TrimSpace(parts[0])
			thresholdStr := strings.TrimSpace(parts[1])
			
			threshold, err := strconv.ParseFloat(thresholdStr, 64)
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid threshold value: %s", thresholdStr)
			}
			
			return metricName, op, threshold, nil
		}
	}
	
	// 如果没有操作符，返回指标名称
	return query, "", 0, nil
}

// evaluateCondition 评估条件
func (mq *MemoryQuerier) evaluateCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "=":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return true // 没有条件时返回true
	}
}

// buildMetricKey 构建指标键
func (mq *MemoryQuerier) buildMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	
	var labelPairs []string
	for k, v := range labels {
		labelPairs = append(labelPairs, fmt.Sprintf("%s=%s", k, v))
	}
	
	return fmt.Sprintf("%s{%s}", name, strings.Join(labelPairs, ","))
}

// IsHealthy 检查查询器是否健康
func (mq *MemoryQuerier) IsHealthy() bool {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	
	// 检查是否有指标数据
	return len(mq.metrics) > 0
}

// Close 关闭查询器
func (mq *MemoryQuerier) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	// 清理指标数据
	mq.metrics = make(map[string]*MetricSeries)
	return nil
}

// GetMetrics 获取所有指标（用于调试）
func (mq *MemoryQuerier) GetMetrics() map[string]*MetricSeries {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	
	result := make(map[string]*MetricSeries)
	for k, v := range mq.metrics {
		result[k] = v
	}
	return result
}