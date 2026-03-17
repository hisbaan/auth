-- Create "events" table
CREATE TABLE "events" (
  "id" bytea NOT NULL,
  "user_id" bytea NULL,
  "type" text NOT NULL,
  "data" jsonb NOT NULL,
  "ip_address" inet NOT NULL,
  "user_agent" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_events_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_events_type" to table: "events"
CREATE INDEX "idx_events_type" ON "events" ("type");
-- Create index "idx_events_user" to table: "events"
CREATE INDEX "idx_events_user" ON "events" ("user_id");
