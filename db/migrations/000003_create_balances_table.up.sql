CREATE TABLE IF NOT EXISTS balances (
    user_id INT PRIMARY KEY,
    current_balance DECIMAL(10, 2) DEFAULT 0.00 NOT NULL,
    withdrawn_balance DECIMAL(10, 2) DEFAULT 0.00 NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);