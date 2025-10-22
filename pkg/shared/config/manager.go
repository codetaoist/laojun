package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager 配置管理
type Manager struct {
	client      *Client
	logger      *logrus.Logger
	watchers    map[string][]func(string, interface{})
	watchersMux sync.RWMutex
	autoRefresh bool
	refreshTick *time.Ticker
	stopChan    chan struct{}
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	Client          *Client
	AutoRefresh     bool
	RefreshInterval time.Duration
}

// NewManager 创建配置管理
func NewManager(config ManagerConfig) *Manager {
	if config.RefreshInterval == 0 {
		config.RefreshInterval = 30 * time.Second
	}

	manager := &Manager{
		client:      config.Client,
		logger:      logrus.New(),
		watchers:    make(map[string][]func(string, interface{})),
		autoRefresh: config.AutoRefresh,
		stopChan:    make(chan struct{}),
	}

	if config.AutoRefresh {
		manager.startAutoRefresh(config.RefreshInterval)
	}

	return manager
}

// LoadFromEnv 从环境变量加载配置到结构
func (m *Manager) LoadFromEnv(target interface{}) error {
	return m.loadFromSource(target, "env")
}

// LoadFromConfigCenter 从配置中心加载配置到结构
func (m *Manager) LoadFromConfigCenter(ctx context.Context, target interface{}) error {
	return m.loadFromSource(target, "config", ctx)
}

// LoadHybrid 混合加载：优先环境变量，然后配置中心
func (m *Manager) LoadHybrid(ctx context.Context, target interface{}) error {
	// 先从配置中心加载
	if err := m.LoadFromConfigCenter(ctx, target); err != nil {
		m.logger.Warnf("Failed to load from config center: %v", err)
	}

	// 再从环境变量覆盖
	if err := m.LoadFromEnv(target); err != nil {
		return fmt.Errorf("failed to load from environment: %w", err)
	}

	return nil
}

