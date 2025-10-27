package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/laojun-admin-api/internal/handlers"
	"github.com/codetaoist/laojun-admin-api/internal/middleware"
	"github.com/codetaoist/laojun-admin-api/internal/services"
)

// SetupPluginAuditRoutes 设置插件审核相关路由
func SetupPluginAuditRoutes(
	router *gin.Engine,
	auditService *services.PluginAuditService,
	logger *logrus.Logger,
) {
	// 创建处理器
	auditHandler := handlers.NewPluginAuditHandler(auditService, logger)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 插件审核路由组
		auditGroup := v1.Group("/plugin-audit")
		auditGroup.Use(middleware.AuthMiddleware()) // 需要认证
		{
			// 提交插件审核
			auditGroup.POST("/submit", auditHandler.SubmitPluginForAudit)

			// 获取审核记录列表
			auditGroup.GET("/records", auditHandler.GetAuditRecords)

			// 获取单个审核记录详情
			auditGroup.GET("/:id", auditHandler.GetAuditRecord)

			// 管理员操作 - 需要管理员权限
			adminGroup := auditGroup.Group("")
			adminGroup.Use(middleware.AdminMiddleware()) // 需要管理员权限
			{
				// 分配审核员
				adminGroup.POST("/:id/assign", auditHandler.AssignAuditor)

				// 获取审核统计信息
				adminGroup.GET("/statistics", auditHandler.GetAuditStatistics)
			}

			// 审核员操作 - 需要审核员权限
			auditorGroup := auditGroup.Group("")
			auditorGroup.Use(middleware.AuditorMiddleware()) // 需要审核员权限
			{
				// 提交审核结果
				auditorGroup.POST("/:id/review", auditHandler.SubmitAuditReview)
			}
		}

		// 开发者认证路由组
		developerGroup := v1.Group("/developer")
		developerGroup.Use(middleware.AuthMiddleware()) // 需要认证
		{
			// 开发者认证申请
			developerGroup.POST("/verify", auditHandler.VerifyDeveloper)
		}

		// 审核员管理路由组
		auditorGroup := v1.Group("/auditor")
		auditorGroup.Use(middleware.AuthMiddleware())  // 需要认证
		auditorGroup.Use(middleware.AdminMiddleware()) // 需要管理员权限
		{
			// 更新审核员档案
			auditorGroup.PUT("/:id/profile", auditHandler.UpdateAuditorProfile)
		}
	}

	// 健康检查路由
	router.GET("/health/plugin-audit", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "plugin-audit",
			"version": "1.0.0",
		})
	})

	logger.Info("Plugin audit routes registered successfully")
}

// SetupPluginAuditWebhookRoutes 设置插件审核Webhook路由
func SetupPluginAuditWebhookRoutes(
	router *gin.Engine,
	auditService *services.PluginAuditService,
	logger *logrus.Logger,
) {
	// Webhook路由组 - 不需要认证，但需要验证签名
	webhookGroup := router.Group("/webhook/plugin-audit")
	webhookGroup.Use(middleware.WebhookSignatureMiddleware()) // 验证Webhook签名
	{
		// 插件构建完成通知
		webhookGroup.POST("/build-completed", func(c *gin.Context) {
			// TODO: 实现构建完成处理逻辑
			c.JSON(200, gin.H{"status": "received"})
		})

		// 插件安全扫描完成通知
		webhookGroup.POST("/security-scan-completed", func(c *gin.Context) {
			// TODO: 实现安全扫描完成处理逻辑
			c.JSON(200, gin.H{"status": "received"})
		})

		// 插件测试完成通知
		webhookGroup.POST("/test-completed", func(c *gin.Context) {
			// TODO: 实现测试完成处理逻辑
			c.JSON(200, gin.H{"status": "received"})
		})
	}

	logger.Info("Plugin audit webhook routes registered successfully")
}