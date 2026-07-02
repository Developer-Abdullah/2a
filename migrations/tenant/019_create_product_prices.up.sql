-- Per-currency pricing for a product. The storefront runs multi-currency (EGP via Paymob, KWD via
-- MyFatoorah), so a product carries one price row per supported currency. NUMERIC(12,3): KWD is a
-- 3-decimal currency; EGP uses 2 but the wider scale stores both losslessly.

CREATE TABLE product_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,

    currency CHAR(3) NOT NULL,

    -- The charged price.
    amount NUMERIC(12, 3) NOT NULL CHECK (amount >= 0),

    -- Optional "was" price shown struck-through to convey a discount. When set it must exceed amount.
    compare_at_amount NUMERIC(12, 3) CHECK (compare_at_amount IS NULL OR compare_at_amount >= amount),

    CONSTRAINT uq_product_prices_currency UNIQUE (product_id, currency)
);

CREATE INDEX idx_product_prices_product ON product_prices (product_id);
