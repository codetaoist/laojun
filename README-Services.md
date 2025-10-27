# Laojun 微服务管理指南

## 概述

本项目采用统一的端口配置管理方案，所有服务的端口配置都集中在 `laojun-shared/config/ports.yaml` 文件中。

## 端口配置

### 配置文件位置
- 主配置文件: `laojun-shared/config/ports.yaml`
- 环境变量文件: `.env` (项目根目录)

### 默认端口分配
```yaml
services:
  discovery: "8500"      # 服务发现
  config-center: "8888"  # 配置中心
  gateway: "8080"        # API网关
  admin-api: "8081"      # 管理API
  marketplace-api: "8082" # 市场API
  plugins: "8083"        # 插件管理
  monitoring: "8084"     # 监控服务
```

### 端口配置优先级
1. 统一端口配置 (`ports.yaml`)
2. 服务专属环境变量 (如 `ADMIN_PORT`)
3. 通用环境变量 (`SERVER_PORT`, `PORT`)
4. 服务默认端口

## 服务管理

### 使用启动脚本

#### 构建所有服务
```powershell
.\start-all-services.ps1 -Build
```

#### 启动所有服务
```powershell
.\start-all-services.ps1
```

#### 查看服务状态
```powershell
.\start-all-services.ps1 -Status
```

#### 停止所有服务
```powershell
.\start-all-services.ps1 -Stop
```

### 单独管理服务

#### 构建单个服务
```powershell
cd laojun-admin-api
go build -o admin-api.exe ./cmd/admin-api
```

#### 启动单个服务
```powershell
cd laojun-admin-api
.\admin-api.exe
```

## 服务列表

| 服务名称 | 目录 | 可执行文件 | 默认端口 |
|---------|------|-----------|---------|
| Discovery | laojun-discovery | discovery.exe | 8500 |
| Config Center | laojun-config-center | config-center.exe | 8888 |
| Gateway | laojun-gateway | gateway.exe | 8080 |
| Admin API | laojun-admin-api | admin-api.exe | 8081 |
| Marketplace API | laojun-marketplace-api | marketplace-api.exe | 8082 |
| Plugins | laojun-plugins | plugin-manager.exe | 8083 |
| Monitoring | laojun-monitoring | monitoring.exe | 8084 |

## 环境变量配置

### .env 文件示例
```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=laojun
DB_USER=laojun
DB_PASSWORD=password

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 服务端口覆盖 (可选)
ADMIN_PORT=8081
MARKETPLACE_PORT=8082
GATEWAY_PORT=8080
```

## 开发指南

### 添加新服务
1. 在 `ports.yaml` 中添加服务端口配置
2. 在新服务的 `main.go` 中集成统一端口配置:
   ```go
   import (
       sharedconfig "github.com/codetaoist/laojun-shared/config"
       "github.com/joho/godotenv"
   )
   
   func main() {
       // 加载.env文件
       godotenv.Load()
       
       // 初始化全局端口配置
       sharedconfig.InitializeGlobalConfig("../laojun-shared/config/ports.yaml")
       
       // 获取服务端口
       port := sharedconfig.GetServicePort("your-service-name")
   }
   ```
3. 在 `start-all-services.ps1` 中添加服务配置

### 修改端口配置
1. 编辑 `laojun-shared/config/ports.yaml`
2. 或在 `.env` 文件中设置环境变量覆盖
3. 重启相关服务

## 故障排除

### 端口冲突
- 检查 `ports.yaml` 配置是否有重复端口
- 使用 `netstat -an | findstr :端口号` 检查端口占用
- 修改配置文件或停止占用端口的进程

### 服务启动失败
- 检查可执行文件是否存在
- 查看服务日志输出
- 确认依赖服务是否已启动
- 检查配置文件路径是否正确

### 配置文件加载失败
- 确认 `laojun-shared/config/ports.yaml` 文件存在
- 检查YAML语法是否正确
- 验证文件路径是否正确

## 监控和日志

### 健康检查端点
- Discovery: `http://localhost:8500/health`
- Config Center: `http://localhost:8888/health`
- Gateway: `http://localhost:8080/health`
- Admin API: `http://localhost:8081/health`
- Marketplace API: `http://localhost:8082/health`
- Plugins: `http://localhost:8083/health`
- Monitoring: `http://localhost:8084/health`

### Prometheus指标
- Monitoring服务提供统一的指标收集: `http://localhost:8084/metrics`

### 日志位置
- 各服务日志输出到控制台
- 生产环境建议配置日志文件输出