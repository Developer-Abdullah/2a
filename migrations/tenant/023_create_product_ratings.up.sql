-- Per-product customer ratings. Distinct from platform_ratings (store-wide): these attach to a
-- specific product and drive the rating shown on its page. One rating per IP per product per day.

CREATE TABLE product_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,

    -- Nullable: anonymous (unactivated) shoppers may rate a product they bought/viewed.
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment VARCHAR(280),

    ip_hash VARCHAR(255) NOT NULL,
    rating_date DATE NOT NULL DEFAULT CURRENT_DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Spam guard + upsert anchor: one rating per IP per product per day (a repeat updates it).
CREATE UNIQUE INDEX idx_product_ratings_unique_daily ON product_ratings (product_id, ip_hash, rating_date);
CREATE INDEX idx_product_ratings_product ON product_ratings (product_id, created_at DESC);

-- Keep the denormalized products.rating_avg / rating_count in sync so the storefront can read them
-- straight off the product row without a join. Recomputed on any change to a product's ratings.
--
-- The body is schema-qualified via TG_TABLE_SCHEMA + dynamic SQL: writers hit product_ratings on a
-- pooled connection whose search_path does NOT include the tenant schema, so unqualified table names
-- inside the trigger would fail to resolve. Resolving the schema at runtime makes it search_path-safe.
CREATE OR REPLACE FUNCTION refresh_product_rating() RETURNS TRIGGER AS $$
DECLARE
    pid UUID := COALESCE(NEW.product_id, OLD.product_id);
BEGIN
    EXECUTE format(
        'UPDATE %1$I.products p SET
            rating_avg = COALESCE((SELECT ROUND(AVG(rating)::numeric, 2) FROM %1$I.product_ratings WHERE product_id = $1), 0),
            rating_count = (SELECT COUNT(*) FROM %1$I.product_ratings WHERE product_id = $1),
            updated_at = NOW()
         WHERE p.id = $1', TG_TABLE_SCHEMA)
    USING pid;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_refresh_product_rating
    AFTER INSERT OR UPDATE OR DELETE ON product_ratings
    FOR EACH ROW EXECUTE FUNCTION refresh_product_rating();
