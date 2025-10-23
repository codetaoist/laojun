# 数据库迁移说明

## 目录结构

```
db/
├── migrations/           # 标准数据库迁移文件
│   ├── 001_create_marketplace_tables.up.sql
│   ├── 001_create_marketplace_tables.down.sql
│   ├── 002_seed_marketplace_data.up.sql
│   ├── 002_seed_marketplace_data.down.sql
│   ├── 003_create_plugin_tables.up.sql
│   ├── 003_create_plugin_tables.down.sql
│   ├── 004_seed_plugin_data.up.sql
│   ├── 004_seed_plugin_data.down.sql
│   ├── 005_add_is_featured_column.up.sql
│   ├── 005_add_is_featured_column.down.sql
│   ├── 006_create_system_tables.up.sql      # 新增：系统表结构
│   ├── 006_create_system_tables.down.sql
│   ├── 007_create_indexes.up.sql            # 新增：性能优化索引
│   └── 007_create_indexes.down.sql
└── README.md            # 本文档
```

## 迁移文件说明

### 001 - 市场基础表
- **mp_users**: 市场用户表
- **mp_categories**: 插件分类表
- **mp_forum_categories**: 论坛分类表

### 002 - 市场基础数据
- 初始化基础分类数据
- 创建默认论坛分类

### 003 - 插件相关表
- **mp_plugins**: 插件信息表
- **mp_plugin_reviews**: 插件评价表

### 004 - 插件基础数据
- 插件示例数据
- 初始评价数据

### 005 - 功能增强
- 为插件表添加 `is_featured` 字段

### 006 - 系统表结构（新增）
包含完整的系统管理表结构：

#### 用户认证模块 (ua_*)
- **ua_admin**: 管理员用户表
- **ua_jwt_keys**: JWT密钥管理表
- **ua_user_sessions**: 用户会话表

#### 授权模块 (az_*)
- **az_roles**: 角色表
- **az_permissions**: 权限表
- **az_user_roles**: 用户角色关联表
- **az_role_permissions**: 角色权限关联表

#### 系统管理模块 (sm_*)
- **sm_menus**: 菜单表
- **sm_device_types**: 设备类型表
- **sm_modules**: 系统模块表

#### 系统功能模块 (sys_*)
- **sys_settings**: 系统设置表
- **sys_icons**: 系统图标表
- **sys_audit_logs**: 审计日志表

### 007 - 性能优化索引（新增）
为所有表创建必要的索引以提升查询性能：
- 主键和外键索引
- 常用查询字段索引
- 复合索引优化

## 表前缀说明

| 前缀 | 模块 | 说明 |
|------|------|------|
| mp_ | Marketplace | 市场相关表 |
| ua_ | User Authentication | 用户认证模块 |
| az_ | Authorization | 授权模块 |
| sm_ | System Management | 系统管理模块 |
| sys_ | System | 系统功能模块 |

## 数据库整合说明

### 整合前的问题
1. **目录冲突**: `sql/` 和 `db/` 目录功能重复
2. **文件重复**: 多个文件定义相同的表结构
3. **版本混乱**: 缺乏统一的迁移管理策略

### 整合后的改进
1. **统一目录**: 保留 `db/migrations/` 作为标准迁移目录
2. **历史保留**: 将原 `sql/` 目录重命名为 `sql-archive/` 保留历史记录
3. **标准化**: 所有迁移文件遵循 `{version}_{description}.{up|down}.sql` 命名规范
4. **完整性**: 补充了缺失的系统表结构和性能优化索引

### 迁移执行顺序
1. 001 - 创建市场基础表
2. 002 - 插入市场基础数据
3. 003 - 创建插件相关表
4. 004 - 插入插件基础数据
5. 005 - 添加功能增强字段
6. 006 - 创建完整系统表结构
7. 007 - 创建性能优化索引

## 使用说明

### 执行迁移
```bash
# 执行所有待执行的迁移
make migrate-up

# 回滚最后一次迁移
make migrate-down

# 查看迁移状态
make migrate-status
```

### 创建新迁移
```bash
# 创建新的迁移文件
make migrate-create name=your_migration_name
```

## 注意事项

1. **备份**: 执行迁移前请备份数据库
2. **测试**: 在开发环境充分测试后再应用到生产环境
3. **顺序**: 严格按照版本号顺序执行迁移
4. **回滚**: 每个 up 迁移都有对应的 down 迁移用于回滚
5. **索引**: 大表创建索引可能耗时较长，建议在低峰期执行