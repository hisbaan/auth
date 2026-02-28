-- Create "roles" table
CREATE TABLE "roles" (
  "name" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("name")
);
-- Create "user_roles" table
CREATE TABLE "user_roles" (
  "user_id" bytea NOT NULL,
  "role" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id", "role"),
  CONSTRAINT "fk_user_roles_role" FOREIGN KEY ("role") REFERENCES "roles" ("name") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_user_roles_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
