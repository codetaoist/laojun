-- Add is_featured column to mp_plugins table

ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS is_featured BOOLEAN DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_mp_plugins_is_featured ON mp_plugins(is_featured);