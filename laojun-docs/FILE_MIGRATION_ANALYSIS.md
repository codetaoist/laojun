# Laojun 文件迁移归属分析

## 概述

本文档详细分析当前根目录下所有文件和目录的归属，为多仓库分离提供具体的迁移指导。

## 详细文件归属分析

### 1. 命令行工具 (cmd/)

#### 迁移到 laojun-core
```
cmd/admin-api/              → laojun-core/cmd/admin-api/
cmd/config-center/          → laojun-core/cmd/config-center/
cmd/db-complete-migrate/    → laojun-core/cmd/db-migrate/
cmd/db-maintenance/         → laojun-core/cmd/db-maintenance/
cmd/db-manager/             → laojun-core/cmd/db-manager/
cmd/db-migrate/             → laojun-core/cmd/db-migrate/
cmd/debug-config/           → laojun-core/cmd/debug-config/
cmd/fix-migration/          → laojun-core/cmd/fix-migration/
cmd/generate_hash/          → laojun-core/cmd/generate_hash/
cmd/migrate/                → laojun-core/cmd/migrate/
cmd/run_migration/          → laojun-core/cmd/run_migration/
```

#### 迁移到 laojun-marketplace
```
cmd/marketplace-api/        → laojun-marketplace/cmd/marketplace-api/
cmd/marketplace-manager/    → laojun-marketplace/cmd/marketplace-manager/
```

#### 迁移到 laojun-admin
```
cmd/create-admin/           → laojun-admin/cmd/create-admin/
cmd/setup-super-admin/      → laojun-admin/cmd/setup-super-admin/
cmd/verify-super-admin/     → laojun-admin/cmd/verify-super-admin/
cmd/menu-manager/           → laojun-admin/cmd/menu-manager/
cmd/project-manager/        → laojun-admin/cmd/project-manager/
```

### 2. 内部模块 (internal/)

#### 迁移到 laojun-core
```
internal/cache/             → laojun-core/internal/cache/
internal/config/            → laojun-core/internal/config/
internal/database/          → laojun-core/internal/database/
internal/server/            → laojun-core/internal/server/
internal/storage/           → laojun-core/internal/storage/
```

#### 按功能模块分离的handlers
```
# 核心认证和用户管理 → laojun-core
internal/handlers/auth_handler.go              → laojun-core/internal/handlers/
internal/handlers/user_handler.go              → laojun-core/internal/handlers/
internal/handlers/user_handler_test.go         → laojun-core/internal/handlers/
internal/handlers/role_handler.go              → laojun-core/internal/handlers/
internal/handlers/permission_handler.go        → laojun-core/internal/handlers/
internal/handlers/config.go                    → laojun-core/internal/handlers/
internal/handlers/config_handler_test.go       → laojun-core/internal/handlers/
internal/handlers/system_handler.go            → laojun-core/internal/handlers/

# 市场相关 → laojun-marketplace
internal/handlers/marketplace_auth_handler.go  → laojun-marketplace/internal/handlers/
internal/handlers/plugin_handler.go            → laojun-marketplace/internal/handlers/
internal/handlers/plugin_review_handler.go     → laojun-marketplace/internal/handlers/
internal/handlers/review_handler.go            → laojun-marketplace/internal/handlers/
internal/handlers/category_handler.go          → laojun-marketplace/internal/handlers/
internal/handlers/developer_handler.go         → laojun-marketplace/internal/handlers/
internal/handlers/rate_limit_handler.go        → laojun-marketplace/internal/handlers/

# 管理后台 → laojun-admin
internal/handlers/admin_plugin_handler.go      → laojun-admin/internal/handlers/
internal/handlers/extended_plugin_handler.go   → laojun-admin/internal/handlers/
internal/handlers/menu_handler.go              → laojun-admin/internal/handlers/
internal/handlers/icon_handler.go              → laojun-admin/internal/handlers/

# 社区功能 → laojun-web
internal/handlers/community_handler.go         → laojun-web/internal/handlers/

# 通用功能 → laojun-shared
internal/handlers/swagger.go                   → laojun-shared/handlers/
```

#### 按功能模块分离的services
```
# 核心服务 → laojun-core
internal/services/auth_service.go              → laojun-core/internal/services/
internal/services/auth_service_test.go         → laojun-core/internal/services/
internal/services/user_service.go              → laojun-core/internal/services/
internal/services/user_service_test.go         → laojun-core/internal/services/
internal/services/role_service.go              → laojun-core/internal/services/
internal/services/permission_service.go        → laojun-core/internal/services/
internal/services/config_service_test.go       → laojun-core/internal/services/
internal/services/system_service.go            → laojun-core/internal/services/
internal/services/jwt_key_service.go           → laojun-core/internal/services/
internal/services/audit_service.go             → laojun-core/internal/services/

# 市场服务 → laojun-marketplace
internal/services/plugin_service.go            → laojun-marketplace/internal/services/
internal/services/plugin_service_test.go       → laojun-marketplace/internal/services/
internal/services/plugin_review_service.go     → laojun-marketplace/internal/services/
internal/services/review_service.go            → laojun-marketplace/internal/services/
internal/services/category_service.go          → laojun-marketplace/internal/services/
internal/services/developer_service.go         → laojun-marketplace/internal/services/

# 管理服务 → laojun-admin
internal/services/admin_auth_service.go        → laojun-admin/internal/services/
internal/services/admin_plugin_service.go      → laojun-admin/internal/services/
internal/services/extended_plugin_service.go   → laojun-admin/internal/services/
internal/services/menu_service.go              → laojun-admin/internal/services/
internal/services/icon_service.go              → laojun-admin/internal/services/

# 社区服务 → laojun-web
internal/services/community_service.go         → laojun-web/internal/services/
```

