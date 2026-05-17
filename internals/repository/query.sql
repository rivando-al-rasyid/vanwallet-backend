WITH new_user AS (
INSERT INTO users(email, password)
        VALUES ('vando@mail.com', 'hashed_password_placeholder') -- Tanda kutip ditambahkan
    RETURNING
        id, email, created_at
), new_profile AS (
INSERT INTO profiles(user_id)
    SELECT
        id
    FROM
        new_user
),
new_pin AS (
INSERT INTO user_pins(user_id)
    SELECT
        id
    FROM
        new_user
),
new_wallet AS (
INSERT INTO wallets(user_id, balance)
    SELECT
        id,
        0
    FROM
        new_user
)
SELECT
    id,
    email,
    created_at
FROM
    new_user;

SELECT
    u.email,
    p.full_name,
    p.phone,
    p.photo
FROM
    users u
    LEFT JOIN profiles p ON u.id = p.user_id
WHERE
    email = 'test@example.com';

SELECT
    w.label,
    w.balance,
    sum(
        CASE WHEN w.direction = "IN") AS income,
    sum(
        CASE WHEN w.direction = "OUT") AS expense
FROM
    users u
    JOIN wallets w ON u.id = w.user_id
    LEFT JOIN transactions t ON w.id = wallet_id
WHERE
    email = 'test@example.com';
