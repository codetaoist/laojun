package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Gateway 统一路由网关
type Gateway struct {
	config           *GatewayConfig
	router           *mux.Router
	server           *http.Server
	pluginRegistry   *PluginRegistry
	loadBalancer     *LoadBalancer
	rateLimiter      *RateLimiter
	authManager      *AuthManager
	metricsCollector *MetricsCollector
	logger           *log.Logger
	upgrader         websocket.Upgrader
	middleware       []Middleware
	mu               sync.RWMutex
}

// GatewayConfig 网关配置
type GatewayConfig struct {
	Host               string        `json:"host"`
	Port               int           `json:"port"`
	ReadTimeout        time.Duration `json:"read_timeout"`
	WriteTimeout       time.Duration `json:"write_timeout"`
	IdleTimeout        time.Duration `json:"idle_timeout"`
	MaxHeaderBytes     int           `json:"max_header_bytes"`
	EnableCORS         bool          `json:"enable_cors"`
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableAuth         bool          `json:"enable_auth"`
	EnableRateLimit    bool          `json:"enable_rate_limit"`
	EnableLoadBalancer bool          `json:"enable_load_balancer"`
	EnableWebSocket    bool          `json:"enable_websocket"`
	TLSCertFile        string        `json:"tls_cert_file"`
	TLSKeyFile         string        `json:"tls_key_file"`
	LogLevel           string        `json:"log_level"`
	HealthCheckPath    string        `json:"health_check_path"`
	MetricsPath        string        `json:"metrics_path"`
	AdminPath          string        `json:"admin_path"`
}

// PluginRegistry 插件注册中心
type PluginRegistry struct {
	plugins map[string]*PluginEndpoint
	routes  map[string]*RouteConfig
	mu      sync.RWMutex
}

// PluginEndpoint 插件端点
type PluginEndpoint struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	BaseURL     string                 `json:"base_url"`
	HealthURL   string                 `json:"health_url"`
	Status      string                 `json:"status"`
	Instances   []*PluginInstance      `json:"instances"`
	Config      map[string]interface{} `json:"config"`
	Metadata    map[string]interface{} `json:"metadata"`
	LastUpdated time.Time              `json:"last_updated"`
}

