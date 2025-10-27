package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	sharedRegistry "github.com/codetaoist/laojun-shared/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PluginServiceRegistry 插件服务注册管理器
type PluginServiceRegistry struct {
	pluginRegistry  PluginRegistry
	serviceRegistry sharedRegistry.ServiceRegistry
	logger          *logrus.Logger
	mu              sync.RWMutex
}

// NewPluginServiceRegistry 创建插件服务注册管理器
func NewPluginServiceRegistry(logger *logrus.Logger) *PluginServiceRegistry {
	pluginRegistry := NewDefaultPluginRegistry(logger)
	serviceRegistry := sharedRegistry.NewDefaultServiceRegistry(&sharedRegistry.RegistryConfig{
		Type: "memory",
		TTL:  30 * time.Second,
	})
	
	return &PluginServiceRegistry{
		pluginRegistry:  pluginRegistry,
		serviceRegistry: serviceRegistry,
		logger:          logger,
	}
}

// RegisterPluginAsService 将插件注册为服务
func (r *PluginServiceRegistry) RegisterPluginAsService(ctx context.Context, registration *PluginRegistration) error {
	// 先注册插件
	if err := r.pluginRegistry.RegisterPlugin(ctx, registration); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}
	
	// 将插件转换为服务信息并注册
	serviceInfo := r.convertPluginToService(registration)
	if err := r.serviceRegistry.Register(ctx, serviceInfo); err != nil {
		// 如果服务注册失败，回滚插件注册
		r.pluginRegistry.UnregisterPlugin(ctx, registration.ID)
		return fmt.Errorf("failed to register plugin as service: %w", err)
	}
	
	return nil
}

// UnregisterPluginService 注销插件服务
func (r *PluginServiceRegistry) UnregisterPluginService(ctx context.Context, pluginID string) error {
	// 先注销服务
	if err := r.serviceRegistry.Deregister(ctx, pluginID); err != nil {
		r.logger.WithError(err).Warnf("Failed to deregister service for plugin %s", pluginID)
	}
	
	// 注销插件
	return r.pluginRegistry.UnregisterPlugin(ctx, pluginID)
}

// GetPluginService 获取插件服务信息
func (r *PluginServiceRegistry) GetPluginService(ctx context.Context, pluginID string) (*sharedRegistry.ServiceInfo, error) {
	return r.serviceRegistry.GetService(ctx, pluginID)
}

// ListPluginServices 列出插件服务
func (r *PluginServiceRegistry) ListPluginServices(ctx context.Context, serviceName string) ([]*sharedRegistry.ServiceInfo, error) {
	return r.serviceRegistry.ListServices(ctx, serviceName)
}

// UpdatePluginHealth 更新插件健康状态
func (r *PluginServiceRegistry) UpdatePluginHealth(ctx context.Context, pluginID string, health *sharedRegistry.HealthCheck) error {
	// 更新服务健康状态
	if err := r.serviceRegistry.UpdateHealth(ctx, pluginID, health); err != nil {
		return fmt.Errorf("failed to update service health: %w", err)
	}
	
	// 更新插件状态
	var status PluginStatus
	switch health.Status {
	case "healthy":
		status = StatusActive
	case "unhealthy":
		status = StatusInactive
	case "degraded":
		status = StatusMaintenance
	default:
		status = StatusInactive
	}
	
	return r.pluginRegistry.UpdatePluginStatus(ctx, pluginID, status)
}

// WatchPluginServices 监听插件服务变化
func (r *PluginServiceRegistry) WatchPluginServices(ctx context.Context, serviceName string) (<-chan *sharedRegistry.ServiceEvent, error) {
	return r.serviceRegistry.Watch(ctx, serviceName)
}

// GetPluginRegistry 获取插件注册中心
func (r *PluginServiceRegistry) GetPluginRegistry() PluginRegistry {
	return r.pluginRegistry
}

// GetServiceRegistry 获取服务注册中心
func (r *PluginServiceRegistry) GetServiceRegistry() sharedRegistry.ServiceRegistry {
	return r.serviceRegistry
}

