CREATE TYPE "token_type" AS ENUM(
    'REFRESH',
    'PASSWORD_RESET',
    'EMAIL_VERIFICATION'
);

CREATE TABLE "tokens"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid NOT NULL,
    "token" varchar UNIQUE NOT NULL,
    "type" token_type NOT NULL,
    "expires_at" timestamp NOT NULL,
    "is_revoked" boolean NOT NULL DEFAULT FALSE,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE INDEX idx_tokens_token ON tokens(token);

CREATE INDEX idx_tokens_user_id ON tokens(user_id);

CREATE INDEX idx_tokens_is_revoked ON tokens(is_revoked);

ALTER TABLE "tokens"
    ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE;

