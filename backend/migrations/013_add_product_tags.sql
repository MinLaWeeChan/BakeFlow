-- Migration: Add product tags
-- Date: 2026-01-30

ALTER TABLE products
ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE INDEX IF NOT EXISTS idx_products_tags_gin ON products USING GIN (tags);
