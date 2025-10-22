package plugin

import (
	"context"
	"time"
)

// PluginInfo 插件基础信息
type PluginInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Metadata    map[string]string `json:"metadata"`
}

// PluginContext 插件执行上下文
type PluginContext struct {
	UserID     string                 `json:"user_id"`
	RequestID  string                 `json:"request_id"`
	ClientType string                 `json:"client_type"` // web, mobile, iot, desktop
	Headers    map[string]string      `json:"headers"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timeout    time.Duration          `json:"timeout"`
}

// PluginResult 插件执行结果
type PluginResult struct {
	Success  bool                   `json:"success"`
	Data     interface{}            `json:"data,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Duration time.Duration          `json:"duration"`
}

// BasePlugin 基础插件接口
type BasePlugin interface {
	// GetInfo 获取插件信息
	GetInfo() PluginInfo

	// Initialize 初始化插件
	Initialize(config map[string]interface{}) error

	// Cleanup 清理资源
	Cleanup() error

	// HealthCheck 健康检查
	HealthCheck() error
}

// DataProcessor 数据处理插件接口
type DataProcessor interface {
	BasePlugin

	// ProcessData 处理数据
	ProcessData(ctx context.Context, input interface{}) (*PluginResult, error)

	// GetInputSchema 获取输入数据结构
	GetInputSchema() map[string]interface{}

	// GetOutputSchema 获取输出数据结构
	GetOutputSchema() map[string]interface{}
}

// ImageFilter 图像过滤器插件接口
type ImageFilter interface {
	BasePlugin

	// FilterImage 过滤图像
	FilterImage(ctx context.Context, imageData []byte, params map[string]interface{}) (*PluginResult, error)

	// GetSupportedFormats 获取支持的图像格式
	GetSupportedFormats() []string

	// GetFilterParams 获取过滤器参数定义
	GetFilterParams() map[string]interface{}
}

// TextAnalyzer 文本分析插件接口
type TextAnalyzer interface {
	BasePlugin

	// AnalyzeText 分析文本
	AnalyzeText(ctx context.Context, text string, options map[string]interface{}) (*PluginResult, error)

	// GetSupportedLanguages 获取支持的语言
	GetSupportedLanguages() []string

	// GetAnalysisTypes 获取分析类型
	GetAnalysisTypes() []string
}

// APIConnector API连接器插件接口
type APIConnector interface {
	BasePlugin

	// CallAPI 调用外部API
	CallAPI(ctx context.Context, endpoint string, method string, headers map[string]string, body interface{}) (*PluginResult, error)

	// GetEndpoints 获取支持的端点
	GetEndpoints() []string

	// ValidateCredentials 验证凭据
	ValidateCredentials(credentials map[string]string) error
}

// DatabaseConnector 数据库连接器插件接口
type DatabaseConnector interface {
	BasePlugin

	// Connect 连接数据库
	Connect(connectionString string) error

	// Query 查询数据
	Query(ctx context.Context, sql string, params []interface{}) (*PluginResult, error)

	// Execute 执行命令
	Execute(ctx context.Context, sql string, params []interface{}) (*PluginResult, error)

	// Disconnect 断开连接
	Disconnect() error
}

// NotificationSender 通知发送器插件接口
type NotificationSender interface {
	BasePlugin

	// SendNotification 发送通知
	SendNotification(ctx context.Context, recipient string, message string, options map[string]interface{}) (*PluginResult, error)

	// GetSupportedChannels 获取支持的通知渠道
	GetSupportedChannels() []string

	// ValidateRecipient 验证接收人
	ValidateRecipient(recipient string, channel string) error
}

// WorkflowStep 工作流步骤插件接口
type WorkflowStep interface {
	BasePlugin

	// Execute 执行工作流步骤
	Execute(ctx context.Context, input interface{}, stepConfig map[string]interface{}) (*PluginResult, error)

	// GetStepSchema 获取步骤配置结构
	GetStepSchema() map[string]interface{}

	// ValidateConfig 验证配置
	ValidateConfig(config map[string]interface{}) error
}

// UIComponent UI组件插件接口（用于前端）
type UIComponent interface {
	BasePlugin

	// Render 渲染组件
	Render(ctx context.Context, props map[string]interface{}) (*PluginResult, error)

	// GetComponentSchema 获取组件属性结构
	GetComponentSchema() map[string]interface{}

	// GetEvents 获取支持的事件
	GetEvents() []string
}

// DeviceController 设备控制器插件接口（用于IoT设备）
type DeviceController interface {
	BasePlugin

	// ControlDevice 控制设备
	ControlDevice(ctx context.Context, deviceID string, command string, params map[string]interface{}) (*PluginResult, error)

	// GetDeviceStatus 获取设备状态
	GetDeviceStatus(ctx context.Context, deviceID string) (*PluginResult, error)

	// GetSupportedCommands 获取支持的命令
	GetSupportedCommands() []string

	// RegisterDevice 注册设备
	RegisterDevice(deviceID string, deviceInfo map[string]interface{}) error
}

// PluginRegistry 插件注册表接口
type PluginRegistry interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(pluginID string, plugin BasePlugin) error

	// UnregisterPlugin 注销插件
	UnregisterPlugin(pluginID string) error

	// GetPlugin 获取插件
	GetPlugin(pluginID string) (BasePlugin, error)

	// ListPlugins 列出所有插件
	ListPlugins() []string

	// GetPluginsByInterface 根据接口类型获取插件
	GetPluginsByInterface(interfaceType string) []string
}

// PluginLoader 插件加载器接口
type PluginLoader interface {
	// LoadPlugin 加载插件
	LoadPlugin(pluginPath string, config map[string]interface{}) (BasePlugin, error)

	// UnloadPlugin 卸载插件
	UnloadPlugin(pluginID string) error

	// ReloadPlugin 重新加载插件
	ReloadPlugin(pluginID string) error

	// ValidatePlugin 验证插件
	ValidatePlugin(pluginPath string) error
}

// PluginManager 插件管理器接口
type PluginManager interface {
	// StartPlugin 启动插件
	StartPlugin(pluginID string) error

	// StopPlugin 停止插件
	StopPlugin(pluginID string) error

	// RestartPlugin 重启插件
	RestartPlugin(pluginID string) error

	// GetPluginStatus 获取插件状态
	GetPluginStatus(pluginID string) (string, error)

	// CallPlugin 调用插件
	CallPlugin(pluginID string, method string, params map[string]interface{}) (*PluginResult, error)

	// GetPluginMetrics 获取插件指标
	GetPluginMetrics(pluginID string) (map[string]interface{}, error)
}

// PluginSecurityManager 插件安全管理器接口
type PluginSecurityManager interface {
	// ValidatePermissions 验证权限
	ValidatePermissions(pluginID string, operation string, resource string) error

	// GrantPermission 授予权限
	GrantPermission(pluginID string, permission string) error

	// RevokePermission 撤销权限
	RevokePermission(pluginID string, permission string) error

	// GetPermissions 获取权限列表
	GetPermissions(pluginID string) ([]string, error)

	// CreateSandbox 创建沙箱环境
	CreateSandbox(pluginID string, config map[string]interface{}) error

	// DestroySandbox 销毁沙箱环境
	DestroySandbox(pluginID string) error
}
