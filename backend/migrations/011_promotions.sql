-- Create promotions table
CREATE TABLE IF NOT EXISTS promotions (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value DECIMAL(10, 2) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT false,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for active promotions
CREATE INDEX idx_promotions_active ON promotions(product_id, is_active);
CREATE INDEX idx_promotions_dates ON promotions(start_date, end_date);

-- Add promotion_id to products for quick lookup (optional, for performance)
ALTER TABLE products ADD COLUMN IF NOT EXISTS active_promotion_id INT REFERENCES promotions(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_products_promotion ON products(active_promotion_id);
