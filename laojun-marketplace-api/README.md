# Laojun Marketplace API 🛒

太上老君插件市场 API 服务

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![API Version](https://img.shields.io/badge/API-v1.0-green.svg)](./docs/api.md)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE)

## 📋 概述

本服务是太上老君平台的插件市场核心 API，提供完整的插件生态功能，包括插件展示、搜索、下载、评价、支付等。采用微服务架构设计，支持高并发访问和弹性扩展。

## ✨ 功能特性

### 🔌 插件管理
- **插件展示**: 插件列表、详情展示、分类浏览
- **搜索功能**: 全文搜索、标签搜索、高级筛选
- **版本管理**: 多版本支持、版本比较、升级提醒
- **依赖管理**: 依赖检查、自动安装、冲突解决

### 👥 用户体验
- **用户评价**: 评分系统、评论管理、反馈收集
- **个人中心**: 购买历史、收藏管理、推荐算法
- **社交功能**: 分享插件、关注开发者、社区互动
- **通知系统**: 更新通知、活动推送、系统消息

### 💰 商业功能
- **支付处理**: 多种支付方式、订单管理、退款处理
- **开发者管理**: 开发者认证、收益分成、数据统计
- **营销工具**: 促销活动、优惠券、会员体系
- **数据分析**: 下载统计、用户行为、收益报表

### 🔧 技术特性
- **高性能**: 基于 Gin 框架，支持高并发
- **缓存优化**: 多级缓存，提升响应速度
- **搜索引擎**: Elasticsearch 全文搜索
- **文件存储**: 对象存储，支持 CDN 加速
- **监控**: 完整的监控和告警体系

## 🏗️ 技术架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │  Mobile Client  │    │   API Gateway   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │              Marketplace API                  │
         │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
         │  │  Handlers   │  │  Services   │  │ Repositories│ │
         │  └─────────────┘  └─────────────┘  └─────────────┘ │
         └───────────────────────┼───────────────────────┘
                                 │
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │ PostgreSQL  │    │Elasticsearch│    │ Object Store│
    └─────────────┘    └─────────────┘    └─────────────┘
```

## 📁 目录结构

```
laojun-marketplace-api/
├── cmd/                    # 应用入口
│   └── marketplace-api/   # 主程序
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
- Elasticsearch 7+

### 本地开发
```bash
# 1. 克隆项目
git clone https://github.com/codetaoist/laojun.git
cd laojun/laojun-marketplace-api

# 2. 安装依赖
go mod download

# 3. 配置环境
cp configs/config.example.yaml configs/config.yaml
# 编辑配置文件

# 4. 初始化数据库
make migrate

# 5. 启动服务
go run cmd/marketplace-api/main.go
```

### Docker 部署
```bash
# 构建镜像
docker build -t laojun-marketplace-api .

# 运行容器
docker run -p 8081:8081 laojun-marketplace-api
```

## 📚 API 文档

### 访问方式
- **Swagger UI**: http://localhost:8081/swagger/index.html
- **API 文档**: [docs/api.md](./docs/api.md)
- **Postman Collection**: [docs/postman/](./docs/postman/)

### 主要接口
- `GET /api/v1/plugins` - 插件列表
- `GET /api/v1/plugins/{id}` - 插件详情
- `POST /api/v1/plugins/{id}/download` - 下载插件
- `POST /api/v1/plugins/{id}/reviews` - 提交评价
- `GET /api/v1/search` - 搜索插件

## 🔧 配置说明

### 主要配置项
```yaml
server:
  port: 8081
  mode: debug

database:
  host: localhost
  port: 5432
  name: laojun_marketplace
  
redis:
  host: localhost
  port: 6379
  
elasticsearch:
  host: localhost
  port: 9200
  
storage:
  type: s3
  bucket: laojun-plugins
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
- `marketplace_api_requests_total` - 请求总数
- `marketplace_api_request_duration_seconds` - 请求耗时
- `marketplace_api_plugin_downloads` - 插件下载数
- `marketplace_api_search_queries` - 搜索查询数

## 🔗 相关链接

- [项目主页](../README.md)
- [API 文档](../../docs/api/README.md)
- [部署指南](../../docs/deployment/README.md)
- [开发指南](../../docs/development/README.md)