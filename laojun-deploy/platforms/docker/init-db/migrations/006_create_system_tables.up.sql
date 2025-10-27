-- Create comprehensive system tables for laojun platform
-- This migration includes user auth, authorization, system management, and extended features

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- 用户认证模块 (User Authentication)
-- =====================================================

-- 管理员用户表
CREATE TABLE IF NOT EXISTS ua_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar VARCHAR(255),
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    is_super_admin BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- JWT密钥管理表
CREATE TABLE IF NOT EXISTS ua_jwt_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(50) UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    algorithm VARCHAR(20) NOT NULL DEFAULT 'RS256',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户会话表
CREATE TABLE IF NOT EXISTS ua_user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device_info TEXT,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 授权模块 (Authorization)
-- =====================================================

-- 角色表
CREATE TABLE IF NOT EXISTS az_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 权限表
CREATE TABLE IF NOT EXISTS az_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS az_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES ua_admin(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, role_id)
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS az_role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES ua_admin(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ua_admin_username ON ua_admin(username);
CREATE INDEX IF NOT EXISTS idx_ua_admin_email ON ua_admin(email);
CREATE INDEX IF NOT EXISTS idx_ua_admin_is_active ON ua_admin(is_active);

CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_key_id ON ua_jwt_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_is_active ON ua_jwt_keys(is_active);

CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_user_id ON ua_user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_token_hash ON ua_user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_expires_at ON ua_user_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_az_roles_name ON az_roles(name);
CREATE INDEX IF NOT EXISTS idx_az_roles_is_active ON az_roles(is_active);

CREATE INDEX IF NOT EXISTS idx_az_permissions_code ON az_permissions(code);
CREATE INDEX IF NOT EXISTS idx_az_permissions_resource ON az_permissions(resource);

CREATE INDEX IF NOT EXISTS idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_role_id ON az_user_roles(role_id);

CREATE INDEX IF NOT EXISTS idx_az_role_permissions_role_id ON az_role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_az_role_permissions_permission_id ON az_role_permissions(permission_id);

-- 插入迁移记录
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('006_create_system_tables', FALSE) 
ON CONFLICT (version) DO NOTHING;