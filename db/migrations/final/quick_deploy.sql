-- 快速部署文件 - 包含核心表和基础数据
-- 适用于快速搭建开发环境

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 权限表
CREATE TABLE az_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 角色表
CREATE TABLE az_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 角色权限关联表
CREATE TABLE az_role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES az_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- 用户角色关联表
CREATE TABLE az_user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- 管理员用户表
CREATE TABLE ua_admin (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_super_admin BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 菜单表
CREATE TABLE sm_menus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(100) NOT NULL,
    path TEXT,
    icon VARCHAR(100),
    component TEXT,
    parent_id UUID REFERENCES sm_menus(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    device_types TEXT,
    permissions TEXT,
    custom_icon TEXT,
    description TEXT,
    keywords TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插入基础权限
INSERT INTO az_permissions (id, name, code, description, resource, action) VALUES
('11111111-1111-1111-1111-111111111111', '系统管理', 'system.manage', '系统管理权限', 'system', 'manage'),
('22222222-2222-2222-2222-222222222222', '用户管理', 'user.manage', '用户管理权限', 'user', 'manage'),
('33333333-3333-3333-3333-333333333333', '菜单管理', 'menu.manage', '菜单管理权限', 'menu', 'manage');

-- 插入基础角色
INSERT INTO az_roles (id, name, display_name, description, is_system) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin', 'Super Admin', 'System Super Administrator', true),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'admin', 'Admin', 'System Administrator', true);

-- 插入角色权限关联
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '22222222-2222-2222-2222-222222222222'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '33333333-3333-3333-3333-333333333333'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333');

-- 插入管理员用户 (密码: admin123)
INSERT INTO ua_admin (id, username, email, password_hash, is_active, is_super_admin) VALUES
('12345678-1234-1234-1234-123456789012', 'admin', 'admin@laojun.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', true, true);

-- 插入用户角色关联
INSERT INTO az_user_roles (user_id, role_id) VALUES
('12345678-1234-1234-1234-123456789012', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');

-- 插入基础菜单
INSERT INTO sm_menus (id, title, path, icon, component, parent_id, sort_order, device_types) VALUES
('m1111111-1111-1111-1111-111111111111', 'Dashboard', '/dashboard', 'dashboard', 'Dashboard', NULL, 1, '["pc","web"]'),
('m2222222-2222-2222-2222-222222222222', 'System Management', '/system', 'setting', NULL, NULL, 2, '["pc","web"]'),
('m2111111-1111-1111-1111-111111111111', 'Menu Management', '/system/menus', 'menu', 'MenuManagement', 'm2222222-2222-2222-2222-222222222222', 1, '["pc","web"]'),
('m2222222-1111-1111-1111-111111111111', 'User Management', '/system/users', 'user', 'UserManagement', 'm2222222-2222-2222-2222-222222222222', 2, '["pc","web"]');

-- 创建基础索引
CREATE INDEX idx_az_role_permissions_role_id ON az_role_permissions(role_id);
CREATE INDEX idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX idx_ua_admin_username ON ua_admin(username);
CREATE INDEX idx_sm_menus_parent_id ON sm_menus(parent_id);

SELECT 'Quick deployment completed!' as status;