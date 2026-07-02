CREATE TYPE platform_admin_role AS ENUM ('super_admin', 'support');

CREATE TABLE public.platform_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Case-insensitive email enforcement via CITEXT or simple CHECK constraint. 
    -- Using CHECK here for standard VARCHAR compatibility.
    email VARCHAR(255) NOT NULL UNIQUE CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
    
    password_hash VARCHAR(255) NOT NULL,
    role platform_admin_role NOT NULL DEFAULT 'support',
    
    -- Nullable because 2FA might be set up post-registration (though enforced later for super_admin)
    "2fa_secret" VARCHAR(255),
    
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);