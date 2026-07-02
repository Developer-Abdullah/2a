CREATE TYPE app_update_type AS ENUM ('optional', 'forced', 'silent');
CREATE TYPE app_signing_status AS ENUM ('pending', 'signing', 'signed', 'failed');

CREATE TABLE application_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    
    version VARCHAR(50) NOT NULL,
    build_number VARCHAR(50) NOT NULL,
    release_notes TEXT,
    
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    
    -- S3 Storage keys
    raw_ipa_s3_key VARCHAR(512),
    signed_ipa_s3_key VARCHAR(512),
    manifest_s3_key VARCHAR(512),
    
    minimum_os_version VARCHAR(50),
    
    update_type app_update_type,
    signing_status app_signing_status NOT NULL DEFAULT 'pending',
    
    -- DEFERRED: The FOREIGN KEY to signing_jobs will be attached in 006_create_signing_jobs.sql
    signing_job_id UUID,
    
    -- CRITICAL CROSS-SCHEMA FK: Points to the public certificates table
    certificate_id UUID REFERENCES public.tenant_certificates(id) ON DELETE SET NULL,
    
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Optimize queries fetching version history for an app
CREATE INDEX idx_app_versions_application_id ON application_versions(application_id, created_at DESC);
