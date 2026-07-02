
DROP TRIGGER IF EXISTS trg_prevent_tenant_audit_tampering ON audit_logs;
DROP FUNCTION IF EXISTS prevent_tenant_audit_tampering();

DROP TABLE IF EXISTS audit_logs CASCADE;