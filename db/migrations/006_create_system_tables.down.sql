-- Drop system tables in reverse dependency order

-- Drop system features tables
DROP TABLE IF EXISTS sys_audit_logs CASCADE;
DROP TABLE IF EXISTS sys_icons CASCADE;
DROP TABLE IF EXISTS sys_settings CASCADE;

-- Drop system management tables
DROP TABLE IF EXISTS sm_modules CASCADE;
DROP TABLE IF EXISTS sm_device_types CASCADE;
DROP TABLE IF EXISTS sm_menus CASCADE;

-- Drop authorization tables
DROP TABLE IF EXISTS az_role_permissions CASCADE;
DROP TABLE IF EXISTS az_user_roles CASCADE;
DROP TABLE IF EXISTS az_permissions CASCADE;
DROP TABLE IF EXISTS az_roles CASCADE;

-- Drop user authentication tables
DROP TABLE IF EXISTS ua_user_sessions CASCADE;
DROP TABLE IF EXISTS ua_jwt_keys CASCADE;
DROP TABLE IF EXISTS ua_admin CASCADE;