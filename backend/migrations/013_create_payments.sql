CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL,
    user_id VARCHAR(255),
    amount DECIMAL(10, 2),
    method VARCHAR(50), -- kpay, wave, cash, manual_upload
    status VARCHAR(50) DEFAULT 'pending', -- pending, verified, rejected
    proof_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
