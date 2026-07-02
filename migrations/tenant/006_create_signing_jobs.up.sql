
CREATE TYPE signing_job_status AS ENUM ('queued', 'processing', 'completed', 'failed');

CREATE TABLE signing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    version_id UUID NOT NULL REFERENCES application_versions(id) ON DELETE CASCADE,
    
    -- CRITICAL CROSS-SCHEMA FK
    certificate_id UUID NOT NULL REFERENCES public.tenant_certificates(id) ON DELETE CASCADE,
    
    status signing_job_status NOT NULL DEFAULT 'queued',
    priority INT NOT NULL DEFAULT 5,
    
    queued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    
    retry_count INT NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    error_message TEXT,
    
    -- Tracks which worker pod/process picked up the job
    worker_id VARCHAR(255)
);

-- Index for queue administration and monitoring
CREATE INDEX idx_signing_jobs_status ON signing_jobs(status, priority DESC, queued_at ASC);

-- Establish the deferred Foreign Key from 005_create_application_versions.sql
-- Using ON DELETE SET NULL ensures if a job history is cleared, the app version record survives.
ALTER TABLE application_versions
    ADD CONSTRAINT fk_app_versions_signing_job_id
    FOREIGN KEY (signing_job_id) REFERENCES signing_jobs(id)
    ON DELETE SET NULL;