package main

import (
	"log"
	"os"
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

	// 创建并启动服务器
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 获取端口，默认为8081
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	env := os.Getenv("APP_ENV")
	mode := cfg.Server.Mode
	log.Printf("Starting Marketplace API server on port %s (env=%s, mode=%s)", port, env, mode)

	if err := srv.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
