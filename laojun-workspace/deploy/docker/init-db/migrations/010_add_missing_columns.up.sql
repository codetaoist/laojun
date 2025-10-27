-- 添加缺失的数据库列
-- 修复API错误：review_priority、data_type、level等列

-- =====================================================
-- 添加插件审核优先级列
-- =====================================================

-- 为mp_plugins表添加review_priority列
ALTER TABLE mp_plugins 
ADD COLUMN IF NOT EXISTS review_priority VARCHAR(20) DEFAULT 'normal';

-- 为review_priority列添加检查约束
ALTER TABLE mp_plugins 
ADD CONSTRAINT chk_review_priority 
CHECK (review_priority IN ('low', 'normal', 'high', 'urgent'));

-- 为review_priority列添加索引
CREATE INDEX IF NOT EXISTS idx_mp_plugins_review_priority ON mp_plugins(review_priority);

-- =====================================================
-- 修复系统设置表
-- =====================================================

-- 为sys_settings表添加data_type列（如果不存在）
ALTER TABLE sys_settings 
ADD COLUMN IF NOT EXISTS data_type VARCHAR(20) DEFAULT 'string';

-- 更新现有记录的data_type
UPDATE sys_settings SET data_type = type WHERE data_type IS NULL OR data_type = '';

-- 为data_type列添加检查约束
ALTER TABLE sys_settings 
ADD CONSTRAINT chk_sys_settings_data_type 
CHECK (data_type IN ('string', 'integer', 'boolean', 'json', 'text'));

-- 为data_type列添加索引
CREATE INDEX IF NOT EXISTS idx_sys_settings_data_type ON sys_settings(data_type);

-- =====================================================
-- 修复审计日志表
-- =====================================================

-- 为sys_audit_logs表添加level列
ALTER TABLE sys_audit_logs 
ADD COLUMN IF NOT EXISTS level VARCHAR(20) DEFAULT 'info';

-- 为level列添加检查约束
ALTER TABLE sys_audit_logs 
ADD CONSTRAINT chk_audit_logs_level 
CHECK (level IN ('debug', 'info', 'warn', 'error', 'fatal'));

-- 为level列添加索引
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_level ON sys_audit_logs(level);

-- =====================================================
-- 添加菜单配置表（修复menu-configs/visual API）
-- =====================================================

-- 创建菜单配置表
CREATE TABLE IF NOT EXISTS sm_menu_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_id UUID REFERENCES sm_menus(id) ON DELETE CASCADE,
    config_type VARCHAR(50) NOT NULL DEFAULT 'visual',
    config_data JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 为菜单配置表添加索引
CREATE INDEX IF NOT EXISTS idx_sm_menu_configs_menu_id ON sm_menu_configs(menu_id);
CREATE INDEX IF NOT EXISTS idx_sm_menu_configs_type ON sm_menu_configs(config_type);
CREATE INDEX IF NOT EXISTS idx_sm_menu_configs_active ON sm_menu_configs(is_active) WHERE is_active = true;

-- =====================================================
-- 插入初始数据
-- =====================================================

-- 更新现有插件的review_priority
UPDATE mp_plugins 
SET review_priority = 'normal' 
WHERE review_priority IS NULL;

-- 插入默认菜单配置数据
INSERT INTO sm_menu_configs (config_type, config_data, is_active) VALUES
('visual', '{"theme": "default", "layout": "sidebar", "colors": {"primary": "#1890ff", "secondary": "#f0f0f0"}}', true),
('navigation', '{"showBreadcrumb": true, "showSidebar": true, "collapsible": true}', true),
('permissions', '{"defaultRole": "user", "adminRequired": false}', true)
ON CONFLICT DO NOTHING;

-- =====================================================
-- 更新时间戳触发器
-- =====================================================

-- 为菜单配置表添加更新时间戳触发器
DROP TRIGGER IF EXISTS update_sm_menu_configs_updated_at ON sm_menu_configs;
CREATE TRIGGER update_sm_menu_configs_updated_at 
    BEFORE UPDATE ON sm_menu_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 记录迁移版本
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('010_add_missing_columns', FALSE) 
ON CONFLICT (version) DO NOTHING;