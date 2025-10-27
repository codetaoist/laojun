# 太上老君监控系统 (Laojun Monitoring System)

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/codetaoist/laojun-monitoring)

太上老君监控系统是一个现代化的、高性能的监控解决方案，专为微服务架构设计。它提供全面的系统监控、应用程序监控、日志收集、告警管理和指标分析功能。

## ✨ 核心特性

### 🔍 多维度监控
- **系统指标**: CPU、内存、磁盘、网络、负载、进程等系统级指标
- **应用指标**: HTTP请求、Goroutine、堆内存、GC、运行时等应用级指标
- **自定义指标**: 支持业务指标和自定义收集器
- **实时监控**: 可配置的收集间隔和实时数据更新

### 📊 灵活的存储
- **内存存储**: 高性能内存时序数据库
- **文件存储**: 本地文件系统持久化存储
- **数据库存储**: 支持SQLite、MySQL、PostgreSQL
- **数据压缩**: 自动数据压缩和清理

### 🚨 智能告警
- **规则引擎**: 基于表达式的灵活告警规则
- **多种通知**: Webhook、邮件、Slack等通知方式
- **告警路由**: 基于标签的智能告警路由
- **静默管理**: 告警静默和抑制功能
- **模板引擎**: 自定义告警消息模板

### 📋 日志管理
- **多源收集**: 文件、systemd、应用日志收集
- **实时处理**: 流式日志处理和解析
- **多种输出**: 文件、控制台、Elasticsearch输出
- **处理链**: 过滤、丰富、解析、转换等处理器

### 🌐 RESTful API
- **指标查询**: 强大的指标查询和聚合API
- **告警管理**: 完整的告警规则和实例管理API
- **配置管理**: 动态配置更新和管理API
- **健康检查**: 系统健康状态和就绪检查API

### 📈 可视化支持
- **Prometheus兼容**: 标准的Prometheus指标导出
- **Grafana集成**: 预配置的监控面板
- **实时图表**: 内置的实时指标图表
- **自定义面板**: 支持自定义监控面板

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Laojun Monitoring                       │
├─────────────────────────────────────────────────────────────────┤
│                         API Layer                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Metrics   │ │   Alerts    │ │    Logs     │ │   Config    ││
│  │     API     │ │     API     │ │     API     │ │     API     ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                      Core Services                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Metric    │ │   Alert     │ │  Pipeline   │ │ Collector   ││
│  │  Registry   │ │  Manager    │ │  Manager    │ │  Manager    ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                     Data Processing                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   System    │ │ Application │ │    Log      │ │   Custom    ││
│  │ Collectors  │ │ Collectors  │ │ Processors  │ │ Processors  ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                      Storage Layer                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Memory    │ │    File     │ │  Database   │ │   Remote    ││
│  │   Storage   │ │   Storage   │ │   Storage   │ │   Storage   ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## 📁 项目结构

```
laojun-monitoring/
├── cmd/                           # 应用程序入口
│   └── main.go                   # 主程序
├── config/                       # 配置文件
│   └── config.yaml              # 默认配置
├── internal/                     # 内部包
│   ├── api/                     # API层
│   │   ├── handlers.go          # HTTP处理器
│   │   ├── middleware.go        # 中间件
│   │   └── server.go           # HTTP服务器
│   ├── alerting/               # 告警系统
│   │   ├── manager.go          # 告警管理器
│   │   ├── notifiers.go        # 通知器
│   │   └── rules.go           # 规则引擎
│   ├── collectors/             # 指标收集器
│   │   ├── collector.go        # 收集器接口
│   │   ├── system.go          # 系统指标收集
│   │   └── application.go     # 应用指标收集
│   ├── config/                # 配置管理
│   │   └── config.go         # 配置结构和管理
│   ├── logging/              # 日志系统
│   │   ├── logger.go         # 日志收集器
│   │   ├── outputs.go        # 日志输出
│   │   ├── processors.go     # 日志处理器
│   │   └── pipeline.go       # 日志管道
│   ├── metrics/              # 指标系统
│   │   ├── registry.go       # 指标注册表
│   │   └── types.go         # 指标类型定义
│   └── storage/             # 存储系统
│       ├── manager.go       # 存储管理器
│       ├── memory.go        # 内存存储
│       ├── file.go         # 文件存储
│       └── database.go     # 数据库存储
├── docs/                    # 文档
├── test/                   # 测试文件
├── Dockerfile             # Docker构建文件
## 🚀 快速开始

### 环境要求

- Go 1.21 或更高版本
- 操作系统: Linux, macOS, Windows

### 安装和运行

1. **克隆项目**
   ```bash
   git clone https://github.com/codetaoist/laojun-monitoring.git
   cd laojun-monitoring
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置系统**
   ```bash
   # 复制默认配置
   cp config/config.yaml config/local.yaml
   
   # 编辑配置文件
   vim config/local.yaml
   ```

