package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/codetaoist/laojun-shared/tracing"
)

func main() {
	// Create a new tracer instance with custom configuration
	config := tracing.DefaultConfig()
	config.Debug = true
	config.ServiceName = "tracing-example"
	config.ServiceVersion = "1.0.0"
	config.Environment = "development"
	config.SamplingRate = 1.0 // 100% sampling for demo
	config.ExportFormat = "json"
	
	tracer, err := tracing.NewTracer(&config)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}
	defer tracer.Close()
	
	ctx := context.Background()
	
	// Example 1: Basic span creation and finishing
	fmt.Println("=== Example 1: Basic Span Operations ===")
	span, _ := tracer.StartSpan(ctx, "example-operation")
	span.SetTag("component", "example")
	span.SetTag("version", "1.0.0")
	span.LogKV("event", "operation started", "timestamp", time.Now())
	
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	
	span.SetStatus(tracing.SpanStatusOK, "Operation completed successfully")
	span.Finish()
	
	fmt.Printf("Created span: %s in trace: %s\n", span.SpanID(), span.TraceID())
	
	// Example 2: Nested spans (parent-child relationship)
	fmt.Println("\n=== Example 2: Nested Spans ===")
	parentSpan, parentCtx := tracer.StartSpan(ctx, "parent-operation")
	parentSpan.SetTag("operation.type", "parent")
	
	// Create child span
	childSpan, childCtx := tracer.StartSpan(parentCtx, "child-operation")
	childSpan.SetTag("operation.type", "child")
	childSpan.SetTag("parent.span.id", parentSpan.SpanID())
	
	// Simulate child work
	time.Sleep(50 * time.Millisecond)
	childSpan.AddEvent("processing", tracing.EventAttribute{Key: "step", Value: "validation"})
	
	// Create grandchild span
	grandchildSpan, _ := tracer.StartSpan(childCtx, "grandchild-operation")
	grandchildSpan.SetTag("operation.type", "grandchild")
	grandchildSpan.LogFields(
		tracing.LogField{Key: "level", Value: "info"},
		tracing.LogField{Key: "message", Value: "Processing grandchild operation"},
	)
	
	time.Sleep(25 * time.Millisecond)
	grandchildSpan.Finish()
	
	childSpan.Finish()
	parentSpan.Finish()
	
	fmt.Printf("Parent span: %s\n", parentSpan.SpanID())
	fmt.Printf("Child span: %s (parent: %s)\n", childSpan.SpanID(), childSpan.ParentSpanID())
	fmt.Printf("Grandchild span: %s (parent: %s)\n", grandchildSpan.SpanID(), grandchildSpan.ParentSpanID())
	
	// Example 3: Error handling and recording
	fmt.Println("\n=== Example 3: Error Handling ===")
	errorSpan, _ := tracer.StartSpan(ctx, "error-operation")
	errorSpan.SetTag("operation.name", "simulate-error")
	
	// Simulate an error
	simulatedError := fmt.Errorf("simulated database connection error")
	errorSpan.RecordError(simulatedError)
	errorSpan.SetStatus(tracing.SpanStatusError, "Database connection failed")
	errorSpan.Finish()
	
	fmt.Printf("Error span: %s with status: %d\n", errorSpan.SpanID(), tracing.SpanStatusError)
	
	// Example 4: Baggage items (cross-cutting concerns)
	fmt.Println("\n=== Example 4: Baggage Items ===")
	baggageSpan, baggageCtx := tracer.StartSpan(ctx, "baggage-operation")
	baggageSpan.SetBaggageItem("user.id", "12345")
	baggageSpan.SetBaggageItem("request.id", "req-abc-123")
	baggageSpan.SetBaggageItem("correlation.id", "corr-xyz-789")
	
	// Create child span that can access baggage
	childWithBaggage, _ := tracer.StartSpan(baggageCtx, "child-with-baggage")
	userID := baggageSpan.GetBaggageItem("user.id")
	requestID := baggageSpan.GetBaggageItem("request.id")
	
	childWithBaggage.SetTag("inherited.user.id", userID)
	childWithBaggage.SetTag("inherited.request.id", requestID)
	childWithBaggage.Finish()
	
	baggageSpan.Finish()
	
	fmt.Printf("Baggage span with user.id: %s, request.id: %s\n", userID, requestID)
	
	// Example 5: Trace context injection and extraction
	fmt.Println("\n=== Example 5: Context Propagation ===")
	propagationSpan, propagationCtx := tracer.StartSpan(ctx, "propagation-operation")
	
	// Inject trace context into HTTP headers
	headers := make(http.Header)
	err = tracer.Inject(propagationCtx, tracing.FormatHTTPHeaders, headers)
	if err != nil {
		log.Printf("Failed to inject trace context: %v", err)
	} else {
		fmt.Printf("Injected traceparent: %s\n", headers.Get("traceparent"))
	}
	
	// Extract trace context from HTTP headers
	extractedCtx, err := tracer.Extract(ctx, tracing.FormatHTTPHeaders, headers)
	if err != nil {
		log.Printf("Failed to extract trace context: %v", err)
	} else {
		fmt.Println("Successfully extracted trace context")
	}
	
	// Create span with extracted context
	extractedSpan, _ := tracer.StartSpan(extractedCtx, "extracted-operation")
	extractedSpan.SetTag("extracted", true)
	extractedSpan.Finish()
	
	propagationSpan.Finish()
	
	// Example 6: Trace and span information
	fmt.Println("\n=== Example 6: Trace Information ===")
	infoSpan, infoCtx := tracer.StartSpan(ctx, "info-operation")
	
	// Get active trace
	activeTrace := tracer.GetActiveTrace(infoCtx)
	if activeTrace != nil {
		fmt.Printf("Active trace ID: %s\n", activeTrace.TraceID())
		fmt.Printf("Root span: %s\n", activeTrace.RootSpan().SpanID())
		fmt.Printf("Trace duration: %v\n", activeTrace.Duration())
	}
	
	infoSpan.Finish()
	
	// Example 7: Export traces
	fmt.Println("\n=== Example 7: Export Traces ===")
	
	// Export in JSON format
	jsonData, err := tracer.Export(ctx, tracing.ExportFormatJSON)
	if err != nil {
		log.Printf("Failed to export JSON: %v", err)
	} else {
		fmt.Printf("Exported JSON data length: %d bytes\n", len(jsonData))
		// Uncomment to see the actual JSON data
		// fmt.Printf("JSON data: %s\n", string(jsonData))
	}
	
	// Export in Jaeger format
	jaegerData, err := tracer.Export(ctx, tracing.ExportFormatJaeger)
	if err != nil {
		log.Printf("Failed to export Jaeger: %v", err)
	} else {
		fmt.Printf("Exported Jaeger data length: %d bytes\n", len(jaegerData))
	}
	
	// Example 8: Health check and flush
	fmt.Println("\n=== Example 8: Health Check and Flush ===")
	
	// Check tracer health
	if tracer.IsHealthy(ctx) {
		fmt.Println("Tracer is healthy")
	} else {
		fmt.Println("Tracer is not healthy")
	}
	
	// Flush pending traces
	err = tracer.Flush(ctx)
	if err != nil {
		log.Printf("Failed to flush traces: %v", err)
	} else {
		fmt.Println("Successfully flushed traces")
	}
	
	// Example 9: Sampling demonstration
	fmt.Println("\n=== Example 9: Sampling ===")
	
	// Create multiple spans to demonstrate sampling
	for i := 0; i < 10; i++ {
		testSpan, _ := tracer.StartSpan(ctx, fmt.Sprintf("test-span-%d", i))
		testSpan.SetTag("iteration", i)
		
		// Check if span should be sampled
		shouldSample := tracer.ShouldSample(ctx, testSpan.TraceID(), fmt.Sprintf("test-span-%d", i))
		testSpan.SetTag("sampled", shouldSample)
		
		testSpan.Finish()
		
		if shouldSample {
			fmt.Printf("Span %d: sampled\n", i)
		} else {
			fmt.Printf("Span %d: not sampled\n", i)
		}
	}
	
	fmt.Println("\n=== Tracing Example Completed Successfully! ===")
}
