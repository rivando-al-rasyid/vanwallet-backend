SELECT
    w.label,
    w.balance,
    sum(
        CASE WHEN t.direction = 'IN' THEN
            t.amount
        ELSE
            0
        END) AS income,
    sum(
        CASE WHEN t.direction = 'OUT' THEN
            t.amount
        ELSE
            0
        END) AS expense
FROM
    users u
    JOIN wallets w ON u.id = w.user_id
    LEFT JOIN transactions t ON w.id = t.wallet_id
WHERE
    u.email = 'test@example.com'
    AND t.status = status
GROUP BY
    u.id,
    w.label,
    w.balance;

