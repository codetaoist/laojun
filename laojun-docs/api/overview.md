# API 概述

本文档提供太上老君系统 API 的总体概述和使用指南。

## API 架构

太上老君系统采用 RESTful API 设计，提供统一的接口规范和响应格式。

### 基础信息

- **Base URL**: `https://api.laojun.dev/v1`
- **协议**: HTTPS
- **数据格式**: JSON
- **认证方式**: JWT Bearer Token
- **API 版本**: v1.2.0

### 认证机制

所有 API 请求都需要通过 JWT Token 进行认证：

```http
Authorization: Bearer <your_jwt_token>
```

获取 Token 的方式：
1. 用户登录接口获取访问令牌
2. 使用刷新令牌延长会话
3. 管理员可以生成 API Key

## 认证

### 获取访问令牌

```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

### 使用访问令牌

在请求头中包含访问令牌：

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 响应格式

所有 API 响应都遵循统一格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "req-123456789"
}
```

### 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 数据验证失败 |
| 500 | 服务器内部错误 |

## 分页

支持分页的 API 使用以下参数：

- `page`: 页码，从 1 开始 (默认: 1)
- `page_size`: 每页数量 (默认: 20，最大: 100)
- `sort`: 排序字段 (默认: id)
- `order`: 排序方向，asc 或 desc (默认: desc)

**示例**:
```http
GET /users?page=1&page_size=20&sort=created_at&order=desc
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## 用户管理

### 用户模型

```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "nickname": "管理员",
  "avatar": "https://example.com/avatar.jpg",
  "status": "active",
  "roles": ["admin"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 获取用户列表

```http
GET /users
```

**查询参数**:
- `username`: 用户名筛选
- `email`: 邮箱筛选
- `status`: 状态筛选 (active, inactive, banned)
- `role`: 角色筛选

### 获取用户详情

```http
GET /users/{id}
```

### 创建用户

```http
POST /users
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123",
  "nickname": "新用户",
  "roles": ["user"]
}
```

### 更新用户

```http
PUT /users/{id}
Content-Type: application/json

{
  "nickname": "更新的昵称",
  "email": "updated@example.com",
  "status": "active"
}
```

### 删除用户

```http
DELETE /users/{id}
```

### 修改密码

```http
POST /users/{id}/password
Content-Type: application/json

{
  "old_password": "oldpassword",
  "new_password": "newpassword"
}
```

## 角色权限管理

### 角色模型

```json
{
  "id": 1,
  "name": "admin",
  "display_name": "管理员",
  "description": "系统管理员角色",
  "permissions": ["user:read", "user:write", "system:admin"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 获取角色列表

```http
GET /roles
```

### 创建角色

```http
POST /roles
Content-Type: application/json

{
  "name": "editor",
  "display_name": "编辑者",
  "description": "内容编辑角色",
  "permissions": ["content:read", "content:write"]
}
```

### 分配角色权限

```http
POST /roles/{id}/permissions
Content-Type: application/json

{
  "permissions": ["user:read", "user:write"]
}
```

## 系统配置

### 获取配置

```http
GET /config
```

### 更新配置

```http
PUT /config
Content-Type: application/json

{
  "app_name": "Laojun",
  "app_version": "1.0.0",
  "maintenance_mode": false,
  "max_upload_size": 10485760
}
```

## 文件上传

### 上传文件

```http
POST /upload
Content-Type: multipart/form-data

file: <binary data>
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "file_id": "file-123456",
    "filename": "document.pdf",
    "size": 1024000,
    "mime_type": "application/pdf",
    "url": "https://example.com/files/document.pdf",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取文件信息

```http
GET /files/{file_id}
```

### 删除文件

```http
DELETE /files/{file_id}
```

## 监控和健康检查

### 健康检查

```http
GET /health
```

**响应**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "uptime": "24h30m15s"
}
```

### 详细健康检查

```http
GET /health/detailed
```

**响应**:
```json
{
  "status": "healthy",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time": "5ms"
    },
    "cache": {
      "status": "healthy",
      "response_time": "2ms"
    },
    "external_api": {
      "status": "degraded",
      "response_time": "500ms"
    }
  }
}
```

### 系统指标

```http
GET /metrics
```

## 日志管理

### 获取日志

```http
GET /logs
```

**查询参数**:
- `level`: 日志级别 (debug, info, warn, error)
- `start_time`: 开始时间 (RFC3339 格式)
- `end_time`: 结束时间 (RFC3339 格式)
- `keyword`: 关键词搜索

### 获取操作日志

```http
GET /audit-logs
```

## 插件管理

### 获取插件列表

```http
GET /plugins
```

### 安装插件

```http
POST /plugins
Content-Type: application/json

