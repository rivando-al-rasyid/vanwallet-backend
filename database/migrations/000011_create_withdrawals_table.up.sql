CREATE TABLE "withdrawals"(
    "transaction_id" uuid PRIMARY KEY,
    "bank_name" varchar NOT NULL,
    "account_number" varchar NOT NULL,
    "account_holder" varchar NOT NULL
);

