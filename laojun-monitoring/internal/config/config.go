package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
)

// Config 监控服务配置
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Logging        LoggingConfig        `mapstructure:"logging"`
	Metrics        MetricsConfig        `mapstructure:"metrics"`
	Alerting       AlertingConfig       `mapstructure:"alerting"`
	Storage        StorageConfig        `mapstructure:"storage"`
	Query          QueryConfig          `mapstructure:"query"`
	Exporters      ExportersConfig      `mapstructure:"exporters"`
	Collectors     CollectorsConfig     `mapstructure:"collectors"`
	Tracing        TracingConfig        `mapstructure:"tracing"`
	BatchProcessor BatchProcessorConfig `mapstructure:"batch_processor"`
	ConfigManager  sharedconfig.ConfigManager `mapstructure:"-"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Path            string        `mapstructure:"path"`
	Interval        time.Duration `mapstructure:"interval"`
	Retention       time.Duration `mapstructure:"retention"`
	EnableBuiltIn   bool          `mapstructure:"enable_builtin"`
	EnableCustom    bool          `mapstructure:"enable_custom"`
	MaxSeries       int           `mapstructure:"max_series"`
	MaxSamples      int           `mapstructure:"max_samples"`
}

// AlertingConfig 告警配置
type AlertingConfig struct {
	Enabled         bool                   `mapstructure:"enabled"`
	EvaluationInterval time.Duration       `mapstructure:"evaluation_interval"`
	Rules           []AlertRule            `mapstructure:"rules"`
	Receivers       map[string]Receiver    `mapstructure:"receivers"`
	Routes          []Route                `mapstructure:"routes"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string            `mapstructure:"name"`
	Query       string            `mapstructure:"query"`
	Duration    time.Duration     `mapstructure:"duration"`
	Severity    string            `mapstructure:"severity"`
	Labels      map[string]string `mapstructure:"labels"`
	Annotations map[string]string `mapstructure:"annotations"`
}

// Receiver 告警接收器
type Receiver struct {
	Name     string            `mapstructure:"name"`
	Type     string            `mapstructure:"type"`
	Config   map[string]string `mapstructure:"config"`
}

// Route 告警路由
type Route struct {
	Match    map[string]string `mapstructure:"match"`
	Receiver string            `mapstructure:"receiver"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string        `mapstructure:"type"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	InfluxDB   InfluxDBConfig   `mapstructure:"influxdb"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

// QueryConfig 查询配置
type QueryConfig struct {
	Type       string                   `mapstructure:"type"`
	Prometheus PrometheusQueryConfig    `mapstructure:"prometheus"`
}

// PrometheusQueryConfig Prometheus查询配置
type PrometheusQueryConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Auth    *AuthConfig   `mapstructure:"auth,omitempty"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type     string `mapstructure:"type"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Token    string `mapstructure:"token"`
}

// PrometheusConfig Prometheus 配置
type PrometheusConfig struct {
	URL      string        `mapstructure:"url"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
}

// InfluxDBConfig InfluxDB 配置
type InfluxDBConfig struct {
	URL      string `mapstructure:"url"`
	Token    string `mapstructure:"token"`
	Org      string `mapstructure:"org"`
	Bucket   string `mapstructure:"bucket"`
}

// ElasticsearchConfig Elasticsearch 配置
type ElasticsearchConfig struct {
	URLs     []string `mapstructure:"urls"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Index    string   `mapstructure:"index"`
}

// ExportersConfig 导出器配置
type ExportersConfig struct {
	Prometheus PrometheusExporterConfig `mapstructure:"prometheus"`
	Jaeger     JaegerExporterConfig     `mapstructure:"jaeger"`
	OTLP       OTLPExporterConfig       `mapstructure:"otlp"`
}

// PrometheusExporterConfig Prometheus 导出器配置
type PrometheusExporterConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// JaegerExporterConfig Jaeger 导出器配置
type JaegerExporterConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// OTLPExporterConfig OTLP 导出器配置
type OTLPExporterConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

// CollectorsConfig 收集器配置
type CollectorsConfig struct {
	System     SystemCollectorConfig     `mapstructure:"system"`
	Application ApplicationCollectorConfig `mapstructure:"application"`
	Network    NetworkCollectorConfig    `mapstructure:"network"`
	Custom     []CustomCollectorConfig   `mapstructure:"custom"`
}

// SystemCollectorConfig 系统收集器配置
type SystemCollectorConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	CPU      bool          `mapstructure:"cpu"`
	Memory   bool          `mapstructure:"memory"`
	Disk     bool          `mapstructure:"disk"`
	Network  bool          `mapstructure:"network"`
}

// ApplicationCollectorConfig 应用收集器配置
type ApplicationCollectorConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	JVM      bool          `mapstructure:"jvm"`
	GC       bool          `mapstructure:"gc"`
	Threads  bool          `mapstructure:"threads"`
}

// NetworkCollectorConfig 网络收集器配置
type NetworkCollectorConfig struct {
	Enabled   bool          `mapstructure:"enabled"`
	Interval  time.Duration `mapstructure:"interval"`
	Interface string        `mapstructure:"interface"`
}

// CustomCollectorConfig 自定义收集器配置
type CustomCollectorConfig struct {
	Name     string            `mapstructure:"name"`
	Type     string            `mapstructure:"type"`
	Interval time.Duration     `mapstructure:"interval"`
	Config   map[string]string `mapstructure:"config"`
}

// TracingConfig 链路追踪配置
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service_name"`
	Sampler     SamplerConfig `mapstructure:"sampler"`
	Jaeger      JaegerConfig  `mapstructure:"jaeger"`
}

