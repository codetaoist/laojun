# 太上老君系统多仓库分离架构优化方案

## 📋 项目概述

### 目标
将当前的单体项目重构为多仓库分离的企业级架构，支持开源核心功能和商业版插件市场功能的独立开发、部署和维护。

### 架构原则
- **模块化分离**：核心后台与插件市场完全解耦
- **开源友好**：核心功能采用开源许可证
- **商业化支持**：插件市场采用商业许可证
- **独立部署**：各模块可独立部署和扩展
- **向后兼容**：保持现有API的兼容性

## 🏗️ 多仓库架构设计

### 仓库分离策略

```
太上老君生态系统
├── 📦 laojun-core (开源仓库)
│   ├── 核心后台管理系统
│   ├── 用户权限管理
│   ├── 系统配置管理
│   ├── 插件运行时框架
│   └── 基础API接口
│
├── 💰 laojun-marketplace (商业仓库)
│   ├── 插件市场系统
│   ├── 支付处理系统
│   ├── 用户订单管理
│   ├── 插件分发系统
│   └── 商业分析系统
│
├── 🔧 laojun-shared (共享仓库)
│   ├── 通用工具库
│   ├── 数据库工具
│   ├── 认证组件
│   ├── 日志组件
│   └── 配置管理
│
├── 🔌 laojun-plugins (插件仓库)
│   ├── 插件开发SDK
│   ├── 插件运行时
│   ├── 示例插件
│   └── 插件开发文档
│
└── 🚀 laojun-deploy (部署仓库)
    ├── Docker配置
    ├── Kubernetes配置
    ├── 部署脚本
    └── 环境配置
```

## 📁 详细仓库结构

### 1. laojun-core (核心仓库)

```
laojun-core/
├── README.md
├── LICENSE (MIT)
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── release.yml
│       └── security.yml
│
├── cmd/
│   ├── admin-api/          # 管理后台API服务
│   ├── config-center/      # 配置中心服务
│   ├── plugin-manager/     # 插件管理服务
│   └── tools/              # 管理工具
│
├── internal/
│   ├── handlers/           # HTTP处理器
│   │   ├── admin/          # 管理后台处理器
│   │   ├── auth/           # 认证处理器
│   │   ├── config/         # 配置处理器
│   │   └── plugin/         # 插件处理器
│   ├── services/           # 业务服务层
│   │   ├── admin/          # 管理服务
│   │   ├── auth/           # 认证服务
│   │   ├── config/         # 配置服务
│   │   └── plugin/         # 插件服务
│   ├── models/             # 数据模型
│   │   ├── admin.go        # 管理员模型
│   │   ├── config.go       # 配置模型
│   │   └── plugin.go       # 插件模型
│   ├── middleware/         # 中间件
│   └── database/           # 数据库操作
│
├── pkg/
│   ├── api/                # API定义
│   ├── config/             # 配置管理
│   └── version/            # 版本信息
│
├── web/
│   ├── admin/              # 管理后台前端
│   │   ├── src/
│   │   ├── public/
│   │   ├── package.json
│   │   └── vite.config.ts
│   └── static/             # 静态资源
│
├── configs/
│   ├── app.yaml            # 应用配置
│   ├── database.yaml       # 数据库配置
│   └── environments/       # 环境配置
│
├── db/
│   └── migrations/         # 数据库迁移
│       ├── core/           # 核心表迁移
│       └── plugins/        # 插件表迁移
│
├── docs/
│   ├── api/                # API文档
│   ├── deployment/         # 部署文档
│   └── development/        # 开发文档
│
└── tests/
    ├── unit/               # 单元测试
    ├── integration/        # 集成测试
    └── e2e/                # 端到端测试
```

### 2. laojun-marketplace (商业仓库)

```
laojun-marketplace/
├── README.md
├── LICENSE (Commercial)
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── deploy.yml
│
├── cmd/
│   ├── marketplace-api/    # 市场API服务
│   ├── payment-service/    # 支付服务
│   ├── analytics-service/  # 分析服务
│   └── notification-service/ # 通知服务
│
├── internal/
│   ├── handlers/           # HTTP处理器
│   │   ├── marketplace/    # 市场处理器
│   │   ├── payment/        # 支付处理器
│   │   ├── order/          # 订单处理器
│   │   └── analytics/      # 分析处理器
│   ├── services/           # 业务服务层
│   │   ├── marketplace/    # 市场服务
│   │   ├── payment/        # 支付服务
│   │   ├── order/          # 订单服务
│   │   ├── license/        # 许可证服务
│   │   └── analytics/      # 分析服务
│   ├── models/             # 数据模型
│   │   ├── marketplace.go  # 市场模型
│   │   ├── payment.go      # 支付模型
│   │   ├── order.go        # 订单模型
│   │   └── license.go      # 许可证模型
│   └── integrations/       # 第三方集成
│       ├── stripe/         # Stripe支付
│       ├── alipay/         # 支付宝
│       └── wechat/         # 微信支付
│
├── web/
│   ├── marketplace/        # 市场前端
│   │   ├── src/
│   │   ├── public/
│   │   └── package.json
│   └── admin/              # 市场管理后台
│
├── configs/
│   ├── marketplace.yaml    # 市场配置
│   ├── payment.yaml        # 支付配置
│   └── analytics.yaml      # 分析配置
│
├── db/
│   └── migrations/         # 数据库迁移
│       ├── marketplace/    # 市场表迁移
│       ├── payment/        # 支付表迁移
│       └── analytics/      # 分析表迁移
│
└── docs/
    ├── api/                # API文档
    ├── payment/            # 支付文档
    └── integration/        # 集成文档
```

