# 太上老君系统数据库部署文件

本目录包含用于快速部署太上老君系统数据库的SQL文件。

## 文件说明

### 1. complete_deployment.sql ⭐ **强烈推荐使用**
- **用途**: 完整的生产环境部署文件（已优化）
- **包含**: 
  - 所有表结构（管理系统、论坛、博客、插件市场）
  - 完整的RBAC权限系统
  - 丰富的测试数据和示例内容
  - 性能优化索引
  - 数据库视图和触发器
  - 系统配置数据
- **特点**: 
  - 2024年最新优化版本
  - 经过全面测试验证
  - 包含完整的业务数据
  - 性能优化，生产就绪
- **适用场景**: 
  - **生产环境部署**（推荐）
  - 开发环境搭建
  - 完整功能演示
  - 系统测试

### 2. final_deploy.sql
- **用途**: 基础部署文件
- **包含**: 核心表结构 + 基础权限系统 + 测试数据
- **适用场景**: 
  - 简化部署需求
  - 快速原型验证

### 3. simple_deploy.sql
- **用途**: 简化部署文件
- **包含**: 基础表结构 + 最小数据集
- **适用场景**: 快速测试和验证

### 4. quick_deploy.sql
- **用途**: 快速开发环境部署
- **包含**: 核心表结构 + 基础数据
- **适用场景**:
  - 开发环境快速搭建
  - 测试环境初始化
  - 最小化功能验证

## 部署方法

### 使用 Docker（推荐）

1. **⭐ 推荐部署方式**（使用complete_deployment.sql）:
```bash
# 复制文件到容器
docker cp complete_deployment.sql laojun-postgres:/tmp/

# 执行部署
docker exec laojun-postgres psql -U laojun -d laojun -f /tmp/complete_deployment.sql
```

2. **基础部署方式**（使用final_deploy.sql）:
```bash
# 复制文件到容器
docker cp final_deploy.sql laojun-postgres:/tmp/

# 执行部署
docker exec laojun-postgres psql -U laojun -d laojun -f /tmp/final_deploy.sql
```

3. **快速部署**:
```bash
# 复制文件到容器
docker cp quick_deploy.sql laojun-postgres:/tmp/

# 执行部署
docker exec laojun-postgres psql -U laojun -d laojun -f /tmp/quick_deploy.sql
```

### 使用 Docker PostgreSQL 容器部署

1. **确保 PostgreSQL 容器正在运行**
```bash
docker ps | grep postgres
```

2. **执行完整部署**
```bash
# 进入项目目录
cd d:\laojun

# 执行完整部署
docker exec -i postgres_container psql -U your_username -d your_database < db\migrations\final\complete_deployment.sql
```

3. **或执行快速部署**
```bash
docker exec -i postgres_container psql -U your_username -d your_database < db\migrations\final\quick_deploy.sql
```

### 直接使用 psql 部署

```bash
# 完整部署
psql -h localhost -U your_username -d your_database -f db\migrations\final\complete_deployment.sql

# 快速部署
psql -h localhost -U your_username -d your_database -f db\migrations\final\quick_deploy.sql
```

## 默认账户信息

部署完成后，系统将创建以下默认账户：

### final_deploy.sql 提供的账户：
- **超级管理员账户**: 
  - 用户名: `admin`
  - 密码: `admin123`
  - 邮箱: `admin@laojun.com`
  - 权限: 超级管理员（所有权限）

- **演示账户**: 
  - 用户名: `demo`
  - 密码: `demo123`
  - 邮箱: `demo@laojun.com`
  - 权限: 普通管理员（部分权限）

### 其他部署文件提供的账户：
- **管理员账户**: 
  - 用户名: `admin`
  - 密码: `admin123`
  - 邮箱: `admin@laojun.com`
  - 权限: 超级管理员（所有权限）

## 数据库结构概览

### 核心表
- `ua_admin` - 管理员用户表
- `az_roles` - 角色表
- `az_permissions` - 权限表
- `az_role_permissions` - 角色权限关联表
- `az_user_roles` - 用户角色关联表
- `sm_menus` - 系统菜单表

### 业务表（完整部署）
- `mp_users` - 市场用户表
- `mp_forum_categories` - 论坛分类表
- `mp_forum_posts` - 论坛帖子表
- `mp_forum_replies` - 论坛回复表
- `mp_blog_categories` - 博客分类表
- `mp_categories` - 插件分类表
- `mp_plugins` - 插件表

## 权限系统

系统采用基于角色的权限控制（RBAC）：

### 默认角色
1. **super_admin** - 超级管理员（所有权限）
2. **admin** - 管理员（基础管理权限）
3. **moderator** - 版主（内容管理权限）
4. **user** - 普通用户（基础查看权限）

### 权限类型
- `system.manage` - 系统管理
- `user.manage` - 用户管理
- `role.manage` - 角色管理
- `permission.manage` - 权限管理
- `menu.manage` - 菜单管理
- `plugin.manage` - 插件管理
- `forum.manage` - 论坛管理
- `blog.manage` - 博客管理

## 优化内容（complete_deployment.sql）

### 🚀 性能优化
- **索引优化**: 为所有关键查询字段添加了索引
- **外键索引**: 为所有外键关系添加了索引
- **复合索引**: 为常用查询组合添加了复合索引

### 📊 数据库视图
- `v_user_permissions`: 用户权限视图，简化权限查询
- `v_plugin_stats`: 插件统计视图，提供插件相关统计信息
- `v_forum_stats`: 论坛统计视图，提供论坛相关统计信息

### ⚡ 触发器
- 自动维护分类帖子数量
- 自动维护插件分类数量
- 数据一致性保证

### 🎯 丰富的测试数据
- 完整的权限和角色数据
- 示例用户和管理员账户
- 论坛、博客、插件市场的示例内容
- 系统配置数据

## 注意事项

1. **密码安全：** 部署后请立即修改默认管理员密码
2. **权限配置：** 根据实际需求调整角色权限配置
3. **数据备份：** 生产环境部署前请做好数据备份
4. **字段兼容性：** 如果遇到字段不存在的错误，请检查表结构是否与预期一致
5. **扩展依赖：** 确保PostgreSQL已安装uuid-ossp和pgcrypto扩展

## 故障排除

### 常见问题

1. **UUID 扩展未安装**
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

2. **权限不足**
确保数据库用户具有创建表和插入数据的权限

3. **外键约束错误**
确保按照文件中的顺序执行 SQL 语句

### 验证部署

部署完成后，可以执行以下查询验证：

```sql
-- 检查表是否创建成功
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

-- 检查管理员用户
SELECT username, email, is_super_admin FROM ua_admin;

-- 检查角色权限
SELECT r.name as role_name, p.name as permission_name 
FROM az_roles r
JOIN az_role_permissions rp ON r.id = rp.role_id
JOIN az_permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.name;
```

## 联系支持

如果在部署过程中遇到问题，请联系开发团队或查看项目文档。