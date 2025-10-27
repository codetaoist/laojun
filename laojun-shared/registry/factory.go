package registry

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RegistryFactory 服务注册工厂
type RegistryFactory struct {
	logger *zap.Logger
}

// NewRegistryFactory 创建注册工厂
func NewRegistryFactory(logger *zap.Logger) *RegistryFactory {
	return &RegistryFactory{
		logger: logger,
	}
}

// CreateRegistry 创建服务注册实例
func (f *RegistryFactory) CreateRegistry(config *RegistryConfig) (ServiceRegistry, error) {
	switch config.Type {
	case RegistryTypeMemory:
		return f.createMemoryRegistry(config)
	case RegistryTypeRedis:
		return f.createRedisRegistry(config)
	case RegistryTypeConsul:
		return f.createConsulRegistry(config)
	case RegistryTypeEtcd:
		return f.createEtcdRegistry(config)
	case RegistryTypeZookeeper:
		return f.createZookeeperRegistry(config)
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", config.Type)
	}
}

// CreateDiscovery 创建服务发现实例
func (f *RegistryFactory) CreateDiscovery(registry ServiceRegistry) ServiceDiscovery {
	return NewDefaultServiceDiscovery(registry, f.logger)
}

// CreateDefaultConfig 创建默认配置
func (f *RegistryFactory) CreateDefaultConfig(registryType RegistryType) *RegistryConfig {
	config := &RegistryConfig{
		Type:              registryType,
		Timeout:           30 * time.Second,
		RetryInterval:     5 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 10 * time.Second,
		TTL:               30 * time.Second,
		Namespace:         "laojun",
		Tags:              []string{},
		Meta:              make(map[string]string),
	}

	switch registryType {
	case RegistryTypeMemory:
		config.Address = "memory://localhost"
	case RegistryTypeRedis:
		config.Address = "redis://localhost:6379"
	case RegistryTypeConsul:
		config.Address = "consul://localhost:8500"
	case RegistryTypeEtcd:
		config.Address = "etcd://localhost:2379"
	case RegistryTypeZookeeper:
		config.Address = "zookeeper://localhost:2181"
	}

	return config
}

// 私有方法

// createMemoryRegistry 创建内存注册中心
func (f *RegistryFactory) createMemoryRegistry(config *RegistryConfig) (ServiceRegistry, error) {
	registry := NewDefaultServiceRegistry(config, f.logger)
	return registry, nil
}

// createRedisRegistry 创建Redis注册中心
func (f *RegistryFactory) createRedisRegistry(config *RegistryConfig) (ServiceRegistry, error) {
	// TODO: 实现Redis注册中心
	f.logger.Warn("Redis registry not implemented, using memory registry")
	return f.createMemoryRegistry(config)
}

// createConsulRegistry 创建Consul注册中心
func (f *RegistryFactory) createConsulRegistry(config *RegistryConfig) (ServiceRegistry, error) {
	// TODO: 实现Consul注册中心
	f.logger.Warn("Consul registry not implemented, using memory registry")
	return f.createMemoryRegistry(config)
}

// createEtcdRegistry 创建Etcd注册中心
func (f *RegistryFactory) createEtcdRegistry(config *RegistryConfig) (ServiceRegistry, error) {
	// TODO: 实现Etcd注册中心
	f.logger.Warn("Etcd registry not implemented, using memory registry")
	return f.createMemoryRegistry(config)
}

// createZookeeperRegistry 创建Zookeeper注册中心
func (f *RegistryFactory) createZookeeperRegistry(config *RegistryConfig) (ServiceRegistry, error) {
	// TODO: 实现Zookeeper注册中心
	f.logger.Warn("Zookeeper registry not implemented, using memory registry")
	return f.createMemoryRegistry(config)
}

// RegistryBuilder 注册中心构建器
type RegistryBuilder struct {
	config *RegistryConfig
	logger *zap.Logger
}

// NewRegistryBuilder 创建注册中心构建器
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		config: &RegistryConfig{
			Type:              RegistryTypeMemory,
			Timeout:           30 * time.Second,
			RetryInterval:     5 * time.Second,
			MaxRetries:        3,
			HeartbeatInterval: 10 * time.Second,
			TTL:               30 * time.Second,
			Namespace:         "laojun",
			Tags:              []string{},
			Meta:              make(map[string]string),
		},
	}
}

// WithType 设置注册中心类型
func (b *RegistryBuilder) WithType(registryType RegistryType) *RegistryBuilder {
	b.config.Type = registryType
	return b
}

// WithAddress 设置地址
func (b *RegistryBuilder) WithAddress(address string) *RegistryBuilder {
	b.config.Address = address
	return b
}

// WithTimeout 设置超时时间
func (b *RegistryBuilder) WithTimeout(timeout time.Duration) *RegistryBuilder {
	b.config.Timeout = timeout
	return b
}

// WithRetryInterval 设置重试间隔
func (b *RegistryBuilder) WithRetryInterval(interval time.Duration) *RegistryBuilder {
	b.config.RetryInterval = interval
	return b
}

// WithMaxRetries 设置最大重试次数
func (b *RegistryBuilder) WithMaxRetries(maxRetries int) *RegistryBuilder {
	b.config.MaxRetries = maxRetries
	return b
}

// WithHeartbeatInterval 设置心跳间隔
func (b *RegistryBuilder) WithHeartbeatInterval(interval time.Duration) *RegistryBuilder {
	b.config.HeartbeatInterval = interval
	return b
}

// WithTTL 设置TTL
func (b *RegistryBuilder) WithTTL(ttl time.Duration) *RegistryBuilder {
	b.config.TTL = ttl
	return b
}

// WithAuth 设置认证
func (b *RegistryBuilder) WithAuth(auth *AuthConfig) *RegistryBuilder {
	b.config.Auth = auth
	return b
}

// WithTLS 设置TLS
func (b *RegistryBuilder) WithTLS(tls *TLSConfig) *RegistryBuilder {
	b.config.TLS = tls
	return b
}

// WithNamespace 设置命名空间
func (b *RegistryBuilder) WithNamespace(namespace string) *RegistryBuilder {
	b.config.Namespace = namespace
	return b
}

// WithTags 设置标签
func (b *RegistryBuilder) WithTags(tags []string) *RegistryBuilder {
	b.config.Tags = tags
	return b
}

// WithMeta 设置元数据
func (b *RegistryBuilder) WithMeta(meta map[string]string) *RegistryBuilder {
	b.config.Meta = meta
	return b
}

// WithLogger 设置日志器
func (b *RegistryBuilder) WithLogger(logger *zap.Logger) *RegistryBuilder {
	b.logger = logger
	return b
}

// Build 构建注册中心
func (b *RegistryBuilder) Build() (ServiceRegistry, error) {
	if b.logger == nil {
		b.logger = zap.NewNop()
	}

	factory := NewRegistryFactory(b.logger)
	return factory.CreateRegistry(b.config)
}

// BuildWithDiscovery 构建注册中心和服务发现
func (b *RegistryBuilder) BuildWithDiscovery() (ServiceRegistry, ServiceDiscovery, error) {
	registry, err := b.Build()
	if err != nil {
		return nil, nil, err
	}

	factory := NewRegistryFactory(b.logger)
	discovery := factory.CreateDiscovery(registry)

	return registry, discovery, nil
}