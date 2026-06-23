-- =============================================================
-- 5. TRANSACTIONS
--    Generate 1 month / 30 days of data
--    - Every wallet gets 1 EXPENSE per day
--    - Every 5 days gets 1 WITHDRAWAL
--    - Every 3 days gets TRANSFER_OUT + TRANSFER_IN pair
-- =============================================================
-- -------------------------------------------------------------
-- 5.1 EXPENSES
-- 5 users x 30 days = 150 expenses
-- -------------------------------------------------------------
WITH expense_rows AS (
    SELECT
        gen_random_uuid() AS transaction_id,
        w.id AS wallet_id,
        (10000 +((day_offset + user_index) % 10) * 7500)::numeric AS amount,
        0::numeric AS admin_fee,
        'SUCCESS' AS status,
        CASE ((day_offset + user_index) % 8)
        WHEN 0 THEN
            'Makan siang'
        WHEN 1 THEN
            'Transport harian'
        WHEN 2 THEN
            'Belanja minimarket'
        WHEN 3 THEN
            'Kopi dan snack'
        WHEN 4 THEN
            'Bensin kendaraan'
        WHEN 5 THEN
            'Makan malam'
        WHEN 6 THEN
            'Belanja kebutuhan rumah'
        ELSE
            'Jajan harian'
        END AS note,
        CASE ((day_offset + user_index) % 8)
        WHEN 0 THEN
            'Food & Beverage'
        WHEN 1 THEN
            'Transport'
        WHEN 2 THEN
            'Groceries'
        WHEN 3 THEN
            'Food & Beverage'
        WHEN 4 THEN
            'Transport'
        WHEN 5 THEN
            'Food & Beverage'
        WHEN 6 THEN
            'Household'
        ELSE
            'Food & Beverage'
        END AS category,
        CASE ((day_offset + user_index) % 8)
        WHEN 0 THEN
            'Warteg Barokah'
        WHEN 1 THEN
            'Gojek'
        WHEN 2 THEN
            'Indomaret'
        WHEN 3 THEN
            'Kopi Kenangan'
        WHEN 4 THEN
            'Shell'
        WHEN 5 THEN
            'RM Padang'
        WHEN 6 THEN
            'Alfamart'
        ELSE
            'Janji Jiwa'
        END AS merchant_name,
(now() -(day_offset || ' days')::interval +(((user_index + day_offset) % 12) || ' hours')::interval) AS created_at
    FROM
        wallets w
        JOIN users u ON u.id = w.user_id
        CROSS JOIN generate_series(0, 29) AS day_offset
        CROSS JOIN LATERAL (
            SELECT
                CASE u.email
                WHEN 'alice@vanwallet.id' THEN
                    1
                WHEN 'budi@vanwallet.id' THEN
                    2
                WHEN 'citra@vanwallet.id' THEN
                    3
                WHEN 'dian@vanwallet.id' THEN
                    4
                WHEN 'eko@vanwallet.id' THEN
                    5
                END AS user_index) user_map
),
inserted_expenses AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        transaction_id,
        wallet_id,
        'EXPENSE',
        amount,
        admin_fee,
        status,
        note,
        created_at
    FROM
        expense_rows
    RETURNING
        id)
INSERT INTO expenses(transaction_id, category, merchant_name)
SELECT
    transaction_id,
    category,
    merchant_name
FROM
    expense_rows;

