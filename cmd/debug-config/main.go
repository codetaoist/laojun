package main

import (
	"fmt"
	"os"

	sharedconfig "github.com/codetaoist/laojun/pkg/shared/config"
)

func main() {
	// 加载配置
	cfg, err := sharedconfig.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 打印关键配置信息
	fmt.Printf("Server.Mode: %s\n", cfg.Server.Mode)
	fmt.Printf("Security.EnableCaptcha: %t\n", cfg.Security.EnableCaptcha)
	fmt.Printf("Security.CaptchaTTL: %v\n", cfg.Security.CaptchaTTL)
	
	// 打印相关环境变量
	fmt.Printf("GIN_MODE: %s\n", os.Getenv("GIN_MODE"))
	fmt.Printf("SECURITY_ENABLE_CAPTCHA: %s\n", os.Getenv("SECURITY_ENABLE_CAPTCHA"))
	fmt.Printf("SECURITY_CAPTCHA_TTL: %s\n", os.Getenv("SECURITY_CAPTCHA_TTL"))
}