-- Migration: Add order_items table for shopping cart functionality
-- Date: 2025-11-17

-- Update orders table structure
ALTER TABLE orders 
  DROP COLUMN IF EXISTS product,
  DROP COLUMN IF EXISTS quantity,
  ADD COLUMN IF NOT EXISTS delivery_type TEXT,
  ADD COLUMN IF NOT EXISTS total_items INT DEFAULT 0;

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
  id SERIAL PRIMARY KEY,
  order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product TEXT NOT NULL,
  quantity INT NOT NULL,
  price DECIMAL(10,2) DEFAULT 0.00,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Comments for documentation
COMMENT ON TABLE order_items IS 'Individual items in each order (supports multiple items per order)';
COMMENT ON COLUMN orders.total_items IS 'Total quantity of all items in the order';
COMMENT ON COLUMN orders.delivery_type IS 'pickup or delivery';
