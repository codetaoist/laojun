# Laojun Swagger 文档生成工具

一个强大的 Swagger/OpenAPI 文档生成工具，专为太上老君平台设计。

## 功能特性

- 🔍 **智能代码扫描** - 自动扫描 Go 源代码并提取 API 信息
- 📝 **多格式输出** - 支持 JSON、YAML、HTML 格式输出
- 🌐 **内置服务器** - 提供 Swagger UI 文档服务器
- ✅ **文档验证** - 验证生成的 Swagger 文档规范
- ⚙️ **灵活配置** - 支持 YAML 配置文件自定义
- 🎨 **现代界面** - 美观的 Swagger UI 界面

## 安装

```bash
cd tools/swagger
go build -o laojun-swagger main.go
```

## 快速开始

### 1. 初始化配置

```bash
./laojun-swagger init
```

这将创建一个默认的 `swagger.yaml` 配置文件。

### 2. 添加 API 注释

在你的 Go 代码中添加 Swagger 注释：

```go
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags users
// @Param id path int true "用户ID"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
    // 实现代码
}
```

### 3. 生成文档

```bash
./laojun-swagger generate
```

### 4. 启动文档服务器

```bash
./laojun-swagger serve
```

访问 http://localhost:8080/swagger-ui 查看文档。

## 命令详解

### generate - 生成文档

```bash
./laojun-swagger generate [flags]
```

**参数：**
- `-c, --config` - 配置文件路径 (默认: swagger.yaml)
- `-o, --output` - 输出文件路径 (默认: ./docs/swagger.json)
- `-f, --format` - 输出格式 (json|yaml|html, 默认: json)

**示例：**
```bash
# 使用默认配置生成 JSON 文档
./laojun-swagger generate

# 生成 YAML 格式文档
./laojun-swagger generate -f yaml -o ./docs/api.yaml

# 使用自定义配置文件
./laojun-swagger generate -c ./config/api-config.yaml
```

### serve - 启动文档服务器

```bash
./laojun-swagger serve [flags]
```

**参数：**
- `-p, --port` - 服务器端口 (默认: 8080)
- `-d, --dir` - 文档目录 (默认: ./docs)

**示例：**
```bash
# 在默认端口启动服务器
./laojun-swagger serve

# 在指定端口启动服务器
./laojun-swagger serve -p 9000

# 指定文档目录
./laojun-swagger serve -d ./api-docs
```

**可用端点：**
- `/` - 静态文件服务
- `/swagger-ui` - Swagger UI 界面
- `/swagger.json` - API 文档 JSON
- `/health` - 健康检查

### validate - 验证文档

```bash
./laojun-swagger validate <file>
```

**示例：**
```bash
# 验证 JSON 文档
./laojun-swagger validate ./docs/swagger.json

# 验证 YAML 文档
./laojun-swagger validate ./docs/api.yaml
```

### init - 初始化配置

```bash
./laojun-swagger init [flags]
```

**参数：**
- `-f, --force` - 强制覆盖已存在的配置文件

## 配置文件

`swagger.yaml` 配置文件示例：

```yaml
source_dir: "./"
output_dir: "./docs"
output_format: "json"
include_files:
  - "**/*.go"
exclude_files:
  - "*_test.go"
  - "vendor/**"
  - "node_modules/**"

info:
  title: "Laojun API"
  description: "太上老君平台 API 文档"
  version: "1.0.0"
  contact:
    name: "Laojun Team"
    email: "team@laojun.com"
  license:
    name: "MIT"
    url: "https://opensource.org/licenses/MIT"

servers:
  - url: "http://localhost:8080"
    description: "开发服务器"
  - url: "https://api.laojun.com"
    description: "生产服务器"

security:
  - bearerAuth: []
```

## API 注释规范

### 基本注释

```go
// @Summary 接口摘要
// @Description 接口详细描述
// @Tags 标签名
// @Accept json
// @Produce json
// @Router /path [method]
```

### 参数注释

```go
// @Param name type dataType required "description"
// @Param id path int true "用户ID"
// @Param name query string false "用户名"
// @Param user body User true "用户信息"
```

**参数类型：**
- `path` - 路径参数
- `query` - 查询参数
- `header` - 请求头参数
- `body` - 请求体参数
- `formData` - 表单参数

### 响应注释

```go
// @Success code {type} model "description"
// @Success 200 {object} User "成功返回用户信息"
// @Success 200 {array} User "成功返回用户列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
```

### 安全注释

```go
// @Security BearerAuth
// @Security ApiKeyAuth
```

## 数据模型

定义数据模型结构体：

```go
// User 用户信息
type User struct {
    ID       int    `json:"id" example:"1"`                    // 用户ID
    Name     string `json:"name" example:"张三"`                // 用户名
    Email    string `json:"email" example:"zhangsan@test.com"` // 邮箱
    Age      int    `json:"age" example:"25"`                  // 年龄
    IsActive bool   `json:"is_active" example:"true"`          // 是否激活
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Code    int    `json:"code" example:"400"`      // 错误码
    Message string `json:"message" example:"错误信息"` // 错误信息
}
```

## 最佳实践

### 1. 组织 API 标签

```go
// @Tags users
// @Tags auth
// @Tags admin
```

### 2. 使用一致的错误响应

```go
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 403 {object} ErrorResponse "禁止访问"
// @Failure 404 {object} ErrorResponse "资源不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
```

### 3. 添加示例数据

```go
type User struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"张三"`
}
```

### 4. 使用描述性的摘要和描述

```go
// @Summary 创建新用户
// @Description 创建一个新的用户账户，需要提供用户名、邮箱等基本信息
```

## 故障排除

### 常见问题

1. **生成的文档为空**
   - 检查源代码目录是否正确
   - 确认 Go 文件中包含 Swagger 注释
   - 检查 include_files 配置

2. **服务器启动失败**
   - 检查端口是否被占用
   - 确认文档目录存在且有读取权限

3. **文档验证失败**
   - 检查 OpenAPI 规范格式
   - 确认必需字段已填写
   - 验证路径格式是否正确

### 调试技巧

1. 使用 `-v` 参数查看详细输出
2. 检查生成的 JSON/YAML 文件格式
3. 使用在线 Swagger 编辑器验证文档

## 集成到 CI/CD

### GitHub Actions 示例

```yaml
name: Generate API Documentation

on:
  push:
    branches: [ main ]

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.19
    
    - name: Build swagger tool
      run: |
        cd tools/swagger
        go build -o laojun-swagger main.go
    
    - name: Generate documentation
      run: |
        ./tools/swagger/laojun-swagger generate
    
    - name: Validate documentation
      run: |
        ./tools/swagger/laojun-swagger validate ./docs/swagger.json
```

## 相关链接

- [OpenAPI 规范](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [Go Swagger 注释指南](https://github.com/swaggo/swag)