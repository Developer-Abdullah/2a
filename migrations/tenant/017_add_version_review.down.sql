DROP TRIGGER IF EXISTS trg_enforce_version_review_transition ON application_versions;
DROP FUNCTION IF EXISTS enforce_version_review_transition();
DROP INDEX IF EXISTS idx_app_versions_review_status;

ALTER TABLE application_versions
    DROP COLUMN IF EXISTS review_status,
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at;

DROP TYPE IF EXISTS version_review_status;
