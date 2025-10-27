package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"io"
	"path/filepath"
	"strings"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
	"go.uber.org/zap"
	"github.com/codetaoist/laojun-marketplace-api/internal/server"
	"github.com/codetaoist/laojun-marketplace-api/internal/discovery"
	"github.com/codetaoist/laojun-marketplace-api/internal/config"
	sharedconfig "github.com/codetaoist/laojun-shared/config"
	unifiedconfig "github.com/codetaoist/laojun-shared/config"
)

func main() {
	// 统一从仓库根目录加载 .env
	sharedconfig.LoadDotenv()

	// 初始化全局端口配置
	if err := unifiedconfig.InitializeGlobalConfig("../../laojun-shared/config/ports.yaml"); err != nil {
		fmt.Printf("Warning: Failed to load global port config: %v\n", err)
	}

	// 创建zap logger（用于配置加载）
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 获取环境变量
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// 尝试从配置中心加载配置
	var cfg *sharedconfig.Config
	var err error

	configCenterURL := os.Getenv("CONFIG_CENTER_URL")
	if configCenterURL == "" {
		configCenterURL = "http://localhost:8083"
	}

	if configCenterURL != "" {
		configClient := config.NewClient(configCenterURL, "marketplace-api", env, logger)
		ctx := context.Background()
		cfg, err = configClient.LoadConfig(ctx)
		if err != nil {
			log.Printf("Failed to load config from config center: %v, falling back to local config", err)
		}
	}

	// 如果配置中心加载失败，回退到本地配置
	if cfg == nil {
		cfg, err = sharedconfig.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}

	// 配置 Gin 日志输出到文件或双写
	logFile := cfg.Log.File
	if logFile == "" {
		logFile = "./logs/marketplace-api.log"
	}
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		log.Printf("Warning: create log dir failed: %v", err)
	} else {
		fileOutput := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
		switch strings.ToLower(cfg.Log.Output) {
		case "file":
			gin.DefaultWriter = fileOutput
			gin.DefaultErrorWriter = fileOutput
		case "both":
			gin.DefaultWriter = io.MultiWriter(os.Stdout, fileOutput)
			gin.DefaultErrorWriter = gin.DefaultWriter
		default:
			// 默认 stdout，无需更改
		}
	}

	// 创建服务�?
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 获取端口，优先使用统一端口配置，其次使用服务专属环境变量，再次通用变量，最后默认8086
	port := unifiedconfig.GetServicePort("marketplace-api")
	if port == "" {
		port = os.Getenv("MARKETPLACE_PORT")
		if port == "" {
			if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
				port = envPort
			} else if envPort := os.Getenv("PORT"); envPort != "" {
				port = envPort
			} else {
				port = "8086"
			}
		}
	}

	mode := cfg.Server.Mode
	log.Printf("Starting Marketplace API server on port %s (env=%s, mode=%s)", port, env, mode)

	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: srv.GetRouter(),
	}

	// 服务发现注册
	var discoveryClient *discovery.Client
	discoveryURL := os.Getenv("DISCOVERY_URL")
	if discoveryURL == "" {
		discoveryURL = "http://localhost:8081"
	}

	if discoveryURL != "" {
		discoveryClient = discovery.NewClient(discoveryURL, logger)
		
		// 注册服务
		serviceInfo := &discovery.ServiceInfo{
			ID:      "marketplace-api-" + port,
			Name:    "marketplace-api",
			Address: "localhost",
			Port:    parsePort(port),
			Tags:    []string{"api", "marketplace", "v1"},
			Meta: map[string]string{
				"version":     "1.0.0",
				"environment": env,
			},
			Health: discovery.HealthCheck{
				HTTP:     fmt.Sprintf("http://localhost:%s/health", port),
				Interval: "30s",
				Timeout:  "10s",
			},
			TTL: 60,
		}

		ctx := context.Background()
		if err := discoveryClient.Register(ctx, serviceInfo); err != nil {
			log.Printf("Failed to register service: %v", err)
		} else {
			log.Printf("Service registered successfully")
			
			// 启动心跳
			go discoveryClient.StartHeartbeat(ctx)
		}
	}

	// 在goroutine中启动服务器
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 注销服务
	if discoveryClient != nil {
		if err := discoveryClient.Deregister(ctx); err != nil {
			log.Printf("Failed to deregister service: %v", err)
		}
	}
	
	// 关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	
	// 关闭数据库连接
	if err := srv.Close(); err != nil {
		log.Printf("Failed to close server resources: %v", err)
	}

	log.Printf("Server exiting")
}

// parsePort 解析端口号
func parsePort(port string) int {
	if port == "" {
		return 8082
	}
	
	var p int
	if _, err := fmt.Sscanf(port, "%d", &p); err != nil {
		return 8082
	}
	return p
}
