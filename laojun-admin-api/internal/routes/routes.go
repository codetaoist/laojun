package routes

import (
	"database/sql"

	"github.com/codetaoist/laojun-admin-api/internal/config"
	"github.com/codetaoist/laojun-admin-api/internal/handlers"
	"github.com/codetaoist/laojun-admin-api/internal/middleware"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	sharedauth "github.com/codetaoist/laojun-shared/auth"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func convertToInternalRateLimitConfig(sharedCfg *sharedconfig.RateLimitConfig) *config.RateLimitConfig {
	return &config.RateLimitConfig{
		Enabled:            sharedCfg.Enabled,
		LoginRequests:      sharedCfg.LoginRequests,
		LoginWindowMinutes: sharedCfg.LoginWindowMinutes,
		APIRequests:        sharedCfg.APIRequests,
		APIWindowHours:     sharedCfg.APIWindowHours,
		ErrorMessage:       sharedCfg.ErrorMessage,
		ErrorDetail:        sharedCfg.ErrorDetail,
		RetryAfterSeconds:  sharedCfg.RetryAfterSeconds,
	}
}

func SetupRoutes(
	adminAuthService *services.AdminAuthService,
	userService *services.UserService,
	permissionService *services.PermissionService,
	redisClient *redis.Client,
	cfg *sharedconfig.Config,
	db *sql.DB,
) *gin.Engine {
	r := gin.Default()

	// 创建处理程序
	jwtManager := sharedauth.NewJWTManager(&cfg.JWT)
	authHandler := handlers.NewAuthHandler(adminAuthService, jwtManager, cfg)
	userHandler := handlers.NewUserHandler(userService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)
	
	// 新增角色服务与处理器
	roleService := services.NewRoleService(db, nil)
	roleHandler := handlers.NewRoleHandler(roleService)

	// 创建菜单服务与处理器
	menuService := services.NewMenuService(db)
	menuHandler := handlers.NewMenuHandler(menuService)

	// 新增系统服务与处理器
	systemService := services.NewSystemService(db)
	systemHandler := handlers.NewSystemHandler(systemService)

	// 创建中间件
	authMiddleware := middleware.NewAdminAuthMiddleware(adminAuthService)
	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"}, true)
	permissionMiddleware := middleware.NewPermissionMiddleware(permissionService)

	// 应用CORS中间件
	r.Use(corsMiddleware)

	// API路由组
	api := r.Group("/api/v1")

	// 公开路由（不需要认证）
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"message": "Admin API is running",
			})
		})
	}

	// 受保护的路由（需要认证）
	protected := api.Group("/")
	protected.Use(authMiddleware)
	{
		// 认证相关
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", authHandler.GetProfile)
			auth.PUT("/profile", authHandler.UpdateProfile)
		}

		// 用户管理 - 需要用户管理权限（list, view, create, edit, delete, reset_password, toggle_status）
		users := protected.Group("/users")
		users.Use(permissionMiddleware.RequireExtendedPermission("system", "user", "list"))
		{
			users.GET("/", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("/",
				permissionMiddleware.RequireExtendedPermission("system", "user", "create"),
				userHandler.CreateUser)
			users.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "user", "edit"),
				userHandler.UpdateUser)
			users.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "user", "delete"),
				userHandler.DeleteUser)
			users.POST("/:id/reset-password",
				permissionMiddleware.RequireExtendedPermission("system", "user", "reset_password"),
				userHandler.ResetPassword)
			users.POST("/:id/toggle-status",
				permissionMiddleware.RequireExtendedPermission("system", "user", "toggle_status"),
				userHandler.ToggleUserStatus)
		}

		// 权限管理 - 需要权限管理权限（list, view, create, edit, delete）
		permissions := protected.Group("/permissions")
		permissions.Use(permissionMiddleware.RequireExtendedPermission("system", "permission", "list"))
		{
			permissions.GET("/", permissionHandler.GetPermissions)
			permissions.GET("/:id", permissionHandler.GetPermission)
			permissions.POST("/",
				permissionMiddleware.RequireExtendedPermission("system", "permission", "create"),
				permissionHandler.CreatePermission)
			permissions.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "permission", "edit"),
				permissionHandler.UpdatePermission)
			permissions.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "permission", "delete"),
				permissionHandler.DeletePermission)
		}

		// 角色管理 - 需要角色管理权限（list, view, create, edit, delete, assign_permissions）
		roles := protected.Group("/roles")
		roles.Use(permissionMiddleware.RequireExtendedPermission("system", "role", "list"))
		{
			roles.GET("/", roleHandler.GetRoles)
			roles.GET("/:id", roleHandler.GetRole)
			roles.POST("/",
				permissionMiddleware.RequireExtendedPermission("system", "role", "create"),
				roleHandler.CreateRole)
			roles.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "role", "edit"),
				roleHandler.UpdateRole)
			roles.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "role", "delete"),
				roleHandler.DeleteRole)
			roles.POST("/:id/permissions",
				permissionMiddleware.RequireExtendedPermission("system", "role", "assign_permissions"),
				roleHandler.AssignPermissions)
		}

		// 菜单管理 - 需要菜单管理权限（list, view, create, edit, delete, move, batch_delete, batch_update, toggle_favorite）
		menus := protected.Group("/menus")
		menus.Use(permissionMiddleware.RequireExtendedPermission("system", "menu", "list"))
		{
			menus.GET("/", menuHandler.GetMenus)
			menus.GET("/:id", menuHandler.GetMenu)
			menus.POST("/",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "create"),
				menuHandler.CreateMenu)
			menus.PUT("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "edit"),
				menuHandler.UpdateMenu)
			menus.DELETE("/:id",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "delete"),
				menuHandler.DeleteMenu)
			menus.POST("/:id/move",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "move"),
				menuHandler.MoveMenu)
			menus.DELETE("/batch",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "batch_delete"),
				menuHandler.BatchDeleteMenus)
			menus.PUT("/batch",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "batch_update"),
				menuHandler.BatchUpdateMenus)
			menus.POST("/:id/toggle-favorite",
				permissionMiddleware.RequireExtendedPermission("system", "menu", "toggle_favorite"),
				menuHandler.ToggleFavorite)
		}

		// 系统管理 - 需要系统管理权（view, edit, manage）
		system := protected.Group("/system")
		system.Use(permissionMiddleware.RequireExtendedPermission("system", "system", "view"))
		{
			// 系统信息
			system.GET("/info", systemHandler.GetSystemInfo)
			
			// 系统配置
			system.GET("/config", systemHandler.GetSystemConfig)
			system.PUT("/config",
				permissionMiddleware.RequireExtendedPermission("system", "system", "edit"),
				systemHandler.UpdateSystemConfig)
			
			// 系统状态
			system.GET("/status", systemHandler.GetSystemStatus)
			
			// 性能指标
			system.GET("/metrics", systemHandler.GetMetrics)
		}
	}

	// Swagger文档
	swaggerHandler := handlers.NewSwaggerHandler("/app/docs/api/swagger")
	r.GET("/swagger/*any", swaggerHandler.ServeSwagger)

	return r
}
