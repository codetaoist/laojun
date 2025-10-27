package query

import (
	"context"
	"time"
)

// QueryResult 查询结果
type QueryResult struct {
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Results []QueryResult `json:"results"`
	Error   string        `json:"error,omitempty"`
}

// MetricQuerier 指标查询器接口
type MetricQuerier interface {
	// Query 执行指标查询
	Query(ctx context.Context, query string) (*QueryResponse, error)
	
	// QueryRange 执行范围查询
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryResponse, error)
	
	// IsHealthy 检查查询器是否健康
	IsHealthy() bool
	
	// Close 关闭查询器
	Close() error
}

// QueryConfig 查询配置
type QueryConfig struct {
	Type     string            `yaml:"type" json:"type"`         // prometheus, influxdb, custom
	Endpoint string            `yaml:"endpoint" json:"endpoint"` // 查询端点
	Timeout  time.Duration     `yaml:"timeout" json:"timeout"`   // 查询超时
	Headers  map[string]string `yaml:"headers" json:"headers"`   // 请求头
	Auth     AuthConfig        `yaml:"auth" json:"auth"`         // 认证配置
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type     string `yaml:"type" json:"type"`         // basic, bearer, none
	Username string `yaml:"username" json:"username"` // 用户名
	Password string `yaml:"password" json:"password"` // 密码
	Token    string `yaml:"token" json:"token"`       // Token
}