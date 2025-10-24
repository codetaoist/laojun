-- 综合优化迁移文件 - 仅包含新增功能和优化
-- 移除了与其他迁移文件重复的索引，避免冲突
-- 版本: 2024年优化版本

-- 启用高级数据库扩展
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- =====================================================
-- 新增表结构
-- =====================================================

-- 博客分类表 (Blog Categories)
CREATE TABLE IF NOT EXISTS mp_blog_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#007bff',
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 系统配置表 (System Configuration)
CREATE TABLE IF NOT EXISTS sm_system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- =====================================================
-- 新增性能优化索引 (仅包含新表和高级优化索引)
-- =====================================================

-- 博客分类索引
CREATE INDEX IF NOT EXISTS idx_mp_blog_categories_sort_order ON mp_blog_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_mp_blog_categories_name ON mp_blog_categories(name);

-- 系统配置索引
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_key ON sm_system_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_type ON sm_system_configs(config_type);
CREATE INDEX IF NOT EXISTS idx_sm_system_configs_public ON sm_system_configs(is_public) WHERE is_public = true;

-- 高级复合索引优化 (不与现有索引重复)
CREATE INDEX IF NOT EXISTS idx_mp_users_email_verified_active ON mp_users(email, is_email_verified) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_category_likes ON mp_forum_posts(category_id, likes_count DESC);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_featured_rating ON mp_plugins(is_featured, rating DESC) WHERE is_featured = true;
CREATE INDEX IF NOT EXISTS idx_az_user_roles_expires ON az_user_roles(user_id, expires_at) WHERE expires_at IS NOT NULL;

-- =====================================================
-- 数据库视图 (Database Views)
-- =====================================================

-- 用户权限视图
CREATE OR REPLACE VIEW v_user_permissions AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    r.name as role_name,
    p.permission_code,
    p.name as permission_name,
    p.resource,
    p.action,
    ur.granted_at,
    ur.expires_at
FROM ua_admin u
JOIN az_user_roles ur ON u.id = ur.user_id
JOIN az_roles r ON ur.role_id = r.id
JOIN az_role_permissions rp ON r.id = rp.role_id
JOIN az_permissions p ON rp.permission_id = p.id
WHERE u.is_active = true 
  AND r.is_active = true
  AND (ur.expires_at IS NULL OR ur.expires_at > NOW());

-- 插件统计视图
CREATE OR REPLACE VIEW v_plugin_stats AS
SELECT 
    p.id,
    p.name,
    p.version,
    p.downloads_count,
    p.rating,
    COUNT(pr.id) as review_count,
    AVG(pr.rating) as avg_rating,
    p.created_at,
    p.updated_at
FROM mp_plugins p
LEFT JOIN mp_plugin_reviews pr ON p.id = pr.plugin_id
GROUP BY p.id, p.name, p.version, p.downloads_count, p.rating, p.created_at, p.updated_at;

-- 论坛统计视图
CREATE OR REPLACE VIEW v_forum_stats AS
SELECT 
    p.id,
    p.title,
    p.likes_count,
    p.replies_count,
    p.views_count,
    p.created_at,
    u.username as author_name,
    c.name as category_name
FROM mp_forum_posts p
JOIN mp_users u ON p.user_id = u.id
JOIN mp_forum_categories c ON p.category_id = c.id;

-- =====================================================
-- 触发器 (Triggers)
-- =====================================================

-- 更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表添加更新时间戳触发器
DROP TRIGGER IF EXISTS update_mp_users_updated_at ON mp_users;
CREATE TRIGGER update_mp_users_updated_at 
    BEFORE UPDATE ON mp_users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_mp_forum_categories_updated_at ON mp_forum_categories;
CREATE TRIGGER update_mp_forum_categories_updated_at 
    BEFORE UPDATE ON mp_forum_categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_mp_blog_categories_updated_at ON mp_blog_categories;
CREATE TRIGGER update_mp_blog_categories_updated_at 
    BEFORE UPDATE ON mp_blog_categories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_sm_system_configs_updated_at ON sm_system_configs;
CREATE TRIGGER update_sm_system_configs_updated_at 
    BEFORE UPDATE ON sm_system_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ua_admin_updated_at ON ua_admin;
CREATE TRIGGER update_ua_admin_updated_at 
    BEFORE UPDATE ON ua_admin 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 插入测试数据
-- =====================================================

-- 博客分类测试数据
INSERT INTO mp_blog_categories (name, description, color, sort_order) VALUES
('技术分享', '分享各种技术心得和经验', '#007bff', 1),
('产品动态', '产品更新和功能介绍', '#28a745', 2),
('行业资讯', '行业最新动态和趋势', '#ffc107', 3),
('开发教程', '详细的开发教程和指南', '#dc3545', 4),
('社区活动', '社区举办的各种活动', '#6f42c1', 5)
ON CONFLICT (name) DO NOTHING;

-- 系统配置测试数据
INSERT INTO sm_system_configs (config_key, config_value, config_type, description, is_public) VALUES
('site_name', '太上老君插件市场', 'string', '网站名称', true),
('site_description', '专业的插件开发与分享平台', 'string', '网站描述', true),
('max_upload_size', '10485760', 'integer', '最大上传文件大小(字节)', false)
ON CONFLICT (config_key) DO NOTHING;

-- 记录迁移版本
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('009_comprehensive_optimization', FALSE) 
ON CONFLICT (version) DO NOTHING;

-- 输出确认信息
\echo 'Comprehensive optimization migration completed successfully';