-- 综合优化迁移文件
-- 基于 complete_deployment.sql 的所有优化

-- 创建系统配置表
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

-- 创建博客分类表
CREATE TABLE IF NOT EXISTS mp_blog_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(7),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表添加更新时间触发器
DROP TRIGGER IF EXISTS update_mp_users_updated_at ON mp_users;
CREATE TRIGGER update_mp_users_updated_at
    BEFORE UPDATE ON mp_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_mp_forum_categories_updated_at ON mp_forum_categories;
CREATE TRIGGER update_mp_forum_categories_updated_at
    BEFORE UPDATE ON mp_forum_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_mp_blog_categories_updated_at ON mp_blog_categories;
CREATE TRIGGER update_mp_blog_categories_updated_at
    BEFORE UPDATE ON mp_blog_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_sm_system_configs_updated_at ON sm_system_configs;
CREATE TRIGGER update_sm_system_configs_updated_at
    BEFORE UPDATE ON sm_system_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ua_admin_updated_at ON ua_admin;
CREATE TRIGGER update_ua_admin_updated_at
    BEFORE UPDATE ON ua_admin
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建性能优化索引
-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_mp_users_username_active ON mp_users(username, is_active);
CREATE INDEX IF NOT EXISTS idx_mp_users_email_verified ON mp_users(email, is_verified);
CREATE INDEX IF NOT EXISTS idx_mp_users_created_at ON mp_users(created_at);
CREATE INDEX IF NOT EXISTS idx_mp_users_last_login ON mp_users(last_login_at);

-- 论坛帖子索引
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_category_created ON mp_forum_posts(category_id, created_at);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_user_created ON mp_forum_posts(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_likes_count ON mp_forum_posts(likes_count);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_views_count ON mp_forum_posts(views_count);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_replies_count ON mp_forum_posts(replies_count);

-- 管理员用户索引
CREATE INDEX IF NOT EXISTS idx_ua_admin_username_active ON ua_admin(username, is_active);
CREATE INDEX IF NOT EXISTS idx_ua_admin_email_active ON ua_admin(email, is_active);
CREATE INDEX IF NOT EXISTS idx_ua_admin_is_super_admin ON ua_admin(is_super_admin);
CREATE INDEX IF NOT EXISTS idx_ua_admin_last_login ON ua_admin(last_login_at);

-- 权限相关索引
CREATE INDEX IF NOT EXISTS idx_az_permissions_resource_action ON az_permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_user_active ON az_user_roles(user_id) WHERE expires_at IS NULL OR expires_at > NOW();
CREATE INDEX IF NOT EXISTS idx_az_role_permissions_role_id ON az_role_permissions(role_id);

-- 系统配置索引
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_key ON sm_system_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_type ON sm_system_configs(config_type);
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_public ON sm_system_configs(is_public);

-- 博客分类索引
CREATE INDEX IF NOT EXISTS idx_mp_blog_categories_sort_order ON mp_blog_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_mp_blog_categories_name ON mp_blog_categories(name);

-- 创建有用的视图
-- 用户权限视图
CREATE OR REPLACE VIEW v_user_permissions AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    r.name as role_name,
    p.code as permission_code,
    p.resource,
    p.action
FROM ua_admin u
JOIN az_user_roles ur ON u.id = ur.user_id
JOIN az_roles r ON ur.role_id = r.id
JOIN az_role_permissions rp ON r.id = rp.role_id
JOIN az_permissions p ON rp.permission_id = p.id
WHERE u.is_active = TRUE 
  AND r.is_active = TRUE
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

-- 插入系统配置数据
INSERT INTO sm_system_configs (config_key, config_value, config_type, description, is_public) VALUES
('site_name', 'Laojun Platform', 'string', '网站名称', TRUE),
('site_description', '老君平台 - 插件市场与开发者社区', 'string', '网站描述', TRUE),
('max_upload_size', '10485760', 'integer', '最大上传文件大小（字节）', FALSE)
ON CONFLICT (config_key) DO NOTHING;

-- 插入博客分类数据
INSERT INTO mp_blog_categories (name, slug, description, sort_order) VALUES
('技术分享', 'tech-sharing', '技术相关的文章和教程', 1),
('产品动态', 'product-updates', '产品更新和新功能介绍', 2),
('行业资讯', 'industry-news', '行业动态和趋势分析', 3),
('开发教程', 'dev-tutorials', '开发相关的教程和指南', 4),
('社区活动', 'community-events', '社区活动和公告', 5)
ON CONFLICT (slug) DO NOTHING;

-- 插入迁移记录
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('009_comprehensive_optimization', FALSE) 
ON CONFLICT (version) DO NOTHING;

-- 输出确认信息
\echo 'Comprehensive optimization migration completed successfully';