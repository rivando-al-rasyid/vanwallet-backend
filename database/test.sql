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

