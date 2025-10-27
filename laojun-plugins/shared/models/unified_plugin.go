package models

import (
	"time"

	"github.com/google/uuid"
)

// UnifiedPluginMetadata 统一插件元数据
type UnifiedPluginMetadata struct {
	// 基础信息
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Version     string    `json:"version" db:"version"`
	Description string    `json:"description" db:"description"`
	Author      string    `json:"author" db:"author"`
	AuthorID    uuid.UUID `json:"author_id" db:"author_id"`
	
	// 分类和标签
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`
	Category    string    `json:"category" db:"category"`
	Tags        []string  `json:"tags" db:"tags"`
	
	// 插件类型和技术信息
	Type         string   `json:"type" db:"type"`         // http, event, scheduled, data, custom
	Runtime      string   `json:"runtime" db:"runtime"`   // go, nodejs, python, docker
	Architecture string   `json:"architecture" db:"architecture"` // amd64, arm64, universal
	
	// 权限和依赖
	Permissions  []string `json:"permissions" db:"permissions"`
	Dependencies []string `json:"dependencies" db:"dependencies"`
	
	// 配置和接口
	Config       map[string]interface{} `json:"config" db:"config"`
	APIEndpoints []APIEndpoint          `json:"api_endpoints" db:"api_endpoints"`
	EventTypes   []string              `json:"event_types" db:"event_types"`
	
	// 市场信息
	Icon        string   `json:"icon" db:"icon"`
	Screenshots []string `json:"screenshots" db:"screenshots"`
	Price       float64  `json:"price" db:"price"`
	IsFeatured  bool     `json:"is_featured" db:"is_featured"`
	IsActive    bool     `json:"is_active" db:"is_active"`
	
	// 统计信息
	Downloads   int     `json:"downloads" db:"downloads"`
	Rating      float64 `json:"rating" db:"rating"`
	ReviewCount int     `json:"review_count" db:"review_count"`
	
	// 审核状态
	ReviewStatus    string     `json:"review_status" db:"review_status"`
	ReviewPriority  string     `json:"review_priority" db:"review_priority"`
	ReviewedAt      *time.Time `json:"reviewed_at" db:"reviewed_at"`
	ReviewerID      *uuid.UUID `json:"reviewer_id" db:"reviewer_id"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// APIEndpoint API端点定义
type APIEndpoint struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
	Response    string            `json:"response"`
}

// UnifiedPluginState 统一插件状态
type UnifiedPluginState struct {
	// 运行时状态
	RuntimeState    string     `json:"runtime_state"`    // unloaded, loaded, running, stopped, error
	MarketState     string     `json:"market_state"`     // draft, submitted, approved, rejected, published
	LoadedAt        *time.Time `json:"loaded_at"`
	StartedAt       *time.Time `json:"started_at"`
	StoppedAt       *time.Time `json:"stopped_at"`
	LastHealthCheck *time.Time `json:"last_health_check"`
	
	// 错误信息
	ErrorMsg     string `json:"error_msg,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorDetails string `json:"error_details,omitempty"`
	
	// 资源使用
	ResourceUsage UnifiedResourceUsage `json:"resource_usage"`
}

// UnifiedResourceUsage 统一资源使用情况
type UnifiedResourceUsage struct {
	CPUPercent     float64   `json:"cpu_percent"`
	MemoryBytes    uint64    `json:"memory_bytes"`
	DiskBytes      uint64    `json:"disk_bytes"`
	NetworkIn      uint64    `json:"network_in"`
	NetworkOut     uint64    `json:"network_out"`
	GoroutineCount int       `json:"goroutine_count"`
	RequestCount   int64     `json:"request_count"`
	ErrorCount     int64     `json:"error_count"`
	ResponseTime   float64   `json:"response_time_ms"`
	Uptime         time.Duration `json:"uptime"`
	LastUpdated    time.Time `json:"last_updated"`
}

// UnifiedPluginVersion 统一插件版本
type UnifiedPluginVersion struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PluginID    uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Version     string    `json:"version" db:"version"`
	Changelog   string    `json:"changelog" db:"changelog"`
	
	// 文件信息
	DownloadURL string `json:"download_url" db:"download_url"`
	FileSize    int64  `json:"file_size" db:"file_size"`
	FileHash    string `json:"file_hash" db:"file_hash"`
	
	// 兼容性
	MinSystemVersion string `json:"min_system_version" db:"min_system_version"`
	MaxSystemVersion string `json:"max_system_version" db:"max_system_version"`
	
	// 状态
	IsStable    bool `json:"is_stable" db:"is_stable"`
	IsActive    bool `json:"is_active" db:"is_active"`
	IsBeta      bool `json:"is_beta" db:"is_beta"`
	
	// 统计
	Downloads int `json:"downloads" db:"downloads"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UnifiedPluginConfig 统一插件配置
type UnifiedPluginConfig struct {
	ID       uuid.UUID              `json:"id" db:"id"`
	PluginID uuid.UUID              `json:"plugin_id" db:"plugin_id"`
	UserID   uuid.UUID              `json:"user_id" db:"user_id"`
	Config   map[string]interface{} `json:"config" db:"config"`
	IsActive bool                   `json:"is_active" db:"is_active"`
	
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UnifiedPluginEvent 统一插件事件
type UnifiedPluginEvent struct {
	ID        uuid.UUID   `json:"id"`
	Type      string      `json:"type"`
	Source    string      `json:"source"`
	Target    string      `json:"target,omitempty"`
	Data      interface{} `json:"data"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Priority  int         `json:"priority"`
	TTL       time.Duration `json:"ttl,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// UnifiedPluginMetrics 统一插件指标
type UnifiedPluginMetrics struct {
	PluginID    uuid.UUID `json:"plugin_id"`
	MetricType  string    `json:"metric_type"`  // performance, usage, error, business
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Tags        map[string]string `json:"tags"`
	Timestamp   time.Time `json:"timestamp"`
}

// UnifiedPluginLog 统一插件日志
type UnifiedPluginLog struct {
	ID        uuid.UUID `json:"id"`
	PluginID  uuid.UUID `json:"plugin_id"`
	Level     string    `json:"level"`    // debug, info, warn, error, fatal
	Message   string    `json:"message"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time `json:"timestamp"`
}

// PluginSyncEvent 插件同步事件
type PluginSyncEvent struct {
	EventType    string                 `json:"event_type"`    // create, update, delete, state_change
	PluginID     uuid.UUID              `json:"plugin_id"`
	Changes      map[string]interface{} `json:"changes"`
	Source       string                 `json:"source"`        // market, runtime, admin
	Timestamp    time.Time              `json:"timestamp"`
	Version      int64                  `json:"version"`       // 用于乐观锁
}

// DataSyncStatus 数据同步状态
type DataSyncStatus struct {
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	LastSync    time.Time `json:"last_sync"`
	Status      string    `json:"status"`      // success, failed, in_progress
	ErrorMsg    string    `json:"error_msg,omitempty"`
	RecordCount int       `json:"record_count"`
}