// PluginInstance 插件实例
type PluginInstance struct {
	ID       string                 `json:"id"`
	Address  string                 `json:"address"`
	Port     int                    `json:"port"`
	Status   string                 `json:"status"`
	Health   *HealthStatus          `json:"health"`
	Metrics  *InstanceMetrics       `json:"metrics"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Path        string            `json:"path"`
	Methods     []string          `json:"methods"`
	PluginID    string            `json:"plugin_id"`
	Rewrite     string            `json:"rewrite"`
	StripPrefix bool              `json:"strip_prefix"`
	Headers     map[string]string `json:"headers"`
	Timeout     time.Duration     `json:"timeout"`
	Retry       *RetryConfig      `json:"retry"`
	Cache       *CacheConfig      `json:"cache"`
	Auth        *AuthConfig       `json:"auth"`
	RateLimit   *RateLimitConfig  `json:"rate_limit"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	Backoff     time.Duration `json:"backoff"`
	Timeout     time.Duration `json:"timeout"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled bool          `json:"enabled"`
	TTL     time.Duration `json:"ttl"`
	Key     string        `json:"key"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Required bool     `json:"required"`
	Roles    []string `json:"roles"`
	Scopes   []string `json:"scopes"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool          `json:"enabled"`
	Rate    int           `json:"rate"`
	Burst   int           `json:"burst"`
	Window  time.Duration `json:"window"`
}

// LoadBalancer 负载均衡中心
type LoadBalancer struct {
	strategy LoadBalanceStrategy
	mu       sync.RWMutex
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy interface {
	SelectInstance(instances []*PluginInstance) *PluginInstance
}

// RoundRobinStrategy 轮询策略
type RoundRobinStrategy struct {
	current int
	mu      sync.Mutex
}

// WeightedRoundRobinStrategy 加权轮询策略
type WeightedRoundRobinStrategy struct {
	weights map[string]int
	current map[string]int
	mu      sync.Mutex
}

// LeastConnectionsStrategy 最少连接策略
type LeastConnectionsStrategy struct{}

// RateLimiter 限流中心
type RateLimiter struct {
	limiters map[string]*TokenBucket
	mu       sync.RWMutex
}

// TokenBucket 令牌桶
type TokenBucket struct {
	capacity   int
	tokens     int
	rate       int
	lastRefill time.Time
	mu         sync.Mutex
}

// AuthManager 认证管理中心
type AuthManager struct {
	jwtSecret   string
	tokenCache  map[string]*TokenInfo
	userStore   UserStore
	roleManager *RoleManager
	mu          sync.RWMutex
}

// TokenInfo 令牌信息
type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserStore 用户存储接口
type UserStore interface {
	GetUser(userID string) (*User, error)
	ValidateCredentials(username, password string) (*User, error)
}

// User 用户信息
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	Active   bool     `json:"active"`
}

// RoleManager 角色管理中心
type RoleManager struct {
	roles       map[string]*Role
	permissions map[string]*Permission
	mu          sync.RWMutex
}

// Role 角色
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Permission 权限
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// MetricsCollector 指标收集中心
type MetricsCollector struct {
	requestCount   map[string]int64
	responseTime   map[string][]float64
	errorCount     map[string]int64
	activeRequests int64
	pluginMetrics  map[string]*PluginMetrics
	mu             sync.RWMutex
}

// PluginMetrics 插件指标
type PluginMetrics struct {
	RequestCount   int64     `json:"request_count"`
	ErrorCount     int64     `json:"error_count"`
	AverageLatency float64   `json:"average_latency"`
	LastRequest    time.Time `json:"last_request"`
}

// InstanceMetrics 实例指标
type InstanceMetrics struct {
	RequestCount    int64     `json:"request_count"`
	ErrorCount      int64     `json:"error_count"`
	AverageLatency  float64   `json:"average_latency"`
	ActiveRequests  int64     `json:"active_requests"`
	LastHealthCheck time.Time `json:"last_health_check"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details"`
	ResponseTime float64                `json:"response_time"`
}

// Middleware 中间件接口
type Middleware interface {
	Handle(next http.Handler) http.Handler
}

// NewGateway 创建新的网关实例
func NewGateway(config *GatewayConfig) *Gateway {
	gateway := &Gateway{
		config:           config,
		router:           mux.NewRouter(),
		pluginRegistry:   NewPluginRegistry(),
		loadBalancer:     NewLoadBalancer(),
		rateLimiter:      NewRateLimiter(),
		authManager:      NewAuthManager(),
		metricsCollector: NewMetricsCollector(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该更严格
			},
		},
		middleware: make([]Middleware, 0),
	}

	gateway.setupRoutes()
	gateway.setupMiddleware()

	return gateway
}

// Start 启动网关
func (g *Gateway) Start() error {
	g.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", g.config.Host, g.config.Port),
		Handler:        g.router,
		ReadTimeout:    g.config.ReadTimeout,
		WriteTimeout:   g.config.WriteTimeout,
		IdleTimeout:    g.config.IdleTimeout,
		MaxHeaderBytes: g.config.MaxHeaderBytes,
	}

	g.logger.Printf("Starting gateway on %s:%d", g.config.Host, g.config.Port)

	if g.config.TLSCertFile != "" && g.config.TLSKeyFile != "" {
		return g.server.ListenAndServeTLS(g.config.TLSCertFile, g.config.TLSKeyFile)
	}

	return g.server.ListenAndServe()
}

// Stop 停止网关
func (g *Gateway) Stop(ctx context.Context) error {
	g.logger.Println("Stopping gateway...")
	return g.server.Shutdown(ctx)
}

// RegisterPlugin 注册插件
func (g *Gateway) RegisterPlugin(plugin *PluginEndpoint) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.pluginRegistry.RegisterPlugin(plugin)
}

// UnregisterPlugin 注销插件
func (g *Gateway) UnregisterPlugin(pluginID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.pluginRegistry.UnregisterPlugin(pluginID)
}

// AddRoute 添加路由
func (g *Gateway) AddRoute(route *RouteConfig) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.pluginRegistry.AddRoute(route)
}

