package routes

import (
	"github.com/gin-gonic/gin"
	
	"github.com/codetaoist/laojun-plugins/internal/handlers"
	"github.com/codetaoist/laojun-plugins/internal/middleware"
	"github.com/codetaoist/laojun-plugins/internal/services"
)

// SetupExtendedPluginRoutes 设置扩展插件路由
func SetupExtendedPluginRoutes(router *gin.Engine, handler *handlers.ExtendedPluginHandler, authService *services.AuthService) {
	// 公开路由组 - 不需要认证
	publicAPI := router.Group("/api/v1/extended-plugins")
	{
		// 获取插件列表（支持筛选）
		publicAPI.GET("", handler.GetExtendedPlugins)

		// 获取插件分类
		publicAPI.GET("/categories", handler.GetExtendedPluginCategories)

		// 获取单个插件详情
		publicAPI.GET("/:id", handler.GetExtendedPlugin)

		// 验证插件配置
		publicAPI.GET("/:id/validate", handler.ValidatePlugin)

		// 获取插件指标（公开数据）
		publicAPI.GET("/:id/metrics", handler.GetPluginMetrics)

		// 调用插件（某些插件可能允许匿名调用）
		publicAPI.POST("/:id/call", handler.CallPlugin)
	}

	// 需要认证的路由组 - 需要用户登录
	authAPI := router.Group("/api/v1/extended-plugins")
	authAPI.Use(middleware.AuthMiddleware(authService))
	{
		// 创建插件
		authAPI.POST("", handler.CreateExtendedPlugin)

		// 上传插件文件
		authAPI.POST("/upload", handler.UploadPlugin)

		// 更新插件
		authAPI.PUT("/:id", handler.UpdateExtendedPlugin)

		// 更新插件状态（例如：部署、停止、重启）
		authAPI.PATCH("/:id/status", handler.UpdatePluginStatus)

		// 获取插件调用日志
		authAPI.GET("/:id/logs", handler.GetPluginCallLogs)
	}

	// 管理员路由组 - 需要管理员权限
	adminAPI := router.Group("/api/v1/admin/extended-plugins")
	adminAPI.Use(middleware.AuthMiddleware(authService))
	adminAPI.Use(middleware.RequireRole("admin"))
	{
		// 部署插件
		adminAPI.POST("/:id/deploy", handler.DeployPlugin)

		// 停止插件
		adminAPI.POST("/:id/stop", handler.StopPlugin)

		// 重启插件
		adminAPI.POST("/:id/restart", handler.RestartPlugin)
	}

	// 开发者路由组 - 需要开发者权限
	developerAPI := router.Group("/api/v1/developer/extended-plugins")
	developerAPI.Use(middleware.AuthMiddleware(authService))
	developerAPI.Use(middleware.RequireRole("developer"))
	{
		// 开发者可以管理自己的插件
		developerAPI.POST("", handler.CreateExtendedPlugin)
		developerAPI.PUT("/:id", handler.UpdateExtendedPlugin)
		developerAPI.PATCH("/:id/status", handler.UpdatePluginStatus)
		developerAPI.GET("/:id/logs", handler.GetPluginCallLogs)

		// 开发者可以部署、停止、重启自己的插件
		developerAPI.POST("/:id/deploy", handler.DeployPlugin)
		developerAPI.POST("/:id/stop", handler.StopPlugin)
		developerAPI.POST("/:id/restart", handler.RestartPlugin)
	}
}

// SetupPluginGatewayRoutes 设置插件网关路由
func SetupPluginGatewayRoutes(router *gin.Engine, handler *handlers.ExtendedPluginHandler) {
	// 插件调用网关 - 统一入口
	gateway := router.Group("/api/v1/gateway")
	{
		// 通用插件调用接口
		gateway.POST("/plugins/:id/call", handler.CallPlugin)

		// 支持不同客户端类型的调用
		gateway.POST("/plugins/:id/call/:client_type", func(c *gin.Context) {
			// 将客户端类型设置到上下文
			c.Set("client_type", c.Param("client_type"))
			handler.CallPlugin(c)
		})

		// 批量调用多个插件
		gateway.POST("/plugins/batch-call", handler.BatchCallPlugins)

		// 获取插件状态（例如：部署、停止、运行中）
		gateway.GET("/plugins/:id/status", handler.GetPluginStatus)

		// 获取插件健康状态（例如：是否响应正常）
		gateway.GET("/plugins/:id/health", handler.GetPluginHealth)
	}

	// WebSocket 路由用于实时通信
	wsAPI := router.Group("/api/v1/ws")
	{
		// 插件实时调用
		wsAPI.GET("/plugins/:id/call", handler.WebSocketPluginCall)

		// 插件状态订阅
		wsAPI.GET("/plugins/:id/status", handler.WebSocketPluginStatus)

		// 插件日志订阅
		wsAPI.GET("/plugins/:id/logs", handler.WebSocketPluginLogs)
	}
}

// SetupPluginWebhookRoutes 设置插件Webhook路由
func SetupPluginWebhookRoutes(router *gin.Engine, handler *handlers.ExtendedPluginHandler, authService *services.AuthService) {
	// Webhook路由组 - 需要认证
	webhook := router.Group("/api/v1/webhooks")
	webhook.Use(middleware.AuthMiddleware(authService))
	{
		// 插件状态变更通知
		webhook.POST("/plugins/:id/status", handler.PluginStatusWebhook)

		// 插件部署完成通知
		webhook.POST("/plugins/:id/deployed", handler.PluginDeployedWebhook)

		// 插件错误通知
		webhook.POST("/plugins/:id/error", handler.PluginErrorWebhook)

		// 插件指标上报
		webhook.POST("/plugins/:id/metrics", handler.PluginMetricsWebhook)
	}
}
