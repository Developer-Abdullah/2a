
CREATE TABLE updates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    from_version_id UUID NOT NULL REFERENCES application_versions(id) ON DELETE CASCADE,
    to_version_id UUID NOT NULL REFERENCES application_versions(id) ON DELETE CASCADE,
    
    -- Reusing the app_update_type ENUM created in 005_create_application_versions
    update_type app_update_type NOT NULL,
    
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to quickly fetch the update history of an application
CREATE INDEX idx_updates_application_id ON updates(application_id, published_at DESC);

