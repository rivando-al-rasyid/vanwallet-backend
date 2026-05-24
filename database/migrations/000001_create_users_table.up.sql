CREATE TABLE "users"(
    "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "email" varchar UNIQUE NOT NULL,
    "password" varchar NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    "updated_at" timestamp
);

