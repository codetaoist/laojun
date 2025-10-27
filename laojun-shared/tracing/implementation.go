package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TracerImpl implements the Tracer interface
type TracerImpl struct {
	config    *Config
	spans     map[string]*spanImpl
	traces    map[string]*traceImpl
	sampler   Sampler
	exporter  Exporter
	mu        sync.RWMutex
	closed    bool
	buffer    []*spanImpl
	bufferMu  sync.Mutex
}

// spanImpl implements the Span interface
type spanImpl struct {
	traceID       string
	spanID        string
	parentSpanID  string
	operationName string
	startTime     time.Time
	endTime       *time.Time
	tags          map[string]interface{}
	logs          []LogEntry
	events        []Event
	status        SpanStatus
	statusDesc    string
	baggage       map[string]string
	finished      bool
	mu            sync.RWMutex
}

// traceImpl implements the Trace interface
type traceImpl struct {
	traceID   string
	rootSpan  *spanImpl
	spans     []*spanImpl
	startTime time.Time
	endTime   *time.Time
	finished  bool
	mu        sync.RWMutex
}

// spanContextImpl implements the SpanContext interface
type spanContextImpl struct {
	traceID    string
	spanID     string
	sampled    bool
	traceFlags byte
	traceState TraceState
}

// traceStateImpl implements the TraceState interface
type traceStateImpl struct {
	entries map[string]string
	mu      sync.RWMutex
}

// LogEntry represents a log entry in a span
type LogEntry struct {
	Timestamp time.Time
	Fields    []LogField
}

// Event represents an event in a span
type Event struct {
	Name       string
	Timestamp  time.Time
	Attributes []EventAttribute
}

// Exporter interface for exporting traces
type Exporter interface {
	Export(ctx context.Context, spans []*spanImpl) error
	Shutdown(ctx context.Context) error
}

// NewTracer creates a new tracer instance
func NewTracer(config *Config) (Tracer, error) {
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	tracer := &TracerImpl{
		config: config,
		spans:  make(map[string]*spanImpl),
		traces: make(map[string]*traceImpl),
		buffer: make([]*spanImpl, 0, config.BufferSize),
	}
	
	// Initialize sampler
	if config.SamplingEnabled {
		tracer.sampler = NewProbabilisticSampler(config.SamplingRate)
	}
	
	// Initialize exporter
	if config.ExportEnabled {
		exporter, err := NewExporter(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter: %w", err)
		}
		tracer.exporter = exporter
	}
	
	return tracer, nil
}

// StartSpan starts a new span
func (t *TracerImpl) StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (Span, context.Context) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.closed {
		return nil, ctx
	}
	
	config := &SpanConfig{
		Tags:      make(map[string]interface{}),
		StartTime: time.Now(),
	}
	
	for _, opt := range opts {
		opt.Apply(config)
	}
	
	span := &spanImpl{
		spanID:        generateID(),
		operationName: operationName,
		startTime:     config.StartTime,
		tags:          config.Tags,
		logs:          make([]LogEntry, 0),
		events:        make([]Event, 0),
		baggage:       make(map[string]string),
		status:        SpanStatusUnset,
	}
	
	// Set trace ID and parent
	if config.Parent != nil {
		span.traceID = config.Parent.TraceID()
		span.parentSpanID = config.Parent.SpanID()
	} else if parentSpan := t.SpanFromContext(ctx); parentSpan != nil {
		span.traceID = parentSpan.TraceID()
		span.parentSpanID = parentSpan.SpanID()
	} else {
		span.traceID = generateID()
	}
	
	// Check sampling
	if t.sampler != nil && !t.ShouldSample(ctx, span.traceID, operationName) {
		return span, ctx
	}
	
	t.spans[span.spanID] = span
	
	// Add to trace
	if trace, exists := t.traces[span.traceID]; exists {
		trace.mu.Lock()
		trace.spans = append(trace.spans, span)
		trace.mu.Unlock()
	} else {
		newTrace := &traceImpl{
			traceID:   span.traceID,
			rootSpan:  span,
			spans:     []*spanImpl{span},
			startTime: span.startTime,
		}
		t.traces[span.traceID] = newTrace
	}
	
	return span, t.ContextWithSpan(ctx, span)
}

