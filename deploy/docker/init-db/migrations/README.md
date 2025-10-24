# Docker 数据库迁移文件

本目录包含用于 Docker 部署的数据库迁移文件，这些文件与主项目的 `db/migrations` 目录保持同步。

## 文件说明

### 核心迁移文件

- `000_init_database.up.sql` - 数据库初始化，创建扩展和模式
- `001_create_marketplace_tables.up.sql` - 创建市场相关表（用户、论坛等）
- `006_create_system_tables.up.sql` - 创建系统管理表（认证、授权等）
- `009_comprehensive_optimization.up.sql` - 综合优化（索引、触发器、视图等）
- `999_complete_deployment.up.sql` - 完整部署脚本（包含所有表结构和基础数据）

### 执行顺序

迁移文件按文件名的数字顺序执行：

1. `000_init_database.up.sql` - 初始化数据库环境
2. `001_create_marketplace_tables.up.sql` - 创建市场功能表
3. `006_create_system_tables.up.sql` - 创建系统管理表
4. `009_comprehensive_optimization.up.sql` - 应用性能优化
5. `999_complete_deployment.up.sql` - 完整部署（可选，包含测试数据）

## 使用方法

### 自动执行（推荐）

在 Docker 容器启动时，`02-run-migrations.sh` 脚本会自动按顺序执行所有 `.up.sql` 文件。

```bash
# 启动 Docker Compose
docker-compose up -d

# 查看迁移执行日志
docker-compose logs postgres
```

### 手动执行

如果需要手动执行特定的迁移文件：

```bash
# 进入 PostgreSQL 容器
docker-compose exec postgres psql -U laojun -d laojun

# 执行特定迁移文件
\i /docker-entrypoint-initdb.d/migrations/001_create_marketplace_tables.up.sql
```

## 与主项目同步

这些迁移文件与主项目的 `db/migrations` 目录保持同步：

- 主项目路径：`d:\laojun\db\migrations\`
- Docker 路径：`d:\laojun\deploy\docker\init-db\migrations\`

当主项目的迁移文件更新时，需要同步更新这里的文件。

## 迁移跟踪

所有迁移的执行状态都记录在 `public.schema_migrations` 表中：

```sql
-- 查看已执行的迁移
SELECT * FROM public.schema_migrations ORDER BY applied_at;

-- 检查特定迁移是否已执行
SELECT * FROM public.schema_migrations WHERE version = '001_create_marketplace_tables';
```

## 故障排除

### 迁移执行失败

1. 检查 PostgreSQL 容器日志：
   ```bash
   docker-compose logs postgres
   ```

2. 检查迁移文件语法：
   ```bash
   # 进入容器验证 SQL 语法
   docker-compose exec postgres psql -U laojun -d laojun -f /docker-entrypoint-initdb.d/migrations/文件名.up.sql
   ```

3. 手动清理和重新执行：
   ```sql
   -- 删除迁移记录（谨慎操作）
   DELETE FROM public.schema_migrations WHERE version = '迁移版本';
   
   -- 重新执行迁移
   \i /docker-entrypoint-initdb.d/migrations/文件名.up.sql
   ```

### 数据冲突

如果遇到数据冲突（如唯一约束违反），检查：

1. 是否重复执行了迁移
2. 是否有手动插入的测试数据与迁移数据冲突
3. 迁移文件中是否使用了 `ON CONFLICT` 子句

## 最佳实践

1. **备份数据**：在执行迁移前备份重要数据
2. **测试环境验证**：先在测试环境验证迁移文件
3. **版本控制**：确保迁移文件与代码版本同步
4. **监控执行**：关注迁移执行日志，及时发现问题

## 联系支持

如果遇到迁移相关问题，请：

1. 查看容器日志获取详细错误信息
2. 检查数据库连接和权限设置
3. 联系开发团队获取技术支持