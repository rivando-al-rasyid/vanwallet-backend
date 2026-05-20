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

