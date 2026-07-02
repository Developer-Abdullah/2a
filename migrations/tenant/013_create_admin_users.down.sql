ALTER TABLE notifications 
    DROP CONSTRAINT IF EXISTS fk_notifications_admin_id;

ALTER TABLE activation_codes 
    DROP CONSTRAINT IF EXISTS fk_activation_codes_admin_id;

DROP TABLE IF EXISTS admin_users CASCADE;
DROP TYPE IF EXISTS tenant_admin_role;