{
  "name": "auth-plugin",
  "version": "1.0.0",
  "source": "registry"
}
```

### 启用/禁用插件

```http
PUT /plugins/{name}/status
Content-Type: application/json

{
  "enabled": true
}
```

### 插件配置

```http
PUT /plugins/{name}/config
Content-Type: application/json

{
  "config": {
    "timeout": 30,
    "retry_count": 3
  }
}
```

## WebSocket API

### 连接

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

// 认证
ws.send(JSON.stringify({
  type: 'auth',
  token: 'your-jwt-token'
}));
```

### 消息格式

```json
{
  "type": "message_type",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 支持的消息类型

- `auth`: 认证
- `ping`: 心跳
- `notification`: 通知
- `log`: 实时日志
- `metric`: 实时指标

## 错误处理

### 错误响应格式

```json
{
  "code": 400,
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "Invalid email format"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "req-123456789"
}
```

### 常见错误码

| 错误码 | 说明 |
|--------|------|
| 10001 | 参数验证失败 |
| 10002 | 资源不存在 |
| 10003 | 权限不足 |
| 10004 | 认证失败 |
| 10005 | 资源冲突 |
| 20001 | 数据库错误 |
| 20002 | 缓存错误 |
| 30001 | 外部服务错误 |

## 限流

API 实施了限流策略：

- **全局限流**: 每秒 1000 请求
- **用户限流**: 每用户每分钟 100 请求
- **IP 限流**: 每 IP 每分钟 200 请求

当达到限流阈值时，API 返回 429 状态码：

```json
{
  "code": 429,
  "message": "Rate limit exceeded",
  "retry_after": 60
}
```

## SDK 和客户端库

### Go SDK

```go
import "github.com/codetaoist/laojun-go-sdk"

client := laojun.NewClient("http://localhost:8080", "your-api-key")
users, err := client.Users.List(context.Background(), &laojun.ListUsersOptions{
    Page:     1,
    PageSize: 20,
})
```

### JavaScript SDK

```javascript
import { LaojunClient } from '@laojun/js-sdk';

const client = new LaojunClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key'
});

const users = await client.users.list({ page: 1, pageSize: 20 });
```

## 测试

### 使用 curl 测试

```bash
# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 获取用户列表
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer your-token"
```

### 使用 Postman

导入 Postman 集合文件: [laojun-api.postman_collection.json](../postman/laojun-api.postman_collection.json)

## 版本控制

API 使用语义化版本控制：

- **主版本号**: 不兼容的 API 修改
- **次版本号**: 向下兼容的功能性新增
- **修订号**: 向下兼容的问题修正

当前版本: `v1.0.0`

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 用户管理 API
- 角色权限 API
- 文件上传 API
- 监控健康检查 API

---

## 开发工具

### Swagger 文档生成

太上老君系统提供了强大的 Swagger 文档生成工具，位于 `tools/swagger/` 目录。

**快速使用**:
```bash
# 生成 API 文档
cd tools/swagger
./laojun-swagger generate

# 启动文档服务器
./laojun-swagger serve
```

**详细配置**: 参考 [Swagger 配置指南](api/swagger/README.md)

### 在线文档

- **开发环境**: http://localhost:8080/swagger-ui
- **API 规范**: http://localhost:8080/swagger.json

更多详细信息请参考 [Swagger 配置文档](api/swagger/README.md)。