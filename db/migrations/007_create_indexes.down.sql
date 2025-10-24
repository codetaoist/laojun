-- Drop indexes created in 007_create_indexes.up.sql

-- =====================================================
-- System Features indexes
-- =====================================================

-- sys_audit_logs indexes
DROP INDEX IF EXISTS idx_sys_audit_logs_ip_address;
DROP INDEX IF EXISTS idx_sys_audit_logs_created_at;
DROP INDEX IF EXISTS idx_sys_audit_logs_resource_id;
DROP INDEX IF EXISTS idx_sys_audit_logs_resource;
DROP INDEX IF EXISTS idx_sys_audit_logs_action;
DROP INDEX IF EXISTS idx_sys_audit_logs_user_id;

-- sys_icons indexes
DROP INDEX IF EXISTS idx_sys_icons_is_active;
DROP INDEX IF EXISTS idx_sys_icons_category;
DROP INDEX IF EXISTS idx_sys_icons_code;

-- sys_settings indexes
DROP INDEX IF EXISTS idx_sys_settings_type;
DROP INDEX IF EXISTS idx_sys_settings_is_public;
DROP INDEX IF EXISTS idx_sys_settings_key;

-- =====================================================
-- System Management indexes
-- =====================================================

-- sm_modules indexes
DROP INDEX IF EXISTS idx_sm_modules_is_active;
DROP INDEX IF EXISTS idx_sm_modules_code;

-- sm_device_types indexes
DROP INDEX IF EXISTS idx_sm_device_types_is_active;
DROP INDEX IF EXISTS idx_sm_device_types_code;

-- sm_menus indexes
DROP INDEX IF EXISTS idx_sm_menus_sort_order;
DROP INDEX IF EXISTS idx_sm_menus_is_hidden;
DROP INDEX IF EXISTS idx_sm_menus_is_active;
DROP INDEX IF EXISTS idx_sm_menus_parent_id;

-- =====================================================
-- Authorization indexes
-- =====================================================

-- az_role_permissions indexes
DROP INDEX IF EXISTS idx_az_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_az_role_permissions_role_id;

-- az_user_roles indexes
DROP INDEX IF EXISTS idx_az_user_roles_expires_at;
DROP INDEX IF EXISTS idx_az_user_roles_role_id;
DROP INDEX IF EXISTS idx_az_user_roles_user_id;

-- az_permissions indexes
DROP INDEX IF EXISTS idx_az_permissions_is_system;
DROP INDEX IF EXISTS idx_az_permissions_action;
DROP INDEX IF EXISTS idx_az_permissions_resource;
DROP INDEX IF EXISTS idx_az_permissions_code;

-- az_roles indexes
DROP INDEX IF EXISTS idx_az_roles_is_system;
DROP INDEX IF EXISTS idx_az_roles_is_active;
DROP INDEX IF EXISTS idx_az_roles_name;

-- =====================================================
-- User Authentication indexes
-- =====================================================

-- ua_jwt_keys indexes
DROP INDEX IF EXISTS idx_ua_jwt_keys_expires_at;
DROP INDEX IF EXISTS idx_ua_jwt_keys_is_active;
DROP INDEX IF EXISTS idx_ua_jwt_keys_key_id;

-- ua_user_sessions indexes
DROP INDEX IF EXISTS idx_ua_user_sessions_ip_address;
DROP INDEX IF EXISTS idx_ua_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_ua_user_sessions_token_hash;
DROP INDEX IF EXISTS idx_ua_user_sessions_user_id;

-- ua_admin indexes
DROP INDEX IF EXISTS idx_ua_admin_last_login_at;
DROP INDEX IF EXISTS idx_ua_admin_is_super_admin;
DROP INDEX IF EXISTS idx_ua_admin_is_active;
DROP INDEX IF EXISTS idx_ua_admin_username;
DROP INDEX IF EXISTS idx_ua_admin_email;

-- =====================================================
-- Marketplace tables indexes
-- =====================================================

-- mp_forum_categories indexes
DROP INDEX IF EXISTS idx_mp_forum_categories_sort_order;
DROP INDEX IF EXISTS idx_mp_forum_categories_is_active;

-- mp_plugin_reviews indexes
DROP INDEX IF EXISTS idx_mp_plugin_reviews_created_at;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_rating;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_user_id;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_plugin_id;

-- mp_plugins indexes
DROP INDEX IF EXISTS idx_mp_plugins_updated_at;
DROP INDEX IF EXISTS idx_mp_plugins_created_at;
DROP INDEX IF EXISTS idx_mp_plugins_is_featured;
DROP INDEX IF EXISTS idx_mp_plugins_status;
DROP INDEX IF EXISTS idx_mp_plugins_author_id;
DROP INDEX IF EXISTS idx_mp_plugins_category_id;

-- mp_categories indexes
DROP INDEX IF EXISTS idx_mp_categories_sort_order;
DROP INDEX IF EXISTS idx_mp_categories_is_active;

-- mp_users indexes
DROP INDEX IF EXISTS idx_mp_users_created_at;
DROP INDEX IF EXISTS idx_mp_users_is_active;
DROP INDEX IF EXISTS idx_mp_users_username;
DROP INDEX IF EXISTS idx_mp_users_email;

-- Insert migration record
INSERT INTO public.schema_migrations (version, applied_at) 
VALUES ('007', NOW())
ON CONFLICT (version) DO NOTHING;