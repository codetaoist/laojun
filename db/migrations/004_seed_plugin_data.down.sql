-- Remove plugin marketplace seed data

DELETE FROM mp_plugins WHERE id IN (
  'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
  'bbbbbbbb-cccc-dddd-eeee-ffffffffffff',
  'cccccccc-dddd-eeee-ffff-000000000000',
  'dddddddd-eeee-ffff-0000-111111111111',
  'eeeeeeee-ffff-0000-1111-222222222222',
  'ffffffff-0000-1111-2222-333333333333',
  '00000000-1111-2222-3333-444444444444'
);

DELETE FROM mp_categories WHERE id IN (
  '11111111-2222-3333-4444-555555555555',
  '22222222-3333-4444-5555-666666666666',
  '33333333-4444-5555-6666-777777777777',
  '44444444-5555-6666-7777-888888888888',
  '55555555-6666-7777-8888-999999999999'
);