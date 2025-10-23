-- Remove seed data

DELETE FROM mp_forum_replies WHERE id = 'dddddddd-dddd-dddd-dddd-dddddddddddd';
DELETE FROM mp_forum_posts WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';
DELETE FROM mp_forum_categories WHERE id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');
DELETE FROM mp_users WHERE id = '11111111-1111-1111-1111-111111111111';