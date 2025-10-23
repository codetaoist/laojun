-- 验证super_admin角色和用户分配

-- 1. 检查super_admin角色是否存在
SELECT 'super_admin角色:' as type, id, name, description, is_system, created_at 
FROM az_roles 
WHERE name = 'super_admin';

-- 2. 检查admin用户信息
SELECT 'admin用户:' as type, id, username, email, is_active, created_at 
FROM ua_admin 
WHERE username = 'admin';

-- 3. 检查admin用户的角色分配
SELECT 'admin用户角色分配:' as type, ur.id, ur.user_id, ur.role_id, r.name as role_name, ur.created_at
FROM az_user_roles ur
INNER JOIN az_roles r ON ur.role_id = r.id
INNER JOIN ua_admin u ON ur.user_id = u.id
WHERE u.username = 'admin';

-- 4. 检查所有用户角色分配
SELECT '所有用户角色分配:' as type, u.username, r.name as role_name, ur.created_at
FROM az_user_roles ur
INNER JOIN az_roles r ON ur.role_id = r.id
INNER JOIN ua_admin u ON ur.user_id = u.id
ORDER BY u.username, r.name;