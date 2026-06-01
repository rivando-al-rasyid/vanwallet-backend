-- =============================================================
-- VanWallet Seeder
-- 5 Users | 1 Wallet each | 10 Transactions per user
-- Password: "password123" → bcrypt hash (cost 10)
-- PIN: "123456" → bcrypt hash (cost 10)
-- =============================================================
BEGIN;
-- -------------------------------------------------------------
-- 1. USERS (no username — email-only identity)
-- -------------------------------------------------------------
INSERT INTO users(id, email, password, created_at)
VALUES
    ('a1000000-0000-0000-0000-000000000001', 'alice@vanwallet.id', '$argon2id$v=19$m=65536,t=2,p=1$voAbbGXbg+iItRUSAQVMvw$rpbnzPGeJKmMPWf0nmYuTiHmgfhwBPdusWW0v7XmslU', now() - interval '30 days'),
('a2000000-0000-0000-0000-000000000002', 'budi@vanwallet.id', '$argon2id$v=19$m=65536,t=2,p=1$voAbbGXbg+iItRUSAQVMvw$rpbnzPGeJKmMPWf0nmYuTiHmgfhwBPdusWW0v7XmslU', now() - interval '29 days'),
('a3000000-0000-0000-0000-000000000003', 'citra@vanwallet.id', '$argon2id$v=19$m=65536,t=2,p=1$voAbbGXbg+iItRUSAQVMvw$rpbnzPGeJKmMPWf0nmYuTiHmgfhwBPdusWW0v7XmslU', now() - interval '28 days'),
('a4000000-0000-0000-0000-000000000004', 'dian@vanwallet.id', '$argon2id$v=19$m=65536,t=2,p=1$voAbbGXbg+iItRUSAQVMvw$rpbnzPGeJKmMPWf0nmYuTiHmgfhwBPdusWW0v7XmslU', now() - interval '27 days'),
('a5000000-0000-0000-0000-000000000005', 'eko@vanwallet.id', '$argon2id$v=19$m=65536,t=2,p=1$voAbbGXbg+iItRUSAQVMvw$rpbnzPGeJKmMPWf0nmYuTiHmgfhwBPdusWW0v7XmslU', now() - interval '26 days');
-- -------------------------------------------------------------
-- 2. PROFILES
-- -------------------------------------------------------------
INSERT INTO profiles(id, user_id, full_name, phone, created_at)
VALUES
    (gen_random_uuid(), 'a1000000-0000-0000-0000-000000000001', 'Alice Pratiwi', '081200000001', now() - interval '30 days'),
(gen_random_uuid(), 'a2000000-0000-0000-0000-000000000002', 'Budi Santoso', '081200000002', now() - interval '29 days'),
(gen_random_uuid(), 'a3000000-0000-0000-0000-000000000003', 'Citra Dewi', '081200000003', now() - interval '28 days'),
(gen_random_uuid(), 'a4000000-0000-0000-0000-000000000004', 'Dian Rahayu', '081200000004', now() - interval '27 days'),
(gen_random_uuid(), 'a5000000-0000-0000-0000-000000000005', 'Eko Prasetyo', '081200000005', now() - interval '26 days');
-- -------------------------------------------------------------
-- 3. USER PINS  (PIN: 123456)
-- -------------------------------------------------------------
INSERT INTO user_pins(id, user_id, pin_hash, failed_attempts, created_at)
VALUES
    (gen_random_uuid(), 'a1000000-0000-0000-0000-000000000001', '$argon2id$v=19$m=65536,t=2,p=1$9DwFBejN6ozLRBnw52d8Cg$UEWko0TO7DDXYH1CBYSFFhBsIA+MP1jQPARZm/O6HbM', 0, now()),
