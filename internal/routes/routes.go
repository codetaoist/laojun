package routes

import (
	"database/sql"

	sharedconfig "github.com/codetaoist/laojun/pkg/shared/config"
	"github.com/codetaoist/laojun/internal/database"
	"github.com/codetaoist/laojun/internal/handlers"
	"github.com/codetaoist/laojun/internal/middleware"
	"github.com/codetaoist/laojun/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	sharedauth "github.com/codetaoist/laojun/pkg/shared/auth"
	shareddb "github.com/codetaoist/laojun/pkg/shared/database"
)

func SetupRoutes(
	authService *services.AuthService,
	userService *services.UserService,
	permissionService *services.PermissionService,
	pluginService *services.PluginService,
	redisClient *redis.Client,
	cfg *sharedconfig.Config,
	db *sql.DB,
) *gin.Engine {
	r := gin.Default()

	// 创建处理程序
	jwtManager := sharedauth.NewJWTManager(&cfg.JWT)
	authHandler := handlers.NewAuthHandler(authService, jwtManager, cfg)
	userHandler := handlers.NewUserHandler(userService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	pluginHandler := handlers.NewPluginHandler(pluginService)
	// 分类与评论处理器（使用共享数据库包装器实例化服务）
	sharedDB, _ := shareddb.NewDB(&cfg.Database)
	categoryService := services.NewCategoryService(sharedDB)
	reviewService := services.NewReviewService(sharedDB)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	rateLimitHandler := handlers.NewRateLimitHandler(redisClient)
	// 新增角色服务与处理器
	roleService := services.NewRoleService(db, nil)
	roleHandler := handlers.NewRoleHandler(roleService)

	// 创建菜单服务与处理器
	menuService := services.NewMenuService(db)
	menuHandler := handlers.NewMenuHandler(menuService)

	// 创建图标服务与处理器
	iconService := services.NewIconService(db)
	iconHandler := handlers.NewIconHandler(iconService)

	// 创建索引管理器和处理程序
	indexManager := database.NewIndexManager(db)
	indexHandler := handlers.NewIndexHandler(indexManager)

	// 新增系统服务与处理器
	systemService := services.NewSystemService(db)
	systemHandler := handlers.NewSystemHandler(systemService)

	// 创建中间件
	authMiddleware := middleware.NewAuthMiddleware(authService)
	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"}, true)
	permissionMiddleware := middleware.NewPermissionMiddleware(permissionService)

	// 应用CORS中间件
	r.Use(corsMiddleware)

	// 应用频率限制中间件（如果启用）
	if cfg.RateLimit.Enabled && redisClient != nil {
		// 全局频率限制
		r.Use(middleware.GlobalRateLimit(redisClient, cfg.RateLimit.GlobalRequests, cfg.RateLimit.GlobalWindow))
	}

	// API路由组
	api := r.Group("/api/v1")

	// 公开路由
	public := api.Group("/")
	{
		// 登录端点使用特殊的频率限制（如果启用）
		if cfg.RateLimit.Enabled && redisClient != nil {
			public.POST("/login",
				middleware.LoginRateLimit(redisClient),
				authHandler.Login)
		} else {
			public.POST("/login", authHandler.Login)
		}

		// 验证码端点按开关启用
		if cfg.Security.EnableCaptcha {
			public.GET("/auth/captcha", authHandler.GetCaptcha)
			// 仅在调试模式暴露验证码明文接口
			if cfg.Server.Mode == "debug" {
				public.GET("/auth/captcha/code", authHandler.GetCaptchaCodeDebug)
			}
		}

		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// 系统信息 - 公开访问
		public.GET("/system/info", systemHandler.GetSystemInfo)
	}

	// 需要认证的路由
	protected := api.Group("/")
	protected.Use(authMiddleware)

	// 为各子路由组准备用户级限流（不作用于 /auth 组）
	var groupUserRateLimiter gin.HandlerFunc
	if cfg.RateLimit.Enabled && redisClient != nil {
		groupUserRateLimiter = middleware.UserRateLimit(redisClient, cfg.RateLimit.UserRequests, cfg.RateLimit.UserWindow)
	}

	{
		// 认证相关（不使用用户级限流）
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", authHandler.GetProfile)
		}

		// 用户信息路由别名（兼容前端调用）
		user := protected.Group("/user")
		{
			user.GET("/profile", authHandler.GetProfile)
		}

		// 权限检查和同步 - 所有认证用户都可以访问
		permissions := protected.Group("/permissions")
		if groupUserRateLimiter != nil {
			permissions.Use(groupUserRateLimiter)
		}
		{
			permissions.POST("/check", permissionHandler.CheckUserPermission)
			permissions.GET("/check", permissionHandler.CheckCurrentUserPermission)
			permissions.POST("/sync", permissionHandler.SyncUserPermissions)
			permissions.GET("/sync", permissionHandler.SyncCurrentUserPermissions)
			permissions.DELETE("/cache", permissionHandler.InvalidateCurrentUserCache)
			// 权限列表（供前端抽屉使用）
			permissions.GET("/", permissionHandler.GetPermissions)

			// 基础权限CRUD
			permissions.GET("/:id", permissionHandler.GetPermissionByID)
			permissions.POST("/",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "create"),
				permissionHandler.CreatePermission)
			permissions.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "edit"),
				permissionHandler.UpdatePermission)
			permissions.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "delete"),
				permissionHandler.DeletePermission)
			permissions.POST("/batch-delete",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "delete"),
				permissionHandler.BatchDeletePermissions)

			// 资源与操作列表（供前端抽屉使用）
			permissions.GET("/resources",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "list"),
				permissionHandler.GetResources)
			permissions.GET("/actions",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "list"),
				permissionHandler.GetActions)

			// 权限使用
			permissions.GET("/:id/usage",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "view"),
				permissionHandler.CheckPermissionUsage)

			// 权限统计
			permissions.GET("/stats", permissionHandler.GetPermissionStats)

			// 导入导出
			permissions.GET("/export",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "view"),
				permissionHandler.ExportPermissions)
			permissions.POST("/import",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "manage"),
				permissionHandler.ImportPermissions)

			// 系统权限同步
			permissions.POST("/sync-system",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "manage"),
				permissionHandler.SyncSystemPermissions)

			// 权限模板管理
			permissions.GET("/templates", permissionHandler.GetPermissionTemplates)
			permissions.POST("/templates",
				permissionMiddleware.RequireExtendedPermission("permission", "template", "create"),
				permissionHandler.CreatePermissionTemplate)
			permissions.POST("/templates/:id/apply",
				permissionMiddleware.RequireExtendedPermission("permission", "template", "apply"),
				permissionHandler.ApplyPermissionTemplate)

			// 扩展权限管理
			permissions.GET("/extended", permissionHandler.GetExtendedPermissions)
			permissions.POST("/extended",
				permissionMiddleware.RequireExtendedPermission("permission", "permission", "create"),
				permissionHandler.CreateExtendedPermission)

			// 设备类型管理
			permissions.GET("/device-types", permissionHandler.GetDeviceTypes)
			permissions.POST("/device-types",
				permissionMiddleware.RequireExtendedPermission("permission", "device_type", "create"),
				permissionHandler.CreateDeviceType)

			// 模块管理
			permissions.GET("/modules", permissionHandler.GetModules)
			permissions.POST("/modules",
				permissionMiddleware.RequireExtendedPermission("permission", "module", "create"),
				permissionHandler.CreateModule)

			// 用户组管理
			permissions.GET("/user-groups", permissionHandler.GetUserGroups)
			permissions.POST("/user-groups",
				permissionMiddleware.RequireExtendedPermission("permission", "user_group", "create"),
				permissionHandler.CreateUserGroup)
			permissions.POST("/user-groups/:id/members",
				permissionMiddleware.RequireExtendedPermission("permission", "user_group", "manage"),
				permissionHandler.AddUsersToGroup)
		}

		// 兼容路由（前端调用路径），与 /permissions/templates 等价
		if groupUserRateLimiter != nil {
			protected.GET("/permission-templates", groupUserRateLimiter, permissionHandler.GetPermissionTemplates)
		} else {
			protected.GET("/permission-templates", permissionHandler.GetPermissionTemplates)
		}

		// 用户管理 - 需要用户管理权（list, view, create, edit, delete, reset_password）
		users := protected.Group("/users")
		users.Use(permissionMiddleware.RequireExtendedPermission("user", "user", "list"))
		if groupUserRateLimiter != nil {
			users.Use(groupUserRateLimiter)
		}
		{
			users.GET("/", userHandler.GetUsers)
			users.GET("/:id",
				permissionMiddleware.RequireExtendedPermission("user", "user", "view"),
				userHandler.GetUser)
			users.POST("/",
				permissionMiddleware.RequireExtendedPermission("user", "user", "create"),
				userHandler.CreateUser)
			users.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("user", "user", "edit"),
				userHandler.UpdateUser)
			users.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("user", "user", "delete"),
				userHandler.DeleteUser)
			users.POST("/:id/change-password",
				permissionMiddleware.RequireExtendedPermission("user", "user", "reset_password"),
				userHandler.ChangePassword)
			users.POST("/:id/reset-password",
				permissionMiddleware.RequireExtendedPermission("user", "user", "reset_password"),
				userHandler.ResetPassword)
			// 分配角色到用户（需要角色管理权）
			users.POST("/:id/roles",
				permissionMiddleware.RequireExtendedPermission("permission", "role", "assign"),
				roleHandler.AssignRolesToUser)
		}

		// 角色管理 - 需要角色管理权（list, view, create, edit, delete, assign）
		roles := protected.Group("/roles")
		roles.Use(permissionMiddleware.RequireExtendedPermission("permission", "role", "list"))
		if groupUserRateLimiter != nil {
			roles.Use(groupUserRateLimiter)
		}
		{
			roles.GET("/", roleHandler.GetRoles)
			roles.GET("/:id",
				permissionMiddleware.RequireExtendedPermission("permission", "role", "view"),
				roleHandler.GetRole)
			roles.POST("/",
				permissionMiddleware.RequireExtendedPermission("permission", "role", "create"),
				roleHandler.CreateRole)
			roles.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("permission", "role", "edit"),
				roleHandler.UpdateRole)
			roles.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("permission", "role", "delete"),
				roleHandler.DeleteRole)

			// 角色权限相关
			roles.GET("/:id/permissions", roleHandler.GetRolePermissions)
			roles.POST("/:id/permissions", roleHandler.AssignPermissionsToRole)
		}

		// 权限管理 - 需要权限管理权（list, view, create, edit, delete, assign）
		permissionMgmt := protected.Group("/permission-management")
		permissionMgmt.Use(permissionMiddleware.RequireExtendedPermission("permission", "permission", "list"))
		if groupUserRateLimiter != nil {
			permissionMgmt.Use(groupUserRateLimiter)
		}
		{
			// 设备类型管理
			deviceTypes := permissionMgmt.Group("/device-types")
			{
				deviceTypes.GET("/", permissionHandler.GetDeviceTypes)
				deviceTypes.POST("/",
					permissionMiddleware.RequireExtendedPermission("system", "device_type", "create"),
					permissionHandler.CreateDeviceType)
			}

			// 模块管理
			modules := permissionMgmt.Group("/modules")
			{
				modules.GET("/", permissionHandler.GetModules)
				modules.POST("/",
					permissionMiddleware.RequireExtendedPermission("system", "module", "create"),
					permissionHandler.CreateModule)
			}

			// 用户组管理
			userGroups := permissionMgmt.Group("/user-groups")
			{
				userGroups.GET("/",
					permissionMiddleware.RequireExtendedPermission("permission", "user_group", "list"),
					permissionHandler.GetUserGroups)
				userGroups.POST("/",
					permissionMiddleware.RequireExtendedPermission("permission", "user_group", "create"),
					permissionHandler.CreateUserGroup)
				userGroups.POST("/members",
					permissionMiddleware.RequireExtendedPermission("permission", "user_group", "edit"),
					permissionHandler.AddUsersToGroup)
			}

			// 扩展权限管理
			extendedPerms := permissionMgmt.Group("/extended-permissions")
			{
				extendedPerms.GET("/",
					permissionMiddleware.RequireExtendedPermission("permission", "permission", "list"),
					permissionHandler.GetExtendedPermissions)
				extendedPerms.POST("/",
					permissionMiddleware.RequireExtendedPermission("permission", "permission", "create"),
					permissionHandler.CreateExtendedPermission)
			}

			// 权限模板管理
			templates := permissionMgmt.Group("/templates")
			{
				templates.GET("/",
					permissionMiddleware.RequireExtendedPermission("permission", "permission", "list"),
					permissionHandler.GetPermissionTemplates)
				templates.POST("/",
					permissionMiddleware.RequireExtendedPermission("permission", "permission", "create"),
					permissionHandler.CreatePermissionTemplate)
				templates.POST("/:id/apply",
					permissionMiddleware.RequireExtendedPermission("permission", "permission", "assign"),
					permissionHandler.ApplyPermissionTemplate)
			}

			// 缓存管理
			cache := permissionMgmt.Group("/cache")
			{
				cache.GET("/stats",
					permissionMiddleware.RequireExtendedPermission("system", "cache", "view"),
					permissionHandler.GetCacheStats)
				cache.DELETE("/users/:user_id",
					permissionMiddleware.RequireExtendedPermission("system", "cache", "clear"),
					permissionHandler.InvalidateUserCache)
				cache.POST("/users/:user_id/warmup",
					permissionMiddleware.RequireExtendedPermission("system", "cache", "manage"),
					permissionHandler.WarmupUserCache)
			}
		}

		// 菜单管理 - 需要菜单管理权（list, view, create, edit, delete, move, batch_delete, batch_update, toggle_favorite）
		menus := protected.Group("/menus")
		menus.Use(permissionMiddleware.RequireExtendedPermission("system", "menu", "list"))
		if groupUserRateLimiter != nil {
			menus.Use(groupUserRateLimiter)
		}
		{
			menus.GET("/", menuHandler.GetMenus)
			menus.GET("/:id", menuHandler.GetMenuByID)
			menus.POST("/",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "create"),
				menuHandler.CreateMenu)
			menus.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "edit"),
				menuHandler.UpdateMenu)
			menus.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "delete"),
				menuHandler.DeleteMenu)
			menus.DELETE("/batch",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "delete"),
				menuHandler.BatchDeleteMenus)
			menus.POST("/:id/move",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "edit"),
				menuHandler.MoveMenu)
			menus.GET("/stats",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "view"),
				menuHandler.GetMenuStats)
			menus.PUT("/batch",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "edit"),
				menuHandler.BatchUpdateMenus)
			menus.POST(":id/toggle-favorite", permissionMiddleware.RequireExtendedPermission("system", "menu", "update"), menuHandler.ToggleFavorite)
		}

		// 图标管理 - 需要图标管理权（list, view, create, edit, delete, move, batch_delete, batch_update, toggle_favorite）
		icons := protected.Group("/icons")
		icons.Use(permissionMiddleware.RequireExtendedPermission("system", "icon", "list"))
		if groupUserRateLimiter != nil {
			icons.Use(groupUserRateLimiter)
		}
		{
			icons.GET("/", iconHandler.GetIcons)
			icons.GET("/:id", iconHandler.GetIcon)
			icons.POST("/",
				permissionMiddleware.RequireExtendedPermission("system", "icon", "create"),
				iconHandler.CreateIcon)
			icons.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "icon", "edit"),
				iconHandler.UpdateIcon)
			icons.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "icon", "delete"),
				iconHandler.DeleteIcon)
			icons.GET("/stats",
				permissionMiddleware.RequireExtendedPermission("system", "icon", "view"),
				iconHandler.GetIconStats)
			icons.GET("/categories", iconHandler.GetIconCategories)
		}

		// 系统管理 - 需要系统管理权（view, edit, manage）
		system := protected.Group("/system")
		system.Use(permissionMiddleware.RequireExtendedPermission("system", "system_config", "view"))
		if groupUserRateLimiter != nil {
			system.Use(groupUserRateLimiter)
		}
		{
			// 配置相关
			system.GET("/configs", systemHandler.GetConfigs)
			system.POST("/configs",
				permissionMiddleware.RequireExtendedPermission("system", "system_config", "edit"),
				systemHandler.SaveConfigs)

			// 审计日志
			system.GET("/logs",
				permissionMiddleware.RequireExtendedPermission("system", "logs", "view"),
				systemHandler.GetLogs)
			system.DELETE("/logs",
				permissionMiddleware.RequireExtendedPermission("system", "system_config", "edit"),
				systemHandler.ClearLogs)

			// 性能指标
			system.GET("/metrics", systemHandler.GetMetrics)
		}

		// 频率限制管理 - 需要系统管理权（view, manage）
		rateLimit := protected.Group("/rate-limit")
		rateLimit.Use(permissionMiddleware.RequireExtendedPermission("system", "rate_limit", "view"))
		if groupUserRateLimiter != nil {
			rateLimit.Use(groupUserRateLimiter)
		}
		{
			rateLimit.GET("/stats",
				permissionMiddleware.RequireExtendedPermission("system", "rate_limit", "view"),
				rateLimitHandler.GetStats)
			rateLimit.POST("/reset/:key",
				permissionMiddleware.RequireExtendedPermission("system", "rate_limit", "manage"),
				rateLimitHandler.ResetLimit)
			rateLimit.GET("/blocked",
				permissionMiddleware.RequireExtendedPermission("system", "rate_limit", "view"),
				rateLimitHandler.GetBlockedIPs)
			rateLimit.DELETE("/blocked/:ip",
				permissionMiddleware.RequireExtendedPermission("system", "rate_limit", "manage"),
				rateLimitHandler.UnblockIP)
		}

		// 索引管理 - 需要系统管理权（manage, view）
		indexes := protected.Group("/indexes")
		if groupUserRateLimiter != nil {
			indexes.Use(groupUserRateLimiter)
		}
		{
			indexes.POST("/apply",
				permissionMiddleware.RequireExtendedPermission("system", "database", "manage"),
				indexHandler.ApplyCompositeIndexes)
			indexes.GET("/info",
				permissionMiddleware.RequireExtendedPermission("system", "database", "view"),
				indexHandler.GetIndexInfo)
			indexes.GET("/stats",
				permissionMiddleware.RequireExtendedPermission("system", "database", "view"),
				indexHandler.GetIndexStats)
			indexes.GET("/analyze",
				permissionMiddleware.RequireExtendedPermission("system", "database", "view"),
				indexHandler.AnalyzeIndexUsage)
			indexes.DELETE("/cleanup",
				permissionMiddleware.RequireExtendedPermission("system", "database", "manage"),
				indexHandler.DropUnusedIndexes)
			indexes.POST("/reindex/:table_name",
				permissionMiddleware.RequireExtendedPermission("system", "database", "manage"),
				indexHandler.ReindexTable)
			indexes.POST("/update-stats",
				permissionMiddleware.RequireExtendedPermission("system", "database", "manage"),
				indexHandler.UpdateIndexStatistics)
			indexes.GET("/recommendations",
				permissionMiddleware.RequireExtendedPermission("system", "database", "view"),
				indexHandler.GetIndexRecommendations)
		}

		// 多端权限示例 - 移动端专用API
		mobile := protected.Group("/mobile")
		mobile.Use(permissionMiddleware.RequireDeviceTypePermission([]string{"mobile"}, "user", "user", "list"))
		if groupUserRateLimiter != nil {
			mobile.Use(groupUserRateLimiter)
		}
		{
			mobile.GET("/users", userHandler.GetUsers)
		}

		// 物联网设备专用API
		iot := protected.Group("/iot")
		iot.Use(permissionMiddleware.RequireDeviceTypePermission([]string{"iot"}, "system", "device_status", "report"))
		if groupUserRateLimiter != nil {
			iot.Use(groupUserRateLimiter)
		}
		{
			iot.POST("/status", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "设备状态上报成功"})
			})
		}
	}

	// Marketplace API路由 - 公开访问
	marketplace := r.Group("/api/marketplace")
	{
		// 插件相关路由
		marketplace.GET("/plugins", pluginHandler.GetPlugins)
		marketplace.GET("/plugins/:id", pluginHandler.GetPlugin)
		marketplace.GET("/plugins/:id/reviews", reviewHandler.GetPluginReviews)

		// 分类相关路由
		marketplace.GET("/plugins/categories", categoryHandler.GetCategories)
	}

	// Marketplace API路由 - v1版本
	marketplaceV1 := api.Group("/marketplace")
	{
		// 插件相关
		marketplaceV1.GET("/plugins", pluginHandler.GetPlugins)
		marketplaceV1.GET("/plugins/:id", pluginHandler.GetPlugin)
		marketplaceV1.GET("/plugins/:id/reviews", reviewHandler.GetPluginReviews)

		// 分类相关（统一至 /marketplace/categories）
		marketplaceV1.GET("/categories", categoryHandler.GetCategories)
	}

	// Marketplace API路由 - 兼容别名（前端 `/plugins/marketplace`）
	pluginsMarketplaceAlias := api.Group("/plugins/marketplace")
	{
		pluginsMarketplaceAlias.GET("/", pluginHandler.GetPlugins)
		pluginsMarketplaceAlias.GET("/:id", pluginHandler.GetPlugin)
		pluginsMarketplaceAlias.GET("/:id/reviews", reviewHandler.GetPluginReviews)
	}

	// 插件API路由别名（兼容前端调用）
	plugins := api.Group("/plugins")
	{
		plugins.GET("/", pluginHandler.GetPlugins)
		plugins.GET("/:id", pluginHandler.GetPlugin)
		plugins.GET("/:id/reviews", reviewHandler.GetPluginReviews)
	}

	// 分类API路由别名（兼容前端调用）
	categories := api.Group("/categories")
	{
		categories.GET("/", categoryHandler.GetCategories)
	}

	return r
}
