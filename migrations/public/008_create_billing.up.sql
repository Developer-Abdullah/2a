-- Billing / subscription layer. Plans already exist (003) with feature limits; here we add pricing,
-- per-tenant subscription state, and an idempotent payment ledger for provider webhooks.

-- Pricing on plans. NUMERIC(12,3): KWD is a 3-decimal currency. Price is data, never hardcoded.
ALTER TABLE public.tenant_plans
    ADD COLUMN price_amount NUMERIC(12,3) NOT NULL DEFAULT 0 CHECK (price_amount >= 0),
    ADD COLUMN currency CHAR(3) NOT NULL DEFAULT 'KWD',
    ADD COLUMN billing_interval VARCHAR(20) NOT NULL DEFAULT 'yearly'
        CHECK (billing_interval IN ('monthly', 'yearly'));

-- Subscription state on tenants. This is distinct from `status` (which is the operational
-- active/suspended/deactivated flag an owner sets manually).
CREATE TYPE subscription_status AS ENUM ('trialing', 'active', 'past_due', 'expired');

ALTER TABLE public.tenants
    ADD COLUMN subscription_status subscription_status NOT NULL DEFAULT 'trialing',
    ADD COLUMN subscription_expires_at TIMESTAMPTZ,
    -- End of the grace window after expiry during which distribution still works.
    ADD COLUMN grace_until TIMESTAMPTZ;

CREATE INDEX idx_tenants_subscription_expiry ON public.tenants (subscription_expires_at);

-- Payment ledger. One row per provider transaction. idempotency_key is the dedupe anchor: a webhook
-- delivered more than once (providers retry) upserts the same row instead of double-crediting.
CREATE TABLE public.subscription_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES public.tenant_plans(id) ON DELETE RESTRICT,

    provider VARCHAR(30) NOT NULL,          -- 'paymob' | 'myfatoorah'
    provider_ref VARCHAR(255) NOT NULL,     -- order/invoice id at the provider
    idempotency_key VARCHAR(255) NOT NULL,  -- provider+ref, unique — dedupes repeated webhooks

    amount NUMERIC(12,3) NOT NULL CHECK (amount >= 0),
    currency CHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed')),

    raw_event JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,

    CONSTRAINT uq_subscription_payments_idem UNIQUE (idempotency_key)
);

CREATE INDEX idx_subscription_payments_tenant ON public.subscription_payments (tenant_id, created_at DESC);

-- Lightweight event log for renewal reminders and subscription state changes (stands in for email
-- until the notifier grows a real channel).
CREATE TABLE public.subscription_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,        -- 'reminder' | 'activated' | 'expired' | 'past_due'
    detail JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscription_events_tenant ON public.subscription_events (tenant_id, created_at DESC);
