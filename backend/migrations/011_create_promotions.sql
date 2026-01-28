-- Migration: Promotions System
-- Supports PERCENT_OFF and BUY_X_GET_Y promotions
-- This is the updated version with constraints, triggers, and indexes

-- Create promotions table
CREATE TABLE IF NOT EXISTS promotions (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('PERCENT_OFF', 'BUY_X_GET_Y')),
    rules JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure end_at is after start_at
    CONSTRAINT check_dates CHECK (end_at > start_at)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_promotions_active ON promotions(active);
CREATE INDEX IF NOT EXISTS idx_promotions_dates ON promotions(start_at, end_at);
CREATE INDEX IF NOT EXISTS idx_promotions_priority ON promotions(priority DESC);
CREATE INDEX IF NOT EXISTS idx_promotions_type ON promotions(type);

-- Trigger to update updated_at timestamp (uses existing function)
CREATE TRIGGER update_promotions_updated_at
    BEFORE UPDATE ON promotions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Example promotions (for testing) - set active=false by default
-- PERCENT_OFF: 20% off all products
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority) VALUES
('20% Off Everything', 'PERCENT_OFF', '{"percent": 20, "productIds": []}'::jsonb, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '30 days', 10)
ON CONFLICT DO NOTHING;

-- BUY_X_GET_Y: Buy 1 Get 1 on Chocolate Cake (product ID 1)
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority) VALUES
('BOGO Chocolate Cake', 'BUY_X_GET_Y', '{"buyQty": 1, "getQty": 1, "productIds": [1]}'::jsonb, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '30 days', 5)
ON CONFLICT DO NOTHING;
