package monitoring

import (
	"context"
	"fmt"
	"time"
)

// Monitor defines the interface for monitoring operations.
// All implementations must be thread-safe.
type Monitor interface {
	// Counter operations
	IncrementCounter(ctx context.Context, name string, labels map[string]string) error
	AddCounter(ctx context.Context, name string, value float64, labels map[string]string) error
	
	// Gauge operations
	SetGauge(ctx context.Context, name string, value float64, labels map[string]string) error
	AddGauge(ctx context.Context, name string, value float64, labels map[string]string) error
	
	// Histogram operations
	RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error
	
	// Summary operations
	RecordSummary(ctx context.Context, name string, value float64, labels map[string]string) error
	
	// Timer operations
	StartTimer(ctx context.Context, name string, labels map[string]string) Timer
	RecordDuration(ctx context.Context, name string, duration time.Duration, labels map[string]string) error
	
	// Registry operations
	RegisterMetric(ctx context.Context, metric Metric) error
	UnregisterMetric(ctx context.Context, name string) error
	GetMetrics(ctx context.Context) ([]Metric, error)
	
	// Export operations
	Export(ctx context.Context, format ExportFormat) ([]byte, error)
	
	// Health check
	IsHealthy(ctx context.Context) bool
	
	// Close releases resources
	Close() error
}

// Timer represents a timing measurement
type Timer interface {
	Stop() time.Duration
	Duration() time.Duration
}

// Metric represents a monitoring metric
type Metric interface {
	Name() string
	Type() MetricType
	Value() interface{}
	Labels() map[string]string
	Timestamp() time.Time
}

// MetricType represents the type of metric
type MetricType int

const (
	MetricTypeCounter MetricType = iota
	MetricTypeGauge
	MetricTypeHistogram
	MetricTypeSummary
)

// ExportFormat represents the export format
type ExportFormat int

const (
	ExportFormatPrometheus ExportFormat = iota
	ExportFormatJSON
	ExportFormatInfluxDB
)

// Config defines the configuration for monitoring.
type Config struct {
	Enabled bool          `yaml:"enabled" env:"MONITORING_ENABLED" default:"true"`
	Debug   bool          `yaml:"debug" env:"MONITORING_DEBUG" default:"false"`
	Timeout time.Duration `yaml:"timeout" env:"MONITORING_TIMEOUT" default:"30s"`
	
	// Metrics configuration
	MetricsEnabled    bool          `yaml:"metrics_enabled" env:"MONITORING_METRICS_ENABLED" default:"true"`
	MetricsPath       string        `yaml:"metrics_path" env:"MONITORING_METRICS_PATH" default:"/metrics"`
	MetricsPort       int           `yaml:"metrics_port" env:"MONITORING_METRICS_PORT" default:"9090"`
	MetricsInterval   time.Duration `yaml:"metrics_interval" env:"MONITORING_METRICS_INTERVAL" default:"15s"`
	
	// Export configuration
	ExportFormat      ExportFormat  `yaml:"export_format" env:"MONITORING_EXPORT_FORMAT" default:"0"`
	ExportEndpoint    string        `yaml:"export_endpoint" env:"MONITORING_EXPORT_ENDPOINT" default:""`
	ExportBatchSize   int           `yaml:"export_batch_size" env:"MONITORING_EXPORT_BATCH_SIZE" default:"100"`
	ExportTimeout     time.Duration `yaml:"export_timeout" env:"MONITORING_EXPORT_TIMEOUT" default:"10s"`
	
	// Storage configuration
	StorageEnabled    bool          `yaml:"storage_enabled" env:"MONITORING_STORAGE_ENABLED" default:"false"`
	StorageRetention  time.Duration `yaml:"storage_retention" env:"MONITORING_STORAGE_RETENTION" default:"24h"`
	StorageMaxSize    int64         `yaml:"storage_max_size" env:"MONITORING_STORAGE_MAX_SIZE" default:"1073741824"` // 1GB
	
	// Labels configuration
	DefaultLabels     map[string]string `yaml:"default_labels" env:"MONITORING_DEFAULT_LABELS"`
	ServiceName       string            `yaml:"service_name" env:"MONITORING_SERVICE_NAME" default:"unknown"`
	ServiceVersion    string            `yaml:"service_version" env:"MONITORING_SERVICE_VERSION" default:"unknown"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if c.MetricsEnabled {
		if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
			return fmt.Errorf("metrics port must be between 1 and 65535")
		}
		if c.MetricsInterval <= 0 {
			return fmt.Errorf("metrics interval must be positive")
		}
		if c.MetricsPath == "" {
			return fmt.Errorf("metrics path cannot be empty")
		}
	}
	
	if c.ExportBatchSize <= 0 {
		return fmt.Errorf("export batch size must be positive")
	}
	
	if c.ExportTimeout <= 0 {
		return fmt.Errorf("export timeout must be positive")
	}
	
	if c.StorageEnabled {
		if c.StorageRetention <= 0 {
			return fmt.Errorf("storage retention must be positive")
		}
		if c.StorageMaxSize <= 0 {
			return fmt.Errorf("storage max size must be positive")
		}
	}
	
	if c.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	
	return nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Debug:   false,
		Timeout: 30 * time.Second,
		
		// Metrics configuration
		MetricsEnabled:  true,
		MetricsPath:     "/metrics",
		MetricsPort:     9090,
		MetricsInterval: 15 * time.Second,
		
		// Export configuration
		ExportFormat:    ExportFormatPrometheus,
		ExportEndpoint:  "",
		ExportBatchSize: 100,
		ExportTimeout:   10 * time.Second,
		
		// Storage configuration
		StorageEnabled:   false,
		StorageRetention: 24 * time.Hour,
		StorageMaxSize:   1024 * 1024 * 1024, // 1GB
		
		// Labels configuration
		DefaultLabels:  make(map[string]string),
		ServiceName:    "unknown",
		ServiceVersion: "unknown",
	}
}
