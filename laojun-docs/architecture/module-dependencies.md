# 太上老君系统模块依赖关系图

## 1. 整体架构依赖图

```mermaid
graph TB
    %% 前端层
    subgraph "前端层 (Frontend Layer)"
        AdminWeb[laojun-admin-web<br/>后台管理前端]
        MarketWeb[laojun-marketplace-web<br/>插件市场前端]
        FrontendShared[laojun-frontend-shared<br/>前端共享库]
    end

    %% API网关层
    subgraph "网关层 (Gateway Layer)"
        Gateway[laojun-gateway<br/>API网关]
    end

    %% 业务服务层
    subgraph "业务服务层 (Business Service Layer)"
        AdminAPI[laojun-admin-api<br/>后台管理API]
        MarketAPI[laojun-marketplace-api<br/>插件市场API]
        Plugins[laojun-plugins<br/>插件系统]
    end

    %% 基础设施层
    subgraph "基础设施层 (Infrastructure Layer)"
        Discovery[laojun-discovery<br/>服务发现]
        ConfigCenter[laojun-config-center<br/>配置中心]
        Monitoring[laojun-monitoring<br/>监控系统]
        Shared[laojun-shared<br/>共享库]
    end

    %% 部署层
    subgraph "部署层 (Deployment Layer)"
        Deploy[laojun-deploy<br/>部署配置]
        Workspace[laojun-workspace<br/>开发环境]
        Docs[laojun-docs<br/>文档]
    end

    %% 依赖关系
    AdminWeb --> FrontendShared
    MarketWeb --> FrontendShared
    
    AdminWeb --> Gateway
    MarketWeb --> Gateway
    
    Gateway --> AdminAPI
    Gateway --> MarketAPI
    Gateway --> Plugins
    
    AdminAPI --> Discovery
    AdminAPI --> ConfigCenter
    AdminAPI --> Shared
    
    MarketAPI --> Discovery
    MarketAPI --> ConfigCenter
    MarketAPI --> Shared
    
    Plugins --> Discovery
    Plugins --> ConfigCenter
    Plugins --> Shared
    
    Discovery --> Shared
    ConfigCenter --> Shared
    Monitoring --> Shared
    
    Deploy --> AdminAPI
    Deploy --> MarketAPI
    Deploy --> Plugins
    Deploy --> Gateway
    
    Workspace --> Deploy
    Docs --> AdminAPI
    Docs --> MarketAPI
    Docs --> Plugins

    %% 样式定义
    classDef frontend fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef gateway fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef business fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef infrastructure fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef deployment fill:#fce4ec,stroke:#880e4f,stroke-width:2px

    class AdminWeb,MarketWeb,FrontendShared frontend
    class Gateway gateway
    class AdminAPI,MarketAPI,Plugins business
    class Discovery,ConfigCenter,Monitoring,Shared infrastructure
    class Deploy,Workspace,Docs deployment
```

## 2. 详细模块依赖关系

### 2.1 前端模块依赖

```mermaid
graph LR
    subgraph "前端依赖关系"
        AdminWeb[laojun-admin-web]
        MarketWeb[laojun-marketplace-web]
        FrontendShared[laojun-frontend-shared]
        
        AdminWeb --> FrontendShared
        MarketWeb --> FrontendShared
        
        AdminWeb -.->|API调用| AdminAPI[laojun-admin-api]
        MarketWeb -.->|API调用| MarketAPI[laojun-marketplace-api]
        MarketWeb -.->|API调用| Plugins[laojun-plugins]
    end
```

**依赖详情:**
- `laojun-admin-web` 依赖 `@laojun/frontend-shared` 共享组件库
- `laojun-marketplace-web` 依赖 `@laojun/frontend-shared` 共享组件库
- 前端通过 API Gateway 调用后端服务

### 2.2 API服务依赖

