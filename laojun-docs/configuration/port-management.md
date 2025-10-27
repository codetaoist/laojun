# 端口管理配置

## 概述

为了避免各个微服务之间的端口冲突，我们实现了统一的端口管理系统。所有服务的端口都在一个地方进行配置和管理。

## 配置文件位置

统一端口配置文件位于：
```
laojun-shared/config/ports.yaml
```

## 配置结构

```yaml
services:
  # 核心业务服务端口 (8080-8099)
  gateway: 8081          # API网关服务
  admin_api: 8082        # 后台管理API服务
  monitoring: 8083       # 监控服务
  discovery: 8084        # 服务注册与发现
  plugin_manager: 8085   # 插件管理服务
  marketplace_api: 8086  # 应用市场API服务
  config_center: 8087    # 配置管理中心

  # 前端服务端口 (3000-3099)
  admin_web: 3000        # 管理后台前端
  marketplace_web: 3001  # 应用市场前端

infra:
  # 基础设施端口
  postgresql: 5432       # 主数据库
  redis: 6379           # 缓存数据库
  prometheus: 9090      # 监控指标收集
  grafana: 3000         # 监控面板
```

## 在代码中使用

### 1. 导入包

```go
import (
    sharedconfig "github.com/taishanglaojun/laojun/laojun-shared/config"
)
```

### 2. 初始化配置

在 `main.go` 中初始化：

```go
func main() {
    // 初始化统一端口配置
    portsConfigPath := filepath.Join("..", "..", "laojun-shared", "config", "ports.yaml")
    if err := sharedconfig.InitGlobalPortConfig(portsConfigPath); err != nil {
        fmt.Printf("Failed to load ports config: %v, using defaults\n", err)
    }

    // 获取当前服务端口
    port := sharedconfig.GetServicePort("admin-api")
    addr := sharedconfig.GetServerAddress("admin-api")
    
    // 启动服务器
    server := &http.Server{
        Addr:    addr,
        Handler: router,
    }
    
    log.Printf("Starting server on %s", addr)
    server.ListenAndServe()
}
```

### 3. 获取端口信息

```go
// 获取特定服务端口
port := sharedconfig.GetServicePort("admin-api")

// 获取服务器地址
addr := sharedconfig.GetServerAddress("admin-api")           // 返回 "0.0.0.0:8082"
addr := sharedconfig.GetServerAddress("admin-api", "localhost") // 返回 "localhost:8082"

// 获取所有服务端口
allPorts := sharedconfig.GetGlobalPortConfig().GetAllServicePorts()
```

## 环境变量覆盖

每个服务都支持通过环境变量覆盖默认端口：

### 服务端口环境变量

| 服务 | 环境变量 | 默认端口 |
|------|----------|----------|
| Gateway | `GATEWAY_PORT` | 8081 |
| Admin API | `ADMIN_API_PORT` | 8082 |
| Monitoring | `MONITORING_PORT` | 8083 |
| Discovery | `DISCOVERY_PORT` | 8084 |
| Plugin Manager | `PLUGIN_MANAGER_PORT` | 8085 |
| Marketplace API | `MARKETPLACE_API_PORT` | 8086 |
| Config Center | `CONFIG_CENTER_PORT` | 8087 |
| Admin Web | `ADMIN_WEB_PORT` | 3000 |
| Marketplace Web | `MARKETPLACE_WEB_PORT` | 3001 |

### 基础设施端口环境变量

| 服务 | 环境变量 | 默认端口 |
|------|----------|----------|
| PostgreSQL | `POSTGRES_PORT` | 5432 |
| Redis | `REDIS_PORT` | 6379 |
| Prometheus | `PROMETHEUS_PORT` | 9090 |
| Grafana | `GRAFANA_PORT` | 3000 |

### 使用示例

