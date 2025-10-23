package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"io"
	"path/filepath"
	"strings"

	sharedconfig "github.com/codetaoist/laojun/pkg/shared/config"
	shareddb "github.com/codetaoist/laojun/pkg/shared/database"
	"github.com/codetaoist/laojun/internal/cache"
	"github.com/codetaoist/laojun/internal/config"
	"github.com/codetaoist/laojun/internal/database"
	"github.com/codetaoist/laojun/internal/routes"
	"github.com/codetaoist/laojun/internal/services"
	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"

	"github.com/codetaoist/laojun/pkg/shared/health"
	"github.com/codetaoist/laojun/pkg/shared/logger"
	"github.com/codetaoist/laojun/pkg/shared/metrics"
	"github.com/codetaoist/laojun/pkg/shared/middleware"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// 统一从仓库根目录加载 .env（一次性）
	sharedconfig.LoadDotenv()

	// 加载配置
	cfg, err := sharedconfig.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logConfig := logger.Config{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		Output:      cfg.Log.Output,
		File:        logger.FileConfig{Filename: cfg.Log.File},
		Service:     "admin-api",
		Version:     "1.0.0",
		Environment: cfg.Server.Mode,
	}

	log := logger.New(logConfig)

	// 将 Gin 的默认日志输出路由到文件或双写
	// 计算日志文件的绝对路径并确保目录存在
	{
		filePath := cfg.Log.File
		if filePath == "" {
			filePath = "./logs/admin-api.log"
		}
		absPath, _ := filepath.Abs(filePath)
		logDir := filepath.Dir(absPath)
		_ = os.MkdirAll(logDir, 0o755)

		lumber := &lumberjack.Logger{
			Filename:   absPath,
			MaxSize:    50,
			MaxBackups: 7,
			MaxAge:     28,
			Compress:   true,
		}
		
		out := strings.ToLower(cfg.Log.Output)
		switch out {
		case "file":
			gin.DefaultWriter = lumber
		case "both":
			gin.DefaultWriter = io.MultiWriter(gin.DefaultWriter, lumber)
		default:
			// 默认保持 stdout
		}
	}

	// 初始化指标收集
	metricsConfig := metrics.Config{
		Enabled:     true,
		Path:        "/metrics",
		Namespace:   "laojun",
		Subsystem:   "admin_api",
		Service:     "admin-api",
		Version:     "1.0.0",
		Environment: cfg.Server.Mode,
	}

	metricsInstance := metrics.New(metricsConfig)

	// 初始化健康检查
	healthConfig := health.Config{
		Enabled:     true,
		Path:        "/health",
		Timeout:     30 * time.Second,
		Service:     "admin-api",
		Version:     "1.0.0",
		Environment: cfg.Server.Mode,
	}

	healthChecker := health.New(healthConfig)

	// 初始化数据库
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Error("Error closing database", "error", err)
		}
	}(db)

	// 添加数据库健康检查
	healthChecker.AddChecker(health.NewDatabaseChecker("postgres", db))

	// 初始化Redis
	var redisClient *redis.Client
	if cfg.Redis.Host != "" {
		// 创建内部配置类型
		redisConfig := &config.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}
		
		// 初始化全局Redis客户端
		if err := cache.InitRedis(redisConfig); err != nil {
			log.Warn("Failed to initialize Redis, continuing without cache", "error", err)
		} else {
			redisClient = cache.GetRedisClient()
			healthChecker.AddChecker(health.NewRedisChecker("redis", redisClient))
		}
	}
	defer func() {
		if err := cache.CloseRedis(); err != nil {
			log.Error("Error closing Redis", "error", err)
		}
	}()

	// 运行数据库迁移
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations", "error", err)
	}

	// 初始化共享数据库包装器
sharedDB, err := shareddb.NewDB(&cfg.Database)
if err != nil {
	log.Fatal("Failed to initialize shared DB", "error", err)
}
defer func() {
	if err := sharedDB.Close(); err != nil {
		log.Error("Error closing shared DB", "error", err)
	}
}()
// 初始化服务
adminAuthService := services.NewAdminAuthService(sharedDB)
userService := services.NewUserService(db)
permissionService := services.NewPermissionService(db)
pluginService := services.NewPluginService(sharedDB)

	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建路由
	router := gin.New()
	// 应用中间件到初始路由（占位，后续将复用到最终路由）
	middlewareChain := middleware.NewMiddlewareChain().
		Use(middleware.RequestID()).
		Use(middleware.SecurityHeaders()).
		Use(middleware.Logger(middleware.LoggerConfig{SkipPaths: []string{"/health", "/metrics"}, LogBody: false})).
		Use(metrics.GinMiddleware(metricsInstance)).
		Use(middleware.Timeout(30 * time.Second)).
		Use(middleware.RequestSizeLimit(10 * 1024 * 1024))
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

	// 设置业务路由
	router = routes.SetupRoutes(adminAuthService, userService, permissionService, pluginService, redisClient, cfg, db)
	// 在最终路由上重新应用中间件链
	middlewareChain.Apply(router)
	// 重新添加健康与指标端点到最终路由
	router.GET("/health", healthChecker.GinHandler())
	router.GET("/healthz", healthChecker.GinHandler())
	router.GET("/metrics", gin.WrapH(metricsInstance.Handler()))

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", cfg.Server.Port)
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 打印启动信息：端口与环境
	log.Info("Starting Admin API server", "addr", ":"+port, "env", os.Getenv("APP_ENV"), "mode", cfg.Server.Mode)

	// 优雅关机
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exiting")
}
