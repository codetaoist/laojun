# 项目目录结构说明

## 根目录结构

太上老君微服务平台采用标准的微服务架构，每个服务独立部署和管理：

```
├── .github/                 # GitHub Actions CI/CD 配置
├── .vscode/                 # VS Code 开发环境配置
├── README.md                # 项目主文档
├── CONTRIBUTING.md          # 贡献指南
├── LICENSE                  # 开源协议
│
├── laojun-admin-api/        # 管理后台 API 服务
├── laojun-admin-web/        # 管理后台前端服务
├── laojun-config-center/    # 配置中心服务
├── laojun-deploy/           # 统一部署中心
├── laojun-discovery/        # 服务发现注册中心
├── laojun-docs/             # 项目文档中心
├── laojun-gateway/          # API 网关服务
├── laojun-marketplace-api/  # 市场服务 API
├── laojun-marketplace-web/  # 市场服务前端
├── laojun-monitoring/       # 监控告警服务
├── laojun-plugins/          # 插件系统
├── laojun-shared/           # 共享组件库
└── laojun-workspace/        # 开发环境工作空间
```

## 微服务架构原则

### 1. 服务独立性
- 每个 `laojun-*` 目录代表一个独立的微服务
- 每个服务有自己的代码、配置、部署文件
- 服务间通过 API 进行通信，避免直接依赖

### 2. 统一命名规范
- 所有服务以 `laojun-` 前缀命名
- 服务名称清晰表达其功能职责
- 避免缩写，使用完整的英文单词

### 3. 清晰的职责分工
- **业务服务**: admin-api, admin-web, marketplace-api, marketplace-web
- **平台服务**: config-center, discovery, gateway, monitoring
- **支撑服务**: deploy, docs, plugins, shared, workspace

## 架构优化说明

### 2025年10月 - 根目录结构优化

#### 已清理的不合理目录
- **configs/** - 全局配置目录，违反微服务独立性原则
- **db/** - 全局数据库目录，应由各服务独立管理
- **deploy/** - 与 laojun-deploy 功能重复
- **logs/** - 全局日志目录，应由各服务独立管理

#### 优化原则
1. **消除全局依赖**: 删除所有全局共享的配置、数据库、日志目录
2. **服务独立性**: 每个服务管理自己的配置、数据、日志
3. **避免重复**: 清理与现有服务功能重复的目录
4. **标准化命名**: 统一使用 `laojun-` 前缀

#### 优化效果
- ✅ 符合微服务架构最佳实践
- ✅ 提高服务独立性和可维护性
- ✅ 减少服务间耦合
- ✅ 便于独立部署和扩展

### 历史清理记录
- 删除了16个临时测试和调试文件（如 `check_*.go`, `test_*.go`, `debug_*.go`）
- 删除了临时测试结果文件 `test_results.json`
- 合并了重复的 `test/` 目录到 `tests/` 目录

## 目录用途说明

- **cmd/**: 应用程序的主要入口点，每个子目录包含一个可执行程序
- **internal/**: 应用程序的内部代码，不对外暴露
- **pkg/**: 可以被外部应用程序使用的库代码
- **scripts/**: 数据库迁移、修复和维护脚本
- **sql/**: 数据库相关的SQL脚本文件
- **tests/**: 所有测试文件，包括单元测试、集成测试和性能测试
- **web/**: 前端应用程序（admin 和 marketplace）
- **docs/**: 项目文档和API文档
- **configs/**: 应用程序配置文件
- **deployments/**: 部署相关的配置文件
- **docker/**: Docker 容器配置
- **k8s/**: Kubernetes 部署配置

这种结构遵循了 Go 项目的标准布局，提高了代码的可维护性和可读性。