# Laojun 架构设计文档

## 概述

Laojun 是一个现代化的云原生微服务平台，采用分层架构设计，支持插件化扩展和多环境部署。

## 整体架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Layer                             │
├─────────────────────────────────────────────────────────────┤
│  Web UI  │  Mobile App  │  CLI Tools  │  Third-party Apps  │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Gateway                               │
├─────────────────────────────────────────────────────────────┤
│  Rate Limiting  │  Authentication  │  Load Balancing       │
│  Request Routing │  Response Cache  │  Circuit Breaker     │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                 Microservices Layer                         │
├─────────────────────────────────────────────────────────────┤
│  User Service  │  Auth Service  │  Plugin Service          │
│  Config Service │ Monitor Service │ Notification Service   │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Data Layer                                 │
├─────────────────────────────────────────────────────────────┤
│  PostgreSQL  │  Redis Cache  │  File Storage  │  Message Queue │
└─────────────────────────────────────────────────────────────┘
```

## 分层架构

### 1. 表示层 (Presentation Layer)

**职责**: 处理 HTTP 请求和响应，数据验证，格式转换

**组件**:
- **API Handler**: 处理 RESTful API 请求
- **Middleware**: 认证、授权、日志、限流等横切关注点
- **Validator**: 请求参数验证
- **Serializer**: 数据序列化和反序列化

**目录**: `internal/api/`

### 2. 业务逻辑层 (Business Logic Layer)

**职责**: 实现核心业务逻辑，协调各个组件

**组件**:
- **Service**: 业务服务实现
- **Domain Model**: 领域模型定义
- **Business Rules**: 业务规则引擎
- **Workflow**: 工作流管理

**目录**: `internal/service/`, `internal/model/`

### 3. 数据访问层 (Data Access Layer)

**职责**: 数据持久化，缓存管理，外部服务调用

**组件**:
- **Repository**: 数据访问接口
- **DAO**: 数据访问对象
- **Cache**: 缓存管理
- **External Client**: 外部服务客户端

**目录**: `internal/repository/`, `internal/cache/`

### 4. 基础设施层 (Infrastructure Layer)

**职责**: 提供技术基础设施支持

**组件**:
- **Database**: 数据库连接和管理
- **Message Queue**: 消息队列
- **File Storage**: 文件存储
- **Monitoring**: 监控和指标收集

**目录**: `internal/database/`, `internal/monitoring/`

## 核心组件设计

### 1. 插件系统

```go
type Plugin interface {
    Name() string
    Version() string
    Initialize(config Config) error
    Execute(ctx Context, input interface{}) (interface{}, error)
    Cleanup() error
}

type PluginManager struct {
    plugins map[string]Plugin
    loader  PluginLoader
}
```

**特性**:
- 动态加载和卸载
- 版本管理
- 依赖解析
- 安全沙箱

### 2. 配置管理

```go
type ConfigManager interface {
    Get(key string) interface{}
    Set(key string, value interface{}) error
    Watch(key string, callback func(value interface{}))
    Reload() error
}
```

**特性**:
- 多环境配置
- 热更新
- 配置验证
- 敏感信息加密

### 3. 服务发现

```go
type ServiceDiscovery interface {
    Register(service ServiceInfo) error
    Deregister(serviceID string) error
    Discover(serviceName string) ([]ServiceInfo, error)
    Watch(serviceName string, callback func([]ServiceInfo))
}
```

**特性**:
- 自动注册和注销
- 健康检查
- 负载均衡
- 故障转移

### 4. 监控系统

```go
type Monitor interface {
    Counter(name string, tags map[string]string) Counter
    Gauge(name string, tags map[string]string) Gauge
    Histogram(name string, tags map[string]string) Histogram
    Timer(name string, tags map[string]string) Timer
}
```

**指标类型**:
- **Counter**: 计数器（请求数、错误数）
- **Gauge**: 仪表盘（内存使用、连接数）
- **Histogram**: 直方图（响应时间分布）
- **Timer**: 计时器（操作耗时）

## 数据模型设计

### 1. 用户模型

```go
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"uniqueIndex"`
    Email     string    `json:"email" gorm:"uniqueIndex"`
    Password  string    `json:"-" gorm:"not null"`
    Nickname  string    `json:"nickname"`
    Avatar    string    `json:"avatar"`
    Status    UserStatus `json:"status" gorm:"default:active"`
    Roles     []Role    `json:"roles" gorm:"many2many:user_roles;"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}
