# Docker 生产环境部署指南

本文档详细说明如何将太上老君系统部署到生产服务器上。

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

## 前置要求

### 服务器要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+ 或 CentOS 8+)
- **内存**: 最少 4GB，推荐 8GB+
- **存储**: 最少 20GB 可用空间
- **CPU**: 最少 2 核，推荐 4 核+

### 软件要求
- Docker 20.10+
- Docker Compose 2.0+
- Git

## 部署步骤

### 1. 服务器准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

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

### 2. 获取代码

```bash
# 克隆项目
git clone <your-repository-url> laojun
cd laojun

# 切换到生产分支（如果有）
git checkout production
```

### 3. 配置环境变量

```bash
# 复制环境变量模板
cp .env.prod .env.production

# 编辑环境变量
nano .env.production
```

**重要配置项说明**:

```bash
# 数据库配置 - 请使用强密码
DB_NAME=laojun_prod
DB_USER=laojun_prod
DB_PASSWORD=your_secure_db_password_here

# Redis配置 - 请使用强密码
REDIS_PASSWORD=your_secure_redis_password_here

# JWT密钥 - 请生成一个强随机字符串
JWT_SECRET=your_jwt_secret_key_here_please_change_this_to_a_secure_random_string

# CORS配置 - 根据实际域名修改
CORS_ALLOW_ORIGINS=https://your-domain.com,https://admin.your-domain.com

# 前端API地址配置
VITE_ADMIN_API_URL=https://admin-api.your-domain.com
VITE_MARKETPLACE_API_URL=https://marketplace-api.your-domain.com

# Grafana管理员密码
GRAFANA_PASSWORD=your_secure_grafana_password_here
```

### 4. SSL证书配置（推荐）

如果使用HTTPS，请准备SSL证书：

```bash
# 创建SSL证书目录
mkdir -p deployments/ssl

# 将证书文件放入目录
# cert.pem - 证书文件
# private.key - 私钥文件
```

### 5. 执行部署

#### 使用自动化脚本（推荐）

```bash
# 给脚本执行权限
chmod +x scripts/deploy-production.sh

# 执行完整部署
./scripts/deploy-production.sh deploy
```

#### 手动部署

```bash
# 构建镜像
docker-compose -f deployments/docker-compose.prod.yml --env-file .env.production build

# 启动服务
docker-compose -f deployments/docker-compose.prod.yml --env-file .env.production up -d
```

### 6. 验证部署

```bash
# 查看服务状态
./scripts/deploy-production.sh status

# 或手动检查
docker-compose -f deployments/docker-compose.prod.yml ps
```

**健康检查**:
- Config Center: http://your-server:8090/health
- Admin API: http://your-server:8080/health
- Marketplace API: http://your-server:8082/health

## 域名和反向代理配置

### Nginx配置示例

如果使用外部Nginx，参考配置：

```nginx
# /etc/nginx/sites-available/laojun
server {
    listen 80;
    server_name your-domain.com admin.your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/private.key;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/private.key;
    
    location / {
        proxy_pass http://localhost:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /api/ {
        proxy_pass http://localhost:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 常用运维命令

### 服务管理

```bash
# 查看服务状态
./scripts/deploy-production.sh status

# 查看日志
./scripts/deploy-production.sh logs

# 重启服务
./scripts/deploy-production.sh restart

# 停止服务
./scripts/deploy-production.sh stop

# 仅构建镜像
./scripts/deploy-production.sh build
```

### 数据库管理

```bash
# 进入数据库容器
docker-compose -f deployments/docker-compose.prod.yml exec postgres psql -U laojun_prod -d laojun_prod

# 备份数据库
docker-compose -f deployments/docker-compose.prod.yml exec postgres pg_dump -U laojun_prod laojun_prod > backup.sql

# 恢复数据库
docker-compose -f deployments/docker-compose.prod.yml exec -T postgres psql -U laojun_prod -d laojun_prod < backup.sql
```

### 日志管理

```bash
# 查看特定服务日志
docker-compose -f deployments/docker-compose.prod.yml logs -f admin-api

# 查看最近100行日志
docker-compose -f deployments/docker-compose.prod.yml logs --tail=100 admin-api

# 清理日志
docker system prune -f
```

## 监控和告警

### Grafana访问
- URL: http://your-server:3002
- 默认用户名: admin
- 密码: 在环境变量中配置的 GRAFANA_PASSWORD

### Prometheus访问
- URL: http://your-server:9090

## 安全建议

1. **防火墙配置**: 只开放必要的端口（80, 443）
2. **定期更新**: 定期更新Docker镜像和系统
3. **密码安全**: 使用强密码，定期更换
4. **SSL证书**: 使用HTTPS加密传输
5. **备份策略**: 定期备份数据库和重要文件
6. **监控告警**: 配置系统监控和告警

## 故障排除

### 常见问题

1. **容器启动失败**
   ```bash
   # 查看容器日志
   docker-compose -f deployments/docker-compose.prod.yml logs [service-name]
   ```

2. **数据库连接失败**
   - 检查数据库容器是否正常运行
   - 验证环境变量配置
   - 检查网络连接

3. **前端无法访问API**
   - 检查CORS配置
   - 验证API服务是否正常
   - 检查网络代理配置

4. **内存不足**
   ```bash
   # 查看系统资源使用
   docker stats
   
   # 清理未使用的镜像和容器
   docker system prune -a
   ```

### 性能优化

1. **数据库优化**
   - 调整PostgreSQL配置参数
   - 定期执行VACUUM和ANALYZE
   - 监控慢查询

2. **缓存优化**
   - 合理配置Redis内存限制
   - 监控缓存命中率

3. **应用优化**
   - 调整Go应用的GOMAXPROCS
   - 配置合适的连接池大小

## 更新部署

```bash
# 拉取最新代码
git pull origin production

# 重新构建和部署
./scripts/deploy-production.sh deploy
```

## 回滚操作

```bash
# 回滚到上一个版本
git checkout <previous-commit-hash>
./scripts/deploy-production.sh deploy
```

## 联系支持

如果遇到部署问题，请提供以下信息：
- 服务器配置信息
- 错误日志
- 环境变量配置（隐藏敏感信息）
- Docker和Docker Compose版本