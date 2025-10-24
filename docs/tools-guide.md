# 工具使用指南

本文档介绍项目中各种工具的使用方法和最佳实践。

## 命令行工具 (cmd/)

### 菜单管理工具 (menu-manager)
统一的菜单数据操作工具，整合了所有菜单相关功能。

```bash
# 备份菜单数据
go run cmd/menu-manager/main.go -action=backup -file=menu_backup.sql

# 检查菜单数据完整性
go run cmd/menu-manager/main.go -action=check

# 清理重复菜单项
go run cmd/menu-manager/main.go -action=clean-duplicates -force

# 恢复菜单数据
go run cmd/menu-manager/main.go -action=restore -file=menu_backup.sql

# 添加唯一约束
go run cmd/menu-manager/main.go -action=add-constraint

# 应用菜单增强迁移
go run cmd/menu-manager/main.go -action=migrate-enhancements
```

### 插件市场管理工具 (marketplace-manager)
统一的插件市场操作工具。

```bash
# 运行插件市场迁移
go run cmd/marketplace-manager/main.go -action=migrate

# 填充演示数据
go run cmd/marketplace-manager/main.go -action=seed-demo

# 迁移插件评审字段
go run cmd/marketplace-manager/main.go -action=migrate-review-fields
```

### 数据库维护工具 (db-maintenance)
统一的数据库维护工具。

```bash
# 执行SQL文件
go run cmd/db-maintenance/main.go -action=execute-sql -sql-file=migration.sql

# 执行SQL查询
go run cmd/db-maintenance/main.go -action=execute-sql -query="SELECT * FROM users"

# 检查数据库架构
go run cmd/db-maintenance/main.go -action=check-schema

# 修复表名问题
go run cmd/db-maintenance/main.go -action=fix-table-names

# 添加审计字段
go run cmd/db-maintenance/main.go -action=add-audit-field

# 更新申诉表结构
go run cmd/db-maintenance/main.go -action=update-appeals
```

### 项目管理工具 (project-manager)
统一的项目管理工具。

```bash
# 整理根目录文件
go run cmd/project-manager/main.go -action=organize-files

# 生成完整迁移
go run cmd/project-manager/main.go -action=generate-migration

# 重置数据库
go run cmd/project-manager/main.go -action=reset-database
```

### 其他工具

#### 管理员相关
- `admin-api`: 启动管理员API服务
- `create-admin`: 创建管理员账户
- `setup-super-admin`: 设置超级管理员
- `verify-super-admin`: 验证超级管理员

#### 数据库相关
- `db-manager`: 数据库管理器
- `db-migrate`: 数据库迁移
- `db-complete-migrate`: 完整数据库迁移
- `migrate`: 通用迁移工具
- `run_migration`: 运行迁移
- `fix-migration`: 修复迁移

#### 配置相关
- `config-center`: 配置中心服务
- `debug-config`: 调试配置

#### 工具类
- `generate_hash`: 生成哈希值
- `marketplace-api`: 插件市场API服务

## 通用选项

所有工具都支持以下通用选项：

- `-force`: 强制执行操作，跳过确认
- `-dry-run`: 预览模式，显示将要执行的操作但不实际执行

## 最佳实践

1. **使用预览模式**: 在执行重要操作前，先使用 `-dry-run` 选项预览
2. **备份数据**: 在执行数据库操作前，先备份相关数据
3. **分步执行**: 对于复杂操作，建议分步执行并验证每一步的结果
4. **查看帮助**: 每个工具都支持不带参数运行来查看帮助信息

## 故障排除

如果遇到问题，请：

1. 检查工具的帮助信息
2. 查看相关的分析报告文档
3. 使用 `-dry-run` 模式验证操作
4. 检查日志文件获取详细错误信息