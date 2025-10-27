package config

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DefaultConfigManager 默认配置管理器实现
type DefaultConfigManager struct {
	storage   ConfigStorage
	watcher   ConfigWatcher
	validator ConfigValidator
	logger    *logrus.Logger
	options   *ConfigOptions

	// 缓存
	cache    map[string]*CacheItem
	cacheMux sync.RWMutex

	// 监听器
	watchers    map[string][]ConfigChangeCallback
	watchersMux sync.RWMutex

	// 状态
	closed bool
	mu     sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
	Version   int64
}

// NewDefaultConfigManager 创建默认配置管理器
func NewDefaultConfigManager(storage ConfigStorage, options *ConfigOptions) *DefaultConfigManager {
	if options == nil {
		options = &ConfigOptions{
			CacheEnabled:    true,
			CacheTTL:        5 * time.Minute,
			CacheSize:       1000,
			WatchEnabled:    true,
			WatchInterval:   30 * time.Second,
			WatchBufferSize: 100,
		}
	}

	manager := &DefaultConfigManager{
		storage:  storage,
		logger:   logrus.New(),
		options:  options,
		cache:    make(map[string]*CacheItem),
		watchers: make(map[string][]ConfigChangeCallback),
	}

	// 初始化监听器
	if options.WatchEnabled {
		manager.watcher = NewDefaultConfigWatcher(storage, options)
	}

	// 初始化验证器
	manager.validator = NewDefaultConfigValidator()

	return manager
}

// Get 获取配置
func (m *DefaultConfigManager) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil, fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	// 检查缓存
	if m.options.CacheEnabled {
		if value, found := m.getFromCache(key); found {
			return value, nil
		}
	}

	// 从存储获取
	item, err := m.storage.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get config from storage: %w", err)
	}

	// 更新缓存
	if m.options.CacheEnabled {
		m.setToCache(key, item.Value, item.Version)
	}

	return item.Value, nil
}

// GetString 获取字符串配置
func (m *DefaultConfigManager) GetString(ctx context.Context, key string) (string, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return "", err
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// GetInt 获取整数配置
func (m *DefaultConfigManager) GetInt(ctx context.Context, key string) (int, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// GetBool 获取布尔配置
func (m *DefaultConfigManager) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// GetFloat64 获取浮点数配置
func (m *DefaultConfigManager) GetFloat64(ctx context.Context, key string) (float64, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// GetDuration 获取时间间隔配置
func (m *DefaultConfigManager) GetDuration(ctx context.Context, key string) (time.Duration, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		return time.ParseDuration(v)
	case int64:
		return time.Duration(v), nil
	case int:
		return time.Duration(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to duration", v)
	}
}

// GetStringSlice 获取字符串切片配置
func (m *DefaultConfigManager) GetStringSlice(ctx context.Context, key string) ([]string, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result, nil
	case string:
		// 尝试解析JSON数组
		var result []string
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			return result, nil
		}
		// 按逗号分割
		return strings.Split(v, ","), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []string", v)
	}
}

// GetStringMap 获取字符串映射配置
func (m *DefaultConfigManager) GetStringMap(ctx context.Context, key string) (map[string]interface{}, error) {
	value, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case map[string]interface{}:
		return v, nil
	case string:
		// 尝试解析JSON对象
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			return result, nil
		}
		return nil, fmt.Errorf("cannot parse string as JSON object: %w", err)
	default:
		return nil, fmt.Errorf("cannot convert %T to map[string]interface{}", v)
	}
}

// Set 设置配置
func (m *DefaultConfigManager) Set(ctx context.Context, key string, value interface{}) error {
	return m.SetWithTTL(ctx, key, value, 0)
}

