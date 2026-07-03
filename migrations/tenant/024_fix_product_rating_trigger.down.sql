-- Revert to the unqualified (search_path-dependent) function body from 023.
CREATE OR REPLACE FUNCTION refresh_product_rating() RETURNS TRIGGER AS $$
DECLARE
    pid UUID := COALESCE(NEW.product_id, OLD.product_id);
BEGIN
    UPDATE products p SET
        rating_avg = COALESCE((SELECT ROUND(AVG(rating)::numeric, 2) FROM product_ratings WHERE product_id = pid), 0),
        rating_count = (SELECT COUNT(*) FROM product_ratings WHERE product_id = pid),
        updated_at = NOW()
    WHERE p.id = pid;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
