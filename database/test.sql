SELECT
    p.user_id,
    p.full_name,
    p.phone,
    p.photo,
    p.created_at,
    p.updated_at
FROM
    profiles p
    JOIN users u ON p.user_id = u.id
WHERE
    u.email = 'vando@example.com';

SELECT
    up.pin_hash
FROM
    user_pins up
    JOIN users u ON up.user_id = u.id
WHERE
    u.email = 'vando@example.com';

-- Update rows in 'user_pins' where condition is met
UPDATE
    users
SET
    PASSWORD = 'new_value'
WHERE
    email = 'vando@example.com';

WITH TargetUser AS (
    -- 1. Cari user berdasarkan email
    SELECT
        id
    FROM
        users
    WHERE
        email = 'user@example.com' -- Ganti dengan email yang dicari
),
UserWallets AS (
    -- 2. Ambil semua wallet milik user tersebut
    SELECT
        w.id,
        w.balance
    FROM
        wallets w
        INNER JOIN TargetUser tu ON w.user_id = tu.id
)
SELECT
    -- 3. Total Balance dari tabel wallets
(
        SELECT
            coalesce(sum(balance), 0)
        FROM
            UserWallets) AS current_balance,
    -- 4. Total Income (Transaksi IN yang SUCCESS)
    coalesce(sum(
            CASE WHEN t.direction = 'IN'
                AND t.status = 'SUCCESS' THEN
                t.amount
            ELSE
                0
            END), 0) AS total_income,
    -- 5. Total Expense (Transaksi OUT yang SUCCESS + Admin Fee)
    coalesce(sum(
            CASE WHEN t.direction = 'OUT'
                AND t.status = 'SUCCESS' THEN
                (t.amount + t.admin_fee)
            ELSE
                0
            END), 0) AS total_expense
FROM
    transactions t
WHERE
    t.wallet_id IN (
        SELECT
            id
        FROM
            UserWallets);

