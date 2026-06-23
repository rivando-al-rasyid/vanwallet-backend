DROP INDEX IF EXISTS idx_profiles_user_id;
DROP INDEX IF EXISTS idx_topups_status;
DROP INDEX IF EXISTS idx_topups_wallet_created;
DROP INDEX IF EXISTS idx_transactions_wallet_status_type;
DROP INDEX IF EXISTS idx_transactions_wallet_created;
DROP INDEX IF EXISTS idx_wallets_user_id;

ALTER TABLE "topups"
    DROP CONSTRAINT IF EXISTS topups_amount_positive;

ALTER TABLE "transactions"
    DROP CONSTRAINT IF EXISTS transactions_admin_fee_non_negative,
    DROP CONSTRAINT IF EXISTS transactions_amount_positive;

ALTER TABLE "wallets"
    DROP CONSTRAINT IF EXISTS wallets_balance_non_negative;
