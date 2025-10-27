package server

import (
	"os"

	"github.com/codetaoist/laojun-admin-api/internal/database"
	"github.com/codetaoist/laojun-admin-api/internal/handlers"
	"github.com/codetaoist/laojun-admin-api/internal/middleware"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/codetaoist/laojun-shared/auth"
	"github.com/codetaoist/laojun-shared/config"
	shareddb "github.com/codetaoist/laojun-shared/database"
	sharedmw "github.com/codetaoist/laojun-shared/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Server 服务器结构体
type Server struct {
	router     *gin.Engine
	db         *shareddb.DB
	jwtManager *auth.JWTManager
	config     *config.Config
}

// NewServer 创建新的服务器实例
func NewServer(cfg *config.Config) (*Server, error) {
	// 创建数据库连接
	db, err := shareddb.NewDB(&cfg.Database)
	if err != nil {
		return nil, err
	}

	// 执行数据库迁移（可通过环境变量禁用）
	if os.Getenv("DISABLE_MIGRATIONS") != "true" {
		migrator := database.NewMigrator(db.DB)
		if err := migrator.RunMigrations(); err != nil {
			return nil, err
		}
	}

	// 创建JWT管理器
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// 创建logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 创建服务层
	pluginService := services.NewPluginService(db)
	categoryService := services.NewCategoryService(db)
	reviewService := services.NewReviewService(db)
	developerService := services.NewDeveloperService(db)
	authService := services.NewAuthService(db)
	communityService := services.NewCommunityService(db)

	// 创建扩展插件系统组件（暂时注释掉未使用的部分）
	// pluginLoaderManager := plugin.NewPluginLoaderManager(logger)
	// microserviceManager, err := plugin.NewMicroservicePluginManager(logger)
	// if err != nil {
	//	return nil, err
	// }
	// pluginGateway := plugin.NewPluginGateway(pluginLoaderManager, microserviceManager, logger)
	// extendedPluginService := services.NewExtendedPluginService(db, pluginLoaderManager, microserviceManager, pluginGateway, logger)

	// 创建处理器层
	pluginHandler := handlers.NewPluginHandler(pluginService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	developerHandler := handlers.NewDeveloperHandler(developerService)
	authHandler := handlers.NewMarketplaceAuthHandler(authService, jwtManager, cfg)
	communityHandler := handlers.NewCommunityHandler(communityService)

	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 初始化Redis客户端（用于限流）
	var redisClient *redis.Client
	if cfg.RateLimit.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.GetRedisAddr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}

	// 添加中间件（按顺序执行）
	router.Use(sharedmw.Logger(sharedmw.LoggerConfig{
		SkipPaths: []string{"/health", "/healthz", "/metrics", "/api/v1/health"},
		LogBody:   false,
	}))
	router.Use(gin.Recovery())
	router.Use(middleware.NewCORSMiddleware([]string{"*"}, true))
	if cfg.RateLimit.Enabled && redisClient != nil {
		router.Use(middleware.GlobalRateLimit(redisClient, cfg.RateLimit.GlobalRequests, cfg.RateLimit.GlobalWindow))
	}

	// 健康检查路由（公开访问）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "marketplace-api",
		})
	})

	// API路由组（统一前缀 /api/v1）
	api := router.Group("/api/v1")
	{
		// 健康检查路由（公开访问）
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "marketplace-api",
				"prefix":  "/api/v1",
			})
		})

		// 认证相关路由（无需认证）
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)

			// 验证码配置接口（公开访问）
			authRoutes.GET("/captcha/config", authHandler.GetCaptchaConfig)

			// 验证码端点（按配置启用）
			if cfg.Security.EnableCaptcha {
				authRoutes.GET("/captcha", authHandler.GetCaptcha)
				// 仅在调试模式暴露验证码明文接口（注意：生产环境不建议开启）
				if cfg.Server.Mode == "debug" {
					authRoutes.GET("/captcha/code", authHandler.GetCaptchaCodeDebug)
				}
			}
		}

		// 插件相关路由
		plugins := api.Group("/plugins")
		{
			plugins.GET("", pluginHandler.GetPlugins)
			plugins.GET("/:id", pluginHandler.GetPlugin)
			plugins.GET("/:id/reviews", reviewHandler.GetPluginReviews)
		}

		// 分类相关路由
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.GET("/:id/plugins", pluginHandler.GetPluginsByCategory)
		}

		// 开发者相关路由（公开访问）
		developers := api.Group("/developers")
		{
			developers.GET("", developerHandler.GetDevelopers)
			developers.GET("/:id", developerHandler.GetDeveloper)
			developers.GET("/:id/plugins", developerHandler.GetDeveloperPlugins)
		}

		// 社区相关路由（公开访问）
		community := api.Group("/community")
		{
			// 社区统计信息
			community.GET("/stats", communityHandler.GetCommunityStats)

			// 论坛分类
			community.GET("/forum/categories", communityHandler.GetForumCategories)
			// 论坛帖子
			community.GET("/forum/posts", communityHandler.GetForumPosts)
			community.GET("/forum/posts/:id", communityHandler.GetForumPost)
			community.GET("/forum/posts/:id/replies", communityHandler.GetForumReplies)

			// 博客分类
			community.GET("/blog/categories", communityHandler.GetBlogCategories)
			// 博客文章
			community.GET("/blog/posts", communityHandler.GetBlogPosts)
			community.GET("/blog/posts/:id", communityHandler.GetBlogPost)

			// 代码片段
			community.GET("/code/snippets", communityHandler.GetCodeSnippets)
		}

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware(authService))
		{
			// 用户资料相关路由
			userRoutes := auth.Group("/user")
			{
				userRoutes.GET("/profile", authHandler.GetProfile)
				userRoutes.PUT("/profile", authHandler.UpdateProfile)
				userRoutes.POST("/change-password", authHandler.ChangePassword)
				userRoutes.GET("/stats", authHandler.GetUserStats)
				userRoutes.POST("/logout", authHandler.Logout)
			}

			// 用户相关的插件操作（如收藏、购买等）
			auth.POST("/plugins/:id/favorite", pluginHandler.ToggleFavorite)
			auth.POST("/plugins/:id/purchase", pluginHandler.PurchasePlugin)
			auth.GET("/user/favorites", pluginHandler.GetUserFavorites)
			auth.GET("/user/purchases", pluginHandler.GetUserPurchases)

			// 评论相关（需要登录）
			auth.POST("/plugins/:id/reviews", reviewHandler.CreateReview)
			auth.PUT("/reviews/:id", reviewHandler.UpdateReview)
			auth.DELETE("/reviews/:id", reviewHandler.DeleteReview)

			// 开发者相关（需要登录）
			auth.POST("/developers/register", developerHandler.RegisterDeveloper)
			auth.GET("/developers/me", developerHandler.GetMyDeveloperProfile)
			auth.PUT("/developers/me", developerHandler.UpdateDeveloper)
			auth.GET("/developers/me/stats", developerHandler.GetDeveloperStats)
			auth.GET("/developers/me/plugins", developerHandler.GetMyPlugins)

			// 分类管理（需要登录）
			auth.POST("/categories", categoryHandler.CreateCategory)
			auth.PUT("/categories/:id", categoryHandler.UpdateCategory)
			auth.DELETE("/categories/:id", categoryHandler.DeleteCategory)

			// 社区功能（需要登录）
			communityAuth := auth.Group("/community")
			{
				// 论坛帖子管理
				communityAuth.POST("/forum/posts", communityHandler.CreateForumPost)
				communityAuth.POST("/forum/posts/:id/replies", communityHandler.CreateForumReply)

				// 博客文章管理
				communityAuth.POST("/blog/posts", communityHandler.CreateBlogPost)

				// 代码片段管理
				communityAuth.POST("/code/snippets", communityHandler.CreateCodeSnippet)

				// 通用功能
				communityAuth.POST("/like", communityHandler.ToggleLike)
			}
		}

		// Marketplace 路径别名以兼容前端调用
		marketplace := api.Group("/marketplace")
		{
			// 列表与分类映射到现有插件与分类接口
			marketplace.GET("/plugins", pluginHandler.GetPlugins)
			marketplace.GET("/plugins/search", pluginHandler.GetPlugins)
			marketplace.GET("/plugins/:id", pluginHandler.GetPlugin)
			marketplace.GET("/categories", categoryHandler.GetCategories)

			// 统计信息（简化聚合）
			marketplace.GET("/plugins/stats", pluginHandler.GetMarketplacePluginStats)

			// 需要认证的 marketplace 子路由（收藏等）
			marketplaceAuth := marketplace.Group("")
			marketplaceAuth.Use(middleware.AuthMiddleware(authService))
			{
				marketplaceAuth.GET("/plugins/favorites", pluginHandler.GetUserFavorites)
			}
		}

	}

	return &Server{
		router:     router,
		db:         db,
		jwtManager: jwtManager,
		config:     cfg,
	}, nil
}

// Start 启动服务
func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

// GetRouter 获取路由器实例
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Close 关闭服务
func (s *Server) Close() error {
	return s.db.Close()
}