-- -------------------------------------------------------------
-- 5.2 WITHDRAWALS
-- 5 users x 6 withdrawals = 30 withdrawals
-- every 5 days
-- -------------------------------------------------------------
WITH withdrawal_rows AS (
    SELECT
        gen_random_uuid() AS transaction_id,
        w.id AS wallet_id,
        (50000 +((day_offset + user_index) % 5) * 25000)::numeric AS amount,
        2500::numeric AS admin_fee,
        'SUCCESS' AS status,
        CASE ((day_offset + user_index) % 4)
        WHEN 0 THEN
            'Tarik tunai BCA'
        WHEN 1 THEN
            'Tarik tunai BRI'
        WHEN 2 THEN
            'Tarik tunai Mandiri'
        ELSE
            'Tarik tunai BNI'
        END AS note,
        CASE ((day_offset + user_index) % 4)
        WHEN 0 THEN
            'BCA'
        WHEN 1 THEN
            'BRI'
        WHEN 2 THEN
            'Mandiri'
        ELSE
            'BNI'
        END AS bank_name,
        CASE u.email
        WHEN 'alice@vanwallet.id' THEN
            '1234567890'
        WHEN 'budi@vanwallet.id' THEN
            '1112223330'
        WHEN 'citra@vanwallet.id' THEN
            '2223334440'
        WHEN 'dian@vanwallet.id' THEN
            '3334445550'
        WHEN 'eko@vanwallet.id' THEN
            '7778889990'
        END AS account_number,
        p.full_name AS account_holder,
(now() -(day_offset || ' days')::interval +(((user_index + day_offset + 3) % 12) || ' hours')::interval) AS created_at
    FROM
        wallets w
        JOIN users u ON u.id = w.user_id
        JOIN profiles p ON p.user_id = u.id
        CROSS JOIN generate_series(0, 29, 5) AS day_offset
        CROSS JOIN LATERAL (
            SELECT
                CASE u.email
                WHEN 'alice@vanwallet.id' THEN
                    1
                WHEN 'budi@vanwallet.id' THEN
                    2
                WHEN 'citra@vanwallet.id' THEN
                    3
                WHEN 'dian@vanwallet.id' THEN
                    4
                WHEN 'eko@vanwallet.id' THEN
                    5
                END AS user_index) user_map
),
inserted_withdrawals AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        transaction_id,
        wallet_id,
        'WITHDRAWAL',
        amount,
        admin_fee,
        status,
        note,
        created_at
    FROM
        withdrawal_rows
    RETURNING
        id)
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
SELECT
    transaction_id,
    bank_name,
    account_number,
    account_holder
FROM
    withdrawal_rows;

-- -------------------------------------------------------------
-- 5.3 TRANSFERS
-- 10 days x 5 transfer pairs = 50 outgoing + 50 incoming
-- every 3 days
-- -------------------------------------------------------------
WITH transfer_plan AS (
    SELECT
        gen_random_uuid() AS outgoing_transaction_id,
        gen_random_uuid() AS incoming_transaction_id,
        sender_wallet.id AS sender_wallet_id,
        receiver_wallet.id AS receiver_wallet_id,
        sender_profile.full_name AS sender_name,
        receiver_profile.full_name AS receiver_name,
(25000 +((day_offset + sender_map.user_index) % 8) * 10000)::numeric AS transfer_amount,
        day_offset,
        sender_map.user_index,
(now() -(day_offset || ' days')::interval +(((sender_map.user_index + day_offset + 6) % 12) || ' hours')::interval) AS created_at
    FROM
        generate_series(0, 29, 3) AS day_offset
        CROSS JOIN users sender_user
        JOIN wallets sender_wallet ON sender_wallet.user_id = sender_user.id
        JOIN profiles sender_profile ON sender_profile.user_id = sender_user.id
        CROSS JOIN LATERAL (
            SELECT
                CASE sender_user.email
                WHEN 'alice@vanwallet.id' THEN
                    1
                WHEN 'budi@vanwallet.id' THEN
                    2
                WHEN 'citra@vanwallet.id' THEN
                    3
                WHEN 'dian@vanwallet.id' THEN
                    4
                WHEN 'eko@vanwallet.id' THEN
                    5
                END AS user_index) sender_map
            JOIN users receiver_user ON receiver_user.email = CASE sender_user.email
            WHEN 'alice@vanwallet.id' THEN
                'budi@vanwallet.id'
            WHEN 'budi@vanwallet.id' THEN
                'citra@vanwallet.id'
            WHEN 'citra@vanwallet.id' THEN
                'dian@vanwallet.id'
            WHEN 'dian@vanwallet.id' THEN
                'eko@vanwallet.id'
            WHEN 'eko@vanwallet.id' THEN
                'alice@vanwallet.id'
            END
            JOIN wallets receiver_wallet ON receiver_wallet.user_id = receiver_user.id
            JOIN profiles receiver_profile ON receiver_profile.user_id = receiver_user.id
),
insert_outgoing AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        outgoing_transaction_id,
        sender_wallet_id,
        'TRANSFER_OUT',
        transfer_amount,
        0,
        'SUCCESS',
        'Transfer ke ' || receiver_name,
        created_at
    FROM
        transfer_plan
    RETURNING
        id
),
insert_incoming AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        incoming_transaction_id,
        receiver_wallet_id,
        'TRANSFER_IN',
        transfer_amount,
        0,
        'SUCCESS',
        'Terima dari ' || sender_name,
        created_at
    FROM
        transfer_plan
    RETURNING
        id)
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
SELECT
    outgoing_transaction_id,
    incoming_transaction_id,
    'TRF-' || upper(substr(replace(sender_name, ' ', ''), 1, 5)) || '-' || upper(substr(replace(receiver_name, ' ', ''), 1, 5)) || '-' || lpad(day_offset::text, 2, '0') || '-' || lpad(user_index::text, 2, '0'),
    created_at
