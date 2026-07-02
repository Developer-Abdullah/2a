ALTER TABLE public.tenants 
    DROP CONSTRAINT IF EXISTS fk_tenants_plan_id;

DROP TABLE IF EXISTS public.tenant_plans CASCADE;