# 环境变量分层加载指南

本项目通过 `pkg/shared/config/LoadDotenv()` 统一加载 `.env` 文件，支持按环境分层覆盖：

- 基础：`/.env`
- 按环境覆盖：`/.env.<APP_ENV>`（例如：`/.env.development`, `/.env.staging`, `/.env.production`）
- 本地覆盖：`/.env.local`
- 兜底：若上述均不存在，尝试当前工作目录的 `./.env`

加载顺序为逐层覆盖，即后者覆盖前者同名变量。

## 使用方式

- 设置环境：
  - Windows PowerShell：``$env:APP_ENV='production'``
  - Linux/macOS：``export APP_ENV=production``
- 启动服务：
  - `go run ./cmd/admin-api`
  - `cd ./cmd/marketplace-api && go run .`

## 模板文件

仓库根目录提供示例模板：
- `.env.development.example`
- `.env.staging.example`
- `.env.production.example`

请复制为实际文件（去掉 `.example` 后缀）并填充具体值：
- 开发：将 `.env.development.example` 复制为 `.env.development`
- 预发布：将 `.env.staging.example` 复制为 `.env.staging`
- 生产：将 `.env.production.example` 复制为 `.env.production`

> 注意：不要将真实密钥写入示例文件。生产环境建议通过 CI/CD 注入环境变量，不依赖 `.env`。

## 变量说明

- 服务器：`GIN_MODE`, `SERVER_HOST`, `SERVER_PORT`
- 数据库：`DB_DRIVER`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_MAX_*`
- Redis：`REDIS_*`
- 安全与认证：`JWT_*`, `ENABLE_CAPTCHA`, `CORS_ALLOW_ORIGINS`
- 限流：`RATE_LIMIT_*`
- 日志：`LOG_LEVEL`, `LOG_FORMAT`, `LOG_OUTPUT`, `LOG_FILE`
- 其他：`APP_ENV`

## 迁移与清理

- 已统一移除子模块散落的 `.env` 文件（`cmd/admin-api/.env`, `cmd/marketplace-api/.env`）。
- 请将所有环境配置集中维护在仓库根目录。

## 常见问题

- 无 `.env` 文件是否会报错？不会，加载器会使用系统环境变量。
- 为什么我的 `APP_ENV` 为空？请确保在启动命令前设置环境变量，并检查启动日志中是否打印 `(env=...)`。
- 本地覆盖如何生效？将临时配置写入 `/.env.local`，它会在最末加载并覆盖前面配置。