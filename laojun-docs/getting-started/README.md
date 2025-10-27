# 快速开始

欢迎使用太上老君系统！本指南将帮助您快速搭建开发环境并开始使用系统。

## 📋 前置要求

在开始之前，请确保您的开发环境满足以下要求：

### 必需组件
- **Go**: 1.19 或更高版本
- **PostgreSQL**: 13 或更高版本  
- **Redis**: 6 或更高版本
- **Git**: 用于代码管理

### 可选组件
- **Docker**: 20.10+ (用于容器化开发)
- **Node.js**: 16+ (如果需要前端开发)
- **Make**: 用于构建脚本

### 系统要求
- **操作系统**: Linux, macOS, Windows
- **内存**: 最少 4GB
- **存储**: 最少 10GB 可用空间
- **网络**: 稳定的互联网连接

## 🚀 5分钟快速启动

### 1. 克隆项目

```bash
git clone https://github.com/your-org/taishanglaojun.git
cd taishanglaojun
```

### 2. 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 验证依赖安装
go mod verify
```

### 3. 配置环境

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
vim .env  # 或使用您喜欢的编辑器
```

基础配置示例：
```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=laojun
DB_PASSWORD=your_password
DB_NAME=laojun

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379

# 应用配置
APP_PORT=8080
APP_ENV=development
JWT_SECRET=your_jwt_secret_key_here
```

详细配置说明请参考 [配置指南](configuration.md)。

### 4. 初始化数据库

```bash
# 创建数据库
createdb laojun

# 运行数据库迁移
make migrate-up

# 或手动运行
go run cmd/migrate/main.go up
```

### 5. 启动服务

```bash
# 使用 Makefile
make run

# 或直接运行
go run main.go
```

### 6. 验证安装

服务启动后，访问以下地址验证安装：

- **健康检查**: http://localhost:8080/health
- **API 文档**: http://localhost:8080/docs
- **API 规范**: http://localhost:8080/swagger.json
- **Swagger UI**: http://localhost:8080/swagger-ui

## 🐳 使用 Docker 快速启动

如果您更喜欢使用 Docker，可以通过以下方式快速启动：

### 1. 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f laojun-api
```

### 2. 初始化数据库

```bash
# 运行数据库迁移
docker-compose exec laojun-api make migrate-up
```

### 3. 访问服务

- **API 服务**: http://localhost:8080
- **数据库**: localhost:5432
- **Redis**: localhost:6379

## 📖 下一步

恭喜！您已经成功启动了太上老君系统。接下来您可以：

### 🔍 探索 API
- 访问 [API 概览](../api.md) 了解接口设计
- 查看 [认证接口](../api/endpoints/auth.md) 学习用户认证
- 使用 Swagger UI 测试接口 (http://localhost:8080/swagger-ui)

### 🏗️ 了解架构
- 阅读 [架构概览](../architecture/overview.md) 理解系统设计
- 查看 [数据库设计](../database/) 了解数据模型

### 🚀 部署到生产
- 学习 [Docker 部署](../deployment/docker.md) 进行容器化部署
- 了解 [Kubernetes 部署](../deployment/kubernetes.md) 进行集群部署

### 🔧 开发新功能
- 查看开发规范和最佳实践
- 了解测试框架和测试方法
- 学习代码贡献流程

## 🛠️ 开发工具推荐

### IDE 和编辑器
- **VS Code**: 推荐安装 Go 扩展
- **GoLand**: JetBrains 的 Go IDE
- **Vim/Neovim**: 配置 Go 插件

### 有用的工具
- **Postman**: API 测试工具
- **TablePlus**: 数据库管理工具
- **Redis Desktop Manager**: Redis 管理工具
- **Docker Desktop**: Docker 图形界面

### Go 工具链
```bash
# 安装有用的 Go 工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## 🐛 常见问题

### Q: 数据库连接失败
**A**: 检查以下几点：
- PostgreSQL 服务是否启动
- 数据库用户和密码是否正确
- 数据库是否已创建
- 防火墙是否阻止连接

### Q: Redis 连接失败
**A**: 确认：
- Redis 服务是否运行
- 端口 6379 是否可访问
- Redis 配置是否正确

### Q: 端口被占用
**A**: 解决方法：
- 修改 `.env` 文件中的 `APP_PORT`
- 或停止占用端口的其他服务

### Q: 依赖下载失败
**A**: 尝试：
- 设置 Go 代理：`go env -w GOPROXY=https://goproxy.cn,direct`
- 清理模块缓存：`go clean -modcache`
- 重新下载：`go mod download`

## 📞 获取帮助

如果您遇到问题，可以通过以下方式获取帮助：

- **文档**: 查看相关模块的详细文档
- **GitHub Issues**: [提交问题](https://github.com/your-org/taishanglaojun/issues)
- **社区讨论**: [GitHub Discussions](https://github.com/your-org/taishanglaojun/discussions)
- **邮件支持**: support@laojun.dev

## 🎉 欢迎贡献

我们欢迎任何形式的贡献！无论是：
- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码

请查看 [贡献指南](../CONTRIBUTING.md) 了解详细信息。

---

**下一步**: [配置指南](configuration.md) | [部署指南](../deployment.md) | [API 文档](../api.md)