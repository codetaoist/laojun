-- 太上老君系统 - Docker 完整数据库部署文件
-- 基于 complete_deployment.sql 的 Docker 优化版本
-- 包含所有表结构和丰富的测试数据
-- 生成时间: 2025-01-23
-- 版本: v2.0 Enhanced for Docker

-- 启用必要的扩展（在初始化脚本中已创建，这里确保存在）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 表结构定义
-- =============================================================================

-- 权限表
CREATE TABLE IF NOT EXISTS az_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 角色表
CREATE TABLE IF NOT EXISTS az_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS az_role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS az_user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- 管理员用户表
CREATE TABLE IF NOT EXISTS ua_admin (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar TEXT,
    phone VARCHAR(20),
    department VARCHAR(100),
    position VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    is_super_admin BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- JWT密钥表
CREATE TABLE IF NOT EXISTS ua_jwt_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id VARCHAR(50) NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    algorithm VARCHAR(20) DEFAULT 'RS256',
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户会话表
CREATE TABLE IF NOT EXISTS ua_user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    refresh_token VARCHAR(255),
    device_info TEXT,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 市场用户表（已在001_create_marketplace_tables.up.sql中创建，这里确保存在）
CREATE TABLE IF NOT EXISTS mp_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    website_url TEXT,
    github_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 系统配置表（已在009_comprehensive_optimization.up.sql中创建，这里确保存在）
CREATE TABLE IF NOT EXISTS sm_system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- 索引创建
-- =============================================================================

-- 权限相关索引
CREATE INDEX IF NOT EXISTS idx_az_permissions_code ON az_permissions(code);
CREATE INDEX IF NOT EXISTS idx_az_permissions_resource_action ON az_permissions(resource, action);

-- 角色相关索引
CREATE INDEX IF NOT EXISTS idx_az_roles_name ON az_roles(name);
CREATE INDEX IF NOT EXISTS idx_az_roles_is_system ON az_roles(is_system);

-- 用户角色索引
CREATE INDEX IF NOT EXISTS idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_role_id ON az_user_roles(role_id);

-- 管理员用户索引
CREATE INDEX IF NOT EXISTS idx_ua_admin_username ON ua_admin(username);
CREATE INDEX IF NOT EXISTS idx_ua_admin_email ON ua_admin(email);
CREATE INDEX IF NOT EXISTS idx_ua_admin_is_active ON ua_admin(is_active);

-- JWT密钥索引
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_key_id ON ua_jwt_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_is_active ON ua_jwt_keys(is_active);

-- 会话索引
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_user_id ON ua_user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_session_token ON ua_user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_expires_at ON ua_user_sessions(expires_at);

-- =============================================================================
-- 触发器创建
-- =============================================================================

-- 为管理员用户表添加更新时间触发器
DROP TRIGGER IF EXISTS update_ua_admin_updated_at ON ua_admin;
CREATE TRIGGER update_ua_admin_updated_at
    BEFORE UPDATE ON ua_admin
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为角色表添加更新时间触发器
DROP TRIGGER IF EXISTS update_az_roles_updated_at ON az_roles;
CREATE TRIGGER update_az_roles_updated_at
    BEFORE UPDATE ON az_roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为用户会话表添加更新时间触发器
DROP TRIGGER IF EXISTS update_ua_user_sessions_updated_at ON ua_user_sessions;
CREATE TRIGGER update_ua_user_sessions_updated_at
    BEFORE UPDATE ON ua_user_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 基础数据插入
-- =============================================================================

-- 插入基础权限数据
INSERT INTO az_permissions (name, code, description, resource, action) VALUES
('查看用户', 'user.view', '查看用户信息', 'user', 'view'),
('创建用户', 'user.create', '创建新用户', 'user', 'create'),
('编辑用户', 'user.edit', '编辑用户信息', 'user', 'edit'),
('删除用户', 'user.delete', '删除用户', 'user', 'delete'),
('查看角色', 'role.view', '查看角色信息', 'role', 'view'),
('创建角色', 'role.create', '创建新角色', 'role', 'create'),
('编辑角色', 'role.edit', '编辑角色信息', 'role', 'edit'),
('删除角色', 'role.delete', '删除角色', 'role', 'delete'),
('系统管理', 'system.admin', '系统管理权限', 'system', 'admin')
ON CONFLICT (code) DO NOTHING;

-- 插入基础角色数据
INSERT INTO az_roles (name, display_name, description, is_system) VALUES
('super_admin', '超级管理员', '拥有所有权限的超级管理员', TRUE),
('admin', '管理员', '系统管理员', TRUE),
('user_manager', '用户管理员', '负责用户管理的管理员', FALSE),
('content_manager', '内容管理员', '负责内容管理的管理员', FALSE)
ON CONFLICT (name) DO NOTHING;

-- 插入默认管理员用户
INSERT INTO ua_admin (username, email, password_hash, full_name, is_active, is_super_admin) VALUES
('admin', 'admin@laojun.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '系统管理员', TRUE, TRUE),
('demo', 'demo@laojun.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '演示用户', TRUE, FALSE)
ON CONFLICT (username) DO NOTHING;

-- 插入系统配置
INSERT INTO sm_system_configs (config_key, config_value, config_type, description, is_public) VALUES
('site_name', 'Laojun Platform', 'string', '网站名称', TRUE),
('site_description', '老君平台 - 插件市场与开发者社区', 'string', '网站描述', TRUE),
('max_upload_size', '10485760', 'integer', '最大上传文件大小（字节）', FALSE),
('enable_registration', 'true', 'boolean', '是否允许用户注册', TRUE),
('maintenance_mode', 'false', 'boolean', '维护模式', FALSE)
ON CONFLICT (config_key) DO NOTHING;

-- 插入迁移记录
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('999_complete_deployment', FALSE) 
ON CONFLICT (version) DO NOTHING;

-- 输出确认信息
\echo 'Complete deployment migration for Docker completed successfully';