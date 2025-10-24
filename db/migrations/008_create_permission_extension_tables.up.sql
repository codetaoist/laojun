-- =====================================================
-- 权限扩展模块表 (Permission Extensions)
-- 添加缺失的权限扩展相关表
-- =====================================================

-- 用户组表
CREATE TABLE IF NOT EXISTS ug_user_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户组成员表
CREATE TABLE IF NOT EXISTS ug_user_group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES ug_user_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, group_id)
);

-- 权限模板表
CREATE TABLE IF NOT EXISTS ug_permission_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户组权限表
CREATE TABLE IF NOT EXISTS ug_user_group_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES ug_user_groups(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    device_type_id UUID REFERENCES sm_device_types(id) ON DELETE CASCADE,
    granted BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, permission_id, device_type_id)
);

-- 扩展权限表，支持细粒度权限控制
CREATE TABLE IF NOT EXISTS pe_extended_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_type_id UUID REFERENCES sm_device_types(id) ON DELETE CASCADE,
    module_id UUID REFERENCES sm_modules(id) ON DELETE CASCADE,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    element_type VARCHAR(255), -- button, menu, field, api
    element_code VARCHAR(255),
    conditions JSONB, -- 额外的权限条件
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 权限继承表，支持权限继承关系
CREATE TABLE IF NOT EXISTS pe_permission_inheritance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    child_permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    inheritance_type VARCHAR(20) DEFAULT 'inherit', -- inherit, exclude, override
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(parent_permission_id, child_permission_id)
);

-- 用户设备权限表，支持用户设备级别的权限控制
CREATE TABLE IF NOT EXISTS pe_user_device_permissions (
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
    UNIQUE(user_id, device_id, device_type_id)
);

-- 创建索引
-- 用户组模块索引
CREATE INDEX IF NOT EXISTS idx_ug_user_groups_name ON ug_user_groups(name);
CREATE INDEX IF NOT EXISTS idx_ug_user_group_members_user_id ON ug_user_group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_ug_user_group_members_group_id ON ug_user_group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_ug_permission_templates_name ON ug_permission_templates(name);
CREATE INDEX IF NOT EXISTS idx_ug_user_group_permissions_group_id ON ug_user_group_permissions(group_id);
CREATE INDEX IF NOT EXISTS idx_ug_user_group_permissions_permission_id ON ug_user_group_permissions(permission_id);

-- 权限扩展模块索引
CREATE INDEX IF NOT EXISTS idx_pe_extended_permissions_device_type_id ON pe_extended_permissions(device_type_id);
CREATE INDEX IF NOT EXISTS idx_pe_extended_permissions_module_id ON pe_extended_permissions(module_id);
CREATE INDEX IF NOT EXISTS idx_pe_extended_permissions_resource_action ON pe_extended_permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_pe_permission_inheritance_parent ON pe_permission_inheritance(parent_permission_id);
CREATE INDEX IF NOT EXISTS idx_pe_permission_inheritance_child ON pe_permission_inheritance(child_permission_id);
CREATE INDEX IF NOT EXISTS idx_pe_user_device_permissions_user_id ON pe_user_device_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_pe_user_device_permissions_device_id ON pe_user_device_permissions(device_id);
CREATE INDEX IF NOT EXISTS idx_pe_user_device_permissions_device_type_id ON pe_user_device_permissions(device_type_id);