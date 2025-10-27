package config

import (
	"fmt"
	"os"
	"strconv"
)

// ServicePorts 定义所有服务的端口配置
type ServicePorts struct {
	Gateway        int `yaml:"gateway" json:"gateway"`
	AdminAPI       int `yaml:"admin_api" json:"admin_api"`
	Monitoring     int `yaml:"monitoring" json:"monitoring"`
	Discovery      int `yaml:"discovery" json:"discovery"`
	PluginManager  int `yaml:"plugin_manager" json:"plugin_manager"`
	MarketplaceAPI int `yaml:"marketplace_api" json:"marketplace_api"`
	ConfigCenter   int `yaml:"config_center" json:"config_center"`
	AdminWeb       int `yaml:"admin_web" json:"admin_web"`
	MarketplaceWeb int `yaml:"marketplace_web" json:"marketplace_web"`
}

// InfraPorts 定义基础设施端口配置
type InfraPorts struct {
	PostgreSQL int `yaml:"postgresql" json:"postgresql"`
	Redis      int `yaml:"redis" json:"redis"`
	Prometheus int `yaml:"prometheus" json:"prometheus"`
	Grafana    int `yaml:"grafana" json:"grafana"`
}

// PortConfig 统一端口配置
type PortConfig struct {
	Services ServicePorts `yaml:"services" json:"services"`
	Infra    InfraPorts   `yaml:"infra" json:"infra"`
}

// DefaultPortConfig 返回默认端口配置
func DefaultPortConfig() *PortConfig {
	return &PortConfig{
		Services: ServicePorts{
			Gateway:        8081,
			AdminAPI:       8082,
			Monitoring:     8083,
			Discovery:      8084,
			PluginManager:  8085,
			MarketplaceAPI: 8086,
			ConfigCenter:   8087,
			AdminWeb:       3000,
			MarketplaceWeb: 3001,
		},
		Infra: InfraPorts{
			PostgreSQL: 5432,
			Redis:      6379,
			Prometheus: 9090,
			Grafana:    3000,
		},
	}
}

// GetServicePort 获取指定服务的端口，支持环境变量覆盖
func (pc *PortConfig) GetServicePort(service string) int {
	var defaultPort int
	var envKey string

	switch service {
	case "gateway":
		defaultPort = pc.Services.Gateway
		envKey = "GATEWAY_PORT"
	case "admin-api":
		defaultPort = pc.Services.AdminAPI
		envKey = "ADMIN_API_PORT"
	case "monitoring":
		defaultPort = pc.Services.Monitoring
		envKey = "MONITORING_PORT"
	case "discovery":
		defaultPort = pc.Services.Discovery
		envKey = "DISCOVERY_PORT"
	case "plugin-manager":
		defaultPort = pc.Services.PluginManager
		envKey = "PLUGIN_MANAGER_PORT"
	case "marketplace-api":
		defaultPort = pc.Services.MarketplaceAPI
		envKey = "MARKETPLACE_API_PORT"
	case "config-center":
		defaultPort = pc.Services.ConfigCenter
		envKey = "CONFIG_CENTER_PORT"
	case "admin-web":
		defaultPort = pc.Services.AdminWeb
		envKey = "ADMIN_WEB_PORT"
	case "marketplace-web":
		defaultPort = pc.Services.MarketplaceWeb
		envKey = "MARKETPLACE_WEB_PORT"
	default:
		return 0
	}

	// 检查环境变量覆盖
	if envPort := os.Getenv(envKey); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil {
			return port
		}
	}

	return defaultPort
}

// GetInfraPort 获取基础设施端口
func (pc *PortConfig) GetInfraPort(service string) int {
	switch service {
	case "postgresql":
		if envPort := os.Getenv("POSTGRES_PORT"); envPort != "" {
			if port, err := strconv.Atoi(envPort); err == nil {
				return port
			}
		}
		return pc.Infra.PostgreSQL
	case "redis":
		if envPort := os.Getenv("REDIS_PORT"); envPort != "" {
			if port, err := strconv.Atoi(envPort); err == nil {
				return port
			}
		}
		return pc.Infra.Redis
	case "prometheus":
		if envPort := os.Getenv("PROMETHEUS_PORT"); envPort != "" {
			if port, err := strconv.Atoi(envPort); err == nil {
				return port
			}
		}
		return pc.Infra.Prometheus
	case "grafana":
		if envPort := os.Getenv("GRAFANA_PORT"); envPort != "" {
			if port, err := strconv.Atoi(envPort); err == nil {
				return port
			}
		}
		return pc.Infra.Grafana
	default:
		return 0
	}
}

// GetServerAddress 获取服务器地址
func (pc *PortConfig) GetServerAddress(service string, host ...string) string {
	port := pc.GetServicePort(service)
	if port == 0 {
		return ""
	}

	hostAddr := "0.0.0.0"
	if len(host) > 0 && host[0] != "" {
		hostAddr = host[0]
	}

	return fmt.Sprintf("%s:%d", hostAddr, port)
}

// IsPortAvailable 检查端口是否可用（简单检查，实际使用时可以扩展）
func (pc *PortConfig) IsPortAvailable(port int) bool {
	// 这里可以实现端口可用性检查逻辑
	// 暂时返回 true，实际项目中可以添加网络检查
	return true
}

// ValidatePortConfig 验证端口配置
func (pc *PortConfig) ValidatePortConfig() error {
	ports := make(map[int]string)
	
	// 检查服务端口冲突
	serviceMap := map[string]int{
		"gateway":         pc.Services.Gateway,
		"admin-api":       pc.Services.AdminAPI,
		"monitoring":      pc.Services.Monitoring,
		"discovery":       pc.Services.Discovery,
		"plugin-manager":  pc.Services.PluginManager,
		"marketplace-api": pc.Services.MarketplaceAPI,
		"config-center":   pc.Services.ConfigCenter,
		"admin-web":       pc.Services.AdminWeb,
		"marketplace-web": pc.Services.MarketplaceWeb,
	}

	for service, port := range serviceMap {
		if existingService, exists := ports[port]; exists {
			return fmt.Errorf("端口冲突: %s 和 %s 都使用端口 %d", service, existingService, port)
		}
		ports[port] = service
	}

	return nil
}

// GetAllServicePorts 获取所有服务端口信息
func (pc *PortConfig) GetAllServicePorts() map[string]int {
	return map[string]int{
		"gateway":         pc.GetServicePort("gateway"),
		"admin-api":       pc.GetServicePort("admin-api"),
		"monitoring":      pc.GetServicePort("monitoring"),
		"discovery":       pc.GetServicePort("discovery"),
		"plugin-manager":  pc.GetServicePort("plugin-manager"),
		"marketplace-api": pc.GetServicePort("marketplace-api"),
		"config-center":   pc.GetServicePort("config-center"),
		"admin-web":       pc.GetServicePort("admin-web"),
		"marketplace-web": pc.GetServicePort("marketplace-web"),
	}
}