-- =====================================================
-- 太上老君系统 - 完整数据库迁移脚本
-- 统一表名前缀，整合所有功能模块
-- =====================================================

-- 清理现有表（按依赖关系逆序删除）
DROP TABLE IF EXISTS pe_user_device_permissions CASCADE;
DROP TABLE IF EXISTS pe_permission_inheritance CASCADE;
DROP TABLE IF EXISTS pe_extended_permissions CASCADE;
DROP TABLE IF EXISTS ug_user_group_permissions CASCADE;
DROP TABLE IF EXISTS ug_user_group_members CASCADE;
DROP TABLE IF EXISTS ug_permission_templates CASCADE;
DROP TABLE IF EXISTS ug_user_groups CASCADE;
DROP TABLE IF EXISTS az_role_permissions CASCADE;
DROP TABLE IF EXISTS az_user_roles CASCADE;
DROP TABLE IF EXISTS az_permissions CASCADE;
DROP TABLE IF EXISTS az_roles CASCADE;
DROP TABLE IF EXISTS ua_user_sessions CASCADE;
DROP TABLE IF EXISTS ua_jwt_keys CASCADE;
DROP TABLE IF EXISTS ua_admin CASCADE;
DROP TABLE IF EXISTS sm_modules CASCADE;
DROP TABLE IF EXISTS sm_device_types CASCADE;
DROP TABLE IF EXISTS sm_menus CASCADE;
DROP TABLE IF EXISTS mp_plugin_reviews CASCADE;
DROP TABLE IF EXISTS mp_plugin_versions CASCADE;
DROP TABLE IF EXISTS mp_user_purchases CASCADE;
DROP TABLE IF EXISTS mp_user_favorites CASCADE;
DROP TABLE IF EXISTS mp_plugins CASCADE;
DROP TABLE IF EXISTS mp_categories CASCADE;
DROP TABLE IF EXISTS mp_developers CASCADE;
DROP TABLE IF EXISTS sys_audit_logs CASCADE;
DROP TABLE IF EXISTS sys_settings CASCADE;
DROP TABLE IF EXISTS sys_icons CASCADE;

-- =====================================================
-- 1. 用户认证模块 (User Authentication - ua_)
-- =====================================================

-- 管理员用户表
CREATE TABLE ua_admin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar VARCHAR(255),
    bio TEXT,
    is_active BOOLEAN DEFAULT true,
    is_email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- 用户会话表
CREATE TABLE ua_user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- JWT密钥管理表
CREATE TABLE ua_jwt_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(50) NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    algorithm VARCHAR(20) DEFAULT 'RS256',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 2. 授权模块 (Authorization - az_)
-- =====================================================

-- 角色表
CREATE TABLE az_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 权限表
CREATE TABLE az_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 模块表
CREATE TABLE az_modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES az_modules(id) ON DELETE CASCADE,
    icon VARCHAR(100),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户角色关联表
CREATE TABLE az_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- 角色权限关联表
CREATE TABLE az_role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- =====================================================
-- 3. 系统管理模块 (System Management - sm_)
-- =====================================================

-- 菜单表
CREATE TABLE sm_menus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    path VARCHAR(255),
    icon VARCHAR(100),
    component VARCHAR(255),
    parent_id UUID REFERENCES sm_menus(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT false,
    permissions JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 设备类型表
CREATE TABLE sm_device_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 4. 用户组模块 (User Groups - ug_)
-- =====================================================

-- 用户组表
CREATE TABLE ug_user_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户组成员表
CREATE TABLE ug_user_group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES ug_user_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, group_id)
);

-- 权限模板表
CREATE TABLE ug_permission_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户组权限表
CREATE TABLE ug_user_group_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES ug_user_groups(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    device_type_id UUID REFERENCES sm_device_types(id) ON DELETE CASCADE,
    granted BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, permission_id, device_type_id)
);

-- =====================================================
-- 5. 权限扩展模块 (Permission Extensions - pe_)
-- =====================================================

