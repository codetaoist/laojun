package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// LoadConfig 加载配置
	LoadConfig(configPath string, target interface{}) error
	
	// SaveConfig 保存配置
	SaveConfig(configPath string, config interface{}) error
	
	// ValidateConfig 验证配置
	ValidateConfig(config interface{}) error
	
	// GetConfigSchema 获取配置模式
	GetConfigSchema(configType reflect.Type) (*ConfigSchema, error)
	
	// MergeConfigs 合并配置
	MergeConfigs(base, override interface{}) (interface{}, error)
	
	// WatchConfig 监听配置变化
	WatchConfig(configPath string, callback func(interface{})) error
}

// ConfigSchema 配置模式定义
type ConfigSchema struct {
	Type        string                    `json:"type" yaml:"type"`
	Properties  map[string]*PropertySpec  `json:"properties" yaml:"properties"`
	Required    []string                  `json:"required" yaml:"required"`
	Description string                    `json:"description" yaml:"description"`
}

// PropertySpec 属性规范
type PropertySpec struct {
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description" yaml:"description"`
	Default     interface{} `json:"default" yaml:"default"`
	Required    bool        `json:"required" yaml:"required"`
	Enum        []string    `json:"enum,omitempty" yaml:"enum,omitempty"`
	Pattern     string      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MinLength   *int        `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength   *int        `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
}

// DefaultConfigManager 默认配置管理器
type DefaultConfigManager struct {
	schemas map[string]*ConfigSchema
	mu      sync.RWMutex
}

// NewDefaultConfigManager 创建默认配置管理器
func NewDefaultConfigManager() *DefaultConfigManager {
	return &DefaultConfigManager{
		schemas: make(map[string]*ConfigSchema),
	}
}

// LoadConfig 加载配置
func (m *DefaultConfigManager) LoadConfig(configPath string, target interface{}) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 根据文件扩展名选择解析器
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".json":
		return json.Unmarshal(data, target)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, target)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// SaveConfig 保存配置
func (m *DefaultConfigManager) SaveConfig(configPath string, config interface{}) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 根据文件扩展名选择序列化器
	ext := strings.ToLower(filepath.Ext(configPath))
	var data []byte
	var err error

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	return ioutil.WriteFile(configPath, data, 0644)
}

// ValidateConfig 验证配置
func (m *DefaultConfigManager) ValidateConfig(config interface{}) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	configType := reflect.TypeOf(config)
	if configType.Kind() == reflect.Ptr {
		configType = configType.Elem()
	}

	// 获取配置模式
	schema, err := m.GetConfigSchema(configType)
	if err != nil {
		return fmt.Errorf("failed to get config schema: %w", err)
	}

	// 验证配置
	return m.validateAgainstSchema(config, schema)
}

// GetConfigSchema 获取配置模式
func (m *DefaultConfigManager) GetConfigSchema(configType reflect.Type) (*ConfigSchema, error) {
	typeName := configType.Name()
	
	m.mu.RLock()
	if schema, exists := m.schemas[typeName]; exists {
		m.mu.RUnlock()
		return schema, nil
	}
	m.mu.RUnlock()

	// 生成配置模式
	schema, err := m.generateSchema(configType)
	if err != nil {
		return nil, err
	}

	// 缓存模式
	m.mu.Lock()
	m.schemas[typeName] = schema
	m.mu.Unlock()

	return schema, nil
}

// generateSchema 生成配置模式
func (m *DefaultConfigManager) generateSchema(t reflect.Type) (*ConfigSchema, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config type must be a struct")
	}

	schema := &ConfigSchema{
		Type:       "object",
		Properties: make(map[string]*PropertySpec),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		// 获取字段标签
		jsonTag := field.Tag.Get("json")
		yamlTag := field.Tag.Get("yaml")
		configTag := field.Tag.Get("config")

		// 确定字段名
		fieldName := field.Name
		if jsonTag != "" && jsonTag != "-" {
			fieldName = strings.Split(jsonTag, ",")[0]
		} else if yamlTag != "" && yamlTag != "-" {
			fieldName = strings.Split(yamlTag, ",")[0]
		}

		// 跳过忽略的字段
		if jsonTag == "-" || yamlTag == "-" {
			continue
		}

		// 创建属性规范
		propSpec := &PropertySpec{
			Type:        m.getFieldType(field.Type),
			Description: field.Tag.Get("description"),
		}

		// 解析配置标签
		if configTag != "" {
			m.parseConfigTag(configTag, propSpec)
		}

		// 检查是否必需
		if strings.Contains(configTag, "required") {
			propSpec.Required = true
			schema.Required = append(schema.Required, fieldName)
		}

		schema.Properties[fieldName] = propSpec
	}

	return schema, nil
}

