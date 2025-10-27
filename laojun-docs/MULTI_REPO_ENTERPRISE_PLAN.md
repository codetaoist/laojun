# Laojun 多仓库分离企业级方案

## 概述

本文档详细描述了 Laojun 平台从单体仓库（Monorepo）向多仓库（Multi-repo）架构迁移的企业级方案。该方案旨在提高代码管理效率、增强模块独立性、优化CI/CD流程，并为团队协作提供更好的支持。

## 目标架构

### 核心仓库结构

```
laojun-ecosystem/
├── laojun-shared/          # 共享组件库
├── laojun-core/           # 核心服务
├── laojun-marketplace/    # 市场服务
├── laojun-plugins/        # 插件系统
├── laojun-admin/          # 管理后台
├── laojun-web/            # 前端应用
├── laojun-deploy/         # 部署配置
├── laojun-docs/           # 文档中心
└── laojun-workspace/      # 开发工作区（可选）
```

### 仓库职责划分

#### 1. laojun-shared（共享组件库）
- **职责**: 提供跨服务的共享组件、工具类、模型定义
- **内容**:
  - 认证授权模块 (auth/)
  - 配置管理 (config/)
  - 数据库连接 (database/)
  - 缓存管理 (cache/)
  - 日志系统 (logger/)
  - 中间件 (middleware/)
  - 通用模型 (models/)
  - 工具函数 (utils/)
  - 健康检查 (health/)
  - 指标监控 (metrics/)
  - 测试工具 (testing/)

#### 2. laojun-core（核心服务）
- **职责**: 核心业务逻辑和API服务
- **内容**:
  - 用户管理
  - 权限控制
  - 系统配置
  - 核心API接口

#### 3. laojun-marketplace（市场服务）
- **职责**: 插件市场相关功能
- **内容**:
  - 插件展示
  - 插件下载
  - 评价系统
  - 支付处理

#### 4. laojun-plugins（插件系统）
- **职责**: 插件开发框架和运行时
- **内容**:
  - 插件SDK
  - 插件加载器
  - 插件注册中心
  - 插件安全管理

#### 5. laojun-admin（管理后台）
- **职责**: 系统管理和运维功能
- **内容**:
  - 管理界面
  - 系统监控
  - 用户管理
  - 插件审核

#### 6. laojun-web（前端应用）
- **职责**: 用户界面和前端资源
- **内容**:
  - 用户前端
  - 管理后台前端
  - 静态资源

#### 7. laojun-deploy（部署配置）
- **职责**: 部署和运维配置
- **内容**:
  - Docker配置
  - Kubernetes配置
  - CI/CD流水线
  - 环境配置

#### 8. laojun-docs（文档中心）
- **职责**: 项目文档和API文档
- **内容**:
  - 架构文档
  - API文档
  - 开发指南
  - 部署文档

## 当前根目录文件规划

### 需要迁移到各仓库的文件/目录

#### 迁移到 laojun-shared
- ✅ `repos/laojun-shared/` (已存在，需要完善)
- ❌ `pkg/shared/` (废弃，内容已合并到laojun-shared)

#### 迁移到 laojun-core
- ✅ `repos/laojun-core/` (已存在)
- 📁 `cmd/admin-api/` → `laojun-core/cmd/admin-api/`
- 📁 `cmd/config-center/` → `laojun-core/cmd/config-center/`
- 📁 `internal/` → `laojun-core/internal/`

#### 迁移到 laojun-marketplace
- ✅ `repos/laojun-marketplace/` (已存在)
- 📁 `cmd/marketplace-api/` → `laojun-marketplace/cmd/marketplace-api/`

#### 迁移到 laojun-plugins
- ✅ `repos/laojun-plugins/` (已存在)
- 📁 `pkg/plugins/` → `laojun-plugins/pkg/`

#### 迁移到 laojun-admin
- 📁 `web/admin/` → `laojun-admin/web/`
- 📁 `cmd/create-admin/` → `laojun-admin/cmd/create-admin/`
- 📁 `cmd/setup-super-admin/` → `laojun-admin/cmd/setup-super-admin/`
- 📁 `cmd/verify-super-admin/` → `laojun-admin/cmd/verify-super-admin/`

#### 迁移到 laojun-web
- 📁 `web/marketplace/` → `laojun-web/marketplace/`
- 📁 `api/` → `laojun-web/api/`

#### 迁移到 laojun-deploy
- 📁 `deploy/` → `laojun-deploy/`
- 📁 `configs/` → `laojun-deploy/configs/`
- 📄 `Makefile` → `laojun-deploy/Makefile`

#### 迁移到 laojun-docs
- 📁 `docs/` → `laojun-docs/`

### 保留在工作区的文件

#### 工作区管理文件
- 📄 `go.work` (工作区配置)
- 📄 `go.work.sum` (工作区依赖锁定)
- 📄 `README.md` (工作区说明)
- 📄 `.gitignore` (工作区忽略规则)

