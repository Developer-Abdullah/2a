DROP INDEX IF EXISTS idx_activation_codes_order;

ALTER TABLE activation_codes
    DROP COLUMN IF EXISTS order_item_id,
    DROP COLUMN IF EXISTS order_id;
