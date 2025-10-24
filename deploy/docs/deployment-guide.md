# 太上老君系统部署指南

## 概述

太上老君系统是一个基于微服务架构的插件化管理系统，支持Docker容器化部署。本指南将帮助您快速部署和管理系统。

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐
│   插件市场Web    │    │   管理后台Web    │
│   (Port: 80)    │    │  (Port: 8888)   │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
            ┌─────────────────┐
            │   Nginx反向代理   │
            │   (Port: 80)    │
            └─────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────────┐ ┌─────────────┐ ┌─────────────┐
│ 管理API    │ │ 插件市场API  │ │  配置中心    │
│(Port:8080)│ │(Port: 8082) │ │(Port: 8081) │
└───────────┘ └─────────────┘ └─────────────┘
        │            │            │
        └────────────┼────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────────┐ ┌─────────────┐ ┌─────────────┐
│PostgreSQL │ │   Redis     │ │   MinIO     │
│(Port:5432)│ │(Port: 6379) │ │(Port: 9000) │
└───────────┘ └─────────────┘ └─────────────┘
```

## 快速开始

### 方式一：一键启动（推荐）

```powershell
# Windows PowerShell
cd d:\laojun\deploy\scripts
.\one-click-deploy.ps1

# 或指定环境
.\one-click-deploy.ps1           # 开发环境（默认）
.\one-click-deploy.ps1 -Production   # 生产环境
.\one-click-deploy.ps1 -Clean        # 清理重新部署
```

```bash
# Linux/macOS
cd /path/to/laojun/deploy/scripts
chmod +x deploy.sh
./deploy.sh start
```

### 方式二：手动部署

#### 1. 环境准备

**系统要求：**
- Docker 20.10+
- Docker Compose 2.0+
- 可用内存：至少 4GB
- 可用磁盘：至少 10GB

**端口要求：**
- 80: Nginx (插件市场)
- 8080: 管理API
- 8081: 配置中心
- 8082: 插件市场API
- 8888: 管理后台
- 5432: PostgreSQL
- 6379: Redis
- 9000: MinIO

#### 2. 配置环境变量

```powershell
# 复制环境配置模板
cd d:\laojun\deploy\configs
copy .env.template .env

# 编辑配置文件
notepad .env
```

**重要配置项：**

```env
# 基础环境
APP_ENV=production
APP_DEBUG=false

# 数据库配置
POSTGRES_DB=laojun
POSTGRES_USER=laojun
POSTGRES_PASSWORD=your-secure-password

# Redis配置
REDIS_PASSWORD=your-redis-password

# JWT密钥（必须修改）
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# 安全配置
SECURITY_ENABLE_CAPTCHA=true
SECURITY_CAPTCHA_TTL=2m
```

#### 3. 构建和启动

```powershell
# 进入Docker目录
cd d:\laojun\deploy\docker

# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 检查服务状态
docker-compose ps
```

#### 4. 数据库初始化

```powershell
# 执行数据库迁移
cd d:\laojun
make migrate-complete-up

# 或手动执行
go run cmd/db-complete-migrate/main.go up
```

## 部署脚本使用

### Windows PowerShell 脚本

```powershell
# 基本命令
.\deploy.ps1 check      # 检查环境
.\deploy.ps1 build      # 构建镜像
.\deploy.ps1 start      # 启动服务
.\deploy.ps1 stop       # 停止服务
.\deploy.ps1 restart    # 重启服务
.\deploy.ps1 status     # 检查状态

# 日志管理
.\deploy.ps1 logs                # 查看所有日志
.\deploy.ps1 logs admin-api      # 查看特定服务日志

# 数据管理
.\deploy.ps1 migrate    # 数据库迁移
.\deploy.ps1 backup     # 备份数据
.\deploy.ps1 update     # 更新系统

# 系统维护
.\deploy.ps1 cleanup    # 清理系统
```

### Linux/macOS Bash 脚本

```bash
# 基本命令
./deploy.sh check      # 检查环境
./deploy.sh build      # 构建镜像
./deploy.sh start      # 启动服务
./deploy.sh stop       # 停止服务
./deploy.sh restart    # 重启服务
./deploy.sh status     # 检查状态

# 日志管理
./deploy.sh logs                # 查看所有日志
./deploy.sh logs admin-api      # 查看特定服务日志