```

### 2. 角色权限模型

```go
type Role struct {
    ID          uint         `json:"id" gorm:"primaryKey"`
    Name        string       `json:"name" gorm:"uniqueIndex"`
    DisplayName string       `json:"display_name"`
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"uniqueIndex"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### 3. 插件模型

```go
type Plugin struct {
    ID          uint         `json:"id" gorm:"primaryKey"`
    Name        string       `json:"name" gorm:"uniqueIndex"`
    Version     string       `json:"version"`
    Type        PluginType   `json:"type"`
    Status      PluginStatus `json:"status"`
    Config      JSON         `json:"config" gorm:"type:jsonb"`
    Metadata    JSON         `json:"metadata" gorm:"type:jsonb"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}
```

## 安全架构

### 1. 认证机制

- **JWT Token**: 无状态认证
- **OAuth2**: 第三方登录
- **API Key**: 服务间认证
- **mTLS**: 双向 TLS 认证

### 2. 授权模型

- **RBAC**: 基于角色的访问控制
- **ABAC**: 基于属性的访问控制
- **Resource-based**: 基于资源的权限控制

### 3. 数据安全

- **加密存储**: 敏感数据加密
- **传输加密**: TLS/SSL 加密
- **数据脱敏**: 日志和监控数据脱敏
- **审计日志**: 操作审计和追踪

## 性能优化

### 1. 缓存策略

- **多级缓存**: 内存缓存 + Redis 缓存
- **缓存预热**: 系统启动时预加载热点数据
- **缓存更新**: 写入时更新，定时刷新
- **缓存穿透**: 布隆过滤器防护

### 2. 数据库优化

- **连接池**: 数据库连接池管理
- **读写分离**: 主从数据库分离
- **分库分表**: 水平分片
- **索引优化**: 查询索引优化

### 3. 并发控制

- **协程池**: Goroutine 池管理
- **限流**: 令牌桶算法限流
- **熔断**: 断路器模式
- **超时控制**: 请求超时管理

## 可扩展性设计

### 1. 水平扩展

- **无状态设计**: 服务无状态化
- **负载均衡**: 多实例负载均衡
- **自动扩缩容**: 基于指标的自动扩缩容

### 2. 垂直扩展

- **资源隔离**: CPU、内存资源隔离
- **性能调优**: JVM 参数调优
- **硬件升级**: 硬件资源升级

### 3. 功能扩展

- **插件机制**: 功能插件化
- **API 版本**: API 版本管理
- **向后兼容**: 版本向后兼容

## 部署架构

### 1. 容器化部署

```yaml
# Docker 部署
services:
  laojun-api:
    image: laojun:latest
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
    depends_on:
      - postgres
      - redis
```

### 2. Kubernetes 部署

```yaml
# K8s 部署
apiVersion: apps/v1
kind: Deployment
metadata:
  name: laojun-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: laojun-api
  template:
    spec:
      containers:
      - name: laojun-api
        image: laojun:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 3. 云原生部署

- **Service Mesh**: Istio 服务网格
- **API Gateway**: Kong/Nginx 网关
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack 日志收集

## 技术选型

### 后端技术栈

- **语言**: Go 1.21+
- **框架**: Gin (HTTP), GORM (ORM)
- **数据库**: PostgreSQL, Redis
- **消息队列**: RabbitMQ, Kafka
- **监控**: Prometheus, Grafana
- **日志**: Logrus, ELK Stack

### 基础设施

- **容器**: Docker, Podman
- **编排**: Kubernetes, Docker Compose
- **CI/CD**: GitHub Actions, GitLab CI
- **云平台**: AWS, Azure, GCP
- **CDN**: CloudFlare, AWS CloudFront

### 开发工具

- **IDE**: VS Code, GoLand
- **版本控制**: Git, GitHub/GitLab
- **包管理**: Go Modules
- **测试**: Testify, GoConvey
- **文档**: Swagger, GitBook

---

本架构设计遵循微服务架构最佳实践，支持高并发、高可用、可扩展的云原生应用部署。