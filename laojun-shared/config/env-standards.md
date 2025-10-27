# 环境变量配置标准

## 概述

本文档定义了太上老君微服务架构中环境变量的标准化配置规范，确保各服务的配置管理一致性和可维护性。

## 环境变量命名规范

### 1. 基础命名规则

- 使用大写字母和下划线
- 服务前缀 + 配置类别 + 具体配置项
- 格式：`{SERVICE_PREFIX}_{CATEGORY}_{ITEM}`

### 2. 服务前缀定义

| 服务 | 前缀 | 示例 |
|------|------|------|
| Gateway | `GATEWAY` | `GATEWAY_SERVER_PORT` |
| Discovery | `DISCOVERY` | `DISCOVERY_SERVER_PORT` |
| Config Center | `CONFIG_CENTER` | `CONFIG_CENTER_SERVER_PORT` |
| Admin API | `ADMIN_API` | `ADMIN_API_SERVER_PORT` |
| Marketplace API | `MARKETPLACE_API` | `MARKETPLACE_API_SERVER_PORT` |
| Plugins | `PLUGINS` | `PLUGINS_SERVER_PORT` |
| Monitoring | `MONITORING` | `MONITORING_SERVER_PORT` |

### 3. 配置类别定义

| 类别 | 说明 | 示例 |
|------|------|------|
| `SERVER` | 服务器配置 | `_SERVER_PORT`, `_SERVER_HOST` |
| `DB` | 数据库配置 | `_DB_HOST`, `_DB_PORT`, `_DB_NAME` |
| `REDIS` | Redis配置 | `_REDIS_HOST`, `_REDIS_PORT` |
| `LOG` | 日志配置 | `_LOG_LEVEL`, `_LOG_FILE` |
| `CONFIG` | 配置文件路径 | `_CONFIG_PATH`, `_CONFIG_FILE` |
| `AUTH` | 认证配置 | `_AUTH_SECRET`, `_AUTH_EXPIRY` |

## 路径配置标准

### 1. 配置文件路径

```bash
# 配置文件路径 - 相对于服务目录
{SERVICE_PREFIX}_CONFIG_PATH=./configs
{SERVICE_PREFIX}_CONFIG_FILE=./configs/config.yaml

# 示例
GATEWAY_CONFIG_PATH=./configs
GATEWAY_CONFIG_FILE=./configs/config.yaml
DISCOVERY_CONFIG_PATH=./configs
DISCOVERY_CONFIG_FILE=./configs/config.yaml
```

### 2. 日志文件路径

```bash
# 日志文件路径 - 相对于服务目录
{SERVICE_PREFIX}_LOG_PATH=./logs
{SERVICE_PREFIX}_LOG_FILE=./logs/{service_name}.log

# 示例
GATEWAY_LOG_PATH=./logs
GATEWAY_LOG_FILE=./logs/gateway.log
DISCOVERY_LOG_PATH=./logs
DISCOVERY_LOG_FILE=./logs/discovery.log
```

### 3. 数据文件路径

```bash
# 数据文件路径 - 相对于服务目录
{SERVICE_PREFIX}_DATA_PATH=./data
{SERVICE_PREFIX}_DB_MIGRATIONS_PATH=./migrations

# 示例
ADMIN_API_DATA_PATH=./data
ADMIN_API_DB_MIGRATIONS_PATH=./migrations
```

## 环境特定配置

### 1. 环境标识

```bash
# 环境标识
ENVIRONMENT=development|staging|production
APP_ENV=dev|staging|prod

# 服务特定环境
{SERVICE_PREFIX}_ENV=development|staging|production
```

### 2. 环境特定配置文件

```bash
# 配置文件命名规范
./configs/config.yaml              # 默认配置
./configs/config.{env}.yaml        # 环境特定配置
./configs/{service}.yaml           # 服务默认配置
./configs/{service}.{env}.yaml     # 服务环境特定配置

# 环境变量指定
{SERVICE_PREFIX}_CONFIG_ENV=development
{SERVICE_PREFIX}_CONFIG_FILE=./configs/config.${ENVIRONMENT}.yaml
```

