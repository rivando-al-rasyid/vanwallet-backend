COMMENT ON TABLE "users" IS 'Core identity table. Credentials only — no PII stored here.';

COMMENT ON TABLE "profiles" IS 'One-to-one with users. Stores PII separately for easier data management.';

COMMENT ON TABLE "favorites" IS 'Links the owner (user_id) to their saved contact (target_user_id) for quick transfers.';

COMMENT ON TABLE "user_pins" IS 'PIN stored as bcrypt hash. locked_until enables brute-force protection.';

COMMENT ON TABLE "wallets" IS 'Supports multiple wallets per user. Balance stored in smallest currency unit (sen/IDR).';

COMMENT ON TABLE "topups" IS 'Standalone top‑up table. Separate category — not recorded in the central transactions ledger.';

COMMENT ON TABLE "transactions" IS 'Central ledger for EXPENSE, WITHDRAWAL, TRANSFER_IN/OUT.
Amount is signed (+ for IN, - for OUT). Top‑ups are handled separately.

PostgreSQL CHECK constraint example:
ALTER TABLE transactions ADD CONSTRAINT chk_amount_sign CHECK (
  (type = ''TRANSFER_IN'' AND amount > 0) OR
  (type IN (''EXPENSE'',''WITHDRAWAL'',''TRANSFER_OUT'') AND amount < 0)
);
';

COMMENT ON TABLE "transfers" IS 'Links BOTH sides of a transfer. transaction_id = OUT row (sender). recipient_transaction_id = IN row (recipient).';

COMMENT ON TABLE "expenses" IS 'Mapped exclusively to EXPENSE type transactions.';

COMMENT ON TABLE "withdrawals" IS 'Manual bank withdrawal snapshot data.';

ALTER TABLE "profiles"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "favorites"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "favorites"
    ADD FOREIGN KEY ("target_user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_pins"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "wallets"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "topups"
    ADD FOREIGN KEY ("wallet_id") REFERENCES "wallets"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transactions"
    ADD FOREIGN KEY ("wallet_id") REFERENCES "wallets"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transfers"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transfers"
    ADD FOREIGN KEY ("recipient_transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "expenses"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "withdrawals"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

