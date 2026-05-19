DROP TABLE IF EXISTS "withdrawals" CASCADE;

DROP TABLE IF EXISTS "expenses" CASCADE;

DROP TABLE IF EXISTS "transfers" CASCADE;

DROP TABLE IF EXISTS "topups" CASCADE;

DROP TABLE IF EXISTS "transactions" CASCADE;

DROP TABLE IF EXISTS "wallets" CASCADE;

DROP TABLE IF EXISTS "user_pins" CASCADE;

DROP TABLE IF EXISTS "profiles" CASCADE;

DROP TABLE IF EXISTS "users" CASCADE;

DROP TYPE IF EXISTS "transaction_status" CASCADE;

DROP TYPE IF EXISTS "payment_method" CASCADE;

DROP TYPE IF EXISTS "direction" CASCADE;

-- ============================================================
-- CREATE TYPES
-- ============================================================
CREATE TYPE "transaction_status" AS ENUM(
    'PENDING',
    'SUCCESS',
    'FAILED',
    'CANCELLED'
);

CREATE TYPE "payment_method" AS ENUM(
    'BRI',
    'BCA',
    'DANA',
    'GOPAY',
    'OVO'
);

CREATE TYPE "direction" AS ENUM(
    'IN',
    'OUT'
);

-- ============================================================
-- CREATE TABLES
-- ============================================================
CREATE TABLE "users"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "email" varchar UNIQUE NOT NULL,
    "password" varchar NOT NULL,
    "token" varchar,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

CREATE TABLE "profiles"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid UNIQUE NOT NULL,
    "full_name" varchar,
    "phone" varchar UNIQUE,
    "photo" varchar,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

CREATE TABLE "user_pins"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid UNIQUE NOT NULL,
    "pin_hash" varchar,
    "failed_attempts" int NOT NULL DEFAULT 0,
    "locked_until" timestamp,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

CREATE TABLE "wallets"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid NOT NULL,
    "label" varchar NOT NULL DEFAULT 'Wallet Utama',
    "balance" bigint NOT NULL DEFAULT 0,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

CREATE TABLE "transactions"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "wallet_id" uuid NOT NULL,
    "amount" bigint NOT NULL,
    "direction" direction NOT NULL,
    "admin_fee" bigint NOT NULL DEFAULT 0,
    "status" transaction_status NOT NULL DEFAULT 'PENDING',
    "note" text,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "topups"(
    "transaction_id" uuid PRIMARY KEY NOT NULL,
    "payment_method" payment_method,
    "external_reference" varchar UNIQUE
);

CREATE TABLE "transfers"(
    "transaction_id" uuid PRIMARY KEY NOT NULL,
    "sender_transaction_id" uuid NOT NULL,
    "recipient_wallet_id" uuid NOT NULL,
    "transfer_code" varchar UNIQUE
);

CREATE TABLE "expenses"(
    "transaction_id" uuid PRIMARY KEY NOT NULL,
    "category" varchar,
    "merchant_name" varchar
);

CREATE TABLE "withdrawals"(
    "transaction_id" uuid PRIMARY KEY NOT NULL,
    "bank_name" varchar NOT NULL,
    "account_number" varchar NOT NULL,
    "account_holder" varchar NOT NULL
);

-- ============================================================
-- COMMENTS
-- ============================================================
COMMENT ON TABLE "users" IS 'Core identity table. Credentials only — no PII stored here.';

COMMENT ON TABLE "profiles" IS 'One-to-one with users. Stores PII separately for easier data management.';

COMMENT ON TABLE "user_pins" IS 'PIN stored as bcrypt hash. locked_until enables brute-force protection — NULL means not locked.';

COMMENT ON TABLE "wallets" IS 'Supports multiple wallets per user. Balance stored in smallest currency unit (sen/IDR). label can be customized by user.';

COMMENT ON TABLE "transactions" IS 'Central ledger. direction is the single source of truth for IN/OUT. Each row is ONE side of a transaction. For transfers: two rows are created — one OUT for sender, one IN for recipient, linked via the transfers table.';

COMMENT ON TABLE "topups" IS 'external_reference: reference ID from payment gateway (DANA, GoPay, etc). unique constraint enables idempotency check.';

COMMENT ON TABLE "transfers" IS 'One row links BOTH sides of a transfer. transaction_id = OUT row (sender). sender_transaction_id = IN row (recipient). This enables full audit trail without ambiguity.';

COMMENT ON TABLE "expenses" IS 'Auto-created by system for every OUT transaction that is NOT a withdrawal. category & merchant_name can be enriched later.';

COMMENT ON TABLE "withdrawals" IS 'Manual bank withdrawal. bank_name, account_number, account_holder are snapshot values at time of withdrawal — not foreign keys to a bank account table.';

