-- Rollback script for 008_create_promotions.sql
-- Run this to undo the first promotions migration

-- Drop trigger if exists
DROP TRIGGER IF EXISTS update_promotions_updated_at ON promotions;

-- Drop indexes
DROP INDEX IF EXISTS idx_promotions_active;
DROP INDEX IF EXISTS idx_promotions_dates;
DROP INDEX IF EXISTS idx_promotions_priority;
DROP INDEX IF EXISTS idx_promotions_type;

-- Drop the table
DROP TABLE IF EXISTS promotions;