-- 扩展权限表，支持细粒度权限控制
CREATE TABLE pe_extended_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_type_id UUID REFERENCES sm_device_types(id) ON DELETE CASCADE,
    module_id UUID REFERENCES az_modules(id) ON DELETE CASCADE,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    element_type VARCHAR(255), -- button, menu, field, api
    element_code VARCHAR(255),
    conditions JSONB, -- 额外的权限条件
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 权限继承表，支持权限继承关系
CREATE TABLE pe_permission_inheritance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    child_permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    inheritance_type VARCHAR(20) DEFAULT 'inherit', -- inherit, exclude, override
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(parent_permission_id, child_permission_id)
);

-- 用户设备权限表，支持用户设备级别的权限控制
CREATE TABLE pe_user_device_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    device_type_id UUID REFERENCES sm_device_types(id) ON DELETE CASCADE,
    permissions JSONB,
    granted_by UUID REFERENCES ua_admin(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_id)
);

-- =====================================================
-- 6. 系统功能模块 (System - sys_)
-- =====================================================

-- 系统设置表
CREATE TABLE sys_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT,
    description TEXT,
    data_type VARCHAR(20) DEFAULT 'string',
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 图标库表
CREATE TABLE sys_icons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    svg_content TEXT NOT NULL,
    category VARCHAR(50),
    tags VARCHAR(200),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 审计日志表
CREATE TABLE sys_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES ua_admin(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 7. 市场模块 (Marketplace - mp_)
-- =====================================================

-- 插件开发者表
CREATE TABLE mp_developers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    company_name VARCHAR(200),
    website VARCHAR(500),
    description TEXT,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插件分类表
CREATE TABLE mp_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES mp_categories(id) ON DELETE SET NULL,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插件表
CREATE TABLE mp_plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    developer_id UUID NOT NULL REFERENCES mp_developers(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES mp_categories(id) ON DELETE RESTRICT,
    price DECIMAL(10,2) DEFAULT 0.00,
    is_free BOOLEAN DEFAULT true,
    status VARCHAR(20) DEFAULT 'draft', -- draft, pending, approved, rejected, suspended
    download_count INTEGER DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER DEFAULT 0,
    tags TEXT[], -- 标签数组
    requirements JSONB, -- 系统要求
    features JSONB, -- 功能特性
    screenshots TEXT[], -- 截图URL数组
    icon_url VARCHAR(500),
    banner_url VARCHAR(500),
    documentation_url VARCHAR(500),
    support_url VARCHAR(500),
    license VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插件版本表
CREATE TABLE mp_plugin_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    changelog TEXT,
    download_url VARCHAR(500) NOT NULL,
    file_size BIGINT,
    file_hash VARCHAR(255),
    is_stable BOOLEAN DEFAULT true,
    min_system_version VARCHAR(50),
    max_system_version VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, version)
);

-- 用户收藏表
CREATE TABLE mp_user_favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, plugin_id)
);

-- 用户购买表
CREATE TABLE mp_user_purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
    version_id UUID REFERENCES mp_plugin_versions(id),
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'pending',
    transaction_id VARCHAR(255),
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, plugin_id)
);

-- 插件评价表
CREATE TABLE mp_plugin_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(200),
    content TEXT,
    is_verified_purchase BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, user_id)
);

-- =====================================================
-- 创建索引
-- =====================================================

-- 用户认证模块索引
CREATE INDEX idx_ua_admin_username ON ua_admin(username);
CREATE INDEX idx_ua_admin_email ON ua_admin(email);
CREATE INDEX idx_ua_admin_is_active ON ua_admin(is_active);
CREATE INDEX idx_ua_user_sessions_user_id ON ua_user_sessions(user_id);
CREATE INDEX idx_ua_user_sessions_token_hash ON ua_user_sessions(token_hash);
CREATE INDEX idx_ua_user_sessions_expires_at ON ua_user_sessions(expires_at);
CREATE INDEX idx_ua_jwt_keys_key_id ON ua_jwt_keys(key_id);
CREATE INDEX idx_ua_jwt_keys_is_active ON ua_jwt_keys(is_active);

