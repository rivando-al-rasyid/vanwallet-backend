CREATE TABLE "wallets"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid NOT NULL,
    "label" varchar NOT NULL DEFAULT 'Wallet Utama',
    "balance" bigint NOT NULL DEFAULT 0,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

