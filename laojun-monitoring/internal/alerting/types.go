package alerting

import (
	"context"
	"time"
)

// Alert状态常量
const (
	AlertStateFiring   = "firing"
	AlertStateResolved = "resolved"
	AlertStateSilenced = "silenced"
)

// Alert 告警实例
type Alert struct {
	ID          string            `json:"id"`
	RuleID      string            `json:"rule_id"`
	RuleName    string            `json:"rule_name"`
	Name        string            `json:"name"`
	State       string            `json:"state"`       // firing, resolved, silenced
	Status      string            `json:"status"`      // 兼容性字段
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	ActiveAt    time.Time         `json:"active_at"`
	StartsAt    time.Time         `json:"starts_at"`
	EndsAt      *time.Time        `json:"ends_at,omitempty"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Expr        string            `json:"expr"`        // 表达式
	For         time.Duration     `json:"for"`         // 持续时间
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Silence 静默规则
type Silence struct {
	ID        string            `json:"id"`
	Matchers  map[string]string `json:"matchers"`
	StartsAt  time.Time         `json:"starts_at"`
	EndsAt    time.Time         `json:"ends_at"`
	CreatedBy string            `json:"created_by"`
	Comment   string            `json:"comment"`
}

// Receiver 接收器
type Receiver struct {
	Name     string                 `yaml:"name" json:"name"`
	Type     string                 `yaml:"type" json:"type"`
	Config   map[string]interface{} `yaml:"config" json:"config"`
	Enabled  bool                   `yaml:"enabled" json:"enabled"`
}

// Querier 查询器接口
type Querier interface {
	Query(ctx context.Context, query string, timestamp time.Time) (QueryResult, error)
}

// QueryResult 查询结果
type QueryResult struct {
	Type   string      `json:"type"`
	Result interface{} `json:"result"`
}

// RuleEvaluationStats 规则评估统计
type RuleEvaluationStats struct {
	TotalEvaluations int64         `json:"total_evaluations"`
	LastEvaluation   time.Time     `json:"last_evaluation"`
	EvaluationTime   time.Duration `json:"evaluation_time"`
	Errors           int64         `json:"errors"`
	LastError        string        `json:"last_error"`
	LastErrorTime    time.Time     `json:"last_error_time"`
}