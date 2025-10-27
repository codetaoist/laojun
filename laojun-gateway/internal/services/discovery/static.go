package discovery

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// StaticService 静态服务发现实现
type StaticService struct {
	services map[string][]*ServiceInstance
	logger   *zap.Logger
}

// NewStaticService 创建静态服务发现
func NewStaticService(staticConfig map[string]string, logger *zap.Logger) (*StaticService, error) {
	services := make(map[string][]*ServiceInstance)

	// 解析静态配置
	for serviceName, addresses := range staticConfig {
		var instances []*ServiceInstance
		
		addressList := strings.Split(addresses, ",")
		for i, addr := range addressList {
			addr = strings.TrimSpace(addr)
			parts := strings.Split(addr, ":")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid address format: %s", addr)
			}

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port in address: %s", addr)
			}

			instance := &ServiceInstance{
				ID:      fmt.Sprintf("%s-%d", serviceName, i),
				Name:    serviceName,
				Address: parts[0],
				Port:    port,
				Tags:    []string{"static"},
				Meta:    map[string]string{"type": "static"},
				Health:  "passing",
			}
			instances = append(instances, instance)
		}
		
		services[serviceName] = instances
	}

	logger.Info("Static service discovery initialized", 
		zap.Int("services", len(services)))

	return &StaticService{
		services: services,
		logger:   logger,
	}, nil
}

// Register 注册服务（静态发现不支持动态注册）
func (ss *StaticService) Register(instance *ServiceInstance) error {
	return fmt.Errorf("static discovery does not support dynamic registration")
}

// Deregister 注销服务（静态发现不支持动态注销）
func (ss *StaticService) Deregister(serviceID string) error {
	return fmt.Errorf("static discovery does not support dynamic deregistration")
}

// Discover 发现服务
func (ss *StaticService) Discover(serviceName string) ([]*ServiceInstance, error) {
	instances, exists := ss.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// 返回副本以避免外部修改
	result := make([]*ServiceInstance, len(instances))
	copy(result, instances)
	
	return result, nil
}

// GetHealthyInstances 获取健康的服务实例
func (ss *StaticService) GetHealthyInstances(serviceName string) ([]*ServiceInstance, error) {
	instances, err := ss.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	// 静态服务假设都是健康的
	var healthyInstances []*ServiceInstance
	for _, instance := range instances {
		if instance.Health == "passing" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances, nil
}

// Close 关闭连接
func (ss *StaticService) Close() error {
	ss.logger.Info("Static service discovery closed")
	return nil
}