-- Migration: Create Products Management System
-- Description: Creates tables for products, logs, analytics, and admin roles

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    image_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'inactive', 'archived')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Index for faster queries
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_created_at ON products(created_at);

-- Admin roles table (for RBAC)
CREATE TABLE IF NOT EXISTS admin_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO admin_roles (name, permissions) VALUES
('viewer', '{"products": ["read"], "analytics": ["read"]}'::jsonb),
('editor', '{"products": ["read", "create", "update"], "analytics": ["read"]}'::jsonb),
('manager', '{"products": ["read", "create", "update", "delete"], "analytics": ["read", "manage"]}'::jsonb),
('owner', '{"products": ["read", "create", "update", "delete"], "analytics": ["read", "manage"], "roles": ["manage"]}'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- Admins table (extend if you have existing admins table)
CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role_id INTEGER REFERENCES admin_roles(id) DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Product logs table (audit trail)
CREATE TABLE IF NOT EXISTS product_logs (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    admin_id INTEGER REFERENCES admins(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    changes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster log queries
CREATE INDEX idx_product_logs_product_id ON product_logs(product_id);
CREATE INDEX idx_product_logs_admin_id ON product_logs(admin_id);
CREATE INDEX idx_product_logs_created_at ON product_logs(created_at);

-- Product analytics table
CREATE TABLE IF NOT EXISTS product_analytics (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    views INTEGER NOT NULL DEFAULT 0,
    purchases INTEGER NOT NULL DEFAULT 0,
    last_viewed_at TIMESTAMP,
    last_purchased_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id)
);

-- Index for analytics
CREATE INDEX idx_product_analytics_product_id ON product_analytics(product_id);
CREATE INDEX idx_product_analytics_views ON product_analytics(views);
CREATE INDEX idx_product_analytics_purchases ON product_analytics(purchases);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_admins_updated_at
    BEFORE UPDATE ON admins
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_analytics_updated_at
    BEFORE UPDATE ON product_analytics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample products (optional - for testing)
INSERT INTO products (name, description, category, price, stock, status) VALUES
('Chocolate Cake', 'Rich chocolate cake with ganache frosting', 'Cakes', 25.99, 10, 'active'),
('Vanilla Cupcake', 'Classic vanilla cupcakes with buttercream', 'Cupcakes', 3.99, 50, 'active'),
('Red Velvet Cake', 'Moist red velvet with cream cheese frosting', 'Cakes', 28.99, 8, 'active'),
('Blueberry Muffin', 'Fresh blueberry muffins', 'Muffins', 2.99, 30, 'active'),
('Strawberry Tart', 'Fresh strawberry tart with pastry cream', 'Tarts', 18.99, 12, 'active')
ON CONFLICT DO NOTHING;

-- Initialize analytics for sample products
INSERT INTO product_analytics (product_id, views, purchases)
SELECT id, 0, 0 FROM products
ON CONFLICT (product_id) DO NOTHING;
