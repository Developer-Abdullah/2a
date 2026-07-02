-- Tie a minted activation code back to the order that paid for it, so the storefront can show the
-- customer their code on the order page and "my orders". Nullable: admin-generated codes (bulk /
-- reseller) have no order.

ALTER TABLE activation_codes
    ADD COLUMN order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    ADD COLUMN order_item_id UUID REFERENCES order_items(id) ON DELETE SET NULL;

CREATE INDEX idx_activation_codes_order ON activation_codes (order_id);
