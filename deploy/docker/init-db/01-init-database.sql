-- PostgreSQL 数据库初始化脚本
-- 此脚本在容器首次启动时自动执行
-- 注意：POSTGRES_DB 环境变量已经创建了 laojun 数据库，我们需要连接到它
-- 基于 complete_deployment.sql 的优化版本

-- 连接到 laojun 数据库
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

-- 创建更新时间触发器函数（用于后续迁移）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 输出确认信息
\echo 'Database laojun initialization completed successfully with enhanced features';