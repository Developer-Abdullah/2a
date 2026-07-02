
CREATE TABLE udid_enrollment_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- CRITICAL CROSS-SCHEMA FK: Ties this session directly to the platform tenant
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    
    one_time_token VARCHAR(64) NOT NULL UNIQUE,
    udid_hash VARCHAR(255),
    
    -- Reuses the device_platform_type ENUM defined in 002_create_devices.sql
    device_type device_platform_type,
    
    ip_address INET,
    completed BOOLEAN NOT NULL DEFAULT false,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- Optimize callback lookups where speed is critical to prevent Apple MDM timeout
CREATE INDEX idx_udid_sessions_token ON udid_enrollment_sessions(one_time_token);

-- Optimize the background cleanup worker querying for expired sessions
CREATE INDEX idx_udid_sessions_expires_at ON udid_enrollment_sessions(expires_at);