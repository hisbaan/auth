-- Normalize existing user emails and enforce uniqueness on normalized values.
UPDATE "users" SET "email" = lower(trim("email"));

DROP INDEX IF EXISTS "idx_users_email_key";
CREATE UNIQUE INDEX "idx_users_email_key" ON "users" (lower("email"));
