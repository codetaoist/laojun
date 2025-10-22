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

	// "github.com/codetaoist/laojun/internal/api"
	"github.com/codetaoist/laojun/internal/config"
	"github.com/codetaoist/laojun/internal/database"

	// "github.com/codetaoist/laojun/internal/middleware"
	// "github.com/codetaoist/laojun/internal/monitoring"
	"github.com/codetaoist/laojun/internal/plugin"
	"github.com/codetaoist/laojun/pkg/shared/logger"
	"github.com/gin-gonic/gin"
)

// BasicExample 展示 Laojun 平台的基础使用方法
type BasicExample struct {
	config    *config.Config
	db        *database.DB
	router    *gin.Engine
	pluginMgr *plugin.Manager
	// monitor   *monitoring.Monitor
	logger *logger.Logger
	server *http.Server
}

// NewBasicExample 创建基础示例实例
func NewBasicExample() *BasicExample {
	return &BasicExample{}
}

// Initialize 初始化所有组件
func (e *BasicExample) Initialize() error {
	// 1. 初始化配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	e.config = cfg

	// 2. 初始化日志
	e.logger = logger.New(cfg.Logger)
	e.logger.Info("Initializing Laojun Basic Example")

	// 3. 初始化数据库
	e.db, err = database.New(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// 4. 初始化监控
	e.monitor, err = monitoring.New(cfg.Monitoring)
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring: %w", err)
	}

	// 5. 初始化插件管理器
	e.pluginMgr, err = plugin.NewManager(cfg.Plugin)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin manager: %w", err)
	}

	// 6. 初始化路由
	e.setupRouter()

	// 7. 创建 HTTP 服务
	e.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      e.router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	return nil
}

// setupRouter 设置路由和中间件
func (e *BasicExample) setupRouter() {
	// 设置 Gin 模式
	if e.config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	e.router = gin.New()

	// 添加中间件
	e.router.Use(middleware.Logger(e.logger))
	e.router.Use(middleware.Recovery(e.logger))
	e.router.Use(middleware.CORS())
	e.router.Use(middleware.Metrics(e.monitor))

	// 健康检查端点
	e.router.GET("/health", e.healthCheck)
	e.router.GET("/ready", e.readinessCheck)

	// API 路由组
	apiGroup := e.router.Group("/api/v1")
	{
		// 用户相关 API
		userAPI := api.NewUserAPI(e.db, e.logger)
		apiGroup.GET("/users", userAPI.List)
		apiGroup.GET("/users/:id", userAPI.Get)
		apiGroup.POST("/users", userAPI.Create)
		apiGroup.PUT("/users/:id", userAPI.Update)
		apiGroup.DELETE("/users/:id", userAPI.Delete)

		// 插件相关 API
		pluginAPI := api.NewPluginAPI(e.pluginMgr, e.logger)
		apiGroup.GET("/plugins", pluginAPI.List)
		apiGroup.GET("/plugins/:name", pluginAPI.Get)
		apiGroup.POST("/plugins/:name/execute", pluginAPI.Execute)

		// 配置相关 API
		configAPI := api.NewConfigAPI(e.config, e.logger)
		apiGroup.GET("/config", configAPI.Get)
		apiGroup.PUT("/config", configAPI.Update)

		// 监控相关 API
		monitorAPI := api.NewMonitorAPI(e.monitor, e.logger)
		apiGroup.GET("/metrics", monitorAPI.Metrics)
		apiGroup.GET("/stats", monitorAPI.Stats)
	}

	// 静态文件服务
	e.router.Static("/static", "./web/static")
	e.router.StaticFile("/", "./web/index.html")
}

// healthCheck 健康检查端点
func (e *BasicExample) healthCheck(c *gin.Context) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   e.config.App.Version,
	}

	// 检查数据库连接
	if err := e.db.Ping(); err != nil {
		status["database"] = "error"
		status["status"] = "error"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}
	status["database"] = "ok"

	// 检查插件状态
	pluginStatus := e.pluginMgr.HealthCheck()
	status["plugins"] = pluginStatus

	c.JSON(http.StatusOK, status)
}