// loadFromSource 从指定源加载配置
func (m *Manager) loadFromSource(target interface{}, source string, args ...interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		if !field.CanSet() {
			continue
		}

		var key string
		var defaultValue string

		switch source {
		case "env":
			key = fieldType.Tag.Get("env")
			defaultValue = fieldType.Tag.Get("default")
		case "config":
			key = fieldType.Tag.Get("config")
			defaultValue = fieldType.Tag.Get("default")
		}

		if key == "" {
			continue
		}

		var value interface{}
		var err error

		switch source {
		case "env":
			value = os.Getenv(key)
			if value == "" && defaultValue != "" {
				value = defaultValue
			}
		case "config":
			if len(args) == 0 {
				return fmt.Errorf("context required for config center")
			}
			ctx := args[0].(context.Context)
			value, err = m.client.Get(ctx, key)
			if err != nil && defaultValue != "" {
				value = defaultValue
			}
		}

		if err != nil {
			m.logger.Warnf("Failed to get config %s: %v", key, err)
			continue
		}

		if err := m.setFieldValue(field, value); err != nil {
			m.logger.Errorf("Failed to set field %s: %v", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func (m *Manager) setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else {
			field.SetString(fmt.Sprintf("%v", value))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intVal int64
		switch v := value.(type) {
		case int:
			intVal = int64(v)
		case int64:
			intVal = v
		case float64:
			intVal = int64(v)
		case string:
			var err error
			intVal, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot convert %v to int: %w", v, err)
			}
		default:
			return fmt.Errorf("cannot convert %T to int", v)
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var uintVal uint64
		switch v := value.(type) {
		case uint:
			uintVal = uint64(v)
		case uint64:
			uintVal = v
		case float64:
			uintVal = uint64(v)
		case string:
			var err error
			uintVal, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot convert %v to uint: %w", v, err)
			}
		default:
			return fmt.Errorf("cannot convert %T to uint", v)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		var floatVal float64
		switch v := value.(type) {
		case float32:
			floatVal = float64(v)
		case float64:
			floatVal = v
		case string:
			var err error
			floatVal, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("cannot convert %v to float: %w", v, err)
			}
		default:
			return fmt.Errorf("cannot convert %T to float", v)
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		var boolVal bool
		switch v := value.(type) {
		case bool:
			boolVal = v
		case string:
			var err error
			boolVal, err = strconv.ParseBool(v)
			if err != nil {
				return fmt.Errorf("cannot convert %v to bool: %w", v, err)
			}
		default:
			return fmt.Errorf("cannot convert %T to bool", v)
		}
		field.SetBool(boolVal)

	case reflect.Slice:
		if str, ok := value.(string); ok {
			// 假设是逗号分隔的字符串
			parts := strings.Split(str, ",")
			slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
			for i, part := range parts {
				if err := m.setFieldValue(slice.Index(i), strings.TrimSpace(part)); err != nil {
					return err
				}
			}
			field.Set(slice)
		}

	default:
		// 对于复杂类型，尝试直接设置值
		if reflect.TypeOf(value).AssignableTo(field.Type()) {
			field.Set(reflect.ValueOf(value))
		} else {
			return fmt.Errorf("unsupported field type: %s", field.Kind())
		}
	}

	return nil
}

// Watch 监听配置变化
func (m *Manager) Watch(key string, callback func(string, interface{})) {
	m.watchersMux.Lock()
	defer m.watchersMux.Unlock()

	if m.watchers[key] == nil {
		m.watchers[key] = make([]func(string, interface{}), 0)
	}
	m.watchers[key] = append(m.watchers[key], callback)
}

// Unwatch 取消监听
func (m *Manager) Unwatch(key string) {
	m.watchersMux.Lock()
	defer m.watchersMux.Unlock()

	delete(m.watchers, key)
}

// notifyWatchers 通知监听者配置变化
func (m *Manager) notifyWatchers(key string, value interface{}) {
	m.watchersMux.RLock()
	defer m.watchersMux.RUnlock()

	if watchers, exists := m.watchers[key]; exists {
		for _, callback := range watchers {
			go callback(key, value)
		}
	}
}

// startAutoRefresh 启动自动刷新
func (m *Manager) startAutoRefresh(interval time.Duration) {
	m.refreshTick = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-m.refreshTick.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := m.client.RefreshCache(ctx); err != nil {
					m.logger.Errorf("Failed to refresh cache: %v", err)
				}
				cancel()
			case <-m.stopChan:
				return
			}
		}
	}()
}

// Stop 停止管理自动刷新
func (m *Manager) Stop() {
	if m.refreshTick != nil {
		m.refreshTick.Stop()
	}
	close(m.stopChan)
}

// GetClient 获取客户
func (m *Manager) GetClient() *Client {
	return m.client
}

// ValidateConfig 验证配置结构
func (m *Manager) ValidateConfig(target interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// 检查必需字段
		if required := fieldType.Tag.Get("required"); required == "true" {
			if field.Kind() == reflect.String && field.String() == "" {
				return fmt.Errorf("required field %s is empty", fieldType.Name)
			}
			if field.Kind() == reflect.Int && field.Int() == 0 {
				return fmt.Errorf("required field %s is zero", fieldType.Name)
			}
		}

		// 检查字段范围
		if min := fieldType.Tag.Get("min"); min != "" {
			if field.Kind() == reflect.Int {
				minVal, err := strconv.ParseInt(min, 10, 64)
				if err == nil && field.Int() < minVal {
					return fmt.Errorf("field %s value %d is less than minimum %d", fieldType.Name, field.Int(), minVal)
				}
			}
		}

		if max := fieldType.Tag.Get("max"); max != "" {
			if field.Kind() == reflect.Int {
				maxVal, err := strconv.ParseInt(max, 10, 64)
				if err == nil && field.Int() > maxVal {
					return fmt.Errorf("field %s value %d is greater than maximum %d", fieldType.Name, field.Int(), maxVal)
				}
			}
		}
	}

	return nil
}