// SamplerConfig 采样器配置
type SamplerConfig struct {
	Type  string  `mapstructure:"type"`
	Param float64 `mapstructure:"param"`
}

// JaegerConfig Jaeger 配置
type JaegerConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// BatchProcessorConfig 批量处理器配置
type BatchProcessorConfig struct {
	Enabled       bool                    `mapstructure:"enabled"`
	Name          string                  `mapstructure:"name"`
	BatchSize     int                     `mapstructure:"batch_size"`
	FlushInterval time.Duration           `mapstructure:"flush_interval"`
	BufferSize    int                     `mapstructure:"buffer_size"`
	MaxRetries    int                     `mapstructure:"max_retries"`
	RetryDelay    time.Duration           `mapstructure:"retry_delay"`
	Exporters     []BatchExporterConfig   `mapstructure:"exporters"`
}

// IsEnabled 返回是否启用
func (c *BatchProcessorConfig) IsEnabled() bool {
	return c.Enabled
}

// BatchExporterConfig 批量导出器配置
type BatchExporterConfig struct {
	Type     string                 `mapstructure:"type"`
	Name     string                 `mapstructure:"name"`
	Enabled  bool                   `mapstructure:"enabled"`
	Config   map[string]interface{} `mapstructure:"config"`
}

// IsEnabled 检查配置是否启用
func (c *BatchExporterConfig) IsEnabled() bool {
	return c.Enabled
}

// GetName 获取配置名称
func (c *BatchExporterConfig) GetName() string {
	return c.Name
}

// GetConfig 返回配置
func (c *BatchExporterConfig) GetConfig() map[string]interface{} {
	return c.Config
}

// SetConfig 设置配置项
func (c *BatchExporterConfig) SetConfig(key string, value interface{}) {
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}
	c.Config[key] = value
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")  // 相对于服务目录
	viper.AddConfigPath(".")          // 当前目录

	// 设置环境变量前缀
	viper.SetEnvPrefix("LAOJUN_MONITORING")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// LoadWithConfigManager 加载配置并集成统一配置管理器
func LoadWithConfigManager(configManager sharedconfig.ConfigManager) (*Config, error) {
	config, err := Load()
	if err != nil {
		return nil, err
	}
	
	// 集成统一配置管理器
	config.ConfigManager = configManager
	
	return config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8082)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)

	// 指标默认配置
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.interval", "15s")
	viper.SetDefault("metrics.retention", "24h")
	viper.SetDefault("metrics.enable_builtin", true)
	viper.SetDefault("metrics.enable_custom", true)
	viper.SetDefault("metrics.max_series", 100000)
	viper.SetDefault("metrics.max_samples", 1000000)

	// 告警默认配置
	viper.SetDefault("alerting.enabled", true)
	viper.SetDefault("alerting.evaluation_interval", "30s")

	// 存储默认配置
	viper.SetDefault("storage.type", "prometheus")
	viper.SetDefault("storage.prometheus.url", "http://localhost:9090")
	viper.SetDefault("storage.prometheus.timeout", "30s")

	// 查询默认配置
	viper.SetDefault("query.type", "memory")
	viper.SetDefault("query.prometheus.url", "http://localhost:9090")
	viper.SetDefault("query.prometheus.timeout", "30s")

	// 导出器默认配置
	viper.SetDefault("exporters.prometheus.enabled", true)
	viper.SetDefault("exporters.prometheus.port", 8082)
	viper.SetDefault("exporters.prometheus.path", "/metrics")

	// 收集器默认配置
	viper.SetDefault("collectors.system.enabled", true)
	viper.SetDefault("collectors.system.interval", "15s")
	viper.SetDefault("collectors.system.cpu", true)
	viper.SetDefault("collectors.system.memory", true)
	viper.SetDefault("collectors.system.disk", true)
	viper.SetDefault("collectors.system.network", true)

	// 链路追踪默认配置
	viper.SetDefault("tracing.enabled", false)
	viper.SetDefault("tracing.service_name", "laojun-monitoring")
	viper.SetDefault("tracing.sampler.type", "const")
	viper.SetDefault("tracing.sampler.param", 1.0)

	// 批量处理器默认配置
	viper.SetDefault("batch_processor.enabled", true)
	viper.SetDefault("batch_processor.name", "default-batch-processor")
	viper.SetDefault("batch_processor.batch_size", 100)
	viper.SetDefault("batch_processor.flush_interval", "5s")
	viper.SetDefault("batch_processor.buffer_size", 1000)
	viper.SetDefault("batch_processor.max_retries", 3)
	viper.SetDefault("batch_processor.retry_delay", "1s")
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	if path := os.Getenv("LAOJUN_MONITORING_CONFIG_PATH"); path != "" {
		return path
	}
	return "./configs/config.yaml"
}