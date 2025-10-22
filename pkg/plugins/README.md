# 太上老君插件系统

太上老君插件系统是一个高性能、安全、可扩展的插件架构，支持多种运行时环境和部署模式。

## 🏗️ 架构概览

### 核心组件

- **PluginManager**: 插件生命周期管理器
- **PluginLoader**: 插件加载器（支持 Go Plugin 和 JavaScript）
- **SecurityManager**: 安全管理器和沙箱系统
- **EventBus**: 事件总线和通信系统
- **PluginRegistry**: 插件注册表和发现服务
- **ConfigManager**: 配置管理系统

### 架构层次

- **核心层**: 基于接口的插件系统，统一生命周期（install/uninstall/activate/deactivate），注册路由/菜单/权限/组件
- **扩展层**: 微服务插件（独立服务），通过API/消息与核心层集成，由核心层聚合权限与菜单

### 支持的插件类型

1. **进程内插件**
   - Go Plugin (.so 文件)
   - JavaScript 插件 (V8 沙箱)

2. **微服务插件**
   - Docker 容器插件
   - Kubernetes 部署插件
   - gRPC 通信

## 📁 目录结构

```
plugins/
├── core/                   # 核心接口定义
│   ├── interfaces.go      # 主要接口定义
│   └── manager.go         # 插件管理器实现
├── loaders/               # 插件加载器
│   ├── go_loader.go       # Go 插件加载器
│   └── js_loader.go       # JavaScript 插件加载器
├── security/              # 安全管理
│   └── manager.go         # 安全管理器和沙箱
├── events/                # 事件系统
│   └── bus.go            # 事件总线实现
├── registry/              # 插件注册表
│   └── registry.go       # 注册表实现
├── config/                # 配置管理
│   └── config.go         # 配置管理器
├── examples/              # 示例插件
├── docs/                  # 文档
└── tests/                 # 测试文件
```

### 插件目录建议
```
plugins/
  plugin-name/
    manifest.json          # 插件清单
    plugin.go             # Go 插件入口（可选）
    plugin.js             # JavaScript 插件入口（可选）
    backend/              # 后端代码
    frontend/             # 前端代码
    migrations/           # 数据库迁移
    permissions/          # 权限定义
    docs/                 # 文档
    config/               # 配置文件
```

## 🚀 快速开始

### 1. 初始化插件系统

```go
package main

import (
    "context"
    "log"
    
    "github.com/taishanglaojun/plugins/core"
    "github.com/taishanglaojun/plugins/config"
    "github.com/taishanglaojun/plugins/security"
    "github.com/taishanglaojun/plugins/events"
    "github.com/taishanglaojun/plugins/registry"
    "github.com/taishanglaojun/plugins/loaders"
)

func main() {
    // 1. 加载配置
    configManager, err := config.NewConfigManager("./config/system.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建安全管理器
    securityConfig := configManager.GetSecurityConfig()
    securityManager := security.NewDefaultSecurityManager(&securityConfig)
    
    // 3. 创建事件总线
    eventConfig := configManager.GetEventSystemConfig()
    eventBus := events.NewDefaultEventBus(eventConfig.WorkerCount)
    
    // 4. 创建插件注册表
    registryConfig := configManager.GetRegistryConfig()
    pluginRegistry := registry.NewDefaultPluginRegistry(&registryConfig)
    
    // 5. 创建插件管理器
    pluginManager := core.NewDefaultPluginManager(
        securityManager,
        eventBus,
        pluginRegistry,
    )
    
    // 6. 注册加载器
    goLoader := loaders.NewGoPluginLoader()
    jsLoader := loaders.NewJSPluginLoader()
    
    pluginManager.RegisterLoader("go", goLoader)
    pluginManager.RegisterLoader("js", jsLoader)
    
    // 7. 扫描并加载插件
    ctx := context.Background()
    if err := pluginRegistry.Scan(ctx); err != nil {
        log.Printf("Plugin scan error: %v", err)
    }
    
    log.Println("Plugin system started successfully")
}
```

### 2. 创建 Go 插件

```go
// plugin.go
package main

import (
    "context"
    "github.com/taishanglaojun/plugins/core"
)

type MyPlugin struct {
    config map[string]any
}

func (p *MyPlugin) GetMetadata() *core.PluginMetadata {
    return &core.PluginMetadata{
        ID:          "my-plugin",
        Name:        "My Plugin",
        Version:     "1.0.0",
        Description: "A sample plugin",
        Type:        core.TypeFilter,
        Runtime:     core.RuntimeGo,
        EntryPoint:  "plugin.so",
        Author:      "Developer",
    }
}

func (p *MyPlugin) Initialize(ctx context.Context, config map[string]any) error {
    p.config = config
    return nil
}

func (p *MyPlugin) HandleRequest(ctx context.Context, req *core.PluginRequest) (*core.PluginResponse, error) {
    return &core.PluginResponse{
        Success: true,
        Data:    map[string]any{"message": "Hello from plugin"},
    }, nil
}

// 插件入口点
var Plugin MyPlugin
```

编译插件：
```bash
go build -buildmode=plugin -o plugin.so plugin.go
```

### 3. 创建 JavaScript 插件

```javascript
// plugin.js
const plugin = {
    metadata: {
        id: "my-js-plugin",
        name: "My JavaScript Plugin",
        version: "1.0.0",
        type: "filter",
        runtime: "js"
    },
    
    initialize: function(configStr) {
        this.config = JSON.parse(configStr);
        return null;
    },
    
    handleRequest: function(requestStr) {
        const request = JSON.parse(requestStr);
        const response = {
            success: true,
            data: { message: "Hello from JavaScript plugin" }
        };
        return JSON.stringify(response);
    }
};
```

## 📋 清单与兼容

### 插件清单文件 (manifest.json)

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "description": "A sample plugin",
  "type": "filter",
  "runtime": "go",
  "entryPoint": "plugin.so",
  "author": "Developer",
  "dependencies": [],
  "permissions": ["network.http"],
  "minSystemVersion": "1.0.0",
  "maxSystemVersion": "2.0.0",
  "supportedDevices": ["server", "desktop"],
  "config": {
    "timeout": 30,
    "retries": 3
  }
}
```

### 兼容性检查

- 安装前校验 `min/max system version` 与依赖闭包
- 设备支持检查
- 权限验证
- 依赖关系解析

## 🔒 安全特性

### 沙箱隔离

- **内存限制**: 每个插件的内存使用限制
- **CPU 时间限制**: 防止插件占用过多 CPU 资源
- **Goroutine 限制**: 控制并发数量
- **API 访问控制**: 限制敏感 API 的访问

### 权限系统

- **细粒度权限**: 网络访问、文件系统、系统调用等
- **权限验证**: 运行时权限检查
- **安全策略**: 可配置的安全策略

## 🔧 配置

系统配置文件支持以下配置项：
- 插件目录和加载器配置
- 安全策略和沙箱设置
- 事件系统配置
- 注册表和缓存设置
- 性能监控配置

## 📊 监控和指标

- 插件资源使用情况
- 请求处理时间
- 错误率统计
- 事件处理性能
- 健康状态监控

## 🛠️ SDK/工具

- Go SDK: 完整的插件开发接口
- JavaScript SDK: V8 沙箱环境支持
- 配置管理工具
- 插件打包和部署工具
- 开发调试工具

## 🧪 测试

运行测试：
```bash
go test ./...
```

## 📚 文档

- [核心接口文档](docs/core-interfaces.md)
- [插件开发指南](docs/plugin-development.md)
- [安全指南](docs/security-guide.md)
- [部署指南](docs/deployment-guide.md)