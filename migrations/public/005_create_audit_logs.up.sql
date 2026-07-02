CREATE TABLE public.system_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(255) NOT NULL,
    attempted_action VARCHAR(50) NOT NULL,
    
    -- Tracks the PostgreSQL DB role that attempted the breach. 
    -- If using PgBouncer/app-level auth, it tracks the active DB connection user.
    user_id VARCHAR(255) NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Create the immutable platform audit logs table
CREATE TABLE public.platform_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- If an admin is deleted, keep the audit log but nullify the reference
    admin_id UUID REFERENCES public.platform_admins(id) ON DELETE SET NULL,
    
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    
    old_value JSONB,
    new_value JSONB,
    
    -- INET type is optimized for IPv4/IPv6 validation and indexing
    ip_address INET,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to quickly filter audit logs by the admin who performed the action or the resource affected
CREATE INDEX idx_platform_audit_logs_admin_id ON public.platform_audit_logs(admin_id);
CREATE INDEX idx_platform_audit_logs_resource ON public.platform_audit_logs(resource_type, resource_id);
CREATE INDEX idx_platform_audit_logs_created_at ON public.platform_audit_logs(created_at);

-- 3. Define the Trigger Function to guarantee immutability
CREATE OR REPLACE FUNCTION public.prevent_audit_tampering()
RETURNS TRIGGER AS $$
BEGIN
    -- Step A: Log the tampering attempt into system_violations
    INSERT INTO public.system_violations (table_name, attempted_action, user_id, created_at)
    VALUES (TG_TABLE_NAME, TG_OP, current_user, NOW());

    -- Step B: Raise a hard exception, rolling back the malicious UPDATE or DELETE transaction
    RAISE EXCEPTION 'CRITICAL SECURITY VIOLATION: Audit log tampering detected and blocked. Incident logged.';
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 4. Attach the trigger to fire BEFORE any UPDATE or DELETE operations
CREATE TRIGGER trg_prevent_audit_tampering
    BEFORE UPDATE OR DELETE ON public.platform_audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION public.prevent_audit_tampering();