### 3. laojun-shared (共享仓库)

```
laojun-shared/
├── README.md
├── LICENSE (MIT)
├── go.mod
├── go.sum
├── .gitignore
│
├── pkg/
│   ├── auth/               # 认证组件
│   │   ├── jwt.go          # JWT处理
│   │   ├── middleware.go   # 认证中间件
│   │   └── validator.go    # 验证器
│   ├── database/           # 数据库工具
│   │   ├── connection.go   # 连接管理
│   │   ├── migration.go    # 迁移工具
│   │   └── transaction.go  # 事务管理
│   ├── logger/             # 日志组件
│   │   ├── logger.go       # 日志接口
│   │   ├── zap.go          # Zap实现
│   │   └── middleware.go   # 日志中间件
│   ├── config/             # 配置管理
│   │   ├── loader.go       # 配置加载器
│   │   ├── validator.go    # 配置验证
│   │   └── watcher.go      # 配置监听
│   ├── cache/              # 缓存组件
│   │   ├── redis.go        # Redis缓存
│   │   ├── memory.go       # 内存缓存
│   │   └── interface.go    # 缓存接口
│   ├── middleware/         # 通用中间件
│   │   ├── cors.go         # CORS中间件
│   │   ├── rate_limit.go   # 限流中间件
│   │   └── recovery.go     # 恢复中间件
│   ├── utils/              # 工具函数
│   │   ├── crypto.go       # 加密工具
│   │   ├── validator.go    # 验证工具
│   │   └── converter.go    # 转换工具
│   ├── models/             # 通用模型
│   │   ├── base.go         # 基础模型
│   │   ├── response.go     # 响应模型
│   │   └── pagination.go   # 分页模型
│   └── errors/             # 错误处理
│       ├── codes.go        # 错误码
│       ├── handler.go      # 错误处理器
│       └── middleware.go   # 错误中间件
│
└── web/
    ├── components/         # 共享前端组件
    ├── utils/              # 前端工具
    └── types/              # TypeScript类型定义
```

## 🔄 迁移策略

### 第一阶段：仓库创建和基础结构搭建

1. **创建新仓库**
   - 在GitHub/GitLab上创建四个新仓库
   - 设置适当的访问权限和分支保护规则
   - 配置CI/CD流水线

2. **建立基础结构**
   - 创建各仓库的目录结构
   - 设置Go模块和依赖管理
   - 配置构建脚本和Makefile

### 第二阶段：代码迁移和重构

1. **共享组件迁移**
   - 提取通用工具和组件到laojun-shared
   - 重构认证、日志、配置等组件
   - 建立统一的接口规范

2. **核心功能迁移**
   - 迁移管理后台相关代码到laojun-core
   - 重构数据库模型和迁移脚本
   - 更新API接口和路由

3. **市场功能迁移**
   - 迁移插件市场相关代码到laojun-marketplace
   - 实现支付和许可证系统
   - 建立商业分析功能

### 第三阶段：集成和测试

1. **模块集成**
   - 配置模块间的依赖关系
   - 实现服务间通信机制
   - 建立统一的配置管理

2. **全面测试**
   - 单元测试覆盖
   - 集成测试验证
   - 端到端测试确认

### 第四阶段：部署和发布

1. **部署配置**
   - 配置Docker容器化部署
   - 设置Kubernetes编排
   - 建立监控和日志系统

2. **版本发布**
   - 发布开源核心版本
   - 发布商业市场版本
   - 建立版本管理策略

## 💼 商业化实现

### 版本策略

#### 开源版本 (laojun-core)
- **功能范围**：基础后台管理、用户权限、系统配置、插件运行时
- **许可证**：MIT License
- **目标用户**：个人开发者、小型团队、开源项目
- **支持方式**：社区支持、文档、GitHub Issues

#### 商业版本 (laojun-core + laojun-marketplace)
- **功能范围**：包含开源版本所有功能 + 插件市场 + 支付系统 + 高级分析
- **许可证**：Commercial License
- **目标用户**：企业用户、商业项目、需要插件生态的团队
- **支持方式**：技术支持、定制开发、SLA保证

### 许可证验证机制

```go
// 许可证验证接口
type LicenseValidator interface {
    ValidateCore() error
    ValidateMarketplace() error
    GetFeatures() []string
    IsExpired() bool
}

// 功能开关
type FeatureFlags struct {
    MarketplaceEnabled bool
    PaymentEnabled     bool
    AnalyticsEnabled   bool
    CustomPlugins      bool
}
```

### 部署模式

