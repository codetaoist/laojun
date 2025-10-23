# Laojun 系统部署指南

## 🏗️ 架构概览

Laojun 系统采用微服务架构，包含以下组件：

### 前端服务
- **插件市场** (端口 80) - 默认首页，展示和管理插件
- **管理后台** (端口 8888) - 系统管理界面

### 后端服务
- **配置中心** (端口 8081) - 统一配置管理
- **管理API** (端口 8080) - 管理后台API服务
- **插件市场API** (端口 8082) - 插件市场API服务

### 基础设施
- **PostgreSQL** - 主数据库
- **Redis** - 缓存和会话存储
- **Nginx** - 反向代理和静态文件服务

## 🚀 快速部署

### 1. 环境准备

```bash
# 确保已安装 Docker 和 Docker Compose
docker --version
docker-compose --version

# 克隆项目
git clone <repository-url>
cd laojun
```

### 2. 配置环境变量

```bash
# 复制环境配置文件
cp deploy/.env.example deploy/.env

# 编辑配置文件（重要：修改生产环境密码）
nano deploy/.env
```

### 3. 启动服务

```bash
# 进入部署目录
cd deploy/docker

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 验证部署

- 插件市场：http://localhost
- 管理后台：http://localhost:8888
- 健康检查：http://localhost/health

## 📁 目录结构

```
deploy/
├── docker/
│   ├── docker-compose.yml          # 主要部署配置
│   ├── Dockerfile.nginx            # Nginx 镜像构建
│   ├── Dockerfile.admin-api        # 管理API镜像构建
│   ├── Dockerfile.marketplace-api  # 插件市场API镜像构建
│   └── Dockerfile.config-center    # 配置中心镜像构建
├── nginx/
│   └── conf.d/
│       └── laojun.conf             # Nginx 统一配置
├── .env.example                    # 环境配置模板
└── README.md                       # 本文档
```

## 🔧 配置说明

### Nginx 配置

- **端口 80**: 插件市场前端 + API代理
- **端口 8888**: 管理后台前端 + API代理
- 支持 SPA 路由
- 静态资源缓存优化
- 安全头配置

### 服务依赖

```
nginx → admin-api → postgres, redis, config-center
nginx → marketplace-api → postgres, redis, config-center
```

### 健康检查

所有服务都配置了健康检查：
- 数据库连接检查
- API服务响应检查
- Redis连接检查

## 🛠️ 运维操作

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs nginx
docker-compose logs admin-api
docker-compose logs marketplace-api

# 实时跟踪日志
docker-compose logs -f
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart nginx
```

### 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

### 数据备份

```bash
# 备份 PostgreSQL
docker-compose exec postgres pg_dump -U laojun laojun > backup.sql

# 备份 Redis
docker-compose exec redis redis-cli --rdb /data/dump.rdb
```

## 🔒 安全配置

### 生产环境必须修改的配置

1. **JWT_SECRET**: 更改为强密码
2. **POSTGRES_PASSWORD**: 更改数据库密码
3. **REDIS_PASSWORD**: 更改Redis密码
4. **CORS_ALLOW_ORIGINS**: 限制允许的域名

### 防火墙配置

```bash
# 只开放必要端口
ufw allow 80/tcp    # 插件市场
ufw allow 8888/tcp  # 管理后台
ufw allow 22/tcp    # SSH
```

## 📊 监控和性能

### 资源限制

- PostgreSQL: 512MB 内存限制
- Redis: 256MB 内存限制
- 其他服务根据需要调整

### 性能优化

1. **数据库连接池**: 配置合适的连接数
2. **Redis缓存**: 启用适当的缓存策略
3. **Nginx缓存**: 静态资源长期缓存
4. **Gzip压缩**: 减少传输大小

## 🐛 故障排除

### 常见问题

1. **端口冲突**: 检查端口是否被占用
2. **权限问题**: 确保Docker有足够权限
3. **网络问题**: 检查防火墙和网络配置
4. **资源不足**: 检查内存和磁盘空间

### 调试命令

```bash
# 检查容器状态
docker-compose ps

# 进入容器调试
docker-compose exec nginx bash
docker-compose exec postgres psql -U laojun

# 检查网络连接
docker-compose exec admin-api ping postgres
```

## 📞 支持

如有问题，请查看：
1. 日志文件：`logs/` 目录
2. 健康检查状态
3. 容器资源使用情况

---

**注意**: 这是生产环境部署配置，请确保在部署前仔细检查所有安全配置。