# 太上老君微服务平台 🏮

[![Build Status](https://github.com/codetaoist/laojun/workflows/CI/badge.svg)](https://github.com/codetaoist/laojun/actions)
[![Security Scan](https://github.com/codetaoist/laojun/workflows/Security/badge.svg)](https://github.com/codetaoist/laojun/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/codetaoist/laojun)](https://goreportcard.com/report/github.com/codetaoist/laojun)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/github/v/release/codetaoist/laojun)](https://github.com/codetaoist/laojun/releases)

太上老君是一个现代化的微服务平台，提供服务发现、监控、链路追踪等核心功能。基于 Go 语言开发，采用云原生架构设计，支持 Kubernetes 部署。

## ✨ 特性

### 🔍 服务发现
- **多注册中心支持**: Consul、Etcd、Nacos
- **健康检查**: HTTP、TCP、gRPC 健康检查
- **负载均衡**: 轮询、随机、加权轮询、一致性哈希
- **服务路由**: 基于标签的智能路由
- **故障转移**: 自动故障检测和恢复

### 📊 监控系统
- **指标收集**: Prometheus 指标采集
- **可视化**: Grafana 仪表板
- **告警**: 多渠道告警通知（邮件、Slack、钉钉、微信）
- **链路追踪**: Jaeger 分布式追踪
- **日志聚合**: ELK/EFK 日志收集

### 🛡️ 安全保障
- **认证授权**: JWT、OAuth2、RBAC
- **网络安全**: 网络策略、Pod 安全策略
- **镜像安全**: 镜像签名、漏洞扫描
- **运行时安全**: Falco 运行时检测
- **合规检查**: CIS 基准、NIST 标准

### 🚀 DevOps
- **CI/CD**: GitHub Actions 自动化流水线
- **容器化**: Docker 镜像构建和管理
- **编排**: Kubernetes 原生支持
- **GitOps**: ArgoCD 持续部署
- **性能测试**: k6 压力测试

## 🏗️ 架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  Load Balancer  │    │   Web Console   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │              Core Services                    │
         │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
         │  │  Discovery  │  │ Marketplace │  │   Admin     │ │
         │  └─────────────┘  └─────────────┘  └─────────────┘ │
         └───────────────────────┼───────────────────────┘
                                 │
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │ PostgreSQL  │    │    Redis    │    │ Prometheus  │
    └─────────────┘    └─────────────┘    └─────────────┘
```

## 📦 模块组成

### 核心服务
- **[laojun-admin-api](./laojun-admin-api/)** 🛡️ - 管理后台 API 服务
- **[laojun-marketplace-api](./laojun-marketplace-api/)** 🛒 - 插件市场 API 服务
- **[laojun-gateway](./laojun-gateway/)** 🚪 - API 网关服务
- **[laojun-discovery](./laojun-discovery/)** 🔍 - 服务发现中心
- **[laojun-config-center](./laojun-config-center/)** ⚙️ - 配置管理中心

### 前端应用
- **[laojun-admin-web](./laojun-admin-web/)** 💻 - 管理后台前端
- **[laojun-marketplace-web](./laojun-marketplace-web/)** 🌐 - 插件市场前端
- **[laojun-frontend-shared](./laojun-frontend-shared/)** 📚 - 前端共享组件库

### 基础设施
- **[laojun-monitoring](./laojun-monitoring/)** 📊 - 监控系统
- **[laojun-shared](./laojun-shared/)** 🔧 - 共享工具库
- **[laojun-plugins](./laojun-plugins/)** 🔌 - 插件开发框架
- **[laojun-deploy](./laojun-deploy/)** 🚀 - 部署配置
- **[laojun-workspace](./laojun-workspace/)** 🏗️ - 开发工作空间

## 📚 文档导航

### 🎯 快速开始
- **[系统概述](./laojun-docs/README.md)** - 系统整体介绍和文档导航
- **[快速部署](./laojun-docs/deployment/overview.md)** - 5分钟快速部署指南
- **[开发指南](./laojun-docs/development/overview.md)** - 开发环境搭建和规范

### 🏗️ 架构设计
- **[系统架构分析](./laojun-docs/architecture/system-analysis-report.md)** - 完整的系统架构分析
- **[技术架构设计](./laojun-docs/architecture/technical-architecture.md)** - 详细技术架构设计
- **[模块依赖关系](./laojun-docs/architecture/module-dependencies.md)** - 模块间依赖分析
- **[架构改进建议](./laojun-docs/architecture/improvement-recommendations.md)** - 架构优化方案

### 🔗 集成方案
- **[插件市场集成](./laojun-docs/integration/marketplace-integration.md)** - 插件市场与总后台集成
- **[API 集成指南](./laojun-docs/api/README.md)** - API 接口文档和集成指南
- **[部署运维指南](./laojun-docs/deployment/overview.md)** - 完整部署和运维文档

### 📊 项目管理
- **[交付成果总结](./laojun-docs/reports/deliverables-summary.md)** - 项目交付成果汇总
- **[贡献指南](./CONTRIBUTING.md)** - 如何参与项目贡献

## 🚀 快速开始

### 环境要求
- **Go**: 1.21+
- **Node.js**: 18+
- **PostgreSQL**: 13+
- **Redis**: 6+
- **Docker**: 20.10+
- **Kubernetes**: 1.25+ (可选)

### 一键启动 (Docker Compose)
```bash
# 1. 克隆项目
git clone https://github.com/codetaoist/laojun.git
cd laojun

# 2. 启动所有服务
docker-compose up -d

# 3. 访问服务
# - 管理后台: http://localhost:3000
# - 插件市场: http://localhost:3001
# - API 网关: http://localhost:8080
# - 监控面板: http://localhost:9090
```

### 本地开发
```bash
# 1. 安装依赖
make deps

# 2. 启动基础服务 (PostgreSQL, Redis)
make infra-up

# 3. 启动开发环境
make dev

# 4. 运行测试
make test
```

### Kubernetes 部署
```bash
# 1. 配置 Helm Values
cp values.example.yaml values.yaml

# 2. 部署到 K8s
helm install laojun ./charts/laojun -f values.yaml

# 3. 检查状态
kubectl get pods -n laojun
```

## 🧪 测试

```bash
# 单元测试
make test

# 集成测试
make test-integration

# 端到端测试
make test-e2e

# 性能测试
make test-performance

# 测试覆盖率
make test-coverage
```

## 📊 监控和运维

### 监控面板
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **Kibana**: http://localhost:5601

### 健康检查
```bash
# 检查所有服务状态
make health-check

# 查看服务日志
make logs

# 性能分析
make profile
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请查看 [贡献指南](./CONTRIBUTING.md) 了解详细信息。

### 贡献方式
- 🐛 报告 Bug
- 💡 提出功能建议
- 💻 提交代码
- 📚 改进文档
- 🧪 编写测试

### 开发流程
1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request
5. 代码审查
6. 合并代码

## 📄 许可证

本项目采用 [MIT 许可证](./LICENSE)。

## 🔗 相关链接

- **[官方网站](https://laojun.dev)** - 项目官方网站
- **[在线文档](https://docs.laojun.dev)** - 完整在线文档
- **[API 文档](https://api.laojun.dev/docs)** - API 接口文档
- **[社区论坛](https://community.laojun.dev)** - 技术交流社区
- **[问题反馈](https://github.com/codetaoist/laojun/issues)** - Bug 报告和功能请求

## 📞 联系我们

- **邮箱**: contact@laojun.dev
- **微信群**: 扫描二维码加入技术交流群
- **QQ 群**: 123456789
- **Slack**: [加入 Slack 频道](https://laojun.slack.com)

---

⭐ 如果这个项目对你有帮助，请给我们一个 Star！