FROM
    transfer_plan;

-- =============================================================
-- VERIFICATION QUERIES
-- =============================================================
-- Total transactions per user
-- Expected around:
-- 30 expenses + 6 withdrawals + 10 transfer_out + 10 transfer_in = 56/user
SELECT
    u.email,
    count(t.id) AS transaction_count,
    min(t.created_at) AS oldest_transaction,
    max(t.created_at) AS newest_transaction
FROM
    users u
    JOIN wallets w ON w.user_id = u.id
    JOIN transactions t ON t.wallet_id = w.id
GROUP BY
    u.email
ORDER BY
    u.email;

-- Transaction count by type
SELECT
    t.type,
    count(*) AS total
FROM
    transactions t
GROUP BY
    t.type
ORDER BY
    t.type;

-- Daily transaction count for last 30 days
SELECT
    date(t.created_at) AS transaction_date,
    count(*) AS total_transactions
FROM
    transactions t
WHERE
    t.created_at >= now() - interval '30 days'
GROUP BY
    date(t.created_at)
ORDER BY
    transaction_date;

-- =============================================================
-- 5. TRANSACTIONS
--    Generate exactly 200 transactions
--
--    100 EXPENSE
--     30 WITHDRAWAL
--     35 TRANSFER_OUT
--     35 TRANSFER_IN
--    ----------------
--    200 TOTAL
-- =============================================================
-- -------------------------------------------------------------
-- 5.1 EXPENSES
-- 5 users x 20 expenses = 100 transactions
-- -------------------------------------------------------------
WITH expense_rows AS (
    SELECT
        gen_random_uuid() AS transaction_id,
        w.id AS wallet_id,
        (10000 +((expense_number + user_index) % 10) * 7500)::numeric AS amount,
        0::numeric AS admin_fee,
        'SUCCESS' AS status,
        CASE ((expense_number + user_index) % 8)
        WHEN 0 THEN
            'Makan siang'
        WHEN 1 THEN
            'Transport harian'
        WHEN 2 THEN
            'Belanja minimarket'
        WHEN 3 THEN
            'Kopi dan snack'
        WHEN 4 THEN
            'Bensin kendaraan'
        WHEN 5 THEN
            'Makan malam'
        WHEN 6 THEN
            'Belanja kebutuhan rumah'
        ELSE
            'Jajan harian'
        END AS note,
        CASE ((expense_number + user_index) % 8)
        WHEN 0 THEN
            'Food & Beverage'
        WHEN 1 THEN
            'Transport'
        WHEN 2 THEN
            'Groceries'
        WHEN 3 THEN
            'Food & Beverage'
        WHEN 4 THEN
            'Transport'
        WHEN 5 THEN
            'Food & Beverage'
        WHEN 6 THEN
            'Household'
        ELSE
            'Food & Beverage'
        END AS category,
        CASE ((expense_number + user_index) % 8)
        WHEN 0 THEN
            'Warteg Barokah'
        WHEN 1 THEN
            'Gojek'
        WHEN 2 THEN
            'Indomaret'
        WHEN 3 THEN
            'Kopi Kenangan'
        WHEN 4 THEN
            'Shell'
        WHEN 5 THEN
            'RM Padang'
        WHEN 6 THEN
            'Alfamart'
        ELSE
            'Janji Jiwa'
        END AS merchant_name,
(now() -(((expense_number * 2 + user_index) % 30) || ' days')::interval +(((expense_number + user_index) % 12) || ' hours')::interval) AS created_at
    FROM
        wallets w
        JOIN users u ON u.id = w.user_id
        CROSS JOIN generate_series(1, 20) AS expense_number
        CROSS JOIN LATERAL (
            SELECT
                CASE u.email
                WHEN 'alice@vanwallet.id' THEN
                    1
                WHEN 'budi@vanwallet.id' THEN
                    2
                WHEN 'citra@vanwallet.id' THEN
                    3
                WHEN 'dian@vanwallet.id' THEN
                    4
                WHEN 'eko@vanwallet.id' THEN
                    5
                END AS user_index) user_map
),
inserted_expenses AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        transaction_id,
        wallet_id,
        'EXPENSE',
        amount,
        admin_fee,
        status,
        note,
        created_at
    FROM
        expense_rows
    RETURNING
        id)
