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

CREATE TYPE "transaction_type" AS ENUM(
    'EXPENSE',
    'WITHDRAWAL',
    'TRANSFER_IN',
    'TRANSFER_OUT'
);

