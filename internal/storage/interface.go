package storage

import (
	"context"
	"time"
)

// ConfigItem 配置项结构
type ConfigItem struct {
	Key         string            `json:"key"`
	Value       interface{}       `json:"value"`
	Type        string            `json:"type"` // string, int, bool, json, yaml
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	Version     string            `json:"version"`
	Tags        map[string]string `json:"tags"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	CreatedBy   string            `json:"createdBy"`
	UpdatedBy   string            `json:"updatedBy"`
}

// ConfigStorage 配置存储接口
type ConfigStorage interface {
	// Get 获取配置项
	Get(ctx context.Context, service, environment, key string) (*ConfigItem, error)

	// Set 设置配置项
	Set(ctx context.Context, item *ConfigItem) error

	// Delete 删除配置项
	Delete(ctx context.Context, service, environment, key string) error

	// List 列出配置项
	List(ctx context.Context, service, environment string) ([]*ConfigItem, error)

	// GetByTags 根据标签获取配置项
	GetByTags(ctx context.Context, tags map[string]string) ([]*ConfigItem, error)

	// Watch 监听配置变化
	Watch(ctx context.Context, service, environment string) (<-chan *ConfigChangeEvent, error)

	// GetHistory 获取配置历史
	GetHistory(ctx context.Context, service, environment, key string, limit int) ([]*ConfigItem, error)

	// Backup 备份配置
	Backup(ctx context.Context, service, environment string) ([]byte, error)

	// Restore 恢复配置
	Restore(ctx context.Context, service, environment string, data []byte) error

	// Close 关闭存储
	Close() error
}

// ConfigChangeEvent 配置变化事件
type ConfigChangeEvent struct {
	Type        string      `json:"type"` // create, update, delete
	Service     string      `json:"service"`
	Environment string      `json:"environment"`
	Key         string      `json:"key"`
	OldValue    interface{} `json:"oldValue,omitempty"`
	NewValue    interface{} `json:"newValue,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
	User        string      `json:"user"`
}

// ConfigQuery 配置查询条件
type ConfigQuery struct {
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	KeyPattern  string            `json:"keyPattern"`
	Tags        map[string]string `json:"tags"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
}
