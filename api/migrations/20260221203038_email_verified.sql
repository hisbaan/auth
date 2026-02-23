-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "email_verified" boolean NULL;
UPDATE "users" set "email_verified" = false;
ALTER TABLE "users" ALTER COLUMN "email_verified" SET NOT NULL;
