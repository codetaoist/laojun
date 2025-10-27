package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/codetaoist/laojun-config-center/internal/config"
	"github.com/codetaoist/laojun-config-center/internal/handlers"
	"github.com/codetaoist/laojun-config-center/internal/middleware"
	"github.com/codetaoist/laojun-config-center/internal/storage"
	"github.com/codetaoist/laojun-config-center/internal/storage/file"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
)

// getVersion 返回应用版本信息
func getVersion() string {
	return "1.0.0"
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		// 忽略 .env 文件不存在的错误
	}

	// 初始化全局端口配置
	if err := sharedconfig.InitializeGlobalConfig("../../laojun-shared/config/ports.yaml"); err != nil {
		fmt.Printf("Warning: Failed to load global port config: %v\n", err)
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 使用统一端口配置覆盖服务端口
	if configCenterPort := sharedconfig.GetServicePort("config-center"); configCenterPort != "" {
		if port, err := strconv.Atoi(configCenterPort); err == nil {
			cfg.Server.Port = port
		}
	}

	// 解析环境名称（优先 APP_ENV，其次 ENVIRONMENT，默认 development）
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting config center",
		zap.String("version", getVersion()),
		zap.String("environment", env),
		zap.Int("port", cfg.Server.Port),
	)

	// 设置 Gin 模式
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化存储后端
	var configStorage storage.ConfigStorage
	switch cfg.Storage.Type {
	case "file":
		configStorage, err = file.NewFileStorage(cfg.Storage.File.BasePath)
		if err != nil {
			logger.Fatal("Failed to initialize file storage", zap.Error(err))
		}

	case "redis":
		// TODO: 实现 Redis 存储
		logger.Fatal("Redis storage not implemented yet")
	case "database":
		// TODO: 实现数据库存储
		logger.Fatal("Database storage not implemented yet")
	default:
		logger.Fatal("Unknown storage type", zap.String("type", cfg.Storage.Type))
	}

	defer configStorage.Close()

	// 初始化处理器
	configHandler := handlers.NewConfigHandler(configStorage, logger)
	
	// 创建统一配置管理器
	configManager := sharedconfig.NewDefaultConfigManager(
		sharedconfig.NewMemoryConfigStorage(),
		&sharedconfig.ConfigOptions{
			HistorySize: 100,
		},
	)
	
	// 初始化统一配置处理器
	unifiedHandler := handlers.NewUnifiedConfigHandler(configManager, logger)

	// 创建路由
	router := gin.New()

	// 应用中间件链
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Metrics())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(100)) // 每分钟100个请求
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(middleware.RequestSizeLimit(10 * 1024 * 1024)) // 10MB限制
	router.Use(middleware.Validation())

	// 系统监控
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			logger.Info("System stats",
				zap.Int("goroutines", runtime.NumGoroutine()),
				zap.Uint64("memory_alloc_mb", m.Alloc/1024/1024),
				zap.Uint64("memory_sys_mb", m.Sys/1024/1024),
				zap.Uint32("gc_cycles", m.NumGC),
			)
		}
	}()

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := configStorage.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"version":     getVersion(),
			"environment": env,
			"storage":     cfg.Storage.Type,
		})
	})

	// 指标端点（简单实现）
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# Config Center Metrics\n# TODO: Implement Prometheus metrics\n")
	})

	// API 路由组
	api := router.Group("/api/v1")

	// 安全中间件 (TODO: 实现认证和IP白名单)
	if cfg.Security.EnableAuth {
		logger.Info("Authentication enabled but not implemented yet")
	}
	if len(cfg.Security.AllowedIPs) > 0 {
		logger.Info("IP whitelist configured but not implemented yet")
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
			configs.POST("/:service/:environment/:key/rollback", configHandler.RollbackConfig)
			configs.GET("/:service/:environment/backup", configHandler.BackupConfigs)
			configs.POST("/:service/:environment/restore", configHandler.RestoreConfigs)
			configs.GET("/:service/:environment/watch", configHandler.WatchConfigs)
		}

		// 统一配置管理路由
		unified := api.Group("/unified")
		{
			unified.GET("/configs/:key", unifiedHandler.GetConfig)
			unified.GET("/configs/:key/typed", unifiedHandler.GetConfigWithType)
			unified.PUT("/configs/:key", unifiedHandler.SetConfig)
			unified.DELETE("/configs/:key", unifiedHandler.DeleteConfig)
			unified.GET("/configs", unifiedHandler.ListConfigs)
			unified.POST("/configs/batch", unifiedHandler.BatchSetConfigs)
			unified.GET("/configs/:key/history", unifiedHandler.GetConfigHistory)
			unified.GET("/configs/watch", unifiedHandler.WatchConfigs)
			unified.HEAD("/configs/:key", unifiedHandler.ExistsConfig)
			unified.GET("/health", unifiedHandler.HealthCheck)
		}

		// 批量操作路由
		api.POST("/configs/batch", configHandler.BatchSetConfigs)
		api.DELETE("/configs/batch", configHandler.BatchDeleteConfigs)

		// 搜索路由
		api.GET("/search", configHandler.SearchConfigs)
	}

	// 创建 HTTP 服务
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Config Center started",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", env),
		zap.String("storage_type", cfg.Storage.Type),
		zap.Bool("security_enabled", cfg.Security.EnableAuth),
	)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Config Center...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Config Center stopped")
}