// convertPluginToService 将插件信息转换为服务信息
func (r *PluginServiceRegistry) convertPluginToService(plugin *PluginRegistration) *sharedRegistry.ServiceInfo {
	// 从插件端点中提取地址和端口信息
	var address string
	var port int
	
	if len(plugin.Endpoints) > 0 {
		// 假设第一个端点包含主要的服务信息
		endpoint := plugin.Endpoints[0]
		// 这里需要根据实际的端点格式来解析地址和端口
		// 简化处理，假设配置中包含地址和端口信息
		if addr, ok := plugin.Config["address"].(string); ok {
			address = addr
		}
		if p, ok := plugin.Config["port"].(float64); ok {
			port = int(p)
		}
	}
	
	// 如果没有配置地址和端口，使用默认值
	if address == "" {
		address = "localhost"
	}
	if port == 0 {
		port = 8080
	}
	
	// 转换健康检查信息
	var healthCheck *sharedRegistry.HealthCheck
	if plugin.Health != nil {
		healthCheck = &sharedRegistry.HealthCheck{
			Status:    plugin.Health.Status,
			Message:   fmt.Sprintf("Plugin %s health check", plugin.Name),
			Timestamp: plugin.Health.LastCheck,
		}
	}
	
	return &sharedRegistry.ServiceInfo{
		ID:          plugin.ID,
		Name:        plugin.Name,
		Version:     plugin.Version,
		Address:     address,
		Port:        port,
		Tags:        plugin.Tags,
		Metadata: map[string]string{
			"type":        "plugin",
			"category":    plugin.Category,
			"author":      plugin.Author,
			"description": plugin.Description,
		},
		Health:       healthCheck,
		RegisteredAt: plugin.RegisteredAt,
		UpdatedAt:    plugin.UpdatedAt,
	}
}

// PluginRegistry 插件注册中心接口 (保持不变，向后兼容)
type PluginRegistry interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(ctx context.Context, registration *PluginRegistration) error

	// UnregisterPlugin 注销插件
	UnregisterPlugin(ctx context.Context, pluginID string) error

	// GetPlugin 获取插件信息
	GetPlugin(ctx context.Context, pluginID string) (*PluginRegistration, error)

	// ListPlugins 列出插件
	ListPlugins(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error)

	// UpdatePluginStatus 更新插件状态
	UpdatePluginStatus(ctx context.Context, pluginID string, status PluginStatus) error

	// DiscoverPlugins 发现插件
	DiscoverPlugins(ctx context.Context, criteria *DiscoveryCriteria) ([]*PluginRegistration, error)

	// Subscribe 订阅插件事件
	Subscribe(ctx context.Context, eventTypes []string) (<-chan *PluginEvent, error)

	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, subscription <-chan *PluginEvent) error

	// GetHealth 获取注册中心健康状态
	GetHealth(ctx context.Context) (*HealthStatus, error)
}

// PluginStatus 插件状态
type PluginStatus string

const (
	StatusRegistered   PluginStatus = "registered"   // 已注册
	StatusActive       PluginStatus = "active"       // 活跃
	StatusInactive     PluginStatus = "inactive"     // 非活跃
	StatusMaintenance  PluginStatus = "maintenance"  // 维护中
	StatusDeprecated   PluginStatus = "deprecated"   // 已弃用
	StatusUnregistered PluginStatus = "unregistered" // 已注销
)

