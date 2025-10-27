# Laojun Admin API 🛡️

太上老君平台管理后台 API 服务

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![API Version](https://img.shields.io/badge/API-v1.0-green.svg)](./docs/api.md)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE)

## 📋 概述

本服务是太上老君平台的管理后台核心 API，提供完整的系统管理功能，包括用户管理、权限控制、角色管理、菜单管理、系统配置等。采用 Clean Architecture 设计，支持高并发和水平扩展。

## ✨ 功能特性

### 🔐 权限管理
- **用户管理**: 用户注册、登录、信息管理、状态控制
- **角色管理**: 角色创建、权限分配、层级管理
- **权限控制**: 基于 RBAC 的细粒度权限控制
- **菜单管理**: 动态菜单配置、权限关联

### 📊 系统管理
- **系统配置**: 动态配置管理、实时更新
- **审计日志**: 操作日志记录、安全审计
- **数据统计**: 用户统计、操作统计、性能指标
- **监控告警**: 系统监控、异常告警

### 🔧 技术特性
- **高性能**: 基于 Gin 框架，支持高并发
- **缓存优化**: Redis 缓存，提升响应速度
- **数据库**: PostgreSQL 主从架构
- **监控**: Prometheus 指标采集
- **文档**: Swagger API 文档

## 🏗️ 技术架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   Load Balancer │    │   API Gateway   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                  Admin API                    │
         │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
         │  │  Handlers   │  │  Services   │  │ Repositories│ │
         │  └─────────────┘  └─────────────┘  └─────────────┘ │
         └───────────────────────┼───────────────────────┘
                                 │
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │ PostgreSQL  │    │    Redis    │    │ Prometheus  │
    └─────────────┘    └─────────────┘    └─────────────┘
```

## 📁 目录结构

```
laojun-admin-api/
├── cmd/                    # 应用入口
│   └── admin-api/         # 主程序
├── internal/              # 内部代码
│   ├── handlers/          # HTTP 处理器
│   ├── services/          # 业务逻辑层
│   ├── repositories/      # 数据访问层
│   ├── models/           # 数据模型
│   ├── middleware/       # 中间件
│   └── config/           # 配置管理
├── configs/              # 配置文件
│   ├── config.yaml       # 主配置文件
│   └── database.yaml     # 数据库配置
├── docs/                 # 文档
│   ├── api.md           # API 文档
│   └── deployment.md    # 部署文档
├── scripts/              # 脚本文件
├── tests/               # 测试文件
└── Dockerfile           # Docker 构建文件
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- PostgreSQL 13+
- Redis 6+

### 本地开发
```bash
# 1. 克隆项目
git clone https://github.com/codetaoist/laojun.git
cd laojun/laojun-admin-api

# 2. 安装依赖
go mod download

# 3. 配置环境
cp configs/config.example.yaml configs/config.yaml
# 编辑配置文件

# 4. 初始化数据库
make migrate

# 5. 启动服务
go run cmd/admin-api/main.go
```

### Docker 部署
```bash
# 构建镜像
docker build -t laojun-admin-api .

# 运行容器
docker run -p 8080:8080 laojun-admin-api
```

## 📚 API 文档

### 访问方式
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API 文档**: [docs/api.md](./docs/api.md)
- **Postman Collection**: [docs/postman/](./docs/postman/)

### 主要接口
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/users` - 用户列表
- `POST /api/v1/users` - 创建用户
- `GET /api/v1/roles` - 角色列表
- `POST /api/v1/roles` - 创建角色

## 🔧 配置说明

### 主要配置项
```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: laojun_admin
  
redis:
  host: localhost
  port: 6379
  
jwt:
  secret: your-secret-key
  expire: 24h
```

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
make test-integration

# 生成测试覆盖率报告
make test-coverage
```

## 📊 监控指标

服务提供以下 Prometheus 指标：
- `admin_api_requests_total` - 请求总数
- `admin_api_request_duration_seconds` - 请求耗时
- `admin_api_active_users` - 活跃用户数
- `admin_api_database_connections` - 数据库连接数

## 🔗 相关链接

- [项目主页](../README.md)
- [API 文档](../../docs/api/README.md)
- [部署指南](../../docs/deployment/README.md)
- [开发指南](../../docs/development/README.md)