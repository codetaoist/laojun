package tracing

import (
	"context"
	"fmt"
	"time"
)

// Tracer defines the interface for distributed tracing operations.
// All implementations must be thread-safe and compatible with OpenTelemetry.
type Tracer interface {
	// Span operations
	StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (Span, context.Context)
	SpanFromContext(ctx context.Context) Span
	ContextWithSpan(ctx context.Context, span Span) context.Context
	
	// Trace operations
	StartTrace(ctx context.Context, traceName string, opts ...TraceOption) (Trace, context.Context)
	GetActiveTrace(ctx context.Context) Trace
	
	// Injection and extraction for distributed tracing
	Inject(ctx context.Context, format Format, carrier interface{}) error
	Extract(ctx context.Context, format Format, carrier interface{}) (context.Context, error)
	
	// Sampling
	ShouldSample(ctx context.Context, traceID string, spanName string) bool
	
	// Export and flush
	Export(ctx context.Context, format ExportFormat) ([]byte, error)
	Flush(ctx context.Context) error
	
	// Health check
	IsHealthy(ctx context.Context) bool
	
	// Close releases resources
	Close() error
}

// Span represents a single span in a trace
type Span interface {
	// Span identification
	TraceID() string
	SpanID() string
	ParentSpanID() string
	
	// Span lifecycle
	Finish()
	FinishWithOptions(opts FinishOptions)
	IsFinished() bool
	
	// Span data
	SetOperationName(operationName string) Span
	SetTag(key string, value interface{}) Span
	SetBaggageItem(key, value string) Span
	GetBaggageItem(key string) string
	
	// Events and logs
	LogFields(fields ...LogField) Span
	LogKV(alternatingKeyValues ...interface{}) Span
	AddEvent(name string, attrs ...EventAttribute) Span
	
	// Status and errors
	SetStatus(status SpanStatus, description string) Span
	RecordError(err error, opts ...ErrorOption) Span
	
	// Context
	Context() SpanContext
	
	// Timing
	StartTime() time.Time
	Duration() time.Duration
}

// Trace represents a complete trace
type Trace interface {
	TraceID() string
	RootSpan() Span
	GetSpans() []Span
	IsFinished() bool
	Duration() time.Duration
}

// SpanContext represents the span context for propagation
type SpanContext interface {
	TraceID() string
	SpanID() string
	IsSampled() bool
	TraceFlags() byte
	TraceState() TraceState
}

// TraceState represents the trace state for propagation
type TraceState interface {
	Get(key string) string
	Set(key, value string) TraceState
	Delete(key string) TraceState
	Len() int
	String() string
}

// SpanOption configures span creation
type SpanOption interface {
	Apply(*SpanConfig)
}

// TraceOption configures trace creation
type TraceOption interface {
	Apply(*TraceConfig)
}

// SpanConfig holds span configuration
type SpanConfig struct {
	Tags       map[string]interface{}
	StartTime  time.Time
	References []SpanReference
	Parent     Span
}

// TraceConfig holds trace configuration
type TraceConfig struct {
	Tags      map[string]interface{}
	StartTime time.Time
	Sampler   Sampler
}

// SpanReference represents a reference between spans
type SpanReference struct {
	Type         ReferenceType
	ReferencedContext SpanContext
}

// Sampler determines if a trace should be sampled
type Sampler interface {
	ShouldSample(ctx context.Context, traceID string, spanName string, parentContext SpanContext) SamplingResult
}

// SamplingResult represents the result of sampling decision
type SamplingResult struct {
	Decision   SamplingDecision
	Attributes map[string]interface{}
	TraceState TraceState
}

// LogField represents a log field
type LogField struct {
	Key   string
	Value interface{}
}

// EventAttribute represents an event attribute
type EventAttribute struct {
	Key   string
	Value interface{}
}

// FinishOptions configures span finishing
type FinishOptions struct {
	FinishTime time.Time
}

