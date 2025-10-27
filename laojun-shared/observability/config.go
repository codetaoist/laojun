package observability

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/codetaoist/laojun-shared/monitoring"
	"github.com/codetaoist/laojun-shared/tracing"
)

// 验证常量
const (
	MinServiceNameLength = 1
	MaxServiceNameLength = 100
	MinSampleRate        = 0.0
	MaxSampleRate        = 1.0
	MinTimeout           = 1 * time.Second
	MaxTimeout           = 5 * time.Minute
	MinBufferSize        = 1
	MaxBufferSize        = 100000
	MinFlushPeriod       = 1 * time.Second
	MaxFlushPeriod       = 1 * time.Hour
	MinBatchSize         = 1
	MaxBatchSize         = 10000
	MinMaxRetries        = 0
	MaxMaxRetries        = 10
	MinQueueSize         = 1
	MaxQueueSize         = 1000000
)

// 验证正则表达式
var (
	serviceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]$`)
	versionRegex     = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_\.]*$`)
	environmentRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_]*$`)
)

// 支持的导出格式
var supportedFormats = map[string]bool{
	"json":       true,
	"jaeger":     true,
	"otlp":       true,
	"prometheus": true,
	"zipkin":     true,
}

// Config 统一的可观测性配置
type Config struct {
	// 服务信息
	ServiceName    string `json:"service_name" yaml:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version"`
	Environment    string `json:"environment" yaml:"environment"`

	// 全局设置
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	SampleRate  float64       `json:"sample_rate" yaml:"sample_rate"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	BufferSize  int           `json:"buffer_size" yaml:"buffer_size"`
	FlushPeriod time.Duration `json:"flush_period" yaml:"flush_period"`

	// 监控配置
	Monitoring *monitoring.Config `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`

	// 追踪配置
	Tracing *tracing.Config `json:"tracing,omitempty" yaml:"tracing,omitempty"`

	// 导出配置
	Export *ExportConfig `json:"export,omitempty" yaml:"export,omitempty"`

	// 资源属性
	ResourceAttributes map[string]string `json:"resource_attributes,omitempty" yaml:"resource_attributes,omitempty"`
}

// ExportConfig 导出配置
type ExportConfig struct {
	// 导出端点
	Endpoints map[string]string `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`

	// 导出格式
	Formats []string `json:"formats,omitempty" yaml:"formats,omitempty"`

	// 导出头部
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// 导出超时
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// 重试配置
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`

	// 批量配置
	BatchSize     int           `json:"batch_size" yaml:"batch_size"`
	BatchTimeout  time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
	MaxQueueSize  int           `json:"max_queue_size" yaml:"max_queue_size"`
	DropOnFailure bool          `json:"drop_on_failure" yaml:"drop_on_failure"`
}

// DefaultConfig 返回默认的可观测性配置
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    "unknown-service",
		ServiceVersion: "unknown",
		Environment:    "development",
		Enabled:        true,
		SampleRate:     1.0,
		Timeout:        30 * time.Second,
		BufferSize:     1000,
		FlushPeriod:    10 * time.Second,
		Export: &ExportConfig{
			Endpoints: map[string]string{
				"jaeger":     "http://localhost:14268/api/traces",
				"prometheus": "http://localhost:9090/api/v1/write",
			},
			Formats:       []string{"json", "jaeger"},
			Timeout:       10 * time.Second,
			MaxRetries:    3,
			RetryDelay:    1 * time.Second,
			BatchSize:     100,
			BatchTimeout:  5 * time.Second,
			MaxQueueSize:  1000,
			DropOnFailure: false,
		},
		ResourceAttributes: map[string]string{
			"service.name":    "unknown-service",
			"service.version": "unknown",
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证服务名称
	if err := c.validateServiceName(); err != nil {
		return err
	}

	// 验证服务版本
	if err := c.validateServiceVersion(); err != nil {
		return err
	}

	// 验证环境
	if err := c.validateEnvironment(); err != nil {
		return err
	}

	// 验证采样率
	if err := c.validateSampleRate(); err != nil {
		return err
	}

	// 验证超时
	if err := c.validateTimeout(); err != nil {
		return err
	}

	// 验证缓冲区大小
	if err := c.validateBufferSize(); err != nil {
		return err
	}

	// 验证刷新周期
	if err := c.validateFlushPeriod(); err != nil {
		return err
	}

	// 验证导出配置
	if c.Export != nil {
		if err := c.Export.Validate(); err != nil {
			return fmt.Errorf("export config validation failed: %w", err)
		}
	}

	// 验证资源属性
	if err := c.validateResourceAttributes(); err != nil {
		return err
	}

	return nil
}

