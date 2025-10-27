# Laojun Discovery - 服务注册与发现

Laojun Discovery 是太上老君微服务架构中的服务注册与发现组件，提供高可用、高性能的服务注册、发现和健康检查功能。

## 核心特性

### 服务注册
- 支持多种服务实例注册方式
- 自动服务健康检查
- 服务元数据管理
- TTL 自动续期机制

### 服务发现
- 实时服务发现
- 负载均衡策略
- 服务过滤和标签匹配
- 健康实例筛选

### 健康检查
- HTTP 健康检查
- TCP 连接检查
- 自定义检查脚本
- 故障自动摘除

### 高可用性
- Redis 持久化存储
- 内存缓存加速
- 集群部署支持
- 故障转移机制

## 目录结构

```
laojun-discovery/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── handlers/         # HTTP 处理器
│   ├── services/         # 业务服务
│   ├── storage/          # 存储层
│   ├── registry/         # 服务注册表
│   └── health/           # 健康检查
├── configs/              # 配置文件
│   └── config.yaml      # 主配置文件
├── docs/                 # 文档
├── deployments/          # 部署配置
└── README.md            # 项目说明
```

## 快速开始

### 环境要求

- Go 1.23+
- Redis 6.0+
- Docker (可选)

### 安装运行

1. **克隆项目**
```bash
git clone https://github.com/codetaoist/laojun-discovery.git
cd laojun-discovery
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置服务**
```bash
# 复制配置文件
cp configs/config.yaml configs/config.local.yaml

# 编辑配置文件
vim configs/config.local.yaml
```

4. **启动 Redis**
```bash
# 使用 Docker 启动 Redis
docker run -d --name redis -p 6379:6379 redis:latest

# 或使用本地 Redis
redis-server
```

5. **启动服务**
```bash
# 开发模式
go run cmd/main.go

# 生产模式
go build -o laojun-discovery cmd/main.go
./laojun-discovery
```

## API 文档

### 服务注册

#### 注册服务
```http
POST /api/v1/registry/register
Content-Type: application/json

{
  "id": "service-001",
  "name": "user-service",
  "address": "192.168.1.100",
  "port": 8080,
  "tags": ["api", "v1"],
  "meta": {
    "version": "1.0.0",
    "environment": "production"
  },
  "health_check": {
    "type": "http",
    "url": "http://192.168.1.100:8080/health",
    "interval": "10s",
    "timeout": "5s"
  }
}
```

#### 注销服务
```http
DELETE /api/v1/registry/deregister/{service_id}
```

#### 更新健康状态
```http
PUT /api/v1/registry/health/{service_id}
Content-Type: application/json

{
  "status": "passing",
  "output": "Service is healthy"
}
```

### 服务发现

#### 发现所有服务
```http
GET /api/v1/discovery/services
```

#### 发现指定服务
```http
GET /api/v1/discovery/services/{service_name}?tag=api&tag=v1&limit=10
```

#### 获取健康实例
```http
GET /api/v1/discovery/services/{service_name}/healthy?limit=5
```

### 健康检查

#### 系统健康状态
```http
GET /health
```

#### 就绪状态检查
```http
GET /ready
```

## 配置说明

### 服务器配置
```yaml
server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 30s
  write_timeout: 30s
```

### 存储配置
```yaml
storage:
  type: "redis"
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

### 健康检查配置
```yaml
health:
  check_interval: 10s
  timeout: 5s
  max_failures: 3
```

## 部署指南

### Docker 部署

1. **构建镜像**
```bash
docker build -t laojun-discovery:latest .
```

2. **运行容器**
```bash
docker run -d \
  --name laojun-discovery \
  -p 8081:8081 \
  -v $(pwd)/configs:/app/configs \
  laojun-discovery:latest
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: laojun-discovery
spec:
  replicas: 3
  selector:
    matchLabels:
      app: laojun-discovery
  template:
    metadata:
      labels:
        app: laojun-discovery
    spec:
      containers:
      - name: laojun-discovery
        image: laojun-discovery:latest
        ports:
        - containerPort: 8081
        env:
        - name: REDIS_HOST
          value: "redis-service"
        - name: REDIS_PORT
          value: "6379"
```

## 监控指标

### Prometheus 指标

- `discovery_services_total` - 注册服务总数
- `discovery_instances_total` - 服务实例总数
- `discovery_health_checks_total` - 健康检查总数
- `discovery_requests_total` - API 请求总数
- `discovery_request_duration_seconds` - 请求处理时间

### 健康检查端点

- `/health` - 系统健康状态
- `/ready` - 服务就绪状态
- `/metrics` - Prometheus 指标

## 开发指南

### 代码结构

- `cmd/` - 应用程序入口点
- `internal/config/` - 配置管理
- `internal/handlers/` - HTTP 请求处理
- `internal/services/` - 业务逻辑
- `internal/storage/` - 数据存储
- `internal/registry/` - 服务注册表
- `internal/health/` - 健康检查

### 开发环境

```bash
# 安装开发工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 代码检查
golangci-lint run

# 运行测试
go test ./...

# 生成文档
go doc -all
```

## 故障排除

### 常见问题

1. **Redis 连接失败**
   - 检查 Redis 服务是否启动
   - 验证连接配置是否正确
   - 检查网络连通性

2. **服务注册失败**
   - 检查请求格式是否正确
   - 验证服务 ID 是否唯一
   - 检查健康检查配置

3. **健康检查失败**
   - 验证健康检查 URL 是否可访问
   - 检查超时配置是否合理
   - 查看服务日志

### 日志分析

```bash
# 查看服务日志
tail -f /var/log/laojun-discovery.log

# 过滤错误日志
grep "ERROR" /var/log/laojun-discovery.log

# 查看健康检查日志
grep "health_check" /var/log/laojun-discovery.log
```

## 性能优化

### 配置优化

1. **Redis 连接池**
```yaml
storage:
  redis:
    pool_size: 20
    min_idle_conns: 10
```

2. **健康检查间隔**
```yaml
health:
  check_interval: 30s
  timeout: 10s
```

3. **缓存配置**
```yaml
registry:
  enable_cache: true
  cache_ttl: 60s
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目主页: https://github.com/codetaoist/laojun-discovery
- 问题反馈: https://github.com/codetaoist/laojun-discovery/issues
- 邮箱: codetaoist@example.com