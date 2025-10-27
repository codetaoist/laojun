package routes

import (
	"github.com/codetaoist/laojun-marketplace-api/internal/handlers"
	"github.com/codetaoist/laojun-marketplace-api/internal/middleware"
	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/gin-gonic/gin"
)

// SetupReviewModerationRoutes 设置评价审核路由
func SetupReviewModerationRoutes(router *gin.Engine, db *shareddb.DB) {
	// 创建审核处理器
	moderationHandler := handlers.NewReviewModerationHandler(db)

	// API v1 路由组
	v1 := router.Group("/api/v1")

	// 用户路由 - 需要认证
	userRoutes := v1.Group("/reviews")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		// 举报评价
		userRoutes.POST("/flag", moderationHandler.FlagReview)
	}

	// 管理员路由 - 需要认证和管理员权限
	adminRoutes := v1.Group("/admin/reviews")
	adminRoutes.Use(middleware.AuthMiddleware())
	adminRoutes.Use(middleware.AdminMiddleware()) // 需要管理员或审核员权限
	{
		// 审核管理
		adminRoutes.POST("/moderate", moderationHandler.ModerateReview)
		
		// 获取待审核评价列表
		adminRoutes.GET("/pending", moderationHandler.GetPendingReviews)
		
		// 获取被举报的评价列表
		adminRoutes.GET("/flagged", moderationHandler.GetFlaggedReviews)
		
		// 处理举报的评价
		adminRoutes.POST("/flags/:flag_id/resolve", moderationHandler.ResolveFlaggedReview)
		
		// 获取审核统计
		adminRoutes.GET("/stats", moderationHandler.GetModerationStats)
		
		// 获取评价审核历史
		adminRoutes.GET("/:review_id/history", moderationHandler.GetModerationHistory)
	}
}