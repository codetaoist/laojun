package tracing

import "errors"

// Define tracing specific errors
var (
	// Tracer errors
	ErrTracerNotInitialized = errors.New("tracer not initialized")
	ErrTracerAlreadyClosed  = errors.New("tracer already closed")
	ErrInvalidConfig        = errors.New("invalid tracer configuration")
	
	// Span errors
	ErrSpanNotFound      = errors.New("span not found")
	ErrSpanAlreadyFinished = errors.New("span already finished")
	ErrInvalidSpanContext  = errors.New("invalid span context")
	ErrSpanNotActive       = errors.New("no active span in context")
	
	// Trace errors
	ErrTraceNotFound      = errors.New("trace not found")
	ErrTraceAlreadyFinished = errors.New("trace already finished")
	ErrInvalidTraceID     = errors.New("invalid trace ID")
	
	// Export errors
	ErrExportFailed       = errors.New("failed to export traces")
	ErrExportTimeout      = errors.New("export operation timeout")
	ErrInvalidExportFormat = errors.New("invalid export format")
	ErrExporterNotConfigured = errors.New("exporter not configured")
	
	// Sampling errors
	ErrInvalidSamplingRate = errors.New("invalid sampling rate")
	ErrSamplerNotConfigured = errors.New("sampler not configured")
	
	// Propagation errors
	ErrInvalidCarrier      = errors.New("invalid carrier for propagation")
	ErrPropagationFailed   = errors.New("failed to propagate trace context")
	ErrUnsupportedFormat   = errors.New("unsupported propagation format")
	
	// Buffer and queue errors
	ErrBufferFull         = errors.New("trace buffer is full")
	ErrQueueFull          = errors.New("trace queue is full")
	ErrFlushFailed        = errors.New("failed to flush traces")
	
	// Connection errors
	ErrConnectionFailed   = errors.New("failed to connect to trace backend")
	ErrConnectionTimeout  = errors.New("connection to trace backend timeout")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// IsTracerNotInitialized checks if the error is a tracer not initialized error.
func IsTracerNotInitialized(err error) bool {
	return errors.Is(err, ErrTracerNotInitialized)
}

// IsSpanNotFound checks if the error is a span not found error.
func IsSpanNotFound(err error) bool {
	return errors.Is(err, ErrSpanNotFound)
}

// IsSpanAlreadyFinished checks if the error is a span already finished error.
func IsSpanAlreadyFinished(err error) bool {
	return errors.Is(err, ErrSpanAlreadyFinished)
}

// IsTraceNotFound checks if the error is a trace not found error.
func IsTraceNotFound(err error) bool {
	return errors.Is(err, ErrTraceNotFound)
}

// IsExportFailed checks if the error is an export failed error.
func IsExportFailed(err error) bool {
	return errors.Is(err, ErrExportFailed)
}

// IsExportTimeout checks if the error is an export timeout error.
func IsExportTimeout(err error) bool {
	return errors.Is(err, ErrExportTimeout)
}

// IsPropagationFailed checks if the error is a propagation failed error.
func IsPropagationFailed(err error) bool {
	return errors.Is(err, ErrPropagationFailed)
}

// IsBufferFull checks if the error is a buffer full error.
func IsBufferFull(err error) bool {
	return errors.Is(err, ErrBufferFull)
}

// IsConnectionFailed checks if the error is a connection failed error.
func IsConnectionFailed(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}
