-- Rollback marketplace tables

-- Drop tables in reverse order to handle foreign key constraints
DROP TABLE IF EXISTS mp_likes;
DROP TABLE IF EXISTS mp_code_snippets;
DROP TABLE IF EXISTS mp_blog_posts;
DROP TABLE IF EXISTS mp_blog_categories;
DROP TABLE IF EXISTS mp_forum_replies;
DROP TABLE IF EXISTS mp_forum_posts;
DROP TABLE IF EXISTS mp_forum_categories;
DROP TABLE IF EXISTS mp_users;