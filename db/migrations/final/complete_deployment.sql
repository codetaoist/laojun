-- 太上老君系统 - 完整数据库部署文件
-- 包含所有表结构和丰富的测试数据
-- 生成时间: 2025-01-23
-- 版本: v2.0 Enhanced

-- 启用必要的扩展
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

-- 市场用户表
CREATE TABLE IF NOT EXISTS mp_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar TEXT,
    bio TEXT,
    website VARCHAR(255),
    github_username VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    reputation_score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- 论坛分类表
CREATE TABLE IF NOT EXISTS mp_forum_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(7) DEFAULT '#1890ff',
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    post_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 论坛帖子表
CREATE TABLE IF NOT EXISTS mp_forum_posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    category_id UUID REFERENCES mp_forum_categories(id) ON DELETE SET NULL,
    user_id UUID REFERENCES mp_users(id) ON DELETE CASCADE,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 论坛回复表
CREATE TABLE IF NOT EXISTS mp_forum_replies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content TEXT NOT NULL,
    post_id UUID NOT NULL REFERENCES mp_forum_posts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES mp_users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES mp_forum_replies(id) ON DELETE CASCADE,
    like_count INTEGER DEFAULT 0,
    is_accepted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 博客分类表
CREATE TABLE IF NOT EXISTS mp_blog_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(7) DEFAULT '#1890ff',
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    post_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 博客文章表
CREATE TABLE IF NOT EXISTS mp_blog_posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    excerpt TEXT,
    category_id UUID REFERENCES mp_blog_categories(id) ON DELETE SET NULL,
    author_id UUID REFERENCES mp_users(id) ON DELETE CASCADE,
    is_published BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    tags TEXT[],
    cover_image TEXT,
    reading_time INTEGER DEFAULT 0,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 菜单表
CREATE TABLE IF NOT EXISTS sm_menus (
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

-- 插件分类表
CREATE TABLE IF NOT EXISTS mp_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    color VARCHAR(7) DEFAULT '#1890ff',
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    plugin_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插件表
CREATE TABLE IF NOT EXISTS mp_plugins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    short_description TEXT,
    author VARCHAR(100) NOT NULL,
    developer_id UUID REFERENCES mp_users(id) ON DELETE SET NULL,
    version VARCHAR(50) NOT NULL,
    category_id UUID REFERENCES mp_categories(id) ON DELETE SET NULL,
    price DECIMAL(10,2) DEFAULT 0.00,
    is_free BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    download_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.00,
    review_count INTEGER DEFAULT 0,
    icon_url TEXT,
    banner_url TEXT,
    screenshots TEXT[],
    tags TEXT[],
    requirements TEXT,
    changelog TEXT,
    license VARCHAR(50) DEFAULT 'MIT',
    repository_url TEXT,
    documentation_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 插件评论表
CREATE TABLE IF NOT EXISTS mp_plugin_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
    user_id UUID REFERENCES mp_users(id) ON DELETE CASCADE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(200),
    content TEXT,
    is_verified_purchase BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS sm_system_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT,
    config_type VARCHAR(50) DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 基础数据插入
-- =============================================================================

-- 插入扩展权限数据
INSERT INTO az_permissions (id, name, code, description, resource, action) VALUES
('11111111-1111-1111-1111-111111111111', '用户管理', 'user.manage', '管理系统用户', 'user', 'manage'),
('22222222-2222-2222-2222-222222222222', '角色管理', 'role.manage', '管理系统角色', 'role', 'manage'),
('33333333-3333-3333-3333-333333333333', '权限管理', 'permission.manage', '管理系统权限', 'permission', 'manage'),
('44444444-4444-4444-4444-444444444444', '菜单管理', 'menu.manage', '管理系统菜单', 'menu', 'manage'),
('55555555-5555-5555-5555-555555555555', '插件管理', 'plugin.manage', '管理插件', 'plugin', 'manage'),
('66666666-6666-6666-6666-666666666666', '系统设置', 'system.setting', '系统设置管理', 'system', 'setting'),
('77777777-7777-7777-7777-777777777777', '用户查看', 'user.view', '查看用户信息', 'user', 'view'),
('88888888-8888-8888-8888-888888888888', '论坛管理', 'forum.manage', '管理论坛内容', 'forum', 'manage'),
('99999999-9999-9999-9999-999999999999', '博客管理', 'blog.manage', '管理博客内容', 'blog', 'manage'),
('aaaaaaaa-1111-1111-1111-111111111111', '插件审核', 'plugin.review', '审核插件发布', 'plugin', 'review'),
('bbbbbbbb-1111-1111-1111-111111111111', '数据统计', 'data.analytics', '查看系统数据统计', 'data', 'analytics'),
('cccccccc-1111-1111-1111-111111111111', '内容发布', 'content.publish', '发布内容', 'content', 'publish'),
('dddddddd-1111-1111-1111-111111111111', '评论管理', 'comment.manage', '管理评论', 'comment', 'manage'),
('eeeeeeee-1111-1111-1111-111111111111', '文件管理', 'file.manage', '管理文件上传', 'file', 'manage'),
('ffffffff-1111-1111-1111-111111111111', '系统监控', 'system.monitor', '系统监控', 'system', 'monitor');

-- 插入角色数据
INSERT INTO az_roles (id, name, display_name, description, is_system) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin', '超级管理员', '系统超级管理员，拥有所有权限', true),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'admin', '管理员', '系统管理员', true),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'moderator', '版主', '论坛版主', false),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'developer', '开发者', '插件开发者', false),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'editor', '编辑', '内容编辑', false),
('ffffffff-ffff-ffff-ffff-ffffffffffff', 'user', '普通用户', '普通用户', false);

-- 插入角色权限关联数据
-- 超级管理员拥有所有权限
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '22222222-2222-2222-2222-222222222222'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '33333333-3333-3333-3333-333333333333'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '44444444-4444-4444-4444-444444444444'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '55555555-5555-5555-5555-555555555555'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '66666666-6666-6666-6666-666666666666'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '77777777-7777-7777-7777-777777777777'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '88888888-8888-8888-8888-888888888888'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '99999999-9999-9999-9999-999999999999'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaaaaaa-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbbbbbb-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'cccccccc-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'dddddddd-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'eeeeeeee-1111-1111-1111-111111111111'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'ffffffff-1111-1111-1111-111111111111');

