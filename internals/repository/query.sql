WITH new_user AS
    (INSERT INTO users (email, password)
     VALUES ('vando@mail.com',
             'hashed_password_placeholder')-- Tanda kutip ditambahkan
 RETURNING id,
           email,
           created_at),
     new_profile AS
    (INSERT INTO profiles (user_id) SELECT id
     FROM new_user),
     new_pin AS
    (INSERT INTO user_pins (user_id) SELECT id
     FROM new_user),
     new_wallet AS
    (INSERT INTO wallets (user_id, balance) SELECT id,
                                                   0
     FROM new_user)
SELECT id,
       email,
       created_at
FROM new_user;


SELECT u.id,
       u.email,
       u.password,
       u.token,
       p.id,
       w.label
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
LEFT JOIN wallets w ON u.id = w.user_id
WHERE email = 'test@example.com';


SELECT u.id,
       u.email,
       u.password,
       u.token,
FROM users u
WHERE email = 'test@example.com'