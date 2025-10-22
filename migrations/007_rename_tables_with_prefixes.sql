-- Description: Rename tables with appropriate prefixes for different modules
-- +migrate Up

BEGIN;

-- ========================================
-- 重命名后台管理相关表 (ua_ 前缀)
-- ========================================

-- 重命名用户相关表
ALTER TABLE IF EXISTS users RENAME TO ua_users;
ALTER TABLE IF EXISTS configs RENAME TO ua_configs;
ALTER TABLE IF EXISTS sessions RENAME TO ua_sessions;
ALTER TABLE IF EXISTS session_activities RENAME TO ua_session_activities;
ALTER TABLE IF EXISTS audit_logs RENAME TO ua_audit_logs;
ALTER TABLE IF EXISTS security_events RENAME TO ua_security_events;
ALTER TABLE IF EXISTS system_settings RENAME TO ua_system_settings;
ALTER TABLE IF EXISTS system_settings_history RENAME TO ua_system_settings_history;

-- 更新索引名称 (后台管理)
-- users表索引
ALTER INDEX IF EXISTS idx_users_email RENAME TO idx_ua_users_email;
ALTER INDEX IF EXISTS idx_users_role RENAME TO idx_ua_users_role;
ALTER INDEX IF EXISTS idx_users_status RENAME TO idx_ua_users_status;
ALTER INDEX IF EXISTS idx_users_created_at RENAME TO idx_ua_users_created_at;
ALTER INDEX IF EXISTS idx_users_last_login_at RENAME TO idx_ua_users_last_login_at;
ALTER INDEX IF EXISTS idx_users_email_verified RENAME TO idx_ua_users_email_verified;

-- configs表索引
ALTER INDEX IF EXISTS idx_configs_key RENAME TO idx_ua_configs_key;
ALTER INDEX IF EXISTS idx_configs_namespace RENAME TO idx_ua_configs_namespace;
ALTER INDEX IF EXISTS idx_configs_environment RENAME TO idx_ua_configs_environment;
ALTER INDEX IF EXISTS idx_configs_type RENAME TO idx_ua_configs_type;
ALTER INDEX IF EXISTS idx_configs_is_public RENAME TO idx_ua_configs_is_public;
ALTER INDEX IF EXISTS idx_configs_created_by RENAME TO idx_ua_configs_created_by;
ALTER INDEX IF EXISTS idx_configs_created_at RENAME TO idx_ua_configs_created_at;
ALTER INDEX IF EXISTS idx_configs_tags RENAME TO idx_ua_configs_tags;

-- sessions表索引
ALTER INDEX IF EXISTS idx_sessions_session_token RENAME TO idx_ua_sessions_session_token;
ALTER INDEX IF EXISTS idx_sessions_user_id RENAME TO idx_ua_sessions_user_id;
ALTER INDEX IF EXISTS idx_sessions_is_active RENAME TO idx_ua_sessions_is_active;
ALTER INDEX IF EXISTS idx_sessions_expires_at RENAME TO idx_ua_sessions_expires_at;
ALTER INDEX IF EXISTS idx_sessions_last_activity_at RENAME TO idx_ua_sessions_last_activity_at;
ALTER INDEX IF EXISTS idx_sessions_ip_address RENAME TO idx_ua_sessions_ip_address;
ALTER INDEX IF EXISTS idx_sessions_created_at RENAME TO idx_ua_sessions_created_at;

-- session_activities表索引
ALTER INDEX IF EXISTS idx_session_activities_session_id RENAME TO idx_ua_session_activities_session_id;
ALTER INDEX IF EXISTS idx_session_activities_activity_type RENAME TO idx_ua_session_activities_activity_type;
ALTER INDEX IF EXISTS idx_session_activities_created_at RENAME TO idx_ua_session_activities_created_at;
ALTER INDEX IF EXISTS idx_session_activities_endpoint RENAME TO idx_ua_session_activities_endpoint;
ALTER INDEX IF EXISTS idx_session_activities_status_code RENAME TO idx_ua_session_activities_status_code;

