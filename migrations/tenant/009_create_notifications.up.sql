
CREATE TYPE notification_type AS ENUM ('alert', 'silent', 'update');

-- 1. Create the base notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    
    target_app_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- DEFERRED: The FOREIGN KEY to admin_users will be attached in 013_create_admin_users.sql
    sent_by_admin_id UUID,
    
    type notification_type NOT NULL
);

-- 2. Create the pivot table for tracking per-user read state
CREATE TABLE user_notification_reads (
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- COMPOSITE PRIMARY KEY: A user can only read a specific notification once
    PRIMARY KEY (notification_id, user_id)
);

-- Index to optimize fetching paginated notification lists for a user
CREATE INDEX idx_notifications_sent_at ON notifications(sent_at DESC);
CREATE INDEX idx_user_notification_reads_user ON user_notification_reads(user_id);
