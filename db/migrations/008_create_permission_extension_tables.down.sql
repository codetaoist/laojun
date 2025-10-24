-- Drop permission extension tables in reverse dependency order

-- Drop indexes first
DROP INDEX IF EXISTS idx_pe_user_device_permissions_device_type_id;
DROP INDEX IF EXISTS idx_pe_user_device_permissions_device_id;
DROP INDEX IF EXISTS idx_pe_user_device_permissions_user_id;
DROP INDEX IF EXISTS idx_pe_permission_inheritance_child;
DROP INDEX IF EXISTS idx_pe_permission_inheritance_parent;
DROP INDEX IF EXISTS idx_pe_extended_permissions_resource_action;
DROP INDEX IF EXISTS idx_pe_extended_permissions_module_id;
DROP INDEX IF EXISTS idx_pe_extended_permissions_device_type_id;
DROP INDEX IF EXISTS idx_ug_user_group_permissions_permission_id;
DROP INDEX IF EXISTS idx_ug_user_group_permissions_group_id;
DROP INDEX IF EXISTS idx_ug_permission_templates_name;
DROP INDEX IF EXISTS idx_ug_user_group_members_group_id;
DROP INDEX IF EXISTS idx_ug_user_group_members_user_id;
DROP INDEX IF EXISTS idx_ug_user_groups_name;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS pe_user_device_permissions CASCADE;
DROP TABLE IF EXISTS pe_permission_inheritance CASCADE;
DROP TABLE IF EXISTS pe_extended_permissions CASCADE;
DROP TABLE IF EXISTS ug_user_group_permissions CASCADE;
DROP TABLE IF EXISTS ug_permission_templates CASCADE;
DROP TABLE IF EXISTS ug_user_group_members CASCADE;
DROP TABLE IF EXISTS ug_user_groups CASCADE;