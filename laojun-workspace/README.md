# Laojun 开发环境管理工作区

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/laojun/laojun-workspace/workflows/CI/badge.svg)](https://github.com/laojun/laojun-workspace/actions)

🎯 **专注开发环境管理的统一工作区**

Laojun Workspace 是专为 Laojun 微服务平台设计的开发环境管理工具，提供统一的开发体验、代码质量保障和本地调试环境。

> 📢 **重要说明**: 本工作区专注于开发环境管理，生产部署请使用 [laojun-deploy](../laojun-deploy/) 统一部署方案。

## 🎯 核心定位

### 🛠️ 开发环境管理
- **统一工作区**: 使用 Go Workspace 管理多个微服务模块
- **开发工具集成**: 集成 linting、格式化、测试等开发工具
- **本地调试环境**: 提供完整的本地开发和调试环境
- **依赖管理**: 统一管理 Go 模块依赖和版本

### 📊 开发监控和观测
- **本地监控栈**: Prometheus、Grafana、Jaeger 本地实例
- **开发日志聚合**: Loki 本地日志收集和查询
- **性能分析**: 集成 pprof 和性能分析工具
- **健康检查**: 开发环境服务健康状态监控

### 🔒 代码质量保障
- **安全扫描**: 集成 Gosec、Trivy 等安全扫描工具
- **代码质量**: SonarQube 代码质量分析
- **测试管理**: 统一的测试执行和覆盖率报告
- **CI/CD 协调**: 与 GitHub Actions 集成的开发工作流

## 📁 开发工作区结构

```
laojun-workspace/           # 🏠 开发环境管理中心
├── configs/                # ⚙️ 开发环境配置
│   ├── dev.yaml               # 本地开发配置
│   ├── tools.yaml             # 开发工具配置
│   └── quality.yaml           # 代码质量配置
├── docs/                   # 📚 开发文档
│   ├── development-guide.md    # 开发指南和最佳实践
│   ├── branching-strategy.md   # Git 分支管理策略
│   ├── code-review-guidelines.md # 代码审查指南
│   ├── security-guide.md       # 安全开发指南
│   └── debugging-guide.md     # 调试和故障排除
├── scripts/                # 🔧 开发自动化脚本
│   ├── setup.ps1              # 开发环境初始化
│   ├── build-all.ps1          # 批量构建工具
│   ├── test-all.ps1           # 统一测试执行
│   ├── dev-tools.ps1          # 开发工具管理
│   ├── quality-check.ps1      # 代码质量检查
│   └── dependency-sync.ps1    # 依赖同步管理
├── monitoring/             # 📊 本地监控环境
│   ├── docker-compose.yml     # 监控服务栈
│   ├── configs/               # 监控工具配置
│   │   ├── prometheus.yml     # Prometheus 配置
│   │   ├── grafana/           # Grafana 仪表板
│   │   └── jaeger.yml         # 链路追踪配置
│   └── start-monitoring.ps1   # 监控环境启动
├── security/               # 🔒 代码安全和质量
│   ├── configs/               # 安全工具配置
│   ├── scripts/               # 安全扫描脚本
│   └── reports/               # 扫描报告存储
├── .github/                # 🚀 CI/CD 工作流
│   └── workflows/             # GitHub Actions 配置
├── go.work                 # 📦 Go 工作区配置
└── README.md               # 📖 本文档
```

> 💡 **设计理念**: 专注于提供高效的本地开发体验，所有配置和工具都围绕开发阶段的需求设计。

## 🏗️ 核心服务

| 服务 | 描述 | 端口 | 状态 |
|------|------|------|------|
| **laojun-admin-api** | 管理后台 API 服务 | 8080 | ✅ |
| **laojun-admin-web** | 管理后台前端应用 | 3000 | ✅ |
| **laojun-config-center** | 配置中心服务 | 8082 | ✅ |
| **laojun-marketplace-api** | 市场 API 服务 | 8081 | ✅ |
| **laojun-marketplace-web** | 市场前端应用 | 3001 | ✅ |
| **laojun-plugins** | 插件系统 | - | ✅ |
| **laojun-shared** | 共享组件库 | - | ✅ |

## 🚀 开发环境快速启动

### 📋 开发环境要求

- **Go**: 1.21+ (微服务开发)
- **Node.js**: 18+ (前端开发工具)
- **Docker**: 20.10+ (本地服务和监控)
- **Git**: 2.30+ (版本控制)
- **PowerShell**: 5.1+ (自动化脚本)

### 1️⃣ 初始化开发工作区

```powershell
# 克隆开发工作区
git clone <repository-url>
cd laojun-workspace

# 一键设置开发环境
.\scripts\setup.ps1
```

### 2️⃣ 启动本地开发环境

```powershell
# 启动本地基础服务（数据库、Redis、消息队列）
.\scripts\dev-tools.ps1 -Action start-services

# 启动本地监控栈（Prometheus、Grafana、Jaeger）
.\monitoring\start-monitoring.ps1

# 验证开发环境
.\scripts\check-env.ps1
```

### 3️⃣ 开发工作流

```powershell
# 构建所有微服务
.\scripts\build-all.ps1

# 运行代码质量检查
.\scripts\quality-check.ps1

# 执行完整测试套件
.\scripts\test-all.ps1
```

### 4️⃣ 本地调试和监控

```powershell
# 访问本地监控面板
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
# Jaeger: http://localhost:16686

# 查看服务日志
.\scripts\dev-tools.ps1 -Action logs

# 性能分析
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 📊 本地开发监控

### 🎯 开发监控目标

本地监控栈专为开发阶段设计，帮助开发者：
- **实时观测**: 监控本地服务性能和健康状态
- **问题诊断**: 快速定位性能瓶颈和错误
- **开发调试**: 提供详细的链路追踪和日志分析
- **性能优化**: 分析代码性能和资源使用情况

### 🔧 监控服务访问

启动本地监控环境后，可访问：

| 服务 | 地址 | 用途 | 凭据 |
|------|------|------|------|
| **Grafana** | http://localhost:3000 | 可视化仪表板和告警 | admin/admin |
| **Prometheus** | http://localhost:9090 | 指标收集和查询 | - |
| **Jaeger** | http://localhost:16686 | 分布式链路追踪 | - |
| **Loki** | http://localhost:3100 | 日志聚合和查询 | - |

### 📈 开发监控最佳实践

- **指标埋点**: 在关键业务逻辑中添加自定义指标
- **链路追踪**: 使用 OpenTelemetry 进行请求链路追踪
- **日志规范**: 遵循结构化日志格式，便于查询和分析
- **性能分析**: 定期使用 pprof 分析内存和 CPU 使用

## 🔒 安全

### 安全扫描工具

- **Gosec**: Go 代码安全扫描
- **Trivy**: 漏洞和密钥扫描
- **SonarQube**: 代码质量和安全分析

### 运行安全扫描

```powershell
# 运行所有安全扫描
.\security\scripts\security-scan.ps1

# 运行特定类型扫描
.\security\scripts\security-scan.ps1 -ScanType code
.\security\scripts\security-scan.ps1 -ScanType dependencies
```

## 🔧 开发工具

### 可用脚本

| 脚本 | 描述 |
|------|------|
| `setup.ps1` | 初始化开发环境 |
| `build-all.ps1` | 构建所有服务 |
| `test-all.ps1` | 运行所有测试 |
| `clean-all.ps1` | 清理构建产物 |
| `dev-tools.ps1` | 开发工具管理 |
| `version-manager.ps1` | 版本管理 |
| `dependency-manager.ps1` | 依赖管理 |

### 开发工具管理

```powershell
# 启动开发服务
.\scripts\dev-tools.ps1 -Action start

# 停止开发服务
.\scripts\dev-tools.ps1 -Action stop

# 重启开发服务
.\scripts\dev-tools.ps1 -Action restart

# 查看服务状态
.\scripts\dev-tools.ps1 -Action status
```

## 📚 文档

- [开发指南](docs/development-guide.md) - 详细的开发指南和最佳实践
- [分支管理策略](docs/branching-strategy.md) - Git 分支管理和发布流程
- [代码审查指南](docs/code-review-guidelines.md) - 代码审查标准和流程
- [安全配置指南](docs/security-guide.md) - 安全工具配置和使用
- [GitHub Secrets 配置](docs/github-secrets-guide.md) - CI/CD 密钥配置

## 🔄 标准开发工作流

### 📝 日常开发流程

1. **环境准备**: 使用 `setup.ps1` 初始化本地开发环境
2. **代码开发**: 在对应的微服务模块中进行功能开发
3. **质量检查**: 运行 `quality-check.ps1` 进行代码质量和安全扫描
4. **本地测试**: 使用 `test-all.ps1` 执行完整测试套件
5. **构建验证**: 使用 `build-all.ps1` 验证构建过程
6. **提交代码**: 遵循 Git 工作流提交代码变更

### 🚀 发布准备流程

1. **依赖同步**: 使用 `dependency-sync.ps1` 同步所有模块依赖
2. **集成测试**: 在本地环境运行完整的集成测试
3. **性能验证**: 使用监控工具验证性能指标
4. **文档更新**: 更新相关的开发文档和 API 文档
5. **版本标记**: 为发布版本创建 Git 标签

> 📢 **部署说明**: 开发完成后的部署工作请转到 [laojun-deploy](../laojun-deploy/) 进行统一部署管理。

## 🐛 故障排除

### 常见问题

1. **Go 模块依赖问题**
   ```powershell
   # 清理模块缓存
   go clean -modcache
   
   # 重新下载依赖
   go mod download
   ```

2. **Docker 服务启动失败**
   ```powershell
   # 检查端口占用
   netstat -ano | findstr :5432
   
   # 重启 Docker 服务
   docker-compose down && docker-compose up -d
   ```

3. **监控服务无法访问**
   ```powershell
   # 检查服务状态
   docker-compose ps
   
   # 查看服务日志
   docker-compose logs grafana
   ```

### 性能优化

- **构建优化**: 使用 Go 构建缓存和并行构建
- **测试优化**: 使用测试缓存和并行测试
- **开发优化**: 使用热重载和增量构建

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

1. **Fork** 项目到你的 GitHub 账户
2. **创建分支** (`git checkout -b feature/amazing-feature`)
3. **提交更改** (`git commit -m 'Add some amazing feature'`)
4. **推送分支** (`git push origin feature/amazing-feature`)
5. **创建 Pull Request**

### 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 添加适当的测试覆盖
- 更新相关文档

### 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
feat: 添加新功能
fix: 修复 bug
docs: 更新文档
style: 代码格式调整
refactor: 代码重构
test: 添加测试
chore: 构建过程或辅助工具的变动
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系我们

- **项目主页**: [GitHub Repository](https://github.com/laojun/laojun-workspace)
- **问题反馈**: [GitHub Issues](https://github.com/laojun/laojun-workspace/issues)
- **讨论交流**: [GitHub Discussions](https://github.com/laojun/laojun-workspace/discussions)

## 🙏 致谢

感谢所有为 Laojun 项目做出贡献的开发者！

---

**注意**: 确保在开发前阅读 [开发指南](docs/development-guide.md) 以了解详细的开发流程和最佳实践。