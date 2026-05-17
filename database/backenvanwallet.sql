CREATE TYPE "transaction_status" AS ENUM ('PENDING',
                                          'SUCCESS',
                                          'FAILED',
                                          'CANCELLED');


CREATE TYPE "payment_method" AS ENUM ('BRI',
                                      'BCA',
                                      'DANA',
                                      'GOPAY',
                                      'OVO');


CREATE TYPE "direction" AS ENUM ('IN',
                                 'OUT');


CREATE TABLE "users" ("id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()), "email" varchar UNIQUE NOT NULL,
                                                                                                "username" varchar UNIQUE NOT NULL,
                                                                                                                          "password" varchar NOT NULL,
                                                                                                                                             "created_at" timestamp NOT NULL DEFAULT (now()), "updated_at" timestamp);


CREATE TABLE "profiles" ("id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()), "user_id" uuid UNIQUE NOT NULL,
                                                                                                  "full_name" varchar, "phone" varchar UNIQUE,
                                                                                                                                       "photo" varchar, "created_at" timestamp NOT NULL DEFAULT (now()), "updated_at" timestamp);


CREATE TABLE "user_pins" ("id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()), "user_id" uuid UNIQUE NOT NULL,
                                                                                                   "pin_hash" varchar NOT NULL,
                                                                                                                      "failed_attempts" int NOT NULL DEFAULT 0,
                                                                                                                                                             "locked_until" timestamp, "created_at" timestamp NOT NULL DEFAULT (now()), "updated_at" timestamp);


CREATE TABLE "wallets" ("id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()), "user_id" uuid NOT NULL,
                                                                                          "label" varchar NOT NULL DEFAULT 'Wallet Utama',
                                                                                                                           "balance" bigint NOT NULL DEFAULT 0,
                                                                                                                                                             "created_at" timestamp NOT NULL DEFAULT (now()), "updated_at" timestamp);


CREATE TABLE "transactions" ("id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()), "wallet_id" uuid NOT NULL,
                                                                                                 "amount" bigint NOT NULL,
                                                                                                                 "direction" direction NOT NULL,
                                                                                                                                       "admin_fee" bigint NOT NULL DEFAULT 0,
                                                                                                                                                                           "status" transaction_status NOT NULL DEFAULT 'PENDING',
                                                                                                                                                                                                                        "note" text, "created_at" timestamp NOT NULL DEFAULT (now()));


CREATE TABLE "topups" ("transaction_id" uuid PRIMARY KEY NOT NULL,
                                                         "payment_method" payment_method,
                                                         "external_reference" varchar UNIQUE);


CREATE TABLE "transfers" ("transaction_id" uuid PRIMARY KEY NOT NULL,
                                                            "sender_transaction_id" uuid NOT NULL,
                                                                                         "recipient_wallet_id" uuid NOT NULL,
                                                                                                                    "transfer_code" varchar UNIQUE);


CREATE TABLE "expenses" ("transaction_id" uuid PRIMARY KEY NOT NULL,
                                                           "category" varchar, "merchant_name" varchar);


CREATE TABLE "withdrawals" ("transaction_id" uuid PRIMARY KEY NOT NULL,
                                                              "bank_name" varchar NOT NULL,
                                                                                  "account_number" varchar NOT NULL,
                                                                                                           "account_holder" varchar NOT NULL);

COMMENT ON TABLE "users" IS 'Core identity table. Credentials only — no PII stored here.';

COMMENT ON TABLE "profiles" IS 'One-to-one with users. Stores PII separately for easier data management.';

COMMENT ON TABLE "user_pins" IS 'PIN stored as bcrypt hash. locked_until enables brute-force protection — NULL means not locked.';

COMMENT ON TABLE "wallets" IS 'Supports multiple wallets per user. Balance stored in smallest currency unit (sen/IDR). label can be customized by user.';

COMMENT ON TABLE "transactions" IS 'Central ledger. direction is the single source of truth for IN/OUT. Each row is ONE side of a transaction. For transfers: two rows are created — one OUT for sender, one IN for recipient, linked via the transfers table.';

COMMENT ON TABLE "topups" IS 'external_reference: reference ID from payment gateway (DANA, GoPay, etc). unique constraint enables idempotency check.';

COMMENT ON TABLE "transfers" IS 'One row links BOTH sides of a transfer. transaction_id = OUT row (sender). sender_transaction_id = IN row (recipient). This enables full audit trail without ambiguity.';

COMMENT ON TABLE "expenses" IS 'Auto-created by system for every OUT transaction that is NOT a withdrawal. category & merchant_name can be enriched later.';

COMMENT ON TABLE "withdrawals" IS 'Manual bank withdrawal. bank_name, account_number, account_holder are snapshot values at time of withdrawal — not foreign keys to a bank account table.';


ALTER TABLE "profiles" ADD
FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "user_pins" ADD
FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "wallets" ADD
FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "transactions" ADD
FOREIGN KEY ("wallet_id") REFERENCES "wallets" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "topups" ADD
FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "transfers" ADD
FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "expenses" ADD
FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "withdrawals" ADD
FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "transfers" ADD
FOREIGN KEY ("sender_transaction_id") REFERENCES "transactions" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE "transfers" ADD
FOREIGN KEY ("recipient_wallet_id") REFERENCES "wallets" ("id") DEFERRABLE INITIALLY IMMEDIATE;