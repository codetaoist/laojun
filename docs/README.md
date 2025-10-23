# 太上老君系统文档

欢迎使用太上老君系统！本文档提供了系统的完整使用指南、部署说明和开发文档。

## 📚 文档导航

### 🚀 快速开始

| 文档 | 描述 |
|------|------|
| [配置指南](getting-started/configuration.md) | 环境变量和系统配置说明 |
| [部署概览](deployment.md) | 部署方式总览和快速入门 |

### 🏗️ 系统架构

| 文档 | 描述 |
|------|------|
| [架构概览](architecture/overview.md) | 系统整体架构设计和组件说明 |

### 🚢 部署指南

| 文档 | 描述 |
|------|------|
| [部署总览](deployment.md) | 部署方式概述和本地开发环境 |
| [Docker 部署](deployment/docker.md) | 容器化部署详细指南 |
| [Kubernetes 部署](deployment/kubernetes.md) | K8s 集群部署完整方案 |

### 🔌 API 文档

| 文档 | 描述 |
|------|------|
| [API 概览](api.md) | API 架构和使用指南 |
| [认证接口](api/endpoints/auth.md) | 用户认证相关接口 |
| [Swagger 配置](api/swagger/README.md) | Swagger 文档配置和使用说明 |

### 🗄️ 数据库

| 文档 | 描述 |
|------|------|
| [数据库迁移](database/migrations.md) | 数据库版本管理和迁移 |
| [命名约定](database/conventions.md) | 数据库表和字段命名规范 |

### 🔧 运维管理

| 文档 | 描述 |
|------|------|
| [监控指南](operations/monitoring.md) | 系统监控和告警配置 |

### 👨‍💻 开发指南

| 文档 | 描述 |
|------|------|
| 开发规范 | 代码规范和最佳实践（待补充） |
| 测试指南 | 单元测试和集成测试（待补充） |

## 🎯 快速导航

### 我是新用户
1. 📖 阅读 [配置指南](getting-started/configuration.md) 了解系统配置
2. 🚀 按照 [部署概览](deployment.md) 搭建本地开发环境
3. 🔍 查看 [API 概览](api.md) 了解接口使用方法

### 我要部署到生产环境
1. 🐳 选择 [Docker 部署](deployment/docker.md) 进行容器化部署
2. ☸️ 或选择 [Kubernetes 部署](deployment/kubernetes.md) 进行集群部署
3. 📊 配置 [监控指南](operations/monitoring.md) 确保系统稳定运行

### 我要开发新功能
1. 🏗️ 了解 [架构概览](architecture/overview.md) 掌握系统设计
2. 🗄️ 查看 [数据库文档](database/) 了解数据模型
3. 🔌 参考 [API 文档](api/) 设计接口规范

### 我遇到了问题
1. 🔍 查看相关模块的文档
2. 📊 检查 [监控指南](operations/monitoring.md) 排查问题
3. 🐛 在 GitHub 提交 Issue

## 📋 系统概述

太上老君系统是一个现代化的微服务架构应用，具有以下特点：

### 🎯 核心特性
- **微服务架构**: 模块化设计，易于扩展和维护
- **RESTful API**: 标准化的 API 接口设计
- **JWT 认证**: 安全的用户认证和授权机制
- **数据库迁移**: 自动化的数据库版本管理
- **容器化部署**: 支持 Docker 和 Kubernetes 部署
- **监控告警**: 完整的系统监控和告警体系

### 🛠️ 技术栈
- **后端**: Go 1.19+
- **数据库**: PostgreSQL 13+
- **缓存**: Redis 6+
- **容器**: Docker & Kubernetes
- **监控**: Prometheus & Grafana
- **文档**: Swagger/OpenAPI

### 🌟 系统优势
- **高性能**: Go 语言高并发特性
- **高可用**: 支持集群部署和自动故障转移
- **易维护**: 清晰的代码结构和完整的文档
- **易扩展**: 微服务架构支持水平扩展
- **易部署**: 容器化部署，一键启动

## 🔗 相关链接

- **项目仓库**: [GitHub](https://github.com/your-org/taishanglaojun)
- **在线文档**: [https://docs.laojun.dev](https://docs.laojun.dev)
- **API 文档**: [http://localhost:8080/docs](http://localhost:8080/docs) (本地开发)
- **问题反馈**: [GitHub Issues](https://github.com/your-org/taishanglaojun/issues)
- **技术支持**: support@laojun.dev

## 📝 文档贡献

我们欢迎社区贡献文档！如果您发现文档有误或需要补充，请：

1. Fork 项目仓库
2. 创建文档分支
3. 修改或添加文档
4. 提交 Pull Request

### 文档规范
- 使用 Markdown 格式
- 遵循现有的文档结构
- 添加适当的示例代码
- 保持语言简洁明了

## 📄 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](../LICENSE) 文件。

## 🙏 致谢

感谢所有为太上老君系统做出贡献的开发者和用户！

---

**最后更新**: 2023-12-01  
**文档版本**: v1.2.0  
**系统版本**: v1.2.0