package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoad(t *testing.T) {
	// 保存原始配置
	originalViper := viper.GetViper()
	defer func() {
		viper.Reset()
		for key, value := range originalViper.AllSettings() {
			viper.Set(key, value)
		}
	}()

	t.Run("加载默认配置", func(t *testing.T) {
		// 重置viper
		viper.Reset()

		config, err := Load()
		if err != nil {
			t.Fatalf("Expected no error loading default config, got %v", err)
		}

		// 验证默认值
		if config.Server.Host != "0.0.0.0" {
			t.Errorf("Expected default host '0.0.0.0', got '%s'", config.Server.Host)
		}
		if config.Server.Port != 8082 {
			t.Errorf("Expected default port 8082, got %d", config.Server.Port)
		}
		if config.Logging.Level != "info" {
			t.Errorf("Expected default log level 'info', got '%s'", config.Logging.Level)
		}
		if !config.Metrics.Enabled {
			t.Error("Expected metrics to be enabled by default")
		}
	})

	t.Run("从配置文件加载", func(t *testing.T) {
		// 创建临时配置文件
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		configContent := `
server:
  host: "127.0.0.1"
  port: 9090
  mode: "release"

logging:
  level: "debug"
  format: "text"

metrics:
  enabled: false
  path: "/custom-metrics"
`

		configFile := filepath.Join(tempDir, "config.yaml")
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// 重置viper并设置配置路径
		viper.Reset()
		
		// 直接使用viper读取配置文件进行测试
		viper.SetConfigFile(configFile)
		err = viper.ReadInConfig()
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		// 设置默认值
		setDefaults()

		// 解析配置
		var config Config
		err = viper.Unmarshal(&config)
		if err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// 验证配置文件中的值
		if config.Server.Host != "127.0.0.1" {
			t.Errorf("Expected host '127.0.0.1', got '%s'", config.Server.Host)
		}
		if config.Server.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", config.Server.Port)
		}
		if config.Server.Mode != "release" {
			t.Errorf("Expected mode 'release', got '%s'", config.Server.Mode)
		}
		if config.Logging.Level != "debug" {
			t.Errorf("Expected log level 'debug', got '%s'", config.Logging.Level)
		}
		if config.Metrics.Enabled {
			t.Error("Expected metrics to be disabled")
		}
		if config.Metrics.Path != "/custom-metrics" {
			t.Errorf("Expected metrics path '/custom-metrics', got '%s'", config.Metrics.Path)
		}
	})

	t.Run("环境变量覆盖", func(t *testing.T) {
		// 设置环境变量
		os.Setenv("LAOJUN_MONITORING_SERVER_HOST", "192.168.1.1")
		os.Setenv("LAOJUN_MONITORING_SERVER_PORT", "8080")
		defer func() {
			os.Unsetenv("LAOJUN_MONITORING_SERVER_HOST")
			os.Unsetenv("LAOJUN_MONITORING_SERVER_PORT")
		}()

		// 重置viper
		viper.Reset()
		
		// 设置环境变量前缀和自动环境变量
		viper.SetEnvPrefix("LAOJUN_MONITORING")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		
		// 绑定特定的环境变量
		viper.BindEnv("server.host", "LAOJUN_MONITORING_SERVER_HOST")
		viper.BindEnv("server.port", "LAOJUN_MONITORING_SERVER_PORT")
		
		// 设置默认值
		setDefaults()

		// 解析配置
		var config Config
		err := viper.Unmarshal(&config)
		if err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// 验证环境变量覆盖了默认值
		if config.Server.Host != "192.168.1.1" {
			t.Errorf("Expected host from env var '192.168.1.1', got '%s'", config.Server.Host)
		}
		if config.Server.Port != 8080 {
			t.Errorf("Expected port from env var 8080, got %d", config.Server.Port)
		}
	})

	t.Run("无效配置文件", func(t *testing.T) {
		// 创建无效的配置文件
		tempDir, err := os.MkdirTemp("", "config_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		invalidConfig := `
server:
  host: "127.0.0.1"
  port: "invalid_port"  # 无效的端口类型
`

		configFile := filepath.Join(tempDir, "config.yaml")
		err = os.WriteFile(configFile, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		// 重置viper并设置配置路径
		viper.Reset()
		viper.SetConfigFile(configFile)
		
		err = viper.ReadInConfig()
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		// 设置默认值
		setDefaults()

		// 尝试解析配置，应该失败
		var config Config
		err = viper.Unmarshal(&config)
		if err == nil {
			t.Error("Expected error when unmarshaling invalid config file")
		}
	})
}

func TestSetDefaults(t *testing.T) {
	// 保存原始配置
	originalViper := viper.GetViper()
	defer func() {
		viper.Reset()
		for key, value := range originalViper.AllSettings() {
			viper.Set(key, value)
		}
	}()

	// 重置viper
	viper.Reset()

	// 调用setDefaults
	setDefaults()

	t.Run("服务器默认值", func(t *testing.T) {
		if viper.GetString("server.host") != "0.0.0.0" {
			t.Errorf("Expected default server host '0.0.0.0', got '%s'", viper.GetString("server.host"))
		}
		if viper.GetInt("server.port") != 8082 {
			t.Errorf("Expected default server port 8082, got %d", viper.GetInt("server.port"))
		}
		if viper.GetString("server.mode") != "debug" {
			t.Errorf("Expected default server mode 'debug', got '%s'", viper.GetString("server.mode"))
		}
	})

	t.Run("日志默认值", func(t *testing.T) {
		if viper.GetString("logging.level") != "info" {
			t.Errorf("Expected default log level 'info', got '%s'", viper.GetString("logging.level"))
		}
		if viper.GetString("logging.format") != "json" {
			t.Errorf("Expected default log format 'json', got '%s'", viper.GetString("logging.format"))
		}
		if viper.GetString("logging.output") != "stdout" {
			t.Errorf("Expected default log output 'stdout', got '%s'", viper.GetString("logging.output"))
		}
	})

	t.Run("指标默认值", func(t *testing.T) {
		if !viper.GetBool("metrics.enabled") {
			t.Error("Expected metrics to be enabled by default")
		}
		if viper.GetString("metrics.path") != "/metrics" {
			t.Errorf("Expected default metrics path '/metrics', got '%s'", viper.GetString("metrics.path"))
		}
		if viper.GetDuration("metrics.interval") != 15*time.Second {
			t.Errorf("Expected default metrics interval 15s, got %v", viper.GetDuration("metrics.interval"))
		}
	})

	t.Run("批量处理器默认值", func(t *testing.T) {
		if !viper.GetBool("batch_processor.enabled") {
			t.Error("Expected batch processor to be enabled by default")
		}
		if viper.GetInt("batch_processor.batch_size") != 100 {
			t.Errorf("Expected default batch size 100, got %d", viper.GetInt("batch_processor.batch_size"))
		}
		if viper.GetDuration("batch_processor.flush_interval") != 5*time.Second {
			t.Errorf("Expected default flush interval 5s, got %v", viper.GetDuration("batch_processor.flush_interval"))
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("默认配置路径", func(t *testing.T) {
		// 确保环境变量未设置
		os.Unsetenv("LAOJUN_MONITORING_CONFIG_PATH")

		path := GetConfigPath()
		expected := "./configs/config.yaml"
		if path != expected {
			t.Errorf("Expected default config path '%s', got '%s'", expected, path)
		}
	})

	t.Run("环境变量配置路径", func(t *testing.T) {
		customPath := "/custom/path/config.yaml"
		os.Setenv("LAOJUN_MONITORING_CONFIG_PATH", customPath)
		defer os.Unsetenv("LAOJUN_MONITORING_CONFIG_PATH")

		path := GetConfigPath()
		if path != customPath {
			t.Errorf("Expected custom config path '%s', got '%s'", customPath, path)
		}
	})
}

func TestBatchProcessorConfig(t *testing.T) {
	t.Run("IsEnabled方法", func(t *testing.T) {
		config := &BatchProcessorConfig{Enabled: true}
		if !config.IsEnabled() {
			t.Error("Expected IsEnabled to return true")
		}

		config.Enabled = false
		if config.IsEnabled() {
			t.Error("Expected IsEnabled to return false")
		}
	})
}

func TestBatchExporterConfig(t *testing.T) {
	t.Run("IsEnabled方法", func(t *testing.T) {
		config := &BatchExporterConfig{Enabled: true}
		if !config.IsEnabled() {
			t.Error("Expected IsEnabled to return true")
		}

		config.Enabled = false
		if config.IsEnabled() {
			t.Error("Expected IsEnabled to return false")
		}
	})

	t.Run("GetName方法", func(t *testing.T) {
		name := "test-exporter"
		config := &BatchExporterConfig{Name: name}
		if config.GetName() != name {
			t.Errorf("Expected name '%s', got '%s'", name, config.GetName())
		}
	})

	t.Run("GetConfig方法", func(t *testing.T) {
		configMap := map[string]interface{}{
			"url":     "http://localhost:8080",
			"timeout": "30s",
		}
		config := &BatchExporterConfig{Config: configMap}
		
		result := config.GetConfig()
		if result["url"] != "http://localhost:8080" {
			t.Errorf("Expected url 'http://localhost:8080', got '%v'", result["url"])
		}
		if result["timeout"] != "30s" {
			t.Errorf("Expected timeout '30s', got '%v'", result["timeout"])
		}
	})

	t.Run("SetConfig方法", func(t *testing.T) {
		config := &BatchExporterConfig{
			Config: make(map[string]interface{}),
		}

		config.SetConfig("key1", "value1")
		config.SetConfig("key2", 42)

		if config.Config["key1"] != "value1" {
			t.Errorf("Expected key1 'value1', got '%v'", config.Config["key1"])
		}
		if config.Config["key2"] != 42 {
			t.Errorf("Expected key2 42, got '%v'", config.Config["key2"])
		}
	})

	t.Run("SetConfig空配置", func(t *testing.T) {
		config := &BatchExporterConfig{}

		config.SetConfig("key", "value")

		if config.Config == nil {
			t.Error("Expected Config to be initialized")
		}
		if config.Config["key"] != "value" {
			t.Errorf("Expected key 'value', got '%v'", config.Config["key"])
		}
	})
}

func TestConfigStructures(t *testing.T) {
	t.Run("完整配置结构", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: 8080,
				Mode: "debug",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
				Compress:   true,
			},
			Metrics: MetricsConfig{
				Enabled:       true,
				Path:          "/metrics",
				Interval:      15 * time.Second,
				Retention:     24 * time.Hour,
				EnableBuiltIn: true,
				EnableCustom:  true,
				MaxSeries:     100000,
				MaxSamples:    1000000,
			},
		}

		// 验证结构体字段
		if config.Server.Host != "localhost" {
			t.Errorf("Expected server host 'localhost', got '%s'", config.Server.Host)
		}
		if config.Logging.Level != "info" {
			t.Errorf("Expected log level 'info', got '%s'", config.Logging.Level)
		}
		if !config.Metrics.Enabled {
			t.Error("Expected metrics to be enabled")
		}
	})

	t.Run("告警配置结构", func(t *testing.T) {
		alertRule := AlertRule{
			Name:     "high_cpu",
			Query:    "cpu_usage > 80",
			Duration: 5 * time.Minute,
			Severity: "critical",
			Labels: map[string]string{
				"team": "infrastructure",
			},
			Annotations: map[string]string{
				"description": "High CPU usage detected",
			},
		}

		receiver := Receiver{
			Name: "slack",
			Type: "slack",
			Config: map[string]string{
				"webhook_url": "https://hooks.slack.com/...",
			},
		}

		route := Route{
			Match: map[string]string{
				"severity": "critical",
			},
			Receiver: "slack",
		}

		alertingConfig := AlertingConfig{
			Enabled:            true,
			EvaluationInterval: 30 * time.Second,
			Rules:              []AlertRule{alertRule},
			Receivers:          map[string]Receiver{"slack": receiver},
			Routes:             []Route{route},
		}

		// 验证告警配置
		if !alertingConfig.Enabled {
			t.Error("Expected alerting to be enabled")
		}
		if len(alertingConfig.Rules) != 1 {
			t.Errorf("Expected 1 alert rule, got %d", len(alertingConfig.Rules))
		}
		if alertingConfig.Rules[0].Name != "high_cpu" {
			t.Errorf("Expected rule name 'high_cpu', got '%s'", alertingConfig.Rules[0].Name)
		}
	})
}