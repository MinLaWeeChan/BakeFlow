CREATE TABLE IF NOT EXISTS preorder_settings (
    id SERIAL PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    product_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_preorder_settings_enabled ON preorder_settings (enabled);

CREATE TRIGGER update_preorder_settings_updated_at
    BEFORE UPDATE ON preorder_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO preorder_settings (id, enabled, product_ids)
VALUES (1, TRUE, '[]'::jsonb)
ON CONFLICT (id) DO NOTHING;
