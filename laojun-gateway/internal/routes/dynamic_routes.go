package routes

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/codetaoist/laojun-gateway/internal/handlers"
	"github.com/codetaoist/laojun-gateway/internal/middleware"
	"github.com/codetaoist/laojun-gateway/internal/proxy"
	"github.com/codetaoist/laojun-gateway/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DynamicRouteManager 动态路由管理器
type DynamicRouteManager struct {
	router         *gin.Engine
	routes         map[string]*RouteInfo
	routesMutex    sync.RWMutex
	config         *config.Config
	serviceManager *services.ServiceManager
	proxyService   *proxy.Service
	logger         *zap.Logger
}

// RouteInfo 路由信息
type RouteInfo struct {
	ID          string                 `json:"id"`
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	Service     string                 `json:"service"`
	Target      string                 `json:"target"`
	StripPrefix bool                   `json:"strip_prefix"`
	Headers     map[string]string      `json:"headers"`
	Auth        bool                   `json:"auth"`
	RateLimit   *config.RateLimitRule  `json:"rate_limit,omitempty"`
	Middleware  []string               `json:"middleware"`
	Timeout     int                    `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      string                 `json:"status"` // active, inactive, deprecated
}

// NewDynamicRouteManager 创建动态路由管理器
func NewDynamicRouteManager(
	router *gin.Engine,
	cfg *config.Config,
	serviceManager *services.ServiceManager,
	logger *zap.Logger,
) *DynamicRouteManager {
	proxyService := proxy.NewService(cfg.Proxy, serviceManager.GetDiscovery(), logger)

	return &DynamicRouteManager{
		router:         router,
		routes:         make(map[string]*RouteInfo),
		config:         cfg,
		serviceManager: serviceManager,
		proxyService:   proxyService,
		logger:         logger,
	}
}

// AddRoute 添加路由
func (drm *DynamicRouteManager) AddRoute(route *RouteInfo) error {
	drm.routesMutex.Lock()
	defer drm.routesMutex.Unlock()

	// 验证路由配置
	if err := drm.validateRoute(route); err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}

	// 生成路由ID
	if route.ID == "" {
		route.ID = drm.generateRouteID(route.Path, route.Method)
	}

	// 检查路由是否已存在
	if _, exists := drm.routes[route.ID]; exists {
		return fmt.Errorf("route with ID %s already exists", route.ID)
	}

	// 设置默认值
	drm.setRouteDefaults(route)

	// 注册路由到Gin
	if err := drm.registerGinRoute(route); err != nil {
		return fmt.Errorf("failed to register route: %w", err)
	}

	// 保存路由信息
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()
	route.Status = "active"
	drm.routes[route.ID] = route

	drm.logger.Info("Route added successfully",
		zap.String("id", route.ID),
		zap.String("path", route.Path),
		zap.String("method", route.Method),
		zap.String("service", route.Service))

	return nil
}

// UpdateRoute 更新路由
func (drm *DynamicRouteManager) UpdateRoute(routeID string, updatedRoute *RouteInfo) error {
	drm.routesMutex.Lock()
	defer drm.routesMutex.Unlock()

	// 检查路由是否存在
	existingRoute, exists := drm.routes[routeID]
	if !exists {
		return fmt.Errorf("route with ID %s not found", routeID)
	}

	// 验证更新的路由配置
	if err := drm.validateRoute(updatedRoute); err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}

	// 保留原有的创建时间和ID
	updatedRoute.ID = routeID
	updatedRoute.CreatedAt = existingRoute.CreatedAt
	updatedRoute.UpdatedAt = time.Now()

	// 设置默认值
	drm.setRouteDefaults(updatedRoute)

	// 注册新路由到Gin（会覆盖原有路由）
	if err := drm.registerGinRoute(updatedRoute); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	// 更新路由信息
	drm.routes[routeID] = updatedRoute

	drm.logger.Info("Route updated successfully",
		zap.String("id", routeID),
		zap.String("path", updatedRoute.Path),
		zap.String("method", updatedRoute.Method))

	return nil
}

// RemoveRoute 删除路由
func (drm *DynamicRouteManager) RemoveRoute(routeID string) error {
	drm.routesMutex.Lock()
	defer drm.routesMutex.Unlock()

	// 检查路由是否存在
	route, exists := drm.routes[routeID]
	if !exists {
		return fmt.Errorf("route with ID %s not found", routeID)
	}

	// 从路由表中删除
	delete(drm.routes, routeID)

	drm.logger.Info("Route removed successfully",
		zap.String("id", routeID),
		zap.String("path", route.Path),
		zap.String("method", route.Method))

	return nil
}

// GetRoute 获取路由信息
func (drm *DynamicRouteManager) GetRoute(routeID string) (*RouteInfo, error) {
	drm.routesMutex.RLock()
	defer drm.routesMutex.RUnlock()

	route, exists := drm.routes[routeID]
	if !exists {
		return nil, fmt.Errorf("route with ID %s not found", routeID)
	}

	return route, nil
}

// ListRoutes 列出所有路由
func (drm *DynamicRouteManager) ListRoutes() map[string]*RouteInfo {
	drm.routesMutex.RLock()
	defer drm.routesMutex.RUnlock()

	// 创建副本以避免并发访问问题
	routes := make(map[string]*RouteInfo)
	for id, route := range drm.routes {
		routes[id] = route
	}

	return routes
}

// GetRoutesByService 根据服务名获取路由
func (drm *DynamicRouteManager) GetRoutesByService(serviceName string) []*RouteInfo {
	drm.routesMutex.RLock()
	defer drm.routesMutex.RUnlock()

	var routes []*RouteInfo
	for _, route := range drm.routes {
		if route.Service == serviceName {
			routes = append(routes, route)
		}
	}

	return routes
}

// ToggleRoute 启用/禁用路由
func (drm *DynamicRouteManager) ToggleRoute(routeID string, active bool) error {
	drm.routesMutex.Lock()
	defer drm.routesMutex.Unlock()

	route, exists := drm.routes[routeID]
	if !exists {
		return fmt.Errorf("route with ID %s not found", routeID)
	}

	if active {
		route.Status = "active"
	} else {
		route.Status = "inactive"
	}

	route.UpdatedAt = time.Now()

	drm.logger.Info("Route status toggled",
		zap.String("id", routeID),
		zap.String("status", route.Status))

	return nil
}

// registerGinRoute 注册路由到Gin
func (drm *DynamicRouteManager) registerGinRoute(route *RouteInfo) error {
	// 创建处理器链
	handlers := drm.buildHandlerChain(route)

	// 根据HTTP方法注册路由
	switch route.Method {
	case "GET":
		drm.router.GET(route.Path, handlers...)
	case "POST":
		drm.router.POST(route.Path, handlers...)
	case "PUT":
		drm.router.PUT(route.Path, handlers...)
	case "DELETE":
		drm.router.DELETE(route.Path, handlers...)
	case "PATCH":
		drm.router.PATCH(route.Path, handlers...)
	case "HEAD":
		drm.router.HEAD(route.Path, handlers...)
	case "OPTIONS":
		drm.router.OPTIONS(route.Path, handlers...)
	case "ANY":
		drm.router.Any(route.Path, handlers...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", route.Method)
	}

	return nil
}

// buildHandlerChain 构建处理器链
func (drm *DynamicRouteManager) buildHandlerChain(route *RouteInfo) []gin.HandlerFunc {
	var handlers []gin.HandlerFunc

	// 添加路由状态检查中间件
	handlers = append(handlers, drm.routeStatusMiddleware(route.ID))

	// 添加自定义中间件
	for _, middlewareName := range route.Middleware {
		if mw := drm.getMiddleware(middlewareName); mw != nil {
			handlers = append(handlers, mw)
		}
	}

	// 添加认证中间件
	if route.Auth {
		authHandler := handlers.NewAuthHandler(drm.config.Auth, drm.logger)
		handlers = append(handlers, middleware.AuthMiddleware(authHandler))
	}

	// 添加路由级限流中间件
	if route.RateLimit != nil {
		handlers = append(handlers, drm.routeRateLimitMiddleware(route.RateLimit))
	}

	// 添加代理处理器
	proxyHandler := handlers.NewProxyHandler(drm.proxyService, drm.logger)
	handlers = append(handlers, drm.proxyMiddleware(route, proxyHandler))

	return handlers
}

// routeStatusMiddleware 路由状态检查中间件
func (drm *DynamicRouteManager) routeStatusMiddleware(routeID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		drm.routesMutex.RLock()
		route, exists := drm.routes[routeID]
		drm.routesMutex.RUnlock()

		if !exists || route.Status != "active" {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Route is not available",
				"route_id": routeID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// routeRateLimitMiddleware 路由级限流中间件
func (drm *DynamicRouteManager) routeRateLimitMiddleware(rule *config.RateLimitRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 实现路由级限流逻辑
		// 这里可以集成现有的限流服务
		rateLimitService := drm.serviceManager.GetRateLimit()
		key := fmt.Sprintf("route:%s:%s", rule.Path, rule.Method)
		
		if !rateLimitService.Allow(key, rule.Rate) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded for this route",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// proxyMiddleware 代理中间件
func (drm *DynamicRouteManager) proxyMiddleware(route *RouteInfo, proxyHandler *handlers.ProxyHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置路由特定的配置
		c.Set("route_config", route)
		
		// 调用代理处理器
		proxyHandler.HandleRequest(c)
	}
}

// getMiddleware 获取中间件
func (drm *DynamicRouteManager) getMiddleware(name string) gin.HandlerFunc {
	switch name {
	case "cors":
		return middleware.CORSMiddleware()
	case "request_id":
		return middleware.RequestIDMiddleware()
	case "monitoring":
		return middleware.MonitoringMiddleware(drm.logger)
	default:
		drm.logger.Warn("Unknown middleware", zap.String("name", name))
		return nil
	}
}

// validateRoute 验证路由配置
func (drm *DynamicRouteManager) validateRoute(route *RouteInfo) error {
	if route.Path == "" {
		return fmt.Errorf("path is required")
	}
	if route.Method == "" {
		return fmt.Errorf("method is required")
	}
	if route.Service == "" && route.Target == "" {
		return fmt.Errorf("either service or target is required")
	}
	return nil
}

// setRouteDefaults 设置路由默认值
func (drm *DynamicRouteManager) setRouteDefaults(route *RouteInfo) {
	if route.Timeout == 0 {
		route.Timeout = drm.config.Proxy.Timeout
	}
	if route.RetryCount == 0 {
		route.RetryCount = drm.config.Proxy.RetryCount
	}
	if route.Headers == nil {
		route.Headers = make(map[string]string)
	}
	if route.Middleware == nil {
		route.Middleware = []string{}
	}
}

// generateRouteID 生成路由ID
func (drm *DynamicRouteManager) generateRouteID(path, method string) string {
	return fmt.Sprintf("%s_%s_%d", method, path, time.Now().Unix())
}