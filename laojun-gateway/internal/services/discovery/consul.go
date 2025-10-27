package discovery

import (
	"fmt"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// ConsulService Consul服务发现实现
type ConsulService struct {
	client *api.Client
	logger *zap.Logger
}

// NewConsulService 创建Consul服务发现
func NewConsulService(cfg config.ConsulConfig, logger *zap.Logger) (*ConsulService, error) {
	config := api.DefaultConfig()
	config.Address = cfg.Address
	config.Scheme = cfg.Scheme

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 测试连接
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to consul: %w", err)
	}

	logger.Info("Connected to Consul", zap.String("address", cfg.Address))

	return &ConsulService{
		client: client,
		logger: logger,
	}, nil
}

// Register 注册服务
func (cs *ConsulService) Register(instance *ServiceInstance) error {
	registration := &api.AgentServiceRegistration{
		ID:      instance.ID,
		Name:    instance.Name,
		Address: instance.Address,
		Port:    instance.Port,
		Tags:    instance.Tags,
		Meta:    instance.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", instance.Address, instance.Port),
			Interval:                       "30s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "90s",
		},
	}

	err := cs.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	cs.logger.Info("Service registered", 
		zap.String("id", instance.ID),
		zap.String("name", instance.Name),
		zap.String("address", fmt.Sprintf("%s:%d", instance.Address, instance.Port)))

	return nil
}

// Deregister 注销服务
func (cs *ConsulService) Deregister(serviceID string) error {
	err := cs.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	cs.logger.Info("Service deregistered", zap.String("id", serviceID))
	return nil
}

// Discover 发现服务
func (cs *ConsulService) Discover(serviceName string) ([]*ServiceInstance, error) {
	services, _, err := cs.client.Health().Service(serviceName, "", false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}

	var instances []*ServiceInstance
	for _, service := range services {
		instance := &ServiceInstance{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
			Health:  cs.getHealthStatus(service.Checks),
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// GetHealthyInstances 获取健康的服务实例
func (cs *ConsulService) GetHealthyInstances(serviceName string) ([]*ServiceInstance, error) {
	services, _, err := cs.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy instances: %w", err)
	}

	var instances []*ServiceInstance
	for _, service := range services {
		instance := &ServiceInstance{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
			Health:  "passing",
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// getHealthStatus 获取健康状态
func (cs *ConsulService) getHealthStatus(checks api.HealthChecks) string {
	for _, check := range checks {
		if check.Status == "critical" {
			return "critical"
		}
		if check.Status == "warning" {
			return "warning"
		}
	}
	return "passing"
}

// Close 关闭连接
func (cs *ConsulService) Close() error {
	cs.logger.Info("Consul service discovery closed")
	return nil
}