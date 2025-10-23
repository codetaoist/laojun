# Docker 部署快速指南

## 快速开始

### 1. 本地构建和测试

```bash
# 构建生产镜像
docker build -f Dockerfile.prod -t laojun:latest .

# 运行容器测试
docker run -d -p 80:80 --name laojun-test laojun:latest

# 查看日志
docker logs -f laojun-test

# 停止和清理
docker stop laojun-test
docker rm laojun-test
```

### 2. 使用 Docker Compose 部署

```bash
# 启动所有服务
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f

# 停止服务
docker-compose -f docker-compose.prod.yml down
```

### 3. 自动化部署

```bash
# Linux/macOS
./deploy.sh prod deploy

# Windows PowerShell
.\deploy.ps1 prod deploy
```

## 镜像构建详解

### 多阶段构建流程

```dockerfile
# 阶段1: 前端构建
FROM node:18-alpine AS frontend-builder
# 构建管理后台和市场前端

# 阶段2: 后端构建  
FROM golang:1.21-alpine AS backend-builder
# 构建所有 Go 服务

# 阶段3: 生产镜像
FROM nginx:alpine AS production
# 集成 Nginx + 后端服务 + 前端资源
```

### 构建参数

```bash
# 指定构建目标
docker build --target backend-builder -t laojun-backend .

# 使用构建参数
docker build --build-arg GO_VERSION=1.21 -t laojun .

# 多平台构建
docker buildx build --platform linux/amd64,linux/arm64 -t laojun .
```

## 服务配置

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `APP_ENV` | 运行环境 | `production` |
| `APP_PORT` | 应用端口 | `80` |
| `POSTGRES_DB` | 数据库名 | `laojun` |
| `POSTGRES_USER` | 数据库用户 | `laojun` |
| `POSTGRES_PASSWORD` | 数据库密码 | - |
| `REDIS_PASSWORD` | Redis 密码 | - |
| `JWT_SECRET` | JWT 密钥 | - |

### 数据卷

```yaml
volumes:
  postgres_data:     # 数据库数据
  redis_data:        # Redis 数据  
  uploads:           # 上传文件
  logs:              # 应用日志
```

### 网络配置

```yaml
networks:
  laojun-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 健康检查

### 容器健康检查

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### 手动健康检查

```bash
# 检查应用健康状态
curl http://localhost/health

# 检查数据库连接
docker-compose exec postgres pg_isready

# 检查 Redis 连接
docker-compose exec redis redis-cli ping
```

## 日志管理

### 日志配置

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f laojun-app

# 查看最近100行日志
docker-compose logs --tail=100 laojun-app

# 查看特定时间段日志
docker-compose logs --since="2024-01-01T00:00:00" laojun-app
```

## 数据管理

### 数据备份

```bash
# 备份数据库
docker-compose exec postgres pg_dump -U laojun laojun > backup.sql

# 备份 Redis
docker-compose exec redis redis-cli BGSAVE

# 备份上传文件
docker cp $(docker-compose ps -q laojun-app):/app/uploads ./uploads-backup
```

### 数据恢复

```bash
# 恢复数据库
docker-compose exec -T postgres psql -U laojun laojun < backup.sql

# 恢复上传文件
docker cp ./uploads-backup $(docker-compose ps -q laojun-app):/app/uploads
```

## 性能调优

### 资源限制

```yaml
deploy:
  resources:
    limits:
      memory: 1G
      cpus: '0.5'
    reservations:
      memory: 512M
      cpus: '0.25'
```

### 并发配置

```bash
# Nginx worker 进程数
worker_processes auto;

# Go 服务 GOMAXPROCS
GOMAXPROCS=2

# 数据库连接池
max_connections=100
```

## 故障排除

### 常见问题

#### 1. 容器无法启动

```bash
# 查看容器状态
docker ps -a

# 查看启动日志
docker logs container_name

# 进入容器调试
docker exec -it container_name /bin/sh
```

#### 2. 网络连接问题

```bash
# 查看网络配置
docker network ls
docker network inspect laojun_laojun-network

# 测试容器间连接
docker exec -it container1 ping container2
```

#### 3. 数据持久化问题

```bash
# 查看数据卷
docker volume ls
docker volume inspect volume_name

# 检查挂载点
docker inspect container_name | grep Mounts -A 20
```

### 调试技巧

```bash
# 进入运行中的容器
docker exec -it laojun-app /bin/sh

# 查看容器资源使用
docker stats

# 查看容器进程
docker exec laojun-app ps aux

# 查看容器网络
docker exec laojun-app netstat -tlnp
```

## 安全最佳实践

### 镜像安全

```dockerfile
# 使用非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

# 最小化镜像层
RUN apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*
```

### 运行时安全

```yaml
# 只读根文件系统
read_only: true

# 禁用特权模式
privileged: false

# 限制能力
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE
```

### 网络安全

```yaml
# 内部网络隔离
networks:
  frontend:
    internal: false
  backend:
    internal: true
```

## 监控和告警

### 基础监控

```bash
# 容器状态监控
docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"

# 健康检查状态
docker inspect --format='{{.State.Health.Status}}' container_name
```

### 日志监控

```bash
# 错误日志监控
docker-compose logs | grep -i error

# 访问日志分析
docker-compose logs nginx | grep "GET\|POST"
```

## 扩展部署

### 水平扩展

```bash
# 扩展应用实例
docker-compose up -d --scale laojun-app=3

# 负载均衡配置
upstream backend {
    server laojun-app-1:8080;
    server laojun-app-2:8080;
    server laojun-app-3:8080;
}
```

### 集群部署

```bash
# Docker Swarm 初始化
docker swarm init

# 部署服务栈
docker stack deploy -c docker-compose.prod.yml laojun

# 查看服务状态
docker service ls
```

## 常用命令速查

```bash
# 构建
docker build -f Dockerfile.prod -t laojun .

# 运行
docker-compose -f docker-compose.prod.yml up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 进入容器
docker exec -it laojun-app /bin/sh

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 清理资源
docker system prune -f
```

---

更多详细信息请参考 [DEPLOYMENT.md](./DEPLOYMENT.md)