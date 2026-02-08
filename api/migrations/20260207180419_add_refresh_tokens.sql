-- Modify "users" table
ALTER TABLE "users" ALTER COLUMN "created_at" SET DEFAULT now(), ALTER COLUMN "updated_at" SET DEFAULT now();
-- Rename an index from "users_email_key" to "idx_users_email_key"
ALTER INDEX "users_email_key" RENAME TO "idx_users_email_key";
-- Rename an index from "users_username_key" to "idx_users_username_key"
ALTER INDEX "users_username_key" RENAME TO "idx_users_username_key";
-- Create "refresh_tokens" table
CREATE TABLE "refresh_tokens" (
  "id" bytea NOT NULL,
  "user_id" bytea NOT NULL,
  "parent_id" bytea NULL,
  "issued_at" timestamptz NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "ip_address" inet NOT NULL,
  "user_agent" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_refresh_tokens_parent_id" FOREIGN KEY ("parent_id") REFERENCES "refresh_tokens" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_refresh_tokens_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_refresh_tokens_user" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_user" ON "refresh_tokens" ("user_id");
