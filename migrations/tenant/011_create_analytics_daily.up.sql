CREATE TABLE analytics_daily (
    date DATE NOT NULL,
    
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    version_id UUID NOT NULL REFERENCES application_versions(id) ON DELETE CASCADE,
    
    installs INT NOT NULL DEFAULT 0 CHECK (installs >= 0),
    updates INT NOT NULL DEFAULT 0 CHECK (updates >= 0),
    opens INT NOT NULL DEFAULT 0 CHECK (opens >= 0),
    errors INT NOT NULL DEFAULT 0 CHECK (errors >= 0),
    
    unique_users INT NOT NULL DEFAULT 0 CHECK (unique_users >= 0),
    unique_devices INT NOT NULL DEFAULT 0 CHECK (unique_devices >= 0),
    
    -- COMPOSITE PRIMARY KEY ensuring exactly one aggregated row per app version per day
    PRIMARY KEY (date, application_id, version_id)
);

-- Optimize querying historical aggregations for an application
CREATE INDEX idx_analytics_daily_app_date ON analytics_daily(application_id, date DESC);
