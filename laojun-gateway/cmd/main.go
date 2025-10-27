package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codetaoist/laojun-gateway/internal/config"
	"github.com/codetaoist/laojun-gateway/internal/routes"
	"github.com/codetaoist/laojun-gateway/internal/services"
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
	configManager := sharedconfig.NewDefaultConfigManager(&sharedconfig.ConfigOptions{
		Storage:     &sharedconfig.MemoryConfigStorage{},
		HistorySize: 100,
	})

	// 加载配置
	cfg, err := config.LoadWithConfigManager(configManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 使用统一端口配置覆盖服务端口
	if gatewayPort := sharedconfig.GetServicePort("gateway"); gatewayPort != "" {
		cfg.Server.Port = gatewayPort
	}

	// 初始化日志
	logger, err := initLogger(cfg.Server.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Laojun Gateway", 
		zap.String("version", "1.0.0"),
		zap.String("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode))

	// 初始化服务管理器
	serviceManager, err := services.NewServiceManager(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize service manager", zap.Error(err))
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 设置路由
	router := routes.SetupRoutes(cfg, serviceManager, logger)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info("Server starting", zap.String("address", server.Addr))
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

	// 关闭服务管理器
	if err := serviceManager.Close(); err != nil {
		logger.Error("Error closing service manager", zap.Error(err))
	}

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server exited gracefully")
	}
}

func initLogger(level string) (*zap.Logger, error) {
	var config zap.Config
	
	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "info", "warn", "error":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(getLogLevel(level))
	default:
		config = zap.NewProductionConfig()
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	return config.Build()
}

func getLogLevel(level string) zap.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}