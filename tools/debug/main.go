package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

// DebugConfig represents the debug configuration
type DebugConfig struct {
	Port          int    `json:"port"`
	Host          string `json:"host"`
	LogLevel      string `json:"log_level"`
	EnablePprof   bool   `json:"enable_pprof"`
	EnableMetrics bool   `json:"enable_metrics"`
	EnableTracing bool   `json:"enable_tracing"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// MetricEntry represents a metric entry
type MetricEntry struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// TraceEntry represents a trace entry
type TraceEntry struct {
	TraceID   string                 `json:"trace_id"`
	SpanID    string                 `json:"span_id"`
	Operation string                 `json:"operation"`
	Duration  time.Duration          `json:"duration"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

var (
	config    DebugConfig
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	logBuffer = make([]LogEntry, 0, 1000)
	clients   = make(map[*websocket.Conn]bool)
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "laojun-debug",
		Short: "Laojun Debug Tool",
		Long: `A comprehensive debugging tool for the Laojun platform.
This tool provides real-time monitoring, logging, profiling, and debugging capabilities.`,
	}

	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(clientCmd())
	rootCmd.AddCommand(profileCmd())
	rootCmd.AddCommand(traceCmd())
	rootCmd.AddCommand(metricsCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func serverCmd() *cobra.Command {
	var port int
	var host string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start debug server",
		Long:  "Start the debug server with web interface",
		Run: func(cmd *cobra.Command, args []string) {
			config = DebugConfig{
				Port:          port,
				Host:          host,
				LogLevel:      "debug",
				EnablePprof:   true,
				EnableMetrics: true,
				EnableTracing: true,
			}
			startDebugServer()
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 9090, "Debug server port")
	cmd.Flags().StringVarP(&host, "host", "h", "localhost", "Debug server host")

	return cmd
}

func startDebugServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Static files
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	// Routes
	setupRoutes(router)

	// Start server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	fmt.Printf("🚀 Debug server starting on http://%s\n", addr)
	fmt.Println("📊 Features enabled:")
	fmt.Printf("   - Web Interface: http://%s\n", addr)
	fmt.Printf("   - WebSocket Logs: ws://%s/ws/logs\n", addr)
	fmt.Printf("   - Metrics: http://%s/metrics\n", addr)
	if config.EnablePprof {
		fmt.Printf("   - Profiling: http://%s/debug/pprof/\n", addr)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down debug server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Debug server stopped")
}

func setupRoutes(router *gin.Engine) {
	// Main dashboard
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Laojun Debug Dashboard",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/status", getStatus)
		api.GET("/logs", getLogs)
		api.POST("/logs", addLog)
		api.GET("/metrics", getMetrics)
		api.POST("/metrics", addMetric)
		api.GET("/traces", getTraces)
		api.POST("/traces", addTrace)
		api.GET("/config", getConfig)
		api.PUT("/config", updateConfig)
	}

	// WebSocket endpoints
	router.GET("/ws/logs", handleLogWebSocket)
	router.GET("/ws/metrics", handleMetricsWebSocket)

	// Profiling endpoints (if enabled)
	if config.EnablePprof {
		setupPprofRoutes(router)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now(),
		"config":    config,
		"stats": gin.H{
			"logs_count":    len(logBuffer),
			"clients_count": len(clients),
		},
	})
}

func getLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	start := 0
	if len(logBuffer) > limit {
		start = len(logBuffer) - limit
	}

	logs := logBuffer[start:]
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logBuffer),
	})
}

func addLog(c *gin.Context) {
	var entry LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.Timestamp = time.Now()
	logBuffer = append(logBuffer, entry)

	// Keep buffer size manageable
	if len(logBuffer) > 1000 {
		logBuffer = logBuffer[100:]
	}

	// Broadcast to WebSocket clients
	broadcastLog(entry)

	c.JSON(http.StatusOK, gin.H{"status": "logged"})
}

