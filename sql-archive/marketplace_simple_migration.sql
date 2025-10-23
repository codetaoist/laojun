-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建marketplace用户表
CREATE TABLE IF NOT EXISTS mp_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(100),
  avatar VARCHAR(255),
  avatar_url VARCHAR(255),
  bio TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 为mp_categories表添加缺少的字段
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_categories' AND column_name='icon') THEN
        ALTER TABLE mp_categories ADD COLUMN icon VARCHAR(50);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_categories' AND column_name='sort_order') THEN
        ALTER TABLE mp_categories ADD COLUMN sort_order INT DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_categories' AND column_name='is_active') THEN
        ALTER TABLE mp_categories ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_categories' AND column_name='created_at') THEN
        ALTER TABLE mp_categories ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_categories' AND column_name='updated_at') THEN
        ALTER TABLE mp_categories ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
    END IF;
END $$;

-- 为mp_plugins表添加缺少的字段
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='version') THEN
        ALTER TABLE mp_plugins ADD COLUMN version VARCHAR(20) NOT NULL DEFAULT '1.0.0';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='author') THEN
        ALTER TABLE mp_plugins ADD COLUMN author VARCHAR(100);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='price') THEN
        ALTER TABLE mp_plugins ADD COLUMN price DECIMAL(10,2) DEFAULT 0.00;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='download_count') THEN
        ALTER TABLE mp_plugins ADD COLUMN download_count INT DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='rating') THEN
        ALTER TABLE mp_plugins ADD COLUMN rating DECIMAL(3,2) DEFAULT 0.00;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='review_count') THEN
        ALTER TABLE mp_plugins ADD COLUMN review_count INT DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='is_active') THEN
        ALTER TABLE mp_plugins ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='created_at') THEN
        ALTER TABLE mp_plugins ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='mp_plugins' AND column_name='updated_at') THEN
        ALTER TABLE mp_plugins ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
    END IF;
END $$;

-- 插入一些测试数据（如果不存在）
INSERT INTO mp_categories (id, name, description, icon, sort_order, is_active)
VALUES
  ('11111111-1111-1111-1111-111111111111', '开发工具', '提高开发效率的工具', 'Code', 1, TRUE),
  ('22222222-2222-2222-2222-222222222222', '主题模板', '美化界面的主题', 'Palette', 2, TRUE),
  ('33333333-3333-3333-3333-333333333333', '实用插件', '日常使用的实用功能', 'Tool', 3, TRUE)
ON CONFLICT (id) DO NOTHING;