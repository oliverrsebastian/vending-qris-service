DROP INDEX IF EXISTS idx_transactions_invoice_number;
ALTER TABLE transactions DROP COLUMN IF EXISTS invoice_number;