// getFieldType 获取字段类型
func (m *DefaultConfigManager) getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	case reflect.Ptr:
		return m.getFieldType(t.Elem())
	default:
		return "string"
	}
}

// parseConfigTag 解析配置标签
func (m *DefaultConfigManager) parseConfigTag(tag string, spec *PropertySpec) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "required" {
			spec.Required = true
		} else if strings.HasPrefix(part, "default=") {
			defaultValue := strings.TrimPrefix(part, "default=")
			spec.Default = defaultValue
		} else if strings.HasPrefix(part, "min=") {
			minStr := strings.TrimPrefix(part, "min=")
			if min, err := parseFloat(minStr); err == nil {
				spec.Minimum = &min
			}
		} else if strings.HasPrefix(part, "max=") {
			maxStr := strings.TrimPrefix(part, "max=")
			if max, err := parseFloat(maxStr); err == nil {
				spec.Maximum = &max
			}
		} else if strings.HasPrefix(part, "pattern=") {
			spec.Pattern = strings.TrimPrefix(part, "pattern=")
		}
	}
}

// parseFloat 解析浮点数
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// validateAgainstSchema 根据模式验证配置
func (m *DefaultConfigManager) validateAgainstSchema(config interface{}, schema *ConfigSchema) error {
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}

	// 检查必需字段
	for _, required := range schema.Required {
		if prop, exists := schema.Properties[required]; exists {
			fieldValue := m.getFieldValue(configValue, required)
			if !fieldValue.IsValid() || fieldValue.IsZero() {
				return fmt.Errorf("required field '%s' is missing or empty", required)
			}
			
			// 验证字段值
			if err := m.validateFieldValue(fieldValue, prop); err != nil {
				return fmt.Errorf("validation failed for field '%s': %w", required, err)
			}
		}
	}

	return nil
}

// getFieldValue 获取字段值
func (m *DefaultConfigManager) getFieldValue(structValue reflect.Value, fieldName string) reflect.Value {
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// 检查JSON标签
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && strings.Split(jsonTag, ",")[0] == fieldName {
			return structValue.Field(i)
		}
		
		// 检查YAML标签
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" && strings.Split(yamlTag, ",")[0] == fieldName {
			return structValue.Field(i)
		}
		
		// 检查字段名
		if field.Name == fieldName {
			return structValue.Field(i)
		}
	}
	
	return reflect.Value{}
}