// validateServiceName 验证服务名称
func (c *Config) validateServiceName() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}

	if len(c.ServiceName) < MinServiceNameLength || len(c.ServiceName) > MaxServiceNameLength {
		return fmt.Errorf("service_name length must be between %d and %d characters", MinServiceNameLength, MaxServiceNameLength)
	}

	if !serviceNameRegex.MatchString(c.ServiceName) {
		return fmt.Errorf("service_name must contain only alphanumeric characters, hyphens, underscores, and dots, and must start and end with alphanumeric characters")
	}

	return nil
}

// validateServiceVersion 验证服务版本
func (c *Config) validateServiceVersion() error {
	if c.ServiceVersion == "" {
		return fmt.Errorf("service_version is required")
	}

	if !versionRegex.MatchString(c.ServiceVersion) {
		return fmt.Errorf("service_version must contain only alphanumeric characters, hyphens, underscores, and dots")
	}

	return nil
}

// validateEnvironment 验证环境
func (c *Config) validateEnvironment() error {
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if !environmentRegex.MatchString(c.Environment) {
		return fmt.Errorf("environment must contain only alphanumeric characters, hyphens, and underscores")
	}

	// 检查常见环境名称
	validEnvironments := []string{"development", "dev", "staging", "stage", "production", "prod", "test", "testing"}
	isValid := false
	for _, env := range validEnvironments {
		if strings.EqualFold(c.Environment, env) {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("environment should be one of: %s (case insensitive)", strings.Join(validEnvironments, ", "))
	}

	return nil
}

// validateSampleRate 验证采样率
func (c *Config) validateSampleRate() error {
	if c.SampleRate < MinSampleRate || c.SampleRate > MaxSampleRate {
		return fmt.Errorf("sample_rate must be between %.1f and %.1f", MinSampleRate, MaxSampleRate)
	}
	return nil
}

// validateTimeout 验证超时
func (c *Config) validateTimeout() error {
	if c.Timeout < MinTimeout || c.Timeout > MaxTimeout {
		return fmt.Errorf("timeout must be between %v and %v", MinTimeout, MaxTimeout)
	}
	return nil
}

// validateBufferSize 验证缓冲区大小
func (c *Config) validateBufferSize() error {
	if c.BufferSize < MinBufferSize || c.BufferSize > MaxBufferSize {
		return fmt.Errorf("buffer_size must be between %d and %d", MinBufferSize, MaxBufferSize)
	}
	return nil
}

// validateFlushPeriod 验证刷新周期
func (c *Config) validateFlushPeriod() error {
	if c.FlushPeriod < MinFlushPeriod || c.FlushPeriod > MaxFlushPeriod {
		return fmt.Errorf("flush_period must be between %v and %v", MinFlushPeriod, MaxFlushPeriod)
	}
	return nil
}

// validateResourceAttributes 验证资源属性
func (c *Config) validateResourceAttributes() error {
	if c.ResourceAttributes == nil {
		return nil
	}

	for key, value := range c.ResourceAttributes {
		if key == "" {
			return fmt.Errorf("resource attribute key cannot be empty")
		}
		if value == "" {
			return fmt.Errorf("resource attribute value for key '%s' cannot be empty", key)
		}
		if len(key) > 100 {
			return fmt.Errorf("resource attribute key '%s' is too long (max 100 characters)", key)
		}
		if len(value) > 500 {
			return fmt.Errorf("resource attribute value for key '%s' is too long (max 500 characters)", key)
		}
	}

	return nil
}

// Validate 验证导出配置
func (e *ExportConfig) Validate() error {
	// 验证超时
	if err := e.validateTimeout(); err != nil {
		return err
	}

	// 验证重试配置
	if err := e.validateRetryConfig(); err != nil {
		return err
	}

	// 验证批量配置
	if err := e.validateBatchConfig(); err != nil {
		return err
	}

	// 验证端点
	if err := e.validateEndpoints(); err != nil {
		return err
	}

	// 验证格式
	if err := e.validateFormats(); err != nil {
		return err
	}

	// 验证头部
	if err := e.validateHeaders(); err != nil {
		return err
	}

	return nil
}

// validateTimeout 验证导出超时
func (e *ExportConfig) validateTimeout() error {
	if e.Timeout <= 0 {
		return fmt.Errorf("export timeout must be positive")
	}
	if e.Timeout > 5*time.Minute {
		return fmt.Errorf("export timeout cannot exceed 5 minutes")
	}
	return nil
}

// validateRetryConfig 验证重试配置
func (e *ExportConfig) validateRetryConfig() error {
	if e.MaxRetries < MinMaxRetries || e.MaxRetries > MaxMaxRetries {
		return fmt.Errorf("max_retries must be between %d and %d", MinMaxRetries, MaxMaxRetries)
	}

	if e.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative")
	}

	if e.RetryDelay > 1*time.Minute {
		return fmt.Errorf("retry_delay cannot exceed 1 minute")
	}

	return nil
}

