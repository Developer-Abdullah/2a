CREATE TYPE app_category AS ENUM (
    'exclusive', 'social_media', 'modified_apps', 'modified_games', 
    'paid_apps', 'paid_games', 'arcade_games', 'notifications', 
    'design', 'productivity', 'entertainment', 'sports', 
    'islamic', 'jailbreak', 'in_house', 'other'
);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    bundle_identifier VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    category app_category NOT NULL,
    
    description TEXT,
    
    -- Array of strings for localized bullet points
    features TEXT[],
    
    icon_s3_key VARCHAR(512),
    
    is_published BOOLEAN NOT NULL DEFAULT false,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for searching and filtering apps efficiently
CREATE INDEX idx_applications_category ON applications(category);
CREATE INDEX idx_applications_is_published ON applications(is_published);