-- 授权模块索引
CREATE INDEX idx_az_roles_name ON az_roles(name);
CREATE INDEX idx_az_permissions_code ON az_permissions(code);
CREATE INDEX idx_az_permissions_resource_action ON az_permissions(resource, action);
CREATE INDEX idx_az_modules_code ON az_modules(code);
CREATE INDEX idx_az_modules_parent_id ON az_modules(parent_id);
CREATE INDEX idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX idx_az_user_roles_role_id ON az_user_roles(role_id);
CREATE INDEX idx_az_role_permissions_role_id ON az_role_permissions(role_id);
CREATE INDEX idx_az_role_permissions_permission_id ON az_role_permissions(permission_id);

-- 系统管理模块索引
CREATE INDEX idx_sm_menus_parent_id ON sm_menus(parent_id);
CREATE INDEX idx_sm_menus_sort_order ON sm_menus(sort_order);
CREATE INDEX idx_sm_device_types_code ON sm_device_types(code);
CREATE INDEX idx_sm_device_types_is_active ON sm_device_types(is_active);

-- 用户组模块索引
CREATE INDEX idx_ug_user_groups_name ON ug_user_groups(name);
CREATE INDEX idx_ug_user_group_members_user_id ON ug_user_group_members(user_id);
CREATE INDEX idx_ug_user_group_members_group_id ON ug_user_group_members(group_id);
CREATE INDEX idx_ug_permission_templates_name ON ug_permission_templates(name);
CREATE INDEX idx_ug_user_group_permissions_group_id ON ug_user_group_permissions(group_id);
CREATE INDEX idx_ug_user_group_permissions_permission_id ON ug_user_group_permissions(permission_id);

-- 权限扩展模块索引
CREATE INDEX idx_pe_extended_permissions_device_type_id ON pe_extended_permissions(device_type_id);
CREATE INDEX idx_pe_extended_permissions_module_id ON pe_extended_permissions(module_id);
CREATE INDEX idx_pe_extended_permissions_resource_action ON pe_extended_permissions(resource, action);
CREATE INDEX idx_pe_permission_inheritance_parent ON pe_permission_inheritance(parent_permission_id);
CREATE INDEX idx_pe_permission_inheritance_child ON pe_permission_inheritance(child_permission_id);
CREATE INDEX idx_pe_user_device_permissions_user_id ON pe_user_device_permissions(user_id);
CREATE INDEX idx_pe_user_device_permissions_device_id ON pe_user_device_permissions(device_id);
CREATE INDEX idx_pe_user_device_permissions_device_type_id ON pe_user_device_permissions(device_type_id);

-- 系统功能模块索引
CREATE INDEX idx_sys_settings_key ON sys_settings(key);
CREATE INDEX idx_sys_icons_code ON sys_icons(code);
CREATE INDEX idx_sys_icons_category ON sys_icons(category);
CREATE INDEX idx_sys_audit_logs_user_id ON sys_audit_logs(user_id);
CREATE INDEX idx_sys_audit_logs_resource ON sys_audit_logs(resource);
CREATE INDEX idx_sys_audit_logs_created_at ON sys_audit_logs(created_at);

-- 市场模块索引
CREATE INDEX idx_mp_developers_user_id ON mp_developers(user_id);
CREATE INDEX idx_mp_categories_parent_id ON mp_categories(parent_id);
CREATE INDEX idx_mp_categories_is_active ON mp_categories(is_active);
CREATE INDEX idx_mp_plugins_slug ON mp_plugins(slug);
CREATE INDEX idx_mp_plugins_developer_id ON mp_plugins(developer_id);
CREATE INDEX idx_mp_plugins_category_id ON mp_plugins(category_id);
CREATE INDEX idx_mp_plugins_status ON mp_plugins(status);
CREATE INDEX idx_mp_plugin_versions_plugin_id ON mp_plugin_versions(plugin_id);
CREATE INDEX idx_mp_user_favorites_user_id ON mp_user_favorites(user_id);
CREATE INDEX idx_mp_user_purchases_user_id ON mp_user_purchases(user_id);
CREATE INDEX idx_mp_plugin_reviews_plugin_id ON mp_plugin_reviews(plugin_id);
CREATE INDEX idx_mp_plugin_reviews_user_id ON mp_plugin_reviews(user_id);