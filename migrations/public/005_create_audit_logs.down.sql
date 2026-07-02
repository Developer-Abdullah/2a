
DROP TRIGGER IF EXISTS trg_prevent_audit_tampering ON public.platform_audit_logs;
DROP FUNCTION IF EXISTS public.prevent_audit_tampering();

DROP TABLE IF EXISTS public.platform_audit_logs CASCADE;
DROP TABLE IF EXISTS public.system_violations CASCADE;