# Docker 部署指南

本文档详细说明如何使用 Docker 部署太上老君系统。

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐
│   Nginx (80/443) │    │  Grafana (3002) │
│   反向代理        │    │   监控面板       │
└─────────┬───────┘    └─────────────────┘
          │
    ┌─────┴─────┐
    │           │
┌───▼───┐   ┌───▼───┐
│Admin  │   │Market │
│Web    │   │Web    │
│(3000) │   │(3001) │
└───┬───┘   └───┬───┘
    │           │
┌───▼───────────▼───┐
│                   │
│   Config Center   │
│      (8090)       │
└───────┬───────────┘
        │
    ┌───▼───┐   ┌───▼───┐
    │Admin  │   │Market │
    │API    │   │API    │
    │(8080) │   │(8082) │
    └───┬───┘   └───┬───┘
        │           │
    ┌───▼───────────▼───┐
    │                   │
    │  PostgreSQL Redis │
    │   (5432)   (6379) │
    └───────────────────┘
```

## 环境要求

### 服务器要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+ 或 CentOS 8+)
- **内存**: 最少 4GB，推荐 8GB+
- **存储**: 最少 20GB 可用空间
- **CPU**: 最少 2 核，推荐 4 核+

### 软件要求
- Docker 20.10+
- Docker Compose 2.0+

## 快速开始

### 1. 安装 Docker

```bash
# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 启动 Docker 服务
sudo systemctl enable docker
sudo systemctl start docker

# 将用户添加到 docker 组
sudo usermod -aG docker $USER
```

### 2. 构建镜像

```bash
# 构建应用镜像
docker build -t laojun:latest .

# 或使用构建脚本
.\scripts\build.ps1 -Target docker -Environment prod
```

### 3. 使用 Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: laojun
      POSTGRES_USER: laojun
      POSTGRES_PASSWORD: laojun123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U laojun"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  laojun-api:
    image: laojun:latest
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=laojun
      - DB_PASSWORD=laojun123
      - DB_NAME=laojun
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:
```

### 4. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f laojun-api
```

## 生产环境部署

### 1. 环境配置

创建生产环境配置文件 `.env.production`：

```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=laojun_prod
DB_PASSWORD=your_secure_password
DB_NAME=laojun_prod

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# 应用配置
GIN_MODE=release
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# JWT配置
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRES_IN=24h

# 前端API地址配置
VITE_ADMIN_API_URL=https://admin-api.your-domain.com
VITE_MARKETPLACE_API_URL=https://marketplace-api.your-domain.com

# CORS配置
CORS_ALLOW_ORIGINS=https://admin.your-domain.com,https://marketplace.your-domain.com

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=file
LOG_FILE=/var/log/laojun/app.log
```

### 2. SSL证书配置

```bash
# 创建SSL证书目录
mkdir -p deployments/ssl

# 将SSL证书文件放入目录
# - deployments/ssl/cert.pem
# - deployments/ssl/key.pem
```

### 3. 生产环境部署

```bash
# 执行完整部署
./scripts/deploy-production.sh

# 或手动部署
docker-compose -f deployments/docker-compose.prod.yml --env-file .env.production build
docker-compose -f deployments/docker-compose.prod.yml --env-file .env.production up -d
```

### 4. 验证部署

```bash
# 检查服务状态
docker-compose -f deployments/docker-compose.prod.yml ps

# 检查健康状态
curl http://your-server:8080/health
curl http://your-server:8082/health
```

## Nginx 反向代理配置

### Admin API 配置

```nginx
server {
    listen 80;
    server_name admin-api.your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Marketplace API 配置

```nginx
server {
    listen 80;
    server_name marketplace-api.your-domain.com;
    
    location / {
        proxy_pass http://localhost:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8082/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 数据管理

### 数据库备份

```bash
# 备份数据库
docker-compose -f deployments/docker-compose.prod.yml exec postgres pg_dump -U laojun_prod laojun_prod > backup.sql

# 恢复数据库
docker-compose -f deployments/docker-compose.prod.yml exec -T postgres psql -U laojun_prod -d laojun_prod < backup.sql
```

### 日志管理

```bash
# 查看实时日志
docker-compose -f deployments/docker-compose.prod.yml logs -f admin-api

# 查看最近100行日志
docker-compose -f deployments/docker-compose.prod.yml logs --tail=100 admin-api

# 清理Docker日志
docker system prune -f
```

## 监控和维护

### 1. 健康检查

所有服务都配置了健康检查，可以通过以下方式监控：

```bash
# 检查服务健康状态
docker-compose ps
curl http://localhost:8080/health
curl http://localhost:8082/health
```

### 2. 性能监控

建议配置以下监控工具：
- **Prometheus**: 指标收集
- **Grafana**: 可视化监控
- **AlertManager**: 告警通知

### 3. 定期维护

1. **日志轮转**: 配置日志轮转避免磁盘空间不足
2. **定期更新**: 定期更新Docker镜像和系统
3. **备份策略**: 建立定期备份策略

## 故障排除

### 常见问题

1. **容器启动失败**
```bash
# 查看容器日志
docker-compose logs [service-name]

# 检查容器状态
docker ps -a
```

2. **数据库连接失败**
- 检查数据库服务是否正常启动
- 验证数据库连接参数
- 检查网络连接

3. **前端无法访问API**
- 检查API服务状态
- 验证API服务是否正常
- 检查防火墙设置

### 性能优化

```bash
# 查看资源使用情况
docker stats

# 清理未使用的镜像和容器
docker system prune -a
```

### 日志分析

```bash
# 查看错误日志
docker-compose logs | grep ERROR

# 分析访问日志
docker-compose logs nginx | grep "GET\|POST"
```

## 更新部署

### 滚动更新

```bash
# 拉取最新代码
git pull origin main

# 重新构建和部署
docker-compose build
docker-compose up -d
```

### 版本回滚

```bash
# 回滚到指定版本
docker-compose down
docker-compose up -d --scale laojun-api=0
docker tag laojun:v1.0.0 laojun:latest
docker-compose up -d
```

## 技术支持

如果遇到部署问题，请提供以下信息：

1. 操作系统版本
2. Docker和Docker Compose版本
3. 错误日志
4. 服务配置文件

更多技术支持，请访问项目文档或提交Issue。