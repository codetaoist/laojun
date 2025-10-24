-- Create plugin marketplace tables

-- Plugin categories table
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

-- Plugins table
CREATE TABLE IF NOT EXISTS mp_plugins (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(200) NOT NULL,
  description TEXT,
  short_description VARCHAR(500),
  author VARCHAR(100),
  developer_id UUID REFERENCES mp_users(id),
  version VARCHAR(50),
  icon_url VARCHAR(255),
  price DECIMAL(10,2) DEFAULT 0.00,
  rating DECIMAL(3,2) DEFAULT 0.00,
  download_count INT DEFAULT 0,
  is_featured BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  category_id UUID REFERENCES mp_categories(id),
  tags TEXT,
  requirements TEXT,
  changelog TEXT,
  documentation_url VARCHAR(255),
  source_url VARCHAR(255),
  demo_url VARCHAR(255),
  status VARCHAR(50) DEFAULT 'active',
  review_status VARCHAR(50) DEFAULT 'pending',
  review_notes TEXT,
  reviewed_at TIMESTAMPTZ,
  reviewed_by UUID REFERENCES mp_users(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Plugin favorites table
CREATE TABLE IF NOT EXISTS mp_favorites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
  plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_mp_favorites_user_plugin ON mp_favorites(user_id, plugin_id);

-- Plugin purchases table
CREATE TABLE IF NOT EXISTS mp_purchases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES mp_users(id) ON DELETE CASCADE,
  plugin_id UUID NOT NULL REFERENCES mp_plugins(id) ON DELETE CASCADE,
  amount DECIMAL(10,2) NOT NULL,
  status VARCHAR(50) DEFAULT 'completed',
  transaction_id VARCHAR(255),
  payment_method VARCHAR(50),
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mp_purchases_user ON mp_purchases(user_id);
CREATE INDEX IF NOT EXISTS idx_mp_purchases_plugin ON mp_purchases(plugin_id);
CREATE INDEX IF NOT EXISTS idx_mp_purchases_status ON mp_purchases(status);
CREATE INDEX IF NOT EXISTS idx_mp_purchases_created_at ON mp_purchases(created_at DESC);