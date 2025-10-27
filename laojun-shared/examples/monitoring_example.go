package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/monitoring"
)

func main() {
	// Create configuration
	config := monitoring.DefaultConfig()
	config.Debug = true
	config.ServiceName = "example-service"
	config.ServiceVersion = "1.0.0"
	config.DefaultLabels = map[string]string{
		"environment": "development",
		"region":      "us-west-2",
	}
	
	// Create monitoring instance
	monitor, err := monitoring.NewMonitor(config)
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}
	defer monitor.Close()
	
	ctx := context.Background()
	
	fmt.Println("=== Monitoring Example ===")
	
	// Example 1: Counter metrics
	fmt.Println("\n1. Counter Metrics:")
	labels := map[string]string{"endpoint": "/api/users", "method": "GET"}
	
	err = monitor.IncrementCounter(ctx, "http_requests_total", labels)
	if err != nil {
		log.Printf("Error incrementing counter: %v", err)
	}
	
	err = monitor.AddCounter(ctx, "http_requests_total", 5, labels)
	if err != nil {
		log.Printf("Error adding to counter: %v", err)
	}
	
	fmt.Println("✅ Counter metrics recorded")
	
	// Example 2: Gauge metrics
	fmt.Println("\n2. Gauge Metrics:")
	err = monitor.SetGauge(ctx, "active_connections", 42, map[string]string{"server": "web-01"})
	if err != nil {
		log.Printf("Error setting gauge: %v", err)
	}
	
	err = monitor.AddGauge(ctx, "memory_usage_bytes", 1024*1024*100, map[string]string{"component": "cache"})
	if err != nil {
		log.Printf("Error adding to gauge: %v", err)
	}
	
	fmt.Println("✅ Gauge metrics recorded")
	
	// Example 3: Histogram metrics
	fmt.Println("\n3. Histogram Metrics:")
	for i := 0; i < 5; i++ {
		responseTime := float64(100 + i*50) // 100ms, 150ms, 200ms, etc.
		err = monitor.RecordHistogram(ctx, "http_request_duration_seconds", responseTime/1000, labels)
		if err != nil {
			log.Printf("Error recording histogram: %v", err)
		}
	}
	
	fmt.Println("✅ Histogram metrics recorded")
	
	// Example 4: Timer usage
	fmt.Println("\n4. Timer Usage:")
	timer := monitor.StartTimer(ctx, "database_query", map[string]string{"query": "SELECT_USERS"})
	
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	
	duration := timer.Stop()
	fmt.Printf("✅ Database query took: %v\n", duration)
	
	// Record the duration as a metric
	err = monitor.RecordDuration(ctx, "database_query", duration, map[string]string{"query": "SELECT_USERS"})
	if err != nil {
		log.Printf("Error recording duration: %v", err)
	}
	
	// Example 5: Summary metrics
	fmt.Println("\n5. Summary Metrics:")
	for i := 0; i < 3; i++ {
		value := float64(50 + i*25) // 50, 75, 100
		err = monitor.RecordSummary(ctx, "request_size_bytes", value, map[string]string{"content_type": "application/json"})
		if err != nil {
			log.Printf("Error recording summary: %v", err)
		}
	}
	
	fmt.Println("✅ Summary metrics recorded")
	
	// Example 6: Get all metrics
	fmt.Println("\n6. Retrieving Metrics:")
	metrics, err := monitor.GetMetrics(ctx)
	if err != nil {
		log.Printf("Error getting metrics: %v", err)
	} else {
		fmt.Printf("Total metrics collected: %d\n", len(metrics))
		
		// Show first few metrics
		for i, metric := range metrics {
			if i >= 3 {
				fmt.Println("... (showing first 3 metrics)")
				break
			}
			fmt.Printf("  - %s: %v (type: %d, labels: %v)\n", 
				metric.Name(), metric.Value(), metric.Type(), metric.Labels())
		}
	}
	
	// Example 7: Export metrics in different formats
	fmt.Println("\n7. Exporting Metrics:")
	
	// Export as JSON
	jsonData, err := monitor.Export(ctx, monitoring.ExportFormatJSON)
	if err != nil {
		log.Printf("Error exporting JSON: %v", err)
	} else {
		fmt.Printf("JSON export size: %d bytes\n", len(jsonData))
		if len(jsonData) > 100 {
			fmt.Printf("JSON preview: %s...\n", string(jsonData[:100]))
		}
	}
	
	// Export as Prometheus format
	promData, err := monitor.Export(ctx, monitoring.ExportFormatPrometheus)
	if err != nil {
		log.Printf("Error exporting Prometheus: %v", err)
	} else {
		fmt.Printf("Prometheus export size: %d bytes\n", len(promData))
		if len(promData) > 200 {
			fmt.Printf("Prometheus preview:\n%s...\n", string(promData[:200]))
		}
	}
	
	// Example 8: Health check
	fmt.Println("\n8. Health Check:")
	if monitor.IsHealthy(ctx) {
		fmt.Println("✅ Monitor is healthy")
	} else {
		fmt.Println("❌ Monitor is not healthy")
	}
	
	// Example 9: Custom metric registration
	fmt.Println("\n9. Custom Metric Registration:")
	customMetric := &customMetricImpl{
		name:      "custom_business_metric",
		value:     123.45,
		labels:    map[string]string{"business_unit": "sales"},
		timestamp: time.Now(),
	}
	
	err = monitor.RegisterMetric(ctx, customMetric)
	if err != nil {
		log.Printf("Error registering custom metric: %v", err)
	} else {
		fmt.Println("✅ Custom metric registered")
	}
	
	// Final metrics count
	finalMetrics, _ := monitor.GetMetrics(ctx)
	fmt.Printf("\nFinal metrics count: %d\n", len(finalMetrics))
	
	fmt.Println("\n=== Monitoring Example Completed Successfully! ===")
}

// customMetricImpl is an example implementation of the Metric interface
type customMetricImpl struct {
	name      string
	value     interface{}
	labels    map[string]string
	timestamp time.Time
}

func (c *customMetricImpl) Name() string                 { return c.name }
func (c *customMetricImpl) Type() monitoring.MetricType { return monitoring.MetricTypeGauge }
func (c *customMetricImpl) Value() interface{}          { return c.value }
func (c *customMetricImpl) Labels() map[string]string   { return c.labels }
func (c *customMetricImpl) Timestamp() time.Time        { return c.timestamp }
