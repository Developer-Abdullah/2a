
CREATE TABLE platform_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Nullable because unactivated users (e.g., on the enrollment page) can leave ratings
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment VARCHAR(280),
    
    ip_hash VARCHAR(255) NOT NULL,
    
    -- Strict uniqueness: One rating per device physically bound to the ecosystem
    device_id UUID UNIQUE REFERENCES devices(id) ON DELETE SET NULL,
    
    rating_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- CRITICAL SPAM PREVENTION: Unique index blocking an IP from submitting more than one rating per day
CREATE UNIQUE INDEX idx_unique_ip_daily ON platform_ratings (ip_hash, rating_date);

-- Optimize rating aggregations (AVG, COUNT)
CREATE INDEX idx_platform_ratings_metrics ON platform_ratings(rating, created_at DESC);
