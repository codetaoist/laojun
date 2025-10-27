package models

import (
	"time"

	sharedmodels "github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
)

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeInProcess    PluginType = "in_process"   // 进程内插件
	PluginTypeMicroservice PluginType = "microservice" // 微服务插件
)

// PluginRuntime 插件运行时枚举
type PluginRuntime string

const (
	RuntimeGo     PluginRuntime = "go"     // Go Plugin (.so)
	RuntimeJS     PluginRuntime = "js"     // JavaScript 模块
	RuntimeDocker PluginRuntime = "docker" // Docker 容器
	RuntimeGRPC   PluginRuntime = "grpc"   // gRPC 服务
)

// PluginStatus 插件状态枚举
type PluginStatus string

const (
	StatusPending   PluginStatus = "pending"   // 待部署状态
	StatusDeploying PluginStatus = "deploying" // 部署中状态
	StatusRunning   PluginStatus = "running"   // 运行中状态
	StatusStopped   PluginStatus = "stopped"   // 已停止状态
	StatusError     PluginStatus = "error"     // 错误状态
)

// ExtendedPlugin 扩展插件模型
type ExtendedPlugin struct {
	// 基础字段
	ID               uuid.UUID  `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Description      *string    `json:"description" db:"description"`
	ShortDescription *string    `json:"short_description" db:"short_description"`
	Author           string     `json:"author" db:"author"`
	DeveloperID      *uuid.UUID `json:"developer_id" db:"developer_id"`
	Version          string     `json:"version" db:"version"`
	CategoryID       *uuid.UUID `json:"category_id" db:"category_id"`
	Price            float64    `json:"price" db:"price"`
	IsFree           bool       `json:"is_free" db:"is_free"`
	IsFeatured       bool       `json:"is_featured" db:"is_featured"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	DownloadCount    int        `json:"download_count" db:"download_count"`
	Rating           float64    `json:"rating" db:"rating"`
	ReviewCount      int        `json:"review_count" db:"review_count"`
	IconURL          *string    `json:"icon_url" db:"icon_url"`
	BannerURL        *string    `json:"banner_url" db:"banner_url"`
	Screenshots      []string   `json:"screenshots" db:"screenshots"`
	Tags             []string   `json:"tags" db:"tags"`
	Requirements     *string    `json:"requirements" db:"requirements"`
	Changelog        *string    `json:"changelog" db:"changelog"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// 新增插件类型相关字段
	Type           PluginType    `json:"type" db:"type"`
	Runtime        PluginRuntime `json:"runtime" db:"runtime"`
	Status         PluginStatus  `json:"status" db:"status"`
	InterfaceSpec  *string       `json:"interface_spec" db:"interface_spec"`   // 接口规范 (JSON)
	Dependencies   []string      `json:"dependencies" db:"dependencies"`       // 依赖列表
	RuntimeConfig  *string       `json:"runtime_config" db:"runtime_config"`   // 运行时配置 (JSON)
	SecurityPolicy *string       `json:"security_policy" db:"security_policy"` // 安全策略
	ResourceLimits *string       `json:"resource_limits" db:"resource_limits"` // 资源限制

	// 进程内插件特有字段
	EntryPoint      *string  `json:"entry_point" db:"entry_point"`           // 入口函数
	ExportedSymbols []string `json:"exported_symbols" db:"exported_symbols"` // 导出符号
	BinaryPath      *string  `json:"binary_path" db:"binary_path"`           // 二进制文件路径
	// 微服务插件特有字段
	DockerImage     *string `json:"docker_image" db:"docker_image"`           // Docker 镜像
	ServicePort     *int    `json:"service_port" db:"service_port"`           // 服务端口
	HealthCheckPath *string `json:"health_check_path" db:"health_check_path"` // 健康检查路径
	GRPCProtoFile   *string `json:"grpc_proto_file" db:"grpc_proto_file"`     // gRPC Proto 文件
	ServiceEndpoint *string `json:"service_endpoint" db:"service_endpoint"`   // 服务端点
	Namespace       *string `json:"namespace" db:"namespace"`                 // K8s 命名空间

	// 关联数据
	Category *sharedmodels.Category `json:"category,omitempty"`
}

// PluginInterface 插件接口定义
type PluginInterface struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PluginID     uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description" db:"description"`
	Methods      []string  `json:"methods" db:"methods"`             // 接口方法列表
	InputSchema  *string   `json:"input_schema" db:"input_schema"`   // 输入参数 JSON Schema
	OutputSchema *string   `json:"output_schema" db:"output_schema"` // 输出参数 JSON Schema
	Version      string    `json:"version" db:"version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PluginDependency 插件依赖关系
type PluginDependency struct {
	ID             uuid.UUID `json:"id" db:"id"`
	PluginID       uuid.UUID `json:"plugin_id" db:"plugin_id"`
	DependencyID   uuid.UUID `json:"dependency_id" db:"dependency_id"`
	DependencyType string    `json:"dependency_type" db:"dependency_type"` // plugin, service, library
	MinVersion     *string   `json:"min_version" db:"min_version"`
	MaxVersion     *string   `json:"max_version" db:"max_version"`
	IsOptional     bool      `json:"is_optional" db:"is_optional"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// PluginInstance 插件实例
type PluginInstance struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	PluginID   uuid.UUID    `json:"plugin_id" db:"plugin_id"`
	UserID     *uuid.UUID   `json:"user_id" db:"user_id"`
	InstanceID string       `json:"instance_id" db:"instance_id"` // 运行时实例ID
	Status     PluginStatus `json:"status" db:"status"`
	Config     *string      `json:"config" db:"config"`       // 实例配置
	Endpoint   *string      `json:"endpoint" db:"endpoint"`   // 服务端点
	LastPing   *time.Time   `json:"last_ping" db:"last_ping"` // 最后心跳时间
	ErrorMsg   *string      `json:"error_msg" db:"error_msg"` // 错误信息
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
}