-- audit_logs表索引
ALTER INDEX IF EXISTS idx_audit_logs_user_id RENAME TO idx_ua_audit_logs_user_id;
ALTER INDEX IF EXISTS idx_audit_logs_action RENAME TO idx_ua_audit_logs_action;
ALTER INDEX IF EXISTS idx_audit_logs_resource_type RENAME TO idx_ua_audit_logs_resource_type;
ALTER INDEX IF EXISTS idx_audit_logs_resource_id RENAME TO idx_ua_audit_logs_resource_id;
ALTER INDEX IF EXISTS idx_audit_logs_severity RENAME TO idx_ua_audit_logs_severity;
ALTER INDEX IF EXISTS idx_audit_logs_status RENAME TO idx_ua_audit_logs_status;
ALTER INDEX IF EXISTS idx_audit_logs_created_at RENAME TO idx_ua_audit_logs_created_at;
ALTER INDEX IF EXISTS idx_audit_logs_partition_date RENAME TO idx_ua_audit_logs_partition_date;
ALTER INDEX IF EXISTS idx_audit_logs_ip_address RENAME TO idx_ua_audit_logs_ip_address;
ALTER INDEX IF EXISTS idx_audit_logs_request_id RENAME TO idx_ua_audit_logs_request_id;
ALTER INDEX IF EXISTS idx_audit_logs_trace_id RENAME TO idx_ua_audit_logs_trace_id;
ALTER INDEX IF EXISTS idx_audit_logs_tags RENAME TO idx_ua_audit_logs_tags;
ALTER INDEX IF EXISTS idx_audit_logs_user_action_date RENAME TO idx_ua_audit_logs_user_action_date;
ALTER INDEX IF EXISTS idx_audit_logs_resource_action_date RENAME TO idx_ua_audit_logs_resource_action_date;

-- system_settings表索引
ALTER INDEX IF EXISTS idx_system_settings_category RENAME TO idx_ua_system_settings_category;
ALTER INDEX IF EXISTS idx_system_settings_key RENAME TO idx_ua_system_settings_key;
ALTER INDEX IF EXISTS idx_system_settings_is_public RENAME TO idx_ua_system_settings_is_public;
ALTER INDEX IF EXISTS idx_system_settings_group_name RENAME TO idx_ua_system_settings_group_name;
ALTER INDEX IF EXISTS idx_system_settings_display_order RENAME TO idx_ua_system_settings_display_order;
ALTER INDEX IF EXISTS idx_system_settings_category_key RENAME TO idx_ua_system_settings_category_key;

-- system_settings_history表索引
ALTER INDEX IF EXISTS idx_system_settings_history_setting_id RENAME TO idx_ua_system_settings_history_setting_id;
ALTER INDEX IF EXISTS idx_system_settings_history_category RENAME TO idx_ua_system_settings_history_category;
ALTER INDEX IF EXISTS idx_system_settings_history_changed_by RENAME TO idx_ua_system_settings_history_changed_by;

-- ========================================
-- 重命名插件市场相关表 (mp_ 前缀)
-- ========================================

-- 重命名插件和分类表
ALTER TABLE IF EXISTS plugins RENAME TO mp_plugins;
ALTER TABLE IF EXISTS categories RENAME TO mp_categories;

-- 更新插件相关表索引
-- plugins表索引
ALTER INDEX IF EXISTS idx_plugins_slug RENAME TO idx_mp_plugins_slug;
ALTER INDEX IF EXISTS idx_plugins_name RENAME TO idx_mp_plugins_name;
ALTER INDEX IF EXISTS idx_plugins_author RENAME TO idx_mp_plugins_author;
ALTER INDEX IF EXISTS idx_plugins_category RENAME TO idx_mp_plugins_category;
ALTER INDEX IF EXISTS idx_plugins_status RENAME TO idx_mp_plugins_status;
ALTER INDEX IF EXISTS idx_plugins_is_featured RENAME TO idx_mp_plugins_is_featured;
ALTER INDEX IF EXISTS idx_plugins_is_verified RENAME TO idx_mp_plugins_is_verified;
ALTER INDEX IF EXISTS idx_plugins_is_premium RENAME TO idx_mp_plugins_is_premium;
ALTER INDEX IF EXISTS idx_plugins_created_by RENAME TO idx_mp_plugins_created_by;
ALTER INDEX IF EXISTS idx_plugins_published_at RENAME TO idx_mp_plugins_published_at;
ALTER INDEX IF EXISTS idx_plugins_download_count RENAME TO idx_mp_plugins_download_count;
ALTER INDEX IF EXISTS idx_plugins_rating_average RENAME TO idx_mp_plugins_rating_average;
ALTER INDEX IF EXISTS idx_plugins_tags RENAME TO idx_mp_plugins_tags;
ALTER INDEX IF EXISTS idx_plugins_keywords RENAME TO idx_mp_plugins_keywords;
ALTER INDEX IF EXISTS idx_plugins_name_version RENAME TO idx_mp_plugins_name_version;

