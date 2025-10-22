package gateway

import (
	"fmt"
	"time"
)

// NewPluginRegistry 创建新的插件注册中心
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]*PluginEndpoint),
		routes:  make(map[string]*RouteConfig),
	}
}

// RegisterPlugin 注册插件
func (pr *PluginRegistry) RegisterPlugin(plugin *PluginEndpoint) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if plugin.ID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	// 验证插件实例
	if len(plugin.Instances) == 0 {
		return fmt.Errorf("plugin must have at least one instance")
	}

	// 设置默认值
	if plugin.Type == "" {
		plugin.Type = "filter"
	}
	if plugin.Status == "" {
		plugin.Status = "active"
	}
	plugin.LastUpdated = time.Now()

	// 初始化实例健康状态
	for _, instance := range plugin.Instances {
		if instance.Health == nil {
			instance.Health = &HealthStatus{
				Status:    "unknown",
				Timestamp: time.Now(),
			}
		}
		if instance.Metrics == nil {
			instance.Metrics = &InstanceMetrics{}
		}
	}

	pr.plugins[plugin.ID] = plugin
	return nil
}

// UnregisterPlugin 注销插件
func (pr *PluginRegistry) UnregisterPlugin(pluginID string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if _, exists := pr.plugins[pluginID]; !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 移除相关路由
	routesToRemove := make([]string, 0)
	for path, route := range pr.routes {
		if route.PluginID == pluginID {
			routesToRemove = append(routesToRemove, path)
		}
	}

	for _, path := range routesToRemove {
		delete(pr.routes, path)
	}

	delete(pr.plugins, pluginID)
	return nil
}

// GetPlugin 获取插件
func (pr *PluginRegistry) GetPlugin(pluginID string) *PluginEndpoint {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return pr.plugins[pluginID]
}

// ListPlugins 列出所有插件
func (pr *PluginRegistry) ListPlugins() []*PluginEndpoint {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]*PluginEndpoint, 0, len(pr.plugins))
	for _, plugin := range pr.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// UpdatePluginStatus 更新插件状态
func (pr *PluginRegistry) UpdatePluginStatus(pluginID, status string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	plugin, exists := pr.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	plugin.Status = status
	plugin.LastUpdated = time.Now()
	return nil
}

// AddRoute 添加路由
func (pr *PluginRegistry) AddRoute(route *RouteConfig) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if route.Path == "" {
		return fmt.Errorf("route path cannot be empty")
	}

	if route.PluginID == "" {
		return fmt.Errorf("route plugin ID cannot be empty")
	}

	// 验证插件是否存在
	if _, exists := pr.plugins[route.PluginID]; !exists {
		return fmt.Errorf("plugin %s not found", route.PluginID)
	}

	// 设置默认值
	if len(route.Methods) == 0 {
		route.Methods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}

	if route.Timeout == 0 {
		route.Timeout = 30 * time.Second
	}

	pr.routes[route.Path] = route
	return nil
}

// RemoveRoute 移除路由
func (pr *PluginRegistry) RemoveRoute(path string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if _, exists := pr.routes[path]; !exists {
		return fmt.Errorf("route %s not found", path)
	}

	delete(pr.routes, path)
	return nil
}

// GetRoute 获取路由
func (pr *PluginRegistry) GetRoute(path string) *RouteConfig {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return pr.routes[path]
}

// ListRoutes 列出所有路由
func (pr *PluginRegistry) ListRoutes() []*RouteConfig {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	routes := make([]*RouteConfig, 0, len(pr.routes))
	for _, route := range pr.routes {
		routes = append(routes, route)
	}

	return routes
}

// UpdateInstanceHealth 更新实例健康状态
func (pr *PluginRegistry) UpdateInstanceHealth(pluginID, instanceID string, health *HealthStatus) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	plugin, exists := pr.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	for _, instance := range plugin.Instances {
		if instance.ID == instanceID {
			instance.Health = health
			instance.Metrics.LastHealthCheck = time.Now()
			return nil
		}
	}

	return fmt.Errorf("instance %s not found in plugin %s", instanceID, pluginID)
}

// UpdateInstanceMetrics 更新实例指标
func (pr *PluginRegistry) UpdateInstanceMetrics(pluginID, instanceID string, metrics *InstanceMetrics) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	plugin, exists := pr.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	for _, instance := range plugin.Instances {
		if instance.ID == instanceID {
			instance.Metrics = metrics
			return nil
		}
	}

	return fmt.Errorf("instance %s not found in plugin %s", instanceID, pluginID)
}

