-- Bind a platform admin to a tenant. NULL tenant_id = platform owner (manages all merchants);
-- a set tenant_id = a merchant admin, scoped to that single store.
ALTER TABLE public.platform_admins
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES public.tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_platform_admins_tenant_id ON public.platform_admins(tenant_id);
