-- Description: Create system settings table for application configuration
-- +migrate Up

CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category VARCHAR(100) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    data_type VARCHAR(50) DEFAULT 'string' CHECK (data_type IN ('string', 'integer', 'float', 'boolean', 'json', 'array', 'text')),
    is_public BOOLEAN DEFAULT FALSE,
    is_readonly BOOLEAN DEFAULT FALSE,
    is_encrypted BOOLEAN DEFAULT FALSE,
    validation_rules JSONB,
    default_value TEXT,
    description TEXT,
    display_name VARCHAR(255),
    display_order INTEGER DEFAULT 0,
    group_name VARCHAR(100),
    help_text TEXT,
    options JSONB, -- For dropdown/select options
    min_value NUMERIC,
    max_value NUMERIC,
    pattern VARCHAR(500), -- Regex pattern for validation
    required BOOLEAN DEFAULT FALSE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(category, key)
);

-- Create indexes for better performance
CREATE INDEX idx_system_settings_category ON system_settings(category);
CREATE INDEX idx_system_settings_key ON system_settings(key);
CREATE INDEX idx_system_settings_is_public ON system_settings(is_public);
CREATE INDEX idx_system_settings_is_readonly ON system_settings(is_readonly);
CREATE INDEX idx_system_settings_group_name ON system_settings(group_name);
CREATE INDEX idx_system_settings_display_order ON system_settings(display_order);
CREATE INDEX idx_system_settings_category_key ON system_settings(category, key);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_system_settings_updated_at 
    BEFORE UPDATE ON system_settings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create system settings history table
CREATE TABLE system_settings_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_id UUID NOT NULL REFERENCES system_settings(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL,
    key VARCHAR(255) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    action VARCHAR(20) NOT NULL CHECK (action IN ('create', 'update', 'delete')),
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    change_reason TEXT,
    ip_address INET,
    user_agent TEXT
);

-- Create indexes for system settings history
CREATE INDEX idx_system_settings_history_setting_id ON system_settings_history(setting_id);
CREATE INDEX idx_system_settings_history_category ON system_settings_history(category);
CREATE INDEX idx_system_settings_history_changed_by ON system_settings_history(changed_by);
CREATE INDEX idx_system_settings_history_changed_at ON system_settings_history(changed_at);
CREATE INDEX idx_system_settings_history_action ON system_settings_history(action);

-- Insert default system settings
INSERT INTO system_settings (category, key, value, data_type, is_public, description, display_name, group_name, display_order) VALUES
-- Application Settings
('application', 'app_name', 'Laojun', 'string', true, 'Application name', 'Application Name', 'General', 1),
('application', 'app_version', '1.0.0', 'string', true, 'Application version', 'Version', 'General', 2),
('application', 'app_description', 'Microservices configuration and plugin management platform', 'text', true, 'Application description', 'Description', 'General', 3),
('application', 'app_url', 'https://laojun.example.com', 'string', true, 'Application base URL', 'Base URL', 'General', 4),
('application', 'maintenance_mode', 'false', 'boolean', true, 'Enable maintenance mode', 'Maintenance Mode', 'General', 5),
('application', 'debug_mode', 'false', 'boolean', false, 'Enable debug mode', 'Debug Mode', 'Development', 1),

-- Security Settings
('security', 'session_timeout', '3600', 'integer', false, 'Session timeout in seconds', 'Session Timeout (seconds)', 'Authentication', 1),
('security', 'max_login_attempts', '5', 'integer', false, 'Maximum login attempts before lockout', 'Max Login Attempts', 'Authentication', 2),
('security', 'lockout_duration', '900', 'integer', false, 'Account lockout duration in seconds', 'Lockout Duration (seconds)', 'Authentication', 3),
('security', 'password_min_length', '8', 'integer', false, 'Minimum password length', 'Min Password Length', 'Password Policy', 1),
('security', 'password_require_uppercase', 'true', 'boolean', false, 'Require uppercase letters in password', 'Require Uppercase', 'Password Policy', 2),
('security', 'password_require_lowercase', 'true', 'boolean', false, 'Require lowercase letters in password', 'Require Lowercase', 'Password Policy', 3),
('security', 'password_require_numbers', 'true', 'boolean', false, 'Require numbers in password', 'Require Numbers', 'Password Policy', 4),
('security', 'password_require_symbols', 'true', 'boolean', false, 'Require symbols in password', 'Require Symbols', 'Password Policy', 5),
('security', 'two_factor_enabled', 'false', 'boolean', false, 'Enable two-factor authentication', 'Two-Factor Authentication', 'Authentication', 4),

-- API Settings
('api', 'rate_limit_enabled', 'true', 'boolean', false, 'Enable API rate limiting', 'Rate Limiting', 'Rate Limiting', 1),
('api', 'rate_limit_requests', '1000', 'integer', false, 'Requests per hour per user', 'Requests per Hour', 'Rate Limiting', 2),
('api', 'rate_limit_burst', '100', 'integer', false, 'Burst limit for rate limiting', 'Burst Limit', 'Rate Limiting', 3),
('api', 'cors_enabled', 'true', 'boolean', false, 'Enable CORS', 'CORS Enabled', 'CORS', 1),
('api', 'cors_origins', '["*"]', 'json', false, 'Allowed CORS origins', 'Allowed Origins', 'CORS', 2),

