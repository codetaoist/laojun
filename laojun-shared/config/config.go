package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 应用配置结构
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Redis     RedisConfig     `json:"redis"`
	JWT       JWTConfig       `json:"jwt"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	Log       LogConfig       `json:"log"`
	Security  SecurityConfig  `json:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Mode         string        `json:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	DBName          string        `json:"dbname"`
	SSLMode         string        `json:"sslmode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `json:"secret"`
	Expiration time.Duration `json:"expiration"`
	Issuer     string        `json:"issuer"`
}

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	Enabled              bool          `json:"enabled"`
	GlobalRequests       int           `json:"global_requests"`
	GlobalWindow         time.Duration `json:"global_window"`
	UserRequests         int           `json:"user_requests"`
	UserWindow           time.Duration `json:"user_window"`
	LoginRequests        int           `json:"login_requests"`
	LoginWindowMinutes   int           `json:"login_window_minutes"`
	APIRequests          int           `json:"api_requests"`
	APIWindowHours       int           `json:"api_window_hours"`
	ErrorMessage         string        `json:"error_message"`
	ErrorDetail          string        `json:"error_detail"`
	RetryAfterSeconds    int           `json:"retry_after_seconds"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // json, text
	Output string `json:"output"` // stdout, file
	File   string `json:"file"`
}

// SecurityConfig 安全相关配置
type SecurityConfig struct {
	EnableCaptcha          bool          `json:"enable_captcha"`
	CaptchaTTL            time.Duration `json:"captcha_ttl"`
	AdminCaptchaEnabled   bool          `json:"admin_captcha_enabled"`
	MarketplaceCaptchaEnabled bool      `json:"marketplace_captcha_enabled"`
	CaptchaType           string        `json:"captcha_type"`
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			Mode:         getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "laojun"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
			Issuer:     getEnv("JWT_ISSUER", "laojun"),
		},
		RateLimit: RateLimitConfig{
			Enabled:              getEnvAsBool("RATE_LIMIT_ENABLED", true),
			GlobalRequests:       getEnvAsInt("RATE_LIMIT_GLOBAL_REQUESTS", 1000),
			GlobalWindow:         getEnvAsDuration("RATE_LIMIT_GLOBAL_WINDOW", time.Minute),
			UserRequests:         getEnvAsInt("RATE_LIMIT_USER_REQUESTS", 100),
			UserWindow:           getEnvAsDuration("RATE_LIMIT_USER_WINDOW", time.Minute),
			LoginRequests:        getEnvAsInt("RATE_LIMIT_LOGIN_REQUESTS", 5),
			LoginWindowMinutes:   getEnvAsInt("RATE_LIMIT_LOGIN_WINDOW_MINUTES", 15),
			APIRequests:          getEnvAsInt("RATE_LIMIT_API_REQUESTS", 100),
			APIWindowHours:       getEnvAsInt("RATE_LIMIT_API_WINDOW_HOURS", 1),
			ErrorMessage:         getEnv("RATE_LIMIT_ERROR_MESSAGE", "请求过于频繁"),
			ErrorDetail:          getEnv("RATE_LIMIT_ERROR_DETAIL", "您的请求频率超过了限制，请稍后再试"),
			RetryAfterSeconds:    getEnvAsInt("RATE_LIMIT_RETRY_AFTER_SECONDS", 900),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
			File:   getEnv("LOG_FILE", "app.log"),
		},
		Security: SecurityConfig{
			EnableCaptcha:             getEnvAsBool("SECURITY_ENABLE_CAPTCHA", true),
			CaptchaTTL:               getEnvAsDuration("SECURITY_CAPTCHA_TTL", 2*time.Minute),
			AdminCaptchaEnabled:      getEnvAsBool("ADMIN_CAPTCHA_ENABLED", false),
			MarketplaceCaptchaEnabled: getEnvAsBool("MARKETPLACE_CAPTCHA_ENABLED", false),
			CaptchaType:              getEnv("CAPTCHA_TYPE", "image"),
		},
	}

	return config, nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
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

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
