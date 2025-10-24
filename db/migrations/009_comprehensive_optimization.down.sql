-- 回滚综合优化迁移
-- 删除所有在 009_comprehensive_optimization.up.sql 中创建的对象

-- 删除触发器
DROP TRIGGER IF EXISTS update_ua_admin_updated_at ON ua_admin;
DROP TRIGGER IF EXISTS update_sm_system_configs_updated_at ON sm_system_configs;
DROP TRIGGER IF EXISTS update_mp_blog_categories_updated_at ON mp_blog_categories;
DROP TRIGGER IF EXISTS update_mp_forum_categories_updated_at ON mp_forum_categories;
DROP TRIGGER IF EXISTS update_mp_users_updated_at ON mp_users;

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除视图
DROP VIEW IF EXISTS v_forum_stats;
DROP VIEW IF EXISTS v_plugin_stats;
DROP VIEW IF EXISTS v_user_permissions;

-- 删除新增的高级复合索引
DROP INDEX IF EXISTS idx_az_user_roles_expires;
DROP INDEX IF EXISTS idx_mp_plugins_featured_rating;
DROP INDEX IF EXISTS idx_mp_forum_posts_category_likes;
DROP INDEX IF EXISTS idx_mp_users_email_verified_active;

-- 删除新表的索引
DROP INDEX IF EXISTS idx_sm_system_configs_public;
DROP INDEX IF EXISTS idx_sm_system_configs_type;
DROP INDEX IF EXISTS idx_sm_system_configs_key;
DROP INDEX IF EXISTS idx_mp_blog_categories_name;
DROP INDEX IF EXISTS idx_mp_blog_categories_sort_order;

-- 删除新增的表
DROP TABLE IF EXISTS sm_system_configs;
DROP TABLE IF EXISTS mp_blog_categories;

-- 删除扩展 (谨慎删除，可能被其他功能使用)
-- DROP EXTENSION IF EXISTS "pg_trgm";
-- DROP EXTENSION IF EXISTS "pg_stat_statements";

-- 删除迁移记录
DELETE FROM public.schema_migrations WHERE version = '009_comprehensive_optimization';