-- Manual review gate: every uploaded version must be approved by the platform owner before its
-- signing job may run. This is an anti-piracy control — there is no auto-approval path.

CREATE TYPE version_review_status AS ENUM ('pending_review', 'approved', 'rejected');

ALTER TABLE application_versions
    ADD COLUMN review_status version_review_status NOT NULL DEFAULT 'pending_review',
    ADD COLUMN rejection_reason TEXT,
    ADD COLUMN reviewed_by UUID,
    ADD COLUMN reviewed_at TIMESTAMPTZ;

-- Fast lookup of the owner's review queue.
CREATE INDEX idx_app_versions_review_status
    ON application_versions(review_status, created_at DESC);

-- Defense in depth: enforce the legal review state machine at the database level so that even a
-- buggy or malicious query cannot skip review. Only pending_review -> approved / pending_review ->
-- rejected are allowed; approved and rejected are terminal. Updates that leave review_status
-- unchanged (e.g. the signing worker advancing signing_status) pass through untouched.
CREATE OR REPLACE FUNCTION enforce_version_review_transition()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.review_status IS DISTINCT FROM OLD.review_status THEN
        IF OLD.review_status <> 'pending_review' THEN
            RAISE EXCEPTION 'illegal review transition: version % is already %', OLD.id, OLD.review_status;
        END IF;
        IF NEW.review_status NOT IN ('approved', 'rejected') THEN
            RAISE EXCEPTION 'illegal review transition: % -> %', OLD.review_status, NEW.review_status;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_enforce_version_review_transition
    BEFORE UPDATE ON application_versions
    FOR EACH ROW
    EXECUTE FUNCTION enforce_version_review_transition();
