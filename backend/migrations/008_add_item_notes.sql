-- Migration: Add note and image_url columns to order_items for special instructions and product images
-- Date: 2026-01-20

-- Add note column to order_items table
ALTER TABLE order_items 
  ADD COLUMN IF NOT EXISTS note TEXT DEFAULT '';

-- Add image_url column to order_items table (stores product image at time of order)
ALTER TABLE order_items 
  ADD COLUMN IF NOT EXISTS image_url TEXT DEFAULT '';

-- Comments for documentation
COMMENT ON COLUMN order_items.note IS 'Special instructions or notes for this item (e.g., "No nuts", "Extra frosting")';
COMMENT ON COLUMN order_items.image_url IS 'Product image URL at time of order (for historical reference)';