// ErrorOption configures error recording
type ErrorOption interface {
	Apply(*ErrorConfig)
}

// ErrorConfig holds error recording configuration
type ErrorConfig struct {
	Timestamp  time.Time
	Attributes map[string]interface{}
}

// Enums and constants

// ReferenceType represents the type of span reference
type ReferenceType int

const (
	ChildOfRef     ReferenceType = iota
	FollowsFromRef
)

// SpanStatus represents the status of a span
type SpanStatus int

const (
	SpanStatusUnset SpanStatus = iota
	SpanStatusOK
	SpanStatusError
)

// SamplingDecision represents the sampling decision
type SamplingDecision int

const (
	SamplingDecisionDrop SamplingDecision = iota
	SamplingDecisionRecordOnly
	SamplingDecisionRecordAndSample
)

// Format represents the format for injection/extraction
type Format int

const (
	FormatTextMap Format = iota
	FormatHTTPHeaders
	FormatBinary
)

// ExportFormat represents the export format
type ExportFormat int

const (
	ExportFormatJaeger ExportFormat = iota
	ExportFormatZipkin
	ExportFormatOTLP
	ExportFormatJSON
)

// Config defines the configuration for tracing.
type Config struct {
	// Basic settings
	Enabled bool   `json:"enabled" yaml:"enabled" env:"TRACING_ENABLED"`
	Debug   bool   `json:"debug" yaml:"debug" env:"TRACING_DEBUG"`
	Timeout string `json:"timeout" yaml:"timeout" env:"TRACING_TIMEOUT"`
	
	// Service information
	ServiceName    string `json:"service_name" yaml:"service_name" env:"TRACING_SERVICE_NAME"`
	ServiceVersion string `json:"service_version" yaml:"service_version" env:"TRACING_SERVICE_VERSION"`
	Environment    string `json:"environment" yaml:"environment" env:"TRACING_ENVIRONMENT"`
	
	// Sampling configuration
	SamplingEnabled    bool    `json:"sampling_enabled" yaml:"sampling_enabled" env:"TRACING_SAMPLING_ENABLED"`
	SamplingRate       float64 `json:"sampling_rate" yaml:"sampling_rate" env:"TRACING_SAMPLING_RATE"`
	SamplingType       string  `json:"sampling_type" yaml:"sampling_type" env:"TRACING_SAMPLING_TYPE"`
	MaxTracesPerSecond int     `json:"max_traces_per_second" yaml:"max_traces_per_second" env:"TRACING_MAX_TRACES_PER_SECOND"`
	
	// Export configuration
	ExportEnabled  bool   `json:"export_enabled" yaml:"export_enabled" env:"TRACING_EXPORT_ENABLED"`
	ExportFormat   string `json:"export_format" yaml:"export_format" env:"TRACING_EXPORT_FORMAT"`
	ExportEndpoint string `json:"export_endpoint" yaml:"export_endpoint" env:"TRACING_EXPORT_ENDPOINT"`
	ExportTimeout  string `json:"export_timeout" yaml:"export_timeout" env:"TRACING_EXPORT_TIMEOUT"`
	ExportBatchSize int   `json:"export_batch_size" yaml:"export_batch_size" env:"TRACING_EXPORT_BATCH_SIZE"`
	ExportInterval  string `json:"export_interval" yaml:"export_interval" env:"TRACING_EXPORT_INTERVAL"`
	
	// Jaeger specific configuration
	JaegerAgentHost     string `json:"jaeger_agent_host" yaml:"jaeger_agent_host" env:"TRACING_JAEGER_AGENT_HOST"`
	JaegerAgentPort     int    `json:"jaeger_agent_port" yaml:"jaeger_agent_port" env:"TRACING_JAEGER_AGENT_PORT"`
	JaegerCollectorURL  string `json:"jaeger_collector_url" yaml:"jaeger_collector_url" env:"TRACING_JAEGER_COLLECTOR_URL"`
	JaegerUser          string `json:"jaeger_user" yaml:"jaeger_user" env:"TRACING_JAEGER_USER"`
	JaegerPassword      string `json:"jaeger_password" yaml:"jaeger_password" env:"TRACING_JAEGER_PASSWORD"`
	
	// Zipkin specific configuration
	ZipkinEndpoint string `json:"zipkin_endpoint" yaml:"zipkin_endpoint" env:"TRACING_ZIPKIN_ENDPOINT"`
	
	// OTLP specific configuration
	OTLPEndpoint    string            `json:"otlp_endpoint" yaml:"otlp_endpoint" env:"TRACING_OTLP_ENDPOINT"`
	OTLPHeaders     map[string]string `json:"otlp_headers" yaml:"otlp_headers"`
	OTLPCompression string            `json:"otlp_compression" yaml:"otlp_compression" env:"TRACING_OTLP_COMPRESSION"`
	OTLPInsecure    bool              `json:"otlp_insecure" yaml:"otlp_insecure" env:"TRACING_OTLP_INSECURE"`
	
	// Buffer and performance settings
	BufferSize      int    `json:"buffer_size" yaml:"buffer_size" env:"TRACING_BUFFER_SIZE"`
	MaxSpanCount    int    `json:"max_span_count" yaml:"max_span_count" env:"TRACING_MAX_SPAN_COUNT"`
	FlushInterval   string `json:"flush_interval" yaml:"flush_interval" env:"TRACING_FLUSH_INTERVAL"`
	MaxQueueSize    int    `json:"max_queue_size" yaml:"max_queue_size" env:"TRACING_MAX_QUEUE_SIZE"`
	
	// Resource attributes
	ResourceAttributes map[string]string `json:"resource_attributes" yaml:"resource_attributes"`
	
	// Propagation settings
	PropagationFormats []string `json:"propagation_formats" yaml:"propagation_formats"`
	
	// Instrumentation settings
	InstrumentationEnabled bool     `json:"instrumentation_enabled" yaml:"instrumentation_enabled" env:"TRACING_INSTRUMENTATION_ENABLED"`
	InstrumentationLibs    []string `json:"instrumentation_libs" yaml:"instrumentation_libs"`
	
	// Security settings
	TLSEnabled  bool   `json:"tls_enabled" yaml:"tls_enabled" env:"TRACING_TLS_ENABLED"`
	TLSCertFile string `json:"tls_cert_file" yaml:"tls_cert_file" env:"TRACING_TLS_CERT_FILE"`
	TLSKeyFile  string `json:"tls_key_file" yaml:"tls_key_file" env:"TRACING_TLS_KEY_FILE"`
	TLSCAFile   string `json:"tls_ca_file" yaml:"tls_ca_file" env:"TRACING_TLS_CA_FILE"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate timeout
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			return fmt.Errorf("invalid timeout format: %v", err)
		}
	}
	
	// Validate service name
	if c.Enabled && c.ServiceName == "" {
		return fmt.Errorf("service_name is required when tracing is enabled")
	}
	
	// Validate sampling configuration
	if c.SamplingEnabled {
		if c.SamplingRate < 0 || c.SamplingRate > 1 {
			return fmt.Errorf("sampling_rate must be between 0 and 1")
		}
		
		if c.SamplingType != "" && c.SamplingType != "probabilistic" && c.SamplingType != "rate_limiting" && c.SamplingType != "adaptive" {
			return fmt.Errorf("invalid sampling_type: %s", c.SamplingType)
		}
		
		if c.MaxTracesPerSecond < 0 {
			return fmt.Errorf("max_traces_per_second must be non-negative")
		}
	}
	
	// Validate export configuration
	if c.ExportEnabled {
		if c.ExportFormat != "" {
			validFormats := []string{"jaeger", "zipkin", "otlp", "json"}
			valid := false
			for _, format := range validFormats {
				if c.ExportFormat == format {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid export_format: %s", c.ExportFormat)
			}
		}
		
		if c.ExportTimeout != "" {
			if _, err := time.ParseDuration(c.ExportTimeout); err != nil {
				return fmt.Errorf("invalid export_timeout format: %v", err)
			}
		}
		
		if c.ExportInterval != "" {
			if _, err := time.ParseDuration(c.ExportInterval); err != nil {
				return fmt.Errorf("invalid export_interval format: %v", err)
			}
		}
		
		if c.ExportBatchSize < 0 {
			return fmt.Errorf("export_batch_size must be non-negative")
		}
	}
	
	// Validate Jaeger configuration
	if c.ExportFormat == "jaeger" {
		if c.JaegerAgentPort < 0 || c.JaegerAgentPort > 65535 {
			return fmt.Errorf("jaeger_agent_port must be between 0 and 65535")
		}
	}
	
	// Validate OTLP configuration
	if c.ExportFormat == "otlp" {
		if c.OTLPCompression != "" && c.OTLPCompression != "gzip" && c.OTLPCompression != "none" {
			return fmt.Errorf("invalid otlp_compression: %s", c.OTLPCompression)
		}
	}
	
	// Validate buffer and performance settings
	if c.BufferSize < 0 {
		return fmt.Errorf("buffer_size must be non-negative")
	}
	
	if c.MaxSpanCount < 0 {
		return fmt.Errorf("max_span_count must be non-negative")
	}
	
	if c.MaxQueueSize < 0 {
		return fmt.Errorf("max_queue_size must be non-negative")
	}
	
	if c.FlushInterval != "" {
		if _, err := time.ParseDuration(c.FlushInterval); err != nil {
			return fmt.Errorf("invalid flush_interval format: %v", err)
		}
	}
	
	// Validate propagation formats
	for _, format := range c.PropagationFormats {
		if format != "tracecontext" && format != "baggage" && format != "b3" && format != "jaeger" && format != "ottrace" {
			return fmt.Errorf("invalid propagation format: %s", format)
		}
	}
	
	return nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		// Basic settings
		Enabled: true,
		Debug:   false,
		Timeout: "30s",
		
		// Service information
		ServiceName:    "unknown-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		
		// Sampling configuration
		SamplingEnabled:    true,
		SamplingRate:       0.1, // 10% sampling rate
		SamplingType:       "probabilistic",
		MaxTracesPerSecond: 100,
		
		// Export configuration
		ExportEnabled:   true,
		ExportFormat:    "jaeger",
		ExportEndpoint:  "http://localhost:14268/api/traces",
		ExportTimeout:   "10s",
		ExportBatchSize: 100,
		ExportInterval:  "5s",
		
		// Jaeger specific configuration
		JaegerAgentHost:    "localhost",
		JaegerAgentPort:    6831,
		JaegerCollectorURL: "http://localhost:14268/api/traces",
		
		// Zipkin specific configuration
		ZipkinEndpoint: "http://localhost:9411/api/v2/spans",
		
		// OTLP specific configuration
		OTLPEndpoint:    "http://localhost:4318/v1/traces",
		OTLPHeaders:     make(map[string]string),
		OTLPCompression: "gzip",
		OTLPInsecure:    true,
		
		// Buffer and performance settings
		BufferSize:    1000,
		MaxSpanCount:  1000,
		FlushInterval: "5s",
		MaxQueueSize:  2048,
		
		// Resource attributes
		ResourceAttributes: map[string]string{
			"service.name":    "unknown-service",
			"service.version": "1.0.0",
		},
		
		// Propagation settings
		PropagationFormats: []string{"tracecontext", "baggage"},
		
		// Instrumentation settings
		InstrumentationEnabled: true,
		InstrumentationLibs:    []string{"http", "grpc", "database"},
		
		// Security settings
		TLSEnabled:  false,
		TLSCertFile: "",
		TLSKeyFile:  "",
		TLSCAFile:   "",
	}
}
