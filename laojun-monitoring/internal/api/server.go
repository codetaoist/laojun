package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/codetaoist/laojun-monitoring/internal/alerting"
	"github.com/codetaoist/laojun-monitoring/internal/collectors"
	"github.com/codetaoist/laojun-monitoring/internal/logging"
	"github.com/codetaoist/laojun-monitoring/internal/metrics"
	"github.com/codetaoist/laojun-monitoring/internal/storage"
)

// Server HTTP服务器
type Server struct {
	config     ServerConfig
	logger     *zap.Logger
	router     *mux.Router
	server     *http.Server
	handler    *Handler
	
	// 组件
	metricRegistry   *metrics.MetricRegistry
	storageManager   *storage.StorageManager
	alertManager     *alerting.AlertManager
	collectorManager *collectors.CollectorManager
	pipelineManager  *logging.PipelineManager
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string            `mapstructure:"host"`
	Port         int               `mapstructure:"port"`
	ReadTimeout  time.Duration     `mapstructure:"read_timeout"`
	WriteTimeout time.Duration     `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration     `mapstructure:"idle_timeout"`
	TLS          TLSConfig         `mapstructure:"tls"`
	Middleware   MiddlewareConfig  `mapstructure:"middleware"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// NewServer 创建新的服务器
func NewServer(
	config ServerConfig,
	logger *zap.Logger,
	metricRegistry *metrics.MetricRegistry,
	storageManager *storage.StorageManager,
	alertManager *alerting.AlertManager,
	collectorManager *collectors.CollectorManager,
	pipelineManager *logging.PipelineManager,
) *Server {
	// 设置默认值
	if config.Host == "" {
		config.Host = "0.0.0.0"
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 30 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 30 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 60 * time.Second
	}
	
	// 创建路由器
	router := mux.NewRouter()
	
	// 创建处理器
	handler := NewHandler(
		logger,
		metricRegistry,
		storageManager,
		alertManager,
		collectorManager,
		pipelineManager,
	)
	
	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}
	
	return &Server{
		config:           config,
		logger:           logger,
		router:           router,
		server:           server,
		handler:          handler,
		metricRegistry:   metricRegistry,
		storageManager:   storageManager,
		alertManager:     alertManager,
		collectorManager: collectorManager,
		pipelineManager:  pipelineManager,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", 
		zap.String("address", s.server.Addr),
		zap.Bool("tls_enabled", s.config.TLS.Enabled))
	
	// 设置中间件
	SetupMiddlewares(s.router, s.logger, s.config.Middleware)
	
	// 注册路由
	s.setupRoutes()
	
	// 启动服务器
	if s.config.TLS.Enabled {
		return s.server.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile)
	}
	
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	
	return s.server.Shutdown(ctx)
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 注册API路由
	s.handler.RegisterRoutes(s.router)
	
	// Prometheus指标端点
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	
	// 静态文件服务（如果需要）
	s.setupStaticRoutes()
	
	// 404处理
	s.router.NotFoundHandler = http.HandlerFunc(s.notFoundHandler)
	
	// 405处理
	s.router.MethodNotAllowedHandler = http.HandlerFunc(s.methodNotAllowedHandler)
}

// setupStaticRoutes 设置静态路由
func (s *Server) setupStaticRoutes() {
	// 可以在这里添加静态文件服务
	// 例如：Web UI、文档等
	
	// 示例：服务静态文件
	// staticDir := "/static/"
	// s.router.PathPrefix(staticDir).Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("./web/static/"))))
	
	// 示例：服务Web UI
	// s.router.HandleFunc("/", s.indexHandler).Methods("GET")
	// s.router.HandleFunc("/ui/{path:.*}", s.uiHandler).Methods("GET")
}

// notFoundHandler 404处理器
func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("404 Not Found", 
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	
	response := fmt.Sprintf(`{
		"error": "Not Found",
		"message": "The requested resource was not found",
		"path": "%s",
		"timestamp": %d
	}`, r.URL.Path, time.Now().Unix())
	
	w.Write([]byte(response))
}

// methodNotAllowedHandler 405处理器
func (s *Server) methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("405 Method Not Allowed", 
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	
	response := fmt.Sprintf(`{
		"error": "Method Not Allowed",
		"message": "The requested method is not allowed for this resource",
		"method": "%s",
		"path": "%s",
		"timestamp": %d
	}`, r.Method, r.URL.Path, time.Now().Unix())
	
	w.Write([]byte(response))
}

// GetRouter 获取路由器
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// GetServer 获取HTTP服务器
func (s *Server) GetServer() *http.Server {
	return s.server
}

// GetAddress 获取服务器地址
func (s *Server) GetAddress() string {
	return s.server.Addr
}

// IsRunning 检查服务器是否运行中
func (s *Server) IsRunning() bool {
	// 简单检查，实际应用中可能需要更复杂的逻辑
	return s.server != nil
}

// ServerManager 服务器管理器
type ServerManager struct {
	servers map[string]*Server
	logger  *zap.Logger
}

// NewServerManager 创建服务器管理器
func NewServerManager(logger *zap.Logger) *ServerManager {
	return &ServerManager{
		servers: make(map[string]*Server),
		logger:  logger,
	}
}

