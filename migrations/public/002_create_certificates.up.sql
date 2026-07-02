CREATE TABLE public.tenant_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Link to the store owner (tenant). 
    -- CASCADE ensures if a tenant record is purged, their cert metadata is cleaned up.
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    
    -- Admin-friendly display name (e.g., "Main Enterprise Cert 2026")
    label VARCHAR(255) NOT NULL,
    
    -- S3 Storage keys for the encrypted certificate assets
    s3_key_p12_enc VARCHAR(512) NOT NULL,
    s3_key_mobileprovision VARCHAR(512) NOT NULL,
    
    -- AES-256-GCM ciphertext of the .p12 password. MUST be BYTEA.
    encrypted_password BYTEA NOT NULL,
    
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Status flags for the "Apple Ban Bypass" rotation logic
    is_active BOOLEAN NOT NULL DEFAULT true,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    
    -- Sanity check: Expiry must be in the future relative to when it was added
    CONSTRAINT chk_expires_after_added CHECK (expires_at > added_at)
);

-- Optimize the query used by the Signing Worker to fetch the active certificate for a tenant
CREATE INDEX idx_tenant_certs_tenant_active ON public.tenant_certificates(tenant_id, is_active);

-- Optimize the query used by the background cron worker to check for expiring certificates
CREATE INDEX idx_tenant_certs_expires_at ON public.tenant_certificates(expires_at);