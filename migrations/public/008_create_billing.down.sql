DROP TABLE IF EXISTS public.subscription_events;
DROP TABLE IF EXISTS public.subscription_payments;

DROP INDEX IF EXISTS public.idx_tenants_subscription_expiry;

ALTER TABLE public.tenants
    DROP COLUMN IF EXISTS subscription_status,
    DROP COLUMN IF EXISTS subscription_expires_at,
    DROP COLUMN IF EXISTS grace_until;

DROP TYPE IF EXISTS subscription_status;

ALTER TABLE public.tenant_plans
    DROP COLUMN IF EXISTS price_amount,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS billing_interval;
