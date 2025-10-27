package routes

import (
	"github.com/gin-gonic/gin"
	
	"github.com/codetaoist/laojun-plugins/internal/handlers"
	"github.com/codetaoist/laojun-plugins/internal/middleware"
	"github.com/codetaoist/laojun-plugins/internal/services"
)

// SetupAdminPluginRoutes 设置总后台插件管理路由
func SetupAdminPluginRoutes(router *gin.Engine, adminPluginHandler *handlers.AdminPluginHandler, adminAuthService *services.AdminAuthService) {
	// 总后台API路由
	adminAPI := router.Group("/admin/api/v1")
	adminAPI.Use(middleware.AdminAuthMiddleware(adminAuthService)) // 管理员认证中间件

	// 插件管理路由
	plugins := adminAPI.Group("/plugins")
	{
		// 插件列表和搜索
		plugins.GET("", adminPluginHandler.GetPluginsForAdmin)
		plugins.GET("/:id", adminPluginHandler.GetPluginForAdmin)

		// 插件状态管理
		plugins.PUT("/:id/status", adminPluginHandler.UpdatePluginStatus)

		// 插件配置管理
		plugins.GET("/:id/config", adminPluginHandler.GetPluginConfig)
		plugins.PUT("/:id/config", adminPluginHandler.UpdatePluginConfig)

		// 插件统计信息
		plugins.GET("/:id/stats", adminPluginHandler.GetPluginStats)
		plugins.GET("/:id/logs", adminPluginHandler.GetPluginLogs)

		// 插件市场同步
		plugins.POST("/:id/sync-from-marketplace", adminPluginHandler.SyncPluginFromMarketplace)
		plugins.POST("/:id/sync-to-marketplace", adminPluginHandler.SyncPluginToMarketplace)
		plugins.POST("/batch-sync", adminPluginHandler.BatchSyncPlugins)
	}

	// 仪表板路由
	dashboard := adminAPI.Group("/dashboard")
	{
		dashboard.GET("/stats", adminPluginHandler.GetDashboardStats)
	}

	// 插件市场集成路由
	marketplace := adminAPI.Group("/marketplace")
	{
		// 插件市场搜索和浏览
		marketplace.GET("/plugins", adminPluginHandler.SearchMarketplacePlugins)
		marketplace.GET("/plugins/:id", adminPluginHandler.GetMarketplacePlugin)
		marketplace.GET("/categories", adminPluginHandler.GetMarketplaceCategories)
		marketplace.GET("/tags", adminPluginHandler.GetMarketplaceTags)

		// 插件安装和管理
		marketplace.POST("/plugins/:id/install", adminPluginHandler.InstallMarketplacePlugin)
		marketplace.DELETE("/plugins/:id/uninstall", adminPluginHandler.UninstallMarketplacePlugin)

		// 插件发布和更新
		marketplace.POST("/plugins", adminPluginHandler.PublishToMarketplace)
		marketplace.PUT("/plugins/:id", adminPluginHandler.UpdateMarketplacePlugin)
		marketplace.DELETE("/plugins/:id", adminPluginHandler.DeleteFromMarketplace)
	}

	// 系统管理路由
	system := adminAPI.Group("/system")
	{
		// 系统配置
		system.GET("/config", adminPluginHandler.GetSystemConfig)
		system.PUT("/config", adminPluginHandler.UpdateSystemConfig)

		// 系统监控
		system.GET("/health", adminPluginHandler.GetSystemHealth)
		system.GET("/metrics", adminPluginHandler.GetSystemMetrics)

		// 缓存管理
		system.POST("/cache/clear", adminPluginHandler.ClearCache)
		system.GET("/cache/stats", adminPluginHandler.GetCacheStats)
	}

	// 开发者管理路由
	developers := adminAPI.Group("/developers")
	{
		developers.GET("", adminPluginHandler.GetDevelopers)
		developers.GET("/:id", adminPluginHandler.GetDeveloper)
		developers.PUT("/:id/status", adminPluginHandler.UpdateDeveloperStatus)
		developers.GET("/:id/plugins", adminPluginHandler.GetDeveloperPlugins)
		developers.GET("/:id/stats", adminPluginHandler.GetDeveloperStats)
	}

	// 审核管理路由
	reviews := adminAPI.Group("/reviews")
	{
		reviews.GET("/queue", adminPluginHandler.GetReviewQueue)
		reviews.GET("/my-tasks", adminPluginHandler.GetMyReviewTasks)
		reviews.POST("/:id/assign", adminPluginHandler.AssignReviewer)
		reviews.POST("/:id/review", adminPluginHandler.ReviewPlugin)
		reviews.POST("/batch-review", adminPluginHandler.BatchReviewPlugins)
	}

	// 报告和分析路由
	reports := adminAPI.Group("/reports")
	{
		reports.GET("/plugins", adminPluginHandler.GetPluginReports)
		reports.GET("/downloads", adminPluginHandler.GetDownloadReports)
		reports.GET("/revenue", adminPluginHandler.GetRevenueReports)
		reports.GET("/users", adminPluginHandler.GetUserReports)
		reports.GET("/performance", adminPluginHandler.GetPerformanceReports)
	}

	// 日志管理路由
	logs := adminAPI.Group("/logs")
	{
		logs.GET("/system", adminPluginHandler.GetSystemLogs)
		logs.GET("/plugins/:id", adminPluginHandler.GetPluginLogs)
		logs.GET("/audit", adminPluginHandler.GetAuditLogs)
		logs.POST("/export", adminPluginHandler.ExportLogs)
	}

	// 备份和恢复路由
	backup := adminAPI.Group("/backup")
	{
		backup.POST("/create", adminPluginHandler.CreateBackup)
		backup.GET("/list", adminPluginHandler.GetBackupList)
		backup.POST("/:id/restore", adminPluginHandler.RestoreBackup)
		backup.DELETE("/:id", adminPluginHandler.DeleteBackup)
	}
}

