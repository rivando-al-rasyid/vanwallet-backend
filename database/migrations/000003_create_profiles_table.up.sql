CREATE TABLE "profiles"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "user_id" uuid UNIQUE NOT NULL,
    "full_name" varchar,
    "phone" varchar UNIQUE,
    "photo" varchar,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

