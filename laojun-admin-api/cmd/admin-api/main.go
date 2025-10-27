package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/database"
	"github.com/codetaoist/laojun-admin-api/internal/routes"
	"github.com/codetaoist/laojun-admin-api/internal/services"
	"github.com/codetaoist/laojun-admin-api/pkg/discovery"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	shareddb "github.com/codetaoist/laojun-shared/database"
	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"

	"github.com/codetaoist/laojun-shared/health"
	"github.com/codetaoist/laojun-shared/logger"
	"github.com/codetaoist/laojun-shared/metrics"
	"github.com/codetaoist/laojun-shared/middleware"
	"go.uber.org/zap"
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
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
		File: logger.FileConfig{
			Filename:   cfg.Log.File,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		Service:     "admin-api",
		Version:     "1.0.0",
		Environment: cfg.Server.Mode,
	}

	log := logger.New(logConfig)

	// 创建zap logger用于discovery client
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

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
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.Connect(dbConfig)
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
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		// 测试Redis连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Warn("Failed to connect to Redis, continuing without cache", "error", err)
			redisClient = nil
		} else {
			healthChecker.AddChecker(health.NewRedisChecker("redis", redisClient))
		}
	}

	if redisClient != nil {
		defer func() {
			if err := redisClient.Close(); err != nil {
				log.Error("Error closing Redis", "error", err)
			}
		}()
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
	permissionService := services.NewPermissionService(db, redisClient)

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
	fmt.Println("About to call SetupRoutes...")
	router = routes.SetupRoutes(adminAuthService, userService, permissionService, redisClient, cfg, db)
	fmt.Println("SetupRoutes completed")
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

	// 创建服务发现客户端
	discoveryClient := discovery.NewClient("http://localhost:8081", zapLogger)

	// 优雅关机
	var serviceInstance *discovery.ServiceInstance
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 注册服务到discovery
	portInt, _ := strconv.Atoi(port)
	registerReq := &discovery.RegisterRequest{
		Name:    "admin-api",
		Address: "localhost",
		Port:    portInt,
		Tags:    []string{"admin", "api", "v1"},
		Meta: map[string]string{
			"version": "1.0.0",
			"env":     os.Getenv("APP_ENV"),
		},
		TTL: 30,
	}

	serviceInstance, err = discoveryClient.Register(registerReq)
	if err != nil {
		log.Error("Failed to register service", "error", err)
	} else {
		log.Info("Service registered successfully", "id", serviceInstance.ID)

		// 启动健康状态更新协程
		go func() {
			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				if err := discoveryClient.UpdateHealth(serviceInstance.ID, "passing"); err != nil {
					log.Error("Failed to update health status", "error", err)
				}
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 注销服务
	if serviceInstance != nil {
		if err := discoveryClient.Deregister(serviceInstance.ID); err != nil {
			log.Error("Failed to deregister service", "error", err)
		} else {
			log.Info("Service deregistered successfully")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Server exiting")
}
