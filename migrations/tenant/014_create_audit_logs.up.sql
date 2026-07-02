
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    
    old_value JSONB,
    new_value JSONB,
    
    ip_address INET,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_admin_id ON audit_logs(admin_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- CRITICAL SECURITY: Trigger Function to prevent tampering
CREATE OR REPLACE FUNCTION prevent_tenant_audit_tampering()
RETURNS TRIGGER AS $$
BEGIN
    -- Cross-Schema Insert: Logs the attempt directly to the public platform violation table.
    -- We use TG_TABLE_SCHEMA to explicitly capture WHICH tenant's audit log was attacked.
    INSERT INTO public.system_violations (table_name, attempted_action, user_id, created_at)
    VALUES (TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME, TG_OP, current_user, NOW());

    -- Block the transaction completely
    RAISE EXCEPTION 'CRITICAL SECURITY VIOLATION: Tenant audit log tampering detected and blocked. Incident reported to platform administrators.';
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Attach the trigger to fire BEFORE any UPDATE or DELETE operations
CREATE TRIGGER trg_prevent_tenant_audit_tampering
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_tenant_audit_tampering();