// SetWithTTL 设置带TTL的配置
func (m *DefaultConfigManager) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	// 创建配置项
	item := &ConfigItem{
		Key:       key,
		Value:     value,
		Type:      m.inferConfigType(value),
		TTL:       ttl,
		Version:   time.Now().UnixNano(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if ttl > 0 {
		expiresAt := time.Now().Add(ttl)
		item.ExpiresAt = &expiresAt
	}

	// 验证配置
	if m.validator != nil {
		if err := m.validator.Validate(ctx, item); err != nil {
			return fmt.Errorf("validate config: %w", err)
		}
	}

	// 获取旧值用于事件
	var oldValue interface{}
	if oldItem, err := m.storage.Get(ctx, key); err == nil {
		oldValue = oldItem.Value
	}

	// 保存到存储
	if err := m.storage.Set(ctx, item); err != nil {
		return fmt.Errorf("set config to storage: %w", err)
	}

	// 更新缓存
	if m.options.CacheEnabled {
		m.setToCache(key, value, item.Version)
	}

	// 触发变化事件
	m.triggerChangeEvent(&ConfigChangeEvent{
		Type:      EventTypeUpdate,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		Version:   item.Version,
		Timestamp: time.Now(),
	})

	return nil
}

// Delete 删除配置
func (m *DefaultConfigManager) Delete(ctx context.Context, key string) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	// 获取旧值用于事件
	var oldValue interface{}
	if oldItem, err := m.storage.Get(ctx, key); err == nil {
		oldValue = oldItem.Value
	}

	// 从存储删除
	if err := m.storage.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete config from storage: %w", err)
	}

	// 从缓存删除
	if m.options.CacheEnabled {
		m.removeFromCache(key)
	}

	// 触发变化事件
	m.triggerChangeEvent(&ConfigChangeEvent{
		Type:      EventTypeDelete,
		Key:       key,
		OldValue:  oldValue,
		Version:   time.Now().UnixNano(),
		Timestamp: time.Now(),
	})

	return nil
}

// Exists 检查配置是否存在
func (m *DefaultConfigManager) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return false, fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	return m.storage.Exists(ctx, key)
}

// Keys 获取所有配置键
func (m *DefaultConfigManager) Keys(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil, fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	items, err := m.storage.List(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("list configs from storage: %w", err)
	}

	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = item.Key
	}

	return keys, nil
}

// Watch 监听配置变化
func (m *DefaultConfigManager) Watch(ctx context.Context, key string, callback ConfigChangeCallback) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	m.watchersMux.Lock()
	defer m.watchersMux.Unlock()

	if m.watchers[key] == nil {
		m.watchers[key] = make([]ConfigChangeCallback, 0)
	}
	m.watchers[key] = append(m.watchers[key], callback)

	// 如果有监听器，添加到监听器
	if m.watcher != nil {
		return m.watcher.AddKey(key, callback)
	}

	return nil
}

// Unwatch 取消监听配置变化
func (m *DefaultConfigManager) Unwatch(ctx context.Context, key string) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	m.watchersMux.Lock()
	defer m.watchersMux.Unlock()

	delete(m.watchers, key)

	// 如果有监听器，从监听器移除
	if m.watcher != nil {
		return m.watcher.RemoveKey(key)
	}

	return nil
}

// GetMultiple 批量获取配置
func (m *DefaultConfigManager) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil, fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	result := make(map[string]interface{})

	// 检查缓存
	uncachedKeys := make([]string, 0)
	if m.options.CacheEnabled {
		for _, key := range keys {
			if value, found := m.getFromCache(key); found {
				result[key] = value
			} else {
				uncachedKeys = append(uncachedKeys, key)
			}
		}
	} else {
		uncachedKeys = keys
	}

	// 从存储获取未缓存的配置
	if len(uncachedKeys) > 0 {
		items, err := m.storage.GetMultiple(ctx, uncachedKeys)
		if err != nil {
			return nil, fmt.Errorf("get multiple configs from storage: %w", err)
		}

		for key, item := range items {
			result[key] = item.Value
			// 更新缓存
			if m.options.CacheEnabled {
				m.setToCache(key, item.Value, item.Version)
			}
		}
	}

	return result, nil
}

// SetMultiple 批量设置配置
func (m *DefaultConfigManager) SetMultiple(ctx context.Context, configs map[string]interface{}) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	items := make([]*ConfigItem, 0, len(configs))
	now := time.Now()

	for key, value := range configs {
		item := &ConfigItem{
			Key:       key,
			Value:     value,
			Type:      m.inferConfigType(value),
			Version:   now.UnixNano(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		// 验证配置
		if m.validator != nil {
			if err := m.validator.Validate(ctx, item); err != nil {
				return fmt.Errorf("validate config %s: %w", key, err)
			}
		}

		items = append(items, item)
	}

	// 批量保存到存储
	if err := m.storage.SetMultiple(ctx, items); err != nil {
		return fmt.Errorf("set multiple configs to storage: %w", err)
	}

	// 更新缓存并触发事件
	for key, value := range configs {
		if m.options.CacheEnabled {
			m.setToCache(key, value, now.UnixNano())
		}

		// 触发变化事件
		m.triggerChangeEvent(&ConfigChangeEvent{
			Type:      EventTypeUpdate,
			Key:       key,
			NewValue:  value,
			Version:   now.UnixNano(),
			Timestamp: now,
		})
	}

	return nil
}

// GetVersion 获取配置版本
func (m *DefaultConfigManager) GetVersion(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return 0, fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	item, err := m.storage.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("get config from storage: %w", err)
	}

	return item.Version, nil
}

