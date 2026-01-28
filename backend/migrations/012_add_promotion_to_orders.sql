-- Migration: Add promotion tracking to orders
-- Allows storing which promotion was applied to each order and the discount amount

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS promotion_id INT REFERENCES promotions(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS discount DECIMAL(10, 2) NOT NULL DEFAULT 0.00;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_orders_promotion_id ON orders(promotion_id);

-- Update existing orders to have discount = 0
UPDATE orders SET discount = 0 WHERE discount IS NULL;

-- Add comments for documentation
COMMENT ON COLUMN orders.promotion_id IS 'Foreign key to the promotion that was applied to this order';
COMMENT ON COLUMN orders.discount IS 'Discount amount applied to this order';
