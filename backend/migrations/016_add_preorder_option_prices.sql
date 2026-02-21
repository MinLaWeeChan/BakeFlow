ALTER TABLE preorder_product_settings
    ADD COLUMN IF NOT EXISTS size_prices JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS layer_prices JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS cream_prices JSONB NOT NULL DEFAULT '{}'::jsonb;