// GetHistory 获取配置历史
func (m *DefaultConfigManager) GetHistory(ctx context.Context, key string, limit int) ([]ConfigHistory, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil, fmt.Errorf("config manager is closed")
	}
	m.mu.RUnlock()

	return m.storage.GetVersions(ctx, key, limit)
}

// Health 健康检查
func (m *DefaultConfigManager) Health(ctx context.Context) (*ConfigHealth, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return &ConfigHealth{
			Status:    HealthStatusUnhealthy,
			Message:   "config manager is closed",
			Timestamp: time.Now(),
		}, nil
	}
	m.mu.RUnlock()

	start := time.Now()
	err := m.storage.Health(ctx)
	latency := time.Since(start)

	health := &ConfigHealth{
		Timestamp: time.Now(),
		Latency:   latency,
		Details:   make(map[string]string),
	}

	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = err.Error()
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "OK"
	}

	// 添加缓存信息
	if m.options.CacheEnabled {
		m.cacheMux.RLock()
		health.Details["cache_size"] = fmt.Sprintf("%d", len(m.cache))
		m.cacheMux.RUnlock()
	}

	// 添加监听器信息
	m.watchersMux.RLock()
	health.Details["watchers"] = fmt.Sprintf("%d", len(m.watchers))
	m.watchersMux.RUnlock()

	return health, nil
}

// Close 关闭配置管理器
func (m *DefaultConfigManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true

	// 关闭监听器
	if m.watcher != nil {
		if err := m.watcher.Stop(); err != nil {
			m.logger.WithError(err).Error("Failed to stop config watcher")
		}
	}

	// 关闭存储
	if err := m.storage.Close(); err != nil {
		m.logger.WithError(err).Error("Failed to close config storage")
		return err
	}

	// 清理缓存
	m.cacheMux.Lock()
	m.cache = nil
	m.cacheMux.Unlock()

	// 清理监听器
	m.watchersMux.Lock()
	m.watchers = nil
	m.watchersMux.Unlock()

	return nil
}

// 缓存相关方法

func (m *DefaultConfigManager) getFromCache(key string) (interface{}, bool) {
	m.cacheMux.RLock()
	defer m.cacheMux.RUnlock()

	item, exists := m.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		delete(m.cache, key)
		return nil, false
	}

	return item.Value, true
}

func (m *DefaultConfigManager) setToCache(key string, value interface{}, version int64) {
	if !m.options.CacheEnabled {
		return
	}

	m.cacheMux.Lock()
	defer m.cacheMux.Unlock()

	// 检查缓存大小限制
	if len(m.cache) >= m.options.CacheSize {
		// 简单的LRU：删除最旧的项
		var oldestKey string
		var oldestTime time.Time
		for k, v := range m.cache {
			if oldestKey == "" || v.ExpiresAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.ExpiresAt
			}
		}
		if oldestKey != "" {
			delete(m.cache, oldestKey)
		}
	}

	m.cache[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(m.options.CacheTTL),
		Version:   version,
	}
}

func (m *DefaultConfigManager) removeFromCache(key string) {
	m.cacheMux.Lock()
	defer m.cacheMux.Unlock()
	delete(m.cache, key)
}

// 辅助方法

func (m *DefaultConfigManager) inferConfigType(value interface{}) ConfigType {
	switch value.(type) {
	case string:
		return ConfigTypeString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return ConfigTypeInt
	case bool:
		return ConfigTypeBool
	case float32, float64:
		return ConfigTypeFloat
	case []interface{}, []string:
		return ConfigTypeArray
	case map[string]interface{}:
		return ConfigTypeObject
	default:
		return ConfigTypeString
	}
}

func (m *DefaultConfigManager) triggerChangeEvent(event *ConfigChangeEvent) {
	m.watchersMux.RLock()
	callbacks := m.watchers[event.Key]
	m.watchersMux.RUnlock()

	for _, callback := range callbacks {
		go func(cb ConfigChangeCallback) {
			if err := cb(event); err != nil {
				m.logger.WithError(err).WithField("key", event.Key).Error("Config change callback failed")
			}
		}(callback)
	}
}