-- ========================================
-- 重命名插件相关表 (plu_ 前缀)
-- ========================================

-- 重命名插件功能相关表
ALTER TABLE IF EXISTS plugin_downloads RENAME TO plu_downloads;
ALTER TABLE IF EXISTS plugin_ratings RENAME TO plu_ratings;
ALTER TABLE IF EXISTS plugin_categories RENAME TO plu_categories;

-- 更新插件功能表索引
-- plugin_downloads表索引
ALTER INDEX IF EXISTS idx_plugin_downloads_plugin_id RENAME TO idx_plu_downloads_plugin_id;
ALTER INDEX IF EXISTS idx_plugin_downloads_user_id RENAME TO idx_plu_downloads_user_id;
ALTER INDEX IF EXISTS idx_plugin_downloads_downloaded_at RENAME TO idx_plu_downloads_downloaded_at;
ALTER INDEX IF EXISTS idx_plugin_downloads_ip_address RENAME TO idx_plu_downloads_ip_address;
ALTER INDEX IF EXISTS idx_plugin_downloads_success RENAME TO idx_plu_downloads_success;

-- plugin_ratings表索引
ALTER INDEX IF EXISTS idx_plugin_ratings_plugin_id RENAME TO idx_plu_ratings_plugin_id;
ALTER INDEX IF EXISTS idx_plugin_ratings_user_id RENAME TO idx_plu_ratings_user_id;
ALTER INDEX IF EXISTS idx_plugin_ratings_rating RENAME TO idx_plu_ratings_rating;
ALTER INDEX IF EXISTS idx_plugin_ratings_created_at RENAME TO idx_plu_ratings_created_at;

-- plugin_categories表索引
ALTER INDEX IF EXISTS idx_plugin_categories_slug RENAME TO idx_plu_categories_slug;
ALTER INDEX IF EXISTS idx_plugin_categories_parent_id RENAME TO idx_plu_categories_parent_id;
ALTER INDEX IF EXISTS idx_plugin_categories_sort_order RENAME TO idx_plu_categories_sort_order;
ALTER INDEX IF EXISTS idx_plugin_categories_is_active RENAME TO idx_plu_categories_is_active;

-- ========================================
-- 更新触发器名称
-- ========================================

-- 后台管理表触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON ua_users;
CREATE TRIGGER update_ua_users_updated_at 
    BEFORE UPDATE ON ua_users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_configs_updated_at ON ua_configs;
CREATE TRIGGER update_ua_configs_updated_at 
    BEFORE UPDATE ON ua_configs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_sessions_updated_at ON ua_sessions;
CREATE TRIGGER update_ua_sessions_updated_at 
    BEFORE UPDATE ON ua_sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_system_settings_updated_at ON ua_system_settings;
CREATE TRIGGER update_ua_system_settings_updated_at 
    BEFORE UPDATE ON ua_system_settings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插件市场表触发器
DROP TRIGGER IF EXISTS update_plugins_updated_at ON mp_plugins;
CREATE TRIGGER update_mp_plugins_updated_at 
    BEFORE UPDATE ON mp_plugins 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插件功能表触发器
DROP TRIGGER IF EXISTS update_plugin_ratings_updated_at ON plu_ratings;
CREATE TRIGGER update_plu_ratings_updated_at 
    BEFORE UPDATE ON plu_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_plugin_categories_updated_at ON plu_categories;