#### 按功能模块分离的models
```
# 核心模型 → laojun-core
internal/models/auth.go                        → laojun-core/internal/models/
internal/models/user.go                        → laojun-core/internal/models/
internal/models/role.go                        → laojun-core/internal/models/
internal/models/permission.go                  → laojun-core/internal/models/
internal/models/system.go                      → laojun-core/internal/models/

# 市场模型 → laojun-marketplace
internal/models/plugin.go                      → laojun-marketplace/internal/models/
internal/models/plugin_extended.go             → laojun-marketplace/internal/models/
internal/models/plugin_review.go               → laojun-marketplace/internal/models/

# 管理模型 → laojun-admin
internal/models/menu.go                        → laojun-admin/internal/models/
internal/models/icon.go                        → laojun-admin/internal/models/
```

#### 中间件分离
```
# 核心中间件 → laojun-core
internal/middleware/auth.go                    → laojun-core/internal/middleware/
internal/middleware/permission.go              → laojun-core/internal/middleware/
internal/middleware/role_validation.go         → laojun-core/internal/middleware/

# 通用中间件 → laojun-shared
internal/middleware/cors.go                    → laojun-shared/middleware/
internal/middleware/rate_limit.go              → laojun-shared/middleware/
```

#### 路由分离
```
# 核心路由 → laojun-core
internal/routes/routes.go                      → laojun-core/internal/routes/

# 管理路由 → laojun-admin
internal/routes/admin_plugin_routes.go         → laojun-admin/internal/routes/
internal/routes/extended_plugin_routes.go      → laojun-admin/internal/routes/
```

#### 客户端和插件系统
```
# 客户端 → 各自对应的仓库
internal/clients/marketplace_client.go         → laojun-core/internal/clients/

# 插件系统 → laojun-plugins
internal/plugin/                               → laojun-plugins/internal/plugin/
```

### 3. 前端应用 (web/)

```
web/admin/                  → laojun-admin/web/
web/marketplace/            → laojun-web/marketplace/
```

### 4. API文档 (api/)

```
api/                        → laojun-docs/api/
```

### 5. 部署配置 (deploy/)

```
deploy/                     → laojun-deploy/
```

### 6. 配置文件 (configs/)

```
configs/                    → laojun-deploy/configs/
```

### 7. 文档 (docs/)

```
docs/                       → laojun-docs/
```

### 8. 构建工具

```
Makefile                    → laojun-deploy/Makefile
```

### 9. 已存在的仓库目录

```
repos/laojun-shared/        → 保持独立，作为共享库
repos/laojun-core/          → 保持独立
repos/laojun-marketplace/   → 保持独立
repos/laojun-plugins/       → 保持独立
```

### 10. 保留在工作区的文件

#### 工作区管理
```
go.work                     → 保留（工作区配置）
go.work.sum                 → 保留（工作区依赖锁定）
README.md                   → 保留（工作区说明）
.gitignore                  → 保留（工作区忽略规则）
```

#### 开发工具
```
tools/                      → 保留（开发工具）
tests/                      → 保留（集成测试）
update_imports.ps1          → 保留（迁移脚本）
```

#### 数据和临时文件
```
db/                         → 保留（数据库相关）
uploads/                    → 保留（上传文件）
bin/                        → 保留（编译输出）
misc/                       → 保留（杂项文件）
etc/                        → 保留（配置文件）
```

#### 包管理（废弃）
```
pkg/                        → 废弃（内容已迁移到laojun-shared）
go.mod                      → 废弃（各仓库有独立的go.mod）
```

## 迁移优先级

### 高优先级（立即迁移）
1. **共享组件库** - `repos/laojun-shared/` 已完成
2. **核心服务** - `cmd/admin-api/`, `cmd/config-center/`, 核心`internal/`模块
3. **市场服务** - `cmd/marketplace-api/`, 市场相关`internal/`模块

### 中优先级（第二阶段）
1. **管理后台** - `web/admin/`, 管理相关`cmd/`和`internal/`模块
2. **前端应用** - `web/marketplace/`, `api/`
3. **插件系统** - `internal/plugin/`

### 低优先级（最后阶段）
1. **部署配置** - `deploy/`, `configs/`, `Makefile`
2. **文档** - `docs/`, API文档
3. **清理工作** - 废弃`pkg/`目录，更新引用

## 依赖关系分析

### 核心依赖链
```
laojun-shared (基础)
    ↑
laojun-core (核心服务)
    ↑
laojun-marketplace, laojun-admin, laojun-web (业务服务)
    ↑
laojun-plugins (插件系统)
```

### 迁移顺序建议
1. 完善 `laojun-shared`
2. 迁移 `laojun-core`
3. 并行迁移 `laojun-marketplace`, `laojun-admin`
4. 迁移 `laojun-web`, `laojun-plugins`
5. 迁移 `laojun-deploy`, `laojun-docs`

## 注意事项

### 代码依赖更新
- 所有import路径需要更新
- go.mod文件需要重新配置
- 版本依赖需要管理

### 数据库迁移
- 数据库相关的cmd工具需要统一管理
- 迁移脚本需要在各仓库中保持一致

### 配置管理
- 配置文件需要在各仓库中独立管理
- 环境变量配置需要更新

### 测试覆盖
- 单元测试随代码迁移
- 集成测试保留在工作区
- CI/CD流水线需要重新配置

---

**文档版本**: v1.0  
**创建日期**: 2024年12月  
**最后更新**: 2024年12月  
**维护者**: Laojun 架构团队