
CREATE TYPE clone_status AS ENUM ('pending', 'signing', 'active', 'revoked');

CREATE TABLE clone_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    original_app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    original_version_id UUID NOT NULL REFERENCES application_versions(id) ON DELETE CASCADE,
    
    -- The new bundle ID generated for this clone (e.g., com.tenant.clone.abc123xyz)
    cloned_bundle_id VARCHAR(255) NOT NULL,
    
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    
    -- Links to the signing job. SET NULL ensures the clone record history survives 
    -- even if old signing jobs are purged from the DB.
    signing_job_id UUID REFERENCES signing_jobs(id) ON DELETE SET NULL,
    
    pin_verified_at TIMESTAMPTZ,
    status clone_status NOT NULL DEFAULT 'pending',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes to look up clones by user or device
CREATE INDEX idx_clone_records_user_id ON clone_records(user_id);
CREATE INDEX idx_clone_records_device_id ON clone_records(device_id);
