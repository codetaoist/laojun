package collectors

import "time"

// Collector 通用收集器接口
type Collector interface {
	// Start 启动收集器
	Start() error
	
	// Stop 停止收集器
	Stop() error
	
	// IsRunning 检查收集器是否运行中
	IsRunning() bool
	
	// GetStats 获取收集器统计信息
	GetStats() CollectorStats
	
	// Health 检查收集器健康状态
	Health() map[string]interface{}
	
	// Name 获取收集器名称
	Name() string
	
	// IsHealthy 检查收集器是否健康
	IsHealthy() bool
	
	// IsReady 检查收集器是否准备就绪
	IsReady() bool
}

// CollectorStats 收集器统计信息
type CollectorStats struct {
	CollectCount    int64     `json:"collect_count"`
	LastCollectTime time.Time `json:"last_collect_time"`
	ErrorCount      int64     `json:"error_count"`
	LastError       string    `json:"last_error,omitempty"`
}