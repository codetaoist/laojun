package main

import (
	"context"
	"fmt"
	"time"

	"github.com/taishanglaojun/plugins/core"
)

// HelloPlugin 示例Go插件
type HelloPlugin struct {
	config    map[string]any
	startTime time.Time
	counter   int64
}

// GetMetadata 获取插件元数据
func (p *HelloPlugin) GetMetadata() *core.PluginMetadata {
	return &core.PluginMetadata{
		ID:           "hello-go-plugin",
		Name:         "Hello Go Plugin",
		Version:      "1.0.0",
		Description:  "A simple Hello World plugin written in Go",
		Type:         core.TypeFilter,
		Runtime:      core.RuntimeGo,
		EntryPoint:   "hello_plugin.so",
		Author:       "Taishanglaojun Team",
		Dependencies: []string{},
		Permissions:  []string{"network.http"},
		Config: map[string]any{
			"greeting": "Hello",
			"language": "en",
		},
	}
}

// Initialize 初始化插件
func (p *HelloPlugin) Initialize(ctx context.Context, config map[string]any) error {
	p.config = config
	p.startTime = time.Now()
	p.counter = 0

	fmt.Printf("Hello Go Plugin initialized with config: %+v\n", config)
	return nil
}

// Start 启动插件
func (p *HelloPlugin) Start(ctx context.Context) error {
	fmt.Println("Hello Go Plugin started")
	return nil
}

// Stop 停止插件
func (p *HelloPlugin) Stop(ctx context.Context) error {
	fmt.Println("Hello Go Plugin stopped")
	return nil
}

// HandleRequest 处理请求
func (p *HelloPlugin) HandleRequest(ctx context.Context, req *core.PluginRequest) (*core.PluginResponse, error) {
	p.counter++

	// 解析请求数据
	var requestData map[string]any
	if req.Data != nil {
		requestData = req.Data
	} else {
		requestData = make(map[string]any)
	}

	// 获取配置中的问候语
	greeting := "Hello"
	if g, ok := p.config["greeting"].(string); ok {
		greeting = g
	}

	// 获取请求中的名称
	name := "World"
	if n, ok := requestData["name"].(string); ok {
		name = n
	}

	// 构造响应数据
	responseData := map[string]any{
		"message":    fmt.Sprintf("%s, %s!", greeting, name),
		"counter":    p.counter,
		"uptime":     time.Since(p.startTime).String(),
		"plugin_id":  p.GetMetadata().ID,
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": req.ID,
	}

	// 如果请求包含特殊参数，添加额外信息
	if includeStats, ok := requestData["include_stats"].(bool); ok && includeStats {
		responseData["stats"] = map[string]any{
			"total_requests": p.counter,
			"start_time":     p.startTime.Format(time.RFC3339),
			"config":         p.config,
		}
	}

	return &core.PluginResponse{
		Success: true,
		Data:    responseData,
		Message: "Request processed successfully",
	}, nil
}

// HandleEvent 处理事件
func (p *HelloPlugin) HandleEvent(ctx context.Context, event *core.PluginEvent) error {
	fmt.Printf("Hello Go Plugin received event: %s from %s\n", event.Type, event.Source)

	// 根据事件类型执行不同的处理逻辑
	switch event.Type {
	case "system.startup":
		fmt.Println("System startup event received")
	case "plugin.loaded":
		if pluginID, ok := event.Data["id"].(string); ok {
			fmt.Printf("Another plugin loaded: %s\n", pluginID)
		}
	case "config.updated":
		fmt.Println("Configuration updated event received")
		// 这里可以重新加载配置
	default:
		fmt.Printf("Unknown event type: %s\n", event.Type)
	}

	return nil
}

// GetStatus 获取插件状态
func (p *HelloPlugin) GetStatus() core.PluginStatus {
	return core.StatusRunning
}

// GetHealth 获取插件健康状态
func (p *HelloPlugin) GetHealth(ctx context.Context) (*core.PluginHealth, error) {
	return &core.PluginHealth{
		Status:    "healthy",
		Message:   "Plugin is running normally",
		Timestamp: time.Now(),
		Details: map[string]any{
			"uptime":         time.Since(p.startTime).String(),
			"total_requests": p.counter,
			"memory_usage":   "N/A", // 在实际实现中可以获取真实的内存使用情况
			"last_request":   time.Now().Format(time.RFC3339),
		},
	}, nil
}

// 插件入口点 - 必须导出 Plugin 变量
var Plugin HelloPlugin

// 可选：提供插件工厂函数
func NewHelloPlugin() *HelloPlugin {
	return &HelloPlugin{}
}

// 可选：提供插件配置验证函数
func ValidateConfig(config map[string]any) error {
	// 验证必需的配置项
	if greeting, ok := config["greeting"]; ok {
		if _, ok := greeting.(string); !ok {
			return fmt.Errorf("greeting must be a string")
		}
	}

	if language, ok := config["language"]; ok {
		if _, ok := language.(string); !ok {
			return fmt.Errorf("language must be a string")
		}

		// 验证支持的语言
		supportedLanguages := []string{"en", "zh", "es", "fr", "de"}
		lang := language.(string)
		supported := false
		for _, supportedLang := range supportedLanguages {
			if lang == supportedLang {
				supported = true
				break
			}
		}
		if !supported {
			return fmt.Errorf("unsupported language: %s", lang)
		}
	}

	return nil
}

// 可选：提供插件信息函数
func GetPluginInfo() map[string]any {
	return map[string]any{
		"build_time":    time.Now().Format(time.RFC3339),
		"go_version":    "1.21+",
		"dependencies":  []string{},
		"capabilities":  []string{"request_handling", "event_handling", "health_check"},
		"documentation": "https://docs.taishanglaojun.com/plugins/hello-go-plugin",
	}
}
