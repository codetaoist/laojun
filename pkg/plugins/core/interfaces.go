// Package core 定义插件系统的核心接口和类型
package core

import (
	"context"
	"time"
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
	StatusUnloaded PluginStatus = "unloaded" // 未加载
	StatusLoaded   PluginStatus = "loaded"   // 已加载
	StatusStarted  PluginStatus = "started"  // 已启动
	StatusStopped  PluginStatus = "stopped"  // 已停止
	StatusError    PluginStatus = "error"    // 错误状态
)

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Author       string         `json:"author"`
	Description  string         `json:"description"`
	Type         PluginType     `json:"type"`
	Runtime      PluginRuntime  `json:"runtime"`
	Dependencies []string       `json:"dependencies"`
	Permissions  []string       `json:"permissions"`
	Config       map[string]any `json:"config"`
	EntryPoint   string         `json:"entry_point"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// PluginRequest 插件请求
type PluginRequest struct {
	ID       string          `json:"id"`
	Method   string          `json:"method"`
	Params   map[string]any  `json:"params"`
	Context  context.Context `json:"-"`
	Metadata map[string]any  `json:"metadata"`
}

// PluginResponse 插件响应
type PluginResponse struct {
	ID       string         `json:"id"`
	Success  bool           `json:"success"`
	Data     any            `json:"data"`
	Error    string         `json:"error,omitempty"`
	Metadata map[string]any `json:"metadata"`
}

// PluginEvent 插件事件
type PluginEvent struct {
	Type      string         `json:"type"`
	Source    string         `json:"source"`
	Target    string         `json:"target,omitempty"`
	Data      any            `json:"data"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:"metadata"`
}

// Plugin 统一插件接口
type Plugin interface {
	// GetMetadata 获取插件元数据
	GetMetadata() *PluginMetadata

	// Initialize 初始化插件
	Initialize(ctx context.Context, config map[string]any) error

	// Start 启动插件
	Start(ctx context.Context) error

	// Stop 停止插件
	Stop(ctx context.Context) error

	// HandleRequest 处理请求
	HandleRequest(ctx context.Context, req *PluginRequest) (*PluginResponse, error)

	// HandleEvent 处理事件
	HandleEvent(ctx context.Context, event *PluginEvent) error

	// GetStatus 获取插件状态
	GetStatus() PluginStatus

	// GetHealth 获取健康状态
	GetHealth(ctx context.Context) (map[string]any, error)
}

// PluginLoader 插件加载器接口
type PluginLoader interface {
	// LoadPlugin 加载插件
	LoadPlugin(ctx context.Context, path string, metadata *PluginMetadata) (Plugin, error)

	// UnloadPlugin 卸载插件
	UnloadPlugin(ctx context.Context, pluginID string) error

	// GetSupportedRuntimes 获取支持的运行时类型
	GetSupportedRuntimes() []PluginRuntime

	// ValidatePlugin 验证插件
	ValidatePlugin(path string, metadata *PluginMetadata) error
}

// PluginManager 插件管理器接口
type PluginManager interface {
	// RegisterLoader 注册插件加载器
	RegisterLoader(runtime PluginRuntime, loader PluginLoader) error

	// LoadPlugin 加载插件
	LoadPlugin(ctx context.Context, path string) error

	// UnloadPlugin 卸载插件
	UnloadPlugin(ctx context.Context, pluginID string) error

	// StartPlugin 启动插件
	StartPlugin(ctx context.Context, pluginID string) error

	// StopPlugin 停止插件
	StopPlugin(ctx context.Context, pluginID string) error

	// GetPlugin 获取插件实例
	GetPlugin(pluginID string) (Plugin, error)

	// ListPlugins 列出所有插件元数据
	ListPlugins() []*PluginMetadata

	// CallPlugin 调用插件方法
	CallPlugin(ctx context.Context, pluginID string, method string, params map[string]any) (*PluginResponse, error)

	// BroadcastEvent 广播事件
	BroadcastEvent(ctx context.Context, event *PluginEvent) error
}

// SecurityManager 安全管理器接口
type SecurityManager interface {
	// ValidatePermissions 验证权限
	ValidatePermissions(pluginID string, permissions []string) error

	// ScanCode 扫描代码安全问题
	ScanCode(path string) (*SecurityReport, error)

	// CreateSandbox 创建沙箱环境
	CreateSandbox(pluginID string, config *SandboxConfig) (Sandbox, error)

	// CheckResourceLimits 检查资源限制
	CheckResourceLimits(pluginID string) (*ResourceUsage, error)
}

// SecurityReport 安全报告
type SecurityReport struct {
	PluginID  string          `json:"plugin_id"`
	Passed    bool            `json:"passed"`
	Issues    []SecurityIssue `json:"issues"`
	Score     int             `json:"score"`
	Timestamp time.Time       `json:"timestamp"`
}

// SecurityIssue 安全问题
type SecurityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	MaxMemory     int64         `json:"max_memory"`
	MaxCPUTime    time.Duration `json:"max_cpu_time"`
	MaxGoroutines int           `json:"max_goroutines"`
	AllowedAPIs   []string      `json:"allowed_apis"`
	NetworkAccess bool          `json:"network_access"`
}

// Sandbox 沙箱接口
type Sandbox interface {
	// Execute 在沙箱中执行代码
	Execute(ctx context.Context, fn func() (any, error)) (any, error)

	// GetResourceUsage 获取资源使用情况
	GetResourceUsage() *ResourceUsage

	// Destroy 销毁沙箱环境
	Destroy() error
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryUsed     int64         `json:"memory_used"`
	CPUTime        time.Duration `json:"cpu_time"`
	GoroutineCount int           `json:"goroutine_count"`
	NetworkIO      int64         `json:"network_io"`
}

// EventBus 事件总线接口
type EventBus interface {
	// Subscribe 订阅事件
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe 取消订阅
	Unsubscribe(eventType string, handler EventHandler) error

	// Publish 发布事件
	Publish(ctx context.Context, event *PluginEvent) error

	// PublishAsync 异步发布事件
	PublishAsync(ctx context.Context, event *PluginEvent) error
}

// EventHandler 事件处理接口
type EventHandler interface {
	// Handle 处理事件
	Handle(ctx context.Context, event *PluginEvent) error

	// GetID 获取处理器ID
	GetID() string
}

// PluginRegistry 插件注册表接口
type PluginRegistry interface {
	// Register 注册插件
	Register(metadata *PluginMetadata) error

	// Unregister 注销插件
	Unregister(pluginID string) error

	// Get 获取插件元数据
	Get(pluginID string) (*PluginMetadata, error)

	// List 列出所有插件元数据
	List() ([]*PluginMetadata, error)

	// Search 搜索插件
	Search(query string, filters map[string]any) ([]*PluginMetadata, error)

	// Update 更新插件元数据
	Update(pluginID string, metadata *PluginMetadata) error
}

// RouteHandler 路由处理接口
type RouteHandler interface {
	// Handle 处理HTTP请求
	Handle(ctx context.Context, req *PluginRequest) (*PluginResponse, error)

	// GetPath 获取路由路径
	GetPath() string

	// GetMethod 获取HTTP方法
	GetMethod() string
}

// PluginConfig 插件配置
type PluginConfig struct {
	PluginPaths      []string        `json:"plugin_paths"`
	EnableHotReload  bool            `json:"enable_hot_reload"`
	EnableSecurity   bool            `json:"enable_security"`
	EnableMonitoring bool            `json:"enable_monitoring"`
	MaxPlugins       int             `json:"max_plugins"`
	DefaultSandbox   *SandboxConfig  `json:"default_sandbox"`
	Security         *SecurityConfig `json:"security"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableSandbox      bool     `json:"enable_sandbox"`
	EnableCodeScanning bool     `json:"enable_code_scanning"`
	AllowedPermissions []string `json:"allowed_permissions"`
	AuditLog           bool     `json:"audit_log"`
}
