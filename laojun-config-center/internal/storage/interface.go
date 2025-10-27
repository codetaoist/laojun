package storage

import (
	"context"
	"time"
)

// ConfigItem 配置项
type ConfigItem struct {
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Type        string                 `json:"type"` // string, json, yaml, toml
	Version     int64                  `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ConfigHistory 配置历史记录
type ConfigHistory struct {
	ID          int64                  `json:"id"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Key         string                 `json:"key"`
	OldValue    interface{}            `json:"old_value"`
	NewValue    interface{}            `json:"new_value"`
	Version     int64                  `json:"version"`
	Operation   string                 `json:"operation"` // create, update, delete
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	Reason      string                 `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SearchQuery 搜索查询
type SearchQuery struct {
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	Key         string            `json:"key"`
	Value       string            `json:"value"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
}

// WatchEvent 监听事件
type WatchEvent struct {
	Type        string      `json:"type"` // create, update, delete
	Service     string      `json:"service"`
	Environment string      `json:"environment"`
	Key         string      `json:"key"`
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Version     int64       `json:"version"`
	Timestamp   time.Time   `json:"timestamp"`
}

// ConfigStorage 配置存储接口
type ConfigStorage interface {
	// 基本操作
	Get(ctx context.Context, service, environment, key string) (*ConfigItem, error)
	Set(ctx context.Context, item *ConfigItem) error
	Delete(ctx context.Context, service, environment, key string) error
	List(ctx context.Context, service, environment string) ([]*ConfigItem, error)
	Exists(ctx context.Context, service, environment, key string) (bool, error)

	// 批量操作
	GetMultiple(ctx context.Context, keys []ConfigKey) ([]*ConfigItem, error)
	SetMultiple(ctx context.Context, items []*ConfigItem) error
	DeleteMultiple(ctx context.Context, keys []ConfigKey) error

	// 搜索
	Search(ctx context.Context, query *SearchQuery) ([]*ConfigItem, error)

	// 版本管理
	GetHistory(ctx context.Context, service, environment, key string, limit int) ([]*ConfigHistory, error)
	GetVersion(ctx context.Context, service, environment, key string, version int64) (*ConfigItem, error)
	Rollback(ctx context.Context, service, environment, key string, version int64, operator string) error

	// 监听
	Watch(ctx context.Context, service, environment string) (<-chan *WatchEvent, error)
	StopWatch(service, environment string)

	// 备份和恢复
	Backup(ctx context.Context, service, environment string) ([]byte, error)
	Restore(ctx context.Context, service, environment string, data []byte, operator string) error

	// 验证
	Validate(ctx context.Context, item *ConfigItem) error

	// 健康检查
	HealthCheck(ctx context.Context) error

	// 关闭
	Close() error
}

// ConfigKey 配置键
type ConfigKey struct {
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Key         string `json:"key"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ConfigNotFoundError 配置未找到错误
type ConfigNotFoundError struct {
	Service     string
	Environment string
	Key         string
}

func (e *ConfigNotFoundError) Error() string {
	return "config not found: " + e.Service + "/" + e.Environment + "/" + e.Key
}

// ConfigExistsError 配置已存在错误
type ConfigExistsError struct {
	Service     string
	Environment string
	Key         string
}

func (e *ConfigExistsError) Error() string {
	return "config already exists: " + e.Service + "/" + e.Environment + "/" + e.Key
}