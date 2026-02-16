-- Create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" (
  "id" bytea NOT NULL,
  "user_id" bytea NOT NULL,
  "token_hash" bytea NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL
);
