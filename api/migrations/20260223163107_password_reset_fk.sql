-- Modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" ADD PRIMARY KEY ("id"), ADD CONSTRAINT "fk_password_reset_tokens_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "idx_password_reset_tokens_user" to table: "password_reset_tokens"
CREATE INDEX "idx_password_reset_tokens_user" ON "password_reset_tokens" ("user_id");