CREATE TRIGGER update_plu_categories_updated_at 
    BEFORE UPDATE ON plu_categories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- +migrate Down

BEGIN;

-- ========================================
-- 恢复触发器名称
-- ========================================

-- 恢复后台管理表触发器
DROP TRIGGER IF EXISTS update_ua_users_updated_at ON ua_users;
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON ua_users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ua_configs_updated_at ON ua_configs;
CREATE TRIGGER update_configs_updated_at 
    BEFORE UPDATE ON ua_configs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ua_sessions_updated_at ON ua_sessions;
CREATE TRIGGER update_sessions_updated_at 
    BEFORE UPDATE ON ua_sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_ua_system_settings_updated_at ON ua_system_settings;
CREATE TRIGGER update_system_settings_updated_at 
    BEFORE UPDATE ON ua_system_settings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 恢复插件市场表触发器
DROP TRIGGER IF EXISTS update_mp_plugins_updated_at ON mp_plugins;
CREATE TRIGGER update_plugins_updated_at 
    BEFORE UPDATE ON mp_plugins 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 恢复插件功能表触发器
DROP TRIGGER IF EXISTS update_plu_ratings_updated_at ON plu_ratings;
CREATE TRIGGER update_plugin_ratings_updated_at 
    BEFORE UPDATE ON plu_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_plu_categories_updated_at ON plu_categories;
CREATE TRIGGER update_plugin_categories_updated_at 
    BEFORE UPDATE ON plu_categories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- 恢复插件相关表名 (plu_ 前缀)
-- ========================================

-- 恢复插件功能表索引
ALTER INDEX IF EXISTS idx_plu_categories_is_active RENAME TO idx_plugin_categories_is_active;
ALTER INDEX IF EXISTS idx_plu_categories_sort_order RENAME TO idx_plugin_categories_sort_order;
ALTER INDEX IF EXISTS idx_plu_categories_parent_id RENAME TO idx_plugin_categories_parent_id;
ALTER INDEX IF EXISTS idx_plu_categories_slug RENAME TO idx_plugin_categories_slug;

ALTER INDEX IF EXISTS idx_plu_ratings_created_at RENAME TO idx_plugin_ratings_created_at;
ALTER INDEX IF EXISTS idx_plu_ratings_rating RENAME TO idx_plugin_ratings_rating;
ALTER INDEX IF EXISTS idx_plu_ratings_user_id RENAME TO idx_plugin_ratings_user_id;
ALTER INDEX IF EXISTS idx_plu_ratings_plugin_id RENAME TO idx_plugin_ratings_plugin_id;

ALTER INDEX IF EXISTS idx_plu_downloads_success RENAME TO idx_plugin_downloads_success;
ALTER INDEX IF EXISTS idx_plu_downloads_ip_address RENAME TO idx_plugin_downloads_ip_address;
ALTER INDEX IF EXISTS idx_plu_downloads_downloaded_at RENAME TO idx_plugin_downloads_downloaded_at;
ALTER INDEX IF EXISTS idx_plu_downloads_user_id RENAME TO idx_plugin_downloads_user_id;
ALTER INDEX IF EXISTS idx_plu_downloads_plugin_id RENAME TO idx_plugin_downloads_plugin_id;

-- 恢复插件功能相关表名
ALTER TABLE IF EXISTS plu_categories RENAME TO plugin_categories;
ALTER TABLE IF EXISTS plu_ratings RENAME TO plugin_ratings;
ALTER TABLE IF EXISTS plu_downloads RENAME TO plugin_downloads;

-- ========================================
-- 恢复插件市场相关表名 (mp_ 前缀)
-- ========================================

