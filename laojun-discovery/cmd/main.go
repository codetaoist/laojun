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

	"github.com/codetaoist/laojun-discovery/internal/config"
	"github.com/codetaoist/laojun-discovery/internal/handlers"
	"github.com/codetaoist/laojun-discovery/internal/registry"
	"github.com/codetaoist/laojun-discovery/internal/services"
	"github.com/codetaoist/laojun-discovery/internal/storage"
	"github.com/gin-gonic/gin"
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

	// 创建统一配置管理器
	configManager := sharedconfig.NewDefaultConfigManager(
		sharedconfig.NewMemoryConfigStorage(),
		&sharedconfig.ConfigOptions{
			HistorySize: 100,
		},
	)

	// 初始化配置
	cfg, err := config.LoadWithConfigManager(configManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 使用统一端口配置覆盖服务端口
	if discoveryPort := sharedconfig.GetServicePort("discovery"); discoveryPort != "" {
		if port, err := strconv.Atoi(discoveryPort); err == nil {
			cfg.Server.Port = port
		}
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化存储
	store, err := storage.NewStorage(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer store.Close()

	// 初始化服务注册表
	serviceRegistry := registry.NewServiceRegistry(store, logger)

	// 初始化服务管理器
	serviceManager := services.NewServiceManager(cfg, serviceRegistry, logger)

	// 初始化处理器
	healthHandler := handlers.NewHealthHandler(serviceManager, logger)
	registryHandler := handlers.NewRegistryHandler(serviceRegistry, logger)
	discoveryHandler := handlers.NewDiscoveryHandler(serviceRegistry, logger)
	
	// 初始化增强处理器
	enhancedRegistryHandler := handlers.NewEnhancedRegistryHandler(serviceRegistry, cfg, logger)
	enhancedDiscoveryHandler := handlers.NewEnhancedDiscoveryHandler(serviceRegistry, cfg, logger)
	
	// 初始化统一配置管理处理器
	unifiedConfigHandler := handlers.NewUnifiedConfigHandler(configManager, logger)

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// 健康检查路由
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 服务注册路由
		api.POST("/services", registryHandler.RegisterService)
		api.DELETE("/services/:id", registryHandler.DeregisterService)
		api.PUT("/services/:id/health", registryHandler.UpdateHealth)
		api.GET("/services", registryHandler.ListServices)
		api.GET("/services/:name", registryHandler.GetService)

		// 服务发现路由
		api.GET("/discovery/services", discoveryHandler.DiscoverServices)
		api.GET("/discovery/services/:name", discoveryHandler.DiscoverService)
		api.GET("/discovery/services/:name/healthy", discoveryHandler.GetHealthyInstances)
		
		// 增强服务注册路由
		enhanced := api.Group("/enhanced")
		{
			// 增强注册功能
			enhanced.POST("/services", enhancedRegistryHandler.RegisterService)
			enhanced.DELETE("/services/:id", enhancedRegistryHandler.DeregisterService)
			enhanced.PUT("/services/:id/health", enhancedRegistryHandler.UpdateServiceHealth)
			enhanced.GET("/services/stats", enhancedRegistryHandler.GetRegistrationStats)
			
			// 增强发现功能
			enhanced.GET("/discovery/services/:name", enhancedDiscoveryHandler.DiscoverWithLoadBalancing)
			enhanced.GET("/discovery/services/:name/multiple", enhancedDiscoveryHandler.DiscoverMultipleInstances)
			enhanced.GET("/discovery/services/:name/circuit-breaker", enhancedDiscoveryHandler.GetCircuitBreakerStatus)
			enhanced.GET("/discovery/services/:name/load-balancer", enhancedDiscoveryHandler.GetLoadBalancerStats)
			enhanced.GET("/discovery/services/:name/rate-limit", enhancedDiscoveryHandler.GetRateLimitStatus)
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

	// 监控路由
	router.GET("/metrics", gin.WrapH(handlers.GetMetricsHandler()))

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 启动服务管理器
	if err := serviceManager.Start(); err != nil {
		logger.Fatal("Failed to start service manager", zap.Error(err))
	}

	// 启动HTTP服务器
	go func() {
		logger.Info("Starting Discovery Service",
			zap.Int("port", cfg.Server.Port),
			zap.String("mode", cfg.Server.Mode))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Discovery Service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止服务管理器
	serviceManager.Stop()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Discovery Service stopped")
}