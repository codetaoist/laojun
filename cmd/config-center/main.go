package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/codetaoist/laojun/internal/config"
	"github.com/codetaoist/laojun/internal/handlers"
	"github.com/codetaoist/laojun/internal/storage"
	"github.com/codetaoist/laojun/pkg/shared/health"
	"github.com/codetaoist/laojun/pkg/shared/logger"
	"github.com/codetaoist/laojun/pkg/shared/metrics"
	"github.com/codetaoist/laojun/pkg/shared/middleware"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 解析环境名称（优先 APP_ENV，其次 ENVIRONMENT，默认 development）
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}

	// 初始化日志（适配共享 logger.Config 结构）
	filePath := os.Getenv("LOG_FILE")
	if filePath == "" {
		filePath = "./logs/config-center.log"
	}
	logConfig := logger.Config{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		Output:      cfg.Log.Output,
		File:        logger.FileConfig{Filename: filePath, MaxSize: 50, MaxBackups: 7, MaxAge: 28, Compress: true},
		Service:     "config-center",
		Version:     "1.0.0",
		Environment: env,
	}

	log := logger.New(logConfig)

	// 设置全局日志
	// logger.SetDefault(log)

	// 初始化指标收集器
	metricsConfig := metrics.Config{
		Enabled:     true,
		Path:        "/metrics",
		Namespace:   "laojun",
		Subsystem:   "config_center",
		Service:     "config-center",
		Version:     "1.0.0",
		Environment: env,
	}

	metricsInstance := metrics.New(metricsConfig)

	// 初始化健康检查器
	healthConfig := health.Config{
		Enabled:     true,
		Path:        "/health",
		Timeout:     30 * time.Second,
		Service:     "config-center",
		Version:     "1.0.0",
		Environment: env,
	}

	healthChecker := health.New(healthConfig)

	// 设置 Gin 模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化存储后端
	var configStorage storage.ConfigStorage
	switch cfg.Storage.Type {
	case "file":
		configStorage, err = storage.NewFileStorage(cfg.Storage.File.BasePath, cfg.Storage.File.WatchDir)
		if err != nil {
			log.Fatal("Failed to initialize file storage", "error", err)
		}

		// 添加文件存储健康检查器
		healthChecker.AddChecker(health.NewCustomChecker(
			"file_storage",
			func(ctx context.Context) (health.Status, string, error) {
				// 检查存储目录是否可访问
				if _, err := os.Stat(cfg.Storage.File.BasePath); err != nil {
					return health.StatusUnhealthy, "Storage directory not accessible", err
				}
				return health.StatusHealthy, "File storage is accessible", nil
			},
		))

	case "redis":
		// TODO: 实现 Redis 存储
		log.Fatal("Redis storage not implemented yet")
	case "database":
		// TODO: 实现数据库存储
		log.Fatal("Database storage not implemented yet")
	default:
		log.Fatal("Unknown storage type", "type", cfg.Storage.Type)
	}

	defer configStorage.Close()

	// 初始化处理器
	configHandler := handlers.NewConfigHandler(configStorage)

	// 创建路由
	router := gin.New()

	// 应用中间件链
	middlewareChain := middleware.NewMiddlewareChain().
		Use(middleware.RequestID()).
		Use(middleware.Recovery()).
		Use(middleware.SecurityHeaders()).
		Use(middleware.Logger(middleware.LoggerConfig{
			SkipPaths: []string{"/health", "/metrics"},
			LogBody:   false,
		})).
		Use(metrics.GinMiddleware(metricsInstance)).
		Use(middleware.CORS(middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: false,
			MaxAge:           86400,
		})).
		Use(middleware.RateLimit(middleware.RateLimitConfig{
			Enabled: true,
			RPS:     100,
			Burst:   200,
		})).
		Use(middleware.Timeout(30 * time.Second)).
		Use(middleware.RequestSizeLimit(10 * 1024 * 1024)) // 10MB

	middlewareChain.Apply(router)

	// 系统监控
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			metricsInstance.RecordMemoryUsage(float64(m.Alloc))
			metricsInstance.RecordGoroutineCount(runtime.NumGoroutine())
		}
	}()

	// 健康检查端点
	router.GET("/health", healthChecker.GinHandler())
	router.GET("/healthz", healthChecker.GinHandler())

	// 监控端点
	router.GET("/metrics", gin.WrapH(metricsInstance.Handler()))

	// API 路由组
	api := router.Group("/api/v1")

	// 如果启用了安全认证
	if cfg.Security.EnableAuth {
		if cfg.Security.APIKey != "" {
			api.Use(middleware.APIKeyAuth([]string{cfg.Security.APIKey}))
		}
		if len(cfg.Security.AllowedIPs) > 0 {
			api.Use(middleware.IPWhitelist(cfg.Security.AllowedIPs))
		}
	}

	{
		// 配置管理路由
		configs := api.Group("/configs")
		{
			configs.GET("/:service/:environment", configHandler.ListConfigs)
			configs.GET("/:service/:environment/:key", configHandler.GetConfig)
			configs.PUT("/:service/:environment/:key", configHandler.SetConfig)
			configs.DELETE("/:service/:environment/:key", configHandler.DeleteConfig)
			configs.GET("/:service/:environment/:key/history", configHandler.GetConfigHistory)
			configs.GET("/:service/:environment/backup", configHandler.BackupConfigs)
			configs.POST("/:service/:environment/restore", configHandler.RestoreConfigs)
			configs.GET("/:service/:environment/watch", configHandler.WatchConfigs)
		}

		// 搜索路由
		api.POST("/search", configHandler.SearchConfigs)
	}

	// 创建 HTTP 服务
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	log.Info("Config Center started",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"environment", env,
		"storage_type", cfg.Storage.Type,
		"security_enabled", cfg.Security.EnableAuth,
	)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Config Center...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Config Center stopped")
}
