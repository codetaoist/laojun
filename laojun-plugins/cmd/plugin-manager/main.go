package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/laojun-plugins/internal/handlers"
	"github.com/codetaoist/laojun-plugins/internal/routes"
	"github.com/codetaoist/laojun-plugins/internal/services"
	"github.com/codetaoist/laojun-shared/database"
)

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// 初始化全局端口配置
	if err := sharedconfig.InitializeGlobalConfig("../../laojun-shared/config/ports.yaml"); err != nil {
		fmt.Printf("Warning: Failed to load global port config: %v\n", err)
	}

	// 初始化配置
	cfg, err := config.Load("./configs/plugin-manager.yaml") // 相对于服务目录
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 使用统一端口配置覆盖服务端口
	if pluginsPort := sharedconfig.GetServicePort("plugins"); pluginsPort != "" {
		if port, err := strconv.Atoi(pluginsPort); err == nil {
			cfg.Server.Port = port
		}
	}

	// 初始化日志
	logger.Init(cfg.Log)

	// 初始化数据库
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 初始化服务
	pluginService := services.NewPluginService(db)
	pluginReviewService := services.NewPluginReviewService(db)
	extendedPluginService := services.NewExtendedPluginService(db)

	// 初始化处理器
	pluginHandler := handlers.NewPluginHandler(pluginService)
	pluginReviewHandler := handlers.NewPluginReviewHandler(pluginReviewService)
	extendedPluginHandler := handlers.NewExtendedPluginHandler(extendedPluginService)

	// 设置 Gin 模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 注册路由
	routes.RegisterPluginRoutes(router, pluginHandler)
	routes.RegisterPluginReviewRoutes(router, pluginReviewHandler)
	routes.RegisterExtendedPluginRoutes(router, extendedPluginHandler)

	// 创建 HTTP 服务
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 启动服务
	go func() {
		log.Printf("Plugin Manager server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Plugin Manager server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Plugin Manager server forced to shutdown: %v", err)
	}

	log.Println("Plugin Manager server exited")
}