```mermaid
graph TB
    subgraph "API服务依赖关系"
        Gateway[laojun-gateway<br/>API网关]
        AdminAPI[laojun-admin-api<br/>后台管理API]
        MarketAPI[laojun-marketplace-api<br/>插件市场API]
        Plugins[laojun-plugins<br/>插件系统]
        
        Gateway --> AdminAPI
        Gateway --> MarketAPI
        Gateway --> Plugins
        
        AdminAPI -.->|数据同步| MarketAPI
        MarketAPI -.->|插件管理| Plugins
        AdminAPI -.->|审核管理| Plugins
    end
```

**依赖详情:**
- API Gateway 作为统一入口，路由到各个业务服务
- 后台管理API与插件市场API之间有数据同步关系
- 插件系统为其他服务提供插件管理能力

### 2.3 基础设施依赖

```mermaid
graph TB
    subgraph "基础设施依赖关系"
        Discovery[laojun-discovery<br/>服务发现]
        ConfigCenter[laojun-config-center<br/>配置中心]
        Monitoring[laojun-monitoring<br/>监控系统]
        Shared[laojun-shared<br/>共享库]
        
        Discovery --> Shared
        ConfigCenter --> Shared
        Monitoring --> Shared
        
        AdminAPI --> Discovery
        AdminAPI --> ConfigCenter
        MarketAPI --> Discovery
        MarketAPI --> ConfigCenter
        Plugins --> Discovery
        Plugins --> ConfigCenter
        Gateway --> Discovery
        Gateway --> ConfigCenter
        
        Monitoring -.->|监控| AdminAPI
        Monitoring -.->|监控| MarketAPI
        Monitoring -.->|监控| Plugins
        Monitoring -.->|监控| Gateway
    end
```

## 3. 数据流向图

### 3.1 用户请求流向

```mermaid
sequenceDiagram
    participant User as 用户
    participant Web as 前端应用
    participant Gateway as API网关
    participant Service as 业务服务
    participant DB as 数据库
    participant Cache as 缓存

    User->>Web: 用户操作
    Web->>Gateway: HTTP请求
    Gateway->>Gateway: 认证授权
    Gateway->>Gateway: 路由解析
    Gateway->>Service: 转发请求
    Service->>Cache: 查询缓存
    alt 缓存命中
        Cache-->>Service: 返回数据
    else 缓存未命中
        Service->>DB: 查询数据库
        DB-->>Service: 返回数据
        Service->>Cache: 更新缓存
    end
    Service-->>Gateway: 返回响应
    Gateway-->>Web: 返回响应
    Web-->>User: 显示结果
```

### 3.2 插件生命周期数据流

```mermaid
sequenceDiagram
    participant Dev as 开发者
    participant Market as 插件市场
    participant Admin as 后台管理
    participant Plugin as 插件系统
    participant User as 用户

    Dev->>Market: 提交插件
    Market->>Admin: 同步插件数据
    Admin->>Admin: 创建审核任务
    Admin->>Market: 审核结果通知
    Market->>Plugin: 发布插件
    Plugin->>Plugin: 注册插件
    User->>Market: 浏览插件
    User->>Plugin: 安装插件
    Plugin->>Market: 更新统计数据
    Market->>Admin: 同步统计数据
```

### 3.3 配置管理数据流

```mermaid
graph LR
    subgraph "配置管理数据流"
        ConfigCenter[配置中心]
        AdminAPI[后台管理API]
        MarketAPI[插件市场API]
        Plugins[插件系统]
        Gateway[API网关]
        
        ConfigCenter -->|推送配置| AdminAPI
        ConfigCenter -->|推送配置| MarketAPI
        ConfigCenter -->|推送配置| Plugins
        ConfigCenter -->|推送配置| Gateway
        
        AdminAPI -.->|配置变更| ConfigCenter
        MarketAPI -.->|配置变更| ConfigCenter
    end
```

## 4. 接口依赖关系

### 4.1 内部API接口依赖

