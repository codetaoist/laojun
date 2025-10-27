package discovery

import (
	"fmt"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"go.uber.org/zap"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
	Health  string            `json:"health"`
}

// Service 服务发现接口
type Service interface {
	// 注册服务
	Register(instance *ServiceInstance) error
	// 注销服务
	Deregister(serviceID string) error
	// 发现服务
	Discover(serviceName string) ([]*ServiceInstance, error)
	// 获取健康的服务实例
	GetHealthyInstances(serviceName string) ([]*ServiceInstance, error)
	// 关闭连接
	Close() error
}

// NewService 创建服务发现服务
func NewService(cfg config.DiscoveryConfig, logger *zap.Logger) (Service, error) {
	switch cfg.Type {
	case "laojun":
		return NewLaojunService(cfg.Laojun, logger)
	case "consul":
		return NewConsulService(cfg.Consul, logger)
	case "static":
		return NewStaticService(cfg.Static, logger)
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", cfg.Type)
	}
}