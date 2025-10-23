-- 创建 super_admin 角色（如果不存在）
INSERT INTO az_roles (id, name, display_name, description, is_system, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'super_admin',
    '超级管理员',
    '系统超级管理员，拥有所有权限',
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (name) DO NOTHING;

-- 为 admin 用户分配 super_admin 角色
INSERT INTO az_user_roles (user_id, role_id, created_at)
SELECT 
    u.id as user_id,
    r.id as role_id,
    CURRENT_TIMESTAMP as created_at
FROM ua_admin u, az_roles r
WHERE u.username = 'admin' 
  AND r.name = 'super_admin'
  AND NOT EXISTS (
    SELECT 1 FROM az_user_roles ur 
    WHERE ur.user_id = u.id AND ur.role_id = r.id
  );

-- 验证结果
SELECT 
    u.username,
    u.email,
    r.name as role_name,
    r.display_name as role_display_name
FROM ua_admin u
JOIN az_user_roles ur ON u.id = ur.user_id
JOIN az_roles r ON ur.role_id = r.id
WHERE u.username = 'admin';