(gen_random_uuid(), 'a2000000-0000-0000-0000-000000000002', '$argon2id$v=19$m=65536,t=2,p=1$9DwFBejN6ozLRBnw52d8Cg$UEWko0TO7DDXYH1CBYSFFhBsIA+MP1jQPARZm/O6HbM', 0, now()),
(gen_random_uuid(), 'a3000000-0000-0000-0000-000000000003', '$argon2id$v=19$m=65536,t=2,p=1$9DwFBejN6ozLRBnw52d8Cg$UEWko0TO7DDXYH1CBYSFFhBsIA+MP1jQPARZm/O6HbM', 0, now()),
(gen_random_uuid(), 'a4000000-0000-0000-0000-000000000004', '$argon2id$v=19$m=65536,t=2,p=1$9DwFBejN6ozLRBnw52d8Cg$UEWko0TO7DDXYH1CBYSFFhBsIA+MP1jQPARZm/O6HbM', 0, now()),
(gen_random_uuid(), 'a5000000-0000-0000-0000-000000000005', '$argon2id$v=19$m=65536,t=2,p=1$9DwFBejN6ozLRBnw52d8Cg$UEWko0TO7DDXYH1CBYSFFhBsIA+MP1jQPARZm/O6HbM', 0, now());
-- -------------------------------------------------------------
-- 4. WALLETS (1 per user, initial balance 1.000.000 IDR)
-- -------------------------------------------------------------
INSERT INTO wallets(id, user_id, label, balance, created_at)
VALUES
    ('b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Wallet Utama', 1000000, now() - interval '30 days'),
