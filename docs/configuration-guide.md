# 太上老君系统 - 配置指南

## 概述

本文档详细说明了太上老君系统的配置管理，包括本地开发环境和Docker环境的配置差异，以及端口分配策略。

## 环境类型

### 1. 本地开发环境 (Local Development)
- 直接在主机上运行Go服务
- 使用本地PostgreSQL和Redis实例
- 配置文件后缀：`.local.yaml`
- 环境变量文件：`.env.local`

### 2. Docker开发环境 (Docker Development)
- 所有服务运行在Docker容器中
- 使用Docker Compose编排
- 配置文件后缀：`.docker.yaml`
- 环境变量文件：`.env.docker`

## 端口分配策略

### 本地开发环境端口

| 服务 | 端口 | 说明 |
|------|------|------|
| config-center | 8093 | 配置中心服务 |
| admin-api | 8080 | 管理后台API |
| marketplace-api | 8082 | 插件市场API |
| PostgreSQL | 5432 | 数据库服务 |
| Redis | 6379 | 缓存服务 |
| Admin Frontend | 3000 | 管理前端（开发服务器） |
| Marketplace Frontend | 5173 | 市场前端（Vite开发服务器） |

### Docker环境端口映射

| 服务 | 主机端口 | 容器端口 | 说明 |
|------|----------|----------|------|
| nginx | 80 | 80 | Web服务器 |
| nginx | 443 | 443 | HTTPS服务器 |
| postgres | 5432 | 5432 | 数据库服务 |
| redis | 6379 | 6379 | 缓存服务 |
| adminer | 8090 | 8080 | 数据库管理工具 |
| redis-commander | 8091 | 8081 | Redis管理工具 |

## 配置文件结构

```
configs/
├── README.md                    # 配置说明
├── config-center.yaml          # 默认配置中心配置
├── config-center.local.yaml    # 本地开发环境配置
├── config-center.docker.yaml   # Docker环境配置
├── admin-api.yaml              # 默认管理API配置
├── admin-api.local.yaml        # 本地开发环境配置
├── admin-api.docker.yaml       # Docker环境配置
├── marketplace-api.local.yaml  # 本地市场API配置
├── database.yaml               # 默认数据库配置
├── database.local.yaml         # 本地数据库配置
└── database.docker.yaml        # Docker数据库配置
```

## 环境变量文件

```
.env                # 通用环境变量（已弃用，保留兼容性）
.env.local          # 本地开发环境变量
.env.docker         # Docker环境变量
```

## 使用方法

### 本地开发环境启动

1. **准备环境**
   ```bash
   # 复制本地环境变量文件
   cp .env.local .env
   
   # 确保本地PostgreSQL和Redis服务运行
   # PostgreSQL: localhost:5432
   # Redis: localhost:6379
   ```

2. **创建本地数据库**
   ```bash
   # 连接PostgreSQL并创建数据库
   psql -U postgres -c "CREATE DATABASE laojun_local;"
   psql -U postgres -c "CREATE USER laojun WITH PASSWORD 'laojun123';"
   psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE laojun_local TO laojun;"
   ```

3. **启动服务**
   ```bash
   # 启动配置中心
   CONFIG_FILE=configs/config-center.local.yaml ./bin/config-center.exe
   
   # 启动管理API
   CONFIG_FILE=configs/admin-api.local.yaml ./bin/admin-api.exe
   
   # 启动市场API
   CONFIG_FILE=configs/marketplace-api.local.yaml ./bin/marketplace-api.exe
   ```

### Docker环境启动

1. **准备环境**
   ```bash
   # 复制Docker环境变量文件
   cp .env.docker .env
   ```

2. **启动基础设施服务**
   ```bash
   # 启动PostgreSQL、Redis、Nginx等基础服务
   docker-compose -f deploy/docker/docker-compose.minimal.yml up -d
   
   # 启动开发工具（Adminer、Redis Commander）
   docker-compose -f deploy/docker/docker-compose.minimal.yml --profile dev-tools up -d
   ```

3. **访问服务**
   - Web服务: http://localhost
   - 数据库管理: http://localhost:8090
   - Redis管理: http://localhost:8091

## 配置文件详解

### 服务器配置
```yaml
server:
  host: 0.0.0.0        # 监听地址
  port: 8080           # 监听端口
  readTimeout: 30s     # 读取超时
  writeTimeout: 30s    # 写入超时
```

### 数据库配置
```yaml
database:
  postgres:
    host: localhost    # 数据库主机（本地：localhost，Docker：postgres）
    port: 5432        # 数据库端口
    user: laojun      # 数据库用户
    password: "laojun123"  # 数据库密码
    dbname: laojun_local   # 数据库名（本地：laojun_local，Docker：laojun_dev）
    sslmode: disable  # SSL模式
```

### Redis配置
```yaml
redis:
  host: localhost     # Redis主机（本地：localhost，Docker：redis）
  port: 6379         # Redis端口
  password: "redis123"  # Redis密码
  db: 0              # Redis数据库编号
```

## 安全配置

### JWT配置
- 本地环境：`laojun-local-dev-secret-key-2024`
- Docker环境：`laojun-docker-dev-secret-key-2024`
- 生产环境：使用强随机密钥

### CORS配置
- 本地环境：允许localhost的各种端口
- Docker环境：允许容器网络访问
- 生产环境：严格限制允许的域名

## 日志配置

### 本地环境
- 格式：text（便于开发调试）
- 级别：debug
- 输出：控制台+文件

### Docker环境
- 格式：json（便于日志收集）
- 级别：info
- 输出：控制台+文件

## 故障排除

### 端口冲突
1. 检查端口占用：`netstat -ano | findstr :8080`
2. 修改配置文件中的端口号
3. 更新环境变量文件

### 数据库连接失败
1. 确认数据库服务运行状态
2. 检查连接参数（主机、端口、用户名、密码）
3. 验证数据库是否存在

### Redis连接失败
1. 确认Redis服务运行状态
2. 检查Redis密码配置
3. 验证网络连通性

## 最佳实践

1. **环境隔离**：不同环境使用不同的配置文件和数据库
2. **密钥管理**：生产环境使用环境变量管理敏感信息
3. **日志管理**：合理设置日志级别和轮转策略
4. **监控配置**：启用健康检查和性能监控
5. **备份策略**：定期备份配置文件和数据库

## 配置模板

### 新环境配置步骤
1. 复制对应的环境变量模板
2. 修改数据库连接信息
3. 更新服务端口配置
4. 设置安全密钥
5. 配置日志路径
6. 测试服务连通性