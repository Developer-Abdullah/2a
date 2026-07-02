
CREATE TYPE tenant_admin_role AS ENUM ('super_admin', 'admin', 'support');

CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    email VARCHAR(255) NOT NULL UNIQUE CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    password_hash VARCHAR(255) NOT NULL,
    role tenant_admin_role NOT NULL DEFAULT 'support',
    
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Establish DEFERRED Foreign Key from 003_create_activation_codes.sql
ALTER TABLE activation_codes
    ADD CONSTRAINT fk_activation_codes_admin_id
    FOREIGN KEY (created_by_admin_id) REFERENCES admin_users(id)
    ON DELETE SET NULL;

-- Establish DEFERRED Foreign Key from 009_create_notifications.sql
ALTER TABLE notifications
    ADD CONSTRAINT fk_notifications_admin_id
    FOREIGN KEY (sent_by_admin_id) REFERENCES admin_users(id)
    ON DELETE SET NULL;