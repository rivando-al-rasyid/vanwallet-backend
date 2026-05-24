CREATE TABLE "transactions"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "wallet_id" uuid NOT NULL,
    "type" transaction_type NOT NULL,
    "amount" bigint NOT NULL,
    "admin_fee" bigint NOT NULL DEFAULT 0,
    "status" transaction_status NOT NULL DEFAULT 'PENDING',
    "idempotency_key" varchar UNIQUE,
    "note" text,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

