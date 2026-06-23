-- Strengthen financial data integrity and query performance.

ALTER TABLE "wallets"
    ADD CONSTRAINT wallets_balance_non_negative CHECK (balance >= 0);

ALTER TABLE "transactions"
    ADD CONSTRAINT transactions_amount_positive CHECK (amount > 0),
    ADD CONSTRAINT transactions_admin_fee_non_negative CHECK (admin_fee >= 0);

ALTER TABLE "topups"
    ADD CONSTRAINT topups_amount_positive CHECK (amount > 0);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_created ON transactions(wallet_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_status_type ON transactions(wallet_id, status, type);
CREATE INDEX IF NOT EXISTS idx_topups_wallet_created ON topups(wallet_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_topups_status ON topups(status);
CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
