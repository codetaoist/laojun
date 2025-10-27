package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"go.uber.org/zap"
)

// LaojunService Laojun服务发现实现
type LaojunService struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// LaojunServiceResponse Laojun服务发现响应
type LaojunServiceResponse struct {
	Services map[string]LaojunServiceInfo `json:"services"`
	Stats    LaojunStats                  `json:"stats"`
}

// LaojunServiceInfo Laojun服务信息
type LaojunServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Health   string            `json:"health"`
	TTL      int               `json:"ttl"`
}

// LaojunStats Laojun统计信息
type LaojunStats struct {
	TotalServices   int `json:"total_services"`
	HealthyServices int `json:"healthy_services"`
}

// NewLaojunService 创建Laojun服务发现
func NewLaojunService(cfg config.LaojunConfig, logger *zap.Logger) (*LaojunService, error) {
	baseURL := fmt.Sprintf("%s://%s", cfg.Scheme, cfg.Address)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &LaojunService{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}, nil
}

// Register 注册服务（网关通常不需要注册自己）
func (ls *LaojunService) Register(instance *ServiceInstance) error {
	// 网关通常不需要注册服务，但可以实现以备将来使用
	ls.logger.Info("Laojun service registration not implemented for gateway")
	return nil
}

// Deregister 注销服务
func (ls *LaojunService) Deregister(serviceID string) error {
	// 网关通常不需要注销服务
	ls.logger.Info("Laojun service deregistration not implemented for gateway")
	return nil
}

// Discover 发现服务
func (ls *LaojunService) Discover(serviceName string) ([]*ServiceInstance, error) {
	url := fmt.Sprintf("%s/api/v1/discovery/services", ls.baseURL)
	
	resp, err := ls.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response LaojunServiceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var instances []*ServiceInstance
	for _, service := range response.Services {
		if service.Name == serviceName {
			instances = append(instances, &ServiceInstance{
				ID:      service.ID,
				Name:    service.Name,
				Address: service.Address,
				Port:    service.Port,
				Tags:    service.Tags,
				Meta:    service.Metadata,
				Health:  service.Health,
			})
		}
	}

	ls.logger.Debug("Discovered services", 
		zap.String("service_name", serviceName),
		zap.Int("instance_count", len(instances)))

	return instances, nil
}

// GetHealthyInstances 获取健康的服务实例
func (ls *LaojunService) GetHealthyInstances(serviceName string) ([]*ServiceInstance, error) {
	instances, err := ls.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Health == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	ls.logger.Debug("Found healthy instances",
		zap.String("service_name", serviceName),
		zap.Int("healthy_count", len(healthyInstances)),
		zap.Int("total_count", len(instances)))

	return healthyInstances, nil
}

// Close 关闭连接
func (ls *LaojunService) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}