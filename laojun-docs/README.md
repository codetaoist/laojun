# 太上老君系统 - 统一文档中心

<div align="center">

![Laojun Logo](https://img.shields.io/badge/Laojun-太上老君-blue?style=for-the-badge)
![Version](https://img.shields.io/badge/Version-2.0-green?style=for-the-badge)
![Docs](https://img.shields.io/badge/Docs-Complete-orange?style=for-the-badge)

**现代化云原生微服务平台 | 插件化架构 | 企业级解决方案**

</div>

---

## 📖 文档概述

欢迎来到太上老君系统的统一文档中心！本文档中心整合了系统的完整文档体系，涵盖架构设计、开发指南、部署运维、插件市场等各个方面，为开发者、架构师、运维人员提供全方位的技术支持。

### 🎯 系统特性

- **🏗️ 微服务架构**: 基于 Go + Gin 的高性能微服务架构
- **🔌 插件化设计**: 完整的插件生态系统，支持动态扩展
- **☁️ 云原生部署**: 支持 Docker、Kubernetes 等现代化部署方案
- **📊 数据驱动**: PostgreSQL + Redis 的高性能数据存储方案
- **🛡️ 企业级安全**: 完善的认证授权和安全防护机制
- **📈 可观测性**: 全链路监控、日志收集、性能分析

---

## 🧭 快速导航

### 📋 文档导航系统
- **[📑 文档导航](./NAVIGATION.md)** - 按角色和功能分类的导航系统
- **[📇 文档索引](./INDEX.md)** - 按字母顺序和主题分类的完整索引
- **[🗺️ 站点地图](./SITEMAP.md)** - 完整的文档结构树状视图

### 🔍 快速查找
- **新手用户**: [快速开始](#-快速开始) → [开发指南](#-开发指南)
- **开发工程师**: [API 文档](#-api-文档) → [开发指南](#-开发指南) → [测试指南](#-测试与质量)
- **系统架构师**: [架构设计](#️-架构设计) → [技术分析](#-技术分析) → [企业方案](#-企业方案)
- **运维工程师**: [部署运维](#-部署运维) → [监控运维](#-监控运维) → [配置管理](#️-配置管理)

---

## 🏗️ 架构设计

### 核心架构文档
- **[系统架构概览](./architecture/overview.md)** - 完整的系统架构设计和技术选型
- **[技术架构图](./architecture/technical-architecture.md)** - 详细的技术架构图和组件说明
- **[模块依赖关系](./architecture/module-dependencies.md)** - 模块间依赖关系分析
- **[架构改进建议](./architecture/improvement-recommendations.md)** - 架构优化和改进方案

### 企业级方案
- **[多仓库分离方案](./MULTI_REPO_ENTERPRISE_PLAN.md)** - 企业级多仓库架构迁移方案
- **[系统分析报告](./architecture/system-analysis-report.md)** - 深度系统架构分析
- **[详细分析报告](./architecture/detailed-analysis-report.md)** - 技术架构详细分析

---

## 🚀 快速开始

### 入门指南
- **[快速入门](./getting-started/README.md)** - 新用户快速上手指南
- **[环境搭建](./getting-started/environment-setup.md)** - 开发环境配置详解
- **[本地开发](./getting-started/local-development.md)** - 本地开发环境搭建
- **[Docker 快速启动](./getting-started/docker-quickstart.md)** - 容器化快速部署

### 项目结构
- **[项目结构说明](./project-structure.md)** - 完整的项目目录结构说明
- **[模块划分](./getting-started/module-overview.md)** - 各模块功能和职责说明

---

## 💻 开发指南

### 开发基础
- **[开发概览](./development/overview.md)** - 开发环境、技术栈、规范总览
- **[编码规范](./development/coding-standards.md)** - Go 和 TypeScript 编码规范
- **[开发工作流](./development/development-workflow.md)** - Git 工作流和协作规范
- **[调试指南](./development/debugging-guide.md)** - 调试技巧和工具使用

### 架构开发
- **[微服务开发](./development/microservice-development.md)** - 微服务开发指南
- **[插件开发](./development/plugin-development.md)** - 插件系统开发指南
- **[中间件开发](./development/middleware-development.md)** - 自定义中间件开发

---

## 📡 API 文档

### API 概览
- **[API 总览](./api/overview.md)** - RESTful API 接口规范和认证
- **[API 设计规范](./api/design-standards.md)** - API 设计标准和最佳实践
- **[Swagger 文档](./api/swagger/)** - 交互式 API 文档

### 核心 API
- **[认证 API](./api/endpoints/auth.md)** - 用户认证和授权接口
- **[用户管理 API](./api/endpoints/users.md)** - 用户管理相关接口
- **[插件市场 API](./api/endpoints/marketplace.md)** - 插件市场相关接口
- **[配置管理 API](./api/endpoints/config.md)** - 系统配置管理接口

---

## 🗄️ 数据库设计

### 数据库文档
- **[数据库概览](./database/README.md)** - 数据库架构和设计原则
- **[表结构设计](./database/schema-design.md)** - 详细的表结构设计
- **[命名约定](./database/conventions.md)** - 数据库命名规范
- **[迁移指南](./database/migrations.md)** - 数据库迁移和版本管理

---

## 🏪 插件市场

### 插件系统
- **[插件市场概览](./marketplace/README.md)** - 插件市场系统总览
- **[业务流程设计](./marketplace/business-flow.md)** - 插件市场业务流程
- **[API 接口设计](./marketplace/api-design.md)** - 插件市场 API 规范
- **[审核工作流程](./marketplace/review-workflow.md)** - 插件审核流程

### 插件开发
- **[插件开发指南](./marketplace/plugin-development-guide.md)** - 插件开发完整指南
- **[插件 API 参考](./marketplace/plugin-api.md)** - 插件开发接口文档
- **[插件示例](./marketplace/examples/)** - 插件开发示例代码

---

## ⚙️ 配置管理

### 配置系统
- **[配置管理指南](./configuration/configuration-guide.md)** - 系统配置完整指南
- **[配置文件参考](./configuration/configs-guide.md)** - 配置文件详细说明
- **[环境配置](./configuration/environments/)** - 多环境配置管理

### 配置中心
- **[配置中心设计](./configuration/config-center.md)** - 配置中心架构设计
- **[动态配置](./configuration/dynamic-config.md)** - 动态配置更新机制

---

## 🚢 部署运维

### 部署方案
- **[部署概览](./deployment/overview.md)** - 部署架构和方案总览
- **[Docker 部署](./deployment/docker.md)** - 容器化部署完整指南
- **[Kubernetes 部署](./deployment/kubernetes.md)** - K8s 集群部署方案
- **[生产环境部署](./deployment/production.md)** - 生产环境部署最佳实践

### 部署优化
- **[部署优化总结](./deployment/DEPLOYMENT_OPTIMIZATION.md)** - 部署性能优化方案
- **[迁移部署总结](./MIGRATION_DEPLOYMENT_SUMMARY.md)** - 部署迁移实施总结

---

## 🛠️ 监控运维

### 运维管理
- **[运维概览](./operations/overview.md)** - 系统监控和运维管理总览
- **[监控系统](./operations/monitoring.md)** - Prometheus + Grafana 监控方案
- **[日志管理](./operations/logging.md)** - 日志收集和分析系统
- **[告警配置](./operations/alerting.md)** - 告警规则和通知配置

### 运维实践
- **[故障排查](./operations/troubleshooting.md)** - 常见问题排查指南
- **[性能优化](./operations/performance-optimization.md)** - 系统性能优化实践
- **[备份恢复](./operations/backup-recovery.md)** - 数据备份和恢复策略

---

## 🔗 集成方案

### 系统集成
- **[插件市场集成](./integration/marketplace-integration.md)** - 插件市场与管理后台集成
- **[管理后台集成](./integration/marketplace-admin-integration.md)** - 管理后台集成方案
- **[插件业务流程](./integration/plugin-business-flow.md)** - 插件业务流程集成

### 第三方集成
- **[外部系统集成](./integration/external-systems.md)** - 第三方系统集成指南
- **[API 网关集成](./integration/api-gateway.md)** - API 网关集成方案

---

## 📊 数据分析

### 分析报告
- **[系统分析报告](./analytics/system-analysis.md)** - 系统性能和使用分析
- **[用户行为分析](./analytics/user-behavior.md)** - 用户行为数据分析
- **[性能分析报告](./analytics/performance-analysis.md)** - 系统性能深度分析

### 数据可视化
- **[数据仪表板](./analytics/dashboards.md)** - 数据可视化仪表板
- **[报表系统](./analytics/reporting.md)** - 自动化报表生成系统

---

## 📋 项目报告

### 交付成果
- **[交付成果总结](./reports/deliverables-summary.md)** - 项目交付成果汇总
- **[文档重构总结](./DOCUMENTATION_SUMMARY.md)** - 文档系统重构总结
- **[文件迁移分析](./FILE_MIGRATION_ANALYSIS.md)** - 文件迁移归属分析

### 项目分析
- **[项目分析报告](./ANALYSIS_REPORT.md)** - 完整的项目分析报告
- **[统一架构规划](./UNIFIED_ARCHITECTURE_PLAN.md)** - 统一架构设计规划

---

## 🧪 测试与质量

### 测试策略
- **[测试指南](./development/testing-guide.md)** - 完整的测试策略和实践
- **[单元测试](./development/unit-testing.md)** - 单元测试编写规范
- **[集成测试](./development/integration-testing.md)** - 集成测试策略
- **[端到端测试](./development/e2e-testing.md)** - E2E 测试实践

### 质量保证
- **[代码质量](./development/code-quality.md)** - 代码质量标准和工具
- **[性能测试](./development/performance-testing.md)** - 性能测试和基准测试

---

## 🛠️ 工具与资源

### 开发工具
- **[工具指南](./tools/development-tools.md)** - 推荐的开发工具和配置
- **[CLI 工具](./tools/cli-tools.md)** - 命令行工具使用指南
- **[IDE 配置](./tools/ide-setup.md)** - 开发环境 IDE 配置

### 资源文件
- **[模板文件](./templates/)** - 各种模板文件和配置示例
- **[脚本工具](./scripts/)** - 自动化脚本和工具
- **[配置示例](./examples/)** - 配置文件示例

---

## 📚 文档维护

### 文档管理
- **[文档规范](./docs-standards.md)** - 文档编写和维护规范
- **[更新日志](./CHANGELOG.md)** - 文档更新历史记录
- **[贡献指南](./CONTRIBUTING.md)** - 文档贡献指南

### 迁移记录
- **[迁移通知](./MIGRATION_NOTICE.md)** - 文档迁移详细说明
- **[迁移分析](./docs/MIGRATION_NOTICE.md)** - 原文档目录迁移通知

---

## 🔍 搜索与索引

### 快速查找
- **按角色查找**: 使用 [文档导航](./NAVIGATION.md) 按角色快速定位
- **按主题查找**: 使用 [文档索引](./INDEX.md) 按主题分类查找
- **按结构查找**: 使用 [站点地图](./SITEMAP.md) 浏览完整结构

### 搜索技巧
- **关键词搜索**: 使用 Ctrl+F 在页面内搜索关键词
- **标签搜索**: 查看文档标签进行分类搜索
- **交叉引用**: 利用文档间的交叉引用快速跳转

---

## 📱 移动端优化

本文档系统针对移动端进行了优化：
- **响应式设计**: 支持各种屏幕尺寸
- **快速加载**: 优化了文档加载速度
- **触摸友好**: 优化了移动端交互体验
- **离线访问**: 支持离线查看核心文档

---

## 🆘 获取帮助

### 技术支持
- **问题反馈**: 通过 GitHub Issues 提交问题
- **功能建议**: 通过 GitHub Discussions 提出建议
- **文档改进**: 通过 Pull Request 贡献文档

### 联系方式
- **技术交流群**: [加入技术交流群]
- **邮件支持**: support@laojun.com
- **在线文档**: [在线文档地址]

---

## 📈 文档统计

- **📄 总文档数**: 60+ 个文档文件
- **📁 目录数量**: 25+ 个分类目录
- **🏷️ 文档类型**: 10 个主要类别
- **🌍 支持语言**: 中文 (主要), English (部分)
- **📊 更新频率**: 持续更新维护

---

## 🎉 开始使用

1. **新用户**: 从 [快速开始](#-快速开始) 开始
2. **开发者**: 查看 [开发指南](#-开发指南) 和 [API 文档](#-api-文档)
3. **架构师**: 阅读 [架构设计](#️-架构设计) 和 [技术分析](#-技术分析)
4. **运维人员**: 参考 [部署运维](#-部署运维) 和 [监控运维](#-监控运维)

---

<div align="center">

**太上老君系统 - 让微服务开发更简单** 🚀

[![GitHub](https://img.shields.io/badge/GitHub-Repository-black?style=flat-square&logo=github)](https://github.com/your-org/laojun)
[![Documentation](https://img.shields.io/badge/Documentation-Latest-blue?style=flat-square&logo=gitbook)](./README.md)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](./LICENSE)

</div>