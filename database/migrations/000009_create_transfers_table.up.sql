CREATE TABLE "transfers"(
    "transaction_id" uuid PRIMARY KEY,
    "recipient_transaction_id" uuid UNIQUE NOT NULL,
    "transfer_code" varchar UNIQUE,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

