# Laojun Gateway

Laojun平台的API网关服务，提供统一的入口点、认证、限流、负载均衡和监控功能。

## 功能特性

### 核心功能
- **统一入口**: 所有微服务的统一访问入口
- **服务发现**: 支持Consul和静态配置两种服务发现方式
- **负载均衡**: 支持轮询、随机和加权负载均衡算法
- **认证授权**: 基于JWT的统一认证和授权
- **限流保护**: 多维度限流（全局、用户、IP、路径）
- **监控指标**: 集成Prometheus指标收集
- **健康检查**: 服务健康状态监控
- **请求代理**: 高性能HTTP请求代理转发

### 高级功能
- **熔断器**: 防止级联故障
- **重试机制**: 自动重试失败请求
- **CORS支持**: 跨域资源共享
- **请求追踪**: 分布式请求追踪
- **配置热更新**: 支持配置动态更新

## 目录结构

```
laojun-gateway/
├── cmd/                    # 应用程序入口
│   └── main.go
├── internal/               # 内部包
│   ├── auth/              # 认证服务
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP处理器
│   ├── middleware/        # 中间件
│   ├── proxy/             # 代理服务
│   ├── routes/            # 路由配置
│   └── services/          # 业务服务
│       ├── discovery/     # 服务发现
│       └── ratelimit/     # 限流服务
├── configs/               # 配置文件
│   └── config.yaml
├── docs/                  # 文档
├── deployments/           # 部署配置
├── go.mod                 # Go模块文件
└── README.md
```

## 快速开始

### 环境要求
- Go 1.23+
- Redis 6.0+
- Consul (可选，用于服务发现)

### 安装依赖
```bash
go mod download
```

### 配置文件
复制并修改配置文件：
```bash
cp configs/config.yaml configs/config.local.yaml
```

主要配置项：
- `server`: 服务器配置（端口、超时等）
- `redis`: Redis连接配置
- `discovery`: 服务发现配置
- `auth`: 认证配置（JWT密钥、过期时间等）
- `ratelimit`: 限流配置
- `proxy`: 代理配置（路由规则、负载均衡等）

### 启动服务
```bash
# 开发模式
go run cmd/main.go

# 编译后运行
go build -o gateway cmd/main.go
./gateway
```

### 验证服务
```bash
# 健康检查
curl http://localhost:8080/health

# 指标监控
curl http://localhost:8080/metrics

# 用户登录
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

## API文档

### 认证接口

#### 用户登录
```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

响应：
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

#### 刷新Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### 用户登出
```http
POST /auth/logout
Authorization: Bearer <access_token>
```

### 代理接口

所有以 `/api/` 开头的请求都会被代理到相应的后端服务：

- `/api/admin/*` → admin-api服务
- `/api/config/*` → config-center服务
- `/api/plugins/*` → plugin-service服务
- `/api/public/*` → 公开API（无需认证）

### 监控接口

- `GET /health` - 健康检查
- `GET /metrics` - Prometheus指标

## 配置说明

### 服务发现配置

#### 静态配置
```yaml
discovery:
  type: static
  static:
    admin-api: "localhost:8081"
    config-center: "localhost:8082"
```

#### Consul配置
```yaml
discovery:
  type: consul
  consul:
    address: localhost:8500
    scheme: http
```

### 路由配置
```yaml
proxy:
  routes:
    - path: "/api/admin/*"      # 路径模式
      method: ""                # HTTP方法（空表示所有方法）
      service: "admin-api"      # 目标服务名
      target: ""                # 静态目标地址（可选）
      strip_prefix: false       # 是否移除路径前缀
      auth: true                # 是否需要认证
      headers:                  # 自定义请求头
        X-Gateway: "laojun-gateway"
```

### 限流配置
```yaml
ratelimit:
  enabled: true
  global_rate: 1000    # 全局限流（每分钟请求数）
  user_rate: 100       # 用户限流
  ip_rate: 50          # IP限流
  rules:               # 自定义限流规则
    - path: "/api/admin/*"
      method: ""
      rate: 20
```

## 部署

### Docker部署
```bash
# 构建镜像
docker build -t laojun-gateway .

# 运行容器
docker run -d \
  --name laojun-gateway \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  laojun-gateway
```

### Kubernetes部署
```bash
kubectl apply -f deployments/k8s/
```

## 监控

### Prometheus指标
- `http_requests_total` - HTTP请求总数
- `http_request_duration_seconds` - HTTP请求持续时间
- `active_connections` - 活跃连接数
- `proxy_requests_total` - 代理请求总数
- `proxy_request_duration_seconds` - 代理请求持续时间

### 健康检查
健康检查端点会检查以下组件：
- Redis连接状态
- 服务发现状态
- 整体服务状态

## 开发

### 添加新的中间件
1. 在 `internal/middleware/` 目录创建新的中间件文件
2. 实现中间件函数
3. 在 `internal/routes/routes.go` 中注册中间件

### 添加新的路由
1. 在配置文件中添加路由规则
2. 重启服务或实现热更新

### 扩展负载均衡算法
1. 在 `internal/proxy/balancer.go` 中实现新的负载均衡器
2. 在配置中指定新的算法类型

## 故障排除

### 常见问题

1. **服务发现失败**
   - 检查Consul连接
   - 验证静态配置格式

2. **认证失败**
   - 检查JWT密钥配置
   - 验证token格式和有效期

3. **限流触发**
   - 检查限流配置
   - 查看Redis连接状态

4. **代理失败**
   - 检查后端服务状态
   - 验证路由配置

### 日志级别
通过环境变量设置日志级别：
```bash
export LOG_LEVEL=debug
```

## 贡献

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License