# 太上老君系统部署指南

## 概述

本文档详细介绍了如何将太上老君系统部署到生产环境，包括 Docker 镜像构建、服务器配置、域名绑定和自动化部署等步骤。

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户浏览器     │────│   Nginx 反向代理  │────│   Docker 容器    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              │                        ├─ Admin API
                              │                        ├─ Config Center
                              │                        ├─ Marketplace API
                              │                        ├─ PostgreSQL
                              │                        └─ Redis
                              │
                       ┌─────────────────┐
                       │   SSL 证书       │
                       │  (Let's Encrypt) │
                       └─────────────────┘
```

## 部署文件说明

### 新的部署目录结构

```
deploy/
├── docker/                    # Docker 相关文件
│   ├── Dockerfile            # 开发环境 Dockerfile
│   ├── Dockerfile.prod       # 生产环境 Dockerfile
│   ├── docker-compose.yml    # 统一服务编排
│   └── README.md            # Docker 使用说明
├── k8s/                      # Kubernetes 配置
│   ├── namespace.yaml
│   ├── postgres.yaml
│   └── redis.yaml
├── nginx/                    # Nginx 配置
│   ├── nginx.conf           # 主配置文件
│   └── conf.d/
│       └── laojun.conf      # 站点配置
├── supervisor/               # Supervisor 配置
│   └── supervisord.conf
├── configs/                  # 配置文件
│   ├── deploy.yaml          # 统一部署配置
│   ├── .env                 # 默认环境配置
│   ├── .env.development     # 开发环境配置
│   ├── .env.staging         # 预发布环境配置
│   └── .env.production      # 生产环境配置
└── docs/                     # 部署文档
    ├── README.md            # 主部署文档
    └── docker-guide.md      # Docker 使用指南
```

### 核心文件

- `deploy/docker/Dockerfile.prod` - 生产环境多阶段构建文件
- `deploy/docker/docker-compose.yml` - 统一服务编排文件
- `deploy/configs/.env.production` - 生产环境配置文件
- `./deploy.ps1` / `./start.ps1` - 自动化部署脚本
- `deploy/configs/deploy.yaml` - 统一部署配置文件

### 配置文件

- `deploy/nginx/nginx.conf` - Nginx 主配置
- `deploy/nginx/conf.d/laojun.conf` - 站点配置
- `deploy/supervisor/supervisord.conf` - 进程管理配置

## 部署步骤

### 1. 服务器环境准备

#### 1.1 系统要求

- **操作系统**: Ubuntu 20.04+ / CentOS 8+ / Debian 11+
- **内存**: 最低 2GB，推荐 4GB+
- **存储**: 最低 20GB，推荐 50GB+
- **网络**: 公网 IP 和域名

#### 1.2 快速环境配置

使用快速配置脚本（推荐）：

```bash
# 下载并运行快速配置脚本
wget https://your-domain.com/quick-deploy.sh
chmod +x quick-deploy.sh
sudo ./quick-deploy.sh your-domain.com admin@your-domain.com
```

#### 1.3 手动环境配置

如果不使用快速配置脚本，请按以下步骤手动配置：

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要工具
sudo apt install -y curl wget git unzip

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装 Nginx
sudo apt install -y nginx

# 安装 Certbot (SSL 证书)
sudo apt install -y certbot python3-certbot-nginx

# 配置防火墙
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. 项目部署

#### 2.1 上传项目代码

```bash
# 创建项目目录
sudo mkdir -p /opt/laojun
sudo chown $USER:$USER /opt/laojun
cd /opt/laojun

# 上传项目代码（选择其中一种方式）

# 方式1: Git 克隆
git clone https://github.com/your-username/laojun.git .

# 方式2: 直接上传压缩包
# 将项目打包上传到服务器并解压
```

#### 2.2 配置环境变量

```bash
# 复制并编辑生产环境配置
cp deploy/configs/.env.prod deploy/configs/.env.prod.local
nano deploy/configs/.env.prod.local
```

重要配置项：

```bash
# 应用配置
APP_ENV=production
APP_DEBUG=false

# 数据库配置
DB_HOST=postgres
DB_NAME=laojun_prod
DB_USER=laojun_prod
DB_PASSWORD=your-secure-db-password

# Redis 配置
REDIS_HOST=redis
REDIS_PASSWORD=your-secure-redis-password

# JWT 配置
JWT_SECRET=your-super-secret-jwt-key

# 域名配置
VITE_ADMIN_API_URL=https://admin-api.your-domain.com
VITE_MARKETPLACE_API_URL=https://marketplace-api.your-domain.com
```

#### 2.3 执行部署

```bash
# 给部署脚本执行权限
chmod +x deploy.sh

# 执行部署 (使用根目录的入口脚本)
./deploy.sh prod deploy

# 或者直接使用 deploy/scripts 目录下的脚本
cd deploy/scripts
chmod +x deploy.sh
./deploy.sh prod deploy
```

### 3. 域名和 SSL 配置

#### 3.1 域名解析

在域名服务商处添加 A 记录：

```
类型: A
主机记录: @
记录值: 你的服务器IP
TTL: 600
```

#### 3.2 SSL 证书配置

```bash
# 自动获取 Let's Encrypt 证书
sudo certbot --nginx -d your-domain.com

# 设置自动续期
sudo crontab -e
# 添加以下行：
# 0 12 * * * /usr/bin/certbot renew --quiet
```

### 4. 验证部署

#### 4.1 检查服务状态

```bash
# 查看容器状态
docker-compose -f docker-compose.prod.yml ps

# 查看服务日志
docker-compose -f docker-compose.prod.yml logs -f

# 健康检查
./deploy.sh prod health
```

#### 4.2 访问测试

- **主站**: https://your-domain.com
- **管理后台**: https://your-domain.com/admin
- **API 文档**: https://your-domain.com/api/docs
- **健康检查**: https://your-domain.com/health

## 运维管理

### 常用命令

```bash
# 查看服务状态
./deploy.sh prod health

# 重启服务
./deploy.sh prod restart

# 查看日志
./deploy.sh prod logs

# 备份数据
./deploy.sh prod backup

# 停止服务
./deploy.sh prod stop

# 清理旧镜像
./deploy.sh prod cleanup
```

### 监控和日志

#### 服务监控

```bash
# 查看系统资源使用
docker stats

# 查看容器详情
docker-compose -f docker-compose.prod.yml ps -a

# 查看网络状态
docker network ls
```

#### 日志管理

```bash
# 查看实时日志
docker-compose -f docker-compose.prod.yml logs -f --tail=100

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f laojun-app

# 日志文件位置
# - Nginx: /var/log/nginx/
# - 应用日志: 容器内 /app/logs/
```

### 数据备份

#### 自动备份

系统已配置自动备份，包括：

- 数据库备份：每天凌晨 2 点
- 文件备份：包含上传文件和配置

#### 手动备份

```bash
# 完整备份
./deploy.sh prod backup

# 仅备份数据库
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U laojun laojun > backup_$(date +%Y%m%d).sql
```

#### 恢复数据

```bash
# 恢复数据库
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U laojun laojun < backup_20240101.sql

# 恢复文件
unzip backup_20240101.zip -d ./restore/
cp -r ./restore/uploads ./
```

### 更新部署

#### 代码更新

```bash
# 拉取最新代码
git pull origin main

# 重新部署
./deploy.sh prod deploy
```

#### 配置更新

```bash
# 修改配置文件
nano .env.production

# 重启服务使配置生效
./deploy.sh prod restart
```

## 性能优化

### 数据库优化

```sql
-- 创建索引（已在迁移文件中包含）
-- 定期分析表
ANALYZE;

-- 清理无用数据
VACUUM;
```

### 缓存优化

- Redis 缓存已配置
- Nginx 静态文件缓存已启用
- 应用层缓存根据需要调整

### 资源限制

在 `docker-compose.prod.yml` 中已配置资源限制：

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

## 故障排除

### 常见问题

#### 1. 容器启动失败

```bash
# 查看详细错误信息
docker-compose -f docker-compose.prod.yml logs laojun-app

# 检查配置文件
docker-compose -f docker-compose.prod.yml config
```

#### 2. 数据库连接失败

```bash
# 检查数据库状态
docker-compose -f docker-compose.prod.yml exec postgres pg_isready

# 检查网络连接
docker-compose -f docker-compose.prod.yml exec laojun-app ping postgres
```

#### 3. 域名访问失败

```bash
# 检查 Nginx 配置
nginx -t

# 重启 Nginx
sudo systemctl restart nginx

# 检查 SSL 证书
certbot certificates
```

#### 4. 性能问题

```bash
# 查看资源使用
docker stats

# 查看系统负载
top
htop

# 查看磁盘使用
df -h
```

### 日志分析

#### 错误日志位置

- **Nginx 错误日志**: `/var/log/nginx/error.log`
- **应用错误日志**: 容器内 `/app/logs/error.log`
- **数据库日志**: 容器内 PostgreSQL 日志

#### 常用日志命令

```bash
# 查看 Nginx 错误日志
sudo tail -f /var/log/nginx/error.log

# 查看应用日志
docker-compose -f docker-compose.prod.yml logs -f laojun-app

# 查看数据库日志
docker-compose -f docker-compose.prod.yml logs -f postgres
```

## 安全配置

### 防火墙配置

```bash
# 只开放必要端口
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### SSL 安全

- 使用 Let's Encrypt 免费 SSL 证书
- 强制 HTTPS 重定向
- 配置安全头部

### 数据库安全

- 使用强密码
- 限制数据库访问
- 定期备份

### 应用安全

- JWT 密钥定期更换
- 输入验证和过滤
- 访问日志记录

## 扩展部署

### 负载均衡

如需处理更高并发，可配置多实例：

```yaml
# docker-compose.prod.yml
services:
  laojun-app-1:
    # ... 配置
  laojun-app-2:
    # ... 配置
```

### 数据库集群

生产环境可考虑：

- PostgreSQL 主从复制
- Redis 集群
- 数据库连接池

### CDN 配置

- 静态资源 CDN 加速
- 图片压缩和优化
- 全球节点分发

## 联系支持

如遇到部署问题，请：

1. 查看本文档的故障排除部分
2. 检查系统日志和错误信息
3. 联系技术支持团队

---

**注意**: 请根据实际情况修改配置文件中的域名、密码等敏感信息，确保生产环境的安全性。