INSERT INTO expenses(transaction_id, category, merchant_name)
SELECT
    transaction_id,
    category,
    merchant_name
FROM
    expense_rows;

-- -------------------------------------------------------------
-- 5.2 WITHDRAWALS
-- 5 users x 6 withdrawals = 30 transactions
-- -------------------------------------------------------------
WITH withdrawal_rows AS (
    SELECT
        gen_random_uuid() AS transaction_id,
        w.id AS wallet_id,
        (50000 +((withdrawal_number + user_index) % 5) * 25000)::numeric AS amount,
        2500::numeric AS admin_fee,
        'SUCCESS' AS status,
        CASE ((withdrawal_number + user_index) % 4)
        WHEN 0 THEN
            'Tarik tunai BCA'
        WHEN 1 THEN
            'Tarik tunai BRI'
        WHEN 2 THEN
            'Tarik tunai Mandiri'
        ELSE
            'Tarik tunai BNI'
        END AS note,
        CASE ((withdrawal_number + user_index) % 4)
        WHEN 0 THEN
            'BCA'
        WHEN 1 THEN
            'BRI'
        WHEN 2 THEN
            'Mandiri'
        ELSE
            'BNI'
        END AS bank_name,
        CASE u.email
        WHEN 'alice@vanwallet.id' THEN
            '1234567890'
        WHEN 'budi@vanwallet.id' THEN
            '1112223330'
        WHEN 'citra@vanwallet.id' THEN
            '2223334440'
        WHEN 'dian@vanwallet.id' THEN
            '3334445550'
        WHEN 'eko@vanwallet.id' THEN
            '7778889990'
        END AS account_number,
        p.full_name AS account_holder,
(now() -(((withdrawal_number * 4 + user_index) % 30) || ' days')::interval +(((withdrawal_number + user_index + 3) % 12) || ' hours')::interval) AS created_at
    FROM
        wallets w
        JOIN users u ON u.id = w.user_id
        JOIN profiles p ON p.user_id = u.id
        CROSS JOIN generate_series(1, 6) AS withdrawal_number
        CROSS JOIN LATERAL (
            SELECT
                CASE u.email
                WHEN 'alice@vanwallet.id' THEN
                    1
                WHEN 'budi@vanwallet.id' THEN
                    2
                WHEN 'citra@vanwallet.id' THEN
                    3
                WHEN 'dian@vanwallet.id' THEN
                    4
                WHEN 'eko@vanwallet.id' THEN
                    5
                END AS user_index) user_map
),
inserted_withdrawals AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        transaction_id,
        wallet_id,
        'WITHDRAWAL',
        amount,
        admin_fee,
        status,
        note,
        created_at
    FROM
        withdrawal_rows
    RETURNING
        id)
INSERT INTO withdrawals(transaction_id, bank_name, account_number, account_holder)
SELECT
    transaction_id,
    bank_name,
    account_number,
    account_holder
FROM
    withdrawal_rows;

