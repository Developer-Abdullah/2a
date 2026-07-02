CREATE TYPE tenant_isolation_mode AS ENUM ('dedicated_schema', 'shared_rls');
CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'deactivated');

CREATE TABLE public.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(63) NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9-]+$'),
    plan_id UUID NOT NULL,
    isolation_mode tenant_isolation_mode NOT NULL,
    schema_name VARCHAR(63) NOT NULL UNIQUE CHECK (schema_name ~ '^[a-z0-9_]+$'),
    s3_prefix VARCHAR(255) NOT NULL UNIQUE,
    status tenant_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON public.tenants (slug);
CREATE INDEX idx_tenants_status ON public.tenants (status);