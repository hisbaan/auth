-- Create "email_verification_tokens" table
CREATE TABLE "email_verification_tokens" (
  "id" bytea NOT NULL,
  "user_id" bytea NOT NULL,
  "token_hash" bytea NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_email_verification_tokens_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_email_verification_tokens_user" to table: "email_verification_tokens"
CREATE INDEX "idx_email_verification_tokens_user" ON "email_verification_tokens" ("user_id");