-- -------------------------------------------------------------
-- 5.3 TRANSFERS
-- 35 transfer pairs = 70 transactions
-- 35 TRANSFER_OUT + 35 TRANSFER_IN
-- -------------------------------------------------------------
WITH transfer_plan AS (
    SELECT
        gen_random_uuid() AS outgoing_transaction_id,
        gen_random_uuid() AS incoming_transaction_id,
        sender_wallet.id AS sender_wallet_id,
        receiver_wallet.id AS receiver_wallet_id,
        sender_profile.full_name AS sender_name,
        receiver_profile.full_name AS receiver_name,
(25000 +((transfer_number + sender_map.user_index) % 8) * 10000)::numeric AS transfer_amount,
        transfer_number,
        sender_map.user_index,
(now() -(((transfer_number * 3 + sender_map.user_index) % 30) || ' days')::interval +(((transfer_number + sender_map.user_index + 6) % 12) || ' hours')::interval) AS created_at
    FROM
        users sender_user
        JOIN wallets sender_wallet ON sender_wallet.user_id = sender_user.id
        JOIN profiles sender_profile ON sender_profile.user_id = sender_user.id
        CROSS JOIN generate_series(1, 7) AS transfer_number
        CROSS JOIN LATERAL (
            SELECT
                CASE sender_user.email
                WHEN 'alice@vanwallet.id' THEN
                    1
                WHEN 'budi@vanwallet.id' THEN
                    2
                WHEN 'citra@vanwallet.id' THEN
                    3
                WHEN 'dian@vanwallet.id' THEN
                    4
                WHEN 'eko@vanwallet.id' THEN
                    5
                END AS user_index) sender_map
            JOIN users receiver_user ON receiver_user.email = CASE sender_user.email
            WHEN 'alice@vanwallet.id' THEN
                'budi@vanwallet.id'
            WHEN 'budi@vanwallet.id' THEN
                'citra@vanwallet.id'
            WHEN 'citra@vanwallet.id' THEN
                'dian@vanwallet.id'
            WHEN 'dian@vanwallet.id' THEN
                'eko@vanwallet.id'
            WHEN 'eko@vanwallet.id' THEN
                'alice@vanwallet.id'
            END
            JOIN wallets receiver_wallet ON receiver_wallet.user_id = receiver_user.id
            JOIN profiles receiver_profile ON receiver_profile.user_id = receiver_user.id
),
insert_outgoing AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        outgoing_transaction_id,
        sender_wallet_id,
        'TRANSFER_OUT',
        transfer_amount,
        0,
        'SUCCESS',
        'Transfer ke ' || receiver_name,
        created_at
    FROM
        transfer_plan
    RETURNING
        id
),
insert_incoming AS (
INSERT INTO transactions(id, wallet_id, type, amount, admin_fee, status, note, created_at)
    SELECT
        incoming_transaction_id,
        receiver_wallet_id,
        'TRANSFER_IN',
        transfer_amount,
        0,
        'SUCCESS',
        'Terima dari ' || sender_name,
        created_at
    FROM
        transfer_plan
    RETURNING
        id)
INSERT INTO transfers(transaction_id, recipient_transaction_id, transfer_code, created_at)
SELECT
    outgoing_transaction_id,
    incoming_transaction_id,
    'TRF-' || upper(substr(replace(sender_name, ' ', ''), 1, 5)) || '-' || upper(substr(replace(receiver_name, ' ', ''), 1, 5)) || '-' || lpad(transfer_number::text, 2, '0') || '-' || lpad(user_index::text, 2, '0'),
    created_at
FROM
    transfer_plan;

-- =============================================================
-- VERIFICATION
-- =============================================================
-- Must return 200
SELECT
    count(*) AS total_transactions
FROM
    transactions;

-- Check by type
SELECT
    type,
    count(*) AS total
FROM
    transactions
GROUP BY
    type
ORDER BY
    type;

-- Check per user
SELECT
    u.email,
    count(t.id) AS transaction_count
FROM
    users u
    JOIN wallets w ON w.user_id = u.id
    JOIN transactions t ON t.wallet_id = w.id
GROUP BY
    u.email
ORDER BY
    u.email;

-- Check date range
SELECT
    min(created_at) AS oldest_transaction,
    max(created_at) AS newest_transaction
FROM
    transactions;

