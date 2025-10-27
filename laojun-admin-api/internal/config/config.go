package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	Expiration time.Duration `yaml:"expiration"`
}

type LogConfig struct {
	Level    string `yaml:"level"`
	Filename string `yaml:"filename"`
	MaxSize  int    `yaml:"max_size"`
	MaxAge   int    `yaml:"max_age"`
}

type RateLimitConfig struct {
	Enabled            bool   `yaml:"enabled"`
	Rate               int    `yaml:"rate"`
	Burst              int    `yaml:"burst"`
	ErrorMessage       string `yaml:"error_message"`
	ErrorDetail        string `yaml:"error_detail"`
	RetryAfterSeconds  int    `yaml:"retry_after_seconds"`
	LoginRequests      int    `yaml:"login_requests"`
	LoginWindowMinutes int    `yaml:"login_window_minutes"`
	APIRequests        int    `yaml:"api_requests"`
	APIWindowHours     int    `yaml:"api_window_hours"`
}