// AddServer 添加服务器
func (sm *ServerManager) AddServer(name string, server *Server) {
	sm.servers[name] = server
	sm.logger.Info("Server added", zap.String("name", name))
}

// RemoveServer 移除服务器
func (sm *ServerManager) RemoveServer(name string) error {
	server, exists := sm.servers[name]
	if !exists {
		return fmt.Errorf("server %s not found", name)
	}
	
	// 停止服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Stop(ctx); err != nil {
		sm.logger.Error("Failed to stop server", 
			zap.String("name", name), 
			zap.Error(err))
		return err
	}
	
	delete(sm.servers, name)
	sm.logger.Info("Server removed", zap.String("name", name))
	
	return nil
}

// GetServer 获取服务器
func (sm *ServerManager) GetServer(name string) (*Server, error) {
	server, exists := sm.servers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}
	
	return server, nil
}

// ListServers 列出所有服务器
func (sm *ServerManager) ListServers() map[string]*Server {
	result := make(map[string]*Server)
	for name, server := range sm.servers {
		result[name] = server
	}
	return result
}

// StartAll 启动所有服务器
func (sm *ServerManager) StartAll() error {
	for name, server := range sm.servers {
		go func(name string, server *Server) {
			if err := server.Start(); err != nil && err != http.ErrServerClosed {
				sm.logger.Error("Server failed", 
					zap.String("name", name), 
					zap.Error(err))
			}
		}(name, server)
		
		sm.logger.Info("Server started", zap.String("name", name))
	}
	
	return nil
}

// StopAll 停止所有服务器
func (sm *ServerManager) StopAll(ctx context.Context) error {
	for name, server := range sm.servers {
		if err := server.Stop(ctx); err != nil {
			sm.logger.Error("Failed to stop server", 
				zap.String("name", name), 
				zap.Error(err))
		} else {
			sm.logger.Info("Server stopped", zap.String("name", name))
		}
	}
	
	return nil
}

// HealthChecker 健康检查器
type HealthChecker struct {
	logger     *zap.Logger
	components map[string]HealthCheckFunc
}

// HealthCheckFunc 健康检查函数
type HealthCheckFunc func() error

// HealthStatus 健康状态
type HealthStatus struct {
	Status     string                 `json:"status"`
	Timestamp  int64                  `json:"timestamp"`
	Components map[string]ComponentStatus `json:"components"`
}

// ComponentStatus 组件状态
type ComponentStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		logger:     logger,
		components: make(map[string]HealthCheckFunc),
	}
}

// RegisterComponent 注册组件
func (hc *HealthChecker) RegisterComponent(name string, checkFunc HealthCheckFunc) {
	hc.components[name] = checkFunc
	hc.logger.Debug("Health check component registered", zap.String("name", name))
}

// Check 执行健康检查
func (hc *HealthChecker) Check() *HealthStatus {
	status := &HealthStatus{
		Status:     "healthy",
		Timestamp:  time.Now().Unix(),
		Components: make(map[string]ComponentStatus),
	}
	
	for name, checkFunc := range hc.components {
		if err := checkFunc(); err != nil {
			status.Components[name] = ComponentStatus{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			status.Status = "unhealthy"
		} else {
			status.Components[name] = ComponentStatus{
				Status: "healthy",
			}
		}
	}
	
	return status
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	logger         *zap.Logger
	metricRegistry *metrics.MetricRegistry
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(logger *zap.Logger, metricRegistry *metrics.MetricRegistry) *MetricsCollector {
	return &MetricsCollector{
		logger:         logger,
		metricRegistry: metricRegistry,
	}
}

// CollectHTTPMetrics 收集HTTP指标
func (mc *MetricsCollector) CollectHTTPMetrics(method, path string, statusCode int, duration time.Duration) {
	if mc.metricRegistry == nil {
		return
	}
	
	labels := map[string]string{
		"method": method,
		"path":   path,
		"status": fmt.Sprintf("%d", statusCode),
	}
	
	// 更新请求计数
	if err := mc.metricRegistry.UpdateMetricValue("http_requests_total", 1, labels); err != nil {
		mc.logger.Debug("Failed to update HTTP request count", zap.Error(err))
	}
	
	// 更新请求持续时间
	if err := mc.metricRegistry.UpdateMetricValue("http_request_duration_seconds", duration.Seconds(), labels); err != nil {
		mc.logger.Debug("Failed to update HTTP request duration", zap.Error(err))
	}
}

// StartMetricsCollection 启动指标收集
func (mc *MetricsCollector) StartMetricsCollection() {
	// 这里可以启动定期的指标收集任务
	go mc.collectSystemMetrics()
}

// collectSystemMetrics 收集系统指标
func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 这里可以收集系统级别的指标
		// 例如：内存使用、CPU使用、goroutine数量等
		mc.collectRuntimeMetrics()
	}
}

// collectRuntimeMetrics 收集运行时指标
func (mc *MetricsCollector) collectRuntimeMetrics() {
	if mc.metricRegistry == nil {
		return
	}
	
	// 这里可以添加运行时指标收集逻辑
	// 例如：
	// - runtime.NumGoroutine()
	// - runtime.ReadMemStats()
	// - 等等
}