('b2000000-0000-0000-0000-000000000002', 'a2000000-0000-0000-0000-000000000002', 'Wallet Utama', 1000000, now() - interval '29 days'),
('b3000000-0000-0000-0000-000000000003', 'a3000000-0000-0000-0000-000000000003', 'Wallet Utama', 1000000, now() - interval '28 days'),
('b4000000-0000-0000-0000-000000000004', 'a4000000-0000-0000-0000-000000000004', 'Wallet Utama', 1000000, now() - interval '27 days'),
('b5000000-0000-0000-0000-000000000005', 'a5000000-0000-0000-0000-000000000005', 'Wallet Utama', 1000000, now() - interval '26 days');
-- =============================================================
-- 5. TRANSACTIONS  (10 per user = 50 total)
--    Mix: EXPENSE(4) | WITHDRAWAL(2) | TRANSFER_OUT(2) | TRANSFER_IN(2)
--    Transfer pairs are linked via the transfers table.
-- =============================================================
-- ─── ALICE (wallet b1) ───────────────────────────────────────
-- EXPENSE x4
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c1010000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'EXPENSE', -50000, 0, 'SUCCESS', 'Makan siang', now() - interval '25 days'),
('c1020000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'EXPENSE', -30000, 0, 'SUCCESS', 'Grab ke kantor', now() - interval '24 days'),
('c1030000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'EXPENSE', -120000, 0, 'SUCCESS', 'Belanja bulanan', now() - interval '23 days'),
('c1040000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'EXPENSE', -20000, 0, 'SUCCESS', 'Kopi dan snack', now() - interval '22 days');
INSERT INTO expenses(transaction_id, category, merchant_name)
VALUES
    ('c1010000-0000-0000-0000-000000000001', 'Food & Beverage', 'Warteg Barokah'),
('c1020000-0000-0000-0000-000000000001', 'Transport', 'Grab'),
('c1030000-0000-0000-0000-000000000001', 'Groceries', 'Indomaret'),
('c1040000-0000-0000-0000-000000000001', 'Food & Beverage', 'Kopi Kenangan');
-- WITHDRAWAL x2
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c1050000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'WITHDRAWAL', -200000, 2500, 'SUCCESS', 'Tarik tunai BCA', now() - interval '21 days'),
('c1060000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'WITHDRAWAL', -100000, 2500, 'SUCCESS', 'Tarik tunai BRI', now() - interval '20 days');
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
VALUES
    ('c1050000-0000-0000-0000-000000000001', 'BCA', '1234567890', 'Alice Pratiwi'),
('c1060000-0000-0000-0000-000000000001', 'BRI', '0987654321', 'Alice Pratiwi');
-- TRANSFER_OUT → Budi (TRANSFER_IN)
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c1070000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'TRANSFER_OUT', -75000, 0, 'SUCCESS', 'Transfer ke Budi', now() - interval '19 days'),
('c2080000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'TRANSFER_IN', 75000, 0, 'SUCCESS', 'Terima dari Alice', now() - interval '19 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c1070000-0000-0000-0000-000000000001', 'c2080000-0000-0000-0000-000000000002', 'TRF-ALICE-BUDI-001', now() - interval '19 days');
-- TRANSFER_OUT → Citra (TRANSFER_IN)
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c1090000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'TRANSFER_OUT', -50000, 0, 'SUCCESS', 'Transfer ke Citra', now() - interval '18 days'),
('c3100000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'TRANSFER_IN', 50000, 0, 'SUCCESS', 'Terima dari Alice', now() - interval '18 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c1090000-0000-0000-0000-000000000001', 'c3100000-0000-0000-0000-000000000003', 'TRF-ALICE-CITRA-001', now() - interval '18 days');
-- ─── BUDI (wallet b2) ────────────────────────────────────────
-- Note: 2 slots already used by Alice's transfers above (c2080000)
-- Budi has 9 own + 1 received = 10 total
-- EXPENSE x4
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c2010000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'EXPENSE', -45000, 0, 'SUCCESS', 'Makan siang', now() - interval '25 days'),
('c2020000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'EXPENSE', -25000, 0, 'SUCCESS', 'Parkir dan tol', now() - interval '24 days'),
('c2030000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'EXPENSE', -90000, 0, 'SUCCESS', 'Bensin', now() - interval '23 days'),
('c2040000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'EXPENSE', -15000, 0, 'SUCCESS', 'Jajan anak', now() - interval '22 days');
INSERT INTO expenses(transaction_id, category, merchant_name)
VALUES
    ('c2010000-0000-0000-0000-000000000002', 'Food & Beverage', 'RM Padang'),
('c2020000-0000-0000-0000-000000000002', 'Transport', 'Jasa Marga'),
('c2030000-0000-0000-0000-000000000002', 'Transport', 'Shell'),
('c2040000-0000-0000-0000-000000000002', 'Food & Beverage', 'Alfamart');
-- WITHDRAWAL x2
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c2050000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'WITHDRAWAL', -150000, 2500, 'SUCCESS', 'Tarik tunai Mandiri', now() - interval '21 days'),
('c2060000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'WITHDRAWAL', -80000, 2500, 'SUCCESS', 'Tarik tunai BNI', now() - interval '20 days');
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
VALUES
    ('c2050000-0000-0000-0000-000000000002', 'Mandiri', '1112223330', 'Budi Santoso'),
('c2060000-0000-0000-0000-000000000002', 'BNI', '4445556660', 'Budi Santoso');
-- TRANSFER_OUT → Dian (TRANSFER_IN)
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c2070000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'TRANSFER_OUT', -60000, 0, 'SUCCESS', 'Transfer ke Dian', now() - interval '17 days'),
('c4090000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'TRANSFER_IN', 60000, 0, 'SUCCESS', 'Terima dari Budi', now() - interval '17 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c2070000-0000-0000-0000-000000000002', 'c4090000-0000-0000-0000-000000000004', 'TRF-BUDI-DIAN-001', now() - interval '17 days');
-- ─── CITRA (wallet b3) ───────────────────────────────────────
-- Note: 1 slot used by Alice's transfer (c3100000)
-- EXPENSE x4
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c3010000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'EXPENSE', -35000, 0, 'SUCCESS', 'Lunch bareng tim', now() - interval '25 days'),
('c3020000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'EXPENSE', -20000, 0, 'SUCCESS', 'Ojek online', now() - interval '24 days'),
('c3030000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'EXPENSE', -75000, 0, 'SUCCESS', 'Skincare minimarket', now() - interval '23 days'),
('c3040000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'EXPENSE', -10000, 0, 'SUCCESS', 'Top up e-toll', now() - interval '22 days');
INSERT INTO expenses(transaction_id, category, merchant_name)
VALUES
    ('c3010000-0000-0000-0000-000000000003', 'Food & Beverage', 'GoPay Food'),
('c3020000-0000-0000-0000-000000000003', 'Transport', 'Gojek'),
('c3030000-0000-0000-0000-000000000003', 'Health & Beauty', 'Guardian'),
('c3040000-0000-0000-0000-000000000003', 'Transport', 'Jasa Marga');
-- WITHDRAWAL x2
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c3050000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'WITHDRAWAL', -100000, 2500, 'SUCCESS', 'Tarik tunai BCA', now() - interval '21 days'),
('c3060000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'WITHDRAWAL', -50000, 2500, 'SUCCESS', 'Tarik tunai BRI', now() - interval '20 days');
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
VALUES
    ('c3050000-0000-0000-0000-000000000003', 'BCA', '2223334440', 'Citra Dewi'),
('c3060000-0000-0000-0000-000000000003', 'BRI', '5556667770', 'Citra Dewi');
-- TRANSFER_OUT → Eko (TRANSFER_IN)
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c3070000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'TRANSFER_OUT', -40000, 0, 'SUCCESS', 'Transfer ke Eko', now() - interval '16 days'),
('c5090000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'TRANSFER_IN', 40000, 0, 'SUCCESS', 'Terima dari Citra', now() - interval '16 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c3070000-0000-0000-0000-000000000003', 'c5090000-0000-0000-0000-000000000005', 'TRF-CITRA-EKO-001', now() - interval '16 days');
-- TRANSFER_OUT → Alice (TRANSFER_IN back)
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c3080000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'TRANSFER_OUT', -30000, 0, 'SUCCESS', 'Bayar utang Alice', now() - interval '15 days'),
('c1110000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'TRANSFER_IN', 30000, 0, 'SUCCESS', 'Terima dari Citra', now() - interval '15 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c3080000-0000-0000-0000-000000000003', 'c1110000-0000-0000-0000-000000000001', 'TRF-CITRA-ALICE-001', now() - interval '15 days');
-- ─── DIAN (wallet b4) ────────────────────────────────────────
-- Note: 1 slot used by Budi's transfer (c4090000)
-- EXPENSE x4
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c4010000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'EXPENSE', -55000, 0, 'SUCCESS', 'Makan malam', now() - interval '25 days'),
('c4020000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'EXPENSE', -18000, 0, 'SUCCESS', 'Parkir mall', now() - interval '24 days'),
('c4030000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'EXPENSE', -60000, 0, 'SUCCESS', 'Buku pelajaran', now() - interval '23 days'),
('c4040000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'EXPENSE', -22000, 0, 'SUCCESS', 'Minuman Chatime', now() - interval '22 days');
INSERT INTO expenses(transaction_id, category, merchant_name)
VALUES
    ('c4010000-0000-0000-0000-000000000004', 'Food & Beverage', 'Sate Khas Senayan'),
('c4020000-0000-0000-0000-000000000004', 'Transport', 'Empark'),
('c4030000-0000-0000-0000-000000000004', 'Education', 'Gramedia'),
('c4040000-0000-0000-0000-000000000004', 'Food & Beverage', 'Chatime');
-- WITHDRAWAL x2
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c4050000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'WITHDRAWAL', -200000, 2500, 'SUCCESS', 'Tarik tunai Mandiri', now() - interval '21 days'),
('c4060000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'WITHDRAWAL', -75000, 2500, 'SUCCESS', 'Tarik tunai BCA', now() - interval '20 days');
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
VALUES
    ('c4050000-0000-0000-0000-000000000004', 'Mandiri', '3334445550', 'Dian Rahayu'),
('c4060000-0000-0000-0000-000000000004', 'BCA', '6667778880', 'Dian Rahayu');
-- TRANSFER_OUT → Eko
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c4070000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'TRANSFER_OUT', -45000, 0, 'SUCCESS', 'Transfer ke Eko', now() - interval '14 days'),
('c5100000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'TRANSFER_IN', 45000, 0, 'SUCCESS', 'Terima dari Dian', now() - interval '14 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c4070000-0000-0000-0000-000000000004', 'c5100000-0000-0000-0000-000000000005', 'TRF-DIAN-EKO-001', now() - interval '14 days');
-- TRANSFER_OUT → Budi (bonus transfer back)
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c4080000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'TRANSFER_OUT', -55000, 0, 'SUCCESS', 'Bayar utang Budi', now() - interval '13 days'),
('c2110000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'TRANSFER_IN', 55000, 0, 'SUCCESS', 'Terima dari Dian', now() - interval '13 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c4080000-0000-0000-0000-000000000004', 'c2110000-0000-0000-0000-000000000002', 'TRF-DIAN-BUDI-001', now() - interval '13 days');
-- ─── EKO (wallet b5) ─────────────────────────────────────────
-- Note: 2 slots used by Citra & Dian transfers (c5090000, c5100000)
-- EXPENSE x4
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c5010000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'EXPENSE', -40000, 0, 'SUCCESS', 'Makan siang warteg', now() - interval '25 days'),
('c5020000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'EXPENSE', -30000, 0, 'SUCCESS', 'KRL commuter', now() - interval '24 days'),
('c5030000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'EXPENSE', -85000, 0, 'SUCCESS', 'Beli baju diskon', now() - interval '23 days'),
('c5040000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'EXPENSE', -12000, 0, 'SUCCESS', 'Rokok & minuman', now() - interval '22 days');
INSERT INTO expenses(transaction_id, category, merchant_name)
VALUES
    ('c5010000-0000-0000-0000-000000000005', 'Food & Beverage', 'Warung Bu Yati'),
('c5020000-0000-0000-0000-000000000005', 'Transport', 'KAI Commuter'),
('c5030000-0000-0000-0000-000000000005', 'Fashion', 'Matahari'),
('c5040000-0000-0000-0000-000000000005', 'Food & Beverage', 'Alfamart');
-- WITHDRAWAL x2
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c5050000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'WITHDRAWAL', -120000, 2500, 'SUCCESS', 'Tarik tunai BRI', now() - interval '21 days'),
('c5060000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'WITHDRAWAL', -60000, 2500, 'SUCCESS', 'Tarik tunai Mandiri', now() - interval '20 days');
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
VALUES
    ('c5050000-0000-0000-0000-000000000005', 'BRI', '7778889990', 'Eko Prasetyo'),
('c5060000-0000-0000-0000-000000000005', 'Mandiri', '0001112220', 'Eko Prasetyo');
-- TRANSFER_OUT → Alice
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c5070000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'TRANSFER_OUT', -35000, 0, 'SUCCESS', 'Transfer ke Alice', now() - interval '12 days'),
('c1120000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'TRANSFER_IN', 35000, 0, 'SUCCESS', 'Terima dari Eko', now() - interval '12 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c5070000-0000-0000-0000-000000000005', 'c1120000-0000-0000-0000-000000000001', 'TRF-EKO-ALICE-001', now() - interval '12 days');
-- TRANSFER_OUT → Budi
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
VALUES
    ('c5080000-0000-0000-0000-000000000005', 'b5000000-0000-0000-0000-000000000005', 'TRANSFER_OUT', -50000, 0, 'SUCCESS', 'Transfer ke Budi', now() - interval '11 days'),
('c2120000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'TRANSFER_IN', 50000, 0, 'SUCCESS', 'Terima dari Eko', now() - interval '11 days');
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
    VALUES ('c5080000-0000-0000-0000-000000000005', 'c2120000-0000-0000-0000-000000000002', 'TRF-EKO-BUDI-001', now() - interval '11 days');
COMMIT;

-- =============================================================
-- VERIFICATION QUERIES (optional, uncomment to check)
-- =============================================================
-- SELECT u.email, COUNT(t.id) AS tx_count
-- FROM users u
-- JOIN wallets w ON w.user_id = u.id
-- JOIN transactions t ON t.wallet_id = w.id
-- GROUP BY u.email
-- ORDER BY u.email;
