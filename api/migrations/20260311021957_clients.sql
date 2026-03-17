-- Create "clients" table
CREATE TABLE "clients" (
  "id" bytea NOT NULL,
  "user_id" bytea NOT NULL,
  "name" text NOT NULL,
  "redirect_uri" text NOT NULL,
  "allowed_scopes" text[] NOT NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_clients_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_clients_user" to table: "clients"
CREATE INDEX "idx_clients_user" ON "clients" ("user_id");
