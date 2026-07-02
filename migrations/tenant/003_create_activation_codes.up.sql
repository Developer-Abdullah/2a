CREATE TYPE activation_code_type AS ENUM ('time_bound', 'usage_count', 'device_bound', 'reseller_bulk');
CREATE TYPE allowed_device_type AS ENUM ('iphone', 'ipad', 'both');

CREATE TABLE activation_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    code VARCHAR(100) NOT NULL UNIQUE,
    type activation_code_type NOT NULL,
    device_type allowed_device_type NOT NULL DEFAULT 'both',
    
    max_devices INT NOT NULL DEFAULT 1 CHECK (max_devices >= 1),
    current_device_count INT NOT NULL DEFAULT 0 CHECK (current_device_count >= 0),
    
    max_uses INT NOT NULL DEFAULT 1 CHECK (max_uses >= 1),
    current_uses INT NOT NULL DEFAULT 0 CHECK (current_uses >= 0),
    
    expires_at TIMESTAMPTZ,
    first_used_at TIMESTAMPTZ,
    
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    revoked_at TIMESTAMPTZ,
    
    notes TEXT,
    
    -- DEFERRED: The FOREIGN KEY to admin_users will be attached in 013_create_admin_users.sql
    created_by_admin_id UUID
);

-- Establish the deferred Foreign Key from 001_create_users.sql.
-- ON DELETE SET NULL prevents deleting users if a code is purged (though codes are generally soft-revoked).
ALTER TABLE users
    ADD CONSTRAINT fk_users_activation_code_id
    FOREIGN KEY (activation_code_id) REFERENCES activation_codes(id)
    ON DELETE SET NULL;
