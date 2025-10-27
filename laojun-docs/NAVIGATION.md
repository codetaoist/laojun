# 太上老君系统文档导航索引

## 📖 文档使用指南

本导航索引帮助您快速找到所需的文档内容。文档按功能模块和用户角色进行分类，支持多种查找方式。

## 🎯 按角色导航

### 👨‍💼 项目经理 / 产品经理
- [系统分析报告](./architecture/system-analysis-report.md) - 了解系统整体架构
- [交付成果总结](./reports/deliverables-summary.md) - 查看项目交付成果
- [开发路线图](./reports/development-roadmap.md) - 了解产品发展规划
- [插件业务流程](./integration/plugin-business-flow.md) - 理解业务流程设计

### 👨‍💻 开发工程师
- [开发概览](./development/overview.md) - 开发环境和规范
- [API 文档](./api/overview.md) - RESTful API 接口文档
- [数据库设计](./database/) - 数据库结构和设计
- [插件开发指南](./marketplace/) - 插件开发和发布流程
- [技术架构设计](./architecture/technical-architecture.md) - 详细技术架构

### 🏗️ 系统架构师
- [技术架构设计](./architecture/technical-architecture.md) - 系统技术架构
- [模块依赖关系](./architecture/module-dependencies.md) - 模块间依赖分析
- [详细分析报告](./architecture/detailed-analysis.md) - 深入系统分析
- [改进建议](./architecture/improvement-recommendations.md) - 架构优化建议
- [统一开发计划](./integration/unified-development-plan.md) - 开发规划

### 🛠️ 运维工程师
- [部署概览](./deployment/overview.md) - 部署架构和方案
- [运维指南](./operations/overview.md) - 系统监控和维护
- [配置指南](./configuration/configuration-guide.md) - 系统配置说明
- [Docker 部署](./deployment/docker.md) - 容器化部署
- [Kubernetes 部署](./deployment/kubernetes.md) - K8s 集群部署

### 🧪 测试工程师
- [测试指南](./development/testing.md) - 测试策略和方法
- [API 测试](./api/testing.md) - API 接口测试
- [集成测试](./development/integration-testing.md) - 集成测试指南
- [性能测试](./operations/performance-testing.md) - 性能测试方案

### 📊 数据分析师
- [插件数据分析](./analytics/plugin-analytics.md) - 插件使用数据分析
- [系统监控指标](./analytics/metrics.md) - 关键性能指标
- [业务数据分析](./analytics/business-analytics.md) - 业务数据洞察

## 🔍 按功能模块导航

### 🏗️ 系统架构
```
architecture/
├── system-analysis-report.md      # 系统整体架构分析
├── technical-architecture.md      # 详细技术架构设计
├── module-dependencies.md         # 各模块间的依赖关系
├── detailed-analysis.md          # 深入的系统分析
└── improvement-recommendations.md # 架构优化建议
```

### 🚀 快速开始
```
getting-started/
├── README.md                      # 快速入门指南
├── environment-setup.md           # 开发环境配置
├── first-steps.md                # 新手入门步骤
└── common-issues.md               # 常见问题解答
```

### 🔧 开发指南
```
development/
├── overview.md                    # 开发环境和规范
├── coding-standards.md            # 代码规范
├── testing.md                     # 测试指南
├── debugging.md                   # 调试技巧
└── best-practices.md              # 最佳实践
```

### 📡 API 文档
```
api/
├── overview.md                    # API 概览
├── authentication.md              # 认证授权
├── endpoints/                     # 接口端点
│   ├── admin-api.md              # 管理后台 API
│   ├── marketplace-api.md        # 插件市场 API
│   └── plugin-api.md             # 插件 API
└── swagger/                       # Swagger 文档
```

### 🗄️ 数据库设计
```
database/
├── schema.md                      # 数据库结构
├── migrations/                    # 数据库迁移
├── indexes.md                     # 索引设计
└── optimization.md                # 性能优化
```

### 🏪 插件市场
```
marketplace/
├── README.md                      # 插件市场概览
├── plugin-development.md         # 插件开发指南
├── plugin-api.md                  # 插件 API 参考
├── examples/                      # 插件开发示例
└── publishing.md                  # 插件发布流程
```

