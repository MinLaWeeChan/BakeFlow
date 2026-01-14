-- Migration: Add scheduling fields to orders
-- Date: 2026-01-14

-- Store when an order should be prepared/delivered.
ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS scheduled_for TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS schedule_type TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_scheduled_for ON orders(scheduled_for);
