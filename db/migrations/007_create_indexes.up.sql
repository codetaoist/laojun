-- Create indexes for performance optimization

-- =====================================================
-- Marketplace tables indexes
-- =====================================================

-- mp_users indexes
CREATE INDEX IF NOT EXISTS idx_mp_users_email ON mp_users(email);
CREATE INDEX IF NOT EXISTS idx_mp_users_username ON mp_users(username);
CREATE INDEX IF NOT EXISTS idx_mp_users_is_active ON mp_users(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_users_created_at ON mp_users(created_at);

-- mp_categories indexes
CREATE INDEX IF NOT EXISTS idx_mp_categories_is_active ON mp_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_categories_sort_order ON mp_categories(sort_order);

-- mp_plugins indexes
CREATE INDEX IF NOT EXISTS idx_mp_plugins_category_id ON mp_plugins(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_author_id ON mp_plugins(author_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_status ON mp_plugins(status);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_is_featured ON mp_plugins(is_featured);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_created_at ON mp_plugins(created_at);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_updated_at ON mp_plugins(updated_at);

-- mp_plugin_reviews indexes
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_plugin_id ON mp_plugin_reviews(plugin_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_user_id ON mp_plugin_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_rating ON mp_plugin_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_mp_plugin_reviews_created_at ON mp_plugin_reviews(created_at);

-- mp_forum_categories indexes
CREATE INDEX IF NOT EXISTS idx_mp_forum_categories_is_active ON mp_forum_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_mp_forum_categories_sort_order ON mp_forum_categories(sort_order);

-- =====================================================
-- User Authentication indexes
-- =====================================================

-- ua_admin indexes
CREATE INDEX IF NOT EXISTS idx_ua_admin_email ON ua_admin(email);
CREATE INDEX IF NOT EXISTS idx_ua_admin_username ON ua_admin(username);
CREATE INDEX IF NOT EXISTS idx_ua_admin_is_active ON ua_admin(is_active);
CREATE INDEX IF NOT EXISTS idx_ua_admin_is_super_admin ON ua_admin(is_super_admin);
CREATE INDEX IF NOT EXISTS idx_ua_admin_last_login_at ON ua_admin(last_login_at);

-- ua_user_sessions indexes
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_user_id ON ua_user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_token_hash ON ua_user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_expires_at ON ua_user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_ua_user_sessions_ip_address ON ua_user_sessions(ip_address);

-- ua_jwt_keys indexes
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_key_id ON ua_jwt_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_is_active ON ua_jwt_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_ua_jwt_keys_expires_at ON ua_jwt_keys(expires_at);

-- =====================================================
-- Authorization indexes
-- =====================================================

-- az_roles indexes
CREATE INDEX IF NOT EXISTS idx_az_roles_name ON az_roles(name);
CREATE INDEX IF NOT EXISTS idx_az_roles_is_active ON az_roles(is_active);
CREATE INDEX IF NOT EXISTS idx_az_roles_is_system ON az_roles(is_system);

-- az_permissions indexes
CREATE INDEX IF NOT EXISTS idx_az_permissions_code ON az_permissions(code);
CREATE INDEX IF NOT EXISTS idx_az_permissions_resource ON az_permissions(resource);
CREATE INDEX IF NOT EXISTS idx_az_permissions_action ON az_permissions(action);
CREATE INDEX IF NOT EXISTS idx_az_permissions_is_system ON az_permissions(is_system);

-- az_user_roles indexes
CREATE INDEX IF NOT EXISTS idx_az_user_roles_user_id ON az_user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_role_id ON az_user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_az_user_roles_expires_at ON az_user_roles(expires_at);

-- az_role_permissions indexes
CREATE INDEX IF NOT EXISTS idx_az_role_permissions_role_id ON az_role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_az_role_permissions_permission_id ON az_role_permissions(permission_id);

-- =====================================================
-- System Management indexes
-- =====================================================

-- sm_menus indexes
CREATE INDEX IF NOT EXISTS idx_sm_menus_parent_id ON sm_menus(parent_id);
CREATE INDEX IF NOT EXISTS idx_sm_menus_is_active ON sm_menus(is_active);
CREATE INDEX IF NOT EXISTS idx_sm_menus_is_hidden ON sm_menus(is_hidden);
CREATE INDEX IF NOT EXISTS idx_sm_menus_sort_order ON sm_menus(sort_order);

-- sm_device_types indexes
CREATE INDEX IF NOT EXISTS idx_sm_device_types_code ON sm_device_types(code);
CREATE INDEX IF NOT EXISTS idx_sm_device_types_is_active ON sm_device_types(is_active);

-- sm_modules indexes
CREATE INDEX IF NOT EXISTS idx_sm_modules_code ON sm_modules(code);
CREATE INDEX IF NOT EXISTS idx_sm_modules_is_active ON sm_modules(is_active);

-- =====================================================
-- System Features indexes
-- =====================================================

-- sys_settings indexes
CREATE INDEX IF NOT EXISTS idx_sys_settings_key ON sys_settings(key);
CREATE INDEX IF NOT EXISTS idx_sys_settings_is_public ON sys_settings(is_public);
CREATE INDEX IF NOT EXISTS idx_sys_settings_type ON sys_settings(type);

-- sys_icons indexes
CREATE INDEX IF NOT EXISTS idx_sys_icons_code ON sys_icons(code);
CREATE INDEX IF NOT EXISTS idx_sys_icons_category ON sys_icons(category);
CREATE INDEX IF NOT EXISTS idx_sys_icons_is_active ON sys_icons(is_active);

-- sys_audit_logs indexes
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_user_id ON sys_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_action ON sys_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_resource ON sys_audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_resource_id ON sys_audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_created_at ON sys_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_sys_audit_logs_ip_address ON sys_audit_logs(ip_address);