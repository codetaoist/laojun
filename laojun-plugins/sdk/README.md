# Laojun Plugin SDK

Laojun Plugin SDK 是一个用于开发 Laojun 插件系统插件的 Go 语言开发工具包。它提供了标准化的接口、工具函数和示例代码，帮助开发者快速构建高质量的插件。

## 特性

- **标准化接口**: 提供统一的插件接口定义
- **多种插件类型**: 支持 HTTP、事件、定时任务、数据处理等多种插件类型
- **配置管理**: 内置配置验证和管理功能
- **客户端支持**: 提供注册中心和事件总线客户端
- **工具函数**: 丰富的工具函数库
- **示例代码**: 完整的示例插件实现
- **构建器模式**: 支持链式调用的插件构建器

## 快速开始

### 安装

```bash
go get github.com/laojun/plugins/sdk
```

### 基础插件

```go
package main

import (
    "context"
    "github.com/laojun/plugins/sdk"
    "github.com/sirupsen/logrus"
)

type MyPlugin struct {
    *sdk.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    base := &sdk.BasePlugin{
        info: &sdk.PluginInfo{
            ID:          "my-plugin",
            Name:        "My Plugin",
            Version:     "1.0.0",
            Description: "My first plugin",
            Author:      "Developer",
            Category:    "utility",
            Tags:        []string{"example"},
        },
        logger: logrus.New(),
    }
    
    return &MyPlugin{BasePlugin: base}
}

func (p *MyPlugin) Initialize(ctx *sdk.PluginContext) error {
    if err := p.BasePlugin.Initialize(ctx); err != nil {
        return err
    }
    
    p.logger.Info("My plugin initialized")
    return nil
}

func main() {
    plugin := NewMyPlugin()
    
    ctx := &sdk.PluginContext{
        PluginID: "my-plugin",
        Config:   make(map[string]interface{}),
    }
    
    if err := plugin.Initialize(ctx); err != nil {
        panic(err)
    }
    
    if err := plugin.Start(); err != nil {
        panic(err)
    }
    
    // 插件运行中...
    
    plugin.Stop()
}
```

## 插件类型

### HTTP 插件

HTTP 插件可以处理 HTTP 请求并提供 Web 服务：

```go
type MyHTTPPlugin struct {
    *sdk.BasePlugin
}

func (p *MyHTTPPlugin) RegisterRoutes(router sdk.Router) {
    router.GET("/hello", p.handleHello)
    router.POST("/data", p.handleData)
}

func (p *MyHTTPPlugin) GetMiddlewares() []sdk.Middleware {
    return []sdk.Middleware{
        p.loggingMiddleware,
        p.authMiddleware,
    }
}

func (p *MyHTTPPlugin) handleHello(ctx *sdk.RequestContext) *sdk.Response {
    return &sdk.Response{
        StatusCode: 200,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       `{"message": "Hello, World!"}`,
    }
}
```

### 事件插件

事件插件可以处理系统事件：

```go
type MyEventPlugin struct {
    *sdk.BasePlugin
}

func (p *MyEventPlugin) HandleEvent(ctx context.Context, event *sdk.Event) error {
    p.logger.WithFields(logrus.Fields{
        "event_id":   event.ID,
        "event_type": event.Type,
        "source":     event.Source,
    }).Info("Event received")
    
    // 处理事件逻辑
    return nil
}

func (p *MyEventPlugin) GetEventTypes() []string {
    return []string{
        "user.created",
        "user.updated",
        "system.alert",
    }
}
```

### 定时任务插件

定时任务插件可以执行周期性任务：

```go
type MyScheduledPlugin struct {
    *sdk.BasePlugin
}

func (p *MyScheduledPlugin) GetSchedule() *sdk.Schedule {
    return &sdk.Schedule{
        Cron:        "0 */10 * * * *", // 每10分钟执行一次
        Timezone:    "UTC",
        MaxRetries:  3,
        RetryDelay:  time.Second * 5,
        Timeout:     time.Minute * 2,
        Description: "Data cleanup task",
    }
}

func (p *MyScheduledPlugin) Execute(ctx context.Context) error {
    p.logger.Info("Executing scheduled task")
    
    // 执行任务逻辑
    
    return nil
}
```

### 数据处理插件

数据处理插件可以处理和转换数据：