4. **运行系统**
   ```bash
   # 使用默认配置运行
   go run cmd/main.go
   
   # 使用自定义配置运行
   go run cmd/main.go -config config/local.yaml
   
   # 查看帮助信息
   go run cmd/main.go -help
   ```

5. **验证安装**
   ```bash
   # 检查健康状态
   curl http://localhost:8080/api/v1/health
   
   # 查看指标
   curl http://localhost:8080/metrics
   
   # 查看系统信息
   curl http://localhost:8080/api/v1/status
   ```

### Docker 部署

1. **构建镜像**
   ```bash
   docker build -t laojun-monitoring:latest .
   ```

2. **运行容器**
   ```bash
   docker run -d \
     --name laojun-monitoring \
     -p 8080:8080 \
     -v $(pwd)/config:/app/config \
     -v $(pwd)/data:/app/data \
     laojun-monitoring:latest
   ```

3. **使用 Docker Compose**
   ```bash
   docker-compose up -d
   ```

## ⚙️ 配置说明

### 基础配置

```yaml
# 基础配置
environment: "production"        # 运行环境: development, production
log_level: "info"               # 日志级别: debug, info, warn, error
data_dir: "./data"              # 数据目录

# 服务器配置
server:
  host: "0.0.0.0"              # 监听地址
  port: 8080                   # 监听端口
  read_timeout: "30s"          # 读取超时
  write_timeout: "30s"         # 写入超时
  idle_timeout: "60s"          # 空闲超时
```

### 指标收集配置

```yaml
# 收集器配置
collectors:
  # 系统收集器
  system:
    enabled: true              # 启用系统指标收集
    interval: "30s"           # 收集间隔
    cpu: true                 # CPU指标
    memory: true              # 内存指标
    disk: true                # 磁盘指标
    network: true             # 网络指标
    load: true                # 负载指标
    processes: true           # 进程指标
  
  # 应用收集器
  application:
    enabled: true             # 启用应用指标收集
    interval: "15s"          # 收集间隔
    http: true               # HTTP指标
    runtime: true            # 运行时指标
    custom: true             # 自定义指标
```

### 存储配置

```yaml
# 存储配置
storage:
  type: "memory"              # 存储类型: memory, file, database
  
  # 内存存储
  memory:
    max_size: "1GB"          # 最大内存使用
    retention_period: "24h"   # 数据保留期
  
  # 文件存储
  file:
    directory: "./data/storage"  # 存储目录
    max_file_size: "100MB"      # 最大文件大小
    retention_period: "168h"    # 数据保留期
  
  # 数据库存储
  database:
    type: "sqlite"             # 数据库类型
    sqlite:
      path: "./data/monitoring.db"  # SQLite文件路径
```

### 告警配置

```yaml
# 告警配置
alerting:
  enabled: true               # 启用告警
  evaluation_interval: "30s"  # 评估间隔
  
  # 告警规则示例
  rules:
    - name: "high_cpu_usage"
      expr: "cpu_usage_percent > 80"
      for: "5m"
      labels:
        severity: "warning"
      annotations:
        summary: "CPU使用率过高"
        description: "CPU使用率超过80%已持续5分钟"
  
  # 接收器配置
  receivers:
    - name: "webhook"
      webhook:
        url: "http://localhost:9093/webhook"
        method: "POST"
        headers:
          Content-Type: "application/json"
```

## 📊 API 文档

### 健康检查

```bash
# 健康检查
GET /api/v1/health

# 就绪检查
GET /api/v1/ready

# 系统状态
GET /api/v1/status
```

### 指标管理

```bash
# 获取所有指标
GET /api/v1/metrics

# 查询指标
GET /api/v1/query?query=cpu_usage_percent

# 范围查询
GET /api/v1/query_range?query=cpu_usage_percent&start=1609459200&end=1609545600&step=60

# 获取指标标签
GET /api/v1/labels

# 获取标签值
GET /api/v1/label/{label}/values
```

