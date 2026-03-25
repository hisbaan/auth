-- Create "authorization_codes" table
CREATE TABLE "authorization_codes" (
  "id" bytea NOT NULL,
  "code_hash" bytea NOT NULL,
  "user_id" bytea NOT NULL,
  "client_id" bytea NOT NULL,
  "redirect_uri" text NOT NULL,
  "scopes" text[] NOT NULL,
  "code_challenge" text NOT NULL,
  "code_challenge_method" text NOT NULL,
  "nonce" text NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_authorization_codes_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_authorization_codes_client_id" FOREIGN KEY ("client_id") REFERENCES "clients" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_authorization_codes_code_hash" to table: "authorization_codes"
CREATE UNIQUE INDEX "idx_authorization_codes_code_hash" ON "authorization_codes" ("code_hash");
-- Create index "idx_authorization_codes_user" to table: "authorization_codes"
CREATE INDEX "idx_authorization_codes_user" ON "authorization_codes" ("user_id");
-- Create index "idx_authorization_codes_client" to table: "authorization_codes"
CREATE INDEX "idx_authorization_codes_client" ON "authorization_codes" ("client_id");