```go
type MyDataPlugin struct {
    *sdk.BasePlugin
}

func (p *MyDataPlugin) ProcessData(ctx context.Context, input *sdk.DataInput) (*sdk.DataOutput, error) {
    // 处理输入数据
    var result map[string]interface{}
    
    switch input.Type {
    case "json":
        // 处理 JSON 数据
    case "xml":
        // 处理 XML 数据
    default:
        return nil, fmt.Errorf("unsupported data type: %s", input.Type)
    }
    
    // 返回处理结果
    outputData, _ := json.Marshal(result)
    return &sdk.DataOutput{
        Type: "json",
        Data: outputData,
    }, nil
}

func (p *MyDataPlugin) GetInputSchema() *sdk.DataSchema {
    return &sdk.DataSchema{
        Type: "object",
        Properties: map[string]*sdk.PropertySpec{
            "type": {Type: "string", Required: true},
            "data": {Type: "string", Required: true},
        },
    }
}
```

### 可配置插件

可配置插件支持动态配置管理：

```go
type MyConfigurablePlugin struct {
    *sdk.BasePlugin
    config *MyConfig
}

type MyConfig struct {
    APIKey     string `json:"api_key"`
    Timeout    int    `json:"timeout"`
    EnableSSL  bool   `json:"enable_ssl"`
}

func (p *MyConfigurablePlugin) GetConfigSchema() *sdk.ConfigSchema {
    return &sdk.ConfigSchema{
        Type: "object",
        Properties: map[string]*sdk.ConfigProperty{
            "api_key": {
                Type:        "string",
                Description: "API key for external service",
                Required:    true,
                Sensitive:   true,
            },
            "timeout": {
                Type:        "integer",
                Description: "Request timeout in seconds",
                Default:     30,
                Minimum:     &[]float64{1}[0],
                Maximum:     &[]float64{300}[0],
            },
            "enable_ssl": {
                Type:        "boolean",
                Description: "Enable SSL/TLS",
                Default:     true,
            },
        },
    }
}

func (p *MyConfigurablePlugin) ValidateConfig(config map[string]interface{}) error {
    validator := sdk.NewConfigValidator(p.GetConfigSchema())
    return validator.Validate(config)
}

func (p *MyConfigurablePlugin) UpdateConfig(config map[string]interface{}) error {
    // 验证配置
    if err := p.ValidateConfig(config); err != nil {
        return err
    }
    
    // 更新配置
    if apiKey, ok := config["api_key"].(string); ok {
        p.config.APIKey = apiKey
    }
    
    // ... 更新其他配置项
    
    return nil
}
```

## 构建器模式

使用构建器模式可以更简洁地创建插件：

```go
plugin := sdk.NewPluginBuilder("my-plugin").
    WithName("My Plugin").
    WithVersion("1.0.0").
    WithDescription("A simple plugin").
    WithAuthor("Developer").
    WithCategory("utility").
    WithTags("example", "demo").
    WithHTTPHandler("/api/status", func(ctx *sdk.RequestContext) *sdk.Response {
        return &sdk.Response{
            StatusCode: 200,
            Body:       `{"status": "ok"}`,
        }
    }).
    WithEventHandler("user.login", func(ctx context.Context, event *sdk.Event) error {
        fmt.Printf("User logged in: %s\n", event.Data["user_id"])
        return nil
    }).
    WithSchedule("0 0 * * * *", func(ctx context.Context) error {
        fmt.Println("Daily task executed")
        return nil
    }).
    Build()
```

## 客户端使用

### 注册中心客户端

```go
// 创建注册中心客户端
registryClient := sdk.NewHTTPRegistryClient(
    "http://localhost:8080",
    "your-api-key",
    logger,
)

// 注册插件
registration := &sdk.PluginRegistration{
    ID:          "my-plugin",
    Name:        "My Plugin",
    Version:     "1.0.0",
    Description: "My plugin description",
    Status:      "running",
    Endpoints: []*sdk.Endpoint{
        {
            Name:   "status",
            URL:    "/api/status",
            Method: "GET",
        },
    },
}

err := registryClient.Register(context.Background(), registration)
if err != nil {
    log.Fatal(err)
}

// 发送心跳
err = registryClient.Heartbeat(context.Background(), "my-plugin")
if err != nil {
    log.Error(err)
}

// 发现插件
criteria := &sdk.DiscoveryCriteria{
    Keywords:   []string{"data", "processing"},
    Categories: []string{"utility"},
    MaxResults: 10,
}

plugins, err := registryClient.Discover(context.Background(), criteria)
if err != nil {
    log.Error(err)
} else {
    for _, plugin := range plugins {
        fmt.Printf("Found plugin: %s\n", plugin.Name)
    }
}
```

