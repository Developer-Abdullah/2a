
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- We map this as a UUID for now. The actual FOREIGN KEY constraint 
    -- will be added in 003_create_activation_codes to avoid a circular dependency.
    activation_code_id UUID,
    
    display_name VARCHAR(255),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ
);

-- Index for querying users by their activation code
CREATE INDEX idx_users_activation_code_id ON users(activation_code_id);