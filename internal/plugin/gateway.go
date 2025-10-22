package plugin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ClientType 客户端类型
type ClientType string

const (
	ClientTypeWeb     ClientType = "web"
	ClientTypeMobile  ClientType = "mobile"
	ClientTypeIoT     ClientType = "iot"
	ClientTypeDesktop ClientType = "desktop"
	ClientTypeAPI     ClientType = "api"
)

// PluginGateway 插件网关类型
type PluginGateway struct {
	inProcessManager    *PluginLoaderManager
	microserviceManager *MicroservicePluginManager
	routes              map[string]*RouteConfig
	middleware          []gin.HandlerFunc
	logger              *logrus.Logger
	mutex               sync.RWMutex
}

// RouteConfig 路由配置
type RouteConfig struct {
	PluginID     string                 `json:"plugin_id"`
	PluginType   string                 `json:"plugin_type"`   // in_process, microservice
	Method       string                 `json:"method"`        // GET, POST, PUT, DELETE
	Path         string                 `json:"path"`          // /api/plugins/{plugin_id}/{method}
	Handler      string                 `json:"handler"`       // 处理函数名称
	ClientTypes  []ClientType           `json:"client_types"`  // 支持的客户端类型
	AuthRequired bool                   `json:"auth_required"` // 是否需要认证
	RateLimit    *RateLimitConfig       `json:"rate_limit"`    // 限流配置
	Timeout      time.Duration          `json:"timeout"`       // 超时时间
	Metadata     map[string]interface{} `json:"metadata"`      // 元数据
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// GatewayRequest 网关请求
type GatewayRequest struct {
	PluginID   string                 `json:"plugin_id"`
	Method     string                 `json:"method"`
	Params     map[string]interface{} `json:"params"`
	Headers    map[string]string      `json:"headers"`
	ClientType ClientType             `json:"client_type"`
	UserID     string                 `json:"user_id"`
	RequestID  string                 `json:"request_id"`
	Context    map[string]interface{} `json:"context"`
}

// GatewayResponse 网关响应
type GatewayResponse struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data,omitempty"`
	Error     *string                `json:"error,omitempty"`
	Duration  int64                  `json:"duration"`
	RequestID string                 `json:"request_id"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewPluginGateway 创建插件网关
func NewPluginGateway(
	inProcessManager *PluginLoaderManager,
	microserviceManager *MicroservicePluginManager,
	logger *logrus.Logger,
) *PluginGateway {
	return &PluginGateway{
		inProcessManager:    inProcessManager,
		microserviceManager: microserviceManager,
		routes:              make(map[string]*RouteConfig),
		logger:              logger,
	}
}

// RegisterRoute 注册路由
func (g *PluginGateway) RegisterRoute(config *RouteConfig) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	routeKey := fmt.Sprintf("%s:%s", config.Method, config.Path)
	g.routes[routeKey] = config

	g.logger.Infof("Registered route: %s %s -> Plugin: %s", config.Method, config.Path, config.PluginID)
	return nil
}

// UnregisterRoute 注销路由
func (g *PluginGateway) UnregisterRoute(method, path string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	routeKey := fmt.Sprintf("%s:%s", method, path)
	delete(g.routes, routeKey)

	g.logger.Infof("Unregistered route: %s %s", method, path)
	return nil
}

// SetupRoutes 设置Gin路由
func (g *PluginGateway) SetupRoutes(router *gin.Engine) {
	// 插件调用API
	api := router.Group("/api/plugins")
	{
		api.POST("/:plugin_id/call", g.handlePluginCall)
		api.GET("/:plugin_id/status", g.handlePluginStatus)
		api.GET("/:plugin_id/metrics", g.handlePluginMetrics)
		api.POST("/:plugin_id/start", g.handlePluginStart)
		api.POST("/:plugin_id/stop", g.handlePluginStop)
		api.POST("/:plugin_id/restart", g.handlePluginRestart)
	}

	// 动态路由（根据注册的路由配置）
	g.setupDynamicRoutes(router)
}

// setupDynamicRoutes 设置动态路由
func (g *PluginGateway) setupDynamicRoutes(router *gin.Engine) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	for routeKey, config := range g.routes {
		parts := strings.SplitN(routeKey, ":", 2)
		if len(parts) != 2 {
			continue
		}

		method := parts[0]
		path := parts[1]

		// 创建处理函数
		handler := g.createDynamicHandler(config)

		// 注册路由
		switch strings.ToUpper(method) {
		case "GET":
			router.GET(path, handler)
		case "POST":
			router.POST(path, handler)
		case "PUT":
			router.PUT(path, handler)
		case "DELETE":
			router.DELETE(path, handler)
		case "PATCH":
			router.PATCH(path, handler)
		}
	}
}

// createDynamicHandler 创建动态处理函数
func (g *PluginGateway) createDynamicHandler(config *RouteConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := uuid.New().String()

		// 检查客户端类型
		clientType := g.detectClientType(c)
		if !g.isClientTypeSupported(config, clientType) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Client type not supported",
				"request_id": requestID,
			})
			return
		}

		// 构建请求
		request := &GatewayRequest{
			PluginID:   config.PluginID,
			Method:     config.Handler,
			Headers:    g.extractHeaders(c),
			ClientType: clientType,
			RequestID:  requestID,
		}

		// 提取参数
		if err := g.extractParams(c, request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      fmt.Sprintf("Invalid parameters: %v", err),
				"request_id": requestID,
			})
			return
		}

		// 调用插件
		result, err := g.callPlugin(request, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      err.Error(),
				"request_id": requestID,
			})
			return
		}

		// 构建响应
		response := &GatewayResponse{
			Success:   result.Success,
			Data:      result.Data,
			Duration:  time.Since(startTime).Milliseconds(),
			RequestID: requestID,
			Metadata:  result.Metadata,
		}

		if !result.Success {
			response.Error = &result.Error
		}

		c.JSON(http.StatusOK, response)
	}
}

// handlePluginCall 处理插件调用
func (g *PluginGateway) handlePluginCall(c *gin.Context) {
	startTime := time.Now()
	pluginID := c.Param("plugin_id")
	requestID := uuid.New().String()

	var request GatewayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      fmt.Sprintf("Invalid request: %v", err),
			"request_id": requestID,
		})
		return
	}

	request.PluginID = pluginID
	request.RequestID = requestID
	request.ClientType = g.detectClientType(c)
	request.Headers = g.extractHeaders(c)

	// 调用插件
	result, err := g.callPlugin(&request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	// 构建响应
	response := &GatewayResponse{
		Success:   result.Success,
		Data:      result.Data,
		Duration:  time.Since(startTime).Milliseconds(),
		RequestID: requestID,
		Metadata:  result.Metadata,
	}

	if !result.Success {
		response.Error = &result.Error
	}

	c.JSON(http.StatusOK, response)
}

// callPlugin 调用插件
func (g *PluginGateway) callPlugin(request *GatewayRequest, config *RouteConfig) (*PluginResult, error) {
	// 确定插件类型
	pluginType, err := g.determinePluginType(request.PluginID)
	if err != nil {
		return nil, err
	}

	// 创建插件上下文
	ctx := context.WithValue(context.Background(), "request_id", request.RequestID)
	if config != nil && config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// 根据插件类型调用
	switch pluginType {
	case "in_process":
		return g.callInProcessPlugin(ctx, request)
	case "microservice":
		return g.callMicroservicePlugin(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", pluginType)
	}
}

// callInProcessPlugin 调用进程内插件
func (g *PluginGateway) callInProcessPlugin(ctx context.Context, request *GatewayRequest) (*PluginResult, error) {
	// 获取插件实例
	plugin, err := g.inProcessManager.GetPlugin(request.PluginID)
	if err != nil {
		return nil, fmt.Errorf("failed to get in-process plugin: %w", err)
	}

	// 根据插件接口类型调用相应方法
	switch p := plugin.(type) {
	case DataProcessor:
		return p.ProcessData(ctx, request.Params)
	case ImageFilter:
		if imageData, ok := request.Params["image_data"].([]byte); ok {
			return p.FilterImage(ctx, imageData, request.Params)
		}
		return nil, fmt.Errorf("missing image_data parameter")
	case TextAnalyzer:
		if text, ok := request.Params["text"].(string); ok {
			return p.AnalyzeText(ctx, text, request.Params)
		}
		return nil, fmt.Errorf("missing text parameter")
	case APIConnector:
		endpoint := request.Params["endpoint"].(string)
		method := request.Params["method"].(string)
		body := request.Params["body"]
		return p.CallAPI(ctx, endpoint, method, request.Headers, body)
	default:
		return nil, fmt.Errorf("unsupported plugin interface type")
	}
}

// callMicroservicePlugin 调用微服务插件
func (g *PluginGateway) callMicroservicePlugin(ctx context.Context, request *GatewayRequest) (*PluginResult, error) {
	return g.microserviceManager.CallPlugin(request.PluginID, request.Method, request.Params)
}

// determinePluginType 确定插件类型
func (g *PluginGateway) determinePluginType(pluginID string) (string, error) {
	// 首先尝试从进程内插件管理器获取插件实例
	if _, err := g.inProcessManager.GetPlugin(pluginID); err == nil {
		return "in_process", nil
	}

	// 然后尝试从微服务插件管理器获取插件实例
	if _, err := g.microserviceManager.GetPlugin(pluginID); err == nil {
		return "microservice", nil
	}

	return "", fmt.Errorf("plugin not found: %s", pluginID)
}

// detectClientType 检测客户端类型
func (g *PluginGateway) detectClientType(c *gin.Context) ClientType {
	userAgent := c.GetHeader("User-Agent")
	clientType := c.GetHeader("X-Client-Type")

	if clientType != "" {
		return ClientType(strings.ToLower(clientType))
	}

	// 根据User-Agent推断客户端类型
	userAgent = strings.ToLower(userAgent)
	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") {
		return ClientTypeMobile
	}
	if strings.Contains(userAgent, "electron") || strings.Contains(userAgent, "desktop") {
		return ClientTypeDesktop
	}
	if strings.Contains(userAgent, "iot") || strings.Contains(userAgent, "device") {
		return ClientTypeIoT
	}

	return ClientTypeWeb
}

// isClientTypeSupported 检查客户端类型是否支持
func (g *PluginGateway) isClientTypeSupported(config *RouteConfig, clientType ClientType) bool {
	if len(config.ClientTypes) == 0 {
		return true // 如果没有限制，则支持所有客户端类型
	}

	for _, supportedType := range config.ClientTypes {
		if supportedType == clientType {
			return true
		}
	}

	return false
}

// extractHeaders 提取请求头
func (g *PluginGateway) extractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// extractParams 提取请求参数
func (g *PluginGateway) extractParams(c *gin.Context, request *GatewayRequest) error {
	params := make(map[string]interface{})

	// 提取URL参数
	for key, value := range c.Request.URL.Query() {
		if len(value) > 0 {
			params[key] = value[0]
		}
	}

	// 提取路径参数
	for _, param := range c.Params {
		params[param.Key] = param.Value
	}

	// 提取请求体参数
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		var bodyParams map[string]interface{}
		if err := c.ShouldBindJSON(&bodyParams); err == nil {
			for key, value := range bodyParams {
				params[key] = value
			}
		}
	}

	request.Params = params
	return nil
}

// handlePluginStatus 处理插件状态查询
func (g *PluginGateway) handlePluginStatus(c *gin.Context) {
	pluginID := c.Param("plugin_id")

	// 尝试获取插件状态
	status := "unknown"
	pluginType := "unknown"

	// 检查进程内插件
	if plugin, err := g.inProcessManager.GetPlugin(pluginID); err == nil {
		if err := plugin.HealthCheck(); err == nil {
			status = "running"
		} else {
			status = "error"
		}
		pluginType = "in_process"
	} else if pluginStatus, err := g.microserviceManager.GetPluginStatus(pluginID); err == nil {
		status = pluginStatus
		pluginType = "microservice"
	}

	c.JSON(http.StatusOK, gin.H{
		"plugin_id":   pluginID,
		"status":      status,
		"plugin_type": pluginType,
		"timestamp":   time.Now(),
	})
}

// handlePluginMetrics 处理插件指标查询
func (g *PluginGateway) handlePluginMetrics(c *gin.Context) {
	pluginID := c.Param("plugin_id")

	// 尝试获取插件指标
	var metrics map[string]interface{}
	var err error

	// 检查微服务插件指标
	if metrics, err = g.microserviceManager.GetPluginMetrics(pluginID); err != nil {
		// 如果是进程内插件，返回基础指标
		if _, pluginErr := g.inProcessManager.GetPlugin(pluginID); pluginErr == nil {
			metrics = map[string]interface{}{
				"plugin_id":    pluginID,
				"plugin_type":  "in_process",
				"status":       "running",
				"last_updated": time.Now(),
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Plugin not found: %s", pluginID),
			})
			return
		}
	}

	c.JSON(http.StatusOK, metrics)
}

// handlePluginStart 处理插件启动
func (g *PluginGateway) handlePluginStart(c *gin.Context) {
	pluginID := c.Param("plugin_id")

	// 只有微服务插件支持启动操作
	if _, err := g.microserviceManager.GetPlugin(pluginID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Microservice plugin not found: %s", pluginID),
		})
		return
	}

	// 这里需要实现启动逻辑，简化处理
	if err := g.microserviceManager.StartPlugin(pluginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Plugin start requested",
		"plugin_id": pluginID,
	})
}

// handlePluginStop 处理插件停止
func (g *PluginGateway) handlePluginStop(c *gin.Context) {
	pluginID := c.Param("plugin_id")

	if err := g.microserviceManager.StopPlugin(pluginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Plugin stopped successfully",
		"plugin_id": pluginID,
	})
}

// handlePluginRestart 处理插件重启
func (g *PluginGateway) handlePluginRestart(c *gin.Context) {
	pluginID := c.Param("plugin_id")

	if err := g.microserviceManager.RestartPlugin(pluginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Plugin restarted successfully",
		"plugin_id": pluginID,
	})
}

// AddMiddleware 添加中间件
func (g *PluginGateway) AddMiddleware(middleware gin.HandlerFunc) {
	g.middleware = append(g.middleware, middleware)
}

// GetRoutes 获取所有路由配置
func (g *PluginGateway) GetRoutes() map[string]*RouteConfig {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	routes := make(map[string]*RouteConfig)
	for key, config := range g.routes {
		routes[key] = config
	}

	return routes
}
