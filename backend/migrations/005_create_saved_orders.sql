-- Migration: Create saved orders tables (cross-device favorites)
-- Date: 2026-01-14

-- Ensure updated_at trigger helper exists (used by other migrations too)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS saved_orders (
    id SERIAL PRIMARY KEY,
    sender_id TEXT NOT NULL,
    name TEXT NOT NULL,
    note TEXT,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_used_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_saved_orders_sender_id ON saved_orders(sender_id);
CREATE INDEX IF NOT EXISTS idx_saved_orders_updated_at ON saved_orders(updated_at);

CREATE TRIGGER update_saved_orders_updated_at
    BEFORE UPDATE ON saved_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS saved_order_items (
    id SERIAL PRIMARY KEY,
    saved_order_id INT NOT NULL REFERENCES saved_orders(id) ON DELETE CASCADE,
    product_id INT,
    name TEXT NOT NULL,
    qty INT NOT NULL DEFAULT 1,
    price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_saved_order_items_saved_order_id ON saved_order_items(saved_order_id);
