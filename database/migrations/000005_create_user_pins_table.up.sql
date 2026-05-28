CREATE TABLE "user_pins"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid UNIQUE NOT NULL,
    "pin_hash" varchar NOT NULL,
    "failed_attempts" int NOT NULL DEFAULT 0,
    "locked_until" timestamp,
    "last_used_at" timestamp,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