// GetHealthyInstances 获取健康的实例
func (pr *PluginRegistry) GetHealthyInstances(pluginID string) []*PluginInstance {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugin, exists := pr.plugins[pluginID]
	if !exists {
		return nil
	}

	healthyInstances := make([]*PluginInstance, 0)
	for _, instance := range plugin.Instances {
		if instance.Status == "running" &&
			instance.Health != nil &&
			instance.Health.Status == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	return healthyInstances
}

// GetHealthSummary 获取健康状态摘要
func (pr *PluginRegistry) GetHealthSummary() map[string]interface{} {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	summary := map[string]interface{}{
		"total_plugins":     len(pr.plugins),
		"total_routes":      len(pr.routes),
		"healthy_plugins":   0,
		"unhealthy_plugins": 0,
		"total_instances":   0,
		"healthy_instances": 0,
		"plugins":           make(map[string]interface{}),
	}

	for pluginID, plugin := range pr.plugins {
		totalInstances := len(plugin.Instances)
		healthyInstances := 0

		for _, instance := range plugin.Instances {
			if instance.Status == "running" &&
				instance.Health != nil &&
				instance.Health.Status == "healthy" {
				healthyInstances++
			}
		}

		pluginHealthy := healthyInstances > 0
		if pluginHealthy {
			summary["healthy_plugins"] = summary["healthy_plugins"].(int) + 1
		} else {
			summary["unhealthy_plugins"] = summary["unhealthy_plugins"].(int) + 1
		}

		summary["total_instances"] = summary["total_instances"].(int) + totalInstances
		summary["healthy_instances"] = summary["healthy_instances"].(int) + healthyInstances

		summary["plugins"].(map[string]interface{})[pluginID] = map[string]interface{}{
			"status":            plugin.Status,
			"total_instances":   totalInstances,
			"healthy_instances": healthyInstances,
			"last_updated":      plugin.LastUpdated,
		}
	}

	return summary
}

// GetPluginsByType 根据类型获取插件
func (pr *PluginRegistry) GetPluginsByType(pluginType string) []*PluginEndpoint {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]*PluginEndpoint, 0)
	for _, plugin := range pr.plugins {
		if plugin.Type == pluginType {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// GetRoutesByPlugin 根据插件ID获取路由
func (pr *PluginRegistry) GetRoutesByPlugin(pluginID string) []*RouteConfig {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	routes := make([]*RouteConfig, 0)
	for _, route := range pr.routes {
		if route.PluginID == pluginID {
			routes = append(routes, route)
		}
	}

	return routes
}

// ValidatePlugin 验证插件配置
func (pr *PluginRegistry) ValidatePlugin(plugin *PluginEndpoint) error {
	if plugin.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}

	if plugin.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if plugin.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	if plugin.Type == "" {
		return fmt.Errorf("plugin type is required")
	}

	if len(plugin.Instances) == 0 {
		return fmt.Errorf("plugin must have at least one instance")
	}

	// 验证实例
	for i, instance := range plugin.Instances {
		if instance.ID == "" {
			return fmt.Errorf("instance %d: ID is required", i)
		}

		if instance.Address == "" {
			return fmt.Errorf("instance %d: address is required", i)
		}

		if instance.Port <= 0 || instance.Port > 65535 {
			return fmt.Errorf("instance %d: invalid port %d", i, instance.Port)
		}
	}

	return nil
}

// ValidateRoute 验证路由配置
func (pr *PluginRegistry) ValidateRoute(route *RouteConfig) error {
	if route.Path == "" {
		return fmt.Errorf("route path is required")
	}

	if route.PluginID == "" {
		return fmt.Errorf("route plugin ID is required")
	}

	// 验证HTTP方法
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "OPTIONS": true, "HEAD": true,
	}

	for _, method := range route.Methods {
		if !validMethods[method] {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
	}

	// 验证重试配置
	if route.Retry != nil {
		if route.Retry.MaxAttempts < 0 {
			return fmt.Errorf("retry max attempts cannot be negative")
		}

		if route.Retry.Backoff < 0 {
			return fmt.Errorf("retry backoff cannot be negative")
		}

		if route.Retry.Timeout < 0 {
			return fmt.Errorf("retry timeout cannot be negative")
		}
	}

	// 验证缓存配置
	if route.Cache != nil && route.Cache.Enabled {
		if route.Cache.TTL <= 0 {
			return fmt.Errorf("cache TTL must be positive")
		}
	}

	// 验证限流配置
	if route.RateLimit != nil && route.RateLimit.Enabled {
		if route.RateLimit.Rate <= 0 {
			return fmt.Errorf("rate limit rate must be positive")
		}

		if route.RateLimit.Burst <= 0 {
			return fmt.Errorf("rate limit burst must be positive")
		}

		if route.RateLimit.Window <= 0 {
			return fmt.Errorf("rate limit window must be positive")
		}
	}

	return nil
}

// GetStatistics 获取统计信息
func (pr *PluginRegistry) GetStatistics() map[string]interface{} {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	stats := map[string]interface{}{
		"plugins": map[string]interface{}{
			"total":     len(pr.plugins),
			"by_type":   make(map[string]int),
			"by_status": make(map[string]int),
		},
		"routes": map[string]interface{}{
			"total":     len(pr.routes),
			"by_plugin": make(map[string]int),
			"by_method": make(map[string]int),
		},
		"instances": map[string]interface{}{
			"total":     0,
			"by_status": make(map[string]int),
			"by_health": make(map[string]int),
		},
	}

	// 统计插件
	for _, plugin := range pr.plugins {
		// 按类型统计		typeStats := stats["plugins"].(map[string]interface{})["by_type"].(map[string]int)
		typeStats[plugin.Type]++

		// 按状态统计		statusStats := stats["plugins"].(map[string]interface{})["by_status"].(map[string]int)
		statusStats[plugin.Status]++

		// 统计实例
		instanceStats := stats["instances"].(map[string]interface{})
		instanceStats["total"] = instanceStats["total"].(int) + len(plugin.Instances)

		for _, instance := range plugin.Instances {
			// 按状态统计			instanceStatusStats := instanceStats["by_status"].(map[string]int)
			instanceStatusStats[instance.Status]++

			// 按健康状态统计			instanceHealthStats := instanceStats["by_health"].(map[string]int)
			if instance.Health != nil {
				instanceHealthStats[instance.Health.Status]++
			} else {
				instanceHealthStats["unknown"]++
			}
		}
	}

	// 统计路由
	for _, route := range pr.routes {
		// 按插件统计		pluginStats := stats["routes"].(map[string]interface{})["by_plugin"].(map[string]int)
		pluginStats[route.PluginID]++

		// 按方法统计		methodStats := stats["routes"].(map[string]interface{})["by_method"].(map[string]int)
		for _, method := range route.Methods {
			methodStats[method]++
		}
	}

	return stats
}
