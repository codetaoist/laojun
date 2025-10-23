-- Seed data for marketplace

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