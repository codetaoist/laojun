-- 数据库初始化脚本
-- 确保 laojun 数据库存在并正确配置

-- 连接到 laojun 数据库（在Docker环境中，数据库已经通过环境变量创建）
\c laojun

-- 创建必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建模式
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS core;
CREATE SCHEMA IF NOT EXISTS plugin;
CREATE SCHEMA IF NOT EXISTS audit;

-- 设置默认权限
GRANT USAGE ON SCHEMA public, auth, core, plugin, audit TO PUBLIC;
GRANT CREATE ON SCHEMA public TO PUBLIC;

-- 创建迁移跟踪表
CREATE TABLE IF NOT EXISTS public.schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    dirty BOOLEAN NOT NULL DEFAULT FALSE,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 插入初始迁移记录
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('000_init_database', FALSE) 
ON CONFLICT (version) DO NOTHING;

-- 输出确认信息
\echo 'Database laojun initialized successfully';