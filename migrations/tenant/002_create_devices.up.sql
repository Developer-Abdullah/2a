
CREATE TYPE enrollment_method_type AS ENUM ('fingerprint', 'udid_profile');
CREATE TYPE device_platform_type AS ENUM ('iphone', 'ipad');

CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    enrollment_method enrollment_method_type NOT NULL,
    
    -- Hashes are unique per tenant schema
    fingerprint_hash VARCHAR(255) UNIQUE,
    udid_hash VARCHAR(255) UNIQUE,
    
    device_type device_platform_type NOT NULL,
    model_string VARCHAR(100),
    
    last_seen_at TIMESTAMPTZ,
    last_seen_ip INET,
    
    jailbreak_flag BOOLEAN NOT NULL DEFAULT false,
    
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    revoked_at TIMESTAMPTZ,
    last_validated_at TIMESTAMPTZ,
    
    -- CRITICAL: Ensure the correct hash is populated based on the enrollment method,
    -- and that they are mutually exclusive.
    CONSTRAINT chk_enrollment_hash_integrity CHECK (
        (enrollment_method = 'fingerprint' AND fingerprint_hash IS NOT NULL AND udid_hash IS NULL) OR
        (enrollment_method = 'udid_profile' AND udid_hash IS NOT NULL AND fingerprint_hash IS NULL)
    )
);

-- Index for fast user device lookups
CREATE INDEX idx_devices_user_id ON devices(user_id);