```yaml
# API接口依赖关系
api_dependencies:
  laojun-admin-api:
    provides:
      - /api/v1/admin/users
      - /api/v1/admin/plugins
      - /api/v1/admin/reviews
      - /api/v1/admin/stats
    consumes:
      - laojun-marketplace-api: /api/v1/marketplace/plugins
      - laojun-plugins: /api/v1/plugins/management
      - laojun-discovery: /api/v1/services
      - laojun-config-center: /api/v1/configs

  laojun-marketplace-api:
    provides:
      - /api/v1/marketplace/plugins
      - /api/v1/marketplace/categories
      - /api/v1/marketplace/reviews
      - /api/v1/marketplace/purchases
    consumes:
      - laojun-plugins: /api/v1/plugins/registry
      - laojun-admin-api: /api/v1/admin/sync
      - laojun-discovery: /api/v1/services
      - laojun-config-center: /api/v1/configs

  laojun-plugins:
    provides:
      - /api/v1/plugins/registry
      - /api/v1/plugins/management
      - /api/v1/plugins/runtime
      - /api/v1/plugins/lifecycle
    consumes:
      - laojun-discovery: /api/v1/services
      - laojun-config-center: /api/v1/configs

  laojun-gateway:
    provides:
      - /api/v1/gateway/routes
      - /api/v1/gateway/health
    consumes:
      - laojun-admin-api: /api/v1/admin/*
      - laojun-marketplace-api: /api/v1/marketplace/*
      - laojun-plugins: /api/v1/plugins/*
      - laojun-discovery: /api/v1/services
```

### 4.2 外部接口依赖

```mermaid
graph TB
    subgraph "外部接口依赖"
        System[太上老君系统]
        
        System -->|数据库| PostgreSQL[(PostgreSQL)]
        System -->|缓存| Redis[(Redis)]
        System -->|消息队列| RabbitMQ[(RabbitMQ)]
        System -->|对象存储| MinIO[(MinIO)]
        System -->|监控| Prometheus[(Prometheus)]
        System -->|日志| ELK[(ELK Stack)]
        System -->|容器| Docker[Docker]
        System -->|编排| Kubernetes[Kubernetes]
    end
```

## 5. 数据库依赖关系

### 5.1 数据库表依赖

```mermaid
erDiagram
    %% 用户相关表
    users ||--o{ user_roles : has
    user_roles }o--|| roles : belongs_to
    roles ||--o{ role_permissions : has
    role_permissions }o--|| permissions : belongs_to

    %% 插件相关表
    users ||--o{ mp_plugins : creates
    mp_categories ||--o{ mp_plugins : contains
    mp_plugins ||--o{ mp_plugin_versions : has
    mp_plugins ||--o{ mp_plugin_reviews : receives
    users ||--o{ mp_plugin_reviews : writes
    mp_plugins ||--o{ mp_plugin_purchases : sold_as
    users ||--o{ mp_plugin_purchases : makes

    %% 审核相关表
    mp_plugins ||--o{ plugin_reviews : reviewed_by
    users ||--o{ plugin_reviews : conducts
    plugin_reviews ||--o{ developer_appeals : appeals_for

    %% 扩展插件表
    mp_plugins ||--|| extended_plugins : extends
    extended_plugins ||--o{ plugin_interfaces : defines
    extended_plugins ||--o{ plugin_dependencies : depends_on

    %% 系统表
    users ||--o{ audit_logs : generates
    audit_logs }o--|| actions : records
```

### 5.2 数据同步依赖

```mermaid
graph LR
    subgraph "数据同步依赖关系"
        MarketDB[(插件市场数据库)]
        AdminDB[(后台管理数据库)]
        PluginDB[(插件系统数据库)]
        SharedDB[(共享数据库)]
        
        MarketDB <-->|双向同步| AdminDB
        MarketDB -->|插件数据| PluginDB
        AdminDB -->|审核数据| PluginDB
        
        SharedDB -->|用户数据| MarketDB
        SharedDB -->|用户数据| AdminDB
        SharedDB -->|配置数据| PluginDB
    end
```

## 6. 服务通信依赖

### 6.1 同步通信依赖

