# Swagger API 文档配置

## 概述

本文档说明如何配置和使用太上老君系统的 Swagger API 文档生成工具。

## 工具位置

Swagger 文档生成工具位于 `tools/swagger/` 目录下，这是一个独立的命令行工具。

## 配置文件

### swagger.yaml 配置示例

```yaml
# Swagger 配置文件
source_dir: "./"
output_dir: "./docs/api"
output_format: "json"  # json, yaml, html
include_files:
  - "**/*.go"
exclude_files:
  - "*_test.go"
  - "vendor/**"
  - "tools/**"

info:
  title: "太上老君 API"
  description: "太上老君系统 RESTful API 文档"
  version: "1.0.0"
  contact:
    name: "太上老君团队"
    email: "team@laojun.com"
  license:
    name: "MIT"
    url: "https://opensource.org/licenses/MIT"

servers:
  - url: "http://localhost:8080"
    description: "开发环境"
  - url: "https://api.laojun.com"
    description: "生产环境"

security:
  - bearerAuth: []
```

## 使用方法

### 1. 生成 API 文档

```bash
# 进入工具目录
cd tools/swagger

# 使用默认配置生成 JSON 格式文档
go run main.go generate

# 指定配置文件和输出格式
go run main.go generate -c swagger.yaml -f json -o ../../docs/api/swagger.json

# 生成 YAML 格式文档
go run main.go generate -f yaml -o ../../docs/api/swagger.yaml

# 生成 HTML 格式文档
go run main.go generate -f html -o ../../docs/api/swagger.html
```

### 2. 验证文档

```bash
# 验证生成的 Swagger 文档
go run main.go validate -f ../../docs/api/swagger.json
```

### 3. 启动文档服务器

```bash
# 启动本地文档服务器
go run main.go serve -p 8081
```

### 4. 初始化配置

```bash
# 生成默认配置文件
go run main.go init
```

## 注释规范

为了正确生成 API 文档，请在 Go 代码中使用以下注释格式：

### 路由注释

```go
// @Summary 用户登录
// @Description 用户通过用户名和密码登录系统
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} LoginResponse "登录成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败"
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
    // 实现代码
}
```

### 模型注释

```go
// LoginRequest 登录请求
type LoginRequest struct {
    Username string `json:"username" binding:"required" example:"admin"`     // 用户名
    Password string `json:"password" binding:"required" example:"password"`  // 密码
}

// LoginResponse 登录响应
type LoginResponse struct {
    Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`      // JWT 令牌
    ExpiresAt int64  `json:"expires_at" example:"1640995200"`              // 过期时间戳
    User      User   `json:"user"`                                         // 用户信息
}
```

## 集成到应用

### 在 Gin 应用中集成 Swagger UI

```go
import (
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func setupSwaggerRoutes(r *gin.Engine) {
    // Swagger 文档路由
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // API 文档重定向
    r.GET("/docs", func(c *gin.Context) {
        c.Redirect(302, "/swagger/index.html")
    })
}
```

### 环境变量配置

```bash
# 开发环境
SWAGGER_ENABLED=true
SWAGGER_HOST=localhost:8080
SWAGGER_BASE_PATH=/api/v1

# 生产环境
SWAGGER_ENABLED=false  # 生产环境建议关闭
```

## 输出文件

生成的文档文件将保存在以下位置：

- `docs/api/swagger.json` - JSON 格式的 OpenAPI 规范
- `docs/api/swagger.yaml` - YAML 格式的 OpenAPI 规范  
- `docs/api/swagger.html` - HTML 格式的可视化文档

## 最佳实践

1. **注释完整性**: 确保所有 API 端点都有完整的注释
2. **模型定义**: 为所有请求和响应模型添加详细注释
3. **示例数据**: 在模型字段中提供示例值
4. **错误处理**: 文档化所有可能的错误响应
5. **版本管理**: 在配置中明确 API 版本信息
6. **安全配置**: 正确配置认证和授权信息

## 故障排除

### 常见问题

1. **文档生成失败**
   - 检查 Go 代码语法是否正确
   - 确认注释格式符合规范
   - 验证配置文件路径

2. **缺少 API 端点**
   - 检查文件包含/排除规则
   - 确认注释中的路由路径正确

3. **模型定义错误**
   - 验证结构体标签格式
   - 检查 JSON 标签是否正确

### 调试模式

```bash
# 启用详细日志输出
go run main.go generate --verbose

# 生成调试信息
go run main.go generate --debug
```

## 相关链接

- [OpenAPI 规范](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [Go Swagger 注释](https://github.com/swaggo/swag)