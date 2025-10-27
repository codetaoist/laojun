package runtime

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PluginState 插件状态枚举
type PluginState int

const (
	StateUnloaded     PluginState = iota // 未加载
	StateLoaded                          // 已加载
	StateInitializing                    // 初始化中
	StateInitialized                     // 已初始化
	StateStarting                        // 启动中
	StateRunning                         // 运行中
	StateStopping                        // 停止中
	StateStopped                         // 已停止
	StateError                           // 错误状态
)

func (s PluginState) String() string {
	switch s {
	case StateUnloaded:
		return "unloaded"
	case StateLoaded:
		return "loaded"
	case StateInitializing:
		return "initializing"
	case StateInitialized:
		return "initialized"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Category    string            `json:"category"`
	Type        string            `json:"type"`        // 插件类型：http, event, scheduled, data, custom
	Tags        []string          `json:"tags"`
	Permissions []string          `json:"permissions"`
	Dependencies []string         `json:"dependencies"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PluginInfo 插件运行时信息
type PluginInfo struct {
	Metadata    PluginMetadata `json:"metadata"`
	State       PluginState    `json:"state"`
	LoadedAt    *time.Time     `json:"loaded_at,omitempty"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	StoppedAt   *time.Time     `json:"stopped_at,omitempty"`
	ErrorMsg    string         `json:"error_msg,omitempty"`
	ResourceUsage ResourceUsage `json:"resource_usage"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryBytes   uint64  `json:"memory_bytes"`
	GoroutineCount int    `json:"goroutine_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// Plugin 插件接口
type Plugin interface {
	// GetMetadata 获取插件元数据
	GetMetadata() PluginMetadata

	// Initialize 初始化插件
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Start 启动插件
	Start(ctx context.Context) error

	// Stop 停止插件
	Stop(ctx context.Context) error

	// Cleanup 清理插件资源
	Cleanup(ctx context.Context) error

	// GetStatus 获取插件状态
	GetStatus() PluginState

	// HandleEvent 处理事件
	HandleEvent(ctx context.Context, event interface{}) error

	// ProcessData 处理数据
	ProcessData(ctx context.Context, data interface{}) (interface{}, error)
}

// PluginEvent 插件事件实现
type PluginEvent struct {
	ID        uuid.UUID   `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func (e *PluginEvent) GetID() uuid.UUID {
	return e.ID
}

func (e *PluginEvent) GetType() string {
	return e.Type
}

func (e *PluginEvent) GetData() interface{} {
	return e.Data
}

func (e *PluginEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// NewPluginEvent 创建新的插件事件
func NewPluginEvent(eventType string, data interface{}) *PluginEvent {
	return &PluginEvent{
		ID:        uuid.New(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
}