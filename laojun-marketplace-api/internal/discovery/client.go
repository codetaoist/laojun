package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/laojun-shared/registry"
	"go.uber.org/zap"
)

// MarketplaceDiscoveryClient 插件市场服务发现客户端
type MarketplaceDiscoveryClient struct {
	registry    registry.ServiceRegistry
	discovery   registry.ServiceDiscovery
	logger      *zap.Logger
	serviceInfo *registry.ServiceInfo
}

// 向后兼容的类型定义（已弃用）
// Deprecated: 使用 registry.ServiceInfo 替代
type ServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Meta     map[string]string `json:"meta"`
	Health   HealthCheck       `json:"health"`
	TTL      int               `json:"ttl"`
}

// Deprecated: 使用 registry.HealthCheck 替代
type HealthCheck struct {
	HTTP     string `json:"http"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

// Deprecated: 使用 NewMarketplaceDiscoveryClient 替代
type Client struct {
	discoveryURL string
	logger       *zap.Logger
	serviceInfo  *ServiceInfo
}

// NewMarketplaceDiscoveryClient 创建插件市场服务发现客户端
func NewMarketplaceDiscoveryClient(registry registry.ServiceRegistry, discovery registry.ServiceDiscovery, logger *zap.Logger) *MarketplaceDiscoveryClient {
	return &MarketplaceDiscoveryClient{
		registry:  registry,
		discovery: discovery,
		logger:    logger,
	}
}

// RegisterService 注册服务
func (c *MarketplaceDiscoveryClient) RegisterService(ctx context.Context, service *registry.ServiceInfo) error {
	c.serviceInfo = service
	
	if err := c.registry.RegisterService(ctx, service); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	c.logger.Info("Marketplace service registered successfully", 
		zap.String("service_id", service.ID),
		zap.String("service_name", service.Name))

	return nil
}

// DeregisterService 注销服务
func (c *MarketplaceDiscoveryClient) DeregisterService(ctx context.Context) error {
	if c.serviceInfo == nil {
		return nil
	}

	if err := c.registry.DeregisterService(ctx, c.serviceInfo.ID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	c.logger.Info("Marketplace service deregistered successfully", 
		zap.String("service_id", c.serviceInfo.ID))

	return nil
}

// UpdateService 更新服务信息
func (c *MarketplaceDiscoveryClient) UpdateService(ctx context.Context, service *registry.ServiceInfo) error {
	if err := c.registry.UpdateService(ctx, service); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	c.serviceInfo = service
	c.logger.Info("Marketplace service updated successfully", 
		zap.String("service_id", service.ID))

	return nil
}

// DiscoverServices 发现服务
func (c *MarketplaceDiscoveryClient) DiscoverServices(ctx context.Context, serviceName string) ([]*registry.ServiceInfo, error) {
	services, err := c.discovery.DiscoverServices(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	c.logger.Debug("Services discovered",
		zap.String("service_name", serviceName),
		zap.Int("count", len(services)))

	return services, nil
}

// GetServiceEndpoint 获取服务端点
func (c *MarketplaceDiscoveryClient) GetServiceEndpoint(ctx context.Context, serviceName string, strategy registry.LoadBalanceStrategy) (*registry.ServiceInfo, error) {
	service, err := c.discovery.GetServiceEndpoint(ctx, serviceName, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to get service endpoint: %w", err)
	}

	return service, nil
}

// StartHeartbeat 启动心跳
func (c *MarketplaceDiscoveryClient) StartHeartbeat(ctx context.Context) {
	if c.serviceInfo == nil {
		c.logger.Warn("No service info available for heartbeat")
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Heartbeat stopped due to context cancellation")
			return
		case <-ticker.C:
			if err := c.sendHeartbeat(ctx); err != nil {
				c.logger.Error("Failed to send heartbeat", zap.Error(err))
			}
		}
	}
}

// WatchServices 监听服务变化
func (c *MarketplaceDiscoveryClient) WatchServices(ctx context.Context, serviceName string) (<-chan *registry.ServiceEvent, error) {
	return c.discovery.Subscribe(ctx, serviceName)
}

// sendHeartbeat 发送心跳
func (c *MarketplaceDiscoveryClient) sendHeartbeat(ctx context.Context) error {
	if err := c.registry.Heartbeat(ctx, c.serviceInfo.ID); err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	c.logger.Debug("Heartbeat sent successfully",
		zap.String("service_id", c.serviceInfo.ID))

	return nil
}

// 向后兼容的方法（已弃用）

// Deprecated: 使用 NewMarketplaceDiscoveryClient 替代
func NewClient(discoveryURL string, logger *zap.Logger) *Client {
	return &Client{
		discoveryURL: discoveryURL,
		logger:       logger,
	}
}

// Deprecated: 使用 MarketplaceDiscoveryClient.RegisterService 替代
func (c *Client) Register(ctx context.Context, service *ServiceInfo) error {
	// 简化的向后兼容实现
	c.logger.Warn("Using deprecated Register method, please migrate to MarketplaceDiscoveryClient")
	return fmt.Errorf("deprecated method, please use MarketplaceDiscoveryClient")
}

// Deprecated: 使用 MarketplaceDiscoveryClient.DeregisterService 替代
func (c *Client) Deregister(ctx context.Context) error {
	c.logger.Warn("Using deprecated Deregister method, please migrate to MarketplaceDiscoveryClient")
	return fmt.Errorf("deprecated method, please use MarketplaceDiscoveryClient")
}

// Deprecated: 使用 MarketplaceDiscoveryClient.StartHeartbeat 替代
func (c *Client) StartHeartbeat(ctx context.Context) {
	c.logger.Warn("Using deprecated StartHeartbeat method, please migrate to MarketplaceDiscoveryClient")
}

// Deprecated: 内部方法，不再使用
func (c *Client) sendHeartbeat(ctx context.Context) error {
	return fmt.Errorf("deprecated method")
}