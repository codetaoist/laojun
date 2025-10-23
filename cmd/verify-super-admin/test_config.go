package main

import (
	"fmt"
	"os"

	sharedconfig "github.com/codetaoist/laojun/pkg/shared/config"
)

func main() {
	// 加载配置
	sharedconfig.LoadDotenv()
	
	cfg, err := sharedconfig.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("配置加载成功:\n")
	fmt.Printf("Server Mode: %s\n", cfg.Server.Mode)
	fmt.Printf("Server Port: %d\n", cfg.Server.Port)
	fmt.Printf("Security.EnableCaptcha: %v\n", cfg.Security.EnableCaptcha)
	fmt.Printf("Security.CaptchaTTL: %v\n", cfg.Security.CaptchaTTL)
	
	// 检查环境变量
	fmt.Printf("\n环境变量:\n")
	fmt.Printf("SECURITY_ENABLE_CAPTCHA: %s\n", os.Getenv("SECURITY_ENABLE_CAPTCHA"))
	fmt.Printf("GIN_MODE: %s\n", os.Getenv("GIN_MODE"))
	fmt.Printf("APP_ENV: %s\n", os.Getenv("APP_ENV"))
}