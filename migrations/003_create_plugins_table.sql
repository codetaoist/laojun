-- Description: Create plugins table for plugin marketplace
-- +migrate Up

CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    long_description TEXT,
    author VARCHAR(255) NOT NULL,
    author_email VARCHAR(255),
    author_url VARCHAR(500),
    category VARCHAR(100) NOT NULL,
    subcategory VARCHAR(100),
    tags TEXT[],
    keywords TEXT[],
    download_url VARCHAR(500) NOT NULL,
    documentation_url VARCHAR(500),
    repository_url VARCHAR(500),
    homepage_url VARCHAR(500),
    license VARCHAR(100) NOT NULL,
    license_url VARCHAR(500),
    file_size BIGINT DEFAULT 0,
    file_hash VARCHAR(255),
    checksum VARCHAR(255),
    is_featured BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    is_premium BOOLEAN DEFAULT FALSE,
    price DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'inactive', 'deprecated', 'banned')),
    download_count INTEGER DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER DEFAULT 0,
    min_version VARCHAR(50),
    max_version VARCHAR(50),
    compatibility JSONB DEFAULT '{}',
    requirements JSONB DEFAULT '{}',
    configuration JSONB DEFAULT '{}',
    screenshots TEXT[],
    changelog TEXT,
    installation_guide TEXT,
    usage_guide TEXT,
    api_documentation TEXT,
    support_email VARCHAR(255),
    support_url VARCHAR(500),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    published_at TIMESTAMP,
    deprecated_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(name, version)
);

-- Create indexes for better performance
CREATE INDEX idx_plugins_slug ON plugins(slug);
CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_author ON plugins(author);
CREATE INDEX idx_plugins_category ON plugins(category);
CREATE INDEX idx_plugins_status ON plugins(status);
CREATE INDEX idx_plugins_is_featured ON plugins(is_featured);
CREATE INDEX idx_plugins_is_verified ON plugins(is_verified);
CREATE INDEX idx_plugins_is_premium ON plugins(is_premium);
CREATE INDEX idx_plugins_created_by ON plugins(created_by);
CREATE INDEX idx_plugins_published_at ON plugins(published_at);
CREATE INDEX idx_plugins_download_count ON plugins(download_count);
CREATE INDEX idx_plugins_rating_average ON plugins(rating_average);
CREATE INDEX idx_plugins_tags ON plugins USING GIN(tags);
CREATE INDEX idx_plugins_keywords ON plugins USING GIN(keywords);
CREATE INDEX idx_plugins_name_version ON plugins(name, version);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_plugins_updated_at 
    BEFORE UPDATE ON plugins 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create plugin downloads table for tracking
CREATE TABLE plugin_downloads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    referer VARCHAR(500),
    download_source VARCHAR(100),
    downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    file_size BIGINT,
    download_duration INTEGER, -- in milliseconds
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT
);

-- Create indexes for plugin downloads
CREATE INDEX idx_plugin_downloads_plugin_id ON plugin_downloads(plugin_id);
CREATE INDEX idx_plugin_downloads_user_id ON plugin_downloads(user_id);
CREATE INDEX idx_plugin_downloads_downloaded_at ON plugin_downloads(downloaded_at);
CREATE INDEX idx_plugin_downloads_ip_address ON plugin_downloads(ip_address);
CREATE INDEX idx_plugin_downloads_success ON plugin_downloads(success);

-- Create plugin ratings table
CREATE TABLE plugin_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    title VARCHAR(255),
    helpful_count INTEGER DEFAULT 0,
    reported_count INTEGER DEFAULT 0,
    is_verified_purchase BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(plugin_id, user_id)
);

-- Create indexes for plugin ratings
CREATE INDEX idx_plugin_ratings_plugin_id ON plugin_ratings(plugin_id);
CREATE INDEX idx_plugin_ratings_user_id ON plugin_ratings(user_id);
CREATE INDEX idx_plugin_ratings_rating ON plugin_ratings(rating);
CREATE INDEX idx_plugin_ratings_created_at ON plugin_ratings(created_at);

-- Create trigger to update updated_at timestamp for ratings
CREATE TRIGGER update_plugin_ratings_updated_at 
    BEFORE UPDATE ON plugin_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create plugin categories table
CREATE TABLE plugin_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(255),
    color VARCHAR(7), -- hex color
    parent_id UUID REFERENCES plugin_categories(id) ON DELETE SET NULL,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for plugin categories
CREATE INDEX idx_plugin_categories_slug ON plugin_categories(slug);
CREATE INDEX idx_plugin_categories_parent_id ON plugin_categories(parent_id);
CREATE INDEX idx_plugin_categories_sort_order ON plugin_categories(sort_order);
CREATE INDEX idx_plugin_categories_is_active ON plugin_categories(is_active);

-- Create trigger to update updated_at timestamp for categories
CREATE TRIGGER update_plugin_categories_updated_at 
    BEFORE UPDATE ON plugin_categories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- +migrate Down

DROP TRIGGER IF EXISTS update_plugin_categories_updated_at ON plugin_categories;
DROP INDEX IF EXISTS idx_plugin_categories_is_active;
DROP INDEX IF EXISTS idx_plugin_categories_sort_order;
DROP INDEX IF EXISTS idx_plugin_categories_parent_id;
DROP INDEX IF EXISTS idx_plugin_categories_slug;
DROP TABLE IF EXISTS plugin_categories;

DROP TRIGGER IF EXISTS update_plugin_ratings_updated_at ON plugin_ratings;
DROP INDEX IF EXISTS idx_plugin_ratings_created_at;
DROP INDEX IF EXISTS idx_plugin_ratings_rating;
DROP INDEX IF EXISTS idx_plugin_ratings_user_id;
DROP INDEX IF EXISTS idx_plugin_ratings_plugin_id;
DROP TABLE IF EXISTS plugin_ratings;

DROP INDEX IF EXISTS idx_plugin_downloads_success;
DROP INDEX IF EXISTS idx_plugin_downloads_ip_address;
DROP INDEX IF EXISTS idx_plugin_downloads_downloaded_at;
DROP INDEX IF EXISTS idx_plugin_downloads_user_id;
DROP INDEX IF EXISTS idx_plugin_downloads_plugin_id;
DROP TABLE IF EXISTS plugin_downloads;

DROP TRIGGER IF EXISTS update_plugins_updated_at ON plugins;
DROP INDEX IF EXISTS idx_plugins_name_version;
DROP INDEX IF EXISTS idx_plugins_keywords;
DROP INDEX IF EXISTS idx_plugins_tags;
DROP INDEX IF EXISTS idx_plugins_rating_average;
DROP INDEX IF EXISTS idx_plugins_download_count;
DROP INDEX IF EXISTS idx_plugins_published_at;
DROP INDEX IF EXISTS idx_plugins_created_by;
DROP INDEX IF EXISTS idx_plugins_is_premium;
DROP INDEX IF EXISTS idx_plugins_is_verified;
DROP INDEX IF EXISTS idx_plugins_is_featured;
DROP INDEX IF EXISTS idx_plugins_status;
DROP INDEX IF EXISTS idx_plugins_category;
DROP INDEX IF EXISTS idx_plugins_author;
DROP INDEX IF EXISTS idx_plugins_name;
DROP INDEX IF EXISTS idx_plugins_slug;
DROP TABLE IF EXISTS plugins;