#### 开源部署
```yaml
# docker-compose.core.yml
version: '3.8'
services:
  laojun-core:
    image: laojun/core:latest
    environment:
      - MARKETPLACE_ENABLED=false
```

#### 商业部署
```yaml
# docker-compose.enterprise.yml
version: '3.8'
services:
  laojun-core:
    image: laojun/core:latest
    environment:
      - MARKETPLACE_ENABLED=true
      - LICENSE_KEY=${LICENSE_KEY}
  
  laojun-marketplace:
    image: laojun/marketplace:latest
    environment:
      - CORE_API_URL=http://laojun-core:8080
```

## 🔧 技术实现细节

### 服务间通信

#### API网关模式
```go
// 统一API网关
type APIGateway struct {
    coreService        CoreService
    marketplaceService MarketplaceService
    authService        AuthService
}

func (g *APIGateway) Route(path string) Handler {
    switch {
    case strings.HasPrefix(path, "/api/admin"):
        return g.coreService.Handler()
    case strings.HasPrefix(path, "/api/marketplace"):
        return g.marketplaceService.Handler()
    default:
        return g.notFoundHandler()
    }
}
```

#### 事件驱动架构
```go
// 事件总线
type EventBus interface {
    Publish(event Event) error
    Subscribe(eventType string, handler EventHandler) error
}

// 插件安装事件
type PluginInstalledEvent struct {
    UserID   string
    PluginID string
    Version  string
    Time     time.Time
}
```

### 数据库分离

#### 核心数据库
```sql
-- 核心系统表
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES roles(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE plugins_registry (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    version VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'installed',
    config JSONB,
    installed_at TIMESTAMP DEFAULT NOW()
);
```

#### 市场数据库
```sql
-- 市场系统表
CREATE TABLE mp_plugins (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) DEFAULT 0,
    category_id UUID REFERENCES mp_categories(id),
    author_id UUID REFERENCES mp_users(id),
    downloads INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE mp_orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    plugin_id UUID REFERENCES mp_plugins(id),
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 📊 监控和运维

### 日志分离
```yaml
# 日志配置
logging:
  core:
    level: info
    output: logs/core/app.log
    format: json
  
  marketplace:
    level: info
    output: logs/marketplace/app.log
    format: json
  
  audit:
    level: info
    output: logs/audit/security.log
    database: true
```

### 监控指标
```go
// 核心系统指标
type CoreMetrics struct {
    ActiveUsers     int64
    PluginsInstalled int64
    APIRequests     int64
    ErrorRate       float64
}

// 市场系统指标
type MarketplaceMetrics struct {
    TotalRevenue    float64
    PluginsSold     int64
    ConversionRate  float64
    TopPlugins      []Plugin
}
```

## 🚀 实施时间表

### 第1-2周：基础设施搭建
- [ ] 创建新仓库和CI/CD配置
- [ ] 建立基础目录结构
- [ ] 配置Go模块和依赖管理

### 第3-4周：共享组件开发
- [ ] 开发laojun-shared组件库
- [ ] 实现认证、日志、配置等基础组件
- [ ] 建立统一的接口规范

### 第5-6周：核心系统迁移
- [ ] 迁移管理后台功能到laojun-core
- [ ] 重构数据库模型和API
- [ ] 实现插件运行时框架

### 第7-8周：市场系统开发
- [ ] 开发laojun-marketplace功能
- [ ] 实现支付和许可证系统
- [ ] 建立商业分析功能

### 第9-10周：集成测试和优化
- [ ] 模块集成和接口对接
- [ ] 全面测试和性能优化
- [ ] 文档编写和部署配置

### 第11-12周：发布和上线
- [ ] 开源版本发布
- [ ] 商业版本发布
- [ ] 社区建设和推广

## 📝 风险评估和应对

### 技术风险
1. **模块依赖复杂性**
   - 风险：模块间依赖关系复杂，可能导致循环依赖
   - 应对：建立清晰的依赖层次，使用接口抽象

2. **数据一致性**
   - 风险：多数据库可能导致数据一致性问题
   - 应对：实现分布式事务或最终一致性机制

3. **性能影响**
   - 风险：服务间通信可能影响性能
   - 应对：优化API设计，使用缓存和异步处理

### 业务风险
1. **用户迁移**
   - 风险：现有用户可能不适应新架构
   - 应对：提供平滑的迁移路径和向后兼容

2. **开发效率**
   - 风险：多仓库可能降低开发效率
   - 应对：建立统一的开发工具和流程

## 📋 成功指标

### 技术指标
- [ ] 代码覆盖率 > 80%
- [ ] API响应时间 < 200ms
- [ ] 系统可用性 > 99.9%
- [ ] 部署时间 < 10分钟

### 业务指标
- [ ] 开源版本下载量 > 1000/月
- [ ] 商业版本转化率 > 5%
- [ ] 插件生态规模 > 50个插件
- [ ] 用户满意度 > 4.5/5

---

**文档版本**: v1.0  
**创建日期**: 2024年12月  
**最后更新**: 2024年12月  
**负责人**: 开发团队