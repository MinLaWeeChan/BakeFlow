-- Migration: Customer identity + blocking (PSID-first)
-- Date: 2026-01-25

CREATE TABLE IF NOT EXISTS customer_phones (
    psid TEXT NOT NULL,
    phone TEXT NOT NULL,
    first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (psid, phone)
);

CREATE INDEX IF NOT EXISTS idx_customer_phones_phone ON customer_phones(phone);
CREATE INDEX IF NOT EXISTS idx_customer_phones_psid ON customer_phones(psid);

CREATE TABLE IF NOT EXISTS blocked_identities (
    id SERIAL PRIMARY KEY,
    identity_type TEXT NOT NULL CHECK (identity_type IN ('psid', 'phone')),
    value TEXT NOT NULL,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (identity_type, value)
);

CREATE INDEX IF NOT EXISTS idx_blocked_identities_type_value ON blocked_identities(identity_type, value);

CREATE TABLE IF NOT EXISTS customer_verifications (
    psid TEXT PRIMARY KEY,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_method TEXT,
    verified_at TIMESTAMP,
    verified_by_admin_id INTEGER REFERENCES admins(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_customer_verifications_verified ON customer_verifications(verified);

CREATE TABLE IF NOT EXISTS customer_verification_requests (
    id SERIAL PRIMARY KEY,
    psid TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'expired', 'cancelled')),
    requested_by_admin_id INTEGER REFERENCES admins(id) ON DELETE SET NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customer_verification_requests_psid_status ON customer_verification_requests(psid, status);

CREATE TABLE IF NOT EXISTS admin_action_logs (
    id SERIAL PRIMARY KEY,
    admin_id INTEGER REFERENCES admins(id) ON DELETE SET NULL,
    action_type TEXT NOT NULL,
    psid TEXT,
    reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_action_logs_psid ON admin_action_logs(psid);
CREATE INDEX IF NOT EXISTS idx_admin_action_logs_created_at ON admin_action_logs(created_at);
