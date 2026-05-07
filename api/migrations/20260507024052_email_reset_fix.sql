DELETE FROM "email_verification_tokens";
-- Modify "email_verification_tokens" table
ALTER TABLE "email_verification_tokens" ADD COLUMN "email" text NOT NULL;
-- Drop index "idx_users_email_key" from table: "users"
DROP INDEX "idx_users_email_key";
-- Create index "idx_users_email_key" to table: "users"
CREATE UNIQUE INDEX "idx_users_email_key" ON "users" ("email");