```bash
# 设置环境变量
export ADMIN_API_PORT=9082
export GATEWAY_PORT=9081

# 或在 Docker Compose 中
services:
  admin-api:
    environment:
      - ADMIN_API_PORT=9082
    ports:
      - "9082:9082"
```

## Docker 集成

### docker-compose.yml 示例

```yaml
version: '3.8'

services:
  gateway:
    build: ./laojun-gateway
    environment:
      - GATEWAY_PORT=8081
    ports:
      - "8081:8081"
    
  admin-api:
    build: ./laojun-admin-api
    environment:
      - ADMIN_API_PORT=8082
    ports:
      - "8082:8082"
    
  monitoring:
    build: ./laojun-monitoring
    environment:
      - MONITORING_PORT=8083
    ports:
      - "8083:8083"
```

## 端口分配原则

1. **8080-8099**: 核心业务服务
2. **3000-3099**: 前端服务
3. **5000-5999**: 数据库服务
4. **6000-6999**: 缓存服务
5. **9000-9999**: 监控相关服务

## 开发环境启动顺序

建议按以下顺序启动服务：

1. **基础设施**: PostgreSQL (5432), Redis (6379)
2. **配置中心**: Config Center (8087)
3. **服务发现**: Discovery (8084)
4. **核心服务**: Admin API (8082), Marketplace API (8086)
5. **插件管理**: Plugin Manager (8085)
6. **监控服务**: Monitoring (8083)
7. **网关服务**: Gateway (8081)
8. **前端服务**: Admin Web (3000), Marketplace Web (3001)

## 端口检查工具

### Windows PowerShell

```powershell
# 检查特定端口
netstat -ano | findstr :8082

# 检查所有监听端口
netstat -ano | findstr LISTENING

# 检查端口范围
netstat -ano | findstr ":808[0-9]"
```

### 批量检查脚本

```powershell
# check-ports.ps1
$ports = @(8081, 8082, 8083, 8084, 8085, 8086, 8087, 3000, 3001)

foreach ($port in $ports) {
    $result = netstat -ano | findstr ":$port"
    if ($result) {
        Write-Host "Port $port is in use:" -ForegroundColor Red
        Write-Host $result
    } else {
        Write-Host "Port $port is available" -ForegroundColor Green
    }
}
```

## 配置验证

系统提供了配置验证功能：

```go
config := sharedconfig.GetGlobalPortConfig()
if err := config.ValidatePortConfig(); err != nil {
    log.Fatalf("Port configuration validation failed: %v", err)
}
```

验证会检查：
- 端口冲突
- 端口范围合法性
- 配置完整性

## 故障排除

### 常见问题

1. **端口被占用**
   ```
   Error: listen tcp :8082: bind: address already in use
   ```
   解决方案：检查端口占用，修改配置或停止占用进程

2. **配置文件不存在**
   ```
   Failed to load ports config: open ports.yaml: no such file or directory
   ```
   解决方案：系统会自动使用默认配置，或手动创建配置文件

3. **端口冲突**
   ```
   Port configuration validation failed: 端口冲突: gateway 和 admin-api 都使用端口 8081
   ```
   解决方案：修改配置文件中的端口分配

### 调试技巧

1. **查看当前配置**
   ```go
   allPorts := sharedconfig.GetGlobalPortConfig().GetAllServicePorts()
   for service, port := range allPorts {
       fmt.Printf("%s: %d\n", service, port)
   }
   ```

2. **测试端口可用性**
   ```go
   config := sharedconfig.GetGlobalPortConfig()
   if !config.IsPortAvailable(8082) {
       log.Warn("Port 8082 may not be available")
   }
   ```

## 最佳实践

1. **统一管理**: 所有端口配置都通过统一配置文件管理
2. **环境隔离**: 不同环境使用不同的端口范围
3. **文档同步**: 及时更新端口分配文档
4. **自动化检查**: 在 CI/CD 中集成端口冲突检查
5. **监控告警**: 监控端口占用情况，及时发现问题