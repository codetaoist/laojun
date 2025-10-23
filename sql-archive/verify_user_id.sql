-- 验证JWT token中的用户ID是否与数据库中的admin用户匹配
-- JWT token中的用户ID: af1f6f34-ced2-4baf-883b-6a588011a74e

-- 查询admin用户的详细信息
SELECT 
    id,
    username,
    email,
    created_at,
    updated_at
FROM ua_admin 
WHERE username = 'admin';

-- 查询特定用户ID的详细信息
SELECT 
    id,
    username,
    email,
    created_at,
    updated_at
FROM ua_admin 
WHERE id = 'af1f6f34-ced2-4baf-883b-6a588011a74e';

-- 验证用户ID匹配
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM ua_admin 
            WHERE username = 'admin' 
            AND id = 'af1f6f34-ced2-4baf-883b-6a588011a74e'
        ) 
        THEN 'JWT token中的用户ID与admin用户匹配' 
        ELSE 'JWT token中的用户ID与admin用户不匹配' 
    END as verification_result;