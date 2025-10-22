package config

import (
	"os"
	"strconv"
	"time"
)

// Config 配置中心自身的配置
type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Storage  StorageConfig  `yaml:"storage" json:"storage"`
	Security SecurityConfig `yaml:"security" json:"security"`
	Log      LogConfig      `yaml:"log" json:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `yaml:"host" json:"host"`
	Port         int           `yaml:"port" json:"port"`
	ReadTimeout  time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string      `yaml:"type" json:"type"` // file, redis, database
	File  FileConfig  `yaml:"file" json:"file"`
	Redis RedisConfig `yaml:"redis" json:"redis"`
	DB    DBConfig    `yaml:"database" json:"database"`
}

// FileConfig 文件存储配置
type FileConfig struct {
	BasePath string `yaml:"basePath" json:"basePath"`
	WatchDir bool   `yaml:"watchDir" json:"watchDir"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"dbname" json:"dbname"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	APIKey    string   `yaml:"apiKey" json:"apiKey"`
	AllowedIPs []string `yaml:"allowedIPs" json:"allowedIPs"`
	EnableAuth bool     `yaml:"enableAuth" json:"enableAuth"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("CONFIG_CENTER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("CONFIG_CENTER_PORT", 8090),
			ReadTimeout:  getEnvAsDuration("CONFIG_CENTER_READ_TIMEOUT", "30s"),
			WriteTimeout: getEnvAsDuration("CONFIG_CENTER_WRITE_TIMEOUT", "30s"),
		},
		Storage: StorageConfig{
			Type: getEnv("CONFIG_STORAGE_TYPE", "file"),
			File: FileConfig{
				BasePath: getEnv("CONFIG_FILE_BASE_PATH", "./etc/laojun"),
				WatchDir: getEnvAsBool("CONFIG_FILE_WATCH_DIR", true),
			},
			Redis: RedisConfig{
				Host:     getEnv("CONFIG_REDIS_HOST", "localhost"),
				Port:     getEnvAsInt("CONFIG_REDIS_PORT", 6379),
				Password: getEnv("CONFIG_REDIS_PASSWORD", ""),
				DB:       getEnvAsInt("CONFIG_REDIS_DB", 1),
			},
			DB: DBConfig{
				Host:     getEnv("CONFIG_DB_HOST", "localhost"),
				Port:     getEnvAsInt("CONFIG_DB_PORT", 5432),
				User:     getEnv("CONFIG_DB_USER", "laojun"),
				Password: getEnv("CONFIG_DB_PASSWORD", "change-me"),
				DBName:   getEnv("CONFIG_DB_NAME", "laojun_config"),
				SSLMode:  getEnv("CONFIG_DB_SSLMODE", "disable"),
			},
		},
		Security: SecurityConfig{
			APIKey:     getEnv("CONFIG_API_KEY", ""),
			AllowedIPs: getEnvAsSlice("CONFIG_ALLOWED_IPS", []string{"127.0.0.1", "::1"}),
			EnableAuth: getEnvAsBool("CONFIG_ENABLE_AUTH", false),
		},
		Log: LogConfig{
			Level:  getEnv("CONFIG_LOG_LEVEL", "info"),
			Format: getEnv("CONFIG_LOG_FORMAT", "json"),
			Output: getEnv("CONFIG_LOG_OUTPUT", "stdout"),
		},
	}

	return config, nil
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 30 * time.Second
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// 简单的逗号分隔解析
		return []string{value}
	}
	return defaultValue
}
