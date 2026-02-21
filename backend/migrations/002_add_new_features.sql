-- Migration: Add new features (Order History, Delivery Fees, Ratings)
-- Date: 2025-11-24

-- Step 1: Add basic columns to orders table (without foreign keys)
ALTER TABLE orders 
  ADD COLUMN IF NOT EXISTS delivery_fee DECIMAL(10,2) DEFAULT 0.00,
  ADD COLUMN IF NOT EXISTS subtotal DECIMAL(10,2) DEFAULT 0.00,
  ADD COLUMN IF NOT EXISTS total_amount DECIMAL(10,2) DEFAULT 0.00,
  ADD COLUMN IF NOT EXISTS reordered_from INT,
  ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP;

-- Step 2: Create ratings table
CREATE TABLE IF NOT EXISTS ratings (
  id SERIAL PRIMARY KEY,
  order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL,
  stars INT NOT NULL CHECK (stars >= 1 AND stars <= 5),
  comment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Step 3: Now add rating_id foreign key to orders (after ratings table exists)
ALTER TABLE orders 
  ADD COLUMN IF NOT EXISTS rating_id INT;

-- Step 4: Add foreign key constraints
ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS orders_rating_id_fkey;
  
ALTER TABLE orders
  ADD CONSTRAINT orders_rating_id_fkey 
  FOREIGN KEY (rating_id) REFERENCES ratings(id) ON DELETE SET NULL;

ALTER TABLE orders
  DROP CONSTRAINT IF EXISTS orders_reordered_from_fkey;
  
ALTER TABLE orders
  ADD CONSTRAINT orders_reordered_from_fkey 
  FOREIGN KEY (reordered_from) REFERENCES orders(id) ON DELETE SET NULL;

-- Step 5: Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_ratings_order_id ON ratings(order_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);

-- Update existing orders to have calculated totals (run this after migration)
-- UPDATE orders SET subtotal = 0.00, delivery_fee = 0.00, total_amount = 0.00 WHERE subtotal IS NULL;
