-- Migration: Improve scheduling query performance
-- Date: 2026-01-14

-- Optional: if you frequently query upcoming scheduled orders.
CREATE INDEX IF NOT EXISTS idx_orders_status_scheduled_for ON orders(status, scheduled_for);
