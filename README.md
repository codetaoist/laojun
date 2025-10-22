# Laojun 太上老君 - 云原生微服务平台

## 快速启动

### 方案一：Docker Compose（推荐）
- 前提：安装并启动 Docker Desktop。
- 构建并启动：
  - `docker compose -f d:\taishanglaojun\deployments\docker-compose.yml build`
  - `docker compose -f d:\taishanglaojun\deployments\docker-compose.yml up -d`
- 服务与端口：
  - `admin-api`：`http://localhost:8080`
  - `config-center`：`http://localhost:8081`
  - `marketplace-api`：`http://localhost:8082`
  - 前端：`admin-web` `http://localhost:3000`，`marketplace-web` `http://localhost:3001`
- 健康检查：
  - `curl http://localhost:8080/health`
  - `curl http://localhost:8081/health`
  - `curl http://localhost:8082/health`
- 关闭：
  - `docker compose -f d:\taishanglaojun\deployments\docker-compose.yml down`

提示：`docker\docker-compose.yml` 提供更完整的栈（含 MinIO/Prometheus/Grafana），在基本栈稳定后再使用。

### 方案二：本地开发（单服务调试）
- 前提：`Go 1.21+`、`Node.js 18+`、已启动 `postgres`/`redis`（可用上面的 Compose 启动依赖）。
- 常用环境变量（PowerShell 示例）：
  - `setx DB_HOST "localhost"`
  - `setx DB_PORT "5432"`
  - `setx DB_USER "laojun"`
  - `setx DB_PASSWORD "laojun123"`
  - `setx DB_NAME "laojun"`
  - `setx REDIS_HOST "localhost"`
  - `setx REDIS_PORT "6379"`
  - `setx JWT_SECRET "your-super-secret-jwt-key"`
  - `setx GIN_MODE "debug"`
  - `setx SECURITY_ENABLE_CAPTCHA "true"`
  - `setx SECURITY_CAPTCHA_TTL "5m"`
- 启动顺序：
  - 配置中心：`go run ./cmd/config-center`
  - 管理后台 API：`go run ./cmd/admin-api`
  - 插件市场 API：`go run ./cmd/marketplace-api`
- 前端开发：
  - `web/admin`：`npm install && npm run dev`（默认 `REACT_APP_API_URL=http://localhost:8080`）
  - `web/marketplace`：`npm install && npm run dev`（默认 `REACT_APP_API_URL=http://localhost:8082`）
- 接口验证（示例）：
  - 获取验证码：`GET http://localhost:8080/api/v1/auth/captcha`
  - Debug 明文（仅 `GIN_MODE=debug`）：`GET http://localhost:8080/api/v1/auth/captcha/code?key=<captcha_key>`
  - 登录（携带验证码）：
    ```
    curl -X POST "http://localhost:8080/api/v1/login" -H "Content-Type: application/json" -d "{\"username\":\"admin\",\"password\":\"admin123\",\"captcha\":\"1234\",\"captcha_key\":\"<key>\"}"
    ```

## 编译提示与最小修复
- 日志配置字段不匹配：
  - `cmd/admin-api/main.go` 与 `cmd/config-center/main.go` 使用了 `pkg/shared/logger.Config` 的 `Filename/MaxSize/MaxBackups/MaxAge/Compress` 字段，但 `internal/config.LogConfig` 缺少这些字段。
  - 建议统一使用 `pkg/shared/config.LoadConfig()` 并按其 `LogConfig` 字段初始化日志，或扩展 `internal/config.LogConfig` 以匹配。
- `marketplace-api` 处理器与中间件：
  - `handlers.NewAuthHandler` 需传入 `cfg`：调整为 `handlers.NewAuthHandler(authService, jwtManager, cfg)`。
  - 中间件使用请参考 `pkg/shared/middleware` 实际可用的函数或 `MiddlewareChain` 组合。
- 完成上述最小改动后，再执行构建与启动。

## 目录与关键文件
- Compose 文件：`d:\taishanglaojun\deployments\docker-compose.yml`
- Dockerfile：`d:\taishanglaojun\docker\Dockerfile`
- 环境配置示例：`d:\taishanglaojun\configs\admin-api.yaml`、`config-center.yaml`
- 数据库初始化：`d:\taishanglaojun\migrations\`（首次启动 `postgres` 自动执行）