-- Remove username — identity is now based on email only.
-- full_name and phone in profiles are used for receiver search.
ALTER TABLE "users" DROP COLUMN IF EXISTS "username";
