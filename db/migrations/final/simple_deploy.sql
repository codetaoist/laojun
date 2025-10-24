-- 简化部署文件
-- 删除现有表
DROP TABLE IF EXISTS az_user_roles CASCADE;
DROP TABLE IF EXISTS az_role_permissions CASCADE;
DROP TABLE IF EXISTS az_permissions CASCADE;
DROP TABLE IF EXISTS az_roles CASCADE;
DROP TABLE IF EXISTS ua_admin CASCADE;
DROP TABLE IF EXISTS sm_menus CASCADE;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 权限表
CREATE TABLE az_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    resource VARCHAR(100),
    action VARCHAR(50),
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 角色表
CREATE TABLE az_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 管理员表
CREATE TABLE ua_admin (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 菜单表
CREATE TABLE sm_menus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(100) NOT NULL,
    path VARCHAR(200),
    icon VARCHAR(50),
    component VARCHAR(100),
    parent_id UUID REFERENCES sm_menus(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT false,
    device_types JSONB DEFAULT '["pc","web"]',
    permissions JSONB DEFAULT '[]',
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
    user_id UUID NOT NULL REFERENCES ua_admin(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES az_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

-- 插入基础权限
INSERT INTO az_permissions (name, display_name, description, resource, action, is_system) VALUES ('system.admin', 'System Admin', 'Full system administration access', 'system', 'admin', true);
INSERT INTO az_permissions (name, display_name, description, resource, action, is_system) VALUES ('menu.manage', 'Menu Management', 'Manage system menus', 'menu', 'manage', true);
INSERT INTO az_permissions (name, display_name, description, resource, action, is_system) VALUES ('user.manage', 'User Management', 'Manage system users', 'user', 'manage', true);

-- 插入基础角色
INSERT INTO az_roles (name, display_name, description, is_system) VALUES ('super_admin', 'Super Admin', 'System Super Administrator', true);
INSERT INTO az_roles (name, display_name, description, is_system) VALUES ('admin', 'Admin', 'System Administrator', true);

-- 插入管理员用户
INSERT INTO ua_admin (username, email, password_hash, full_name, is_active) VALUES ('admin', 'admin@laojun.com', crypt('admin123', gen_salt('bf')), 'System Administrator', true);

-- 插入基础菜单
INSERT INTO sm_menus (title, path, icon, component, parent_id, sort_order, device_types) VALUES ('Dashboard', '/dashboard', 'dashboard', 'Dashboard', NULL, 1, '["pc","web"]');
INSERT INTO sm_menus (title, path, icon, component, parent_id, sort_order, device_types) VALUES ('System Management', '/system', 'setting', NULL, NULL, 2, '["pc","web"]');

-- 创建基础索引
CREATE INDEX idx_az_permissions_name ON az_permissions(name);
CREATE INDEX idx_az_roles_name ON az_roles(name);
CREATE INDEX idx_ua_admin_username ON ua_admin(username);
CREATE INDEX idx_ua_admin_email ON ua_admin(email);
CREATE INDEX idx_sm_menus_parent_id ON sm_menus(parent_id);
CREATE INDEX idx_sm_menus_sort_order ON sm_menus(sort_order);

SELECT 'Simple deployment completed successfully!' as status;