-- ============================================================
-- FOREIGN KEYS
-- ============================================================
ALTER TABLE "profiles"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_pins"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "wallets"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transactions"
    ADD FOREIGN KEY ("wallet_id") REFERENCES "wallets"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "topups"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transfers"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transfers"
    ADD FOREIGN KEY ("sender_transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "transfers"
    ADD FOREIGN KEY ("recipient_wallet_id") REFERENCES "wallets"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "expenses"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "withdrawals"
    ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions"("id") DEFERRABLE INITIALLY IMMEDIATE;

-- ============================================================
-- USERS (2 users)
-- ============================================================
INSERT INTO users(id, email, password, token, created_at)
VALUES
    ('a1b2c3d4-0001-0001-0001-000000000001', 'vando@example.com', '$2b$12$hashedpassword1111111111111111111111111111111111', 'jwt_token_user_1', now()),
('a1b2c3d4-0002-0002-0002-000000000002', 'budi@example.com', '$2b$12$hashedpassword2222222222222222222222222222222222', NULL, now());

-- ============================================================
-- PROFILES
-- ============================================================
INSERT INTO profiles(id, user_id, full_name, phone, photo, created_at)
VALUES
    ('b1000000-0001-0001-0001-000000000001', 'a1b2c3d4-0001-0001-0001-000000000001', -- vando
        'Rivando Al Rasyid', '081234567890', 'https://example.com/photos/vando.jpg', now()),
('b1000000-0002-0002-0002-000000000002', 'a1b2c3d4-0002-0002-0002-000000000002', -- budi
        'Budi Santoso', '089876543210', NULL, now());

-- ============================================================
-- USER_PINS
-- pin_hash = bcrypt("123456") — ganti dengan hash asli di production
-- ============================================================
INSERT INTO user_pins(id, user_id, pin_hash, failed_attempts, locked_until, created_at)
VALUES
    ('c1000000-0001-0001-0001-000000000001', 'a1b2c3d4-0001-0001-0001-000000000001', -- vando
        '$2b$12$pinhashabcdef111111111111111111111111111111', 0, NULL, now()),
('c1000000-0002-0002-0002-000000000002', 'a1b2c3d4-0002-0002-0002-000000000002', -- budi
        '$2b$12$pinhashabcdef222222222222222222222222222222', 3, now() + interval '30 minutes', -- locked!
        now());

-- ============================================================
-- WALLETS
-- balance dalam satuan sen/IDR terkecil
-- vando punya 2 wallet, budi 1 wallet
-- ============================================================
INSERT INTO wallets(id, user_id, label, balance, created_at)
VALUES
    ('d1000000-0001-0001-0001-000000000001', 'a1b2c3d4-0001-0001-0001-000000000001', -- vando wallet utama
        'Wallet Utama', 5000000, -- Rp 50.000 (dalam sen)
        now()),
('d1000000-0002-0002-0002-000000000002', 'a1b2c3d4-0001-0001-0001-000000000001', -- vando wallet tabungan
        'Tabungan', 10000000, -- Rp 100.000
        now()),
('d1000000-0003-0003-0003-000000000003', 'a1b2c3d4-0002-0002-0002-000000000002', -- budi wallet utama
        'Wallet Utama', 2000000, -- Rp 20.000
        now());

-- ============================================================
-- TRANSACTIONS
-- tx-0001: topup IN  (vando)
-- tx-0002: transfer OUT dari vando ke budi
-- tx-0003: transfer IN  ke budi (pasangan tx-0002)
-- tx-0004: expense OUT (vando)
-- tx-0005: withdrawal OUT (vando)
-- ============================================================
INSERT INTO transactions(id, wallet_id, amount, direction, admin_fee, status, note, created_at)
VALUES
    ('e1000000-0001-0001-0001-000000000001', -- topup IN
        'd1000000-0001-0001-0001-000000000001', -- vando wallet utama
        5000000, 'IN', 0, 'SUCCESS', 'Top up via GoPay', now()),
('e1000000-0002-0002-0002-000000000002', -- transfer OUT sender
        'd1000000-0001-0001-0001-000000000001', -- vando wallet utama
        2000000, 'OUT', 1000, 'SUCCESS', 'Transfer ke Budi', now()),
('e1000000-0003-0003-0003-000000000003', -- transfer IN recipient
        'd1000000-0003-0003-0003-000000000003', -- budi wallet utama
        2000000, 'IN', 0, 'SUCCESS', 'Terima dari Vando', now()),
('e1000000-0004-0004-0004-000000000004', -- expense OUT
        'd1000000-0001-0001-0001-000000000001', -- vando wallet utama
        1500000, 'OUT', 0, 'SUCCESS', 'Makan siang', now()),
('e1000000-0005-0005-0005-000000000005', -- withdrawal OUT
        'd1000000-0002-0002-0002-000000000002', -- vando tabungan
        3000000, 'OUT', 5000, 'SUCCESS', 'Tarik ke BCA', now());

-- ============================================================
-- TOPUPS
-- refs transaction tx-0001 (IN topup)
-- ============================================================
INSERT INTO topups(transaction_id, payment_method, external_reference)
    VALUES ('e1000000-0001-0001-0001-000000000001', 'GOPAY', 'GOPAY-REF-20240518-001');

-- ============================================================
-- TRANSFERS
-- transaction_id     = OUT row (sender)   = tx-0002
-- sender_transaction_id = IN row (recipient) = tx-0003
-- ⚠ Naming counterintuitive — lihat COMMENT di schema
-- ============================================================
INSERT INTO transfers(transaction_id, sender_transaction_id, recipient_wallet_id, transfer_code)
    VALUES ('e1000000-0002-0002-0002-000000000002', -- OUT row sender
        'e1000000-0003-0003-0003-000000000003', -- IN row recipient
        'd1000000-0003-0003-0003-000000000003', -- budi wallet
        'TRF-20240518-VANDO-BUDI');

-- ============================================================
-- EXPENSES
-- Wajib ada untuk setiap OUT yang BUKAN withdrawal
-- tx-0002 (transfer OUT) dan tx-0004 (expense OUT) keduanya butuh row di sini
-- ============================================================
INSERT INTO expenses(transaction_id, category, merchant_name)
VALUES
    ('e1000000-0002-0002-0002-000000000002', -- transfer OUT
        'transfer', NULL),
('e1000000-0004-0004-0004-000000000004', -- expense OUT
        'food', 'Warung Makan Bu Sari');

-- ============================================================
-- WITHDRAWALS
-- snapshot data — bukan FK ke bank account table
-- ============================================================
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
    VALUES ('e1000000-0005-0005-0005-000000000005', -- withdrawal OUT
        'BCA', '1234567890', 'Rivando Al Rasyid');

