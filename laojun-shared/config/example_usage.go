package config

import (
	"fmt"
	"log"
)

// ExampleUsage 展示如何使用统一端口配置
func ExampleUsage() {
	// 方式1: 使用默认配置
	fmt.Println("=== 使用默认配置 ===")
	defaultConfig := DefaultPortConfig()
	fmt.Printf("Gateway端口: %d\n", defaultConfig.GetServicePort("gateway"))
	fmt.Printf("Admin API端口: %d\n", defaultConfig.GetServicePort("admin-api"))
	fmt.Printf("监控服务端口: %d\n", defaultConfig.GetServicePort("monitoring"))

	// 方式2: 从配置文件加载
	fmt.Println("\n=== 从配置文件加载 ===")
	loader := NewConfigLoader("../laojun-shared/config/ports.yaml")
	config, err := loader.LoadPortConfig()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		config = DefaultPortConfig()
	}

	// 获取服务地址
	fmt.Printf("Gateway地址: %s\n", config.GetServerAddress("gateway"))
	fmt.Printf("Admin API地址: %s\n", config.GetServerAddress("admin-api"))
	fmt.Printf("监控服务地址: %s\n", config.GetServerAddress("monitoring"))

	// 方式3: 使用全局配置
	fmt.Println("\n=== 使用全局配置 ===")
	// 初始化全局配置
	if err := InitGlobalPortConfig("../laojun-shared/config/ports.yaml"); err != nil {
		log.Printf("初始化全局配置失败: %v", err)
	}

	// 直接使用全局函数
	fmt.Printf("Gateway端口: %d\n", GetServicePort("gateway"))
	fmt.Printf("Admin API地址: %s\n", GetServerAddress("admin-api"))
	fmt.Printf("监控服务地址: %s\n", GetServerAddress("monitoring", "localhost"))

	// 获取所有服务端口
	fmt.Println("\n=== 所有服务端口 ===")
	allPorts := config.GetAllServicePorts()
	for service, port := range allPorts {
		fmt.Printf("%s: %d\n", service, port)
	}

	// 验证配置
	fmt.Println("\n=== 配置验证 ===")
	if err := config.ValidatePortConfig(); err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
	} else {
		fmt.Println("配置验证通过")
	}
}

// ExampleServiceIntegration 展示如何在具体服务中集成
func ExampleServiceIntegration() {
	fmt.Println("=== 服务集成示例 ===")

	// 在 main.go 中的使用示例
	fmt.Println("\n// 在 main.go 中初始化")
	fmt.Println(`
func main() {
    // 初始化端口配置
    configPath := filepath.Join("configs", "ports.yaml")
    if err := config.InitGlobalPortConfig(configPath); err != nil {
        log.Printf("初始化端口配置失败: %v", err)
    }

    // 获取当前服务端口
    port := config.GetServicePort("admin-api")
    addr := config.GetServerAddress("admin-api")
    
    fmt.Printf("启动Admin API服务，端口: %d\n", port)
    fmt.Printf("服务地址: %s\n", addr)
    
    // 启动HTTP服务器
    if err := http.ListenAndServe(addr, router); err != nil {
        log.Fatal(err)
    }
}`)

	// 在配置结构体中的使用示例
	fmt.Println("\n// 在配置结构体中集成")
	fmt.Println(`
type ServerConfig struct {
    Host string
    Port int
    // 其他配置...
}

func LoadServerConfig() *ServerConfig {
    return &ServerConfig{
        Host: "0.0.0.0",
        Port: config.GetServicePort("admin-api"),
    }
}`)

	// 在Docker环境中的使用示例
	fmt.Println("\n// Docker环境变量覆盖示例")
	fmt.Println(`
# docker-compose.yml
services:
  admin-api:
    environment:
      - ADMIN_API_PORT=8082
    ports:
      - "8082:8082"
      
  gateway:
    environment:
      - GATEWAY_PORT=8081
    ports:
      - "8081:8081"`)
}