-- Description: Create configs table for configuration management
-- +migrate Up

CREATE TABLE configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'string' CHECK (type IN ('string', 'integer', 'float', 'boolean', 'json', 'array')),
    namespace VARCHAR(100) DEFAULT 'default',
    environment VARCHAR(50) DEFAULT 'production' CHECK (environment IN ('development', 'testing', 'staging', 'production')),
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    is_encrypted BOOLEAN DEFAULT FALSE,
    version INTEGER DEFAULT 1,
    tags TEXT[],
    validation_rules JSONB,
    default_value TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(key, namespace, environment)
);

-- Create indexes for better performance
CREATE INDEX idx_configs_key ON configs(key);
CREATE INDEX idx_configs_namespace ON configs(namespace);
CREATE INDEX idx_configs_environment ON configs(environment);
CREATE INDEX idx_configs_type ON configs(type);
CREATE INDEX idx_configs_is_public ON configs(is_public);
CREATE INDEX idx_configs_created_by ON configs(created_by);
CREATE INDEX idx_configs_created_at ON configs(created_at);
CREATE INDEX idx_configs_key_namespace_env ON configs(key, namespace, environment);
CREATE INDEX idx_configs_tags ON configs USING GIN(tags);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_configs_updated_at 
    BEFORE UPDATE ON configs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create config history table for audit trail
CREATE TABLE config_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_id UUID NOT NULL REFERENCES configs(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    old_type VARCHAR(50),
    new_type VARCHAR(50),
    action VARCHAR(20) NOT NULL CHECK (action IN ('create', 'update', 'delete')),
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_reason TEXT,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for config history
CREATE INDEX idx_config_history_config_id ON config_history(config_id);
CREATE INDEX idx_config_history_changed_by ON config_history(changed_by);
CREATE INDEX idx_config_history_changed_at ON config_history(changed_at);
CREATE INDEX idx_config_history_action ON config_history(action);

-- +migrate Down

DROP INDEX IF EXISTS idx_config_history_action;
DROP INDEX IF EXISTS idx_config_history_changed_at;
DROP INDEX IF EXISTS idx_config_history_changed_by;
DROP INDEX IF EXISTS idx_config_history_config_id;
DROP TABLE IF EXISTS config_history;

DROP TRIGGER IF EXISTS update_configs_updated_at ON configs;
DROP INDEX IF EXISTS idx_configs_tags;
DROP INDEX IF EXISTS idx_configs_key_namespace_env;
DROP INDEX IF EXISTS idx_configs_created_at;
DROP INDEX IF EXISTS idx_configs_created_by;
DROP INDEX IF EXISTS idx_configs_is_public;
DROP INDEX IF EXISTS idx_configs_type;
DROP INDEX IF EXISTS idx_configs_environment;
DROP INDEX IF EXISTS idx_configs_namespace;
DROP INDEX IF EXISTS idx_configs_key;
DROP TABLE IF EXISTS configs;