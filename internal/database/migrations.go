package database

const migrationSQL = `
CREATE TABLE IF NOT EXISTS lj_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    bio TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_lj_users_username ON lj_users(username);
CREATE INDEX IF NOT EXISTS idx_lj_users_email ON lj_users(email);
CREATE INDEX IF NOT EXISTS idx_lj_users_is_active ON lj_users(is_active);

-- 角色表
CREATE TABLE IF NOT EXISTS lj_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_roles_name ON lj_roles(name);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS lj_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES lj_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_lj_user_roles_user_id ON lj_user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_roles_role_id ON lj_user_roles(role_id);

-- 权限表
CREATE TABLE IF NOT EXISTS lj_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_permissions_code ON lj_permissions(code);
CREATE INDEX IF NOT EXISTS idx_lj_permissions_resource ON lj_permissions(resource);
-- 新增列：系统权限标记
ALTER TABLE lj_permissions ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_lj_permissions_is_system ON lj_permissions(is_system);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS lj_role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES lj_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES lj_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_lj_role_permissions_role_id ON lj_role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_lj_role_permissions_permission_id ON lj_role_permissions(permission_id);

-- 菜单表
CREATE TABLE IF NOT EXISTS lj_menus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    path VARCHAR(255),
    icon VARCHAR(100),
    component VARCHAR(255),
    parent_id UUID REFERENCES lj_menus(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    is_hidden BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(title, path, component)
);

CREATE INDEX IF NOT EXISTS idx_lj_menus_parent_id ON lj_menus(parent_id);
CREATE INDEX IF NOT EXISTS idx_lj_menus_sort_order ON lj_menus(sort_order);

-- 用户会话表
CREATE TABLE IF NOT EXISTS lj_user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_id ON lj_user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_token_hash ON lj_user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_expires_at ON lj_user_sessions(expires_at);

-- JWT密钥管理表
CREATE TABLE IF NOT EXISTS lj_jwt_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_lj_jwt_keys_is_active ON lj_jwt_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_lj_jwt_keys_expires_at ON lj_jwt_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_lj_jwt_keys_key_hash ON lj_jwt_keys(key_hash);

-- 设备类型表
CREATE TABLE IF NOT EXISTS lj_device_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_device_types_code ON lj_device_types(code);
CREATE INDEX IF NOT EXISTS idx_lj_device_types_is_active ON lj_device_types(is_active);

-- 模块表
CREATE TABLE IF NOT EXISTS lj_modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES lj_modules(id) ON DELETE CASCADE,
    icon VARCHAR(100),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_modules_code ON lj_modules(code);
CREATE INDEX IF NOT EXISTS idx_lj_modules_parent_id ON lj_modules(parent_id);
CREATE INDEX IF NOT EXISTS idx_lj_modules_is_active ON lj_modules(is_active);

-- 用户组表
CREATE TABLE IF NOT EXISTS lj_user_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_user_groups_name ON lj_user_groups(name);
CREATE INDEX IF NOT EXISTS idx_lj_user_groups_is_active ON lj_user_groups(is_active);

-- 用户组成员表
CREATE TABLE IF NOT EXISTS lj_user_group_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES lj_user_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_user_id ON lj_user_group_members(user_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_group_id ON lj_user_group_members(group_id);

-- 权限模板表
CREATE TABLE IF NOT EXISTS lj_permission_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_permission_templates_name ON lj_permission_templates(name);
CREATE INDEX IF NOT EXISTS idx_lj_permission_templates_is_system ON lj_permission_templates(is_system);

-- 扩展权限表，支持细粒度权限控制
CREATE TABLE IF NOT EXISTS lj_extended_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_type_id UUID REFERENCES lj_device_types(id) ON DELETE CASCADE,
    module_id UUID REFERENCES lj_modules(id) ON DELETE CASCADE,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    element_type VARCHAR(50), -- button, menu, field, api
    element_code VARCHAR(100),
    conditions JSONB, -- 额外的权限条件
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_device_type_id ON lj_extended_permissions(device_type_id);
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_module_id ON lj_extended_permissions(module_id);
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_element_type ON lj_extended_permissions(element_type);
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_resource ON lj_extended_permissions(resource);
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_action ON lj_extended_permissions(action);

-- 权限继承表，支持权限继承关系
CREATE TABLE IF NOT EXISTS lj_permission_inheritance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_permission_id UUID NOT NULL REFERENCES lj_permissions(id) ON DELETE CASCADE,
    child_permission_id UUID NOT NULL REFERENCES lj_permissions(id) ON DELETE CASCADE,
    inheritance_type VARCHAR(20) DEFAULT 'inherit', -- inherit, exclude, override
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(parent_permission_id, child_permission_id)
);

CREATE INDEX IF NOT EXISTS idx_lj_permission_inheritance_parent ON lj_permission_inheritance(parent_permission_id);
CREATE INDEX IF NOT EXISTS idx_lj_permission_inheritance_child ON lj_permission_inheritance(child_permission_id);

-- 用户组权限表
CREATE TABLE IF NOT EXISTS lj_user_group_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES lj_user_groups(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES lj_permissions(id) ON DELETE CASCADE,
    device_type_id UUID REFERENCES lj_device_types(id) ON DELETE CASCADE,
    granted BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(group_id, permission_id, device_type_id)
);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_group_id ON lj_user_group_permissions(group_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_permission_id ON lj_user_group_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_device_type_id ON lj_user_group_permissions(device_type_id);

-- 用户设备权限表，支持用户设备级别的权限控制
CREATE TABLE IF NOT EXISTS lj_user_device_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    device_type_id UUID NOT NULL REFERENCES lj_device_types(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES lj_permissions(id) ON DELETE CASCADE,
    granted BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, device_type_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_id ON lj_user_device_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_device_type_id ON lj_user_device_permissions(device_type_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_permission_id ON lj_user_device_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_expires_at ON lj_user_device_permissions(expires_at);

-- 插入默认角色
INSERT INTO lj_roles (name, display_name, description, is_system) VALUES
('super_admin', '超级管理员', '系统超级管理员，拥有所有权', true),
('admin', '管理员', '系统管理员，拥有大部分管理权', true),
('user_manager', '用户管理', '负责用户管理的管理员', false),
('viewer', '查看器', '只能查看数据，无修改权限', false)
ON CONFLICT (name) DO NOTHING;

-- 插入默认权限
INSERT INTO lj_permissions (name, code, description, resource, action) VALUES
('用户查看', 'user:read', '查看用户信息', 'user', 'read'),
('用户创建', 'user:create', '创建新用户', 'user', 'create'),
('用户更新', 'user:update', '更新用户信息', 'user', 'update'),
('用户删除', 'user:delete', '删除用户', 'user', 'delete'),
('角色查看', 'role:read', '查看角色信息', 'role', 'read'),
('角色创建', 'role:create', '创建新角色', 'role', 'create'),
('角色更新', 'role:update', '更新角色信息', 'role', 'update'),
('角色删除', 'role:delete', '删除角色', 'role', 'delete'),
('权限查看', 'permission:read', '查看权限信息', 'permission', 'read'),
('菜单查看', 'menu:read', '查看菜单信息', 'menu', 'read'),
('菜单管理', 'menu:manage', '管理菜单结构', 'menu', 'manage'),
('系统设置', 'system:settings', '系统设置管理', 'system', 'settings')
ON CONFLICT (code) DO NOTHING;
-- 标记默认权限为系统权限
UPDATE lj_permissions SET is_system = true WHERE code IN (
  'user:read','user:create','user:update','user:delete',
  'role:read','role:create','role:update','role:delete',
  'permission:read','menu:read','menu:manage','system:settings'
);

-- 插入默认菜单
INSERT INTO lj_menus (title, path, icon, component, sort_order) VALUES
('仪表板', '/dashboard', 'DashboardOutlined', 'Dashboard', 1),
('用户管理', '/users', 'UserOutlined', 'UserManagement', 2),
('角色管理', '/roles', 'TeamOutlined', 'RoleManagement', 3),
('权限管理', '/permissions', 'SafetyOutlined', 'PermissionManagement', 4),
('用户组管理', '/user-groups', 'UsergroupAddOutlined', 'UserGroupManagement', 5),
('菜单管理', '/menus', 'MenuOutlined', 'MenuManagement', 6),
('系统设置', '/settings', 'SettingOutlined', 'SystemSettings', 7)
ON CONFLICT (title, path, component) DO NOTHING;

-- 插入默认设备类型
INSERT INTO lj_device_types (code, name, description, icon, sort_order) VALUES
('web', 'WEB端', '网页浏览器端应用', 'DesktopOutlined', 1),
('desktop', '桌面端', '桌面应用程序', 'LaptopOutlined', 2),
('mobile', '手机端', '手机移动应用', 'MobileOutlined', 3),
('tablet', '平板端', '平板设备应用', 'TabletOutlined', 4),
('watch', '手表端', '智能手表应用', 'WatchOutlined', 5),
('glasses', '眼镜端', '智能眼镜应用', 'EyeOutlined', 6),
('iot', '物联端', '物联网设备应用', 'ApiOutlined', 7),
('robot', '机器人端', '机器人控制端', 'RobotOutlined', 8)
ON CONFLICT (code) DO NOTHING;

-- 插入默认模块
INSERT INTO lj_modules (code, name, description, icon, sort_order) VALUES
('system', '系统管理', '系统核心管理功能', 'SettingOutlined', 1),
('user', '用户管理', '用户相关管理功能', 'UserOutlined', 2),
('permission', '权限管理', '权限控制管理功能', 'SafetyOutlined', 3),
('content', '内容管理', '内容发布管理功能', 'FileTextOutlined', 4),
('analytics', '数据分析', '数据统计分析功能', 'BarChartOutlined', 5),
('notification', '消息通知', '消息推送管理功能', 'BellOutlined', 6),
('workflow', '工作流程', '业务流程管理功能', 'NodeIndexOutlined', 7),
('integration', '系统集成', '第三方系统集成功能', 'LinkOutlined', 8)
ON CONFLICT (code) DO NOTHING;

-- 创建默认超级管理员用户(密码: admin123)
INSERT INTO lj_users (username, email, password_hash, is_active) VALUES
('admin', 'admin@laojun.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', true)
ON CONFLICT (username) DO NOTHING;

-- 为超级管理员分配角色
INSERT INTO lj_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM lj_users u, lj_roles r
WHERE u.username = 'admin' AND r.name = 'super_admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- 为超级管理员角色分配所有权权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为管理员角色分配基本权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'admin' AND p.code IN ('user:read', 'user:create', 'user:update', 'role:read', 'menu:read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为用户管理员分配用户相关权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'user_manager' AND p.code IN ('user:read', 'user:create', 'user:update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为查看者分配只读权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'viewer' AND p.code IN ('user:read', 'role:read', 'permission:read', 'menu:read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 插入扩展权限（细粒度权限）
INSERT INTO lj_extended_permissions (module_id, device_type_id, resource, action, description) VALUES
-- 系统管理模块权限
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'dashboard', 'view', 'WEB端查看仪表盘'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'system_config', 'view', 'WEB端查看系统配置'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'system_config', 'edit', 'WEB端编辑系统配置'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'logs', 'view', 'WEB端查看系统日志'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'backup', 'create', 'WEB端创建系统备份'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'backup', 'restore', 'WEB端恢复系统备份'),

-- 用户管理模块权限
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'list', 'WEB端查看用户列表'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'view', 'WEB端查看用户详情'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'create', 'WEB端创建用户'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'edit', 'WEB端编辑用户'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'delete', 'WEB端删除用户'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'reset_password', 'WEB端重置用户密码'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'change_status', 'WEB端修改用户状态'),

-- 权限管理模块权限
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'list', 'WEB端查看角色列表'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'view', 'WEB端查看角色详情'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'create', 'WEB端创建角色'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'edit', 'WEB端编辑角色'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'delete', 'WEB端删除角色'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'permission', 'list', 'WEB端查看权限列表'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'permission', 'assign', 'WEB端分配权限'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'list', 'WEB端查看用户组列表'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'create', 'WEB端创建用户组'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'edit', 'WEB端编辑用户组'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'delete', 'WEB端删除用户组'),

-- 移动端权限示例
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'mobile'), 'user', 'list', '手机端查看用户列表'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'mobile'), 'user', 'view', '手机端查看用户详情'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'mobile'), 'dashboard', 'view', '手机端查看仪表盘'),

-- 物联网设备权限示例
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'iot'), 'device_status', 'report', '物联网设备状态上报'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'iot'), 'device_control', 'execute', '物联网设备控制执行')
ON CONFLICT DO NOTHING;

-- 插入权限模板
INSERT INTO lj_permission_templates (name, description, template_data) VALUES
('超级管理员模板', '拥有所有权限的超级管理员模板', '{"all_permissions": true, "description": "完全访问权限"}'),
('普通管理员模板', '普通管理员权限模板，不包含系统核心配置权限', '{"modules": ["user", "permission"], "exclude_actions": ["system_config.edit", "backup.restore"]}'),
('用户管理员模板', '专门负责用户管理的管理员模板', '{"modules": ["user"], "actions": ["user.*", "user_group.*"]}'),
('只读用户模板', '只能查看信息，无法进行任何修改操作', '{"actions": ["*.view", "*.list"], "exclude_actions": ["*.create", "*.edit", "*.delete"]}'),
('移动端用户模板', '移动端用户基础权限模板', '{"device_types": ["mobile"], "modules": ["user"], "actions": ["dashboard.view", "user.list", "user.view"]}'),
('物联网设备模板', '物联网设备基础权限模板', '{"device_types": ["iot"], "modules": ["system"], "actions": ["device_status.report", "device_control.execute"]}')
ON CONFLICT (name) DO NOTHING;

-- ========================================
-- 复合索引优化 - 提升查询性能
-- ========================================

-- 用户认证和会话管理复合索引
CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_expires 
ON lj_user_sessions(user_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_token_expires 
ON lj_user_sessions(token_hash, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_active 
ON lj_user_sessions(user_id, expires_at, last_used_at);

CREATE INDEX IF NOT EXISTS idx_lj_jwt_keys_active_expires 
ON lj_jwt_keys(is_active, expires_at);

-- 权限查询优化复合索引
CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_group_device 
ON lj_user_group_permissions(group_id, device_type_id, permission_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_perm_granted 
ON lj_user_group_permissions(permission_id, granted, group_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_device 
ON lj_user_device_permissions(user_id, device_type_id, permission_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_valid 
ON lj_user_device_permissions(user_id, granted, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_cleanup 
ON lj_user_device_permissions(expires_at) WHERE expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lj_permission_inheritance_parent_type 
ON lj_permission_inheritance(parent_permission_id, inheritance_type);

-- 扩展权限查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_device_module 
ON lj_extended_permissions(device_type_id, module_id, resource, action);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_module_element 
ON lj_extended_permissions(module_id, element_type, element_code);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_resource_action 
ON lj_extended_permissions(resource, action, device_type_id);

-- 层级结构查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_menus_parent_sort 
ON lj_menus(parent_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_menus_visible_sort 
ON lj_menus(is_hidden, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_modules_parent_active_sort 
ON lj_modules(parent_id, is_active, sort_order);

-- 状态和时间范围查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_users_active_created 
ON lj_users(is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_users_active_login 
ON lj_users(is_active, last_login_at) WHERE last_login_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lj_device_types_active_sort 
ON lj_device_types(is_active, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_user_groups_active_created 
ON lj_user_groups(is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_user_created 
ON lj_user_group_members(user_id, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_group_created 
ON lj_user_group_members(group_id, created_at);

-- 权限模板查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_permission_templates_system_created 
ON lj_permission_templates(is_system, created_at);

-- 角色和权限关联复合索引
CREATE INDEX IF NOT EXISTS idx_lj_roles_system_created 
ON lj_roles(is_system, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_permissions_resource_action 
ON lj_permissions(resource, action);

-- 性能监控和统计查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_users_stats 
ON lj_users(created_at, is_active);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_stats 
ON lj_user_sessions(created_at, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_stats 
ON lj_user_device_permissions(permission_id, device_type_id, created_at);

-- ========================================
-- Marketplace 相关索引
-- ========================================

-- 删除现有的marketplace表（如果存在）
DROP TABLE IF EXISTS purchases CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;

-- ========================================
-- 系统设置与审计日志索引
-- ========================================

-- 系统设置索引
CREATE TABLE IF NOT EXISTS lj_system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'string', -- string, number, boolean, json, list
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    is_sensitive BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_system_settings_key ON lj_system_settings(key);
CREATE INDEX IF NOT EXISTS idx_lj_system_settings_category ON lj_system_settings(category);
CREATE INDEX IF NOT EXISTS idx_lj_system_settings_public ON lj_system_settings(is_public);

-- 审计日志索引
CREATE TABLE IF NOT EXISTS lj_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES lj_users(id) ON DELETE SET NULL,
    target_id UUID,
    target_type VARCHAR(50),
    action VARCHAR(50) NOT NULL,
    level VARCHAR(20) DEFAULT 'info', -- debug, info, warn, error, fatal
    description VARCHAR(500),
    old_data TEXT,
    new_data TEXT,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lj_audit_logs_created_at ON lj_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_lj_audit_logs_action ON lj_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_lj_audit_logs_target_type ON lj_audit_logs(target_type);

-- 预置部分系统设置，方便前端展示
INSERT INTO lj_system_settings (key, value, type, category, description, is_public)
VALUES
('system.name', '{"value": "太上老君管理系统"}', 'string', 'general', '系统名称', true),
('system.version', '{"value": "1.0.0"}', 'string', 'general', '系统版本', true),
('security.session_timeout', '{"value": 3600}', 'number', 'security', '会话超时时间（秒）', false),
('logging.level', '{"value": "info"}', 'string', 'logging', '日志级别', false)
ON CONFLICT (key) DO NOTHING;

-- 插入默认设备类型
INSERT INTO lj_device_types (code, name, description, icon, sort_order) VALUES
('web', 'WEB端', '网页浏览器端应用', 'DesktopOutlined', 1),
('desktop', '桌面端', '桌面应用程序', 'LaptopOutlined', 2),
('mobile', '手机端', '手机移动应用', 'MobileOutlined', 3),
('tablet', '平板端', '平板设备应用', 'TabletOutlined', 4),
('watch', '手表端', '智能手表应用', 'WatchOutlined', 5),
('glasses', '眼镜端', '智能眼镜应用', 'EyeOutlined', 6),
('iot', '物联端', '物联网设备应用', 'ApiOutlined', 7),
('robot', '机器人端', '机器人控制端', 'RobotOutlined', 8)
ON CONFLICT (code) DO NOTHING;

-- 插入默认模块
INSERT INTO lj_modules (code, name, description, icon, sort_order) VALUES
('system', '系统管理', '系统核心管理功能', 'SettingOutlined', 1),
('user', '用户管理', '用户相关管理功能', 'UserOutlined', 2),
('permission', '权限管理', '权限控制管理功能', 'SafetyOutlined', 3),
('content', '内容管理', '内容发布管理功能', 'FileTextOutlined', 4),
('analytics', '数据分析', '数据统计分析功能', 'BarChartOutlined', 5),
('notification', '消息通知', '消息推送管理功能', 'BellOutlined', 6),
('workflow', '工作流程', '业务流程管理功能', 'NodeIndexOutlined', 7),
('integration', '系统集成', '第三方系统集成功能', 'LinkOutlined', 8)
ON CONFLICT (code) DO NOTHING;

-- 创建默认超级管理员用户(密码: admin123)
INSERT INTO lj_users (username, email, password_hash, is_active) VALUES
('admin', 'admin@laojun.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', true)
ON CONFLICT (username) DO NOTHING;

-- 为超级管理员分配角色
INSERT INTO lj_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM lj_users u, lj_roles r
WHERE u.username = 'admin' AND r.name = 'super_admin'
ON CONFLICT (user_id, role_id) DO NOTHING;

-- 为超级管理员角色分配所有权权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为管理员角色分配基本权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'admin' AND p.code IN ('user:read', 'user:create', 'user:update', 'role:read', 'menu:read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为用户管理员分配用户相关权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'user_manager' AND p.code IN ('user:read', 'user:create', 'user:update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 为查看者分配只读权限
INSERT INTO lj_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM lj_roles r, lj_permissions p
WHERE r.name = 'viewer' AND p.code IN ('user:read', 'role:read', 'permission:read', 'menu:read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 插入扩展权限（细粒度权限）
INSERT INTO lj_extended_permissions (module_id, device_type_id, resource, action, description) VALUES
-- 系统管理模块权限
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'dashboard', 'view', 'WEB端查看仪表盘'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'system_config', 'view', 'WEB端查看系统配置'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'system_config', 'edit', 'WEB端编辑系统配置'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'logs', 'view', 'WEB端查看系统日志'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'backup', 'create', 'WEB端创建系统备份'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'backup', 'restore', 'WEB端恢复系统备份'),

-- 用户管理模块权限
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'list', 'WEB端查看用户列表'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'view', 'WEB端查看用户详情'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'create', 'WEB端创建用户'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'edit', 'WEB端编辑用户'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'delete', 'WEB端删除用户'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'reset_password', 'WEB端重置用户密码'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user', 'change_status', 'WEB端修改用户状态'),

-- 权限管理模块权限
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'list', 'WEB端查看角色列表'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'view', 'WEB端查看角色详情'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'create', 'WEB端创建角色'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'edit', 'WEB端编辑角色'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'role', 'delete', 'WEB端删除角色'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'permission', 'list', 'WEB端查看权限列表'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'permission', 'assign', 'WEB端分配权限'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'list', 'WEB端查看用户组列表'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'create', 'WEB端创建用户组'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'edit', 'WEB端编辑用户组'),
((SELECT id FROM lj_modules WHERE code = 'permission'), (SELECT id FROM lj_device_types WHERE code = 'web'), 'user_group', 'delete', 'WEB端删除用户组'),

-- 移动端权限示例
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'mobile'), 'user', 'list', '手机端查看用户列表'),
((SELECT id FROM lj_modules WHERE code = 'user'), (SELECT id FROM lj_device_types WHERE code = 'mobile'), 'user', 'view', '手机端查看用户详情'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'mobile'), 'dashboard', 'view', '手机端查看仪表盘'),

-- 物联网设备权限示例
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'iot'), 'device_status', 'report', '物联网设备状态上报'),
((SELECT id FROM lj_modules WHERE code = 'system'), (SELECT id FROM lj_device_types WHERE code = 'iot'), 'device_control', 'execute', '物联网设备控制执行')
ON CONFLICT DO NOTHING;

-- 插入权限模板
INSERT INTO lj_permission_templates (name, description, template_data) VALUES
('超级管理员模板', '拥有所有权限的超级管理员模板', '{"all_permissions": true, "description": "完全访问权限"}'),
('普通管理员模板', '普通管理员权限模板，不包含系统核心配置权限', '{"modules": ["user", "permission"], "exclude_actions": ["system_config.edit", "backup.restore"]}'),
('用户管理员模板', '专门负责用户管理的管理员模板', '{"modules": ["user"], "actions": ["user.*", "user_group.*"]}'),
('只读用户模板', '只能查看信息，无法进行任何修改操作', '{"actions": ["*.view", "*.list"], "exclude_actions": ["*.create", "*.edit", "*.delete"]}'),
('移动端用户模板', '移动端用户基础权限模板', '{"device_types": ["mobile"], "modules": ["user"], "actions": ["dashboard.view", "user.list", "user.view"]}'),
('物联网设备模板', '物联网设备基础权限模板', '{"device_types": ["iot"], "modules": ["system"], "actions": ["device_status.report", "device_control.execute"]}')
ON CONFLICT (name) DO NOTHING;

-- ========================================
-- 复合索引优化 - 提升查询性能
-- ========================================

-- 用户认证和会话管理复合索引
CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_expires 
ON lj_user_sessions(user_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_token_expires 
ON lj_user_sessions(token_hash, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_user_active 
ON lj_user_sessions(user_id, expires_at, last_used_at);

CREATE INDEX IF NOT EXISTS idx_lj_jwt_keys_active_expires 
ON lj_jwt_keys(is_active, expires_at);

-- 权限查询优化复合索引
CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_group_device 
ON lj_user_group_permissions(group_id, device_type_id, permission_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_permissions_perm_granted 
ON lj_user_group_permissions(permission_id, granted, group_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_device 
ON lj_user_device_permissions(user_id, device_type_id, permission_id);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_user_valid 
ON lj_user_device_permissions(user_id, granted, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_cleanup 
ON lj_user_device_permissions(expires_at) WHERE expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lj_permission_inheritance_parent_type 
ON lj_permission_inheritance(parent_permission_id, inheritance_type);

-- 扩展权限查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_device_module 
ON lj_extended_permissions(device_type_id, module_id, resource, action);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_module_element 
ON lj_extended_permissions(module_id, element_type, element_code);

CREATE INDEX IF NOT EXISTS idx_lj_extended_permissions_resource_action 
ON lj_extended_permissions(resource, action, device_type_id);

-- 层级结构查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_menus_parent_sort 
ON lj_menus(parent_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_menus_visible_sort 
ON lj_menus(is_hidden, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_modules_parent_active_sort 
ON lj_modules(parent_id, is_active, sort_order);

-- 状态和时间范围查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_users_active_created 
ON lj_users(is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_users_active_login 
ON lj_users(is_active, last_login_at) WHERE last_login_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lj_device_types_active_sort 
ON lj_device_types(is_active, sort_order);

CREATE INDEX IF NOT EXISTS idx_lj_user_groups_active_created 
ON lj_user_groups(is_active, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_user_created 
ON lj_user_group_members(user_id, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_group_members_group_created 
ON lj_user_group_members(group_id, created_at);

-- 权限模板查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_permission_templates_system_created 
ON lj_permission_templates(is_system, created_at);

-- 角色和权限关联复合索引
CREATE INDEX IF NOT EXISTS idx_lj_roles_system_created 
ON lj_roles(is_system, created_at);

CREATE INDEX IF NOT EXISTS idx_lj_permissions_resource_action 
ON lj_permissions(resource, action);

-- 性能监控和统计查询复合索引
CREATE INDEX IF NOT EXISTS idx_lj_users_stats 
ON lj_users(created_at, is_active);

CREATE INDEX IF NOT EXISTS idx_lj_user_sessions_stats 
ON lj_user_sessions(created_at, expires_at);

CREATE INDEX IF NOT EXISTS idx_lj_user_device_permissions_stats 
ON lj_user_device_permissions(permission_id, device_type_id, created_at);

-- ========================================
-- Marketplace 相关索引
-- ========================================

-- 清理旧表（按依赖顺序）
DROP TABLE IF EXISTS user_favorites CASCADE;
DROP TABLE IF EXISTS purchases CASCADE;
DROP TABLE IF EXISTS reviews CASCADE;
DROP TABLE IF EXISTS plugin_versions CASCADE;
DROP TABLE IF EXISTS plugins CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS developers CASCADE;

-- 开发者表
CREATE TABLE IF NOT EXISTS developers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    display_name VARCHAR(100) NOT NULL,
    company_name VARCHAR(200),
    website_url VARCHAR(500),
    github_url VARCHAR(500),
    bio TEXT,
    avatar_url VARCHAR(500),
    is_verified BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    total_downloads INTEGER DEFAULT 0,
    total_revenue DECIMAL(12,2) DEFAULT 0.00,
    plugin_count INTEGER DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX IF NOT EXISTS idx_developers_user_id ON developers(user_id);
CREATE INDEX IF NOT EXISTS idx_developers_is_verified ON developers(is_verified);
CREATE INDEX IF NOT EXISTS idx_developers_is_active ON developers(is_active);
CREATE INDEX IF NOT EXISTS idx_developers_total_downloads ON developers(total_downloads);
CREATE INDEX IF NOT EXISTS idx_developers_average_rating ON developers(average_rating);
CREATE INDEX IF NOT EXISTS idx_developers_created_at ON developers(created_at);

-- 插件分类表
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    color VARCHAR(20) DEFAULT '#1890ff',
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order);

-- 插件表
CREATE TABLE IF NOT EXISTS plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    author VARCHAR(100) NOT NULL,
    developer_id UUID REFERENCES developers(id) ON DELETE SET NULL,
    version VARCHAR(50) NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    price DECIMAL(10,2) DEFAULT 0.00,
    is_free BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    download_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    review_count INTEGER DEFAULT 0,
    icon_url VARCHAR(500),
    banner_url VARCHAR(500),
    screenshots TEXT[],
    tags TEXT[],
    requirements JSONB,
    changelog TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plugins_category_id ON plugins(category_id);
CREATE INDEX IF NOT EXISTS idx_plugins_developer_id ON plugins(developer_id);
CREATE INDEX IF NOT EXISTS idx_plugins_is_featured ON plugins(is_featured);
CREATE INDEX IF NOT EXISTS idx_plugins_is_active ON plugins(is_active);
CREATE INDEX IF NOT EXISTS idx_plugins_rating ON plugins(rating);
CREATE INDEX IF NOT EXISTS idx_plugins_download_count ON plugins(download_count);
CREATE INDEX IF NOT EXISTS idx_plugins_created_at ON plugins(created_at);

-- 插件版本表
CREATE TABLE IF NOT EXISTS plugin_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    release_notes TEXT,
    download_url VARCHAR(500),
    file_size BIGINT,
    checksum VARCHAR(255),
    is_stable BOOLEAN DEFAULT true,
    min_system_version VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(plugin_id, version)
);
CREATE INDEX IF NOT EXISTS idx_plugin_versions_plugin_id ON plugin_versions(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_versions_is_stable ON plugin_versions(is_stable);

-- 插件评论表
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    user_id UUID REFERENCES lj_users(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(200),
    content TEXT,
    is_verified BOOLEAN DEFAULT false,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reviews_plugin_id ON reviews(plugin_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at);

-- 插件购买记录表
CREATE TABLE IF NOT EXISTS purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(50),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, plugin_id)
);
CREATE INDEX IF NOT EXISTS idx_purchases_user_id ON purchases(user_id);
CREATE INDEX IF NOT EXISTS idx_purchases_plugin_id ON purchases(plugin_id);
CREATE INDEX IF NOT EXISTS idx_purchases_status ON purchases(status);
CREATE INDEX IF NOT EXISTS idx_purchases_created_at ON purchases(created_at);

-- 用户收藏表
CREATE TABLE IF NOT EXISTS user_favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES lj_users(id) ON DELETE CASCADE,
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, plugin_id)
);
CREATE INDEX IF NOT EXISTS idx_user_favorites_user_id ON user_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_plugin_id ON user_favorites(plugin_id);
CREATE INDEX IF NOT EXISTS idx_user_favorites_created_at ON user_favorites(created_at);

-- 插入默认分类
INSERT INTO categories (name, description, icon, color, sort_order, is_active) VALUES
('开发工具', '提升开发效率的工具和插件', 'CodeOutlined', '#1890ff', 1, true),
('UI组件', '用户界面组件和主界面元素', 'AppstoreOutlined', '#52c41a', 2, true),
('数据分析', '数据处理和分析工具', 'BarChartOutlined', '#fa8c16', 3, true),
('安全工具', '安全防护和检测工具', 'SafetyOutlined', '#f5222d', 4, true),
('效率工具', '提升工作效率的实用工具', 'ThunderboltOutlined', '#722ed1', 5, true),
('娱乐游戏', '休闲娱乐和游戏插件', 'SmileOutlined', '#eb2f96', 6, true)
ON CONFLICT DO NOTHING;
`
