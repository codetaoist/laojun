# 路径配置标准文档

## 概述

本文档定义了太上老君微服务架构中路径配置的标准化规范，确保各服务的目录结构一致性和可维护性，解决项目根目录下自动生成文件夹的问题。

## 问题背景

### 当前问题

项目根目录下自动生成以下文件夹：
- `configs/` - 配置文件目录
- `db/` - 数据库相关文件
- `docs/` - 文档目录  
- `logs/` - 日志文件目录

### 问题原因

1. **硬编码相对路径**: 各服务代码中使用了相对于项目根目录的路径
2. **配置路径不统一**: Viper配置搜索路径包含项目根目录
3. **日志路径默认值**: 日志组件默认输出到项目根目录
4. **Docker挂载路径**: 容器配置使用相对路径挂载

## 路径标准化规范

### 1. 服务目录结构标准

每个服务应遵循以下目录结构：

```
laojun-{service}/
├── cmd/                    # 主程序入口
├── internal/              # 内部代码
├── configs/               # 配置文件目录
│   ├── config.yaml        # 默认配置
│   ├── config.dev.yaml    # 开发环境配置
│   ├── config.prod.yaml   # 生产环境配置
│   └── config.test.yaml   # 测试环境配置
├── logs/                  # 日志文件目录
│   └── {service}.log      # 服务日志文件
├── data/                  # 数据文件目录（可选）
├── migrations/            # 数据库迁移文件（可选）
├── docs/                  # 服务文档（可选）
├── Dockerfile             # Docker构建文件
├── go.mod                 # Go模块文件
└── README.md              # 服务说明文档
```

### 2. 路径引用规范

#### 2.1 配置文件路径

```go
// ✅ 正确：相对于服务目录
viper.AddConfigPath("./configs")
viper.AddConfigPath(".")

// ❌ 错误：相对于项目根目录
viper.AddConfigPath("../configs")
viper.AddConfigPath("../../configs")
```

#### 2.2 日志文件路径

```go
// ✅ 正确：相对于服务目录
defaultLogPath := "./logs/{service_name}.log"

// ❌ 错误：相对于项目根目录
defaultLogPath := "./logs/app.log"
```

#### 2.3 数据文件路径

```go
// ✅ 正确：相对于服务目录
dataPath := "./data"
migrationsPath := "./migrations"

// ❌ 错误：相对于项目根目录
dataPath := "../data"
migrationsPath := "../db/migrations"
```

### 3. 环境变量配置

#### 3.1 路径环境变量

```bash
# 服务特定路径配置
{SERVICE_PREFIX}_CONFIG_PATH=./configs
{SERVICE_PREFIX}_LOG_PATH=./logs
{SERVICE_PREFIX}_DATA_PATH=./data

# 示例
GATEWAY_CONFIG_PATH=./configs
GATEWAY_LOG_PATH=./logs
DISCOVERY_CONFIG_PATH=./configs
DISCOVERY_LOG_PATH=./logs
```

#### 3.2 配置文件环境变量

```bash
# 配置文件指定
{SERVICE_PREFIX}_CONFIG_FILE=./configs/config.yaml
{SERVICE_PREFIX}_CONFIG_ENV=development

# 示例
GATEWAY_CONFIG_FILE=./configs/config.yaml
MONITORING_CONFIG_ENV=production
```

## Docker 配置标准

### 1. 容器内路径标准

```yaml
# 容器内标准路径
environment:
  - CONFIG_PATH=/app/configs
  - LOG_PATH=/app/logs
  - DATA_PATH=/app/data
```

### 2. 卷挂载标准

```yaml
# ✅ 正确：使用绝对路径挂载
volumes:
  - ../../laojun-gateway/configs:/app/configs:ro    # 只读挂载配置
  - ../../laojun-gateway/logs:/app/logs             # 读写挂载日志
  - ../../laojun-gateway/data:/app/data             # 读写挂载数据

# ❌ 错误：使用相对路径挂载
volumes:
  - ./configs:/app/configs
  - ./logs:/app/logs
```

### 3. 构建上下文标准

```yaml
# ✅ 正确：相对于docker-compose.yml的绝对路径
build:
  context: ../../laojun-gateway
  dockerfile: Dockerfile

# ❌ 错误：相对路径可能导致构建失败
build:
  context: ./laojun-gateway
  dockerfile: Dockerfile
```

## 配置加载优先级

### 1. 配置文件搜索顺序

1. 环境变量指定的配置文件路径
2. `./configs/{service}.{env}.yaml`
3. `./configs/{service}.yaml`
4. `./configs/config.{env}.yaml`
5. `./configs/config.yaml`
6. `./config.yaml`

### 2. 配置值优先级

1. 环境变量（最高优先级）
2. 命令行参数
3. 环境特定配置文件
4. 默认配置文件
5. 代码默认值（最低优先级）

## 实施指南

### 1. 现有服务迁移步骤

#### 步骤1：创建服务目录结构

```bash
# 为每个服务创建标准目录
cd laojun-{service}
mkdir -p configs logs data docs
```

#### 步骤2：移动配置文件

```bash
# 将配置文件移动到服务目录
mv ../configs/{service}*.yaml ./configs/
```

