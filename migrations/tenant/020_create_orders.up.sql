-- Customer orders. One row per checkout. The status machine is:
--   pending  -> paid -> fulfilled   (happy path)
--   pending  -> failed              (payment declined / abandoned)
-- A paid order is fulfilled by minting activation codes (see 022) exactly once; idempotency is
-- anchored on (provider, provider_ref) so a redelivered webhook cannot double-fulfill.

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Guest checkout is allowed, so user_id is nullable. It is linked when the buyer is known.
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    email VARCHAR(255) NOT NULL,
    phone VARCHAR(40),

    currency CHAR(3) NOT NULL,
    subtotal NUMERIC(12, 3) NOT NULL CHECK (subtotal >= 0),
    total NUMERIC(12, 3) NOT NULL CHECK (total >= 0),

    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed', 'fulfilled')),

    -- Payment gateway used and the reference we correlate the webhook against.
    provider VARCHAR(30),
    provider_ref VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    fulfilled_at TIMESTAMPTZ
);

-- One provider reference maps to exactly one order (dedupe anchor for webhooks). Partial so multiple
-- orders can sit at NULL provider_ref before checkout starts.
CREATE UNIQUE INDEX uq_orders_provider_ref
    ON orders (provider, provider_ref)
    WHERE provider_ref IS NOT NULL;

CREATE INDEX idx_orders_user ON orders (user_id, created_at DESC);
CREATE INDEX idx_orders_status ON orders (status, created_at DESC);
