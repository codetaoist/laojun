# 数据库迁移系统

本项目使用 `golang-migrate` 工具进行数据库迁移管理，提供版本化的数据库架构管理。

## 概述

迁移系统的主要特点：
- **版本化管理**：每个迁移都有唯一的版本号
- **双向迁移**：支持上行（up）和下行（down）迁移
- **原子性操作**：每个迁移在事务中执行
- **状态跟踪**：自动跟踪迁移状态和版本
- **脏状态检测**：检测并防止不一致的迁移状态

## 目录结构

```
db/
└── migrations/
    ├── 001_create_user_auth_tables.up.sql
    ├── 001_create_user_auth_tables.down.sql
    ├── 002_create_authorization_tables.up.sql
    ├── 002_create_authorization_tables.down.sql
    ├── ...
    └── 009_assign_role_permissions.down.sql
```

## 迁移文件命名规范

迁移文件遵循以下命名规范：
```
{version}_{description}.{direction}.sql
```

- `version`: 3位数字版本号（如 001, 002, 003）
- `description`: 迁移描述（使用下划线分隔）
- `direction`: `up` 或 `down`
- 扩展名: `.sql`

示例：
- `001_create_user_auth_tables.up.sql`
- `001_create_user_auth_tables.down.sql`

## 使用方法

### 1. 执行迁移

```bash
# 执行所有待执行的迁移
go run cmd/db-migrate/main.go -action=up -db="postgres://user:pass@localhost/dbname?sslmode=disable"

# 使用环境变量
export DATABASE_URL="postgres://user:pass@localhost/dbname?sslmode=disable"
go run cmd/db-migrate/main.go -action=up
```

### 2. 查看当前版本

```bash
go run cmd/db-migrate/main.go -action=version -db="your_database_url"
```

### 3. 回滚迁移

```bash
# 回滚到指定版本
go run cmd/db-migrate/main.go -action=rollback -version=3 -db="your_database_url"

# 回滚所有迁移
go run cmd/db-migrate/main.go -action=down -db="your_database_url"
```

### 4. 重置数据库

```bash
# 重置数据库（交互式确认）
go run cmd/db-migrate/main.go -action=reset -db="your_database_url"
```

## 迁移版本说明

### 001 - 用户认证表
- `ua_users`: 用户基础信息表
- `lj_user_sessions`: 用户会话管理表
- `lj_jwt_keys`: JWT 密钥管理表

### 002 - 权限管理表
- `az_roles`: 角色表
- `az_permissions`: 权限表
- `az_user_roles`: 用户角色关联表
- `az_role_permissions`: 角色权限关联表

### 003 - 系统管理表
- `lj_menus`: 菜单表
- `lj_device_types`: 设备类型表
- `lj_modules`: 模块表

### 004 - 用户组和扩展权限
- `lj_user_groups`: 用户组表
- `lj_user_group_members`: 用户组成员表
- `lj_permission_templates`: 权限模板表
- `lj_extended_permissions`: 扩展权限表

### 005 - 权限继承和设备权限
- `lj_permission_inheritance`: 权限继承表
- `lj_user_group_permissions`: 用户组权限表
- `lj_user_device_permissions`: 用户设备权限表

### 006 - 插件市场基础表
- `developers`: 开发者表
- `categories`: 插件分类表
- `plugins`: 插件表

### 007 - 插件版本和交互
- `plugin_versions`: 插件版本表
- `reviews`: 插件评论表
- `purchases`: 插件购买记录表
- `user_favorites`: 用户收藏表

### 008 - 初始数据
- 默认角色和权限
- 默认菜单
- 默认分类

### 009 - 角色权限分配
- 为默认角色分配相应权限

## 开发指南

### 创建新的迁移

1. 确定版本号（下一个递增的3位数字）
2. 创建上行迁移文件：`{version}_{description}.up.sql`
3. 创建下行迁移文件：`{version}_{description}.down.sql`
4. 在上行文件中编写创建/修改语句
5. 在下行文件中编写相应的回滚语句

### 最佳实践

1. **原子性**：每个迁移应该是原子的，要么全部成功，要么全部失败
2. **可逆性**：每个上行迁移都应该有对应的下行迁移
3. **幂等性**：迁移应该是幂等的，多次执行结果相同
4. **测试**：在开发环境中充分测试迁移
5. **备份**：在生产环境执行迁移前备份数据库

### 注意事项

1. **不要修改已执行的迁移文件**：一旦迁移在生产环境执行，就不应该修改
2. **谨慎使用 DROP 操作**：删除表或列可能导致数据丢失
3. **索引管理**：大表上创建索引可能需要很长时间
4. **外键约束**：注意外键约束的创建和删除顺序

## 故障排除

### 脏状态

如果数据库处于脏状态（dirty state），说明某个迁移执行失败：

1. 检查数据库日志确定失败原因
2. 手动修复数据库状态
3. 使用 `golang-migrate` 工具清除脏状态标记

### 版本不一致

如果本地迁移文件与数据库版本不一致：

1. 检查是否有未提交的迁移文件
2. 确认当前分支的迁移文件是否完整
3. 必要时回滚到一致的版本

## 相关文件

- `internal/database/migrator.go`: 迁移器实现
- `cmd/db-migrate/main.go`: 迁移管理命令
- `db/migrations/`: 迁移文件目录