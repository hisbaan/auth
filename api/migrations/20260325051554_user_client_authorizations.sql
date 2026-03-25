-- Create "user_client_authorizations" table
CREATE TABLE "user_client_authorizations" (
  "id" bytea NOT NULL,
  "user_id" bytea NOT NULL,
  "client_id" bytea NOT NULL,
  "granted_scopes" text[] NOT NULL,
  "last_authorized_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_client_authorizations_client_id" FOREIGN KEY ("client_id") REFERENCES "clients" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_user_client_authorizations_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_user_client_authorizations_client" to table: "user_client_authorizations"
CREATE INDEX "idx_user_client_authorizations_client" ON "user_client_authorizations" ("client_id");
-- Create index "idx_user_client_authorizations_user" to table: "user_client_authorizations"
CREATE INDEX "idx_user_client_authorizations_user" ON "user_client_authorizations" ("user_id");
-- Create index "idx_user_client_authorizations_user_client" to table: "user_client_authorizations"
CREATE UNIQUE INDEX "idx_user_client_authorizations_user_client" ON "user_client_authorizations" ("user_id", "client_id");
-- Modify "refresh_tokens" table
ALTER TABLE "refresh_tokens" ADD COLUMN "client_id" bytea NULL, ADD COLUMN "authorization_id" bytea NULL, ADD COLUMN "token_source" text NOT NULL DEFAULT 'self', ADD CONSTRAINT "fk_refresh_tokens_authorization_id" FOREIGN KEY ("authorization_id") REFERENCES "user_client_authorizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_refresh_tokens_client_id" FOREIGN KEY ("client_id") REFERENCES "clients" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "idx_refresh_tokens_authorization" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_authorization" ON "refresh_tokens" ("authorization_id");
-- Create index "idx_refresh_tokens_client" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_client" ON "refresh_tokens" ("client_id");
