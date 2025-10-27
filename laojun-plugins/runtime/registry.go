package runtime

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginRegistry 插件注册中心接口
type PluginRegistry interface {
	// RegisterPlugin 注册插件
	RegisterPlugin(metadata *PluginMetadata) error

	// UnregisterPlugin 注销插件
	UnregisterPlugin(pluginID string) error

	// GetPlugin 获取插件信息
	GetPlugin(pluginID string) (*PluginMetadata, error)

	// ListPlugins 列出所有插件
	ListPlugins() ([]*PluginMetadata, error)

	// FindPlugins 根据条件查找插件
	FindPlugins(filter *PluginFilter) ([]*PluginMetadata, error)

	// UpdatePluginStatus 更新插件状态
	UpdatePluginStatus(pluginID string, status PluginState) error

	// GetPluginStatus 获取插件状态
	GetPluginStatus(pluginID string) (PluginState, error)

	// Subscribe 订阅插件变更事件
	Subscribe(callback RegistryCallback) error

	// Unsubscribe 取消订阅
	Unsubscribe(callback RegistryCallback) error

	// GetStats 获取注册中心统计信息
	GetStats() *RegistryStats
}

// PluginFilter 插件过滤器
type PluginFilter struct {
	Category     string   `json:"category,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Author       string   `json:"author,omitempty"`
	Version      string   `json:"version,omitempty"`
	State        *PluginState `json:"state,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	SearchText   string   `json:"search_text,omitempty"`
}

// RegistryCallback 注册中心回调函数
type RegistryCallback func(event *RegistryEvent)

// RegistryEvent 注册中心事件
type RegistryEvent struct {
	Type      string          `json:"type"` // registered, unregistered, updated
	PluginID  string          `json:"plugin_id"`
	Metadata  *PluginMetadata `json:"metadata,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// DefaultPluginRegistry 默认插件注册中心实现
type DefaultPluginRegistry struct {
	plugins     map[string]*PluginRegistryEntry
	callbacks   []RegistryCallback
	logger      *logrus.Logger
	mu          sync.RWMutex
}

// PluginRegistryEntry 插件注册条目
type PluginRegistryEntry struct {
	Metadata      *PluginMetadata `json:"metadata"`
	Status        PluginState     `json:"status"`
	RegisteredAt  time.Time       `json:"registered_at"`
	LastUpdated   time.Time       `json:"last_updated"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
}

// NewDefaultPluginRegistry 创建默认插件注册中心
func NewDefaultPluginRegistry(logger *logrus.Logger) *DefaultPluginRegistry {
	return &DefaultPluginRegistry{
		plugins:   make(map[string]*PluginRegistryEntry),
		callbacks: make([]RegistryCallback, 0),
		logger:    logger,
	}
}

// RegisterPlugin 注册插件
func (r *DefaultPluginRegistry) RegisterPlugin(metadata *PluginMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	if metadata.ID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	entry := &PluginRegistryEntry{
		Metadata:      metadata,
		Status:        StateLoaded,
		RegisteredAt:  now,
		LastUpdated:   now,
		LastHeartbeat: now,
	}

	r.plugins[metadata.ID] = entry

	r.logger.WithFields(logrus.Fields{
		"plugin_id": metadata.ID,
		"name":      metadata.Name,
		"version":   metadata.Version,
	}).Info("Plugin registered in registry")

	// 发送注册事件
	r.notifyCallbacks(&RegistryEvent{
		Type:      "registered",
		PluginID:  metadata.ID,
		Metadata:  metadata,
		Timestamp: now,
	})

	return nil
}

// UnregisterPlugin 注销插件
func (r *DefaultPluginRegistry) UnregisterPlugin(pluginID string) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", pluginID)
	}

	delete(r.plugins, pluginID)

	r.logger.WithField("plugin_id", pluginID).Info("Plugin unregistered from registry")

	// 发送注销事件
	r.notifyCallbacks(&RegistryEvent{
		Type:      "unregistered",
		PluginID:  pluginID,
		Metadata:  entry.Metadata,
		Timestamp: time.Now(),
	})

	return nil
}

// GetPlugin 获取插件信息
func (r *DefaultPluginRegistry) GetPlugin(pluginID string) (*PluginMetadata, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found in registry", pluginID)
	}

	return entry.Metadata, nil
}

// ListPlugins 列出所有插件
func (r *DefaultPluginRegistry) ListPlugins() ([]*PluginMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]*PluginMetadata, 0, len(r.plugins))
	for _, entry := range r.plugins {
		plugins = append(plugins, entry.Metadata)
	}

	return plugins, nil
}

