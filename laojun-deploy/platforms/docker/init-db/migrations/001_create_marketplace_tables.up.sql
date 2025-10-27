-- Marketplace 相关表创建脚本

-- 用户表
CREATE TABLE IF NOT EXISTS mp_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    bio TEXT,
    website_url TEXT,
    github_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 论坛分类表
CREATE TABLE IF NOT EXISTS mp_forum_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(7),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 论坛帖子表
CREATE TABLE IF NOT EXISTS mp_forum_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES mp_forum_categories(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    likes_count INTEGER DEFAULT 0,
    views_count INTEGER DEFAULT 0,
    replies_count INTEGER DEFAULT 0,
    last_reply_at TIMESTAMP WITH TIME ZONE,
    last_reply_user_id UUID REFERENCES mp_users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_mp_users_username ON mp_users(username);
CREATE INDEX IF NOT EXISTS idx_mp_users_email ON mp_users(email);
CREATE INDEX IF NOT EXISTS idx_mp_users_is_active ON mp_users(is_active);

CREATE INDEX IF NOT EXISTS idx_mp_forum_categories_sort_order ON mp_forum_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_mp_forum_categories_is_active ON mp_forum_categories(is_active);

CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_category_id ON mp_forum_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_user_id ON mp_forum_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_created_at ON mp_forum_posts(created_at);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_is_pinned ON mp_forum_posts(is_pinned);

-- 插入迁移记录
INSERT INTO public.schema_migrations (version, dirty) 
VALUES ('001_create_marketplace_tables', FALSE) 
ON CONFLICT (version) DO NOTHING;