func getMetrics(c *gin.Context) {
	// Mock metrics data
	metrics := []MetricEntry{
		{
			Name:      "http_requests_total",
			Type:      "counter",
			Value:     1234,
			Labels:    map[string]string{"method": "GET", "status": "200"},
			Timestamp: time.Now(),
		},
		{
			Name:      "http_request_duration_seconds",
			Type:      "histogram",
			Value:     0.123,
			Labels:    map[string]string{"method": "POST"},
			Timestamp: time.Now(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

func addMetric(c *gin.Context) {
	var metric MetricEntry
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric.Timestamp = time.Now()
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

func getTraces(c *gin.Context) {
	// Mock trace data
	traces := []TraceEntry{
		{
			TraceID:   "trace-123",
			SpanID:    "span-456",
			Operation: "http.request",
			Duration:  time.Millisecond * 150,
			Tags:      map[string]interface{}{"http.method": "GET", "http.url": "/api/users"},
			Timestamp: time.Now(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"traces": traces,
	})
}

func addTrace(c *gin.Context) {
	var trace TraceEntry
	if err := c.ShouldBindJSON(&trace); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trace.Timestamp = time.Now()
	c.JSON(http.StatusOK, gin.H{"status": "traced"})
}

func getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, config)
}

func updateConfig(c *gin.Context) {
	var newConfig DebugConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config = newConfig
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func handleLogWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	defer delete(clients, conn)

	// Send recent logs
	for _, entry := range logBuffer {
		if err := conn.WriteJSON(entry); err != nil {
			break
		}
	}

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func handleMetricsWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Send metrics updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := []MetricEntry{
				{
					Name:      "cpu_usage",
					Type:      "gauge",
					Value:     float64(time.Now().Unix() % 100),
					Timestamp: time.Now(),
				},
			}
			if err := conn.WriteJSON(metrics); err != nil {
				return
			}
		}
	}
}

func broadcastLog(entry LogEntry) {
	for client := range clients {
		if err := client.WriteJSON(entry); err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}

func setupPprofRoutes(router *gin.Engine) {
	pprof := router.Group("/debug/pprof")
	{
		pprof.GET("/", gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
		})))
		pprof.GET("/cmdline", gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, strings.Join(os.Args, " "))
		})))
		pprof.GET("/profile", gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			fmt.Fprintf(w, "CPU profiling not implemented in this demo")
		})))
		pprof.GET("/symbol", gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Symbol lookup not implemented in this demo")
		})))
		pprof.GET("/trace", gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			fmt.Fprintf(w, "Trace not implemented in this demo")
		})))
	}
}

func clientCmd() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Debug client",
		Long:  "Connect to a debug server and send test data",
		Run: func(cmd *cobra.Command, args []string) {
			startDebugClient(target)
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "http://localhost:9090", "Debug server URL")

	return cmd
}

func startDebugClient(target string) {
	fmt.Printf("🔗 Connecting to debug server: %s\n", target)

	// Interactive client
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Debug client started. Commands:")
	fmt.Println("  log <level> <message> - Send a log entry")
	fmt.Println("  metric <name> <value> - Send a metric")
	fmt.Println("  trace <operation> <duration> - Send a trace")
	fmt.Println("  status - Get server status")
	fmt.Println("  quit - Exit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "log":
			if len(parts) >= 3 {
				level := parts[1]
				message := strings.Join(parts[2:], " ")
				sendLog(target, level, message)
			} else {
				fmt.Println("Usage: log <level> <message>")
			}

		case "metric":
			if len(parts) >= 3 {
				name := parts[1]
				if value, err := strconv.ParseFloat(parts[2], 64); err == nil {
					sendMetric(target, name, value)
				} else {
					fmt.Printf("Invalid value: %s\n", parts[2])
				}
			} else {
				fmt.Println("Usage: metric <name> <value>")
			}

		case "trace":
			if len(parts) >= 3 {
				operation := parts[1]
				if duration, err := time.ParseDuration(parts[2]); err == nil {
					sendTrace(target, operation, duration)
				} else {
					fmt.Printf("Invalid duration: %s\n", parts[2])
				}
			} else {
				fmt.Println("Usage: trace <operation> <duration>")
			}

		case "status":
			getServerStatus(target)

		case "quit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s\n", parts[0])
		}
	}
}

func sendLog(target, level, message string) {
	entry := LogEntry{
		Level:   level,
		Message: message,
		Source:  "debug-client",
	}

	data, _ := json.Marshal(entry)
	resp, err := http.Post(target+"/api/v1/logs", "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Printf("Error sending log: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Log sent: %s\n", message)
}

func sendMetric(target, name string, value float64) {
	metric := MetricEntry{
		Name:  name,
		Type:  "gauge",
		Value: value,
	}

	data, _ := json.Marshal(metric)
	resp, err := http.Post(target+"/api/v1/metrics", "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Printf("Error sending metric: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Metric sent: %s = %f\n", name, value)
}

func sendTrace(target, operation string, duration time.Duration) {
	trace := TraceEntry{
		TraceID:   fmt.Sprintf("trace-%d", time.Now().Unix()),
		SpanID:    fmt.Sprintf("span-%d", time.Now().UnixNano()),
		Operation: operation,
		Duration:  duration,
	}

	data, _ := json.Marshal(trace)
	resp, err := http.Post(target+"/api/v1/traces", "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Printf("Error sending trace: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Trace sent: %s (%v)\n", operation, duration)
}

func getServerStatus(target string) {
	resp, err := http.Get(target + "/api/v1/status")
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var status map[string]interface{}
	if err := json.Unmarshal(body, &status); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	fmt.Printf("📊 Server Status:\n")
	data, _ := json.MarshalIndent(status, "", "  ")
	fmt.Println(string(data))
}

func profileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "profile",
		Short: "Profile analysis",
		Long:  "Analyze performance profiles",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 Profile analysis not implemented yet")
		},
	}
}

func traceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trace",
		Short: "Trace analysis",
		Long:  "Analyze distributed traces",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 Trace analysis not implemented yet")
		},
	}
}

func metricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "Metrics analysis",
		Long:  "Analyze system metrics",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📊 Metrics analysis not implemented yet")
		},
	}
}