### 告警管理

```bash
# 获取所有告警
GET /api/v1/alerts

# 获取告警规则
GET /api/v1/rules

# 创建告警规则
POST /api/v1/rules

# 更新告警规则
PUT /api/v1/rules/{id}

# 删除告警规则
DELETE /api/v1/rules/{id}
```

### 收集器管理

```bash
# 获取所有收集器
GET /api/v1/collectors

# 获取收集器状态
GET /api/v1/collectors/{name}

# 启动收集器
POST /api/v1/collectors/{name}/start

# 停止收集器
POST /api/v1/collectors/{name}/stop

# 获取收集器统计
GET /api/v1/collectors/{name}/stats
```

## 🔧 开发指南

### 添加自定义收集器

1. **实现收集器接口**
   ```go
   type CustomCollector struct {
       name     string
       interval time.Duration
       logger   *zap.Logger
   }
   
   func (c *CustomCollector) Start() error {
       // 启动收集逻辑
       return nil
   }
   
   func (c *CustomCollector) Stop() error {
       // 停止收集逻辑
       return nil
   }
   
   func (c *CustomCollector) IsRunning() bool {
       // 返回运行状态
       return true
   }
   ```

2. **注册收集器**
   ```go
   collector := &CustomCollector{
       name:     "custom",
       interval: 30 * time.Second,
       logger:   logger,
   }
   
   collectorManager.RegisterCollector("custom", collector)
   ```

### 添加自定义通知器

1. **实现通知器接口**
   ```go
   type CustomNotifier struct {
       config CustomNotifierConfig
       logger *zap.Logger
   }
   
   func (n *CustomNotifier) Send(alert *Alert) error {
       // 发送通知逻辑
       return nil
   }
   
   func (n *CustomNotifier) Name() string {
       return "custom"
   }
   ```

2. **注册通知器**
   ```go
   notifier := &CustomNotifier{
       config: config,
       logger: logger,
   }
   
   alertManager.AddNotifier(notifier)
   ```

### 添加自定义存储

1. **实现存储接口**
   ```go
   type CustomStorage struct {
       config CustomStorageConfig
       logger *zap.Logger
   }
   
   func (s *CustomStorage) Store(metrics []Metric) error {
       // 存储指标逻辑
       return nil
   }
   
   func (s *CustomStorage) Query(query Query) ([]Metric, error) {
       // 查询指标逻辑
       return nil, nil
   }
   ```

2. **注册存储**
   ```go
   storage := &CustomStorage{
       config: config,
       logger: logger,
   }
   
   ## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/metrics

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 基准测试

```bash
# 运行基准测试
go test -bench=. ./...

# 运行内存分析
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## 📈 监控和可视化

### Grafana 集成

1. **导入仪表板**
   - 使用提供的 Grafana 仪表板模板
   - 配置数据源指向监控系统

2. **自定义面板**
   ```json
   {
     "targets": [
       {
         "expr": "cpu_usage_percent",
         "legendFormat": "CPU使用率"
       }
     ]
   }
   ```

### Prometheus 集成

1. **配置 Prometheus**
   ```yaml
   scrape_configs:
     - job_name: 'laojun-monitoring'
       static_configs:
         - targets: ['localhost:8080']
   ```

2. **查询示例**
   ```promql
   # CPU使用率
   cpu_usage_percent
   
   # 内存使用率
   memory_usage_percent
   
   # 磁盘使用率
   disk_usage_percent{device="/"}
   
   # 网络流量
   rate(network_bytes_total[5m])
   ```

## 🔍 故障排除

### 常见问题

1. **服务启动失败**
   ```bash
   # 检查配置文件
   go run cmd/main.go -config config/config.yaml -validate
   
   # 检查端口占用
   netstat -tlnp | grep 8080
   
   # 查看详细日志
   go run cmd/main.go -log-level debug
   ```

2. **指标收集异常**
   ```bash
   # 检查收集器状态
   curl http://localhost:8080/api/v1/collectors
   
   # 重启收集器
   curl -X POST http://localhost:8080/api/v1/collectors/system/restart
   ```

3. **告警不工作**
   ```bash
   # 检查告警规则
   curl http://localhost:8080/api/v1/rules
   
   # 测试告警规则
   curl -X POST http://localhost:8080/api/v1/rules/test \
     -d '{"expr": "cpu_usage_percent > 80"}'
   ```