// validateBatchConfig 验证批量配置
func (e *ExportConfig) validateBatchConfig() error {
	if e.BatchSize < MinBatchSize || e.BatchSize > MaxBatchSize {
		return fmt.Errorf("batch_size must be between %d and %d", MinBatchSize, MaxBatchSize)
	}

	if e.BatchTimeout <= 0 {
		return fmt.Errorf("batch_timeout must be positive")
	}

	if e.BatchTimeout > 1*time.Hour {
		return fmt.Errorf("batch_timeout cannot exceed 1 hour")
	}

	if e.MaxQueueSize < MinQueueSize || e.MaxQueueSize > MaxQueueSize {
		return fmt.Errorf("max_queue_size must be between %d and %d", MinQueueSize, MaxQueueSize)
	}

	return nil
}

// validateEndpoints 验证端点
func (e *ExportConfig) validateEndpoints() error {
	if e.Endpoints == nil || len(e.Endpoints) == 0 {
		return fmt.Errorf("at least one export endpoint must be configured")
	}

	for name, endpoint := range e.Endpoints {
		if name == "" {
			return fmt.Errorf("endpoint name cannot be empty")
		}
		if endpoint == "" {
			return fmt.Errorf("endpoint URL for '%s' cannot be empty", name)
		}

		// 验证URL格式
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("invalid URL for endpoint '%s': %w", name, err)
		}

		// 确保URL有有效的scheme和host
		if parsedURL.Scheme == "" {
			return fmt.Errorf("endpoint URL for '%s' must include a scheme (http or https)", name)
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return fmt.Errorf("endpoint URL for '%s' must use http or https scheme", name)
		}
		if parsedURL.Host == "" {
			return fmt.Errorf("endpoint URL for '%s' must include a host", name)
		}
	}

	return nil
}

// validateFormats 验证格式
func (e *ExportConfig) validateFormats() error {
	if len(e.Formats) == 0 {
		return fmt.Errorf("at least one export format must be specified")
	}

	for _, format := range e.Formats {
		if !supportedFormats[format] {
			supportedList := make([]string, 0, len(supportedFormats))
			for f := range supportedFormats {
				supportedList = append(supportedList, f)
			}
			return fmt.Errorf("unsupported export format '%s', supported formats: %s", format, strings.Join(supportedList, ", "))
		}
	}

	return nil
}

// validateHeaders 验证头部
func (e *ExportConfig) validateHeaders() error {
	if e.Headers == nil {
		return nil
	}

	for key, value := range e.Headers {
		if key == "" {
			return fmt.Errorf("header key cannot be empty")
		}
		if value == "" {
			return fmt.Errorf("header value for key '%s' cannot be empty", key)
		}
		if len(key) > 100 {
			return fmt.Errorf("header key '%s' is too long (max 100 characters)", key)
		}
		if len(value) > 1000 {
			return fmt.Errorf("header value for key '%s' is too long (max 1000 characters)", key)
		}
	}

	return nil
}

// ApplyDefaults 应用默认值
func (c *Config) ApplyDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "unknown-service"
	}

	if c.ServiceVersion == "" {
		c.ServiceVersion = "unknown"
	}

	if c.Environment == "" {
		c.Environment = "development"
	}

	if c.SampleRate == 0 {
		c.SampleRate = 1.0
	}

	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}

	if c.BufferSize == 0 {
		c.BufferSize = 1000
	}

	if c.FlushPeriod == 0 {
		c.FlushPeriod = 10 * time.Second
	}

	// 应用导出配置默认值
	if c.Export != nil {
		c.Export.ApplyDefaults()
	}

	// 初始化资源属性
	if c.ResourceAttributes == nil {
		c.ResourceAttributes = make(map[string]string)
	}

	// 设置基本资源属性
	c.ResourceAttributes["service.name"] = c.ServiceName
	c.ResourceAttributes["service.version"] = c.ServiceVersion
	c.ResourceAttributes["environment"] = c.Environment
}

