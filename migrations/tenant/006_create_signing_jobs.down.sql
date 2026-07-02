ALTER TABLE application_versions
    DROP CONSTRAINT IF EXISTS fk_app_versions_signing_job_id;

DROP TABLE IF EXISTS signing_jobs CASCADE;
DROP TYPE IF EXISTS signing_job_status;