### ⚙️ 配置管理
```
configuration/
├── configuration-guide.md         # 系统配置详细说明
├── configs-guide.md              # 配置文件参考
├── environments/                  # 环境配置
│   ├── development.md            # 开发环境
│   ├── staging.md                # 测试环境
│   └── production.md             # 生产环境
└── secrets-management.md         # 密钥管理
```

### 🚢 部署运维
```
deployment/
├── overview.md                    # 部署概览
├── docker.md                     # Docker 部署
├── kubernetes.md                 # Kubernetes 部署
├── ci-cd/                        # CI/CD 流程
└── environments/                  # 环境配置
```

### 🔧 运维监控
```
operations/
├── overview.md                    # 运维概览
├── monitoring.md                  # 监控系统
├── logging.md                     # 日志管理
├── alerting.md                    # 告警配置
├── backup.md                      # 备份策略
└── troubleshooting.md             # 故障排查
```

### 🔗 集成方案
```
integration/
├── marketplace-integration.md     # 插件市场集成
├── plugin-business-flow.md       # 插件业务流程
├── unified-development-plan.md   # 统一开发计划
└── third-party-integration.md    # 第三方集成
```

### 📊 数据分析
```
analytics/
├── plugin-analytics.md           # 插件数据分析
├── metrics.md                     # 系统监控指标
├── business-analytics.md         # 业务数据分析
└── reporting.md                   # 报表系统
```

### 📋 项目报告
```
reports/
├── deliverables-summary.md       # 交付成果总结
├── development-roadmap.md        # 开发路线图
├── change-log.md                  # 变更日志
└── release-notes.md               # 发布说明
```

## 🔗 快速链接

### 🌟 最常用文档
1. [系统分析报告](./architecture/system-analysis-report.md) - 系统整体架构
2. [开发概览](./development/overview.md) - 开发环境搭建
3. [部署概览](./deployment/overview.md) - 部署指南
4. [API 概览](./api/overview.md) - API 接口文档
5. [配置指南](./configuration/configuration-guide.md) - 系统配置

### 🚀 新手必读
1. [快速入门指南](./getting-started/) - 新用户入门
2. [环境搭建](./getting-started/environment-setup.md) - 开发环境配置
3. [第一步](./getting-started/first-steps.md) - 新手入门步骤
4. [常见问题](./getting-started/common-issues.md) - FAQ

### 🔧 开发必备
1. [开发规范](./development/coding-standards.md) - 代码规范
2. [API 认证](./api/authentication.md) - 认证授权
3. [数据库结构](./database/schema.md) - 数据库设计
4. [插件开发](./marketplace/plugin-development.md) - 插件开发

### 🚢 部署运维
1. [Docker 部署](./deployment/docker.md) - 容器化部署
2. [K8s 部署](./deployment/kubernetes.md) - 集群部署
3. [监控系统](./operations/monitoring.md) - 系统监控
4. [故障排查](./operations/troubleshooting.md) - 问题解决

## 🔍 搜索技巧

### 按关键词搜索
- **架构**: `architecture/`, `technical-architecture.md`, `system-analysis-report.md`
- **API**: `api/`, `endpoints/`, `swagger/`
- **部署**: `deployment/`, `docker.md`, `kubernetes.md`
- **配置**: `configuration/`, `configs-guide.md`
- **插件**: `marketplace/`, `plugin-development.md`
- **监控**: `operations/`, `monitoring.md`
- **数据库**: `database/`, `schema.md`

### 按文件类型搜索
- **概览文档**: `overview.md`, `README.md`
- **指南文档**: `*-guide.md`, `*-guidelines.md`
- **参考文档**: `*-reference.md`, `*-api.md`
- **示例文档**: `examples/`, `*-examples.md`

## 📱 移动端优化

所有文档都针对移动设备进行了优化，支持：
- 📱 响应式布局
- 🔍 移动端搜索
- 📖 离线阅读
- 🔗 快速导航

## 🆘 获取帮助

### 文档问题
- **GitHub Issues**: [提交文档问题](https://github.com/your-org/laojun/issues)
- **文档团队**: docs-team@example.com

### 技术支持
- **技术支持**: tech-support@example.com
- **开发者社区**: [Discord](https://discord.gg/laojun)
- **在线文档**: [https://docs.laojun.com](https://docs.laojun.com)

---

**文档版本**: v1.0  
**最后更新**: 2024年12月19日  
**维护团队**: 太上老君文档团队