// PluginCallLog 插件调用日志
type PluginCallLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	PluginID   uuid.UUID  `json:"plugin_id" db:"plugin_id"`
	InstanceID *string    `json:"instance_id" db:"instance_id"`
	UserID     *uuid.UUID `json:"user_id" db:"user_id"`
	Method     string     `json:"method" db:"method"`
	InputData  *string    `json:"input_data" db:"input_data"`
	OutputData *string    `json:"output_data" db:"output_data"`
	Duration   int64      `json:"duration" db:"duration"` // 执行时间(毫秒)
	Success    bool       `json:"success" db:"success"`
	ErrorMsg   *string    `json:"error_msg" db:"error_msg"`
	ClientType string     `json:"client_type" db:"client_type"` // web, mobile, iot, desktop
	ClientIP   *string    `json:"client_ip" db:"client_ip"`
	UserAgent  *string    `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PluginID      uuid.UUID `json:"plugin_id" db:"plugin_id"`
	Date          time.Time `json:"date" db:"date"`
	CallCount     int64     `json:"call_count" db:"call_count"`
	SuccessCount  int64     `json:"success_count" db:"success_count"`
	ErrorCount    int64     `json:"error_count" db:"error_count"`
	AvgDuration   float64   `json:"avg_duration" db:"avg_duration"`
	MaxDuration   int64     `json:"max_duration" db:"max_duration"`
	MinDuration   int64     `json:"min_duration" db:"min_duration"`
	TotalDuration int64     `json:"total_duration" db:"total_duration"`
	UniqueUsers   int64     `json:"unique_users" db:"unique_users"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// PluginCallRequest 插件调用请求
type PluginCallRequest struct {
	PluginID   string                 `json:"plugin_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	Method     string                 `json:"method"`
	Params     map[string]interface{} `json:"params"`
	Context    map[string]interface{} `json:"context"`
	InputData  map[string]interface{} `json:"input_data,omitempty"`
	ClientType string                 `json:"client_type,omitempty"`
	ClientIP   string                 `json:"client_ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
}

// PluginCallResponse 插件调用响应
type PluginCallResponse struct {
	Success  bool                   `json:"success"`
	Data     interface{}            `json:"data,omitempty"`
	Error    *string                `json:"error,omitempty"`
	Duration int64                  `json:"duration"` // 执行时间(毫秒)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PluginSearchParams 扩展插件搜索参数
type PluginSearchParams struct {
	Query      string         `form:"q"`
	CategoryID *uuid.UUID     `form:"category_id"`
	Type       *PluginType    `form:"type"`
	Runtime    *PluginRuntime `form:"runtime"`
	Status     *PluginStatus  `form:"status"`
	Featured   *bool          `form:"featured"`
	Free       *bool          `form:"free"`
	MinPrice   *float64       `form:"min_price"`
	MaxPrice   *float64       `form:"max_price"`
	MinRating  *float64       `form:"min_rating"`
	SortBy     string         `form:"sort_by"`    // name, rating, downloads, created_at
	SortOrder  string         `form:"sort_order"` // asc, desc
	Page       int            `form:"page"`
	PageSize   int            `form:"page_size"`
}