// readinessCheck 就绪检查端点
func (e *BasicExample) readinessCheck(c *gin.Context) {
	ready := true
	status := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().Unix(),
	}

	// 检查各组件是否就绪
	checks := map[string]bool{
		"database": e.db.IsReady(),
		"plugins":  e.pluginMgr.IsReady(),
		"monitor":  e.monitor.IsReady(),
	}

	for component, isReady := range checks {
		status[component] = isReady
		if !isReady {
			ready = false
		}
	}

	status["ready"] = ready

	if ready {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// Start 启动服务
func (e *BasicExample) Start() error {
	e.logger.Info("Starting Laojun Basic Example server",
		"port", e.config.Server.Port,
		"mode", e.config.Server.Mode)

	// 启动插件
	if err := e.pluginMgr.StartAll(); err != nil {
		return fmt.Errorf("failed to start plugins: %w", err)
	}

	// 启动监控
	if err := e.monitor.Start(); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// 启动 HTTP 服务
	go func() {
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.logger.Error("Failed to start server", "error", err)
		}
	}()

	e.logger.Info("Server started successfully")
	return nil
}

// Stop 停止服务
func (e *BasicExample) Stop() error {
	e.logger.Info("Stopping Laojun Basic Example server")

	// 创建超时上下文，确保服务在 30 秒内关闭	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止 HTTP 服务
	if err := e.server.Shutdown(ctx); err != nil {
		e.logger.Error("Failed to shutdown server gracefully", "error", err)
		return err
	}

	// 停止监控
	if err := e.monitor.Stop(); err != nil {
		e.logger.Error("Failed to stop monitoring", "error", err)
	}

	// 停止插件
	if err := e.pluginMgr.StopAll(); err != nil {
		e.logger.Error("Failed to stop plugins", "error", err)
	}

	// 关闭数据库连接
	if err := e.db.Close(); err != nil {
		e.logger.Error("Failed to close database", "error", err)
	}

	e.logger.Info("Server stopped successfully")
	return nil
}

// 演示插件使用
func (e *BasicExample) demonstratePlugins() {
	e.logger.Info("Demonstrating plugin usage")

	// 列出所有可用插件
	plugins := e.pluginMgr.ListPlugins()
	e.logger.Info("Available plugins", "count", len(plugins))

	for _, plugin := range plugins {
		e.logger.Info("Plugin info",
			"name", plugin.Name(),
			"version", plugin.Version(),
			"status", plugin.Status())

		// 执行插件（示例）
		if plugin.Status() == "active" {
			result, err := plugin.Execute(context.Background(), map[string]interface{}{
				"action": "demo",
				"data":   "Hello from basic example",
			})
			if err != nil {
				e.logger.Error("Plugin execution failed",
					"plugin", plugin.Name(),
					"error", err)
			} else {
				e.logger.Info("Plugin execution result",
					"plugin", plugin.Name(),
					"result", result)
			}
		}
	}
}

// 演示监控功能
func (e *BasicExample) demonstrateMonitoring() {
	e.logger.Info("Demonstrating monitoring features")

	// 创建计数器
	requestCounter := e.monitor.Counter("http_requests_total", map[string]string{
		"method": "GET",
		"path":   "/api/v1/users",
	})
	requestCounter.Inc()

	// 创建仪表
	activeConnections := e.monitor.Gauge("active_connections", nil)
	activeConnections.Set(42)

	// 创建直方图
	responseTime := e.monitor.Histogram("http_request_duration_seconds", map[string]string{
		"method": "GET",
	})
	responseTime.Observe(0.123)

	// 创建计时器
	timer := e.monitor.Timer("database_query_duration", map[string]string{
		"query": "select_users",
	})
	defer timer.ObserveDuration()

	// 模拟数据库查询
	time.Sleep(50 * time.Millisecond)

	e.logger.Info("Monitoring metrics updated")
}

func main() {
	// 创建基础示例实例
	example := NewBasicExample()

	// 初始化示例
	if err := example.Initialize(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// 启动服务
	if err := example.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 演示功能
	go func() {
		time.Sleep(2 * time.Second)
		example.demonstratePlugins()
		example.demonstrateMonitoring()
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	if err := example.Stop(); err != nil {
		log.Printf("Failed to stop server gracefully: %v", err)
	}
}
