package registry

import (
	"context"
	"time"
)

// ServiceRegistry 统一的服务注册接口
type ServiceRegistry interface {
	// RegisterService 注册服务
	RegisterService(ctx context.Context, service *ServiceInfo) error
	
	// DeregisterService 注销服务
	DeregisterService(ctx context.Context, serviceID string) error
	
	// UpdateService 更新服务信息
	UpdateService(ctx context.Context, service *ServiceInfo) error
	
	// GetService 获取服务信息
	GetService(ctx context.Context, serviceID string) (*ServiceInfo, error)
	
	// ListServices 列出服务
	ListServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	
	// ListAllServices 列出所有服务
	ListAllServices(ctx context.Context) (map[string][]*ServiceInfo, error)
	
	// GetHealthyServices 获取健康的服务实例
	GetHealthyServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	
	// WatchServices 监听服务变化
	WatchServices(ctx context.Context, serviceName string) (<-chan *ServiceEvent, error)
	
	// Heartbeat 发送心跳
	Heartbeat(ctx context.Context, serviceID string) error
	
	// GetRegistryHealth 获取注册中心健康状态
	GetRegistryHealth(ctx context.Context) (*RegistryHealth, error)
}

// ServiceDiscovery 服务发现接口
type ServiceDiscovery interface {
	// DiscoverServices 发现服务
	DiscoverServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	
	// DiscoverServicesByTag 根据标签发现服务
	DiscoverServicesByTag(ctx context.Context, tag string) ([]*ServiceInfo, error)
	
	// DiscoverServicesByMeta 根据元数据发现服务
	DiscoverServicesByMeta(ctx context.Context, meta map[string]string) ([]*ServiceInfo, error)
	
	// GetServiceEndpoint 获取服务端点
	GetServiceEndpoint(ctx context.Context, serviceName string, strategy LoadBalanceStrategy) (*ServiceInfo, error)
	
	// Subscribe 订阅服务变化事件
	Subscribe(ctx context.Context, serviceName string) (<-chan *ServiceEvent, error)
	
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, subscription <-chan *ServiceEvent) error
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Tags        []string          `json:"tags"`
	Meta        map[string]string `json:"meta"`
	Health      *HealthCheck      `json:"health"`
	TTL         int               `json:"ttl"`
	Weight      int               `json:"weight"`
	Status      ServiceStatus     `json:"status"`
	RegisteredAt time.Time        `json:"registered_at"`
	LastSeen    time.Time         `json:"last_seen"`
}

// HealthCheck 健康检查配置
type HealthCheck struct {
	Type     HealthCheckType `json:"type"`
	HTTP     *HTTPCheck      `json:"http,omitempty"`
	TCP      *TCPCheck       `json:"tcp,omitempty"`
	Script   *ScriptCheck    `json:"script,omitempty"`
	Interval time.Duration   `json:"interval"`
	Timeout  time.Duration   `json:"timeout"`
	Retries  int             `json:"retries"`
}

// HTTPCheck HTTP健康检查
type HTTPCheck struct {
	URL                string            `json:"url"`
	Method             string            `json:"method"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	ExpectedBody       string            `json:"expected_body"`
	TLSSkipVerify      bool              `json:"tls_skip_verify"`
}

// TCPCheck TCP健康检查
type TCPCheck struct {
	Address string        `json:"address"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

// ScriptCheck 脚本健康检查
type ScriptCheck struct {
	Command []string          `json:"command"`
	Env     map[string]string `json:"env"`
	Timeout time.Duration     `json:"timeout"`
}

// ServiceEvent 服务事件
type ServiceEvent struct {
	Type      EventType    `json:"type"`
	Service   *ServiceInfo `json:"service"`
	Timestamp time.Time    `json:"timestamp"`
}

// RegistryHealth 注册中心健康状态
type RegistryHealth struct {
	Status           HealthStatus      `json:"status"`
	TotalServices    int               `json:"total_services"`
	HealthyServices  int               `json:"healthy_services"`
	UnhealthyServices int              `json:"unhealthy_services"`
	LastCheck        time.Time         `json:"last_check"`
	Details          map[string]string `json:"details"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Type              RegistryType      `json:"type"`
	Address           string            `json:"address"`
	Timeout           time.Duration     `json:"timeout"`
	RetryInterval     time.Duration     `json:"retry_interval"`
	MaxRetries        int               `json:"max_retries"`
	HeartbeatInterval time.Duration     `json:"heartbeat_interval"`
	TTL               time.Duration     `json:"ttl"`
	Auth              *AuthConfig       `json:"auth,omitempty"`
	TLS               *TLSConfig        `json:"tls,omitempty"`
	Namespace         string            `json:"namespace"`
	Tags              []string          `json:"tags"`
	Meta              map[string]string `json:"meta"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled    bool   `json:"enabled"`
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
	CAFile     string `json:"ca_file"`
	SkipVerify bool   `json:"skip_verify"`
}

// 枚举类型定义

// ServiceStatus 服务状态
type ServiceStatus string

const (
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusInactive ServiceStatus = "inactive"
	ServiceStatusDraining ServiceStatus = "draining"
	ServiceStatusMaintenance ServiceStatus = "maintenance"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheckType 健康检查类型
type HealthCheckType string

const (
	HealthCheckTypeHTTP   HealthCheckType = "http"
	HealthCheckTypeTCP    HealthCheckType = "tcp"
	HealthCheckTypeScript HealthCheckType = "script"
)

// EventType 事件类型
type EventType string

const (
	EventTypeRegister   EventType = "register"
	EventTypeDeregister EventType = "deregister"
	EventTypeUpdate     EventType = "update"
	EventTypeHealthy    EventType = "healthy"
	EventTypeUnhealthy  EventType = "unhealthy"
)

// RegistryType 注册中心类型
type RegistryType string

const (
	RegistryTypeConsul    RegistryType = "consul"
	RegistryTypeEtcd      RegistryType = "etcd"
	RegistryTypeZookeeper RegistryType = "zookeeper"
	RegistryTypeRedis     RegistryType = "redis"
	RegistryTypeMemory    RegistryType = "memory"
)

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	LoadBalanceRoundRobin LoadBalanceStrategy = "round_robin"
	LoadBalanceRandom     LoadBalanceStrategy = "random"
	LoadBalanceWeighted   LoadBalanceStrategy = "weighted"
	LoadBalanceLeastConn  LoadBalanceStrategy = "least_conn"
	LoadBalanceConsistentHash LoadBalanceStrategy = "consistent_hash"
)