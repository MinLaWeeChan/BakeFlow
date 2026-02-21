-- Migration: Production-Grade Stock Reservation System
-- Similar to Amazon, GrabFood, Shopify inventory management

-- Add reserved_stock column to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS reserved_stock INTEGER NOT NULL DEFAULT 0;

-- Create stock_reservations table for tracking active reservations
CREATE TABLE IF NOT EXISTS stock_reservations (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    session_id VARCHAR(255),              -- For cart reservations before order
    quantity INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, confirmed, released, expired
    expires_at TIMESTAMP NOT NULL,        -- Auto-release after timeout
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMP,
    released_at TIMESTAMP,
    release_reason VARCHAR(100)           -- timeout, cancelled, payment_failed, manual
);

-- Create stock_transactions for audit trail
CREATE TABLE IF NOT EXISTS stock_transactions (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    transaction_type VARCHAR(50) NOT NULL, -- reserve, confirm, release, adjust, restock, sale
    quantity INTEGER NOT NULL,             -- Positive or negative
    previous_stock INTEGER NOT NULL,
    new_stock INTEGER NOT NULL,
    previous_reserved INTEGER NOT NULL DEFAULT 0,
    new_reserved INTEGER NOT NULL DEFAULT 0,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    reservation_id INTEGER REFERENCES stock_reservations(id) ON DELETE SET NULL,
    admin_id INTEGER,
    reason VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_reservations_product ON stock_reservations(product_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON stock_reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires ON stock_reservations(expires_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_reservations_session ON stock_reservations(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_stock_transactions_product ON stock_transactions(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_transactions_order ON stock_transactions(order_id);

-- Function to calculate available stock (total - reserved)
-- This is the ONLY source of truth for availability
CREATE OR REPLACE FUNCTION get_available_stock(p_product_id INTEGER)
RETURNS INTEGER AS $$
DECLARE
    v_stock INTEGER;
    v_reserved INTEGER;
BEGIN
    SELECT stock, reserved_stock INTO v_stock, v_reserved
    FROM products
    WHERE id = p_product_id AND deleted_at IS NULL;
    
    IF v_stock IS NULL THEN
        RETURN 0;
    END IF;
    
    RETURN GREATEST(v_stock - v_reserved, 0);
END;
$$ LANGUAGE plpgsql;

-- Atomic stock reservation with row locking
-- Returns reservation_id on success, 0 on insufficient stock
CREATE OR REPLACE FUNCTION reserve_stock(
    p_product_id INTEGER,
    p_quantity INTEGER,
    p_session_id VARCHAR(255),
    p_timeout_minutes INTEGER DEFAULT 15
)
RETURNS INTEGER AS $$
DECLARE
    v_available INTEGER;
    v_reservation_id INTEGER;
    v_current_stock INTEGER;
    v_current_reserved INTEGER;
BEGIN
    -- Lock the product row to prevent concurrent modifications
    SELECT stock, reserved_stock INTO v_current_stock, v_current_reserved
    FROM products
    WHERE id = p_product_id AND deleted_at IS NULL AND status = 'active'
    FOR UPDATE;
    
    IF v_current_stock IS NULL THEN
        RETURN 0; -- Product not found or inactive
    END IF;
    
    v_available := v_current_stock - v_current_reserved;
    
    IF v_available < p_quantity THEN
        RETURN 0; -- Insufficient stock
    END IF;
    
    -- Create reservation
    INSERT INTO stock_reservations (product_id, session_id, quantity, expires_at)
    VALUES (p_product_id, p_session_id, p_quantity, NOW() + (p_timeout_minutes || ' minutes')::INTERVAL)
    RETURNING id INTO v_reservation_id;
    
    -- Update reserved count
    UPDATE products
    SET reserved_stock = reserved_stock + p_quantity
    WHERE id = p_product_id;
    
    -- Log transaction
    INSERT INTO stock_transactions (
        product_id, transaction_type, quantity,
        previous_stock, new_stock, previous_reserved, new_reserved,
        reservation_id, reason
    ) VALUES (
        p_product_id, 'reserve', p_quantity,
        v_current_stock, v_current_stock, v_current_reserved, v_current_reserved + p_quantity,
        v_reservation_id, 'Cart reservation'
    );
    
    RETURN v_reservation_id;
END;
$$ LANGUAGE plpgsql;

-- Confirm reservation (convert to actual sale)
-- Called when order is confirmed/paid
CREATE OR REPLACE FUNCTION confirm_reservation(
    p_reservation_id INTEGER,
    p_order_id INTEGER
)
RETURNS BOOLEAN AS $$
DECLARE
    v_product_id INTEGER;
    v_quantity INTEGER;
    v_status VARCHAR(50);
    v_current_stock INTEGER;
    v_current_reserved INTEGER;
BEGIN
    -- Lock and get reservation
    SELECT product_id, quantity, status INTO v_product_id, v_quantity, v_status
    FROM stock_reservations
    WHERE id = p_reservation_id
    FOR UPDATE;
    
    IF v_status IS NULL OR v_status != 'pending' THEN
        RETURN FALSE; -- Invalid or already processed
    END IF;
    
    -- Lock product row
    SELECT stock, reserved_stock INTO v_current_stock, v_current_reserved
    FROM products
    WHERE id = v_product_id
    FOR UPDATE;
    
    -- Update reservation status
    UPDATE stock_reservations
    SET status = 'confirmed', order_id = p_order_id, confirmed_at = NOW()
    WHERE id = p_reservation_id;
    
    -- Deduct from both stock and reserved
    UPDATE products
    SET stock = stock - v_quantity,
        reserved_stock = reserved_stock - v_quantity
    WHERE id = v_product_id;
    
    -- Log transaction
    INSERT INTO stock_transactions (
        product_id, transaction_type, quantity,
        previous_stock, new_stock, previous_reserved, new_reserved,
        order_id, reservation_id, reason
    ) VALUES (
        v_product_id, 'confirm', -v_quantity,
        v_current_stock, v_current_stock - v_quantity,
        v_current_reserved, v_current_reserved - v_quantity,
        p_order_id, p_reservation_id, 'Order confirmed'
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Release reservation (timeout, cancel, payment failure)
CREATE OR REPLACE FUNCTION release_reservation(
    p_reservation_id INTEGER,
    p_reason VARCHAR(100)
)
RETURNS BOOLEAN AS $$
DECLARE
    v_product_id INTEGER;
    v_quantity INTEGER;
    v_status VARCHAR(50);
    v_current_stock INTEGER;
    v_current_reserved INTEGER;
BEGIN
    -- Lock and get reservation
    SELECT product_id, quantity, status INTO v_product_id, v_quantity, v_status
    FROM stock_reservations
    WHERE id = p_reservation_id
    FOR UPDATE;
    
    IF v_status IS NULL OR v_status NOT IN ('pending') THEN
        RETURN FALSE; -- Invalid or already processed
    END IF;
    
    -- Lock product row
    SELECT stock, reserved_stock INTO v_current_stock, v_current_reserved
    FROM products
    WHERE id = v_product_id
    FOR UPDATE;
    
    -- Update reservation status
    UPDATE stock_reservations
    SET status = 'released', released_at = NOW(), release_reason = p_reason
    WHERE id = p_reservation_id;
    
    -- Release reserved stock
    UPDATE products
    SET reserved_stock = GREATEST(reserved_stock - v_quantity, 0)
    WHERE id = v_product_id;
    
    -- Log transaction
    INSERT INTO stock_transactions (
        product_id, transaction_type, quantity,
        previous_stock, new_stock, previous_reserved, new_reserved,
        reservation_id, reason
    ) VALUES (
        v_product_id, 'release', v_quantity,
        v_current_stock, v_current_stock,
        v_current_reserved, GREATEST(v_current_reserved - v_quantity, 0),
        p_reservation_id, p_reason
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Atomic direct purchase (reserve + confirm in one step)
-- For immediate checkout without cart holding
CREATE OR REPLACE FUNCTION atomic_purchase(
    p_product_id INTEGER,
    p_quantity INTEGER,
    p_order_id INTEGER
)
RETURNS BOOLEAN AS $$
DECLARE
    v_available INTEGER;
    v_current_stock INTEGER;
    v_current_reserved INTEGER;
BEGIN
    -- Lock the product row
    SELECT stock, reserved_stock INTO v_current_stock, v_current_reserved
    FROM products
    WHERE id = p_product_id AND deleted_at IS NULL AND status = 'active'
    FOR UPDATE;
    
    IF v_current_stock IS NULL THEN
        RETURN FALSE; -- Product not found or inactive
    END IF;
    
    v_available := v_current_stock - v_current_reserved;
    
    IF v_available < p_quantity THEN
        RETURN FALSE; -- Insufficient stock
    END IF;
    
    -- Directly deduct stock (no reservation step)
    UPDATE products
    SET stock = stock - p_quantity
    WHERE id = p_product_id;
    
    -- Log transaction
    INSERT INTO stock_transactions (
        product_id, transaction_type, quantity,
        previous_stock, new_stock, previous_reserved, new_reserved,
        order_id, reason
    ) VALUES (
        p_product_id, 'sale', -p_quantity,
        v_current_stock, v_current_stock - p_quantity,
        v_current_reserved, v_current_reserved,
        p_order_id, 'Direct purchase'
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Restore stock (cancellation, refund)
CREATE OR REPLACE FUNCTION restore_stock(
    p_product_id INTEGER,
    p_quantity INTEGER,
    p_order_id INTEGER,
    p_reason VARCHAR(255)
)
RETURNS BOOLEAN AS $$
DECLARE
    v_current_stock INTEGER;
    v_current_reserved INTEGER;
BEGIN
    -- Lock the product row
    SELECT stock, reserved_stock INTO v_current_stock, v_current_reserved
    FROM products
    WHERE id = p_product_id
    FOR UPDATE;
    
    IF v_current_stock IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Restore stock
    UPDATE products
    SET stock = stock + p_quantity
    WHERE id = p_product_id;
    
    -- Log transaction
    INSERT INTO stock_transactions (
        product_id, transaction_type, quantity,
        previous_stock, new_stock, previous_reserved, new_reserved,
        order_id, reason
    ) VALUES (
        p_product_id, 'restock', p_quantity,
        v_current_stock, v_current_stock + p_quantity,
        v_current_reserved, v_current_reserved,
        p_order_id, p_reason
    );
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Cleanup expired reservations (run periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_reservations()
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER := 0;
    v_reservation RECORD;
BEGIN
    FOR v_reservation IN
        SELECT id FROM stock_reservations
        WHERE status = 'pending' AND expires_at < NOW()
        FOR UPDATE SKIP LOCKED  -- Skip locked rows to avoid blocking
    LOOP
        PERFORM release_reservation(v_reservation.id, 'timeout');
        v_count := v_count + 1;
    END LOOP;
    
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- View for real-time stock status
CREATE OR REPLACE VIEW product_stock_status AS
SELECT 
    p.id,
    p.name,
    p.stock as total_stock,
    p.reserved_stock,
    (p.stock - p.reserved_stock) as available_stock,
    CASE 
        WHEN (p.stock - p.reserved_stock) <= 0 THEN 'out_of_stock'
        WHEN (p.stock - p.reserved_stock) <= 5 THEN 'low_stock'
        ELSE 'in_stock'
    END as stock_status,
    p.status as product_status
FROM products p
WHERE p.deleted_at IS NULL;

COMMENT ON TABLE stock_reservations IS 'Tracks temporary stock reservations during checkout flow';
COMMENT ON TABLE stock_transactions IS 'Complete audit trail of all stock movements';
COMMENT ON FUNCTION reserve_stock IS 'Atomically reserve stock with row locking - prevents overselling';
COMMENT ON FUNCTION confirm_reservation IS 'Convert pending reservation to actual sale';
COMMENT ON FUNCTION release_reservation IS 'Release reserved stock back to available pool';
COMMENT ON FUNCTION atomic_purchase IS 'Single-step purchase without reservation (for instant checkout)';
COMMENT ON FUNCTION restore_stock IS 'Return stock on cancellation/refund';
