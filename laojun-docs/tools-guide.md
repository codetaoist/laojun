# 工具使用指南

## 端口分配方案

为了避免各个模块之间的端口冲突，我们制定了统一的端口分配方案：

### 核心服务端口分配

| 服务名称 | 端口 | 描述 |
|---------|------|------|
| Gateway (网关) | 8081 | API网关服务，统一入口 |
| Admin API (管理API) | 8082 | 后台管理API服务 |
| Monitoring (监控) | 8083 | 监控服务 |
| Discovery (服务发现) | 8084 | 服务注册与发现 |
| Plugin Manager (插件管理) | 8085 | 插件管理服务 |
| Marketplace API (市场API) | 8086 | 应用市场API服务 |
| Config Center (配置中心) | 8087 | 配置管理中心 |

### 前端服务端口分配

| 服务名称 | 端口 | 描述 |
|---------|------|------|
| Admin Web (管理后台) | 3000 | 管理后台前端 |
| Marketplace Web (市场前端) | 3001 | 应用市场前端 |

### 基础设施端口分配

| 服务名称 | 端口 | 描述 |
|---------|------|------|
| PostgreSQL | 5432 | 主数据库 |
| Redis | 6379 | 缓存数据库 |
| Prometheus | 9090 | 监控指标收集 |
| Grafana | 3000 | 监控面板 |

### 端口分配原则

1. **8080-8099**: 核心业务服务
2. **3000-3099**: 前端服务
3. **5000-5999**: 数据库服务
4. **6000-6999**: 缓存服务
5. **9000-9999**: 监控相关服务

### 环境变量配置

每个服务都支持通过环境变量覆盖默认端口：

```bash
# 网关服务
export SERVER_PORT=8081

# 管理API
export ADMIN_API_PORT=8082

# 监控服务
export MONITORING_PORT=8083

# 服务发现
export DISCOVERY_PORT=8084

# 插件管理
export PLUGIN_MANAGER_PORT=8085

# 市场API
export MARKETPLACE_API_PORT=8086

# 配置中心
export CONFIG_CENTER_PORT=8087
```

### 开发环境启动顺序

建议按以下顺序启动服务：

1. **基础设施**: PostgreSQL, Redis
2. **配置中心**: Config Center (8087)
3. **服务发现**: Discovery (8084)
4. **核心服务**: Admin API (8082), Marketplace API (8086)
5. **插件管理**: Plugin Manager (8085)
6. **监控服务**: Monitoring (8083)
7. **网关服务**: Gateway (8081)
8. **前端服务**: Admin Web (3000), Marketplace Web (3001)

### 端口检查命令

在Windows环境下，可以使用以下命令检查端口占用：

```powershell
# 检查特定端口
netstat -ano | findstr :8081

# 检查所有监听端口
netstat -ano | findstr LISTENING
```

### 注意事项

1. 确保防火墙允许相应端口的访问
2. 在生产环境中，建议使用负载均衡器统一对外提供服务
3. 定期检查端口占用情况，避免冲突
4. 在Docker环境中，注意容器端口映射配置