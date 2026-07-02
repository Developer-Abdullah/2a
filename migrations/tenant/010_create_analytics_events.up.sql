
CREATE TYPE analytics_event_type AS ENUM ('install', 'update', 'open', 'error', 'download');

-- Create the base partitioned table
CREATE TABLE analytics_events (
    id UUID DEFAULT gen_random_uuid(),
    
    -- Not using explicit FOREIGN KEY constraints here to ensure maximum 
    -- write throughput for high-volume analytics and to prevent lock contention.
    user_id UUID,
    device_id UUID,
    application_id UUID,
    version_id UUID,
    
    event_type analytics_event_type NOT NULL,
    metadata JSONB,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- CRITICAL: PostgreSQL requires the partition key to be part of the Primary Key
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create a default partition. If pg_partman falls behind on creating specific 
-- monthly/daily partitions, events will safely land here instead of rejecting the insert.
CREATE TABLE analytics_events_default PARTITION OF analytics_events DEFAULT;

-- Indexes applied to the parent table automatically propagate to all partitions
CREATE INDEX idx_analytics_events_app_type_date ON analytics_events(application_id, event_type, created_at DESC);
CREATE INDEX idx_analytics_events_device_date ON analytics_events(device_id, created_at DESC);

