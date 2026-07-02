-- Line items of an order. Product name and unit price are snapshotted at purchase time so a later
-- price/name change on the catalog does not rewrite historical orders.

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,

    -- RESTRICT: a product with sales history cannot be hard-deleted (unpublish instead).
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

    product_name VARCHAR(255) NOT NULL,
    qty INT NOT NULL CHECK (qty >= 1),
    unit_amount NUMERIC(12, 3) NOT NULL CHECK (unit_amount >= 0),
    currency CHAR(3) NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items (order_id);