### 事件总线客户端

```go
// 创建事件总线客户端
eventClient := sdk.NewWebSocketEventBusClient(
    "ws://localhost:8081/events",
    logger,
)

// 发布事件
event := &sdk.Event{
    ID:        "event-123",
    Type:      "user.action",
    Source:    "my-plugin",
    Timestamp: time.Now(),
    Data: map[string]interface{}{
        "user_id": "user-456",
        "action":  "login",
    },
}

err := eventClient.Publish(context.Background(), event)
if err != nil {
    log.Error(err)
}

// 订阅事件
subscription, err := eventClient.Subscribe(
    context.Background(),
    []string{"user.login", "user.logout"},
    func(ctx context.Context, event *sdk.Event) error {
        fmt.Printf("Received event: %s\n", event.Type)
        return nil
    },
)
if err != nil {
    log.Error(err)
}

// 取消订阅
err = eventClient.Unsubscribe(context.Background(), subscription.ID)
if err != nil {
    log.Error(err)
}
```

## 工具函数

SDK 提供了丰富的工具函数：

### 文件工具

```go
fileUtils := sdk.NewFileUtils()

// 检查文件是否存在
if fileUtils.Exists("/path/to/file") {
    fmt.Println("File exists")
}

// 读取 JSON 文件
var config map[string]interface{}
err := fileUtils.ReadJSON("/path/to/config.json", &config)

// 写入 JSON 文件
err = fileUtils.WriteJSON("/path/to/output.json", config)

// 复制文件
err = fileUtils.CopyFile("/src/file", "/dst/file")

// 获取文件哈希
hash, err := fileUtils.GetFileHash("/path/to/file")
```

### 字符串工具

```go
stringUtils := sdk.NewStringUtils()

// 生成随机 ID
id := stringUtils.GenerateID()

// 生成 URL 友好的字符串
slug := stringUtils.Slugify("Hello World!")  // "hello-world"

// 截断字符串
truncated := stringUtils.TruncateString("Long text...", 10)

// 清理字符串
cleaned := stringUtils.SanitizeString("  Text with\twhitespace  ")
```

### 时间工具

```go
timeUtils := sdk.NewTimeUtils()

// 格式化持续时间
formatted := timeUtils.FormatDuration(time.Hour * 2)  // "2.0h"

// 解析持续时间
duration, err := timeUtils.ParseDuration("2h30m")

// 检查是否过期
expired := timeUtils.IsExpired(timestamp, time.Hour)
```

### HTTP 工具

```go
httpUtils := sdk.NewHTTPUtils()

// 解析 Content-Type
mediaType, params := httpUtils.ParseContentType("application/json; charset=utf-8")

// 构建 Content-Type
contentType := httpUtils.BuildContentType("application/json", map[string]string{
    "charset": "utf-8",
})

// 获取客户端 IP
ip := httpUtils.GetClientIP(headers)
```

## 配置验证

SDK 提供了强大的配置验证功能：

```go
// 字符串验证器
stringValidator := &sdk.StringValidator{
    MinLength: 5,
    MaxLength: 50,
    Pattern:   regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
    Required:  true,
}

errors := stringValidator.Validate("test_value")
if errors.HasErrors() {
    for _, err := range errors {
        fmt.Printf("Validation error: %s\n", err.Error())
    }
}

// 数字验证器
numberValidator := &sdk.NumberValidator{
    Min:      &[]float64{0}[0],
    Max:      &[]float64{100}[0],
    Required: true,
}

// 数组验证器
arrayValidator := &sdk.ArrayValidator{
    MinItems: 1,
    MaxItems: 10,
    ItemValidator: stringValidator,
    Required: true,
}
```

## 示例插件

SDK 包含了多个完整的示例插件：

- **ExampleHTTPPlugin**: HTTP 服务插件示例
- **ExampleEventPlugin**: 事件处理插件示例
- **ExampleScheduledPlugin**: 定时任务插件示例
- **ExampleDataPlugin**: 数据处理插件示例
- **ExampleConfigurablePlugin**: 可配置插件示例

查看 `examples.go` 文件获取完整的示例代码。

## 最佳实践

### 1. 错误处理