## Docker 环境配置

### 1. 容器内路径标准

```bash
# 容器内标准路径
CONFIG_PATH=/app/configs
LOG_PATH=/app/logs
DATA_PATH=/app/data

# 服务特定容器配置
{SERVICE_PREFIX}_CONFIG_PATH=/app/configs
{SERVICE_PREFIX}_LOG_PATH=/app/logs
{SERVICE_PREFIX}_DATA_PATH=/app/data
```

### 2. 卷挂载配置

```yaml
# docker-compose.yml 示例
services:
  gateway:
    environment:
      - GATEWAY_CONFIG_PATH=/app/configs
      - GATEWAY_LOG_PATH=/app/logs
    volumes:
      - ./laojun-gateway/configs:/app/configs:ro
      - ./laojun-gateway/logs:/app/logs
```

## 配置优先级

### 1. 配置加载顺序

1. 环境变量（最高优先级）
2. 环境特定配置文件
3. 默认配置文件
4. 代码中的默认值（最低优先级）

### 2. 配置文件搜索路径

```go
// Viper 配置搜索路径标准
viper.AddConfigPath("./configs")     // 服务目录下的configs
viper.AddConfigPath(".")             // 服务根目录
```

## 实施指南

### 1. 现有服务迁移

```bash
# 1. 更新环境变量
export GATEWAY_CONFIG_PATH=./configs
export GATEWAY_LOG_PATH=./logs

# 2. 更新配置文件路径
# 确保配置文件在服务目录的 ./configs 下

# 3. 更新日志路径
# 确保日志输出到服务目录的 ./logs 下
```

### 2. 新服务开发

```go
// 配置加载示例
func LoadConfig() (*Config, error) {
    viper.SetEnvPrefix(strings.ToUpper(serviceName))
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // 设置默认值
    viper.SetDefault("config.path", "./configs")
    viper.SetDefault("log.path", "./logs")
    
    // 配置文件搜索路径
    viper.AddConfigPath(viper.GetString("config.path"))
    viper.AddConfigPath(".")
    
    return config, nil
}
```

## 验证和测试

### 1. 配置验证脚本

```bash
#!/bin/bash
# validate-env-config.sh

echo "验证环境变量配置..."

# 检查必需的环境变量
required_vars=(
    "GATEWAY_CONFIG_PATH"
    "GATEWAY_LOG_PATH"
    "DISCOVERY_CONFIG_PATH"
    "DISCOVERY_LOG_PATH"
)

for var in "${required_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
        echo "错误: 环境变量 $var 未设置"
        exit 1
    fi
done

echo "环境变量配置验证通过"
```

### 2. 路径验证

```bash
# 验证配置文件路径
for service in gateway discovery monitoring; do
    config_path="./laojun-${service}/configs"
    if [[ ! -d "$config_path" ]]; then
        echo "警告: 配置目录不存在: $config_path"
    fi
done
```

## 最佳实践

### 1. 开发环境

- 使用 `.env` 文件管理本地开发环境变量
- 配置文件使用相对路径
- 日志输出到服务目录下的 `logs` 文件夹

### 2. 生产环境

- 使用容器编排工具管理环境变量
- 配置文件通过 ConfigMap 或 Secret 挂载
- 日志输出到持久化存储

### 3. 安全考虑

- 敏感配置（密码、密钥）使用环境变量
- 配置文件不包含敏感信息
- 生产环境配置文件不提交到版本控制

## 迁移检查清单

- [ ] 更新所有服务的环境变量命名
- [ ] 修改配置文件搜索路径
- [ ] 更新日志文件输出路径
- [ ] 修改 Docker 配置文件
- [ ] 更新部署脚本
- [ ] 验证各环境配置正确性
- [ ] 更新文档和示例

## 相关文档

- [路径配置标准文档](./path-standards.md)
- [Docker 配置指南](../deploy/docker-guide.md)
- [服务配置示例](./config-examples.md)