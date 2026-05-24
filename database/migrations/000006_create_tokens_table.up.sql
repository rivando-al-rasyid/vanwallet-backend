CREATE TABLE "tokens"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid NOT NULL,
    "token" varchar UNIQUE NOT NULL,
    "type" token_type NOT NULL,
    "expires_at" timestamp NOT NULL,
    "is_revoked" boolean NOT NULL DEFAULT FALSE,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

