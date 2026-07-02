ALTER TABLE users 
    DROP CONSTRAINT IF EXISTS fk_users_activation_code_id;

DROP TABLE IF EXISTS activation_codes CASCADE;
DROP TYPE IF EXISTS allowed_device_type;
DROP TYPE IF EXISTS activation_code_type;