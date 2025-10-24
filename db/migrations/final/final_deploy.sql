-- 太上老君系统 - 最终部署文件
-- 包含完整的表结构和测试数据
-- 使用方法: docker exec laojun-postgres psql -U laojun -d laojun -f /tmp/final_deploy.sql

-- 删除现有表（如果存在）
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
INSERT INTO az_permissions (name, display_name, description, resource, action, is_system) VALUES 
('system.admin', 'System Admin', 'Full system administration access', 'system', 'admin', true),
('menu.manage', 'Menu Management', 'Manage system menus', 'menu', 'manage', true),
('user.manage', 'User Management', 'Manage system users', 'user', 'manage', true),
('role.manage', 'Role Management', 'Manage system roles', 'role', 'manage', true),
('permission.manage', 'Permission Management', 'Manage system permissions', 'permission', 'manage', true);

-- 插入基础角色
INSERT INTO az_roles (name, display_name, description, is_system) VALUES 
('super_admin', 'Super Admin', 'System Super Administrator', true),
('admin', 'Admin', 'System Administrator', true),
('user', 'User', 'Regular User', true);

-- 插入管理员用户
INSERT INTO ua_admin (username, email, password_hash, full_name, is_active) VALUES 
('admin', 'admin@laojun.com', crypt('admin123', gen_salt('bf')), 'System Administrator', true),
('demo', 'demo@laojun.com', crypt('demo123', gen_salt('bf')), 'Demo User', true);

-- 插入基础菜单
INSERT INTO sm_menus (title, path, icon, component, parent_id, sort_order, device_types) VALUES 
('Dashboard', '/dashboard', 'dashboard', 'Dashboard', NULL, 1, '["pc","web"]'),
('System Management', '/system', 'setting', NULL, NULL, 2, '["pc","web"]'),
('User Management', '/system/users', 'user', 'UserManagement', (SELECT id FROM sm_menus WHERE path = '/system'), 1, '["pc","web"]'),
('Role Management', '/system/roles', 'team', 'RoleManagement', (SELECT id FROM sm_menus WHERE path = '/system'), 2, '["pc","web"]'),
('Menu Management', '/system/menus', 'menu', 'MenuManagement', (SELECT id FROM sm_menus WHERE path = '/system'), 3, '["pc","web"]'),
('Permission Management', '/system/permissions', 'lock', 'PermissionManagement', (SELECT id FROM sm_menus WHERE path = '/system'), 4, '["pc","web"]');

-- 插入角色权限关联（超级管理员拥有所有权限）
INSERT INTO az_role_permissions (role_id, permission_id) 
SELECT r.id, p.id FROM az_roles r, az_permissions p WHERE r.name = 'super_admin';

-- 插入角色权限关联（普通管理员拥有部分权限）
INSERT INTO az_role_permissions (role_id, permission_id) 
SELECT r.id, p.id FROM az_roles r, az_permissions p 
WHERE r.name = 'admin' AND p.name IN ('menu.manage', 'user.manage');

-- 插入用户角色关联
INSERT INTO az_user_roles (user_id, role_id) 
SELECT u.id, r.id FROM ua_admin u, az_roles r 
WHERE u.username = 'admin' AND r.name = 'super_admin';

INSERT INTO az_user_roles (user_id, role_id) 
SELECT u.id, r.id FROM ua_admin u, az_roles r 
WHERE u.username = 'demo' AND r.name = 'admin';

-- 创建基础索引
CREATE INDEX idx_az_permissions_name ON az_permissions(name);
CREATE INDEX idx_az_permissions_resource ON az_permissions(resource);
CREATE INDEX idx_az_roles_name ON az_roles(name);
CREATE INDEX idx_az_role_permissions_role_id ON az_role_permissions(role_id);
CREATE INDEX idx_az_role_permissions_permission_id ON az_role_permissions(permission_id);
CREATE INDEX idx_ua_admin_username ON ua_admin(username);
CREATE INDEX idx_ua_admin_email ON ua_admin(email);
CREATE INDEX idx_ua_admin_is_active ON ua_admin(is_active);
CREATE INDEX idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX idx_az_user_roles_role_id ON az_user_roles(role_id);
CREATE INDEX idx_sm_menus_parent_id ON sm_menus(parent_id);
CREATE INDEX idx_sm_menus_sort_order ON sm_menus(sort_order);
CREATE INDEX idx_sm_menus_path ON sm_menus(path);

-- 验证部署结果
SELECT 
    'Tables created: ' || COUNT(*) as table_count
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('az_permissions', 'az_roles', 'ua_admin', 'sm_menus', 'az_role_permissions', 'az_user_roles');

SELECT 
    'Admin users: ' || COUNT(*) as admin_count
FROM ua_admin;

SELECT 
    'Permissions: ' || COUNT(*) as permission_count
FROM az_permissions;

SELECT 
    'Roles: ' || COUNT(*) as role_count
FROM az_roles;

SELECT 
    'Menus: ' || COUNT(*) as menu_count
FROM sm_menus;

SELECT 'Final deployment completed successfully!' as status;

-- 显示默认登录信息
SELECT 
    'Default Login Info:' as info,
    'Username: admin, Password: admin123' as admin_account,
    'Username: demo, Password: demo123' as demo_account;