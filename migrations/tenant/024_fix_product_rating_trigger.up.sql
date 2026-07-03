-- Replace refresh_product_rating() with a search_path-safe version. The original (023) used
-- unqualified table names, which fail when the trigger fires on a pooled connection whose search_path
-- does not include the tenant schema (e.g. a public rating submit). Resolve the schema at runtime.
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