-- 恢复插件相关表索引
ALTER INDEX IF EXISTS idx_mp_plugins_name_version RENAME TO idx_plugins_name_version;
ALTER INDEX IF EXISTS idx_mp_plugins_keywords RENAME TO idx_plugins_keywords;
ALTER INDEX IF EXISTS idx_mp_plugins_tags RENAME TO idx_plugins_tags;
ALTER INDEX IF EXISTS idx_mp_plugins_rating_average RENAME TO idx_plugins_rating_average;
ALTER INDEX IF EXISTS idx_mp_plugins_download_count RENAME TO idx_plugins_download_count;
ALTER INDEX IF EXISTS idx_mp_plugins_published_at RENAME TO idx_plugins_published_at;
ALTER INDEX IF EXISTS idx_mp_plugins_created_by RENAME TO idx_plugins_created_by;
ALTER INDEX IF EXISTS idx_mp_plugins_is_premium RENAME TO idx_plugins_is_premium;
ALTER INDEX IF EXISTS idx_mp_plugins_is_verified RENAME TO idx_plugins_is_verified;
ALTER INDEX IF EXISTS idx_mp_plugins_is_featured RENAME TO idx_plugins_is_featured;
ALTER INDEX IF EXISTS idx_mp_plugins_status RENAME TO idx_plugins_status;
ALTER INDEX IF EXISTS idx_mp_plugins_category RENAME TO idx_plugins_category;
ALTER INDEX IF EXISTS idx_mp_plugins_author RENAME TO idx_plugins_author;
ALTER INDEX IF EXISTS idx_mp_plugins_name RENAME TO idx_plugins_name;
ALTER INDEX IF EXISTS idx_mp_plugins_slug RENAME TO idx_plugins_slug;

-- 恢复插件和分类表名
ALTER TABLE IF EXISTS mp_categories RENAME TO categories;
ALTER TABLE IF EXISTS mp_plugins RENAME TO plugins;

-- ========================================
-- 恢复后台管理相关表名 (ua_ 前缀)
-- ========================================

-- 恢复system_settings_history表索引
ALTER INDEX IF EXISTS idx_ua_system_settings_history_changed_by RENAME TO idx_system_settings_history_changed_by;
ALTER INDEX IF EXISTS idx_ua_system_settings_history_category RENAME TO idx_system_settings_history_category;
ALTER INDEX IF EXISTS idx_ua_system_settings_history_setting_id RENAME TO idx_system_settings_history_setting_id;

-- 恢复system_settings表索引
ALTER INDEX IF EXISTS idx_ua_system_settings_category_key RENAME TO idx_system_settings_category_key;
ALTER INDEX IF EXISTS idx_ua_system_settings_display_order RENAME TO idx_system_settings_display_order;
ALTER INDEX IF EXISTS idx_ua_system_settings_group_name RENAME TO idx_system_settings_group_name;
ALTER INDEX IF EXISTS idx_ua_system_settings_is_public RENAME TO idx_system_settings_is_public;
ALTER INDEX IF EXISTS idx_ua_system_settings_key RENAME TO idx_system_settings_key;
ALTER INDEX IF EXISTS idx_ua_system_settings_category RENAME TO idx_system_settings_category;

-- 恢复audit_logs表索引
ALTER INDEX IF EXISTS idx_ua_audit_logs_resource_action_date RENAME TO idx_audit_logs_resource_action_date;
ALTER INDEX IF EXISTS idx_ua_audit_logs_user_action_date RENAME TO idx_audit_logs_user_action_date;
ALTER INDEX IF EXISTS idx_ua_audit_logs_tags RENAME TO idx_audit_logs_tags;
ALTER INDEX IF EXISTS idx_ua_audit_logs_trace_id RENAME TO idx_audit_logs_trace_id;
ALTER INDEX IF EXISTS idx_ua_audit_logs_request_id RENAME TO idx_audit_logs_request_id;
ALTER INDEX IF EXISTS idx_ua_audit_logs_ip_address RENAME TO idx_audit_logs_ip_address;
ALTER INDEX IF EXISTS idx_ua_audit_logs_partition_date RENAME TO idx_audit_logs_partition_date;
ALTER INDEX IF EXISTS idx_ua_audit_logs_created_at RENAME TO idx_audit_logs_created_at;
ALTER INDEX IF EXISTS idx_ua_audit_logs_status RENAME TO idx_audit_logs_status;
ALTER INDEX IF EXISTS idx_ua_audit_logs_severity RENAME TO idx_audit_logs_severity;
ALTER INDEX IF EXISTS idx_ua_audit_logs_resource_id RENAME TO idx_audit_logs_resource_id;
ALTER INDEX IF EXISTS idx_ua_audit_logs_resource_type RENAME TO idx_audit_logs_resource_type;
ALTER INDEX IF EXISTS idx_ua_audit_logs_action RENAME TO idx_audit_logs_action;
ALTER INDEX IF EXISTS idx_ua_audit_logs_user_id RENAME TO idx_audit_logs_user_id;

