ALTER TABLE transactions ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(255);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_invoice_number
    ON transactions (invoice_number)
    WHERE invoice_number IS NOT NULL AND invoice_number <> '';