// ApplyDefaults 应用导出配置默认值
func (e *ExportConfig) ApplyDefaults() {
	if e.Timeout == 0 {
		e.Timeout = 10 * time.Second
	}

	if e.MaxRetries == 0 {
		e.MaxRetries = 3
	}

	if e.RetryDelay == 0 {
		e.RetryDelay = 1 * time.Second
	}

	if e.BatchSize == 0 {
		e.BatchSize = 100
	}

	if e.BatchTimeout == 0 {
		e.BatchTimeout = 5 * time.Second
	}

	if e.MaxQueueSize == 0 {
		e.MaxQueueSize = 1000
	}

	if e.Endpoints == nil {
		e.Endpoints = map[string]string{
			"jaeger":     "http://localhost:14268/api/traces",
			"prometheus": "http://localhost:9090/api/v1/write",
		}
	}

	if len(e.Formats) == 0 {
		e.Formats = []string{"json", "jaeger"}
	}
}

// ValidateAndApplyDefaults 验证配置并应用默认值
func (c *Config) ValidateAndApplyDefaults() error {
	// 先应用默认值
	c.ApplyDefaults()

	// 然后验证
	return c.Validate()
}

// GetMonitoringConfig 获取监控配置，如果为空则返回默认配置
func (c *Config) GetMonitoringConfig() *monitoring.Config {
	if c.Monitoring != nil {
		return c.Monitoring
	}

	// 返回默认监控配置
	defaultConfig := monitoring.DefaultConfig()
	defaultConfig.ServiceName = c.ServiceName
	defaultConfig.ServiceVersion = c.ServiceVersion
	return &defaultConfig
}

// GetTracingConfig 获取追踪配置，如果为空则返回默认配置
func (c *Config) GetTracingConfig() *tracing.Config {
	if c.Tracing != nil {
		return c.Tracing
	}

	// 返回默认追踪配置
	defaultConfig := tracing.DefaultConfig()
	defaultConfig.ServiceName = c.ServiceName
	defaultConfig.ServiceVersion = c.ServiceVersion
	defaultConfig.Environment = c.Environment
	return &defaultConfig
}

// EnableMonitoring 启用监控
func (c *Config) EnableMonitoring() *Config {
	if c.Monitoring == nil {
		c.Monitoring = c.GetMonitoringConfig()
	}
	return c
}

// EnableTracing 启用追踪
func (c *Config) EnableTracing() *Config {
	if c.Tracing == nil {
		c.Tracing = c.GetTracingConfig()
	}
	return c
}

// WithServiceInfo 设置服务信息
func (c *Config) WithServiceInfo(name, version, environment string) *Config {
	c.ServiceName = name
	c.ServiceVersion = version
	c.Environment = environment

	// 更新资源属性
	if c.ResourceAttributes == nil {
		c.ResourceAttributes = make(map[string]string)
	}
	c.ResourceAttributes["service.name"] = name
	c.ResourceAttributes["service.version"] = version
	c.ResourceAttributes["environment"] = environment

	return c
}

// WithExportEndpoints 设置导出端点
func (c *Config) WithExportEndpoints(endpoints map[string]string) *Config {
	if c.Export == nil {
		c.Export = &ExportConfig{}
	}
	if c.Export.Endpoints == nil {
		c.Export.Endpoints = make(map[string]string)
	}
	for k, v := range endpoints {
		c.Export.Endpoints[k] = v
	}
	return c
}

// WithExportFormats 设置导出格式
func (c *Config) WithExportFormats(formats ...string) *Config {
	if c.Export == nil {
		c.Export = &ExportConfig{}
	}
	c.Export.Formats = formats
	return c
}

// WithResourceAttribute 添加资源属性
func (c *Config) WithResourceAttribute(key, value string) *Config {
	if c.ResourceAttributes == nil {
		c.ResourceAttributes = make(map[string]string)
	}
	c.ResourceAttributes[key] = value
	return c
}

// WithExportHeader 添加导出头部
func (c *Config) WithExportHeader(key, value string) *Config {
	if c.Export == nil {
		c.Export = &ExportConfig{}
	}
	if c.Export.Headers == nil {
		c.Export.Headers = make(map[string]string)
	}
	c.Export.Headers[key] = value
	return c
}