CREATE TABLE IF NOT EXISTS preorder_product_settings (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    start_date DATE NULL,
    end_date DATE NULL,
    sizes JSONB NOT NULL DEFAULT '[]'::jsonb,
    layers JSONB NOT NULL DEFAULT '[]'::jsonb,
    creams JSONB NOT NULL DEFAULT '[]'::jsonb,
    flavors JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT preorder_product_settings_product_id_unique UNIQUE (product_id)
);

CREATE INDEX IF NOT EXISTS idx_preorder_product_settings_enabled ON preorder_product_settings (enabled);
CREATE INDEX IF NOT EXISTS idx_preorder_product_settings_product_id ON preorder_product_settings (product_id);

CREATE TRIGGER update_preorder_product_settings_updated_at
    BEFORE UPDATE ON preorder_product_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