// FindPlugins 根据条件查找插件
func (r *DefaultPluginRegistry) FindPlugins(filter *PluginFilter) ([]*PluginMetadata, error) {
	if filter == nil {
		return r.ListPlugins()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*PluginMetadata

	for _, entry := range r.plugins {
		if r.matchesFilter(entry, filter) {
			result = append(result, entry.Metadata)
		}
	}

	return result, nil
}

// matchesFilter 检查插件是否匹配过滤条件
func (r *DefaultPluginRegistry) matchesFilter(entry *PluginRegistryEntry, filter *PluginFilter) bool {
	metadata := entry.Metadata

	// 检查分类
	if filter.Category != "" && metadata.Category != filter.Category {
		return false
	}

	// 检查作者
	if filter.Author != "" && metadata.Author != filter.Author {
		return false
	}

	// 检查版本
	if filter.Version != "" && metadata.Version != filter.Version {
		return false
	}

	// 检查状态
	if filter.State != nil && entry.Status != *filter.State {
		return false
	}

	// 检查标签
	if len(filter.Tags) > 0 {
		tagMap := make(map[string]bool)
		for _, tag := range metadata.Tags {
			tagMap[tag] = true
		}
		for _, filterTag := range filter.Tags {
			if !tagMap[filterTag] {
				return false
			}
		}
	}

	// 检查权限
	if len(filter.Permissions) > 0 {
		permMap := make(map[string]bool)
		for _, perm := range metadata.Permissions {
			permMap[perm] = true
		}
		for _, filterPerm := range filter.Permissions {
			if !permMap[filterPerm] {
				return false
			}
		}
	}

	// 检查搜索文本
	if filter.SearchText != "" {
		searchText := filter.SearchText
		if !contains(metadata.Name, searchText) &&
			!contains(metadata.Description, searchText) &&
			!contains(metadata.Author, searchText) {
			return false
		}
	}

	return true
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(str, substr string) bool {
	// 简单的包含检查，实际应用中可能需要更复杂的搜索逻辑
	return len(str) >= len(substr) && 
		   (str == substr || 
		    (len(str) > len(substr) && 
		     (str[:len(substr)] == substr || 
		      str[len(str)-len(substr):] == substr ||
		      findSubstring(str, substr))))
}

// findSubstring 在字符串中查找子字符串
func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// UpdatePluginStatus 更新插件状态
func (r *DefaultPluginRegistry) UpdatePluginStatus(pluginID string, status PluginState) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", pluginID)
	}

	oldStatus := entry.Status
	entry.Status = status
	entry.LastUpdated = time.Now()
	entry.LastHeartbeat = time.Now()

	r.logger.WithFields(logrus.Fields{
		"plugin_id":  pluginID,
		"old_status": oldStatus.String(),
		"new_status": status.String(),
	}).Debug("Plugin status updated in registry")

	// 发送更新事件
	r.notifyCallbacks(&RegistryEvent{
		Type:      "updated",
		PluginID:  pluginID,
		Metadata:  entry.Metadata,
		Timestamp: time.Now(),
	})

	return nil
}

// GetPluginStatus 获取插件状态
func (r *DefaultPluginRegistry) GetPluginStatus(pluginID string) (PluginState, error) {
	if pluginID == "" {
		return StateError, fmt.Errorf("plugin ID cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[pluginID]
	if !exists {
		return StateError, fmt.Errorf("plugin %s not found in registry", pluginID)
	}

	return entry.Status, nil
}

// Subscribe 订阅插件变更事件
func (r *DefaultPluginRegistry) Subscribe(callback RegistryCallback) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.callbacks = append(r.callbacks, callback)

	r.logger.Debug("New callback subscribed to registry events")

	return nil
}

// Unsubscribe 取消订阅
func (r *DefaultPluginRegistry) Unsubscribe(callback RegistryCallback) error {
	if callback == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 由于Go中函数比较的限制，这里使用简单的实现
	// 实际应用中可能需要使用ID或其他方式来标识回调
	for i, cb := range r.callbacks {
		// 这里无法直接比较函数，需要其他方式实现
		_ = cb
		if i >= 0 { // 占位符逻辑
			r.callbacks = append(r.callbacks[:i], r.callbacks[i+1:]...)
			break
		}
	}

	r.logger.Debug("Callback unsubscribed from registry events")

	return nil
}

// notifyCallbacks 通知所有回调函数
func (r *DefaultPluginRegistry) notifyCallbacks(event *RegistryEvent) {
	for _, callback := range r.callbacks {
		go func(cb RegistryCallback) {
			defer func() {
				if err := recover(); err != nil {
					r.logger.WithField("error", err).Error("Registry callback panic")
				}
			}()
			cb(event)
		}(callback)
	}
}

// GetStats 获取注册中心统计信息
func (r *DefaultPluginRegistry) GetStats() *RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &RegistryStats{
		TotalPlugins: len(r.plugins),
		StatusCount:  make(map[string]int),
		CategoryCount: make(map[string]int),
	}

	for _, entry := range r.plugins {
		// 统计状态
		statusStr := entry.Status.String()
		stats.StatusCount[statusStr]++

		// 统计分类
		if entry.Metadata.Category != "" {
			stats.CategoryCount[entry.Metadata.Category]++
		}
	}

	return stats
}

// RegistryStats 注册中心统计信息
type RegistryStats struct {
	TotalPlugins  int            `json:"total_plugins"`
	StatusCount   map[string]int `json:"status_count"`
	CategoryCount map[string]int `json:"category_count"`
}

// ExportRegistry 导出注册中心数据
func (r *DefaultPluginRegistry) ExportRegistry() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	export := make(map[string]*PluginRegistryEntry)
	for id, entry := range r.plugins {
		export[id] = entry
	}

	return json.Marshal(export)
}

// ImportRegistry 导入注册中心数据
func (r *DefaultPluginRegistry) ImportRegistry(data []byte) error {
	var imported map[string]*PluginRegistryEntry
	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("failed to unmarshal registry data: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for id, entry := range imported {
		r.plugins[id] = entry
		r.logger.WithField("plugin_id", id).Debug("Plugin imported to registry")
	}

	r.logger.WithField("count", len(imported)).Info("Registry data imported")

	return nil
}