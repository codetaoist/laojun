package routes

import (
	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/codetaoist/laojun-gateway/internal/handlers"
	"github.com/codetaoist/laojun-gateway/internal/middleware"
	"github.com/codetaoist/laojun-gateway/internal/proxy"
	"github.com/codetaoist/laojun-gateway/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes 设置路由
func SetupRoutes(cfg *config.Config, serviceManager *services.ServiceManager, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// 基础中间件
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.MonitoringMiddleware(logger))

	// 初始化增强中间件
	enhancedAuth := middleware.NewEnhancedAuthMiddleware(cfg.Auth, logger)
	enhancedRateLimit := middleware.NewEnhancedRateLimitMiddleware(&cfg.RateLimit, logger)
	circuitBreaker := middleware.NewCircuitBreakerMiddleware(&cfg.Proxy.CircuitBreaker, logger)

	// 全局限流中间件
	if cfg.RateLimit.Enabled {
		router.Use(enhancedRateLimit.GlobalRateLimitMiddleware())
		router.Use(enhancedRateLimit.IPRateLimitMiddleware())
	}

	// 熔断器中间件
	if cfg.Proxy.CircuitBreaker.Enabled {
		router.Use(circuitBreaker.CircuitBreakerMiddleware())
	}

	// 初始化服务
	proxyService := proxy.NewService(cfg.Proxy, serviceManager.GetDiscovery(), logger)

	// 初始化动态路由管理器
	dynamicRouteManager := NewDynamicRouteManager(router, cfg, serviceManager, logger)

	// 初始化处理器
	healthHandler := handlers.NewHealthHandler(serviceManager, logger)
	authHandler := handlers.NewAuthHandler(cfg.Auth, logger)
	proxyHandler := handlers.NewProxyHandler(proxyService, logger)
	routeHandler := handlers.NewRouteHandler(dynamicRouteManager, logger)
	unifiedConfigHandler := handlers.NewUnifiedConfigHandler(cfg.ConfigManager, logger)

	// 健康检查路由
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 监控指标
		api.GET("/metrics", healthHandler.Metrics)

		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", enhancedAuth.JWTRefreshMiddleware())
			auth.POST("/logout", authHandler.Logout)
		}

		// 管理API（需要管理员权限）
		admin := api.Group("/admin")
		admin.Use(enhancedAuth.RoleBasedAuthMiddleware("admin"))
		{
			// 路由管理
			routes := admin.Group("/routes")
			{
				routes.GET("", routeHandler.ListRoutes)
				routes.POST("", routeHandler.CreateRoute)
				routes.GET("/:id", routeHandler.GetRoute)
				routes.PUT("/:id", routeHandler.UpdateRoute)
				routes.DELETE("/:id", routeHandler.DeleteRoute)
				routes.POST("/:id/toggle", routeHandler.ToggleRoute)
				routes.GET("/stats", routeHandler.GetRouteStats)
				routes.POST("/validate", routeHandler.ValidateRoute)
				routes.GET("/export", routeHandler.ExportRoutes)
				routes.POST("/import", routeHandler.ImportRoutes)
				routes.GET("/service/:service", routeHandler.GetRoutesByService)
			}

			// 中间件管理
			middlewares := admin.Group("/middlewares")
			{
				// 限流管理
				rateLimit := middlewares.Group("/ratelimit")
				{
					rateLimit.GET("/stats", func(c *gin.Context) {
						stats := enhancedRateLimit.GetStats()
						c.JSON(200, gin.H{"stats": stats})
					})
				}

				// 熔断器管理
				circuitBreakerGroup := middlewares.Group("/circuit-breaker")
				{
					circuitBreakerGroup.GET("/status", func(c *gin.Context) {
						status := circuitBreaker.GetAllBreakerStatus()
						c.JSON(200, gin.H{"breakers": status})
					})
					circuitBreakerGroup.GET("/status/:name", func(c *gin.Context) {
						name := c.Param("name")
						status := circuitBreaker.GetBreakerStatus(name)
						if status == nil {
							c.JSON(404, gin.H{"error": "Circuit breaker not found"})
							return
						}
						c.JSON(200, gin.H{"breaker": status})
					})
					circuitBreakerGroup.POST("/reset/:name", func(c *gin.Context) {
						name := c.Param("name")
						if err := circuitBreaker.ResetBreaker(name); err != nil {
							c.JSON(400, gin.H{"error": err.Error()})
							return
						}
						c.JSON(200, gin.H{"message": "Circuit breaker reset successfully"})
					})
					circuitBreakerGroup.POST("/force-open/:name", func(c *gin.Context) {
						name := c.Param("name")
						if err := circuitBreaker.ForceOpen(name); err != nil {
							c.JSON(400, gin.H{"error": err.Error()})
							return
						}
						c.JSON(200, gin.H{"message": "Circuit breaker forced open"})
					})
					circuitBreakerGroup.POST("/force-close/:name", func(c *gin.Context) {
						name := c.Param("name")
						if err := circuitBreaker.ForceClose(name); err != nil {
							c.JSON(400, gin.H{"error": err.Error()})
							return
						}
						c.JSON(200, gin.H{"message": "Circuit breaker forced close"})
					})
				}

				// 认证管理
				authGroup := middlewares.Group("/auth")
				{
					authGroup.GET("/sessions", func(c *gin.Context) {
						sessions := enhancedAuth.GetActiveSessions()
						c.JSON(200, gin.H{"sessions": sessions})
					})
					authGroup.DELETE("/sessions/:user_id/:ip", func(c *gin.Context) {
						userID := c.Param("user_id")
						ip := c.Param("ip")
						enhancedAuth.RevokeSession(userID, ip)
						c.JSON(200, gin.H{"message": "Session revoked successfully"})
					})
				}
			}

			// 统一配置管理
			unified := admin.Group("/unified")
			{
				unified.GET("/config/:key", unifiedConfigHandler.GetConfig)
				unified.GET("/config/:key/type", unifiedConfigHandler.GetConfigWithType)
				unified.PUT("/config/:key", unifiedConfigHandler.SetConfig)
				unified.DELETE("/config/:key", unifiedConfigHandler.DeleteConfig)
				unified.GET("/configs", unifiedConfigHandler.ListConfigs)
				unified.POST("/configs/batch", unifiedConfigHandler.BatchSetConfigs)
				unified.GET("/config/:key/history", unifiedConfigHandler.GetConfigHistory)
				unified.GET("/configs/watch", unifiedConfigHandler.WatchConfigs)
				unified.GET("/config/:key/exists", unifiedConfigHandler.ExistsConfig)
				unified.GET("/health", unifiedConfigHandler.HealthCheck)
			}
		}

		// API Key认证的路由
		apiKey := api.Group("/api-key")
		apiKey.Use(enhancedAuth.APIKeyAuthMiddleware())
		{
			apiKey.Any("/*path", proxyHandler.HandleRequest)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(enhancedAuth.AuthMiddleware())
		protected.Use(enhancedRateLimit.UserRateLimitMiddleware())
		protected.Use(enhancedRateLimit.PathRateLimitMiddleware())
		{
			// 代理所有其他请求
			protected.Any("/*path", proxyHandler.HandleRequest)
		}
	}

	return router
}

