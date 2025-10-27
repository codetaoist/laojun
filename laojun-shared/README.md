# Laojun Shared Library

太上老君共享库 - 一个功能丰富的Go语言工具包，提供缓存管理、工具函数、健康检查、日志记录等核心功能。

## 🚀 特性

- **缓存管理**: 支持内存和Redis缓存，提供统一的缓存接口
- **工具函数**: 丰富的字符串、切片、数字、时间、类型转换等工具函数
- **健康检查**: 灵活的健康检查框架，支持自定义检查器
- **日志记录**: 结构化日志记录，支持多种输出格式和目标
- **JWT认证**: JWT令牌生成和验证功能
- **配置管理**: 统一的配置结构定义
- **验证器**: 数据验证工具

## 📦 安装

```bash
go get github.com/codetaoist/laojun-shared
```

## 🛠️ 快速开始

### 缓存管理

```go
package main

import (
    "context"
    "time"
    "github.com/codetaoist/laojun-shared/cache"
)

func main() {
    // 创建内存缓存
    config := &cache.CacheConfig{
        Type:              cache.CacheTypeMemory,
        DefaultExpiration: time.Minute * 10,
    }
    
    manager, err := cache.NewManager(config, nil)
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // 设置缓存
    err = manager.Set(ctx, "user:1001", "张三", time.Minute*5)
    
    // 获取缓存
    value, err := manager.Get(ctx, "user:1001")
    
    // JSON对象缓存
    user := map[string]interface{}{
        "id": 1001,
        "name": "张三",
    }
    err = manager.SetJSON(ctx, "user:detail:1001", user, time.Minute*10)
}
```

### 工具函数

```go
package main

import (
    "fmt"
    "github.com/codetaoist/laojun-shared/utils"
)

func main() {
    // 字符串工具
    fmt.Println(utils.String.IsEmpty(""))           // true
    fmt.Println(utils.String.Reverse("hello"))      // "olleh"
    fmt.Println(utils.String.ToCamelCase("user_name")) // "userName"
    
    // 切片工具
    slice := []string{"apple", "banana", "cherry"}
    fmt.Println(utils.Slice.Contains(slice, "banana")) // true
    
    // 数字工具
    fmt.Println(utils.Number.Max(10.5, 20.3))       // 20.3
    fmt.Println(utils.Number.IsEven(16))             // true
    
    // 类型转换
    intVal, _ := utils.Convert.ToInt("123")          // 123
    jsonStr, _ := utils.JSON.ToJSON(map[string]string{"key": "value"})
    
    // 验证工具
    fmt.Println(utils.Validate.IsValidEmail("test@example.com")) // true
    
    // 加密工具
    uuid := utils.Crypto.GenerateUUID()
    hash := utils.Crypto.MD5Hash("hello world")
}
```

### 健康检查

```go
package main

import (
    "context"
    "time"
    "github.com/codetaoist/laojun-shared/health"
)

func main() {
    // 创建健康检查器
    h := health.New(health.Config{
        Timeout: time.Second * 5,
    })
    
    // 添加数据库检查
    dbChecker := health.NewCustomChecker("database", func(ctx context.Context) error {
        // 检查数据库连接
        return nil // 或返回错误
    })
    h.AddChecker(dbChecker)
    
    // 执行健康检查
    report := h.Check(context.Background())
    fmt.Printf("健康状态: %s\n", report.Status)
}
```

### 日志记录

```go
package main

import (
    "github.com/codetaoist/laojun-shared/logger"
)

func main() {
    // 创建日志记录器
    config := logger.Config{
        Level:  "info",
        Format: "json",
        Output: "console",
    }
    
    log := logger.New(config)
    
    // 记录不同级别的日志
    log.Info("应用启动", map[string]interface{}{
        "version": "1.0.0",
        "port":    8080,
    })
    
    log.Error("数据库连接失败", map[string]interface{}{
        "database": "mysql",
        "error":    "connection refused",
    })
}
```

## 📚 目录结构

```
laojun-shared/
├── auth/           # JWT认证模块
├── cache/          # 缓存管理
├── config/         # 配置管理
├── health/         # 健康检查
├── logger/         # 日志系统
├── utils/          # 工具函数
├── validator/      # 验证器
├── examples/       # 使用示例
└── test/           # 集成测试
```

## 📖 详细文档

### 模块说明

#### 1. 缓存管理 (cache)
- 支持内存缓存和Redis缓存
- 统一的缓存接口
- JSON对象缓存支持
- 批量操作支持

#### 2. 工具函数 (utils)
- **字符串工具**: 空值检查、截断、反转、命名转换等
- **切片工具**: 包含检查、去重、合并等
- **数字工具**: 最大值、最小值、绝对值、奇偶判断等
- **时间工具**: 格式化、解析等
- **类型转换**: 字符串、数字、布尔值之间的转换
- **JSON工具**: 序列化和反序列化
- **验证工具**: 邮箱、URL验证等
- **加密工具**: UUID生成、哈希计算等
- **分页工具**: 分页计算和管理
- **Map工具**: 键值操作

#### 3. 健康检查 (health)
- 灵活的健康检查框架
- 自定义检查器支持
- 超时控制
- JSON格式报告

#### 4. 日志记录 (logger)
- 结构化日志记录
- 多种输出格式（JSON、文本）
- 多种输出目标（控制台、文件）
- 日志轮转支持

#### 5. JWT认证 (auth)
- JWT令牌生成和验证
- 自定义声明支持
- 过期时间控制

#### 6. 配置管理 (config)
- 统一的配置结构
- 数据库、Redis、JWT等配置

#### 7. 验证器 (validator)
- 数据验证功能
- 自定义验证规则

## 📖 示例

项目提供了完整的使用示例，位于 `examples/` 目录：

- `cache_example.go` - 缓存管理示例
- `utils_example.go` - 工具函数示例
- `health_example.go` - 健康检查示例
- `logger_example.go` - 日志记录示例

运行示例：

```bash
# 缓存示例
go run examples/cache_example.go

# 工具函数示例
go run examples/utils_example.go

# 健康检查示例
go run examples/health_example.go

# 日志记录示例
go run examples/logger_example.go
```

## 🧪 测试

运行所有测试：

```bash
go test ./...
```

运行集成测试：

```bash
go test ./test/
```

## 版本

当前版本：v1.0.0

---

**太上老君共享库** - 让Go开发更简单、更高效！