// validateFieldValue 验证字段值
func (m *DefaultConfigManager) validateFieldValue(value reflect.Value, spec *PropertySpec) error {
	if !value.IsValid() {
		return nil
	}

	switch spec.Type {
	case "string":
		if value.Kind() != reflect.String {
			return fmt.Errorf("expected string, got %s", value.Kind())
		}
		str := value.String()
		
		// 检查长度限制
		if spec.MinLength != nil && len(str) < *spec.MinLength {
			return fmt.Errorf("string length %d is less than minimum %d", len(str), *spec.MinLength)
		}
		if spec.MaxLength != nil && len(str) > *spec.MaxLength {
			return fmt.Errorf("string length %d is greater than maximum %d", len(str), *spec.MaxLength)
		}
		
		// 检查枚举值
		if len(spec.Enum) > 0 {
			found := false
			for _, enum := range spec.Enum {
				if str == enum {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value '%s' is not in allowed enum values: %v", str, spec.Enum)
			}
		}

	case "integer":
		if !value.Type().ConvertibleTo(reflect.TypeOf(int64(0))) {
			return fmt.Errorf("expected integer, got %s", value.Kind())
		}
		num := value.Convert(reflect.TypeOf(int64(0))).Int()
		
		// 检查数值范围
		if spec.Minimum != nil && float64(num) < *spec.Minimum {
			return fmt.Errorf("value %d is less than minimum %f", num, *spec.Minimum)
		}
		if spec.Maximum != nil && float64(num) > *spec.Maximum {
			return fmt.Errorf("value %d is greater than maximum %f", num, *spec.Maximum)
		}

	case "number":
		if !value.Type().ConvertibleTo(reflect.TypeOf(float64(0))) {
			return fmt.Errorf("expected number, got %s", value.Kind())
		}
		num := value.Convert(reflect.TypeOf(float64(0))).Float()
		
		// 检查数值范围
		if spec.Minimum != nil && num < *spec.Minimum {
			return fmt.Errorf("value %f is less than minimum %f", num, *spec.Minimum)
		}
		if spec.Maximum != nil && num > *spec.Maximum {
			return fmt.Errorf("value %f is greater than maximum %f", num, *spec.Maximum)
		}

	case "boolean":
		if value.Kind() != reflect.Bool {
			return fmt.Errorf("expected boolean, got %s", value.Kind())
		}
	}

	return nil
}

// MergeConfigs 合并配置
func (m *DefaultConfigManager) MergeConfigs(base, override interface{}) (interface{}, error) {
	if base == nil {
		return override, nil
	}
	if override == nil {
		return base, nil
	}

	baseValue := reflect.ValueOf(base)
	overrideValue := reflect.ValueOf(override)

	if baseValue.Type() != overrideValue.Type() {
		return nil, fmt.Errorf("config types do not match")
	}

	// 创建结果副本
	result := reflect.New(baseValue.Type()).Elem()
	result.Set(baseValue)

	// 合并字段
	return m.mergeStructs(result, overrideValue).Interface(), nil
}

// mergeStructs 合并结构体
func (m *DefaultConfigManager) mergeStructs(base, override reflect.Value) reflect.Value {
	if base.Kind() == reflect.Ptr {
		base = base.Elem()
	}
	if override.Kind() == reflect.Ptr {
		override = override.Elem()
	}

	for i := 0; i < override.NumField(); i++ {
		overrideField := override.Field(i)
		baseField := base.Field(i)

		if !overrideField.IsZero() {
			if baseField.CanSet() {
				baseField.Set(overrideField)
			}
		}
	}

	return base
}

// WatchConfig 监听配置变化
func (m *DefaultConfigManager) WatchConfig(configPath string, callback func(interface{})) error {
	// 这里可以实现文件监听逻辑
	// 由于简化实现，这里只是一个占位符
	return fmt.Errorf("config watching not implemented yet")
}

// PluginConfig 插件配置基础结构
type PluginConfig struct {
	Name        string            `json:"name" yaml:"name" config:"required"`
	Version     string            `json:"version" yaml:"version" config:"required"`
	Description string            `json:"description" yaml:"description"`
	Author      string            `json:"author" yaml:"author"`
	License     string            `json:"license" yaml:"license"`
	Homepage    string            `json:"homepage" yaml:"homepage"`
	Repository  string            `json:"repository" yaml:"repository"`
	Tags        []string          `json:"tags" yaml:"tags"`
	Categories  []string          `json:"categories" yaml:"categories"`
	Settings    map[string]string `json:"settings" yaml:"settings"`
	Resources   *ResourceConfig   `json:"resources" yaml:"resources"`
	Security    *SecurityConfig   `json:"security" yaml:"security"`
	Logging     *LoggingConfig    `json:"logging" yaml:"logging"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	CPU    string `json:"cpu" yaml:"cpu" config:"default=100m"`
	Memory string `json:"memory" yaml:"memory" config:"default=128Mi"`
	Disk   string `json:"disk" yaml:"disk" config:"default=1Gi"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Permissions []string `json:"permissions" yaml:"permissions"`
	Sandbox     bool     `json:"sandbox" yaml:"sandbox" config:"default=true"`
	NetworkAccess bool   `json:"networkAccess" yaml:"networkAccess" config:"default=false"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `json:"level" yaml:"level" config:"default=info"`
	Format string `json:"format" yaml:"format" config:"default=json"`
	Output string `json:"output" yaml:"output" config:"default=stdout"`
}