// RemoveRoute 移除路由
func (g *Gateway) RemoveRoute(path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.pluginRegistry.RemoveRoute(path)
}

// setupRoutes 设置路由
func (g *Gateway) setupRoutes() {
	// 健康检查
	g.router.HandleFunc(g.config.HealthCheckPath, g.handleHealth).Methods("GET")

	// 指标
	if g.config.EnableMetrics {
		g.router.HandleFunc(g.config.MetricsPath, g.handleMetrics).Methods("GET")
	}

	// 管理接口
	adminRouter := g.router.PathPrefix(g.config.AdminPath).Subrouter()
	adminRouter.HandleFunc("/plugins", g.handleListPlugins).Methods("GET")
	adminRouter.HandleFunc("/plugins/{id}", g.handleGetPlugin).Methods("GET")
	adminRouter.HandleFunc("/plugins/{id}", g.handleUpdatePlugin).Methods("PUT")
	adminRouter.HandleFunc("/plugins/{id}", g.handleDeletePlugin).Methods("DELETE")
	adminRouter.HandleFunc("/routes", g.handleListRoutes).Methods("GET")
	adminRouter.HandleFunc("/routes", g.handleCreateRoute).Methods("POST")
	adminRouter.HandleFunc("/routes/{path:.*}", g.handleDeleteRoute).Methods("DELETE")

	// WebSocket支持
	if g.config.EnableWebSocket {
		g.router.HandleFunc("/ws/{plugin}/{path:.*}", g.handleWebSocket)
	}

	// 插件路由（通配符路由，放在最后）
	g.router.PathPrefix("/").HandlerFunc(g.handlePluginRequest)
}

// setupMiddleware 设置中间件
func (g *Gateway) setupMiddleware() {
	// CORS中间件
	if g.config.EnableCORS {
		g.router.Use(g.corsMiddleware)
	}

	// 日志中间件
	g.router.Use(g.loggingMiddleware)

	// 指标中间件
	if g.config.EnableMetrics {
		g.router.Use(g.metricsMiddleware)
	}

	// 认证中间件
	if g.config.EnableAuth {
		g.router.Use(g.authMiddleware)
	}

	// 限流中间件
	if g.config.EnableRateLimit {
		g.router.Use(g.rateLimitMiddleware)
	}
}

// handlePluginRequest 处理插件请求
func (g *Gateway) handlePluginRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 查找匹配的路由
	route := g.findMatchingRoute(r.URL.Path, r.Method)
	if route == nil {
		g.sendError(w, http.StatusNotFound, "Route not found")
		return
	}

	// 获取插件实例
	plugin := g.pluginRegistry.GetPlugin(route.PluginID)
	if plugin == nil {
		g.sendError(w, http.StatusServiceUnavailable, "Plugin not available")
		return
	}

	// 选择实例
	instance := g.loadBalancer.SelectInstance(plugin.Instances)
	if instance == nil {
		g.sendError(w, http.StatusServiceUnavailable, "No healthy instances available")
		return
	}

	// 创建反向代理
	proxy := g.createReverseProxy(instance, route)

	// 执行请求
	proxy.ServeHTTP(w, r)

	// 记录指标
	latency := time.Since(startTime).Seconds()
	g.metricsCollector.RecordRequest(route.PluginID, latency)
}

// findMatchingRoute 查找匹配的路由
func (g *Gateway) findMatchingRoute(path, method string) *RouteConfig {
	g.pluginRegistry.mu.RLock()
	defer g.pluginRegistry.mu.RUnlock()

	for routePath, route := range g.pluginRegistry.routes {
		if g.matchRoute(routePath, path) && g.matchMethod(route.Methods, method) {
			return route
		}
	}

	return nil
}

// matchRoute 匹配路由路径
func (g *Gateway) matchRoute(routePath, requestPath string) bool {
	// 简单的路径匹配，可以扩展为更复杂的模式匹配
	if strings.HasSuffix(routePath, "*") {
		prefix := strings.TrimSuffix(routePath, "*")
		return strings.HasPrefix(requestPath, prefix)
	}
	return routePath == requestPath
}

