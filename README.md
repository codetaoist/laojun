# Laojun 太上老君 - 云原生微服务平台

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Node.js Version](https://img.shields.io/badge/Node.js-18+-green.svg)](https://nodejs.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个现代化的云原生微服务平台，提供管理后台、插件市场、配置中心等核心功能，支持插件化扩展和多环境部署。

## ✨ 特性

- 🚀 **云原生架构**: 基于 Docker 和 Kubernetes 的容器化部署
- 🔧 **插件化设计**: 支持动态插件加载和管理
- 🎯 **统一管理**: 集成管理后台、插件市场、配置中心
- 🛡️ **安全可靠**: JWT 认证、RBAC 权限控制、验证码保护
- 📊 **监控完善**: 健康检查、指标监控、日志管理
- 🔄 **配置继承**: 支持多环境配置继承和覆盖
- 🧪 **测试完备**: 单元测试、集成测试、性能测试

## 🚀 快速启动

### 方案一：一键部署（推荐）

**Windows 用户：**
```batch
# 一键部署（推荐）
.\一键部署.bat

# 或使用 PowerShell 脚本
.\deploy\scripts\deploy.ps1 prod deploy
```

**Linux/macOS 用户：**
```bash
# 一键部署
./deploy/scripts/deploy.sh prod deploy
```

**部署完成后访问：**
- 🌐 **Swagger API 文档**: http://localhost:8080/swagger
- ⚙️ **配置中心**: http://localhost:8081
- 🏪 **插件市场**: http://localhost:8082
- 👨‍💼 **管理后台**: http://localhost:3000
- 🛒 **市场前端**: http://localhost:3001

### 方案二：Docker Compose（手动）

```bash
# 构建并启动所有服务
docker compose -f deploy/docker/docker-compose.yml up -d --build

# 查看服务状态
docker compose -f deploy/docker/docker-compose.yml ps

# 查看日志
docker compose -f deploy/docker/docker-compose.yml logs -f

# 停止服务
docker compose -f deploy/docker/docker-compose.yml down
```

### 方案三：本地开发

**前提条件：**
- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Redis 6+

**环境变量配置：**
```powershell
# 数据库配置
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="laojun"
$env:DB_PASSWORD="laojun123"
$env:DB_NAME="laojun"

# Redis 配置
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="6379"

# JWT 配置
$env:JWT_SECRET="your-super-secret-jwt-key"

# 安全配置
$env:SECURITY_ENABLE_CAPTCHA="true"
$env:SECURITY_CAPTCHA_TTL="5m"

# 开发模式
$env:GIN_MODE="debug"
```

**启动服务：**
```bash
# 1. 启动配置中心
go run ./cmd/config-center

# 2. 启动管理后台 API
go run ./cmd/admin-api

# 3. 启动插件市场 API
go run ./cmd/marketplace-api

# 4. 启动前端（新终端）
cd web/admin && npm install && npm run dev
cd web/marketplace && npm install && npm run dev
```

## 🛠️ 工具使用

项目提供了统一的命令行工具，支持 `-dry-run` 预览模式和详细帮助信息。

### 菜单管理工具
```bash
# 备份菜单数据
go run cmd/menu-manager/main.go -action=backup -file=menu_backup.sql

# 检查菜单完整性
go run cmd/menu-manager/main.go -action=check -dry-run

# 清理重复菜单
go run cmd/menu-manager/main.go -action=clean-duplicates -force
```

### 数据库维护工具
```bash
# 检查数据库架构
go run cmd/db-maintenance/main.go -action=check-schema -dry-run

# 修复表名
go run cmd/db-maintenance/main.go -action=fix-table-names

# 执行 SQL 文件
go run cmd/db-maintenance/main.go -action=execute-sql -sql-file=migration.sql
```

### 插件市场管理工具
```bash
# 运行市场迁移
go run cmd/marketplace-manager/main.go -action=migrate

# 填充演示数据
go run cmd/marketplace-manager/main.go -action=seed-demo
```

### 项目管理工具
```bash
# 组织根目录文件
go run cmd/project-manager/main.go -action=organize-root

# 生成完整迁移
go run cmd/project-manager/main.go -action=generate-migration

# 重置数据库
go run cmd/project-manager/main.go -action=reset-database -force
```

更多工具使用详情请参考：[工具使用指南](docs/tools-guide.md)

## 🧪 测试

项目提供了完整的测试框架，支持单元测试、集成测试和性能测试。

```bash
# 运行所有测试
go run tests/run_tests.go -type=all -v

# 运行单元测试
go run tests/run_tests.go -type=unit -v

# 运行集成测试
go run tests/run_tests.go -type=integration -v

# 生成覆盖率报告
go run tests/run_tests.go -type=unit -coverage -v
```

## 📁 项目结构

```
laojun/
├── api/                    # API 定义和文档
├── cmd/                    # 命令行工具和服务入口
│   ├── admin-api/         # 管理后台 API 服务
│   ├── config-center/     # 配置中心服务
│   ├── marketplace-api/   # 插件市场 API 服务
│   ├── menu-manager/      # 菜单管理工具
│   ├── db-maintenance/    # 数据库维护工具
│   ├── marketplace-manager/ # 插件市场管理工具
│   └── project-manager/   # 项目管理工具
├── configs/               # 配置文件
│   ├── base.yaml         # 基础配置（可继承）
│   ├── admin-api.yaml    # 管理 API 配置
│   └── environments/     # 环境特定配置
├── deploy/               # 部署相关文件
│   ├── docker/          # Docker 配置
│   ├── k8s/            # Kubernetes 配置
│   └── scripts/        # 部署脚本
├── docs/                # 文档
│   ├── api/            # API 文档
│   ├── deployment/     # 部署文档
│   ├── reports/        # 分析报告
│   └── tools-guide.md  # 工具使用指南
├── internal/           # 内部包
├── pkg/               # 公共包
│   ├── plugins/       # 插件系统
│   └── shared/        # 共享组件
├── tests/             # 测试文件
│   ├── unit/         # 单元测试
│   ├── integration/  # 集成测试
│   ├── performance/  # 性能测试
│   └── run_tests.go  # 测试运行脚本
└── web/              # 前端应用
    ├── admin/        # 管理后台前端
    └── marketplace/  # 插件市场前端
```

## ⚙️ 配置管理

项目采用配置继承机制，支持多环境配置管理：

- **基础配置**: `configs/base.yaml` - 所有环境的通用配置
- **服务配置**: `configs/{service}.yaml` - 特定服务配置，继承基础配置
- **环境配置**: `configs/environments/{env}.yaml` - 环境特定配置

配置优先级：环境配置 > 服务配置 > 基础配置

## 🔧 常用操作

### 部署管理
```bash
# 查看部署状态
docker compose ps

# 重启服务
docker compose restart

# 查看日志
docker compose logs -f [service_name]

# 资源监控
docker stats

# 健康检查
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
```

### 数据库管理
```bash
# 连接数据库
docker exec -it laojun-postgres psql -U laojun -d laojun

# 备份数据库
docker exec laojun-postgres pg_dump -U laojun laojun > backup.sql

# 恢复数据库
docker exec -i laojun-postgres psql -U laojun -d laojun < backup.sql
```

### 日志查看
```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f admin-api
docker compose logs -f config-center
docker compose logs -f marketplace-api
```

## 🚨 故障排除

### 常见问题

1. **端口冲突**
   - 检查端口占用：`netstat -ano | findstr :8080`
   - 修改配置文件中的端口设置

2. **数据库连接失败**
   - 确认 PostgreSQL 服务运行：`docker compose ps postgres`
   - 检查数据库配置和网络连接

3. **Redis 连接失败**
   - 确认 Redis 服务运行：`docker compose ps redis`
   - 检查 Redis 配置和连接参数

4. **前端访问失败**
   - 检查前端构建状态：`docker compose logs nginx`
   - 确认 API 服务正常运行

### 性能优化

- 调整数据库连接池大小
- 配置 Redis 缓存策略
- 优化 Nginx 配置
- 监控资源使用情况

## 📚 文档

- 📖 [详细部署指南](deploy/docs/README.md)
- 🐳 [Docker 使用指南](deploy/docs/docker-guide.md)
- 🔧 [工具使用指南](docs/tools-guide.md)
- 🏗️ [项目架构文档](docs/architecture/README.md)
- 📊 [API 文档](docs/api/README.md)
- 🔍 [优化报告](docs/reports/optimization-summary.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

感谢所有贡献者和开源社区的支持！

---

**快速链接**
- [🚀 快速开始](#-快速启动)
- [🛠️ 工具使用](#️-工具使用)
- [📁 项目结构](#-项目结构)
- [📚 完整文档](docs/README.md)