// SetupAdminPluginWebRoutes 设置总后台插件管理Web路由
func SetupAdminPluginWebRoutes(router *gin.Engine, adminPluginHandler *handlers.AdminPluginHandler, adminAuthService *services.AdminAuthService) {
	// 总后台Web路由
	admin := router.Group("/admin")
	admin.Use(middleware.AdminAuthMiddleware(adminAuthService)) // 管理员认证中间件

	// 插件管理页面
	plugins := admin.Group("/plugins")
	{
		plugins.GET("", adminPluginHandler.PluginListPage)
		plugins.GET("/:id", adminPluginHandler.PluginDetailPage)
		plugins.GET("/:id/config", adminPluginHandler.PluginConfigPage)
		plugins.GET("/:id/logs", adminPluginHandler.PluginLogsPage)
	}

	// 仪表板页面
	admin.GET("/dashboard", adminPluginHandler.DashboardPage)

	// 插件市场页面
	marketplace := admin.Group("/marketplace")
	{
		marketplace.GET("", adminPluginHandler.MarketplacePage)
		marketplace.GET("/plugins/:id", adminPluginHandler.MarketplacePluginPage)
	}

	// 开发者管理页面
	developers := admin.Group("/developers")
	{
		developers.GET("", adminPluginHandler.DevelopersPage)
		developers.GET("/:id", adminPluginHandler.DeveloperDetailPage)
	}

	// 审核管理页面
	reviews := admin.Group("/reviews")
	{
		reviews.GET("/queue", adminPluginHandler.ReviewQueuePage)
		reviews.GET("/my-tasks", adminPluginHandler.MyReviewTasksPage)
		reviews.GET("/:id", adminPluginHandler.ReviewDetailPage)
	}

	// 报告页面
	reports := admin.Group("/reports")
	{
		reports.GET("", adminPluginHandler.ReportsPage)
		reports.GET("/plugins", adminPluginHandler.PluginReportsPage)
		reports.GET("/downloads", adminPluginHandler.DownloadReportsPage)
		reports.GET("/revenue", adminPluginHandler.RevenueReportsPage)
	}

	// 系统管理页面
	system := admin.Group("/system")
	{
		system.GET("/config", adminPluginHandler.SystemConfigPage)
		system.GET("/health", adminPluginHandler.SystemHealthPage)
		system.GET("/logs", adminPluginHandler.SystemLogsPage)
		system.GET("/backup", adminPluginHandler.BackupPage)
	}
}
