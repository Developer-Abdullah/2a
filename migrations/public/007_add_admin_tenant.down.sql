DROP INDEX IF EXISTS idx_platform_admins_tenant_id;
ALTER TABLE public.platform_admins DROP COLUMN IF EXISTS tenant_id;