-- 管理员权限
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444444'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '55555555-5555-5555-5555-555555555555'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '77777777-7777-7777-7777-777777777777'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '88888888-8888-8888-8888-888888888888'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '99999999-9999-9999-9999-999999999999'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaaa-1111-1111-1111-111111111111'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'bbbbbbbb-1111-1111-1111-111111111111'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'dddddddd-1111-1111-1111-111111111111'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'eeeeeeee-1111-1111-1111-111111111111');

-- 版主权限
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('cccccccc-cccc-cccc-cccc-cccccccccccc', '77777777-7777-7777-7777-777777777777'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '88888888-8888-8888-8888-888888888888'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '99999999-9999-9999-9999-999999999999'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'cccccccc-1111-1111-1111-111111111111'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'dddddddd-1111-1111-1111-111111111111');

-- 开发者权限
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('dddddddd-dddd-dddd-dddd-dddddddddddd', '77777777-7777-7777-7777-777777777777'),
('dddddddd-dddd-dddd-dddd-dddddddddddd', '55555555-5555-5555-5555-555555555555'),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'cccccccc-1111-1111-1111-111111111111'),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'eeeeeeee-1111-1111-1111-111111111111');

-- 编辑权限
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '77777777-7777-7777-7777-777777777777'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '99999999-9999-9999-9999-999999999999'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'cccccccc-1111-1111-1111-111111111111'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'dddddddd-1111-1111-1111-111111111111');

-- 普通用户权限
INSERT INTO az_role_permissions (role_id, permission_id) VALUES
('ffffffff-ffff-ffff-ffff-ffffffffffff', '77777777-7777-7777-7777-777777777777'),
('ffffffff-ffff-ffff-ffff-ffffffffffff', 'cccccccc-1111-1111-1111-111111111111');

