-- 创建更新时间戳的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为菜单配置表添加更新时间戳触发器
CREATE TRIGGER update_sm_menu_configs_updated_at 
    BEFORE UPDATE ON sm_menu_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();