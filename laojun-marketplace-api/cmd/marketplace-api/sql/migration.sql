-- Marketplace mp_* schema migration (v1)
-- Creates core user, forum, blog, code, and like tables

-- Ensure UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

BEGIN;

-- 创建市场用户表 (继承自用户认证域)
-- 注意：这个表可能需要与 ua_users 表进行关联，而不是重复创建
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

-- 创建论坛分类表 (属于社区功能域)
CREATE TABLE IF NOT EXISTS cm_forum_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon VARCHAR(100),
  sort_order INT DEFAULT 0,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 创建论坛帖子表 (属于社区功能域)
CREATE TABLE IF NOT EXISTS cm_forum_posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id UUID NOT NULL REFERENCES cm_forum_categories(id),
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
CREATE INDEX IF NOT EXISTS idx_cm_forum_posts_category ON cm_forum_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_cm_forum_posts_created_at ON cm_forum_posts(created_at DESC);

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

-- 创建插件分类表
CREATE TABLE IF NOT EXISTS mp_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon VARCHAR(100),
  color VARCHAR(50),
  sort_order INT DEFAULT 0,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 创建插件表
CREATE TABLE IF NOT EXISTS mp_plugins (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(200) NOT NULL,
  description TEXT,
  short_description TEXT,
  author VARCHAR(100),
  developer_id UUID REFERENCES mp_users(id),
  version VARCHAR(50),
  icon_url VARCHAR(255),
  banner_url VARCHAR(255),
  price DECIMAL(10,2) DEFAULT 0.00,
  rating DECIMAL(3,2) DEFAULT 0.00,
  download_count INT DEFAULT 0,
  is_featured BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  category_id UUID REFERENCES mp_categories(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_category ON mp_plugins(category_id);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_featured ON mp_plugins(is_featured);
CREATE INDEX IF NOT EXISTS idx_mp_plugins_rating ON mp_plugins(rating DESC);

-- 创建插件收藏表
CREATE TABLE IF NOT EXISTS mp_favorites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
  plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_mp_favorites_user_plugin ON mp_favorites(user_id, plugin_id);

-- 创建插件购买表
CREATE TABLE IF NOT EXISTS mp_purchases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
  plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
  amount DECIMAL(10,2) NOT NULL,
  status VARCHAR(50) DEFAULT 'completed',
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mp_purchases_user ON mp_purchases(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_purchases_plugin ON mp_purchases(plugin_id);

-- Seed data: one user, categories, a post, and a reply
INSERT INTO mp_users (id, username, email, password_hash, full_name, avatar, avatar_url, bio)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'testuser',
  'test@example.com',
  '$2a$10$KnZfQwKk7QnQkO8k5HqY7eHkUO8Ck8s4dEwUQJcKcVqVbYI2HqG3a', -- bcrypt placeholder
  '测试用户',
  'https://cdn.example.com/avatar/test.png',
  'https://cdn.example.com/avatar/test.png',
  '这是一个用于测试的用户'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO mp_forum_categories (id, name, description, icon, sort_order, is_active)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '通用讨论', '通用话题与交流', 'MessageSquare', 1, TRUE),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '问答互助', '提问与解答', 'HelpCircle', 2, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO mp_forum_posts (id, category_id, user_id, title, content, likes_count, replies_count, views_count, is_pinned, is_locked)
VALUES (
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  '11111111-1111-1111-1111-111111111111',
  '欢迎来到论坛',
  '这是一个测试帖子，用于验证详情页面与服务端接口。\n\n请在下方回复，参与讨论！',
  3, 1, 12, FALSE, FALSE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO mp_forum_replies (id, post_id, user_id, content, likes_count)
VALUES (
  'dddddddd-dddd-dddd-dddd-dddddddddddd',
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  '11111111-1111-1111-1111-111111111111',
  '这是第一条测试回复。',
  1
) ON CONFLICT (id) DO NOTHING;

-- 插入插件分类示例数据
INSERT INTO mp_categories (id, name, description, icon, color, sort_order, is_active)
VALUES
  ('11111111-2222-3333-4444-555555555555', '开发工具', '提升开发效率的工具插件', 'Code', '#3B82F6', 1, TRUE),
  ('22222222-3333-4444-5555-666666666666', '数据处理', '数据分析和处理相关插件', 'Database', '#10B981', 2, TRUE),
  ('33333333-4444-5555-6666-777777777777', '图像处理', '图像编辑和处理工具', 'Image', '#F59E0B', 3, TRUE),
  ('44444444-5555-6666-7777-888888888888', '文本分析', '文本处理和分析工具', 'FileText', '#8B5CF6', 4, TRUE),
  ('55555555-6666-7777-8888-999999999999', 'API连接器', 'API集成和连接工具', 'Link', '#EF4444', 5, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 插入插件示例数据
INSERT INTO mp_plugins (id, name, description, short_description, author, developer_id, version, icon_url, price, rating, download_count, is_featured, category_id)
VALUES
  ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', 'Code Formatter Pro', '强大的代码格式化工具，支持多种编程语言的自动格式化和代码美化。', '专业的代码格式化工具', 'DevTools Team', '11111111-1111-1111-1111-111111111111', '1.2.0', 'https://cdn.example.com/icons/code-formatter.png', 29.99, 4.8, 1250, TRUE, '11111111-2222-3333-4444-555555555555'),
  ('bbbbbbbb-cccc-dddd-eeee-ffffffffffff', 'Data Analyzer', '高效的数据分析插件，提供数据可视化、统计分析和报表生成功能。', '数据分析和可视化工具', 'Analytics Pro', '11111111-1111-1111-1111-111111111111', '2.1.5', 'https://cdn.example.com/icons/data-analyzer.png', 49.99, 4.6, 890, TRUE, '22222222-3333-4444-5555-666666666666'),
  ('cccccccc-dddd-eeee-ffff-000000000000', 'Image Filter Studio', '专业的图像滤镜工具，提供丰富的滤镜效果和图像处理功能。', '图像滤镜和处理工具', 'ImageTech', '11111111-1111-1111-1111-111111111111', '1.8.3', 'https://cdn.example.com/icons/image-filter.png', 19.99, 4.5, 2100, FALSE, '33333333-4444-5555-6666-777777777777'),
  ('dddddddd-eeee-ffff-0000-111111111111', 'Text Sentiment Analyzer', '智能文本情感分析工具，支持多语言文本的情感识别和分析。', '文本情感分析工具', 'AI Text Labs', '11111111-1111-1111-1111-111111111111', '1.5.2', 'https://cdn.example.com/icons/text-analyzer.png', 39.99, 4.7, 650, TRUE, '44444444-5555-6666-7777-888888888888'),
  ('eeeeeeee-ffff-0000-1111-222222222222', 'REST API Connector', '通用的REST API连接器，简化API集成和数据交换。', 'REST API集成工具', 'API Solutions', '11111111-1111-1111-1111-111111111111', '3.0.1', 'https://cdn.example.com/icons/api-connector.png', 0.00, 4.4, 3200, FALSE, '55555555-6666-7777-8888-999999999999'),
  ('ffffffff-0000-1111-2222-333333333333', 'Quick Debugger', '快速调试工具，提供断点设置、变量监控和性能分析功能。', '快速调试和性能分析', 'Debug Masters', '11111111-1111-1111-1111-111111111111', '2.3.0', 'https://cdn.example.com/icons/debugger.png', 24.99, 4.9, 1800, TRUE, '11111111-2222-3333-4444-555555555555'),
  ('00000000-1111-2222-3333-444444444444', 'CSV Data Processor', '专业的CSV数据处理工具，支持大文件处理和数据转换。', 'CSV数据处理和转换', 'Data Tools Inc', '11111111-1111-1111-1111-111111111111', '1.4.7', 'https://cdn.example.com/icons/csv-processor.png', 15.99, 4.3, 950, FALSE, '22222222-3333-4444-5555-666666666666'),
  ('11111111-2222-3333-4444-555555555556', 'Photo Enhancer', '智能照片增强工具，自动优化照片质量和色彩。', '智能照片增强和优化', 'PhotoAI', '11111111-1111-1111-1111-111111111111', '1.9.1', 'https://cdn.example.com/icons/photo-enhancer.png', 34.99, 4.6, 1400, TRUE, '33333333-4444-5555-6666-777777777777')
ON CONFLICT (id) DO NOTHING;

COMMIT;