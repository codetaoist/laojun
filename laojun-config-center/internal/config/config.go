package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置中心配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Storage  StorageConfig  `yaml:"storage"`
	Security SecurityConfig `yaml:"security"`
	Log      LogConfig      `yaml:"log"`
	Logging  LoggingConfig  `yaml:"logging"` // 兼容性字段
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string           `yaml:"type"`
	File     FileStorageConfig     `yaml:"file"`
	Redis    RedisStorageConfig    `yaml:"redis"`
	Database DatabaseStorageConfig `yaml:"database"`
}

// FileStorageConfig 文件存储配置
type FileStorageConfig struct {
	BasePath string `yaml:"basePath"`
	WatchDir bool   `yaml:"watchDir"`
}

// RedisStorageConfig Redis存储配置
type RedisStorageConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// DatabaseStorageConfig 数据库存储配置
type DatabaseStorageConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableAuth bool     `yaml:"enableAuth"`
	APIKey     string   `yaml:"apiKey"`
	AllowedIPs []string `yaml:"allowedIPs"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// LoggingConfig 详细日志配置（兼容性）
type LoggingConfig struct {
	Level  string         `yaml:"level"`
	Format string         `yaml:"format"`
	Output string         `yaml:"output"`
	File   LogFileConfig  `yaml:"file"`
}

// LogFileConfig 日志文件配置
type LogFileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    string `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
}

// Load 加载配置
func Load() (*Config, error) {
	configFile := getConfigFile()
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// 兼容性处理：如果Logging字段有值，使用它覆盖Log字段
	if config.Logging.Level != "" {
		config.Log.Level = config.Logging.Level
	}
	if config.Logging.Format != "" {
		config.Log.Format = config.Logging.Format
	}
	if config.Logging.Output != "" {
		config.Log.Output = config.Logging.Output
	}

	// 设置默认值
	setDefaults(&config)

	// 验证配置
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// getConfigFile 获取配置文件路径
func getConfigFile() string {
	// 优先使用环境变量指定的配置文件
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		return configFile
	}

	// 根据环境选择配置文件
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "local"
	}

	switch env {
	case "docker":
		return "config-center.docker.yaml"
	case "production", "prod":
		return "config-center.yaml"
	case "development", "dev":
		return "config-center.yaml"
	default:
		return "config-center.local.yaml"
	}
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8087
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}

	if config.Storage.Type == "" {
		config.Storage.Type = "file"
	}
	if config.Storage.File.BasePath == "" {
		config.Storage.File.BasePath = "./etc/laojun"
	}

	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "json"
	}
	if config.Log.Output == "" {
		config.Log.Output = "stdout"
	}
}

// validate 验证配置
func validate(config *Config) error {
	// 验证服务器配置
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// 验证存储配置
	switch config.Storage.Type {
	case "file":
		if config.Storage.File.BasePath == "" {
			return fmt.Errorf("file storage base path is required")
		}
	case "redis":
		if config.Storage.Redis.Host == "" {
			return fmt.Errorf("redis host is required")
		}
		if config.Storage.Redis.Port < 1 || config.Storage.Redis.Port > 65535 {
			return fmt.Errorf("invalid redis port: %d", config.Storage.Redis.Port)
		}
	case "database":
		if config.Storage.Database.Host == "" {
			return fmt.Errorf("database host is required")
		}
		if config.Storage.Database.Port < 1 || config.Storage.Database.Port > 65535 {
			return fmt.Errorf("invalid database port: %d", config.Storage.Database.Port)
		}
		if config.Storage.Database.User == "" {
			return fmt.Errorf("database user is required")
		}
		if config.Storage.Database.DBName == "" {
			return fmt.Errorf("database name is required")
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", config.Storage.Type)
	}

	// 验证日志配置
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[config.Log.Level] {
		return fmt.Errorf("invalid log level: %s", config.Log.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[config.Log.Format] {
		return fmt.Errorf("invalid log format: %s", config.Log.Format)
	}

	return nil
}

// Reload 重新加载配置
func (c *Config) Reload() error {
	newConfig, err := Load()
	if err != nil {
		return err
	}

	*c = *newConfig
	return nil
}