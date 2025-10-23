-- 为 mp_plugins 表添加审核相关字段
-- 这些字段是插件审核系统所需的

BEGIN;

-- 添加审核状态字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS review_status VARCHAR(50) DEFAULT 'pending';

-- 添加审核优先级字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS review_priority VARCHAR(20) DEFAULT 'normal';

-- 添加自动审核评分字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS auto_review_score DECIMAL(3,2);

-- 添加自动审核结果字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS auto_review_result VARCHAR(50);

-- 添加审核备注字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS review_notes TEXT;

-- 添加审核完成时间字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP WITH TIME ZONE;

-- 添加审核员ID字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS reviewer_id UUID;

-- 添加提交审核时间字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS submitted_for_review_at TIMESTAMP WITH TIME ZONE;

-- 添加拒绝原因字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

-- 添加申诉次数字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS appeal_count INTEGER DEFAULT 0;

-- 添加最后申诉时间字段
ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS last_appeal_at TIMESTAMP WITH TIME ZONE;

-- 创建审核状态的检查约束
ALTER TABLE mp_plugins ADD CONSTRAINT IF NOT EXISTS chk_review_status 
    CHECK (review_status IN ('pending', 'in_review', 'approved', 'rejected', 'appealing', 'suspended'));

-- 创建审核优先级的检查约束
ALTER TABLE mp_plugins ADD CONSTRAINT IF NOT EXISTS chk_review_priority 
    CHECK (review_priority IN ('low', 'normal', 'high', 'urgent'));

-- 创建自动审核结果的检查约束
ALTER TABLE mp_plugins ADD CONSTRAINT IF NOT EXISTS chk_auto_review_result 
    CHECK (auto_review_result IN ('pass', 'fail', 'manual_review', 'pending'));

-- 为审核相关字段创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_mp_plugins_review_status ON mp_plugins(review_status);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_review_priority ON mp_plugins(review_priority);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_reviewer_id ON mp_plugins(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_submitted_for_review_at ON mp_plugins(submitted_for_review_at);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_reviewed_at ON mp_plugins(reviewed_at);

-- 为现有插件设置默认的审核状态
-- 根据当前的 status 字段来设置 review_status
UPDATE mp_plugins SET 
    review_status = CASE 
        WHEN status = 'approved' THEN 'approved'
        WHEN status = 'rejected' THEN 'rejected'
        WHEN status = 'suspended' THEN 'suspended'
        WHEN status = 'pending' THEN 'pending'
        ELSE 'pending'
    END,
    submitted_for_review_at = CASE 
        WHEN status IN ('pending', 'approved', 'rejected') THEN created_at
        ELSE NULL
    END,
    reviewed_at = CASE 
        WHEN status IN ('approved', 'rejected', 'suspended') THEN updated_at
        ELSE NULL
    END
WHERE review_status IS NULL OR review_status = 'pending';

COMMIT;

-- 验证迁移结果
SELECT 
    'mp_plugins表审核字段迁移完成' as message,
    COUNT(*) as total_plugins,
    COUNT(CASE WHEN review_status = 'pending' THEN 1 END) as pending_review,
    COUNT(CASE WHEN review_status = 'approved' THEN 1 END) as approved,
    COUNT(CASE WHEN review_status = 'rejected' THEN 1 END) as rejected
FROM mp_plugins;