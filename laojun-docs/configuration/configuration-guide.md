# 太上老君系统配置指南

## 概述

本文档提供太上老君系统的完整配置指南，包括环境配置、服务配置、数据库配置等各个方面的详细说明。

## 配置架构

### 配置层次结构
```
配置管理
├── 环境配置 (Environment)
│   ├── 开发环境 (development)
│   ├── 测试环境 (staging)
│   └── 生产环境 (production)
├── 服务配置 (Services)
│   ├── 核心服务配置
│   ├── 业务服务配置
│   └── 支撑服务配置
└── 基础设施配置 (Infrastructure)
    ├── 数据库配置
    ├── 缓存配置
    └── 消息队列配置
```

## 环境变量配置

### 核心环境变量
```bash
# 应用基础配置
APP_NAME=laojun
APP_VERSION=1.0.0
APP_ENV=development
APP_DEBUG=true

# 服务端口配置
ADMIN_API_PORT=8080
MARKETPLACE_API_PORT=8081
CONFIG_CENTER_PORT=8082
GATEWAY_PORT=8083

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=laojun
DB_PASSWORD=laojun123
DB_NAME=laojun
DB_SSL_MODE=disable

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT 配置
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRE_HOURS=24

# 文件存储配置
STORAGE_TYPE=local
STORAGE_PATH=./uploads
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# 监控配置
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
```

### 安全配置
```bash
# 安全相关配置
SECURITY_ENABLE_CAPTCHA=true
SECURITY_CAPTCHA_TTL=5m
SECURITY_MAX_LOGIN_ATTEMPTS=5
SECURITY_LOCKOUT_DURATION=30m

# CORS 配置
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# 限流配置
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100
```

## 服务配置文件

### 管理后台 API 配置 (admin-api.yaml)
```yaml
server:
  port: 8080
  mode: debug
  
database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:5432}
  user: ${DB_USER:laojun}
  password: ${DB_PASSWORD:laojun123}
  dbname: ${DB_NAME:laojun}
  sslmode: ${DB_SSL_MODE:disable}
  
redis:
  host: ${REDIS_HOST:localhost}
  port: ${REDIS_PORT:6379}
  password: ${REDIS_PASSWORD:}
  db: ${REDIS_DB:0}
  
jwt:
  secret: ${JWT_SECRET:your-super-secret-jwt-key}
  expire_hours: ${JWT_EXPIRE_HOURS:24}
  
logging:
  level: info
  format: json
  output: stdout
```

### 插件市场 API 配置 (marketplace-api.yaml)
```yaml
server:
  port: 8081
  mode: debug
  
database:
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:5432}
  user: ${DB_USER:laojun}
  password: ${DB_PASSWORD:laojun123}
  dbname: ${DB_NAME:laojun}
  sslmode: ${DB_SSL_MODE:disable}
  
storage:
  type: ${STORAGE_TYPE:local}
  local_path: ${STORAGE_PATH:./uploads}
  minio:
    endpoint: ${MINIO_ENDPOINT:localhost:9000}
    access_key: ${MINIO_ACCESS_KEY:minioadmin}
    secret_key: ${MINIO_SECRET_KEY:minioadmin}
    bucket: laojun-plugins
    
plugin:
  max_size: 100MB
  allowed_types: [".zip", ".tar.gz"]
  scan_timeout: 30s
```

### 配置中心配置 (config-center.yaml)
```yaml
server:
  port: 8082
  mode: debug
  
etcd:
  endpoints:
    - localhost:2379
  timeout: 5s
  
consul:
  address: localhost:8500
  scheme: http
  
nacos:
  server_configs:
    - ip_addr: localhost
      port: 8848
  client_config:
    namespace_id: public
    timeout_ms: 5000
```

## 数据库配置

### PostgreSQL 配置
```sql
-- 数据库初始化
CREATE DATABASE laojun;
CREATE USER laojun WITH PASSWORD 'laojun123';
GRANT ALL PRIVILEGES ON DATABASE laojun TO laojun;

-- 连接池配置
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
```

### Redis 配置
```conf
# 基础配置
port 6379
bind 127.0.0.1
protected-mode yes
timeout 0
keepalive 300

# 内存配置
maxmemory 256mb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000
```

## Docker 配置

### Docker Compose 配置
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: laojun
      POSTGRES_USER: laojun
      POSTGRES_PASSWORD: laojun123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

## Kubernetes 配置

### ConfigMap 示例
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: laojun-config
data:
  admin-api.yaml: |
    server:
      port: 8080
    database:
      host: postgres-service
      port: 5432
      user: laojun
      password: laojun123
      dbname: laojun
```

### Secret 示例
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: laojun-secrets
type: Opaque
data:
  jwt-secret: eW91ci1zdXBlci1zZWNyZXQtand0LWtleQ==
  db-password: bGFvanVuMTIz
```

## 监控配置

### Prometheus 配置
```yaml
global:
  scrape_interval: 15s
  
scrape_configs:
  - job_name: 'laojun-admin-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    
  - job_name: 'laojun-marketplace-api'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: /metrics
```

### Grafana 配置
```yaml
datasources:
  - name: Prometheus
    type: prometheus
    url: http://localhost:9090
    access: proxy
    isDefault: true
```

## 配置最佳实践

### 1. 环境分离
- 开发、测试、生产环境使用不同的配置文件
- 敏感信息使用环境变量或密钥管理系统
- 配置文件版本控制，但排除敏感信息

### 2. 配置验证
- 启动时验证配置的完整性和正确性
- 提供配置模板和示例
- 实现配置热重载机制

### 3. 安全考虑
- 数据库密码、JWT密钥等敏感信息加密存储
- 定期轮换密钥和密码
- 限制配置文件的访问权限

### 4. 监控告警
- 监控配置变更
- 配置错误时及时告警
- 记录配置变更日志

## 故障排查

### 常见配置问题
1. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数是否正确
   - 确认网络连通性

2. **Redis 连接失败**
   - 检查 Redis 服务状态
   - 验证连接参数
   - 检查防火墙设置

3. **JWT 认证失败**
   - 确认 JWT 密钥配置正确
   - 检查 token 过期时间设置
   - 验证签名算法一致性

### 配置调试
```bash
# 检查配置文件语法
go run cmd/debug-config/main.go -config=configs/admin-api.yaml

# 测试数据库连接
go run cmd/test-db/main.go

# 验证 Redis 连接
go run cmd/test-redis/main.go
```

---

**文档版本**: v1.0  
**最后更新**: 2024年12月  
**维护团队**: 太上老君运维团队