package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
)

// Config 网关配置结构
type Config struct {
	Server        ServerConfig                `mapstructure:"server"`
	Redis         RedisConfig                 `mapstructure:"redis"`
	Discovery     DiscoveryConfig             `mapstructure:"discovery"`
	Auth          AuthConfig                  `mapstructure:"auth"`
	RateLimit     RateLimitConfig             `mapstructure:"ratelimit"`
	Proxy         ProxyConfig                 `mapstructure:"proxy"`
	Monitoring    MonitoringConfig            `mapstructure:"monitoring"`
	ConfigManager sharedconfig.ConfigManager `json:"-" mapstructure:"-"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
	Type   string            `mapstructure:"type"`
	Laojun LaojunConfig      `mapstructure:"laojun"`
	Consul ConsulConfig      `mapstructure:"consul"`
	Static map[string]string `mapstructure:"static"`
}

// LaojunConfig Laojun服务发现配置
type LaojunConfig struct {
	Address string `mapstructure:"address"`
	Scheme  string `mapstructure:"scheme"`
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address string `mapstructure:"address"`
	Scheme  string `mapstructure:"scheme"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret     string   `mapstructure:"jwt_secret"`
	TokenExpiry   int      `mapstructure:"token_expiry"`
	RefreshExpiry int      `mapstructure:"refresh_expiry"`
	WhiteList     []string `mapstructure:"whitelist"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled    bool              `mapstructure:"enabled"`
	GlobalRate int               `mapstructure:"global_rate"`
	UserRate   int               `mapstructure:"user_rate"`
	IPRate     int               `mapstructure:"ip_rate"`
	Rules      []RateLimitRule   `mapstructure:"rules"`
}

// RateLimitRule 限流规则
type RateLimitRule struct {
	Path   string `mapstructure:"path"`
	Method string `mapstructure:"method"`
	Rate   int    `mapstructure:"rate"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Timeout         int                    `mapstructure:"timeout"`
	RetryCount      int                    `mapstructure:"retry_count"`
	LoadBalancer    string                 `mapstructure:"load_balancer"`
	HealthCheck     HealthCheckConfig      `mapstructure:"health_check"`
	CircuitBreaker  CircuitBreakerConfig   `mapstructure:"circuit_breaker"`
	Routes          []RouteConfig          `mapstructure:"routes"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Interval int    `mapstructure:"interval"`
	Timeout  int    `mapstructure:"timeout"`
	Path     string `mapstructure:"path"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	FailureThreshold  int  `mapstructure:"failure_threshold"`
	RecoveryTimeout   int  `mapstructure:"recovery_timeout"`
	HalfOpenRequests  int  `mapstructure:"half_open_requests"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Path        string            `mapstructure:"path"`
	Method      string            `mapstructure:"method"`
	Service     string            `mapstructure:"service"`
	Target      string            `mapstructure:"target"`
	StripPrefix bool              `mapstructure:"strip_prefix"`
	Headers     map[string]string `mapstructure:"headers"`
	Auth        bool              `mapstructure:"auth"`
	RateLimit   *RateLimitRule    `mapstructure:"rate_limit"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	MetricsPath string `mapstructure:"metrics_path"`
	HealthPath  string `mapstructure:"health_path"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 设置配置文件名和搜索路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")  // 相对于服务目录
	viper.AddConfigPath(".")          // 当前目录

	// 设置环境变量前缀
	viper.SetEnvPrefix("GATEWAY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，使用默认配置
			fmt.Println("Config file not found, using defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
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
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// 服务发现默认配置
	viper.SetDefault("discovery.type", "static")
	viper.SetDefault("discovery.consul.address", "localhost:8500")
	viper.SetDefault("discovery.consul.scheme", "http")

	// 认证默认配置
	viper.SetDefault("auth.jwt_secret", "your-secret-key")
	viper.SetDefault("auth.token_expiry", 3600)
	viper.SetDefault("auth.refresh_expiry", 86400)

	// 限流默认配置
	viper.SetDefault("ratelimit.enabled", true)
	viper.SetDefault("ratelimit.global_rate", 1000)
	viper.SetDefault("ratelimit.user_rate", 100)
	viper.SetDefault("ratelimit.ip_rate", 50)

	// 代理默认配置
	viper.SetDefault("proxy.timeout", 30)
	viper.SetDefault("proxy.retry_count", 3)
	viper.SetDefault("proxy.load_balancer", "round_robin")
	viper.SetDefault("proxy.health_check.enabled", true)
	viper.SetDefault("proxy.health_check.interval", 30)
	viper.SetDefault("proxy.health_check.timeout", 5)
	viper.SetDefault("proxy.health_check.path", "/health")
	viper.SetDefault("proxy.circuit_breaker.enabled", true)
	viper.SetDefault("proxy.circuit_breaker.failure_threshold", 5)
	viper.SetDefault("proxy.circuit_breaker.recovery_timeout", 60)
	viper.SetDefault("proxy.circuit_breaker.half_open_requests", 3)

	// 监控默认配置
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.health_path", "/health")
}