#### 步骤3：更新代码中的路径引用

```go
// 更新Viper配置路径
viper.AddConfigPath("./configs")  // 相对于服务目录
viper.AddConfigPath(".")          // 当前目录

// 更新日志路径
defaultLogFile := "./logs/{service}.log"
```

#### 步骤4：更新Docker配置

```yaml
# 更新docker-compose.yml中的路径
volumes:
  - ../../laojun-{service}/configs:/app/configs:ro
  - ../../laojun-{service}/logs:/app/logs
```

### 2. 新服务开发规范

#### 配置加载模板

```go
package config

import (
    "strings"
    "github.com/spf13/viper"
)

func Load(serviceName string) (*Config, error) {
    // 设置环境变量前缀
    viper.SetEnvPrefix(strings.ToUpper(serviceName))
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // 设置配置文件搜索路径
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")  // 服务配置目录
    viper.AddConfigPath(".")          // 服务根目录
    
    // 设置默认值
    viper.SetDefault("log.path", "./logs")
    viper.SetDefault("log.file", fmt.Sprintf("./logs/%s.log", serviceName))
    viper.SetDefault("data.path", "./data")
    
    // 读取配置
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}
```

#### 日志配置模板

```go
package main

import (
    "github.com/codetaoist/laojun-shared/logger"
)

func initLogger(serviceName string) error {
    logConfig := logger.Config{
        Level:  "info",
        Output: "both", // console, file, both
        File: logger.FileConfig{
            Filename: fmt.Sprintf("./logs/%s.log", serviceName),
            MaxSize:  100, // MB
            MaxAge:   30,  // days
            Compress: true,
        },
    }
    
    return logger.Init(logConfig)
}
```

## 验证和测试

### 1. 路径验证脚本

```bash
#!/bin/bash
# validate-paths.sh

echo "验证服务路径配置..."

services=("gateway" "discovery" "monitoring" "admin-api" "marketplace-api" "plugins")

for service in "${services[@]}"; do
    service_dir="laojun-${service}"
    
    # 检查服务目录是否存在
    if [[ ! -d "$service_dir" ]]; then
        echo "❌ 服务目录不存在: $service_dir"
        continue
    fi
    
    # 检查配置目录
    if [[ ! -d "$service_dir/configs" ]]; then
        echo "❌ 配置目录不存在: $service_dir/configs"
    else
        echo "✅ 配置目录存在: $service_dir/configs"
    fi
    
    # 检查日志目录
    if [[ ! -d "$service_dir/logs" ]]; then
        echo "⚠️  日志目录不存在: $service_dir/logs (运行时创建)"
    else
        echo "✅ 日志目录存在: $service_dir/logs"
    fi
done

echo "路径验证完成"
```

### 2. Docker配置验证

```bash
#!/bin/bash
# validate-docker-paths.sh

echo "验证Docker配置路径..."

compose_file="laojun-deploy/platforms/docker/docker-compose.yml"

if [[ ! -f "$compose_file" ]]; then
    echo "❌ Docker Compose文件不存在: $compose_file"
    exit 1
fi

# 检查是否使用了相对路径挂载
if grep -q "\./.*:" "$compose_file"; then
    echo "❌ 发现相对路径挂载，请使用绝对路径"
    grep -n "\./.*:" "$compose_file"
else
    echo "✅ Docker配置路径验证通过"
fi
```

## 最佳实践

### 1. 开发环境

- 使用相对路径，便于本地开发
- 配置文件放在服务目录的 `configs/` 下
- 日志输出到服务目录的 `logs/` 下
- 使用 `.env` 文件管理环境变量

### 2. 测试环境

- 使用环境变量覆盖默认配置
- 配置文件通过CI/CD管道部署
- 日志输出到持久化存储
- 使用测试专用的配置文件

### 3. 生产环境

- 使用容器编排工具管理路径
- 配置文件通过ConfigMap挂载
- 日志输出到集中式日志系统
- 敏感配置使用Secret管理

### 4. 安全考虑

- 配置文件挂载为只读模式
- 敏感信息不写入配置文件
- 日志文件权限控制
- 数据目录访问权限限制

## 迁移检查清单

### 代码层面
- [ ] 更新所有服务的配置文件搜索路径
- [ ] 修改日志文件输出路径
- [ ] 更新数据文件访问路径
- [ ] 移除硬编码的相对路径引用

### 配置层面
- [ ] 创建各服务的configs目录
- [ ] 移动配置文件到对应服务目录
- [ ] 更新环境变量配置
- [ ] 验证配置文件加载正确性

### Docker层面
- [ ] 更新docker-compose.yml中的路径
- [ ] 修改Dockerfile中的路径引用
- [ ] 验证容器构建和运行
- [ ] 测试卷挂载功能

### 文档层面
- [ ] 更新README文档中的路径说明
- [ ] 修改部署文档中的路径配置
- [ ] 更新开发指南中的目录结构
- [ ] 创建路径迁移指南

## 相关文档

- [环境变量配置标准](./env-standards.md)
- [Docker配置指南](../deploy/docker-guide.md)
- [服务开发规范](../docs/development-standards.md)
- [部署最佳实践](../deploy/best-practices.md)