```mermaid
graph TB
    subgraph "同步通信 (HTTP/gRPC)"
        Gateway[API网关]
        AdminAPI[后台管理API]
        MarketAPI[插件市场API]
        Plugins[插件系统]
        Discovery[服务发现]
        ConfigCenter[配置中心]
        
        Gateway -->|HTTP| AdminAPI
        Gateway -->|HTTP| MarketAPI
        Gateway -->|HTTP| Plugins
        
        AdminAPI -->|gRPC| MarketAPI
        MarketAPI -->|gRPC| Plugins
        AdminAPI -->|gRPC| Plugins
        
        AdminAPI -->|HTTP| Discovery
        MarketAPI -->|HTTP| Discovery
        Plugins -->|HTTP| Discovery
        
        AdminAPI -->|HTTP| ConfigCenter
        MarketAPI -->|HTTP| ConfigCenter
        Plugins -->|HTTP| ConfigCenter
    end
```

### 6.2 异步通信依赖

```mermaid
graph TB
    subgraph "异步通信 (消息队列/事件)"
        EventBus[事件总线]
        AdminAPI[后台管理API]
        MarketAPI[插件市场API]
        Plugins[插件系统]
        Monitoring[监控系统]
        
        AdminAPI -->|发布事件| EventBus
        MarketAPI -->|发布事件| EventBus
        Plugins -->|发布事件| EventBus
        
        EventBus -->|订阅事件| AdminAPI
        EventBus -->|订阅事件| MarketAPI
        EventBus -->|订阅事件| Plugins
        EventBus -->|订阅事件| Monitoring
    end
```

## 7. 部署依赖关系

### 7.1 容器依赖

```mermaid
graph TB
    subgraph "容器部署依赖"
        BaseImage[基础镜像<br/>golang:1.21-alpine]
        
        BaseImage --> AdminAPIImage[laojun-admin-api:latest]
        BaseImage --> MarketAPIImage[laojun-marketplace-api:latest]
        BaseImage --> PluginsImage[laojun-plugins:latest]
        BaseImage --> GatewayImage[laojun-gateway:latest]
        BaseImage --> DiscoveryImage[laojun-discovery:latest]
        BaseImage --> ConfigImage[laojun-config-center:latest]
        
        AdminAPIImage --> AdminAPIPod[admin-api-pod]
        MarketAPIImage --> MarketAPIPod[marketplace-api-pod]
        PluginsImage --> PluginsPod[plugins-pod]
        GatewayImage --> GatewayPod[gateway-pod]
        DiscoveryImage --> DiscoveryPod[discovery-pod]
        ConfigImage --> ConfigPod[config-center-pod]
    end
```

### 7.2 Kubernetes资源依赖

```yaml
# Kubernetes资源依赖关系
k8s_dependencies:
  namespaces:
    - laojun-system    # 系统组件
    - laojun-business  # 业务组件
    - laojun-frontend  # 前端组件

  config_maps:
    - laojun-config    # 全局配置
    - database-config  # 数据库配置
    - redis-config     # Redis配置

  secrets:
    - database-secret  # 数据库密码
    - jwt-secret      # JWT密钥
    - tls-secret      # TLS证书

  services:
    dependencies:
      gateway-service:
        - admin-api-service
        - marketplace-api-service
        - plugins-service
      
      admin-api-service:
        - database-service
        - redis-service
        - discovery-service
        - config-center-service
      
      marketplace-api-service:
        - database-service
        - redis-service
        - discovery-service
        - config-center-service
      
      plugins-service:
        - database-service
        - redis-service
        - discovery-service
        - config-center-service
```

## 8. 开发依赖关系

### 8.1 Go模块依赖

```mermaid
graph TB
    subgraph "Go模块依赖关系"
        SharedModule[laojun-shared]
        
        SharedModule --> AdminAPI[laojun-admin-api]
        SharedModule --> MarketAPI[laojun-marketplace-api]
        SharedModule --> Plugins[laojun-plugins]
        SharedModule --> Gateway[laojun-gateway]
        SharedModule --> Discovery[laojun-discovery]
        SharedModule --> ConfigCenter[laojun-config-center]
        
        AdminAPI -.->|可选依赖| MarketAPI
        MarketAPI -.->|可选依赖| Plugins
    end
```

### 8.2 前端依赖关系