// setupProxyRoute 设置代理路由
func setupProxyRoute(group *gin.RouterGroup, route config.RouteConfig, proxyHandler *handlers.ProxyHandler, authMiddleware gin.HandlerFunc, logger *zap.Logger) {
	// 创建路由处理函数
	handler := func(c *gin.Context) {
		// 设置代理服务信息到上下文
		c.Set("proxy_service", route.Service)
		proxyHandler.ProxyRequest(c, route)
	}

	// 创建中间件链
	var middlewares []gin.HandlerFunc
	
	// 如果需要认证，添加认证中间件
	if route.Auth {
		middlewares = append(middlewares, authMiddleware)
	}

	// 添加处理函数
	middlewares = append(middlewares, handler)

	// 根据HTTP方法注册路由
	switch route.Method {
	case "GET":
		group.GET(route.Path, middlewares...)
	case "POST":
		group.POST(route.Path, middlewares...)
	case "PUT":
		group.PUT(route.Path, middlewares...)
	case "DELETE":
		group.DELETE(route.Path, middlewares...)
	case "PATCH":
		group.PATCH(route.Path, middlewares...)
	case "HEAD":
		group.HEAD(route.Path, middlewares...)
	case "OPTIONS":
		group.OPTIONS(route.Path, middlewares...)
	default:
		// 默认支持所有方法
		group.Any(route.Path, middlewares...)
	}

	logger.Info("Proxy route registered",
		zap.String("method", route.Method),
		zap.String("path", route.Path),
		zap.String("service", route.Service),
		zap.String("target", route.Target),
		zap.Bool("auth", route.Auth))
}