### 日志分析

```bash
# 查看错误日志
grep "ERROR" logs/monitoring.log

# 查看告警日志
grep "ALERT" logs/monitoring.log

# 实时监控日志
tail -f logs/monitoring.log
```

## 🤝 贡献指南

### 开发环境设置

1. **Fork 项目**
2. **创建功能分支**
   ```bash
   git checkout -b feature/new-feature
   ```

3. **提交更改**
   ```bash
   git commit -am 'Add new feature'
   ```

4. **推送分支**
   ```bash
   git push origin feature/new-feature
   ```

5. **创建 Pull Request**

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档
- 编写单元测试

### 提交规范

```
type(scope): description

[optional body]

[optional footer]
```

类型:
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建过程或辅助工具的变动

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Prometheus](https://prometheus.io/) - 指标收集和查询
- [Grafana](https://grafana.com/) - 数据可视化
- [Go](https://golang.org/) - 编程语言
- [Zap](https://github.com/uber-go/zap) - 高性能日志库
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP路由器

## 📞 联系方式

- 项目主页: https://github.com/codetaoist/laojun-monitoring
- 问题反馈: https://github.com/codetaoist/laojun-monitoring/issues
- 邮箱: codetaoist@example.com

---

**太上老君监控系统** - 让监控变得简单而强大 🚀
├── configs/              # 配置文件
│   └── config.yaml      # 主配置文件
├── docs/                 # 文档
├── deployments/          # 部署配置
└── README.md            # 项目文档
```

## 快速开始

### 环境要求

- Go 1.23+
- Prometheus (可选)
- Grafana (可选)
- InfluxDB (可选)

### 安装

1. **克隆项目**
```bash
git clone https://github.com/codetaoist/laojun-monitoring.git
cd laojun-monitoring
```

2. **安装依赖**
```bash
go mod download
```

3. **配置文件**
```bash
cp configs/config.yaml.example configs/config.yaml
# 编辑配置文件
vim configs/config.yaml
```

4. **运行服务**
```bash
go run cmd/main.go
```

### 配置说明

#### 基础配置

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8082
  mode: "debug"

# 指标配置
metrics:
  enabled: true
  path: "/metrics"
  interval: "15s"
```

#### 收集器配置

```yaml
collectors:
  # 系统收集器
  system:
    enabled: true
    interval: "15s"
    metrics:
      - "cpu"
      - "memory"
      - "disk"
      - "network"
  
  # 应用程序收集器
  application:
    enabled: true
    interval: "15s"
    metrics:
      - "goroutines"
      - "heap"
      - "gc"
```

#### 告警配置

```yaml
alerting:
  enabled: true
  evaluation_interval: "30s"
  
  rules:
    - id: "high_cpu_usage"
      name: "High CPU Usage"
      query: "system_cpu_usage_percent > 80"
      duration: "5m"
      severity: "warning"
```

## API 文档

### 健康检查

#### GET /health
获取服务健康状态

**响应示例:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "components": {
    "collectors": "healthy",
    "exporters": "healthy",
    "alert_manager": "healthy"
  }
}
```

#### GET /ready
检查服务就绪状态

### 指标管理

#### GET /api/v1/metrics
获取指标统计信息

**响应示例:**
```json
{
  "status": "success",
  "data": {
    "collectors": {
      "system": {
        "status": "running",
        "collect_count": 1234,
        "last_collect_time": "2024-01-15T10:29:45Z"
      }
    },
    "exporters": {
      "prometheus": {
        "status": "running",
        "request_count": 567
      }
    }
  }
}
```

#### POST /api/v1/metrics/query
执行指标查询

**请求体:**
```json
{
  "query": "system_cpu_usage_percent",
  "time": "2024-01-15T10:30:00Z"
}
```

### 告警管理

#### GET /api/v1/alerts
获取告警列表

**查询参数:**
- `status`: 告警状态 (firing, resolved, silenced)
- `severity`: 严重级别 (critical, warning, info)
- `limit`: 返回数量限制
- `offset`: 偏移量

#### POST /api/v1/alerts
创建告警规则

**请求体:**
```json
{
  "name": "High Memory Usage",
  "query": "system_memory_usage_percent > 85",
  "duration": "5m",
  "severity": "critical",
  "labels": {
    "component": "system"
  }
}
```

#### POST /api/v1/alerts/{id}/silence
静默告警

**请求体:**
```json
{
  "duration": "1h",
  "comment": "Maintenance window"
}
```

## 部署

### Docker 部署

1. **构建镜像**
```bash
docker build -t laojun-monitoring:latest .
```

2. **运行容器**
```bash
docker run -d \
  --name laojun-monitoring \
  -p 8082:8082 \
  -v $(pwd)/configs:/app/configs \
  laojun-monitoring:latest
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: laojun-monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: laojun-monitoring
  template:
    metadata:
      labels:
        app: laojun-monitoring
    spec:
      containers:
      - name: monitoring
        image: laojun-monitoring:latest
        ports:
        - containerPort: 8082
        env:
        - name: CONFIG_PATH
          value: "/app/configs/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /app/configs
      volumes:
      - name: config
        configMap:
          name: monitoring-config
```

### Docker Compose

```yaml
version: '3.8'
services:
  monitoring:
    build: .
    ports:
      - "8082:8082"
    volumes:
      - ./configs:/app/configs
    environment:
      - CONFIG_PATH=/app/configs/config.yaml
    depends_on:
      - prometheus
      - grafana
  
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
  
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

## 监控集成

### Prometheus 配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'laojun-monitoring'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana 面板

预配置的 Grafana 面板包括：

1. **系统监控面板**
   - CPU 使用率
   - 内存使用率
   - 磁盘使用率
   - 网络 I/O

2. **应用监控面板**
   - Goroutine 数量
   - 堆内存使用
   - GC 性能
   - 分配速率

3. **告警面板**
   - 活跃告警
   - 告警历史
   - 告警统计

## 性能优化

### 收集器优化

1. **调整收集间隔**
```yaml
collectors:
  system:
    interval: "30s"  # 降低频率以减少开销
```

2. **选择性收集**
```yaml
collectors:
  system:
    metrics:
      - "cpu"
      - "memory"
      # 只收集必要的指标
```

### 存储优化

1. **Prometheus 保留策略**
```yaml
storage:
  prometheus:
    retention: "7d"  # 根据需要调整保留时间
```

2. **批量写入**
```yaml
exporters:
  prometheus:
    batch_size: 1000
    flush_interval: "10s"
```

## 故障排除

### 常见问题

1. **指标收集失败**
```bash
# 检查收集器状态
curl http://localhost:8082/api/v1/metrics

# 查看日志
docker logs laojun-monitoring
```

2. **告警不触发**
```bash
# 检查告警规则
curl http://localhost:8082/api/v1/alerts

# 验证查询语句
curl -X POST http://localhost:8082/api/v1/metrics/query \
  -d '{"query": "system_cpu_usage_percent"}'
```

3. **Prometheus 连接失败**
```bash
# 测试连接
curl http://localhost:9090/-/healthy

# 检查配置
grep -A 5 prometheus configs/config.yaml
```

### 日志分析

```bash
# 查看错误日志
grep ERROR /var/log/laojun-monitoring.log

# 查看告警日志
grep "Alert" /var/log/laojun-monitoring.log

# 实时监控日志
tail -f /var/log/laojun-monitoring.log
```

## 开发指南

### 添加自定义收集器

1. **创建收集器**
```go
type CustomCollector struct {
    // 实现 Collector 接口
}

func (c *CustomCollector) Start() error {
    // 启动逻辑
}

func (c *CustomCollector) Stop() error {
    // 停止逻辑
}
```

2. **注册收集器**
```go
// 在 monitoring.go 中注册
func (ms *MonitoringService) initCollectors() {
    ms.collectors["custom"] = NewCustomCollector()
}
```

### 添加自定义导出器

1. **实现导出器接口**
```go
type Exporter interface {
    Start() error
    Stop() error
    IsRunning() bool
    Health() map[string]interface{}
}
```

2. **配置导出器**
```yaml
exporters:
  custom:
    enabled: true
    endpoint: "http://custom-endpoint"
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 代码规范
- 添加适当的注释和文档
- 编写单元测试
- 确保所有测试通过

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目主页: https://github.com/codetaoist/laojun-monitoring
- 问题反馈: https://github.com/codetaoist/laojun-monitoring/issues
- 邮箱: codetaoist@example.com

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 基础监控功能
- Prometheus 集成
- 告警管理
- 系统和应用指标收集