```json
{
  "frontend_dependencies": {
    "laojun-admin-web": {
      "dependencies": {
        "@laojun/frontend-shared": "workspace:*",
        "react": "^18.2.0",
        "antd": "^5.0.0",
        "zustand": "^4.4.0",
        "axios": "^1.5.0"
      }
    },
    "laojun-marketplace-web": {
      "dependencies": {
        "@laojun/frontend-shared": "workspace:*",
        "react": "^18.2.0",
        "antd": "^5.0.0",
        "zustand": "^4.4.0",
        "axios": "^1.5.0"
      }
    },
    "laojun-frontend-shared": {
      "peerDependencies": {
        "react": "^18.2.0",
        "antd": "^5.0.0"
      }
    }
  }
}
```

## 9. 运行时依赖关系

### 9.1 服务启动顺序

```mermaid
graph TB
    subgraph "服务启动依赖顺序"
        Infrastructure[基础设施层]
        Discovery[服务发现]
        ConfigCenter[配置中心]
        Database[(数据库)]
        Redis[(Redis)]
        
        BusinessServices[业务服务层]
        AdminAPI[后台管理API]
        MarketAPI[插件市场API]
        Plugins[插件系统]
        
        Gateway[API网关]
        Frontend[前端应用]
        
        Infrastructure --> Database
        Infrastructure --> Redis
        Database --> Discovery
        Redis --> Discovery
        Discovery --> ConfigCenter
        
        ConfigCenter --> BusinessServices
        BusinessServices --> AdminAPI
        BusinessServices --> MarketAPI
        BusinessServices --> Plugins
        
        AdminAPI --> Gateway
        MarketAPI --> Gateway
        Plugins --> Gateway
        
        Gateway --> Frontend
    end
```

### 9.2 健康检查依赖

```yaml
# 健康检查依赖配置
health_check_dependencies:
  laojun-gateway:
    depends_on:
      - laojun-admin-api
      - laojun-marketplace-api
      - laojun-plugins
    
  laojun-admin-api:
    depends_on:
      - postgresql
      - redis
      - laojun-discovery
      - laojun-config-center
    
  laojun-marketplace-api:
    depends_on:
      - postgresql
      - redis
      - laojun-discovery
      - laojun-config-center
    
  laojun-plugins:
    depends_on:
      - postgresql
      - redis
      - laojun-discovery
      - laojun-config-center
```

## 10. 依赖管理策略

### 10.1 版本兼容性矩阵

| 模块 | laojun-shared | PostgreSQL | Redis | Go版本 |
|------|---------------|------------|-------|--------|
| laojun-admin-api | v1.0.x | 14+ | 6+ | 1.21+ |
| laojun-marketplace-api | v1.0.x | 14+ | 6+ | 1.21+ |
| laojun-plugins | v1.0.x | 14+ | 6+ | 1.21+ |
| laojun-gateway | v1.0.x | - | 6+ | 1.21+ |
| laojun-discovery | v1.0.x | - | 6+ | 1.21+ |
| laojun-config-center | v1.0.x | 14+ | 6+ | 1.21+ |

### 10.2 依赖更新策略

```yaml
dependency_update_strategy:
  # 自动更新策略
  auto_update:
    patch_versions: true    # 自动更新补丁版本
    minor_versions: false   # 手动更新次版本
    major_versions: false   # 手动更新主版本
  
  # 测试策略
  testing:
    unit_tests: required
    integration_tests: required
    e2e_tests: required
    performance_tests: optional
  
  # 回滚策略
  rollback:
    automatic: true
    conditions:
      - health_check_failure
      - performance_degradation
      - error_rate_increase
```

---

## 总结

本文档详细描述了太上老君系统的模块依赖关系，包括：

1. **整体架构依赖**: 展示了各层级之间的依赖关系
2. **数据流向**: 描述了数据在系统中的流转路径
3. **接口依赖**: 定义了内部和外部接口的依赖关系
4. **部署依赖**: 说明了容器和Kubernetes资源的依赖
5. **开发依赖**: 展示了代码模块之间的依赖关系
6. **运行时依赖**: 描述了服务启动和运行时的依赖顺序

这些依赖关系图为系统的开发、部署、运维和扩展提供了重要的参考依据。