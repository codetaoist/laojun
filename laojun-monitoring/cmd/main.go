package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/codetaoist/laojun-monitoring/internal/config"
	"github.com/codetaoist/laojun-monitoring/internal/handlers"
	"github.com/codetaoist/laojun-monitoring/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	"github.com/joho/godotenv"
)

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// 初始化全局端口配置
	if err := sharedconfig.InitializeGlobalConfig("../laojun-shared/config/ports.yaml"); err != nil {
		fmt.Printf("Warning: Failed to load global port config: %v\n", err)
	}

	// 初始化统一配置管理器
	configManager := sharedconfig.NewDefaultConfigManager(
		sharedconfig.WithStorage(sharedconfig.NewMemoryConfigStorage()),
		sharedconfig.WithHistorySize(100),
	)

	// 初始化配置
	cfg, err := config.LoadWithConfigManager(configManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 使用统一端口配置覆盖服务端口
	if monitoringPort := sharedconfig.GetServicePort("monitoring"); monitoringPort != "" {
		if port, err := strconv.Atoi(monitoringPort); err == nil {
			cfg.Server.Port = port
		}
	}

	// 初始化日志
	logger, err := initLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Laojun Monitoring Service",
		zap.String("version", "1.0.0"),
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port))

	// 初始化监控服务
	monitoringService, err := services.NewMonitoringService(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create monitoring service", zap.Error(err))
	}

	// 启动监控服务
	if err := monitoringService.Start(); err != nil {
		logger.Fatal("Failed to start monitoring service", zap.Error(err))
	}
	defer monitoringService.Stop()

	// 初始化处理器
	healthHandler := handlers.NewHealthHandler(monitoringService, logger)
	metricsHandler := handlers.NewMetricsHandler(monitoringService, logger)
	alertHandler := handlers.NewAlertHandler(monitoringService, logger)
	
	// 初始化统一配置管理处理器
	unifiedConfigHandler := handlers.NewUnifiedConfigHandler(configManager, logger)

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// 健康检查路由
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API 路由组
	api := router.Group("/api/v1")
	{
		// 指标相关路由
		metrics := api.Group("/metrics")
		{
			metrics.GET("", metricsHandler.GetMetrics)
			metrics.GET("/query", metricsHandler.QueryMetrics)
			metrics.GET("/range", metricsHandler.QueryRangeMetrics)
			metrics.POST("/custom", metricsHandler.CreateCustomMetric)
		}

		// 告警相关路由
		alerts := api.Group("/alerts")
		{
			alerts.GET("", alertHandler.ListAlerts)
			alerts.POST("", alertHandler.CreateAlert)
			alerts.GET("/:id", alertHandler.GetAlert)
			alerts.PUT("/:id", alertHandler.UpdateAlert)
			alerts.DELETE("/:id", alertHandler.DeleteAlert)
			alerts.POST("/:id/silence", alertHandler.SilenceAlert)
		}

		// 配置相关路由
		config := api.Group("/config")
		{
			config.GET("", handlers.GetConfig)
			config.PUT("", handlers.UpdateConfig)
			config.POST("/reload", handlers.ReloadConfig)
		}
		
		// 统一配置管理路由
		unified := api.Group("/unified")
		{
			unified.GET("/config/:key", unifiedConfigHandler.GetConfig)
			unified.GET("/config/:key/type", unifiedConfigHandler.GetConfigWithType)
			unified.POST("/config/:key", unifiedConfigHandler.SetConfig)
			unified.DELETE("/config/:key", unifiedConfigHandler.DeleteConfig)
			unified.GET("/configs", unifiedConfigHandler.ListConfigs)
			unified.POST("/configs/batch", unifiedConfigHandler.BatchSetConfigs)
			unified.GET("/config/:key/history", unifiedConfigHandler.GetConfigHistory)
			unified.GET("/config/:key/watch", unifiedConfigHandler.WatchConfigs)
			unified.GET("/config/:key/exists", unifiedConfigHandler.ExistsConfig)
			unified.GET("/health", unifiedConfigHandler.HealthCheck)
		}
	}

	// Prometheus 指标端点
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("address", server.Addr))
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// initLogger 初始化日志
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Logging.Format == "json" {
		if cfg.Server.Mode == "release" {
			logger, err = zap.NewProduction()
		} else {
			logger, err = zap.NewDevelopment()
		}
	} else {
		config := zap.NewDevelopmentConfig()
		config.Encoding = "console"
		logger, err = config.Build()
	}

	if err != nil {
		return nil, err
	}

	// 设置日志级别
	switch cfg.Logging.Level {
	case "debug":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	case "info":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	case "warn":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))
	case "error":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	}

	return logger, nil
}