-- Email Settings
('email', 'smtp_host', '', 'string', false, 'SMTP server host', 'SMTP Host', 'SMTP', 1),
('email', 'smtp_port', '587', 'integer', false, 'SMTP server port', 'SMTP Port', 'SMTP', 2),
('email', 'smtp_username', '', 'string', false, 'SMTP username', 'SMTP Username', 'SMTP', 3),
('email', 'smtp_password', '', 'string', false, 'SMTP password', 'SMTP Password', 'SMTP', 4),
('email', 'smtp_encryption', 'tls', 'string', false, 'SMTP encryption method', 'Encryption', 'SMTP', 5),
('email', 'from_email', 'noreply@laojun.example.com', 'string', false, 'Default from email address', 'From Email', 'General', 1),
('email', 'from_name', 'Laojun Platform', 'string', false, 'Default from name', 'From Name', 'General', 2),

-- Storage Settings
('storage', 'default_driver', 'local', 'string', false, 'Default storage driver', 'Default Driver', 'General', 1),
('storage', 'max_file_size', '10485760', 'integer', false, 'Maximum file size in bytes (10MB)', 'Max File Size (bytes)', 'Limits', 1),
('storage', 'allowed_file_types', '["jpg", "jpeg", "png", "gif", "pdf", "doc", "docx", "zip"]', 'json', false, 'Allowed file types', 'Allowed File Types', 'Limits', 2),

-- Cache Settings
('cache', 'default_ttl', '3600', 'integer', false, 'Default cache TTL in seconds', 'Default TTL (seconds)', 'General', 1),
('cache', 'enabled', 'true', 'boolean', false, 'Enable caching', 'Cache Enabled', 'General', 2),

-- Logging Settings
('logging', 'level', 'info', 'string', false, 'Log level', 'Log Level', 'General', 1),
('logging', 'max_file_size', '100', 'integer', false, 'Maximum log file size in MB', 'Max File Size (MB)', 'Files', 1),
('logging', 'max_files', '10', 'integer', false, 'Maximum number of log files to keep', 'Max Files', 'Files', 2),

-- Plugin Settings
('plugins', 'auto_update_enabled', 'false', 'boolean', false, 'Enable automatic plugin updates', 'Auto Update', 'Updates', 1),
('plugins', 'update_check_interval', '86400', 'integer', false, 'Update check interval in seconds', 'Check Interval (seconds)', 'Updates', 2),
('plugins', 'allow_beta_versions', 'false', 'boolean', false, 'Allow beta version plugins', 'Allow Beta Versions', 'General', 1),

-- Monitoring Settings
('monitoring', 'metrics_enabled', 'true', 'boolean', false, 'Enable metrics collection', 'Metrics Enabled', 'General', 1),
('monitoring', 'health_check_interval', '30', 'integer', false, 'Health check interval in seconds', 'Health Check Interval', 'Health Checks', 1),
('monitoring', 'alert_email', '', 'string', false, 'Email for system alerts', 'Alert Email', 'Alerts', 1);

-- Create function to get setting value with type conversion
CREATE OR REPLACE FUNCTION get_setting(setting_category VARCHAR, setting_key VARCHAR)
RETURNS TEXT AS $$
DECLARE
    setting_value TEXT;
BEGIN
    SELECT value INTO setting_value
    FROM system_settings
    WHERE category = setting_category AND key = setting_key;
    
    RETURN setting_value;
END;
$$ LANGUAGE plpgsql;

-- Create function to update setting value
CREATE OR REPLACE FUNCTION update_setting(
    setting_category VARCHAR, 
    setting_key VARCHAR, 
    new_value TEXT,
    user_id UUID DEFAULT NULL,
    reason TEXT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    old_value TEXT;
    setting_exists BOOLEAN;
BEGIN
    -- Check if setting exists and get old value
    SELECT value, TRUE INTO old_value, setting_exists
    FROM system_settings
    WHERE category = setting_category AND key = setting_key;
    
    IF NOT setting_exists THEN
        RETURN FALSE;
    END IF;
    
    -- Update the setting
    UPDATE system_settings
    SET value = new_value, updated_by = user_id, updated_at = CURRENT_TIMESTAMP
    WHERE category = setting_category AND key = setting_key;
    
    -- Log the change
    INSERT INTO system_settings_history (
        setting_id, category, key, old_value, new_value, action, changed_by, change_reason
    )
    SELECT id, setting_category, setting_key, old_value, new_value, 'update', user_id, reason
    FROM system_settings
    WHERE category = setting_category AND key = setting_key;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- +migrate Down

DROP FUNCTION IF EXISTS update_setting(VARCHAR, VARCHAR, TEXT, UUID, TEXT);
DROP FUNCTION IF EXISTS get_setting(VARCHAR, VARCHAR);

DROP INDEX IF EXISTS idx_system_settings_history_action;
DROP INDEX IF EXISTS idx_system_settings_history_changed_at;
DROP INDEX IF EXISTS idx_system_settings_history_changed_by;
DROP INDEX IF EXISTS idx_system_settings_history_category;
DROP INDEX IF EXISTS idx_system_settings_history_setting_id;
DROP TABLE IF EXISTS system_settings_history;

DROP TRIGGER IF EXISTS update_system_settings_updated_at ON system_settings;
DROP INDEX IF EXISTS idx_system_settings_category_key;
DROP INDEX IF EXISTS idx_system_settings_display_order;
DROP INDEX IF EXISTS idx_system_settings_group_name;
DROP INDEX IF EXISTS idx_system_settings_is_readonly;
DROP INDEX IF EXISTS idx_system_settings_is_public;
DROP INDEX IF EXISTS idx_system_settings_key;
DROP INDEX IF EXISTS idx_system_settings_category;
DROP TABLE IF EXISTS system_settings;