// SpanFromContext extracts span from context
func (t *TracerImpl) SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value("span").(Span); ok {
		return span
	}
	return nil
}

// ContextWithSpan adds span to context
func (t *TracerImpl) ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, "span", span)
}

// StartTrace starts a new trace
func (t *TracerImpl) StartTrace(ctx context.Context, traceName string, opts ...TraceOption) (Trace, context.Context) {
	_, newCtx := t.StartSpan(ctx, traceName)
	if trace := t.GetActiveTrace(newCtx); trace != nil {
		return trace, newCtx
	}
	return nil, ctx
}

// GetActiveTrace gets the active trace from context
func (t *TracerImpl) GetActiveTrace(ctx context.Context) Trace {
	if span := t.SpanFromContext(ctx); span != nil {
		t.mu.RLock()
		defer t.mu.RUnlock()
		if trace, exists := t.traces[span.TraceID()]; exists {
			return trace
		}
	}
	return nil
}

// Inject injects trace context into carrier
func (t *TracerImpl) Inject(ctx context.Context, format Format, carrier interface{}) error {
	span := t.SpanFromContext(ctx)
	if span == nil {
		return ErrSpanNotActive
	}
	
	switch format {
	case FormatHTTPHeaders:
		if headers, ok := carrier.(http.Header); ok {
			headers.Set("traceparent", fmt.Sprintf("00-%s-%s-01", span.TraceID(), span.SpanID()))
			return nil
		}
		return ErrInvalidCarrier
	case FormatTextMap:
		if textMap, ok := carrier.(map[string]string); ok {
			textMap["traceparent"] = fmt.Sprintf("00-%s-%s-01", span.TraceID(), span.SpanID())
			return nil
		}
		return ErrInvalidCarrier
	default:
		return ErrUnsupportedFormat
	}
}

// Extract extracts trace context from carrier
func (t *TracerImpl) Extract(ctx context.Context, format Format, carrier interface{}) (context.Context, error) {
	var traceparent string
	
	switch format {
	case FormatHTTPHeaders:
		if headers, ok := carrier.(http.Header); ok {
			traceparent = headers.Get("traceparent")
		} else {
			return ctx, ErrInvalidCarrier
		}
	case FormatTextMap:
		if textMap, ok := carrier.(map[string]string); ok {
			traceparent = textMap["traceparent"]
		} else {
			return ctx, ErrInvalidCarrier
		}
	default:
		return ctx, ErrUnsupportedFormat
	}
	
	if traceparent == "" {
		return ctx, nil
	}
	
	// Parse traceparent header (simplified)
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 {
		return ctx, ErrInvalidSpanContext
	}
	
	spanContext := &spanContextImpl{
		traceID:    parts[1],
		spanID:     parts[2],
		sampled:    parts[3] == "01",
		traceFlags: 1,
	}
	
	return context.WithValue(ctx, "spanContext", spanContext), nil
}

// ShouldSample determines if a trace should be sampled
func (t *TracerImpl) ShouldSample(ctx context.Context, traceID string, spanName string) bool {
	if t.sampler == nil {
		return true
	}
	
	result := t.sampler.ShouldSample(ctx, traceID, spanName, nil)
	return result.Decision != SamplingDecisionDrop
}

// Export exports traces in the specified format
func (t *TracerImpl) Export(ctx context.Context, format ExportFormat) ([]byte, error) {
	t.mu.RLock()
	spans := make([]*spanImpl, 0, len(t.spans))
	for _, span := range t.spans {
		if span.finished {
			spans = append(spans, span)
		}
	}
	t.mu.RUnlock()
	
	switch format {
	case ExportFormatJSON:
		return t.exportJSON(spans)
	case ExportFormatJaeger:
		return t.exportJaeger(spans)
	default:
		return nil, ErrInvalidExportFormat
	}
}