-- 恢复session_activities表索引
ALTER INDEX IF EXISTS idx_ua_session_activities_status_code RENAME TO idx_session_activities_status_code;
ALTER INDEX IF EXISTS idx_ua_session_activities_endpoint RENAME TO idx_session_activities_endpoint;
ALTER INDEX IF EXISTS idx_ua_session_activities_created_at RENAME TO idx_session_activities_created_at;
ALTER INDEX IF EXISTS idx_ua_session_activities_activity_type RENAME TO idx_session_activities_activity_type;
ALTER INDEX IF EXISTS idx_ua_session_activities_session_id RENAME TO idx_session_activities_session_id;

-- 恢复sessions表索引
ALTER INDEX IF EXISTS idx_ua_sessions_created_at RENAME TO idx_sessions_created_at;
ALTER INDEX IF EXISTS idx_ua_sessions_ip_address RENAME TO idx_sessions_ip_address;
ALTER INDEX IF EXISTS idx_ua_sessions_last_activity_at RENAME TO idx_sessions_last_activity_at;
ALTER INDEX IF EXISTS idx_ua_sessions_expires_at RENAME TO idx_sessions_expires_at;
ALTER INDEX IF EXISTS idx_ua_sessions_is_active RENAME TO idx_sessions_is_active;
ALTER INDEX IF EXISTS idx_ua_sessions_user_id RENAME TO idx_sessions_user_id;
ALTER INDEX IF EXISTS idx_ua_sessions_session_token RENAME TO idx_sessions_session_token;

-- 恢复configs表索引
ALTER INDEX IF EXISTS idx_ua_configs_tags RENAME TO idx_configs_tags;
ALTER INDEX IF EXISTS idx_ua_configs_created_at RENAME TO idx_configs_created_at;
ALTER INDEX IF EXISTS idx_ua_configs_created_by RENAME TO idx_configs_created_by;
ALTER INDEX IF EXISTS idx_ua_configs_is_public RENAME TO idx_configs_is_public;
ALTER INDEX IF EXISTS idx_ua_configs_type RENAME TO idx_configs_type;
ALTER INDEX IF EXISTS idx_ua_configs_environment RENAME TO idx_configs_environment;
ALTER INDEX IF EXISTS idx_ua_configs_namespace RENAME TO idx_configs_namespace;
ALTER INDEX IF EXISTS idx_ua_configs_key RENAME TO idx_configs_key;

-- 恢复users表索引
ALTER INDEX IF EXISTS idx_ua_users_email_verified RENAME TO idx_users_email_verified;
ALTER INDEX IF EXISTS idx_ua_users_last_login_at RENAME TO idx_users_last_login_at;
ALTER INDEX IF EXISTS idx_ua_users_created_at RENAME TO idx_users_created_at;
ALTER INDEX IF EXISTS idx_ua_users_status RENAME TO idx_users_status;
ALTER INDEX IF EXISTS idx_ua_users_role RENAME TO idx_users_role;
ALTER INDEX IF EXISTS idx_ua_users_email RENAME TO idx_users_email;

-- 恢复用户相关表名
ALTER TABLE IF EXISTS ua_system_settings_history RENAME TO system_settings_history;
ALTER TABLE IF EXISTS ua_system_settings RENAME TO system_settings;
ALTER TABLE IF EXISTS ua_security_events RENAME TO security_events;
ALTER TABLE IF EXISTS ua_audit_logs RENAME TO audit_logs;
ALTER TABLE IF EXISTS ua_session_activities RENAME TO session_activities;
ALTER TABLE IF EXISTS ua_sessions RENAME TO sessions;
ALTER TABLE IF EXISTS ua_configs RENAME TO configs;
ALTER TABLE IF EXISTS ua_users RENAME TO users;

COMMIT;