# 数据管理
./deploy.sh migrate    # 数据库迁移
./deploy.sh backup     # 备份数据
./deploy.sh update     # 更新系统

# 系统维护
./deploy.sh cleanup    # 清理系统
```

## 环境配置详解

### 开发环境 (.env.development)

适用于本地开发和测试：

```env
APP_ENV=development
APP_DEBUG=true
LOG_LEVEL=debug

# 使用默认密码（仅开发环境）
POSTGRES_PASSWORD=laojun123
REDIS_PASSWORD=redis123
JWT_SECRET=dev-jwt-secret-key

# 开发配置
CORS_ALLOW_ORIGINS=*
SECURITY_ENABLE_CAPTCHA=false
```

### 生产环境 (.env.production)

适用于生产部署：

```env
APP_ENV=production
APP_DEBUG=false
LOG_LEVEL=info

# 强密码（必须修改）
POSTGRES_PASSWORD=your-super-secure-db-password
REDIS_PASSWORD=your-super-secure-redis-password
JWT_SECRET=your-super-secret-jwt-key-at-least-32-chars

# 生产安全配置
CORS_ALLOW_ORIGINS=https://yourdomain.com
SECURITY_ENABLE_CAPTCHA=true
SECURITY_RATE_LIMIT_ENABLED=true
```

## 服务管理

### 服务列表

| 服务名 | 端口 | 描述 | 健康检查 |
|--------|------|------|----------|
| nginx | 80, 8888 | 反向代理和静态文件服务 | http://localhost/health |
| admin-api | 8080 | 管理后台API | http://localhost:8080/health |
| marketplace-api | 8082 | 插件市场API | http://localhost:8082/health |
| config-center | 8081 | 配置中心 | http://localhost:8081/health |
| postgres | 5432 | PostgreSQL数据库 | pg_isready |
| redis | 6379 | Redis缓存 | redis-cli ping |

### 服务操作

```powershell
# 查看特定服务状态
docker-compose ps admin-api

# 重启特定服务
docker-compose restart admin-api

# 查看服务日志
docker-compose logs -f admin-api

# 进入服务容器
docker-compose exec admin-api sh

# 查看服务资源使用
docker stats
```

## 数据管理

### 数据库迁移

```powershell
# 查看迁移状态
go run cmd/db-complete-migrate/main.go status

# 执行迁移
go run cmd/db-complete-migrate/main.go up

# 重置数据库
go run cmd/db-complete-migrate/main.go reset
```

### 数据备份

```powershell
# 自动备份（推荐）
.\deploy.ps1 backup

# 手动备份数据库
docker-compose exec postgres pg_dump -U laojun laojun > backup.sql

# 手动备份Redis
docker-compose exec redis redis-cli --rdb /data/dump.rdb

# 备份上传文件
copy-item uploads backup/uploads -Recurse
```

### 数据恢复

```powershell
# 恢复数据库
docker-compose exec -T postgres psql -U laojun laojun < backup.sql

# 恢复Redis
docker-compose exec redis redis-cli --rdb /data/dump.rdb

# 恢复上传文件
copy-item backup/uploads uploads -Recurse
```

## 监控和日志

### 日志管理

```powershell
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f admin-api

# 查看最近的日志
docker-compose logs --tail=100 admin-api

# 导出日志
docker-compose logs admin-api > admin-api.log
```

### 性能监控

```powershell
# 查看容器资源使用
docker stats

# 查看系统资源
docker system df

# 查看网络状态
docker network ls
docker network inspect laojun_laojun-network
```

### 健康检查

```powershell
# 检查所有服务健康状态
.\deploy.ps1 status

# 手动检查API健康状态
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
```

## 故障排除

### 常见问题

#### 1. 端口被占用

**问题：** 启动时提示端口被占用

**解决：**
```powershell
# 查看端口占用
netstat -ano | findstr :80
netstat -ano | findstr :8080

# 结束占用进程
taskkill /PID <进程ID> /F
```

#### 2. 数据库连接失败

**问题：** 应用无法连接数据库

**解决：**
```powershell
# 检查数据库服务状态
docker-compose ps postgres

# 查看数据库日志
docker-compose logs postgres

# 重启数据库服务
docker-compose restart postgres

