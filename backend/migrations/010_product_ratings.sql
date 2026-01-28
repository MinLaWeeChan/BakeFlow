-- Migration: Product Ratings System
-- Allows customers to rate products after order delivery

-- Create product_ratings table
CREATE TABLE IF NOT EXISTS product_ratings (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,           -- Messenger sender_id
    stars INTEGER NOT NULL CHECK (stars >= 1 AND stars <= 5),
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- One rating per product per order per user
    UNIQUE(product_id, order_id, user_id)
);

-- Add average rating cache columns to products for fast queries
ALTER TABLE products ADD COLUMN IF NOT EXISTS avg_rating DECIMAL(2,1) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS rating_count INTEGER DEFAULT 0;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ratings_product ON product_ratings(product_id);
CREATE INDEX IF NOT EXISTS idx_ratings_order ON product_ratings(order_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user ON product_ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_ratings_stars ON product_ratings(stars);

-- Function to update product average rating
CREATE OR REPLACE FUNCTION update_product_rating()
RETURNS TRIGGER AS $$
DECLARE
    v_avg DECIMAL(2,1);
    v_count INTEGER;
BEGIN
    -- Calculate new average
    SELECT 
        COALESCE(ROUND(AVG(stars)::numeric, 1), 0),
        COUNT(*)
    INTO v_avg, v_count
    FROM product_ratings
    WHERE product_id = COALESCE(NEW.product_id, OLD.product_id);
    
    -- Update product
    UPDATE products
    SET avg_rating = v_avg,
        rating_count = v_count
    WHERE id = COALESCE(NEW.product_id, OLD.product_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update product rating on insert/update/delete
DROP TRIGGER IF EXISTS trigger_update_product_rating ON product_ratings;
CREATE TRIGGER trigger_update_product_rating
AFTER INSERT OR UPDATE OR DELETE ON product_ratings
FOR EACH ROW
EXECUTE FUNCTION update_product_rating();

-- View for product ratings summary
CREATE OR REPLACE VIEW product_ratings_summary AS
SELECT 
    p.id as product_id,
    p.name as product_name,
    p.avg_rating,
    p.rating_count,
    COUNT(CASE WHEN pr.stars = 5 THEN 1 END) as five_star,
    COUNT(CASE WHEN pr.stars = 4 THEN 1 END) as four_star,
    COUNT(CASE WHEN pr.stars = 3 THEN 1 END) as three_star,
    COUNT(CASE WHEN pr.stars = 2 THEN 1 END) as two_star,
    COUNT(CASE WHEN pr.stars = 1 THEN 1 END) as one_star
FROM products p
LEFT JOIN product_ratings pr ON p.id = pr.product_id
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.avg_rating, p.rating_count;

COMMENT ON TABLE product_ratings IS 'Customer ratings for products (1-5 stars)';
COMMENT ON COLUMN products.avg_rating IS 'Cached average rating for fast queries';
COMMENT ON COLUMN products.rating_count IS 'Cached total number of ratings';