// Flush flushes all pending traces
func (t *TracerImpl) Flush(ctx context.Context) error {
	if t.exporter == nil {
		return nil
	}
	
	t.bufferMu.Lock()
	spans := make([]*spanImpl, len(t.buffer))
	copy(spans, t.buffer)
	t.buffer = t.buffer[:0]
	t.bufferMu.Unlock()
	
	if len(spans) > 0 {
		return t.exporter.Export(ctx, spans)
	}
	
	return nil
}

// IsHealthy checks if the tracer is healthy
func (t *TracerImpl) IsHealthy(ctx context.Context) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.closed
}

// Close closes the tracer
func (t *TracerImpl) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.closed {
		return ErrTracerAlreadyClosed
	}
	
	t.closed = true
	
	if t.exporter != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return t.exporter.Shutdown(ctx)
	}
	
	return nil
}

// Span implementation methods

func (s *spanImpl) TraceID() string {
	return s.traceID
}

func (s *spanImpl) SpanID() string {
	return s.spanID
}

func (s *spanImpl) ParentSpanID() string {
	return s.parentSpanID
}

func (s *spanImpl) Finish() {
	s.FinishWithOptions(FinishOptions{FinishTime: time.Now()})
}

func (s *spanImpl) FinishWithOptions(opts FinishOptions) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.finished {
		return
	}
	
	s.finished = true
	s.endTime = &opts.FinishTime
}

func (s *spanImpl) IsFinished() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.finished
}

func (s *spanImpl) SetOperationName(operationName string) Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operationName = operationName
	return s
}

func (s *spanImpl) SetTag(key string, value interface{}) Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tags[key] = value
	return s
}

func (s *spanImpl) SetBaggageItem(key, value string) Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.baggage[key] = value
	return s
}

func (s *spanImpl) GetBaggageItem(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baggage[key]
}

func (s *spanImpl) LogFields(fields ...LogField) Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Fields:    fields,
	}
	s.logs = append(s.logs, entry)
	return s
}

func (s *spanImpl) LogKV(alternatingKeyValues ...interface{}) Span {
	fields := make([]LogField, 0, len(alternatingKeyValues)/2)
	for i := 0; i < len(alternatingKeyValues)-1; i += 2 {
		key := fmt.Sprintf("%v", alternatingKeyValues[i])
		value := alternatingKeyValues[i+1]
		fields = append(fields, LogField{Key: key, Value: value})
	}
	return s.LogFields(fields...)
}

func (s *spanImpl) AddEvent(name string, attrs ...EventAttribute) Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	event := Event{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: attrs,
	}
	s.events = append(s.events, event)
	return s
}

func (s *spanImpl) SetStatus(status SpanStatus, description string) Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
	s.statusDesc = description
	return s
}

func (s *spanImpl) RecordError(err error, opts ...ErrorOption) Span {
	config := &ErrorConfig{
		Timestamp:  time.Now(),
		Attributes: make(map[string]interface{}),
	}
	
	for _, opt := range opts {
		opt.Apply(config)
	}
	
	s.SetStatus(SpanStatusError, err.Error())
	s.SetTag("error", true)
	s.SetTag("error.message", err.Error())
	
	return s
}

func (s *spanImpl) Context() SpanContext {
	return &spanContextImpl{
		traceID:    s.traceID,
		spanID:     s.spanID,
		sampled:    true,
		traceFlags: 1,
	}
}

func (s *spanImpl) StartTime() time.Time {
	return s.startTime
}

func (s *spanImpl) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.endTime != nil {
		return s.endTime.Sub(s.startTime)
	}
	return time.Since(s.startTime)
}

// Helper functions

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (t *TracerImpl) exportJSON(spans []*spanImpl) ([]byte, error) {
	data := make([]map[string]interface{}, len(spans))
	for i, span := range spans {
		data[i] = map[string]interface{}{
			"traceID":       span.traceID,
			"spanID":        span.spanID,
			"parentSpanID":  span.parentSpanID,
			"operationName": span.operationName,
			"startTime":     span.startTime,
			"endTime":       span.endTime,
			"tags":          span.tags,
			"logs":          span.logs,
			"events":        span.events,
			"status":        span.status,
			"duration":      span.Duration(),
		}
	}
	return json.Marshal(data)
}

