-- Add is_featured column to mp_plugins table

ALTER TABLE mp_plugins ADD COLUMN IF NOT EXISTS is_featured BOOLEAN DEFAULT FALSE;