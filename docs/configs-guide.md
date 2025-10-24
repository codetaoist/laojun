# 配置文件说明

## 📁 目录结构

```
config/
├── README.md                           # 本说明文件
└── environments/                       # 环境配置文件
    ├── .env.local.example             # 本地开发环境模板
    ├── .env.docker.example            # Docker环境模板
    ├── .env.development.example       # 开发环境模板
    ├── .env.staging.example           # 预发布环境模板
    └── .env.production.example        # 生产环境模板
```

## 🚀 使用方法

### 1. 本地开发环境

```bash
# 复制本地开发模板
cp config/environments/.env.local.example .env.local

# 根据需要修改配置
# 编辑 .env.local 文件中的数据库连接、Redis配置等
```

### 2. Docker开发环境

```bash
# 复制Docker环境模板
cp config/environments/.env.docker.example .env.docker

# 使用Docker启动
.\start-docker.ps1
```

### 3. 生产环境部署

```bash
# 复制生产环境模板
cp config/environments/.env.production.example .env.production

# 修改生产环境配置
# 务必修改JWT_SECRET、数据库密码等敏感信息
```

## ⚙️ 配置项说明

### 应用配置
- `APP_ENV`: 应用环境 (development/staging/production)
- `APP_DEBUG`: 调试模式开关
- `LOG_LEVEL`: 日志级别 (debug/info/warn/error)

### 数据库配置
- `POSTGRES_HOST`: PostgreSQL主机地址
- `POSTGRES_PORT`: PostgreSQL端口
- `POSTGRES_DB`: 数据库名称
- `POSTGRES_USER`: 数据库用户名
- `POSTGRES_PASSWORD`: 数据库密码

### Redis配置
- `REDIS_HOST`: Redis主机地址
- `REDIS_PORT`: Redis端口
- `REDIS_PASSWORD`: Redis密码
- `REDIS_DB`: Redis数据库编号

### 安全配置
- `JWT_SECRET`: JWT签名密钥 (生产环境必须修改)
- `JWT_EXPIRE_HOURS`: JWT过期时间(小时)

## 🔒 安全注意事项

1. **生产环境配置**：
   - 必须修改默认的JWT_SECRET
   - 使用强密码配置数据库和Redis
   - 禁用调试模式 (APP_DEBUG=false)

2. **配置文件管理**：
   - 实际的 `.env.*` 文件不应提交到版本控制
   - 只提交 `.env.*.example` 模板文件
   - 在 `.gitignore` 中排除实际配置文件

3. **环境隔离**：
   - 不同环境使用不同的数据库和Redis实例
   - 生产环境配置独立管理，避免泄露