# 检查数据库连接
docker-compose exec postgres psql -U laojun -d laojun -c "SELECT 1;"
```

#### 3. 内存不足

**问题：** 服务启动失败，提示内存不足

**解决：**
```powershell
# 清理未使用的镜像和容器
docker system prune -f

# 调整Docker内存限制
# 在Docker Desktop设置中增加内存分配

# 减少并发服务数量
docker-compose up -d postgres redis  # 先启动基础服务
docker-compose up -d config-center   # 再启动应用服务
```

#### 4. 镜像构建失败

**问题：** Docker镜像构建失败

**解决：**
```powershell
# 清理构建缓存
docker builder prune -f

# 重新构建镜像
docker-compose build --no-cache

# 检查Dockerfile语法
docker build -t test-image -f Dockerfile.admin-api .
```

### 日志分析

#### 应用日志位置

```
logs/
├── admin-api/          # 管理API日志
├── marketplace-api/    # 插件市场API日志
├── config-center/      # 配置中心日志
├── nginx/             # Nginx日志
├── postgres/          # PostgreSQL日志
└── redis/             # Redis日志
```

#### 常用日志命令

```powershell
# 实时查看错误日志
docker-compose logs -f | findstr ERROR

# 查看特定时间段日志
docker-compose logs --since="2024-01-01T00:00:00" admin-api

# 搜索特定关键词
docker-compose logs admin-api | findstr "database"
```

## 性能优化

### 数据库优化

```sql
-- 查看数据库性能
SELECT * FROM pg_stat_activity;

-- 查看慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

### Redis优化

```bash
# 查看Redis内存使用
docker-compose exec redis redis-cli info memory

# 查看Redis性能统计
docker-compose exec redis redis-cli info stats
```

### 容器资源限制

```yaml
# docker-compose.yml 中的资源限制示例
services:
  admin-api:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
```

## 安全配置

### SSL/TLS 配置

```nginx
# nginx.conf SSL配置示例
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    ssl_certificate /etc/ssl/certs/yourdomain.crt;
    ssl_certificate_key /etc/ssl/private/yourdomain.key;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;
}
```

### 防火墙配置

```powershell
# Windows防火墙规则
netsh advfirewall firewall add rule name="Laojun HTTP" dir=in action=allow protocol=TCP localport=80
netsh advfirewall firewall add rule name="Laojun HTTPS" dir=in action=allow protocol=TCP localport=443
netsh advfirewall firewall add rule name="Laojun Admin" dir=in action=allow protocol=TCP localport=8888
```

### 密码安全

```env
# 强密码要求
POSTGRES_PASSWORD=至少16位，包含大小写字母、数字、特殊字符
REDIS_PASSWORD=至少16位，包含大小写字母、数字、特殊字符
JWT_SECRET=至少32位随机字符串
```

## 更新和维护

### 系统更新

```powershell
# 自动更新（包含备份）
.\deploy.ps1 update

# 手动更新步骤
.\deploy.ps1 backup     # 1. 备份数据
.\deploy.ps1 stop       # 2. 停止服务
git pull                # 3. 更新代码
.\deploy.ps1 build      # 4. 重新构建
.\deploy.ps1 migrate    # 5. 数据库迁移
.\deploy.ps1 start      # 6. 启动服务
```

### 定期维护

```powershell
# 每周维护脚本
# 1. 清理日志
Get-ChildItem logs -Recurse | Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-7)} | Remove-Item

# 2. 清理Docker
docker system prune -f

# 3. 备份数据
.\deploy.ps1 backup

# 4. 检查服务状态
.\deploy.ps1 status
```

## 扩展部署

### 多实例部署

```yaml
# docker-compose.scale.yml
services:
  admin-api:
    deploy:
      replicas: 3
  marketplace-api:
    deploy:
      replicas: 3
```

```powershell
# 启动多实例
docker-compose -f docker-compose.yml -f docker-compose.scale.yml up -d
```

### 负载均衡

```nginx
# nginx负载均衡配置
upstream admin_api {
    server admin-api-1:8080;
    server admin-api-2:8080;
    server admin-api-3:8080;
}

server {
    location /api/ {
        proxy_pass http://admin_api;
    }
}
```

## 联系支持

如果您在部署过程中遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查系统日志：`docker-compose logs`
3. 查看GitHub Issues
4. 联系技术支持

---

**注意：** 本文档会持续更新，请定期查看最新版本。