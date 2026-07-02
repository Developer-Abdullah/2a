CREATE TABLE public.tenant_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Display name for the plan tier (e.g., "Basic", "Premium", "Enterprise")
    name VARCHAR(255) NOT NULL UNIQUE,
    
    -- Feature limits
    max_apps INT NOT NULL CHECK (max_apps >= 0),
    max_users INT NOT NULL CHECK (max_users >= 0),
    max_storage_gb NUMERIC(10, 2) NOT NULL CHECK (max_storage_gb >= 0),
    max_devices_per_code INT NOT NULL CHECK (max_devices_per_code >= 0),
    
    -- Queue priority weight for the Signing Worker (e.g., 10 for Premium, 5 for Basic)
    signing_priority INT NOT NULL DEFAULT 5,
    
    -- Feature toggles
    allows_clone BOOLEAN NOT NULL DEFAULT false,
    
    -- The schema isolation mode this plan utilizes (reuses existing ENUM from 001)
    isolation_mode tenant_isolation_mode NOT NULL,
    
    -- Specialized platform capabilities
    supports_ipad BOOLEAN NOT NULL DEFAULT true,
    allows_udid_enrollment BOOLEAN NOT NULL DEFAULT true,
    allows_ratings_widget BOOLEAN NOT NULL DEFAULT true,
    
    -- Maximum number of certificates this plan is allowed to hold for the Ban Bypass pool
    cert_pool_size_limit INT NOT NULL DEFAULT 3 CHECK (cert_pool_size_limit >= 1)
);

-- Establish the deferred Foreign Key constraint from 001_create_tenants.sql.
-- Using ON DELETE RESTRICT guarantees we cannot delete a plan if any tenant is assigned to it.
ALTER TABLE public.tenants
    ADD CONSTRAINT fk_tenants_plan_id
    FOREIGN KEY (plan_id) REFERENCES public.tenant_plans(id) 
    ON DELETE RESTRICT;