-- 插入管理员用户数据 (密码: admin123)
INSERT INTO ua_admin (id, username, email, password_hash, full_name, department, position, is_active, is_super_admin, login_count) VALUES
('12345678-1234-1234-1234-123456789012', 'admin', 'admin@laojun.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '系统管理员', '技术部', '系统管理员', true, true, 156),
('12345678-1234-1234-1234-123456789013', 'manager', 'manager@laojun.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '项目经理', '产品部', '项目经理', true, false, 89),
('12345678-1234-1234-1234-123456789014', 'moderator', 'mod@laojun.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '社区版主', '运营部', '社区版主', true, false, 234),
('12345678-1234-1234-1234-123456789015', 'editor', 'editor@laojun.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '内容编辑', '内容部', '高级编辑', true, false, 67);

-- 插入用户角色关联
INSERT INTO az_user_roles (user_id, role_id) VALUES
('12345678-1234-1234-1234-123456789012', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
('12345678-1234-1234-1234-123456789013', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
('12345678-1234-1234-1234-123456789014', 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
('12345678-1234-1234-1234-123456789015', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee');

-- 插入市场用户数据
INSERT INTO mp_users (id, username, email, password_hash, full_name, bio, github_username, is_active, is_verified, reputation_score) VALUES
('87654321-4321-4321-4321-210987654321', 'developer1', 'dev1@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '张开发', '全栈开发工程师，专注于Go和Vue.js开发', 'zhangdev', true, true, 1250),
('87654321-4321-4321-4321-210987654322', 'developer2', 'dev2@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '李程序', '后端开发专家，微服务架构师', 'lichengxu', true, true, 980),
('87654321-4321-4321-4321-210987654323', 'designer1', 'design@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '王设计', 'UI/UX设计师，热爱创造美好的用户体验', 'wangdesign', true, true, 750),
('87654321-4321-4321-4321-210987654324', 'tester1', 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '赵测试', '软件测试工程师，质量保证专家', 'zhaoceshi', true, false, 420),
('87654321-4321-4321-4321-210987654325', 'blogger1', 'blog@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '刘博主', '技术博主，分享编程心得和技术见解', 'liubozhu', true, true, 890),
('87654321-4321-4321-4321-210987654326', 'student1', 'student@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '陈学生', '计算机专业学生，正在学习软件开发', 'chenxuesheng', true, false, 180),
('87654321-4321-4321-4321-210987654327', 'entrepreneur', 'startup@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1VdLSbn9RQoQHKI6qIqg.z.nQ9QdvKe', '孙创业', '技术创业者，专注于企业级解决方案', 'sunchuangye', true, true, 1100);

-- 插入论坛分类数据
INSERT INTO mp_forum_categories (id, name, description, icon, color, sort_order, post_count) VALUES
('f1111111-1111-1111-1111-111111111111', '技术讨论', '技术相关话题讨论，包括编程语言、框架、工具等', 'code', '#1890ff', 1, 45),
('f2222222-2222-2222-2222-222222222222', '插件开发', '插件开发相关讨论，分享开发经验和技巧', 'plugin', '#52c41a', 2, 28),
('f3333333-3333-3333-3333-333333333333', '问题反馈', '系统问题反馈和bug报告', 'bug', '#ff4d4f', 3, 12),
('f4444444-4444-4444-4444-444444444444', '新手指南', '新手使用指南和入门教程', 'book', '#faad14', 4, 23),
('f5555555-5555-5555-5555-555555555555', '项目展示', '展示你的项目和作品', 'project', '#722ed1', 5, 18),
('f6666666-6666-6666-6666-666666666666', '求职招聘', '技术岗位招聘和求职信息', 'team', '#13c2c2', 6, 8);

-- 插入博客分类数据
INSERT INTO mp_blog_categories (id, name, description, icon, color, sort_order, post_count) VALUES
('b1111111-1111-1111-1111-111111111111', '技术分享', '技术文章分享，深度技术解析', 'share', '#1890ff', 1, 32),
('b2222222-2222-2222-2222-222222222222', '开发日志', '开发过程记录和项目总结', 'edit', '#52c41a', 2, 18),
('b3333333-3333-3333-3333-333333333333', '教程指南', '详细的使用教程和指南', 'book', '#faad14', 3, 25),
('b4444444-4444-4444-4444-444444444444', '行业动态', '技术行业新闻和趋势分析', 'global', '#722ed1', 4, 14),
('b5555555-5555-5555-5555-555555555555', '工具推荐', '开发工具和资源推荐', 'tool', '#13c2c2', 5, 21);

-- 插入菜单数据
INSERT INTO sm_menus (id, title, path, icon, component, parent_id, sort_order, is_hidden, device_types, permissions) VALUES
-- 主菜单
('m1111111-1111-1111-1111-111111111111', '仪表盘', '/dashboard', 'dashboard', 'Dashboard', NULL, 1, false, '["pc","web"]', '["user.view"]'),
('m2222222-2222-2222-2222-222222222222', '用户管理', '/users', 'user', NULL, NULL, 2, false, '["pc","web"]', '["user.manage","user.view"]'),
('m3333333-3333-3333-3333-333333333333', '系统管理', '/system', 'setting', NULL, NULL, 3, false, '["pc","web"]', '["system.setting"]'),
('m4444444-4444-4444-4444-444444444444', '插件市场', '/marketplace', 'appstore', NULL, NULL, 4, false, '["pc","web"]', '["plugin.manage","plugin.view"]'),
('m5555555-5555-5555-5555-555555555555', '论坛', '/forum', 'message', NULL, NULL, 5, false, '["pc","web","mobile"]', '["forum.manage","content.publish"]'),
('m6666666-6666-6666-6666-666666666666', '博客', '/blog', 'read', NULL, NULL, 6, false, '["pc","web","mobile"]', '["blog.manage","content.publish"]'),
('m7777777-7777-7777-7777-777777777777', '数据统计', '/analytics', 'bar-chart', NULL, NULL, 7, false, '["pc","web"]', '["data.analytics"]'),

-- 用户管理子菜单
('m2111111-1111-1111-1111-111111111111', '用户列表', '/users/list', 'team', 'UserList', 'm2222222-2222-2222-2222-222222222222', 1, false, '["pc","web"]', '["user.view"]'),
('m2222222-1111-1111-1111-111111111111', '角色管理', '/users/roles', 'safety', 'RoleManagement', 'm2222222-2222-2222-2222-222222222222', 2, false, '["pc","web"]', '["role.manage"]'),
('m2333333-1111-1111-1111-111111111111', '权限管理', '/users/permissions', 'lock', 'PermissionManagement', 'm2222222-2222-2222-2222-222222222222', 3, false, '["pc","web"]', '["permission.manage"]'),

-- 系统管理子菜单
('m3111111-1111-1111-1111-111111111111', '菜单管理', '/system/menus', 'menu', 'MenuManagement', 'm3333333-3333-3333-3333-333333333333', 1, false, '["pc","web"]', '["menu.manage"]'),
('m3222222-1111-1111-1111-111111111111', '系统设置', '/system/settings', 'tool', 'SystemSettings', 'm3333333-3333-3333-3333-333333333333', 2, false, '["pc","web"]', '["system.setting"]'),
('m3333333-1111-1111-1111-111111111111', '日志管理', '/system/logs', 'file-text', 'LogManagement', 'm3333333-3333-3333-3333-333333333333', 3, false, '["pc","web"]', '["system.monitor"]'),
('m3444444-1111-1111-1111-111111111111', '系统监控', '/system/monitor', 'monitor', 'SystemMonitor', 'm3333333-3333-3333-3333-333333333333', 4, false, '["pc","web"]', '["system.monitor"]'),

-- 插件市场子菜单
('m4111111-1111-1111-1111-111111111111', '插件列表', '/marketplace/plugins', 'appstore', 'PluginList', 'm4444444-4444-4444-4444-444444444444', 1, false, '["pc","web"]', '["plugin.view"]'),
('m4222222-1111-1111-1111-111111111111', '分类管理', '/marketplace/categories', 'tags', 'CategoryManagement', 'm4444444-4444-4444-4444-444444444444', 2, false, '["pc","web"]', '["plugin.manage"]'),
('m4333333-1111-1111-1111-111111111111', '插件审核', '/marketplace/review', 'audit', 'PluginReview', 'm4444444-4444-4444-4444-444444444444', 3, false, '["pc","web"]', '["plugin.review"]'),
('m4444444-1111-1111-1111-111111111111', '开发者中心', '/marketplace/developer', 'code', 'DeveloperCenter', 'm4444444-4444-4444-4444-444444444444', 4, false, '["pc","web"]', '["plugin.manage"]'),

-- 论坛子菜单
('m5111111-1111-1111-1111-111111111111', '帖子列表', '/forum/posts', 'file-text', 'ForumPosts', 'm5555555-5555-5555-5555-555555555555', 1, false, '["pc","web","mobile"]', '["forum.view"]'),
('m5222222-1111-1111-1111-111111111111', '分类管理', '/forum/categories', 'folder', 'ForumCategories', 'm5555555-5555-5555-5555-555555555555', 2, false, '["pc","web"]', '["forum.manage"]'),
('m5333333-1111-1111-1111-111111111111', '内容审核', '/forum/moderation', 'eye', 'ForumModeration', 'm5555555-5555-5555-5555-555555555555', 3, false, '["pc","web"]', '["forum.manage"]'),

-- 博客子菜单
('m6111111-1111-1111-1111-111111111111', '文章列表', '/blog/posts', 'read', 'BlogPosts', 'm6666666-6666-6666-6666-666666666666', 1, false, '["pc","web","mobile"]', '["blog.view"]'),
('m6222222-1111-1111-1111-111111111111', '分类管理', '/blog/categories', 'folder', 'BlogCategories', 'm6666666-6666-6666-6666-666666666666', 2, false, '["pc","web"]', '["blog.manage"]'),
('m6333333-1111-1111-1111-111111111111', '内容编辑', '/blog/editor', 'edit', 'BlogEditor', 'm6666666-6666-6666-6666-666666666666', 3, false, '["pc","web"]', '["content.publish"]');

-- 插入插件分类数据
INSERT INTO mp_categories (id, name, description, icon, color, sort_order, plugin_count) VALUES
('c1111111-1111-1111-1111-111111111111', '开发工具', '开发相关的工具插件，提升开发效率', 'code', '#1890ff', 1, 15),
('c2222222-2222-2222-2222-222222222222', '系统工具', '系统管理和监控工具', 'setting', '#52c41a', 2, 8),
('c3333333-3333-3333-3333-333333333333', '娱乐游戏', '娱乐和游戏插件', 'smile', '#faad14', 3, 5),
('c4444444-4444-4444-4444-444444444444', '效率工具', '提高工作效率的工具', 'rocket', '#722ed1', 4, 12),
('c5555555-5555-5555-5555-555555555555', '数据分析', '数据处理和分析工具', 'bar-chart', '#13c2c2', 5, 6),
('c6666666-6666-6666-6666-666666666666', '安全工具', '安全防护和检测工具', 'safety', '#f5222d', 6, 4),
('c7777777-7777-7777-7777-777777777777', '通信工具', '通信和协作工具', 'message', '#fa8c16', 7, 7),
('c8888888-8888-8888-8888-888888888888', '文档工具', '文档处理和管理工具', 'file-text', '#a0d911', 8, 9);

-- 插入丰富的插件数据
INSERT INTO mp_plugins (id, name, description, short_description, author, developer_id, version, category_id, price, is_free, is_featured, download_count, rating, review_count, tags, license, repository_url) VALUES
('p1111111-1111-1111-1111-111111111111', 'Code Editor Pro', '强大的代码编辑器插件，支持多种编程语言的语法高亮、智能提示、代码格式化等功能。内置Git集成，支持多主题切换。', '专业代码编辑器', 'LaojunTeam', '87654321-4321-4321-4321-210987654321', '2.1.0', 'c1111111-1111-1111-1111-111111111111', 0.00, true, true, 2580, 4.8, 156, '["编辑器","开发","语法高亮","Git"]', 'MIT', 'https://github.com/laojun/code-editor-pro'),
('p2222222-2222-2222-2222-222222222222', 'System Monitor', '系统监控插件，实时监控CPU、内存、磁盘、网络等系统资源使用情况，支持历史数据查看和告警设置。', '系统资源监控', 'LaojunTeam', '87654321-4321-4321-4321-210987654322', '1.5.2', 'c2222222-2222-2222-2222-222222222222', 29.99, false, true, 1890, 4.6, 89, '["监控","系统","性能","告警"]', 'Apache-2.0', 'https://github.com/laojun/system-monitor'),
('p3333333-3333-3333-3333-333333333333', 'Task Manager Pro', '高效的任务管理插件，支持项目管理、时间跟踪、团队协作等功能。集成甘特图和看板视图。', '专业任务管理', 'LaojunTeam', '87654321-4321-4321-4321-210987654321', '3.0.1', 'c4444444-4444-4444-4444-444444444444', 49.99, false, true, 1456, 4.7, 78, '["任务","项目","协作","甘特图"]', 'Commercial', NULL),
('p4444444-4444-4444-4444-444444444444', 'API Tester', 'RESTful API测试工具，支持多种HTTP方法、请求头设置、响应验证等功能。内置环境变量管理。', 'API接口测试', 'DevTools Inc', '87654321-4321-4321-4321-210987654322', '1.8.0', 'c1111111-1111-1111-1111-111111111111', 0.00, true, false, 3240, 4.5, 201, '["API","测试","HTTP","开发"]', 'MIT', 'https://github.com/devtools/api-tester'),
('p5555555-5555-5555-5555-555555555555', 'Database Manager', '数据库管理工具，支持MySQL、PostgreSQL、MongoDB等多种数据库。提供可视化查询界面和数据导入导出功能。', '数据库管理工具', 'DataSoft', '87654321-4321-4321-4321-210987654323', '2.3.0', 'c1111111-1111-1111-1111-111111111111', 39.99, false, true, 1678, 4.4, 92, '["数据库","SQL","管理","可视化"]', 'GPL-3.0', 'https://github.com/datasoft/db-manager'),
('p6666666-6666-6666-6666-666666666666', 'Markdown Editor', '功能丰富的Markdown编辑器，支持实时预览、数学公式、流程图、思维导图等扩展语法。', 'Markdown编辑器', 'WriteWell', '87654321-4321-4321-4321-210987654325', '1.6.5', 'c8888888-8888-8888-8888-888888888888', 0.00, true, false, 2890, 4.6, 167, '["Markdown","编辑器","预览","文档"]', 'MIT', 'https://github.com/writewell/md-editor'),
('p7777777-7777-7777-7777-777777777777', 'Security Scanner', '安全扫描工具，检测常见的安全漏洞，包括SQL注入、XSS、CSRF等。支持自定义规则配置。', '安全漏洞扫描', 'SecureTech', '87654321-4321-4321-4321-210987654324', '1.2.8', 'c6666666-6666-6666-6666-666666666666', 99.99, false, true, 856, 4.3, 45, '["安全","扫描","漏洞","检测"]', 'Commercial', NULL),
('p8888888-8888-8888-8888-888888888888', 'Chat Bot', '智能聊天机器人插件，支持自然语言处理、多轮对话、知识库集成等功能。可用于客服和技术支持。', '智能聊天机器人', 'AI Solutions', '87654321-4321-4321-4321-210987654327', '2.0.3', 'c7777777-7777-7777-7777-777777777777', 79.99, false, false, 1234, 4.2, 67, '["AI","聊天","机器人","客服"]', 'Apache-2.0', 'https://github.com/aisolutions/chatbot'),
('p9999999-9999-9999-9999-999999999999', 'Data Visualizer', '数据可视化工具，支持多种图表类型，包括柱状图、折线图、饼图、散点图等。支持实时数据更新。', '数据可视化工具', 'ChartMaster', '87654321-4321-4321-4321-210987654322', '1.9.2', 'c5555555-5555-5555-5555-555555555555', 59.99, false, true, 1567, 4.5, 89, '["图表","可视化","数据","分析"]', 'MIT', 'https://github.com/chartmaster/data-viz'),
('pa111111-1111-1111-1111-111111111111', 'File Organizer', '文件整理工具，自动按照规则整理文件夹，支持批量重命名、重复文件检测、文件分类等功能。', '智能文件整理', 'FileUtils', '87654321-4321-4321-4321-210987654326', '1.4.1', 'c4444444-4444-4444-4444-444444444444', 0.00, true, false, 2145, 4.1, 123, '["文件","整理","批量","工具"]', 'GPL-2.0', 'https://github.com/fileutils/organizer'),
('pb222222-2222-2222-2222-222222222222', 'Password Manager', '密码管理器，安全存储和管理密码，支持密码生成、自动填充、多设备同步等功能。', '安全密码管理', 'SecureVault', '87654321-4321-4321-4321-210987654324', '2.1.5', 'c6666666-6666-6666-6666-666666666666', 29.99, false, false, 987, 4.7, 56, '["密码","安全","管理","加密"]', 'Commercial', NULL),
('pc333333-3333-3333-3333-333333333333', 'Mini Games', '小游戏合集，包含俄罗斯方块、贪吃蛇、2048等经典小游戏，支持排行榜和成就系统。', '经典小游戏合集', 'GameStudio', '87654321-4321-4321-4321-210987654326', '1.0.8', 'c3333333-3333-3333-3333-333333333333', 0.00, true, false, 3456, 4.0, 234, '["游戏","娱乐","休闲","经典"]', 'MIT', 'https://github.com/gamestudio/mini-games');

-- 插入论坛帖子数据
INSERT INTO mp_forum_posts (id, title, content, category_id, user_id, is_pinned, is_featured, view_count, reply_count, like_count, tags) VALUES
('fp111111-1111-1111-1111-111111111111', 'Go语言微服务架构最佳实践', '在这篇文章中，我将分享在使用Go语言构建微服务架构时的一些最佳实践和经验总结...', 'f1111111-1111-1111-1111-111111111111', '87654321-4321-4321-4321-210987654321', true, true, 1250, 23, 89, '["Go","微服务","架构","最佳实践"]'),
('fp222222-2222-2222-2222-222222222222', '如何开发一个高质量的插件', '插件开发是一门艺术，需要考虑用户体验、性能优化、兼容性等多个方面...', 'f2222222-2222-2222-2222-222222222222', '87654321-4321-4321-4321-210987654322', false, true, 890, 15, 67, '["插件开发","质量","用户体验"]'),
('fp333333-3333-3333-3333-333333333333', '系统登录时出现500错误', '我在使用系统时遇到了登录问题，每次输入用户名密码后都会出现500错误...', 'f3333333-3333-3333-3333-333333333333', '87654321-4321-4321-4321-210987654324', false, false, 234, 8, 12, '["bug","登录","500错误"]'),
('fp444444-4444-4444-4444-444444444444', '新手入门指南：如何快速上手太上老君系统', '作为一个新手，刚开始使用太上老君系统时可能会感到困惑，这里我整理了一份详细的入门指南...', 'f4444444-4444-4444-4444-444444444444', '87654321-4321-4321-4321-210987654325', true, false, 567, 12, 45, '["新手","入门","指南","教程"]'),
('fp555555-5555-5555-5555-555555555555', '我的开源项目：基于Vue3的管理后台模板', '经过几个月的开发，我的开源项目终于完成了第一个版本，这是一个基于Vue3的现代化管理后台模板...', 'f5555555-5555-5555-5555-555555555555', '87654321-4321-4321-4321-210987654323', false, true, 789, 18, 78, '["Vue3","开源","后台","模板"]'),
('fp666666-6666-6666-6666-666666666666', '招聘：高级Go开发工程师', '我们公司正在寻找有经验的Go开发工程师，主要负责微服务架构的设计和开发...', 'f6666666-6666-6666-6666-666666666666', '87654321-4321-4321-4321-210987654327', false, false, 345, 5, 23, '["招聘","Go","开发","微服务"]');

-- 插入博客文章数据
INSERT INTO mp_blog_posts (id, title, content, excerpt, category_id, author_id, is_published, is_featured, view_count, like_count, tags, reading_time, published_at) VALUES
('bp111111-1111-1111-1111-111111111111', '深入理解Go语言的并发模型', '# 深入理解Go语言的并发模型\n\nGo语言的并发模型是其最大的特色之一，通过goroutine和channel，Go提供了一种简洁而强大的并发编程方式...', 'Go语言的并发模型是其最大的特色之一，本文将深入探讨goroutine和channel的工作原理。', 'b1111111-1111-1111-1111-111111111111', '87654321-4321-4321-4321-210987654321', true, true, 2340, 156, '["Go","并发","goroutine","channel"]', 15, '2025-01-20 10:00:00+08'),
('bp222222-2222-2222-2222-222222222222', '项目重构日志：从单体到微服务', '# 项目重构日志：从单体到微服务\n\n在过去的六个月里，我们团队完成了一次大规模的架构重构，将原有的单体应用拆分为微服务架构...', '记录了一次从单体应用到微服务架构的完整重构过程，包括遇到的问题和解决方案。', 'b2222222-2222-2222-2222-222222222222', '87654321-4321-4321-4321-210987654322', true, true, 1890, 123, '["重构","微服务","架构","经验"]', 20, '2025-01-18 14:30:00+08'),
('bp333333-3333-3333-3333-333333333333', 'Docker容器化部署完全指南', '# Docker容器化部署完全指南\n\n容器化技术已经成为现代应用部署的标准方式，本文将详细介绍如何使用Docker进行应用的容器化部署...', '详细介绍Docker容器化部署的完整流程，从基础概念到实际应用。', 'b3333333-3333-3333-3333-333333333333', '87654321-4321-4321-4321-210987654325', true, false, 1567, 89, '["Docker","容器","部署","教程"]', 25, '2025-01-15 09:15:00+08'),
('bp444444-4444-4444-4444-444444444444', '2025年前端技术趋势预测', '# 2025年前端技术趋势预测\n\n随着技术的不断发展，前端领域也在快速演进。本文将分析2025年前端技术的发展趋势...', '分析2025年前端技术的发展趋势，包括新兴框架、工具和最佳实践。', 'b4444444-4444-4444-4444-444444444444', '87654321-4321-4321-4321-210987654323', true, true, 2100, 178, '["前端","趋势","2025","技术"]', 12, '2025-01-22 16:45:00+08'),
('bp555555-5555-5555-5555-555555555555', '推荐10个提升开发效率的VS Code插件', '# 推荐10个提升开发效率的VS Code插件\n\nVS Code作为最受欢迎的代码编辑器之一，拥有丰富的插件生态。本文推荐10个能显著提升开发效率的插件...', '精选10个VS Code插件，帮助开发者提升编码效率和开发体验。', 'b5555555-5555-5555-5555-555555555555', '87654321-4321-4321-4321-210987654325', true, false, 1345, 67, '["VS Code","插件","效率","工具"]', 8, '2025-01-19 11:20:00+08');

-- 插入插件评论数据
INSERT INTO mp_plugin_reviews (id, plugin_id, user_id, rating, title, content, is_verified_purchase) VALUES
('pr111111-1111-1111-1111-111111111111', 'p1111111-1111-1111-1111-111111111111', '87654321-4321-4321-4321-210987654322', 5, '非常棒的代码编辑器', '这个插件真的很棒！语法高亮很准确，智能提示也很实用。Git集成功能让我的工作效率提升了很多。', true),
('pr222222-2222-2222-2222-222222222222', 'p1111111-1111-1111-1111-111111111111', '87654321-4321-4321-4321-210987654323', 4, '功能丰富，值得推荐', '功能确实很丰富，但是有时候会有点卡顿。希望后续版本能优化一下性能。', true),
('pr333333-3333-3333-3333-333333333333', 'p2222222-2222-2222-2222-222222222222', '87654321-4321-4321-4321-210987654324', 5, '系统监控必备工具', '作为运维人员，这个工具对我来说非常有用。界面清晰，数据准确，告警功能也很实用。', true),
('pr444444-4444-4444-4444-444444444444', 'p3333333-3333-3333-3333-333333333333', '87654321-4321-4321-4321-210987654325', 4, '项目管理好帮手', '任务管理功能很强大，甘特图视图特别有用。不过价格有点贵，希望能有更多的免费功能。', true),
('pr555555-5555-5555-5555-555555555555', 'p4444444-4444-4444-4444-444444444444', '87654321-4321-4321-4321-210987654326', 5, '免费的API测试神器', '完全免费却功能强大，比很多付费工具都好用。环境变量管理功能特别方便。', false);

-- 插入系统配置数据
INSERT INTO sm_system_configs (id, config_key, config_value, config_type, description, is_public) VALUES
('sc111111-1111-1111-1111-111111111111', 'site.name', '太上老君系统', 'string', '网站名称', true),
('sc222222-2222-2222-2222-222222222222', 'site.description', '基于Go和Vue.js的现代化管理系统', 'string', '网站描述', true),
('sc333333-3333-3333-3333-333333333333', 'site.logo', '/assets/logo.png', 'string', '网站Logo', true),
('sc444444-4444-4444-4444-444444444444', 'site.favicon', '/assets/favicon.ico', 'string', '网站图标', true),
('sc555555-5555-5555-5555-555555555555', 'site.keywords', '管理系统,插件市场,论坛,博客', 'string', '网站关键词', true),
('sc666666-6666-6666-6666-666666666666', 'system.version', '2.0.0', 'string', '系统版本', true),
('sc777777-7777-7777-7777-777777777777', 'system.timezone', 'Asia/Shanghai', 'string', '系统时区', false),
('sc888888-8888-8888-8888-888888888888', 'system.language', 'zh-CN', 'string', '系统语言', false),
('sc999999-9999-9999-9999-999999999999', 'email.smtp_host', 'smtp.example.com', 'string', 'SMTP服务器', false),
('scaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'email.smtp_port', '587', 'number', 'SMTP端口', false),
('scbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'email.from_address', 'noreply@laojun.com', 'string', '发件人邮箱', false),
('sccccccc-cccc-cccc-cccc-cccccccccccc', 'upload.max_file_size', '10485760', 'number', '最大文件上传大小(字节)', false),
('scdddddd-dddd-dddd-dddd-dddddddddddd', 'upload.allowed_types', 'jpg,jpeg,png,gif,pdf,doc,docx,zip', 'string', '允许上传的文件类型', false),
('sceeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'security.session_timeout', '3600', 'number', '会话超时时间(秒)', false),
('scffffff-ffff-ffff-ffff-ffffffffffff', 'security.password_min_length', '8', 'number', '密码最小长度', false),
('sc111111-2222-2222-2222-222222222222', 'forum.posts_per_page', '20', 'number', '论坛每页帖子数', true),
('sc222222-3333-3333-3333-333333333333', 'blog.posts_per_page', '10', 'number', '博客每页文章数', true),
('sc333333-4444-4444-4444-444444444444', 'marketplace.featured_count', '6', 'number', '首页推荐插件数量', true),
('sc444444-5555-5555-5555-555555555555', 'analytics.enable_tracking', 'true', 'boolean', '启用数据统计', false),
('sc555555-6666-6666-6666-666666666666', 'maintenance.mode', 'false', 'boolean', '维护模式', false);

-- =============================================================================
-- 创建索引以提高查询性能
-- =============================================================================

-- 权限和角色相关索引
CREATE INDEX IF NOT EXISTS idx_az_role_permissions_role_id ON az_role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_az_role_permissions_permission_id ON az_role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_role_id ON az_user_roles(role_id);

-- 用户相关索引
CREATE INDEX IF NOT EXISTS idx_ua_admin_username ON ua_admin(username);
CREATE INDEX IF NOT EXISTS idx_ua_admin_email ON ua_admin(email);
CREATE INDEX IF NOT EXISTS idx_ua_admin_is_active ON ua_admin(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_users_username ON mp_users(username);
CREATE INDEX IF NOT EXISTS idx_mp_users_email ON mp_users(email);
CREATE INDEX IF NOT EXISTS idx_mp_users_is_active ON mp_users(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_users_is_verified ON mp_users(is_verified);

-- 会话相关索引
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_user_id ON ua_user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_session_token ON ua_user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_expires_at ON ua_user_sessions(expires_at);

-- 菜单相关索引
CREATE INDEX IF NOT EXISTS idx_sm_menus_parent_id ON sm_menus(parent_id);
CREATE INDEX IF NOT EXISTS idx_sm_menus_sort_order ON sm_menus(sort_order);
CREATE INDEX IF NOT EXISTS idx_sm_menus_is_hidden ON sm_menus(is_hidden);

-- 论坛相关索引
CREATE INDEX IF NOT EXISTS idx_mp_forum_categories_sort_order ON mp_forum_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_mp_forum_categories_is_active ON mp_forum_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_category_id ON mp_forum_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_user_id ON mp_forum_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_is_pinned ON mp_forum_posts(is_pinned);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_is_featured ON mp_forum_posts(is_featured);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_created_at ON mp_forum_posts(created_at);
CREATE INDEX IF NOT EXISTS idx_mp_forum_replies_post_id ON mp_forum_replies(post_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_replies_user_id ON mp_forum_replies(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_replies_parent_id ON mp_forum_replies(parent_id);

-- 博客相关索引
CREATE INDEX IF NOT EXISTS idx_mp_blog_categories_sort_order ON mp_blog_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_mp_blog_categories_is_active ON mp_blog_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_blog_posts_category_id ON mp_blog_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_blog_posts_author_id ON mp_blog_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_mp_blog_posts_is_published ON mp_blog_posts(is_published);
CREATE INDEX IF NOT EXISTS idx_mp_blog_posts_is_featured ON mp_blog_posts(is_featured);
CREATE INDEX IF NOT EXISTS idx_mp_blog_posts_published_at ON mp_blog_posts(published_at);

-- 插件相关索引
CREATE INDEX IF NOT EXISTS idx_mp_categories_sort_order ON mp_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_mp_categories_is_active ON mp_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_category_id ON mp_plugins(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_developer_id ON mp_plugins(developer_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_is_active ON mp_plugins(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_is_featured ON mp_plugins(is_featured);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_is_free ON mp_plugins(is_free);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_rating ON mp_plugins(rating);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_download_count ON mp_plugins(download_count);
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_plugin_id ON mp_plugin_reviews(plugin_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_user_id ON mp_plugin_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_rating ON mp_plugin_reviews(rating);

-- 系统配置索引
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_config_key ON sm_system_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_is_public ON sm_system_configs(is_public);

-- JWT密钥索引
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_key_id ON ua_jwt_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_is_active ON ua_jwt_keys(is_active);

-- =============================================================================
-- 创建视图以简化常用查询
-- =============================================================================

-- 用户权限视图
CREATE OR REPLACE VIEW v_user_permissions AS
SELECT 
    u.id as user_id,
    u.username,
    r.name as role_name,
    p.permission_code,
    p.name as permission_name,
    p.resource,
    p.action
FROM ua_admin u
JOIN az_user_roles ur ON u.id = ur.user_id
JOIN az_roles r ON ur.role_id = r.id
JOIN az_role_permissions rp ON r.id = rp.role_id
JOIN az_permissions p ON rp.permission_id = p.id
WHERE u.is_active = true;

-- 插件统计视图
CREATE OR REPLACE VIEW v_plugin_stats AS
SELECT 
    c.id as category_id,
    c.name as category_name,
    COUNT(p.id) as plugin_count,
    AVG(p.rating) as avg_rating,
    SUM(p.download_count) as total_downloads
FROM mp_categories c
LEFT JOIN mp_plugins p ON c.id = p.category_id AND p.is_active = true
GROUP BY c.id, c.name;

-- 论坛统计视图
CREATE OR REPLACE VIEW v_forum_stats AS
SELECT 
    c.id as category_id,
    c.name as category_name,
    COUNT(p.id) as post_count,
    SUM(p.views_count) as total_views,
    SUM(p.replies_count) as total_replies
FROM mp_forum_categories c
LEFT JOIN mp_forum_posts p ON c.id = p.category_id
GROUP BY c.id, c.name;

-- =============================================================================
-- 创建触发器以维护数据一致性
-- =============================================================================

-- 更新论坛分类帖子数量的触发器函数
CREATE OR REPLACE FUNCTION update_forum_category_post_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE mp_forum_categories 
        SET post_count = post_count + 1 
        WHERE id = NEW.category_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE mp_forum_categories 
        SET post_count = post_count - 1 
        WHERE id = OLD.category_id;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.category_id != NEW.category_id THEN
            UPDATE mp_forum_categories 
            SET post_count = post_count - 1 
            WHERE id = OLD.category_id;
            UPDATE mp_forum_categories 
            SET post_count = post_count + 1 
            WHERE id = NEW.category_id;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_update_forum_category_post_count ON mp_forum_posts;
CREATE TRIGGER trigger_update_forum_category_post_count
    AFTER INSERT OR UPDATE OR DELETE ON mp_forum_posts
    FOR EACH ROW EXECUTE FUNCTION update_forum_category_post_count();

-- 更新博客分类文章数量的触发器函数
CREATE OR REPLACE FUNCTION update_blog_category_post_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.is_published = true THEN
            UPDATE mp_blog_categories 
            SET post_count = post_count + 1 
            WHERE id = NEW.category_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.is_published = true THEN
            UPDATE mp_blog_categories 
            SET post_count = post_count - 1 
            WHERE id = OLD.category_id;
        END IF;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.is_published != NEW.is_published THEN
            IF NEW.is_published = true THEN
                UPDATE mp_blog_categories 
                SET post_count = post_count + 1 
                WHERE id = NEW.category_id;
            ELSE
                UPDATE mp_blog_categories 
                SET post_count = post_count - 1 
                WHERE id = OLD.category_id;
            END IF;
        ELSIF OLD.category_id != NEW.category_id AND NEW.is_published = true THEN
            UPDATE mp_blog_categories 
            SET post_count = post_count - 1 
            WHERE id = OLD.category_id;
            UPDATE mp_blog_categories 
            SET post_count = post_count + 1 
            WHERE id = NEW.category_id;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_update_blog_category_post_count ON mp_blog_posts;
CREATE TRIGGER trigger_update_blog_category_post_count
    AFTER INSERT OR UPDATE OR DELETE ON mp_blog_posts
    FOR EACH ROW EXECUTE FUNCTION update_blog_category_post_count();

-- 更新插件分类插件数量的触发器函数
CREATE OR REPLACE FUNCTION update_plugin_category_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.is_active = true THEN
            UPDATE mp_categories 
            SET plugin_count = plugin_count + 1 
            WHERE id = NEW.category_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.is_active = true THEN
            UPDATE mp_categories 
            SET plugin_count = plugin_count - 1 
            WHERE id = OLD.category_id;
        END IF;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.is_active != NEW.is_active THEN
            IF NEW.is_active = true THEN
                UPDATE mp_categories 
                SET plugin_count = plugin_count + 1 
                WHERE id = NEW.category_id;
            ELSE
                UPDATE mp_categories 
                SET plugin_count = plugin_count - 1 
                WHERE id = OLD.category_id;
            END IF;
        ELSIF OLD.category_id != NEW.category_id AND NEW.is_active = true THEN
            UPDATE mp_categories 
            SET plugin_count = plugin_count - 1 
            WHERE id = OLD.category_id;
            UPDATE mp_categories 
            SET plugin_count = plugin_count + 1 
            WHERE id = NEW.category_id;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_update_plugin_category_count ON mp_plugins;
CREATE TRIGGER trigger_update_plugin_category_count
    AFTER INSERT OR UPDATE OR DELETE ON mp_plugins
    FOR EACH ROW EXECUTE FUNCTION update_plugin_category_count();

-- =============================================================================
-- 数据完整性检查和统计信息
-- =============================================================================

-- 显示部署统计信息
DO $$
DECLARE
    permission_count INTEGER;
    role_count INTEGER;
    admin_count INTEGER;
    user_count INTEGER;
    category_count INTEGER;
    plugin_count INTEGER;
    forum_category_count INTEGER;
    forum_post_count INTEGER;
    blog_category_count INTEGER;
    blog_post_count INTEGER;
    menu_count INTEGER;
    config_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO permission_count FROM az_permissions;
    SELECT COUNT(*) INTO role_count FROM az_roles;
    SELECT COUNT(*) INTO admin_count FROM ua_admin;
    SELECT COUNT(*) INTO user_count FROM mp_users;
    SELECT COUNT(*) INTO category_count FROM mp_categories;
    SELECT COUNT(*) INTO plugin_count FROM mp_plugins;
    SELECT COUNT(*) INTO forum_category_count FROM mp_forum_categories;
    SELECT COUNT(*) INTO forum_post_count FROM mp_forum_posts;
    SELECT COUNT(*) INTO blog_category_count FROM mp_blog_categories;
    SELECT COUNT(*) INTO blog_post_count FROM mp_blog_posts;
    SELECT COUNT(*) INTO menu_count FROM sm_menus;
    SELECT COUNT(*) INTO config_count FROM sm_system_configs;
    
    RAISE NOTICE '=== 太上老君系统数据库部署完成 ===';
    RAISE NOTICE '权限数量: %', permission_count;
    RAISE NOTICE '角色数量: %', role_count;
    RAISE NOTICE '管理员数量: %', admin_count;
    RAISE NOTICE '市场用户数量: %', user_count;
    RAISE NOTICE '插件分类数量: %', category_count;
    RAISE NOTICE '插件数量: %', plugin_count;
    RAISE NOTICE '论坛分类数量: %', forum_category_count;
    RAISE NOTICE '论坛帖子数量: %', forum_post_count;
    RAISE NOTICE '博客分类数量: %', blog_category_count;
    RAISE NOTICE '博客文章数量: %', blog_post_count;
    RAISE NOTICE '菜单数量: %', menu_count;
    RAISE NOTICE '系统配置数量: %', config_count;
    RAISE NOTICE '=== 部署成功！===';
END $$;

-- 完成部署
SELECT 'Database deployment completed successfully! 🎉' as status;