```go
func (p *MyPlugin) HandleRequest(ctx *sdk.RequestContext) *sdk.Response {
    // 使用结构化日志记录错误
    if err := p.processRequest(ctx); err != nil {
        p.logger.WithError(err).WithFields(logrus.Fields{
            "path":   ctx.Path,
            "method": ctx.Method,
        }).Error("Failed to process request")
        
        return &sdk.Response{
            StatusCode: 500,
            Body:       `{"error": "Internal server error"}`,
        }
    }
    
    return &sdk.Response{StatusCode: 200}
}
```

### 2. 配置管理

```go
func (p *MyPlugin) Initialize(ctx *sdk.PluginContext) error {
    // 验证必需的配置项
    if apiKey, ok := ctx.Config["api_key"].(string); !ok || apiKey == "" {
        return fmt.Errorf("api_key is required")
    }
    
    // 设置默认值
    if timeout, ok := ctx.Config["timeout"].(float64); ok {
        p.timeout = time.Duration(timeout) * time.Second
    } else {
        p.timeout = 30 * time.Second
    }
    
    return p.BasePlugin.Initialize(ctx)
}
```

### 3. 资源管理

```go
func (p *MyPlugin) Start() error {
    if err := p.BasePlugin.Start(); err != nil {
        return err
    }
    
    // 启动资源
    p.ticker = time.NewTicker(time.Minute)
    go p.backgroundTask()
    
    return nil
}

func (p *MyPlugin) Stop() error {
    // 清理资源
    if p.ticker != nil {
        p.ticker.Stop()
    }
    
    return p.BasePlugin.Stop()
}
```

### 4. 并发安全

```go
type MyPlugin struct {
    *sdk.BasePlugin
    mu    sync.RWMutex
    data  map[string]interface{}
}

func (p *MyPlugin) SetData(key string, value interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.data[key] = value
}

func (p *MyPlugin) GetData(key string) interface{} {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.data[key]
}
```

## 测试

### 单元测试

```go
func TestMyPlugin(t *testing.T) {
    plugin := NewMyPlugin()
    
    ctx := &sdk.PluginContext{
        PluginID: "test-plugin",
        Config: map[string]interface{}{
            "api_key": "test-key",
        },
    }
    
    err := plugin.Initialize(ctx)
    assert.NoError(t, err)
    
    err = plugin.Start()
    assert.NoError(t, err)
    
    // 测试插件功能
    
    err = plugin.Stop()
    assert.NoError(t, err)
}
```

### 模拟客户端

```go
func TestWithMockClient(t *testing.T) {
    logger := logrus.New()
    mockClient := sdk.NewMockRegistryClient(logger)
    
    registration := &sdk.PluginRegistration{
        ID:   "test-plugin",
        Name: "Test Plugin",
    }
    
    err := mockClient.Register(context.Background(), registration)
    assert.NoError(t, err)
    
    plugin, err := mockClient.GetPlugin(context.Background(), "test-plugin")
    assert.NoError(t, err)
    assert.Equal(t, "Test Plugin", plugin.Name)
}
```

## 部署

### 构建插件

```bash
# 构建为可执行文件
go build -o my-plugin main.go

# 构建为共享库（Linux/macOS）
go build -buildmode=plugin -o my-plugin.so main.go

# 构建为 Windows DLL
go build -buildmode=c-shared -o my-plugin.dll main.go
```

### 插件配置

创建 `plugin.json` 配置文件：

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "description": "My plugin description",
  "author": "Developer",
  "category": "utility",
  "tags": ["example"],
  "main": "my-plugin",
  "config_schema": {
    "type": "object",
    "properties": {
      "api_key": {
        "type": "string",
        "required": true,
        "sensitive": true
      }
    }
  }
}
```

## 故障排除

### 常见问题

1. **插件无法启动**
   - 检查配置文件格式
   - 验证必需的配置项
   - 查看日志文件

2. **注册失败**
   - 检查注册中心连接
   - 验证 API 密钥
   - 确认插件 ID 唯一性

3. **事件处理异常**
   - 检查事件类型订阅
   - 验证事件处理器实现
   - 查看事件总线连接状态

### 调试技巧

```go
// 启用调试日志
logger.SetLevel(logrus.DebugLevel)

// 添加详细的日志记录
p.logger.WithFields(logrus.Fields{
    "plugin_id": p.info.ID,
    "version":   p.info.Version,
    "config":    p.config,
}).Debug("Plugin configuration loaded")

// 使用性能分析
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。