// PluginRegistration 插件注册信息
type PluginRegistration struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	Permissions  []string               `json:"permissions"`
	Dependencies []string               `json:"dependencies"`
	Endpoints    []PluginEndpoint       `json:"endpoints"`
	Config       map[string]interface{} `json:"config"`
	Status       PluginStatus           `json:"status"`
	Health       *HealthInfo            `json:"health,omitempty"`
	Metrics      *PluginMetrics         `json:"metrics,omitempty"`
	RegisteredAt time.Time              `json:"registered_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastSeen     time.Time              `json:"last_seen"`
}

// PluginEndpoint 插件端点信息
type PluginEndpoint struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
	Response    map[string]string `json:"response"`
}

// HealthInfo 健康信息
type HealthInfo struct {
	Status      string            `json:"status"`
	LastCheck   time.Time         `json:"last_check"`
	CheckCount  int               `json:"check_count"`
	FailCount   int               `json:"fail_count"`
	Details     map[string]string `json:"details"`
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	RequestCount    int64     `json:"request_count"`
	ErrorCount      int64     `json:"error_count"`
	AvgResponseTime float64   `json:"avg_response_time"`
	LastRequest     time.Time `json:"last_request"`
	Uptime          float64   `json:"uptime"`
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	Category     string       `json:"category,omitempty"`
	Status       PluginStatus `json:"status,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	Author       string       `json:"author,omitempty"`
	NamePattern  string       `json:"name_pattern,omitempty"`
	Limit        int          `json:"limit,omitempty"`
	Offset       int          `json:"offset,omitempty"`
}

// DiscoveryCriteria 发现条件
type DiscoveryCriteria struct {
	RequiredCapabilities []string `json:"required_capabilities"`
	PreferredTags        []string `json:"preferred_tags"`
	ExcludePlugins       []string `json:"exclude_plugins"`
	MinVersion           string   `json:"min_version,omitempty"`
	MaxResults           int      `json:"max_results,omitempty"`
}

