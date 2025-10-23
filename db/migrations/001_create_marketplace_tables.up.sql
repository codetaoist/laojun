-- Marketplace mp_* schema migration (v1)
-- Creates core user, forum, blog, code, and like tables

-- Ensure UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users
CREATE TABLE IF NOT EXISTS mp_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(100),
  avatar VARCHAR(255),
  avatar_url VARCHAR(255),
  bio TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  is_email_verified BOOLEAN DEFAULT FALSE,
  email_verification_token VARCHAR(255),
  password_reset_token VARCHAR(255),
  password_reset_expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  last_login_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_mp_users_username ON mp_users(username);
CREATE INDEX IF NOT EXISTS idx_mp_users_email ON mp_users(email);

-- Forum categories
CREATE TABLE IF NOT EXISTS mp_forum_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon VARCHAR(100),
  sort_order INT DEFAULT 0,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Forum posts
CREATE TABLE IF NOT EXISTS mp_forum_posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id UUID NOT NULL REFERENCES mp_forum_categories(id),
  user_id UUID NOT NULL REFERENCES mp_users(id),
  title VARCHAR(200) NOT NULL,
  content TEXT NOT NULL,
  likes_count INT DEFAULT 0,
  replies_count INT DEFAULT 0,
  views_count INT DEFAULT 0,
  is_pinned BOOLEAN DEFAULT FALSE,
  is_locked BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_category ON mp_forum_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_forum_posts_created_at ON mp_forum_posts(created_at DESC);

-- Forum replies
CREATE TABLE IF NOT EXISTS mp_forum_replies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID NOT NULL REFERENCES mp_forum_posts(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  likes_count INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mp_forum_replies_post ON mp_forum_replies(post_id);

-- Blog categories
CREATE TABLE IF NOT EXISTS mp_blog_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  color VARCHAR(50),
  sort_order INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Blog posts
CREATE TABLE IF NOT EXISTS mp_blog_posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id UUID REFERENCES mp_blog_categories(id),
  user_id UUID NOT NULL REFERENCES mp_users(id),
  title VARCHAR(200) NOT NULL,
  summary TEXT,
  content TEXT NOT NULL,
  cover_image VARCHAR(255),
  tags TEXT,
  likes_count INT DEFAULT 0,
  comments_count INT DEFAULT 0,
  views_count INT DEFAULT 0,
  is_published BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Code snippets
CREATE TABLE IF NOT EXISTS mp_code_snippets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES mp_users(id),
  title VARCHAR(200) NOT NULL,
  description TEXT,
  code TEXT NOT NULL,
  language VARCHAR(50),
  tags TEXT,
  likes_count INT DEFAULT 0,
  views_count INT DEFAULT 0,
  is_public BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mp_code_snippets_user ON mp_code_snippets(user_id);

-- Likes (generic for posts, replies, blog, code)
CREATE TABLE IF NOT EXISTS mp_likes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
  target_type VARCHAR(50) NOT NULL,
  target_id UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_mp_likes_user_target ON mp_likes(user_id, target_type, target_id);