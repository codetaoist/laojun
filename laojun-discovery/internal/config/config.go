package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
)

// Config 应用配置
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Registry    RegistryConfig    `yaml:"registry"`
	Storage     StorageConfig     `yaml:"storage"`
	Health      HealthConfig      `yaml:"health"`
	Log         LogConfig         `yaml:"log"`
	LoadBalance LoadBalanceConfig `yaml:"load_balance"`
	Circuit     CircuitConfig     `yaml:"circuit"`
	RateLimit   RateLimitConfig   `yaml:"rate_limit"`
	
	// 统一配置管理器
	ConfigManager sharedconfig.ConfigManager `yaml:"-"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string      `mapstructure:"type"`
	Redis RedisConfig `mapstructure:"redis"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RegistryConfig 注册表配置
type RegistryConfig struct {
	TTL                int  `mapstructure:"ttl"`
	CleanupInterval    int  `mapstructure:"cleanup_interval"`
	EnableAutoCleanup  bool `mapstructure:"enable_auto_cleanup"`
	MaxServices        int  `mapstructure:"max_services"`
	MaxInstancesPerSvc int  `mapstructure:"max_instances_per_service"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	CheckInterval   int  `mapstructure:"check_interval"`
	Timeout         int  `mapstructure:"timeout"`
	FailureThreshold int  `mapstructure:"failure_threshold"`
	SuccessThreshold int  `mapstructure:"success_threshold"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	MetricsPath string `mapstructure:"metrics_path"`
	HealthPath  string `mapstructure:"health_path"`
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address    string `mapstructure:"address"`
	Scheme     string `mapstructure:"scheme"`
	Datacenter string `mapstructure:"datacenter"`
	Token      string `mapstructure:"token"`
}

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	Timeout   int      `mapstructure:"timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// LoadBalanceConfig 负载均衡配置
type LoadBalanceConfig struct {
	Algorithm          string         `mapstructure:"algorithm"`
	HealthCheckEnabled bool           `mapstructure:"health_check_enabled"`
	StatsEnabled       bool           `mapstructure:"stats_enabled"`
	HashKey            string         `mapstructure:"hash_key"`
	Weights            map[string]int `mapstructure:"weights"`
}

// CircuitConfig 熔断器配置
type CircuitConfig struct {
	Enabled          bool    `mapstructure:"enabled"`
	FailureThreshold uint32  `mapstructure:"failure_threshold"`
	SuccessThreshold uint32  `mapstructure:"success_threshold"`
	Timeout          int     `mapstructure:"timeout"`
	MaxRequests      uint32  `mapstructure:"max_requests"`
	Interval         int     `mapstructure:"interval"`
	MinRequests      uint32  `mapstructure:"min_requests"`
	FailureRatio     float64 `mapstructure:"failure_ratio"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	Algorithm string  `mapstructure:"algorithm"`
	Rate      float64 `mapstructure:"rate"`
	Burst     int     `mapstructure:"burst"`
	Window    int     `mapstructure:"window"`
	Capacity  int     `mapstructure:"capacity"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 设置配置文件名和搜索路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")  // 相对于服务目录
	viper.AddConfigPath(".")          // 当前目录

	// 设置环境变量前缀
	viper.SetEnvPrefix("DISCOVERY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 从环境变量覆盖敏感配置
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Storage.Redis.Password = redisPassword
	}
	if consulToken := os.Getenv("CONSUL_TOKEN"); consulToken != "" {
		config.Consul.Token = consulToken
	}
	if etcdPassword := os.Getenv("ETCD_PASSWORD"); etcdPassword != "" {
		config.Etcd.Password = etcdPassword
	}

	return &config, nil
}

// LoadWithConfigManager 使用统一配置管理器加载配置
func LoadWithConfigManager(configManager sharedconfig.ConfigManager) (*Config, error) {
	// 先加载传统配置
	config, err := Load()
	if err != nil {
		return nil, err
	}
	
	// 设置统一配置管理器
	config.ConfigManager = configManager
	
	return config, nil
}

// setDefaults 设置默认值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", 8084)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// 存储默认配置
	viper.SetDefault("storage.type", "redis")
	viper.SetDefault("storage.redis.host", "localhost")
	viper.SetDefault("storage.redis.port", 6379)
	viper.SetDefault("storage.redis.password", "")
	viper.SetDefault("storage.redis.db", 0)

	// 注册表默认配置
	viper.SetDefault("registry.ttl", 30)
	viper.SetDefault("registry.cleanup_interval", 60)
	viper.SetDefault("registry.enable_auto_cleanup", true)
	viper.SetDefault("registry.max_services", 1000)
	viper.SetDefault("registry.max_instances_per_service", 100)

	// 健康检查默认配置
	viper.SetDefault("health.enabled", true)
	viper.SetDefault("health.check_interval", 30)
	viper.SetDefault("health.timeout", 5)
	viper.SetDefault("health.failure_threshold", 3)
	viper.SetDefault("health.success_threshold", 2)

	// 监控默认配置
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.health_path", "/health")

	// Consul默认配置
	viper.SetDefault("consul.address", "localhost:8500")
	viper.SetDefault("consul.scheme", "http")
	viper.SetDefault("consul.datacenter", "dc1")

	// Etcd默认配置
	viper.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	viper.SetDefault("etcd.timeout", 5)

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
	viper.SetDefault("log.compress", true)

	// 负载均衡默认配置
	viper.SetDefault("load_balance.algorithm", "round_robin")
	viper.SetDefault("load_balance.health_check_enabled", true)
	viper.SetDefault("load_balance.stats_enabled", true)
	viper.SetDefault("load_balance.hash_key", "")

	// 熔断器默认配置
	viper.SetDefault("circuit.enabled", true)
	viper.SetDefault("circuit.failure_threshold", 5)
	viper.SetDefault("circuit.success_threshold", 3)
	viper.SetDefault("circuit.timeout", 60)
	viper.SetDefault("circuit.max_requests", 10)
	viper.SetDefault("circuit.interval", 60)
	viper.SetDefault("circuit.min_requests", 10)
	viper.SetDefault("circuit.failure_ratio", 0.5)

	// 限流默认配置
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.algorithm", "token_bucket")
	viper.SetDefault("rate_limit.rate", 100.0)
	viper.SetDefault("rate_limit.burst", 200)
	viper.SetDefault("rate_limit.window", 60)
	viper.SetDefault("rate_limit.capacity", 1000)
}