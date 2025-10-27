package exporters

import (
	"context"
	"time"
)

// Exporter 导出器接口
type Exporter interface {
	// Name 返回导出器名称
	Name() string
	
	// Start 启动导出器
	Start(ctx context.Context) error
	
	// Stop 停止导出器
	Stop() error
	
	// IsHealthy 检查导出器健康状态
	IsHealthy() bool
	
	// IsReady 检查导出器就绪状态
	IsReady() bool
	
	// Export 导出数据
	Export(data interface{}) error
	
	// GetStats 获取导出器统计信息
	GetStats() ExporterStats
}



// ExporterConfig 导出器配置接口
type ExporterConfig interface {
	IsEnabled() bool
	GetName() string
	GetConfig() map[string]interface{}
	SetConfig(key string, value interface{})
}

// BatchExporter 批量导出器接口
type BatchExporter interface {
	Exporter
	
	// ExportBatch 批量导出数据
	ExportBatch(data []interface{}) error
	
	// GetBatchSize 获取批量大小
	GetBatchSize() int
	
	// SetBatchSize 设置批量大小
	SetBatchSize(size int)
	
	// GetFlushInterval 获取刷新间隔
	GetFlushInterval() time.Duration
	
	// SetFlushInterval 设置刷新间隔
	SetFlushInterval(interval time.Duration)
}

// MetricData 指标数据结构
type MetricData struct {
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventData 事件数据结构
type EventData struct {
	Name      string                 `json:"name"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TraceData 追踪数据结构
type TraceData struct {
	TraceID   string                 `json:"trace_id"`
	SpanID    string                 `json:"span_id"`
	Operation string                 `json:"operation"`
	Duration  time.Duration          `json:"duration"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}