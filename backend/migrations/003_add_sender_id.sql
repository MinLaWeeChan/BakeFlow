-- Migration: Add sender_id column to orders for Messenger recipient tracking
-- Date: 2025-11-25

ALTER TABLE orders
	ADD COLUMN IF NOT EXISTS sender_id TEXT; -- Messenger PSID or similar identifier

-- Optional: index if lookups by sender_id become frequent
CREATE INDEX IF NOT EXISTS idx_orders_sender_id ON orders(sender_id);

COMMENT ON COLUMN orders.sender_id IS 'Messenger sender/recipient ID used for status notifications';
