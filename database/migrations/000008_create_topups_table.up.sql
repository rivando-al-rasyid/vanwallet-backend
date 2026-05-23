CREATE TABLE "topups"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "wallet_id" uuid NOT NULL,
    "amount" bigint NOT NULL,
    "status" transaction_status NOT NULL DEFAULT 'PENDING',
    "payment_method" payment_method,
    "payment_metadata" jsonb,
    "external_reference" varchar UNIQUE,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

