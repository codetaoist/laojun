-- Drop all indexes created in 007_create_indexes.up.sql

-- System Features indexes
DROP INDEX IF EXISTS idx_sys_audit_logs_ip_address;
DROP INDEX IF EXISTS idx_sys_audit_logs_created_at;
DROP INDEX IF EXISTS idx_sys_audit_logs_resource_id;
DROP INDEX IF EXISTS idx_sys_audit_logs_resource;
DROP INDEX IF EXISTS idx_sys_audit_logs_action;
DROP INDEX IF EXISTS idx_sys_audit_logs_user_id;
DROP INDEX IF EXISTS idx_sys_icons_is_active;
DROP INDEX IF EXISTS idx_sys_icons_category;
DROP INDEX IF EXISTS idx_sys_icons_code;
DROP INDEX IF EXISTS idx_sys_settings_type;
DROP INDEX IF EXISTS idx_sys_settings_is_public;
DROP INDEX IF EXISTS idx_sys_settings_key;

-- System Management indexes
DROP INDEX IF EXISTS idx_sm_modules_is_active;
DROP INDEX IF EXISTS idx_sm_modules_code;
DROP INDEX IF EXISTS idx_sm_device_types_is_active;
DROP INDEX IF EXISTS idx_sm_device_types_code;
DROP INDEX IF EXISTS idx_sm_menus_sort_order;
DROP INDEX IF EXISTS idx_sm_menus_is_hidden;
DROP INDEX IF EXISTS idx_sm_menus_is_active;
DROP INDEX IF EXISTS idx_sm_menus_parent_id;

-- Authorization indexes
DROP INDEX IF EXISTS idx_az_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_az_role_permissions_role_id;
DROP INDEX IF EXISTS idx_az_user_roles_expires_at;
DROP INDEX IF EXISTS idx_az_user_roles_role_id;
DROP INDEX IF EXISTS idx_az_user_roles_user_id;
DROP INDEX IF EXISTS idx_az_permissions_is_system;
DROP INDEX IF EXISTS idx_az_permissions_action;
DROP INDEX IF EXISTS idx_az_permissions_resource;
DROP INDEX IF EXISTS idx_az_permissions_code;
DROP INDEX IF EXISTS idx_az_roles_is_system;
DROP INDEX IF EXISTS idx_az_roles_is_active;
DROP INDEX IF EXISTS idx_az_roles_name;

-- User Authentication indexes
DROP INDEX IF EXISTS idx_ua_jwt_keys_expires_at;
DROP INDEX IF EXISTS idx_ua_jwt_keys_is_active;
DROP INDEX IF EXISTS idx_ua_jwt_keys_key_id;
DROP INDEX IF EXISTS idx_ua_user_sessions_ip_address;
DROP INDEX IF EXISTS idx_ua_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_ua_user_sessions_token_hash;
DROP INDEX IF EXISTS idx_ua_user_sessions_user_id;
DROP INDEX IF EXISTS idx_ua_admin_last_login_at;
DROP INDEX IF EXISTS idx_ua_admin_is_super_admin;
DROP INDEX IF EXISTS idx_ua_admin_is_active;
DROP INDEX IF EXISTS idx_ua_admin_username;
DROP INDEX IF EXISTS idx_ua_admin_email;

-- Marketplace indexes
DROP INDEX IF EXISTS idx_mp_forum_categories_sort_order;
DROP INDEX IF EXISTS idx_mp_forum_categories_is_active;
DROP INDEX IF EXISTS idx_mp_forum_categories_parent_id;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_created_at;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_rating;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_user_id;
DROP INDEX IF EXISTS idx_mp_plugin_reviews_plugin_id;
DROP INDEX IF EXISTS idx_mp_plugins_updated_at;
DROP INDEX IF EXISTS idx_mp_plugins_created_at;
DROP INDEX IF EXISTS idx_mp_plugins_is_featured;
DROP INDEX IF EXISTS idx_mp_plugins_status;
DROP INDEX IF EXISTS idx_mp_plugins_author_id;
DROP INDEX IF EXISTS idx_mp_plugins_category_id;
DROP INDEX IF EXISTS idx_mp_categories_sort_order;
DROP INDEX IF EXISTS idx_mp_categories_is_active;
DROP INDEX IF EXISTS idx_mp_categories_parent_id;
DROP INDEX IF EXISTS idx_mp_users_created_at;
DROP INDEX IF EXISTS idx_mp_users_is_active;
DROP INDEX IF EXISTS idx_mp_users_username;
DROP INDEX IF EXISTS idx_mp_users_email;