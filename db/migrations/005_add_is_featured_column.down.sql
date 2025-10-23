-- Remove is_featured column from mp_plugins table

DROP INDEX IF EXISTS idx_mp_plugins_is_featured;
ALTER TABLE mp_plugins DROP COLUMN IF EXISTS is_featured;