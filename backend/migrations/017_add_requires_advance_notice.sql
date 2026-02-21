-- Migration 017: Add requires_advance_notice to products
-- This replaces text-based preorder detection with product-level configuration
-- Products with requires_advance_notice = true are "Custom Cakes" 
--   (availability is based on preorder_period dates set by admin)
-- Products with requires_advance_notice = false are "Regular" (ready same day)

ALTER TABLE products ADD COLUMN IF NOT EXISTS requires_advance_notice BOOLEAN DEFAULT false;

-- Add order_type to orders table for clear classification
ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_type VARCHAR(20) DEFAULT 'regular';

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_products_advance_notice ON products(requires_advance_notice);
CREATE INDEX IF NOT EXISTS idx_orders_order_type ON orders(order_type);

-- Mark existing custom cake products (adjust WHERE clause to match your data)
-- UPDATE products SET requires_advance_notice = true
-- WHERE id IN (SELECT UNNEST(product_ids) FROM preorder_settings WHERE enabled = true);
