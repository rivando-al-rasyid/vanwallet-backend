CREATE TABLE "favorites"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid NOT NULL,
    "target_user_id" uuid NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

