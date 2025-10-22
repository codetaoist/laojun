package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PluginServer 插件服务端
type PluginServer struct {
	config     *PluginConfig
	httpServer *http.Server
	grpcClient *grpc.ClientConn
	logger     *log.Logger
	stats      *PluginStats
	startTime  time.Time
}

// PluginConfig 插件配置
type PluginConfig struct {
	PluginID    string `json:"plugin_id"`
	InstanceID  string `json:"instance_id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Port        int    `json:"port"`
	GRPCServer  string `json:"grpc_server"`
	LogLevel    string `json:"log_level"`
	Environment string `json:"environment"`
	HealthCheck bool   `json:"health_check"`
	MetricsPath string `json:"metrics_path"`
}

// PluginStats 插件统计信息
type PluginStats struct {
	RequestCount    int64     `json:"request_count"`
	ErrorCount      int64     `json:"error_count"`
	LastRequestTime time.Time `json:"last_request_time"`
	TotalLatency    float64   `json:"total_latency"`
	AverageLatency  float64   `json:"average_latency"`
	Uptime          string    `json:"uptime"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// RequestData 请求数据
type RequestData struct {
	ID        string                 `json:"id"`
	Method    string                 `json:"method"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// ResponseData 响应数据
type ResponseData struct {
	ID        string                 `json:"id"`
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Latency   float64                `json:"latency"`
}

func main() {
	// 加载配置
	config := loadConfig()

	// 创建插件服务端
	server := NewPluginServer(config)

	// 启动服务端
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start plugin server: %v", err)
	}

	// 等待信号
	server.WaitForShutdown()
}

// NewPluginServer 创建新的插件服务端
func NewPluginServer(config *PluginConfig) *PluginServer {
	return &PluginServer{
		config:    config,
		logger:    log.New(os.Stdout, fmt.Sprintf("[%s] ", config.Name), log.LstdFlags),
		stats:     &PluginStats{},
		startTime: time.Now(),
	}
}

// Start 启动插件服务端
func (s *PluginServer) Start() error {
	s.logger.Printf("Starting plugin server: %s (%s)", s.config.Name, s.config.Version)

	// 连接到gRPC服务端
	if err := s.connectToGRPC(); err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// 注册插件
	if err := s.registerPlugin(); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	// 启动HTTP服务端
	if err := s.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	s.logger.Printf("Plugin server started on port %d", s.config.Port)
	return nil
}

// connectToGRPC 连接到gRPC服务端
func (s *PluginServer) connectToGRPC() error {
	conn, err := grpc.Dial(s.config.GRPCServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	s.grpcClient = conn
	s.logger.Printf("Connected to gRPC server: %s", s.config.GRPCServer)
	return nil
}

// registerPlugin 注册插件
func (s *PluginServer) registerPlugin() error {
	// 这里应该调用gRPC的RegisterPlugin方法
	// 为了演示，我们只是记录日志
	s.logger.Printf("Registering plugin with gRPC server...")

	// 模拟注册过程
	time.Sleep(100 * time.Millisecond)

	s.logger.Printf("Plugin registered successfully")
	return nil
}

// startHTTPServer 启动HTTP服务端
func (s *PluginServer) startHTTPServer() error {
	router := mux.NewRouter()

	// 注册路由
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/stats", s.handleStats).Methods("GET")
	router.HandleFunc("/process", s.handleProcess).Methods("POST")
	router.HandleFunc("/config", s.handleConfig).Methods("GET", "PUT")
	router.HandleFunc("/metadata", s.handleMetadata).Methods("GET")

	// 添加中间件
	router.Use(s.loggingMiddleware)
	router.Use(s.metricsMiddleware)

	// 创建HTTP服务端
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务端
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// handleHealth 健康检查处理器
func (s *PluginServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime)

	health := HealthResponse{
		Status:    "healthy",
		Message:   "Plugin is running normally",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"plugin_id":     s.config.PluginID,
			"instance_id":   s.config.InstanceID,
			"name":          s.config.Name,
			"version":       s.config.Version,
			"uptime":        uptime.String(),
			"request_count": s.stats.RequestCount,
			"error_count":   s.stats.ErrorCount,
			"environment":   s.config.Environment,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleStats 统计信息处理器
func (s *PluginServer) handleStats(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime)
	s.stats.Uptime = uptime.String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.stats)
}

// handleProcess 请求处理处理器
func (s *PluginServer) handleProcess(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var requestData RequestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		s.sendErrorResponse(w, "Invalid request format", err)
		return
	}

	// 处理请求
	responseData := s.processRequest(&requestData)
	responseData.Latency = time.Since(startTime).Seconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// handleConfig 配置处理处理器
func (s *PluginServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.config)
	case "PUT":
		// 更新配置的逻辑
		s.sendSuccessResponse(w, "Configuration updated", nil)
	}
}

// handleMetadata 元数据处理器
func (s *PluginServer) handleMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := map[string]interface{}{
		"id":          s.config.PluginID,
		"name":        s.config.Name,
		"version":     s.config.Version,
		"type":        "microservice",
		"runtime":     "docker",
		"description": "Example microservice plugin",
		"author":      "Taishanglaojun Team",
		"capabilities": []string{
			"request_processing",
			"health_check",
			"metrics",
			"configuration",
		},
		"endpoints": map[string]string{
			"health":   "/health",
			"stats":    "/stats",
			"process":  "/process",
			"config":   "/config",
			"metadata": "/metadata",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// processRequest 处理业务请求
func (s *PluginServer) processRequest(request *RequestData) *ResponseData {
	response := &ResponseData{
		ID:        request.ID,
		Success:   true,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	// 根据方法处理不同的请求
	switch request.Method {
	case "hello":
		name := "World"
		if request.Data["name"] != nil {
			name = request.Data["name"].(string)
		}
		response.Data["message"] = fmt.Sprintf("Hello, %s! From %s", name, s.config.Name)
		response.Message = "Greeting processed successfully"

	case "calculate":
		if operation, ok := request.Data["operation"].(string); ok {
			a, aOk := request.Data["a"].(float64)
			b, bOk := request.Data["b"].(float64)

			if aOk && bOk {
				var result float64
				var err error

				switch operation {
				case "add":
					result = a + b
				case "subtract":
					result = a - b
				case "multiply":
					result = a * b
				case "divide":
					if b != 0 {
						result = a / b
					} else {
						response.Success = false
						response.Message = "Division by zero"
						return response
					}
				default:
					response.Success = false
					response.Message = "Unknown operation"
					return response
				}

				if err == nil {
					response.Data["result"] = result
					response.Data["operation"] = operation
					response.Data["operands"] = map[string]float64{"a": a, "b": b}
					response.Message = "Calculation completed successfully"
				}
			} else {
				response.Success = false
				response.Message = "Invalid operands"
			}
		} else {
			response.Success = false
			response.Message = "Missing operation"
		}

	case "echo":
		response.Data = request.Data
		response.Message = "Echo processed successfully"

	case "time":
		response.Data["server_time"] = time.Now().Format(time.RFC3339)
		response.Data["timezone"] = time.Now().Location().String()
		response.Data["unix_timestamp"] = time.Now().Unix()
		response.Message = "Time information retrieved"

	case "stats":
		response.Data["plugin_stats"] = s.stats
		response.Data["uptime"] = time.Since(s.startTime).String()
		response.Message = "Statistics retrieved"

	default:
		response.Success = false
		response.Message = fmt.Sprintf("Unknown method: %s", request.Method)
	}

	return response
}

// sendErrorResponse 发送错误响应处理器
func (s *PluginServer) sendErrorResponse(w http.ResponseWriter, message string, err error) {
	s.stats.ErrorCount++

	response := map[string]interface{}{
		"success":   false,
		"message":   message,
		"timestamp": time.Now(),
	}

	if err != nil {
		response["error"] = err.Error()
		s.logger.Printf("Error: %s - %v", message, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// sendSuccessResponse 发送成功响应处理器
func (s *PluginServer) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := map[string]interface{}{
		"success":   true,
		"message":   message,
		"timestamp": time.Now(),
	}

	if data != nil {
		response["data"] = data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loggingMiddleware 日志中间件
func (s *PluginServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		s.logger.Printf("%s %s %v", r.Method, r.URL.Path, duration)
	})
}

// metricsMiddleware 指标中间件
func (s *PluginServer) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.stats.RequestCount++
		s.stats.LastRequestTime = start

		next.ServeHTTP(w, r)

		latency := time.Since(start).Seconds()
		s.stats.TotalLatency += latency
		s.stats.AverageLatency = s.stats.TotalLatency / float64(s.stats.RequestCount)
	})
}

// WaitForShutdown 等待关闭信号
func (s *PluginServer) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Println("Shutting down plugin server...")

	// 注销插件
	s.unregisterPlugin()

	// 关闭HTTP服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Printf("HTTP server shutdown error: %v", err)
	}

	// 关闭gRPC连接
	if s.grpcClient != nil {
		s.grpcClient.Close()
	}

	s.logger.Println("Plugin server stopped")
}

// unregisterPlugin 注销插件
func (s *PluginServer) unregisterPlugin() {
	// 这里应该调用gRPC的UnregisterPlugin方法
	s.logger.Printf("Unregistering plugin from gRPC server...")

	// 模拟注销过程
	time.Sleep(100 * time.Millisecond)

	s.logger.Printf("Plugin unregistered successfully")
}

// loadConfig 加载配置
func loadConfig() *PluginConfig {
	config := &PluginConfig{
		PluginID:    getEnv("PLUGIN_ID", "example-microservice-plugin"),
		InstanceID:  getEnv("INSTANCE_ID", "instance-1"),
		Name:        getEnv("PLUGIN_NAME", "Example Microservice Plugin"),
		Version:     getEnv("PLUGIN_VERSION", "1.0.0"),
		Port:        getEnvInt("PLUGIN_PORT", 8080),
		GRPCServer:  getEnv("GRPC_SERVER", "localhost:9090"),
		LogLevel:    getEnv("PLUGIN_LOG_LEVEL", "info"),
		Environment: getEnv("PLUGIN_ENV", "development"),
		HealthCheck: getEnvBool("HEALTH_CHECK", true),
		MetricsPath: getEnv("METRICS_PATH", "/metrics"),
	}

	return config
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