func (t *TracerImpl) exportJaeger(spans []*spanImpl) ([]byte, error) {
	// Simplified Jaeger format
	jaegerSpans := make([]map[string]interface{}, len(spans))
	for i, span := range spans {
		jaegerSpans[i] = map[string]interface{}{
			"traceID":       span.traceID,
			"spanID":        span.spanID,
			"parentSpanID":  span.parentSpanID,
			"operationName": span.operationName,
			"startTime":     span.startTime.UnixMicro(),
			"duration":      span.Duration().Microseconds(),
			"tags":          span.tags,
			"logs":          span.logs,
		}
	}
	
	jaegerTrace := map[string]interface{}{
		"traceID": spans[0].traceID,
		"spans":   jaegerSpans,
	}
	
	return json.Marshal(jaegerTrace)
}

// Sampler implementations

type ProbabilisticSampler struct {
	rate float64
}

func NewProbabilisticSampler(rate float64) *ProbabilisticSampler {
	return &ProbabilisticSampler{rate: rate}
}

func (s *ProbabilisticSampler) ShouldSample(ctx context.Context, traceID string, spanName string, parentContext SpanContext) SamplingResult {
	// Simple hash-based sampling
	hash := 0
	for _, b := range traceID {
		hash = hash*31 + int(b)
	}
	
	decision := SamplingDecisionDrop
	if float64(hash%100)/100.0 < s.rate {
		decision = SamplingDecisionRecordAndSample
	}
	
	return SamplingResult{
		Decision:   decision,
		Attributes: make(map[string]interface{}),
	}
}

// Exporter implementations

type JSONExporter struct {
	endpoint string
}

func NewExporter(config *Config) (Exporter, error) {
	switch config.ExportFormat {
	case "json":
		return &JSONExporter{endpoint: config.ExportEndpoint}, nil
	default:
		return &JSONExporter{endpoint: config.ExportEndpoint}, nil
	}
}

func (e *JSONExporter) Export(ctx context.Context, spans []*spanImpl) error {
	// This is a simplified implementation
	// In a real implementation, you would send the spans to the configured endpoint
	return nil
}

func (e *JSONExporter) Shutdown(ctx context.Context) error {
	return nil
}

// SpanContext implementation

func (sc *spanContextImpl) TraceID() string {
	return sc.traceID
}

func (sc *spanContextImpl) SpanID() string {
	return sc.spanID
}

func (sc *spanContextImpl) IsSampled() bool {
	return sc.sampled
}

func (sc *spanContextImpl) TraceFlags() byte {
	return sc.traceFlags
}

func (sc *spanContextImpl) TraceState() TraceState {
	return sc.traceState
}

// TraceState implementation

func (ts *traceStateImpl) Get(key string) string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.entries[key]
}

func (ts *traceStateImpl) Set(key, value string) TraceState {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.entries == nil {
		ts.entries = make(map[string]string)
	}
	ts.entries[key] = value
	return ts
}

func (ts *traceStateImpl) Delete(key string) TraceState {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.entries, key)
	return ts
}

func (ts *traceStateImpl) Len() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.entries)
}

func (ts *traceStateImpl) String() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	var parts []string
	for k, v := range ts.entries {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

// Trace implementation

func (t *traceImpl) TraceID() string {
	return t.traceID
}

func (t *traceImpl) RootSpan() Span {
	return t.rootSpan
}

func (t *traceImpl) GetSpans() []Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	spans := make([]Span, len(t.spans))
	for i, span := range t.spans {
		spans[i] = span
	}
	return spans
}

func (t *traceImpl) IsFinished() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.finished
}

func (t *traceImpl) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if t.endTime != nil {
		return t.endTime.Sub(t.startTime)
	}
	return time.Since(t.startTime)
}
