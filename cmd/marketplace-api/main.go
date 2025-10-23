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

	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/codetaoist/laojun/internal/server"
	"github.com/codetaoist/laojun/pkg/shared/config"
)

func main() {
	// 统一从仓库根目录加载 .env
	config.LoadDotenv()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	// 创建服务器
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 获取端口，优先使用服务专属环境变量，其次通用变量，最后默认 8082
	port := os.Getenv("MARKETPLACE_PORT")
	if port == "" {
		if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
			port = envPort
		} else if envPort := os.Getenv("PORT"); envPort != "" {
			port = envPort
		} else {
			port = "8082"
		}
	}

	env := os.Getenv("APP_ENV")
	mode := cfg.Server.Mode
	log.Printf("Starting Marketplace API server on port %s (env=%s, mode=%s)", port, env, mode)

	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: srv.GetRouter(),
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
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