// matchMethod 匹配HTTP方法
func (g *Gateway) matchMethod(allowedMethods []string, method string) bool {
	if len(allowedMethods) == 0 {
		return true // 允许所有方法
	}

	for _, allowedMethod := range allowedMethods {
		if allowedMethod == method {
			return true
		}
	}

	return false
}

// createReverseProxy 创建反向代理
func (g *Gateway) createReverseProxy(instance *PluginInstance, route *RouteConfig) *httputil.ReverseProxy {
	target, _ := url.Parse(fmt.Sprintf("http://%s:%d", instance.Address, instance.Port))

	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 重写路径
		if route.Rewrite != "" {
			req.URL.Path = route.Rewrite
		} else if route.StripPrefix {
			// 移除路由前缀
			req.URL.Path = strings.TrimPrefix(req.URL.Path, route.Path)
		}

		// 添加自定义头
		for key, value := range route.Headers {
			req.Header.Set(key, value)
		}

		// 添加网关信息头部
		req.Header.Set("X-Gateway-Plugin", route.PluginID)
		req.Header.Set("X-Gateway-Instance", instance.ID)
		req.Header.Set("X-Gateway-Timestamp", time.Now().Format(time.RFC3339))
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		g.logger.Printf("Proxy error: %v", err)
		g.metricsCollector.RecordError(route.PluginID)
		g.sendError(w, http.StatusBadGateway, "Service temporarily unavailable")
	}

	return proxy
}

// handleWebSocket 处理WebSocket连接
func (g *Gateway) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginID := vars["plugin"]
	path := vars["path"]

	// 获取插件实例
	plugin := g.pluginRegistry.GetPlugin(pluginID)
	if plugin == nil {
		g.sendError(w, http.StatusServiceUnavailable, "Plugin not available")
		return
	}

	instance := g.loadBalancer.SelectInstance(plugin.Instances)
	if instance == nil {
		g.sendError(w, http.StatusServiceUnavailable, "No healthy instances available")
		return
	}

	// 升级到WebSocket连接
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.logger.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 连接到后端WebSocket
	backendURL := fmt.Sprintf("ws://%s:%d/%s", instance.Address, instance.Port, path)
	backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
	if err != nil {
		g.logger.Printf("Backend WebSocket connection error: %v", err)
		return
	}
	defer backendConn.Close()

	// 双向代理WebSocket消息
	g.proxyWebSocket(conn, backendConn)
}

// proxyWebSocket 代理WebSocket消息
func (g *Gateway) proxyWebSocket(client, backend *websocket.Conn) {
	done := make(chan struct{})

	// 客户端到后端
	go func() {
		defer close(done)
		for {
			messageType, message, err := client.ReadMessage()
			if err != nil {
				return
			}
			if err := backend.WriteMessage(messageType, message); err != nil {
				return
			}
		}
	}()

	// 后端到客户端
	go func() {
		defer close(done)
		for {
			messageType, message, err := backend.ReadMessage()
			if err != nil {
				return
			}
			if err := client.WriteMessage(messageType, message); err != nil {
				return
			}
		}
	}()

	<-done
}

// handleHealth 处理健康检查请求
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"plugins":   g.pluginRegistry.GetHealthSummary(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleMetrics 处理指标请求
func (g *Gateway) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := g.metricsCollector.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// sendError 发送错误响应
func (g *Gateway) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     true,
		"message":   message,
		"timestamp": time.Now(),
	})
}

// 中间件实现
func (g *Gateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		g.logger.Printf("%s %s %v", r.Method, r.URL.Path, duration)
	})
}

func (g *Gateway) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.metricsCollector.IncrementActiveRequests()
		defer g.metricsCollector.DecrementActiveRequests()
		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 跳过健康检查和指标端点
		if r.URL.Path == g.config.HealthCheckPath || r.URL.Path == g.config.MetricsPath {
			next.ServeHTTP(w, r)
			return
		}

		// 验证认证
		if !g.authManager.ValidateRequest(r) {
			g.sendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := g.getClientIP(r)
		if !g.rateLimiter.Allow(clientIP) {
			g.sendError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP 获取客户端IP
func (g *Gateway) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头部
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// 检查X-Real-IP头部
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
