-- Description: Create audit logs table for system activity tracking
-- +migrate Up

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    severity VARCHAR(20) DEFAULT 'info' CHECK (severity IN ('debug', 'info', 'warning', 'error', 'critical')),
    status VARCHAR(20) DEFAULT 'success' CHECK (status IN ('success', 'failure', 'pending')),
    error_message TEXT,
    ip_address INET,
    user_agent TEXT,
    endpoint VARCHAR(255),
    method VARCHAR(10),
    request_id VARCHAR(255),
    trace_id VARCHAR(255),
    span_id VARCHAR(255),
    duration INTEGER, -- in milliseconds
    tags TEXT[],
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Partitioning key for better performance
    partition_date DATE GENERATED ALWAYS AS (DATE(created_at)) STORED
);

-- Create indexes for better performance
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_session_id ON audit_logs(session_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_partition_date ON audit_logs(partition_date);
CREATE INDEX idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id);
CREATE INDEX idx_audit_logs_trace_id ON audit_logs(trace_id);
CREATE INDEX idx_audit_logs_tags ON audit_logs USING GIN(tags);

-- Create composite indexes for common queries
CREATE INDEX idx_audit_logs_user_action_date ON audit_logs(user_id, action, created_at);
CREATE INDEX idx_audit_logs_resource_action_date ON audit_logs(resource_type, action, created_at);

-- Create security events table for specific security-related logs
CREATE TABLE security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) DEFAULT 'warning' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    location_country VARCHAR(100),
    location_city VARCHAR(255),
    threat_level VARCHAR(20) DEFAULT 'low' CHECK (threat_level IN ('low', 'medium', 'high', 'critical')),
    is_blocked BOOLEAN DEFAULT FALSE,
    is_automated BOOLEAN DEFAULT FALSE,
    detection_method VARCHAR(100),
    rule_id VARCHAR(255),
    description TEXT,
    details JSONB DEFAULT '{}',
    remediation_action VARCHAR(255),
    remediation_status VARCHAR(50) DEFAULT 'pending',
    false_positive BOOLEAN DEFAULT FALSE,
    investigated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    investigated_at TIMESTAMP,
    investigation_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for security events
CREATE INDEX idx_security_events_event_type ON security_events(event_type);
CREATE INDEX idx_security_events_severity ON security_events(severity);
CREATE INDEX idx_security_events_user_id ON security_events(user_id);
CREATE INDEX idx_security_events_ip_address ON security_events(ip_address);
CREATE INDEX idx_security_events_threat_level ON security_events(threat_level);
CREATE INDEX idx_security_events_is_blocked ON security_events(is_blocked);
CREATE INDEX idx_security_events_created_at ON security_events(created_at);
CREATE INDEX idx_security_events_remediation_status ON security_events(remediation_status);

-- Create trigger to update updated_at timestamp for security events
CREATE TRIGGER update_security_events_updated_at 
    BEFORE UPDATE ON security_events 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to archive old audit logs
CREATE OR REPLACE FUNCTION archive_old_audit_logs(days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER;
    cutoff_date DATE;
BEGIN
    cutoff_date := CURRENT_DATE - INTERVAL '1 day' * days_to_keep;
    
    -- Create archive table if it doesn't exist
    CREATE TABLE IF NOT EXISTS audit_logs_archive (LIKE audit_logs INCLUDING ALL);
    
    -- Move old records to archive
    WITH moved_rows AS (
        DELETE FROM audit_logs 
        WHERE partition_date < cutoff_date
        RETURNING *
    )
    INSERT INTO audit_logs_archive SELECT * FROM moved_rows;
    
    GET DIAGNOSTICS archived_count = ROW_COUNT;
    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- Create function to get audit log statistics
CREATE OR REPLACE FUNCTION get_audit_log_stats(start_date DATE DEFAULT CURRENT_DATE - 7, end_date DATE DEFAULT CURRENT_DATE)
RETURNS TABLE (
    action VARCHAR(100),
    resource_type VARCHAR(100),
    count BIGINT,
    success_count BIGINT,
    failure_count BIGINT,
    avg_duration NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        al.action,
        al.resource_type,
        COUNT(*) as count,
        COUNT(*) FILTER (WHERE al.status = 'success') as success_count,
        COUNT(*) FILTER (WHERE al.status = 'failure') as failure_count,
        ROUND(AVG(al.duration), 2) as avg_duration
    FROM audit_logs al
    WHERE al.partition_date BETWEEN start_date AND end_date
    GROUP BY al.action, al.resource_type
    ORDER BY count DESC;
END;
$$ LANGUAGE plpgsql;

-- +migrate Down

DROP FUNCTION IF EXISTS get_audit_log_stats(DATE, DATE);
DROP FUNCTION IF EXISTS archive_old_audit_logs(INTEGER);

DROP TRIGGER IF EXISTS update_security_events_updated_at ON security_events;
DROP INDEX IF EXISTS idx_security_events_remediation_status;
DROP INDEX IF EXISTS idx_security_events_created_at;
DROP INDEX IF EXISTS idx_security_events_is_blocked;
DROP INDEX IF EXISTS idx_security_events_threat_level;
DROP INDEX IF EXISTS idx_security_events_ip_address;
DROP INDEX IF EXISTS idx_security_events_user_id;
DROP INDEX IF EXISTS idx_security_events_severity;
DROP INDEX IF EXISTS idx_security_events_event_type;
DROP TABLE IF EXISTS security_events;

DROP INDEX IF EXISTS idx_audit_logs_resource_action_date;
DROP INDEX IF EXISTS idx_audit_logs_user_action_date;
DROP INDEX IF EXISTS idx_audit_logs_tags;
DROP INDEX IF EXISTS idx_audit_logs_trace_id;
DROP INDEX IF EXISTS idx_audit_logs_request_id;
DROP INDEX IF EXISTS idx_audit_logs_ip_address;
DROP INDEX IF EXISTS idx_audit_logs_partition_date;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_status;
DROP INDEX IF EXISTS idx_audit_logs_severity;
DROP INDEX IF EXISTS idx_audit_logs_resource_id;
DROP INDEX IF EXISTS idx_audit_logs_resource_type;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_session_id;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP TABLE IF EXISTS audit_logs_archive;
DROP TABLE IF EXISTS audit_logs;