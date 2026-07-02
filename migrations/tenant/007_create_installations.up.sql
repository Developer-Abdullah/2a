
CREATE TYPE install_type AS ENUM ('fresh', 'update', 'reinstall');

CREATE TABLE installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    version_id UUID NOT NULL REFERENCES application_versions(id) ON DELETE CASCADE,
    
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    install_type install_type NOT NULL
);

-- Indexes for fast analytics aggregation and user history lookups
CREATE INDEX idx_installations_user_id ON installations(user_id, installed_at DESC);
CREATE INDEX idx_installations_app_id ON installations(application_id, installed_at DESC);
CREATE INDEX idx_installations_device_id ON installations(device_id);