// PluginEvent 插件事件
type PluginEvent struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	PluginID  string                 `json:"plugin_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status         string            `json:"status"`
	TotalPlugins   int               `json:"total_plugins"`
	ActivePlugins  int               `json:"active_plugins"`
	FailedPlugins  int               `json:"failed_plugins"`
	LastCheck      time.Time         `json:"last_check"`
	Details        map[string]string `json:"details"`
}

// DefaultPluginRegistry 默认插件注册中心实现
type DefaultPluginRegistry struct {
	plugins       map[string]*PluginRegistration
	subscribers   map[string][]chan *PluginEvent
	logger        *logrus.Logger
	mu            sync.RWMutex
	eventChan     chan *PluginEvent
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewDefaultPluginRegistry 创建默认插件注册中心
func NewDefaultPluginRegistry(logger *logrus.Logger) *DefaultPluginRegistry {
	ctx, cancel := context.WithCancel(context.Background())
	
	registry := &DefaultPluginRegistry{
		plugins:     make(map[string]*PluginRegistration),
		subscribers: make(map[string][]chan *PluginEvent),
		logger:      logger,
		eventChan:   make(chan *PluginEvent, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}

	// 启动事件处理器
	registry.startEventProcessor()

	return registry
}

// RegisterPlugin 注册插件
func (r *DefaultPluginRegistry) RegisterPlugin(ctx context.Context, registration *PluginRegistration) error {
	r.logger.WithField("plugin_id", registration.ID).Info("Registering plugin")

	r.mu.Lock()
	defer r.mu.Unlock()

	// 验证注册信息
	if err := r.validateRegistration(registration); err != nil {
		return fmt.Errorf("invalid registration: %w", err)
	}

	// 检查插件是否已存在
	if existing, exists := r.plugins[registration.ID]; exists {
		// 如果版本相同，更新注册信息
		if existing.Version == registration.Version {
			registration.RegisteredAt = existing.RegisteredAt
			registration.UpdatedAt = time.Now()
		} else {
			// 版本不同，作为新注册处理
			registration.RegisteredAt = time.Now()
			registration.UpdatedAt = time.Now()
		}
	} else {
		// 新插件注册
		registration.RegisteredAt = time.Now()
		registration.UpdatedAt = time.Now()
	}

	registration.Status = StatusRegistered
	registration.LastSeen = time.Now()

	// 存储注册信息
	r.plugins[registration.ID] = registration

	// 发送注册事件
	event := &PluginEvent{
		ID:        uuid.New(),
		Type:      "plugin.registered",
		PluginID:  registration.ID,
		Data: map[string]interface{}{
			"name":    registration.Name,
			"version": registration.Version,
		},
		Timestamp: time.Now(),
	}

	select {
	case r.eventChan <- event:
	default:
		r.logger.Warn("Event channel is full, dropping event")
	}

	r.logger.WithField("plugin_id", registration.ID).Info("Plugin registered successfully")
	return nil
}

// UnregisterPlugin 注销插件
func (r *DefaultPluginRegistry) UnregisterPlugin(ctx context.Context, pluginID string) error {
	r.logger.WithField("plugin_id", pluginID).Info("Unregistering plugin")

	r.mu.Lock()
	defer r.mu.Unlock()

	registration, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	// 更新状态
	registration.Status = StatusUnregistered
	registration.UpdatedAt = time.Now()

	// 发送注销事件
	event := &PluginEvent{
		ID:        uuid.New(),
		Type:      "plugin.unregistered",
		PluginID:  pluginID,
		Data: map[string]interface{}{
			"name": registration.Name,
		},
		Timestamp: time.Now(),
	}

	select {
	case r.eventChan <- event:
	default:
		r.logger.Warn("Event channel is full, dropping event")
	}

	// 删除插件
	delete(r.plugins, pluginID)

	r.logger.WithField("plugin_id", pluginID).Info("Plugin unregistered successfully")
	return nil
}

// GetPlugin 获取插件信息
func (r *DefaultPluginRegistry) GetPlugin(ctx context.Context, pluginID string) (*PluginRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registration, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginID)
	}

	// 返回副本
	regCopy := *registration
	return &regCopy, nil
}

// ListPlugins 列出插件
func (r *DefaultPluginRegistry) ListPlugins(ctx context.Context, filter *PluginFilter) ([]*PluginRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*PluginRegistration
	
	for _, registration := range r.plugins {
		if r.matchesFilter(registration, filter) {
			regCopy := *registration
			result = append(result, &regCopy)
		}
	}

	// 应用分页
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(result) {
			result = result[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(result) {
			result = result[:filter.Limit]
		}
	}

	return result, nil
}

// UpdatePluginStatus 更新插件状态
func (r *DefaultPluginRegistry) UpdatePluginStatus(ctx context.Context, pluginID string, status PluginStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	registration, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginID)
	}

	oldStatus := registration.Status
	registration.Status = status
	registration.UpdatedAt = time.Now()
	registration.LastSeen = time.Now()

	// 发送状态变更事件
	if oldStatus != status {
		event := &PluginEvent{
			ID:       uuid.New(),
			Type:     "plugin.status_changed",
			PluginID: pluginID,
			Data: map[string]interface{}{
				"old_status": oldStatus,
				"new_status": status,
			},
			Timestamp: time.Now(),
		}

		select {
		case r.eventChan <- event:
		default:
			r.logger.Warn("Event channel is full, dropping event")
		}
	}

	return nil
}

// DiscoverPlugins 发现插件
func (r *DefaultPluginRegistry) DiscoverPlugins(ctx context.Context, criteria *DiscoveryCriteria) ([]*PluginRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*PluginRegistration

	for _, registration := range r.plugins {
		if registration.Status != StatusActive {
			continue
		}

		// 检查排除列表
		excluded := false
		for _, excludeID := range criteria.ExcludePlugins {
			if registration.ID == excludeID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// 检查必需能力（这里简化为检查权限）
		hasRequiredCapabilities := true
		for _, capability := range criteria.RequiredCapabilities {
			found := false
			for _, permission := range registration.Permissions {
				if permission == capability {
					found = true
					break
				}
			}
			if !found {
				hasRequiredCapabilities = false
				break
			}
		}
		if !hasRequiredCapabilities {
			continue
		}

		regCopy := *registration
		candidates = append(candidates, &regCopy)
	}

	// 应用结果限制
	if criteria.MaxResults > 0 && len(candidates) > criteria.MaxResults {
		candidates = candidates[:criteria.MaxResults]
	}

	return candidates, nil
}

// Subscribe 订阅插件事件
func (r *DefaultPluginRegistry) Subscribe(ctx context.Context, eventTypes []string) (<-chan *PluginEvent, error) {
	eventChan := make(chan *PluginEvent, 100)

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, eventType := range eventTypes {
		if r.subscribers[eventType] == nil {
			r.subscribers[eventType] = make([]chan *PluginEvent, 0)
		}
		r.subscribers[eventType] = append(r.subscribers[eventType], eventChan)
	}

	return eventChan, nil
}

// Unsubscribe 取消订阅
func (r *DefaultPluginRegistry) Unsubscribe(ctx context.Context, subscription <-chan *PluginEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for eventType, channels := range r.subscribers {
		for i, ch := range channels {
			if ch == subscription {
				// 移除订阅
				r.subscribers[eventType] = append(channels[:i], channels[i+1:]...)
				close(ch)
				return nil
			}
		}
	}

	return fmt.Errorf("subscription not found")
}

// GetHealth 获取注册中心健康状态
func (r *DefaultPluginRegistry) GetHealth(ctx context.Context) (*HealthStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalPlugins := len(r.plugins)
	activePlugins := 0
	failedPlugins := 0

	for _, registration := range r.plugins {
		switch registration.Status {
		case StatusActive:
			activePlugins++
		case StatusInactive:
			failedPlugins++
		}
	}

	status := "healthy"
	if failedPlugins > totalPlugins/2 {
		status = "unhealthy"
	} else if failedPlugins > 0 {
		status = "degraded"
	}

	return &HealthStatus{
		Status:        status,
		TotalPlugins:  totalPlugins,
		ActivePlugins: activePlugins,
		FailedPlugins: failedPlugins,
		LastCheck:     time.Now(),
		Details: map[string]string{
			"registry_type": "default",
		},
	}, nil
}

// validateRegistration 验证注册信息
func (r *DefaultPluginRegistry) validateRegistration(registration *PluginRegistration) error {
	if registration.ID == "" {
		return fmt.Errorf("plugin ID is required")
	}
	if registration.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if registration.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	return nil
}

// matchesFilter 检查插件是否匹配过滤器
func (r *DefaultPluginRegistry) matchesFilter(registration *PluginRegistration, filter *PluginFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Category != "" && registration.Category != filter.Category {
		return false
	}

	if filter.Status != "" && registration.Status != filter.Status {
		return false
	}

	if filter.Author != "" && registration.Author != filter.Author {
		return false
	}

	// 检查标签匹配
	if len(filter.Tags) > 0 {
		hasMatchingTag := false
		for _, filterTag := range filter.Tags {
			for _, regTag := range registration.Tags {
				if regTag == filterTag {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if !hasMatchingTag {
			return false
		}
	}

	return true
}

// startEventProcessor 启动事件处理器
func (r *DefaultPluginRegistry) startEventProcessor() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		for {
			select {
			case <-r.ctx.Done():
				return
			case event := <-r.eventChan:
				r.processEvent(event)
			}
		}
	}()
}

// processEvent 处理事件
func (r *DefaultPluginRegistry) processEvent(event *PluginEvent) {
	r.mu.RLock()
	subscribers := r.subscribers[event.Type]
	r.mu.RUnlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- event:
		default:
			r.logger.Warn("Subscriber channel is full, dropping event")
		}
	}
}

// Shutdown 关闭注册中心
func (r *DefaultPluginRegistry) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down plugin registry")

	r.cancel()
	r.wg.Wait()

	// 关闭所有订阅通道
	r.mu.Lock()
	for _, channels := range r.subscribers {
		for _, ch := range channels {
			close(ch)
		}
	}
	r.subscribers = make(map[string][]chan *PluginEvent)
	r.mu.Unlock()

	close(r.eventChan)

	r.logger.Info("Plugin registry shutdown completed")
	return nil
}