#### 开发工具和脚本
- 📁 `tools/` (开发工具)
- 📁 `tests/` (集成测试)
- 📄 `update_imports.ps1` (迁移脚本)

#### 数据和临时文件
- 📁 `db/` (数据库相关)
- 📁 `uploads/` (上传文件)
- 📁 `bin/` (编译输出)
- 📁 `misc/` (杂项文件)
- 📁 `etc/` (配置文件)

## 迁移策略

### 阶段一：准备阶段（1-2周）
1. **完善共享组件库**
   - ✅ 统一 `laojun-shared` 组件库
   - ✅ 废弃 `pkg/shared`
   - ✅ 更新所有引用

2. **创建独立仓库**
   - 在组织下创建各个独立仓库
   - 设置仓库权限和分支保护
   - 配置CI/CD基础设施

### 阶段二：核心迁移（2-3周）
1. **迁移核心服务**
   - 迁移 `laojun-core` 相关代码
   - 迁移 `laojun-marketplace` 相关代码
   - 更新依赖关系

2. **迁移插件系统**
   - 迁移 `laojun-plugins` 相关代码
   - 更新插件SDK

### 阶段三：前端和部署（1-2周）
1. **迁移前端应用**
   - 迁移管理后台到 `laojun-admin`
   - 迁移用户前端到 `laojun-web`

2. **迁移部署配置**
   - 迁移部署脚本到 `laojun-deploy`
   - 更新CI/CD配置

### 阶段四：文档和优化（1周）
1. **迁移文档**
   - 迁移文档到 `laojun-docs`
   - 更新README和架构文档

2. **优化和测试**
   - 端到端测试
   - 性能优化
   - 文档完善

## 技术实施细节

### 依赖管理
```go
// 各仓库的 go.mod 示例
module github.com/codetaoist/laojun-core

require (
    github.com/codetaoist/laojun-shared v1.0.0
    // 其他依赖...
)
```

### 版本管理策略
- **语义化版本**: 所有仓库采用语义化版本 (SemVer)
- **标签管理**: 使用 Git 标签管理版本发布
- **依赖锁定**: 使用 go.mod 锁定依赖版本

### CI/CD 流水线
```yaml
# .github/workflows/ci.yml 示例
name: CI/CD Pipeline
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      - run: go test ./...
      
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: docker build -t laojun-core .
```

## 工作区配置

### 开发工作区 (laojun-workspace)
```
laojun-workspace/
├── go.work                 # 工作区配置
├── README.md              # 工作区说明
├── scripts/               # 开发脚本
│   ├── setup.sh          # 环境设置
│   ├── build-all.sh      # 批量构建
│   └── test-all.sh       # 批量测试
├── tools/                 # 开发工具
├── tests/                 # 集成测试
└── docs/                  # 开发文档
```

### go.work 配置
```go
go 1.21

use (
    ./laojun-shared
    ./laojun-core
    ./laojun-marketplace
    ./laojun-plugins
    ./laojun-admin
)
```

## 团队协作模式

### 代码审查流程
1. **功能分支**: 每个功能在独立分支开发
2. **Pull Request**: 通过PR进行代码审查
3. **自动化测试**: CI/CD自动运行测试
4. **代码质量**: 使用代码质量检查工具

### 发布流程
1. **版本规划**: 制定版本发布计划
2. **集成测试**: 在工作区进行集成测试
3. **版本发布**: 按顺序发布各仓库版本
4. **部署验证**: 验证部署结果

## 监控和运维

### 仓库监控
- **依赖更新**: 监控依赖版本更新
- **安全扫描**: 定期进行安全漏洞扫描
- **性能监控**: 监控构建和测试性能

### 文档维护
- **API文档**: 自动生成和更新API文档
- **架构文档**: 定期更新架构文档
- **变更日志**: 维护详细的变更日志

## 风险评估和缓解

### 主要风险
1. **依赖复杂性**: 多仓库间依赖关系复杂
2. **版本兼容性**: 版本升级可能导致兼容性问题
3. **开发效率**: 初期可能影响开发效率

### 缓解措施
1. **依赖图管理**: 使用工具可视化依赖关系
2. **自动化测试**: 完善的自动化测试覆盖
3. **渐进式迁移**: 分阶段逐步迁移
4. **回滚计划**: 制定详细的回滚计划

## 成功指标

### 技术指标
- 构建时间减少 30%
- 测试执行时间减少 40%
- 代码复用率提高 50%

### 团队指标
- 开发效率提升 20%
- 代码审查时间减少 25%
- 发布频率提高 2倍

## 总结

多仓库分离方案将显著提升 Laojun 平台的可维护性、可扩展性和团队协作效率。通过合理的仓库划分、完善的工具支持和渐进式的迁移策略，可以确保迁移过程的平稳进行，并为未来的发展奠定坚实基础。

---

**文档版本**: v1.0  
**创建日期**: 2024年12月  
**最后更新**: 2024年12月  
**维护者**: Laojun 架构团队