schema "public" {}

table "users" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "email" {
    type = text
    null = false
  }
  column "username" {
    type = text
    null = false
  }
  column "password_hash" {
    type = text
    null = false
  }
  column "email_verified" {
    type = boolean
    null = false
  }
  column "created_at" {
    type = timestamptz
    default = sql("now()")
    null = false
  }
  column "updated_at" {
    type = timestamptz
    default = sql("now()")
    null = false
  }

  primary_key {
    columns = [column.id]
  }
  index "idx_users_email_key" {
    unique  = true
    columns = [column.email]
  }
  index "idx_users_username_key" {
    unique  = true
    columns = [column.username]
  }
}

table "refresh_tokens" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = false
  }
  column "parent_id" {
    type = bytea
    null = true
  }
  column "issued_at" {
    type = timestamptz
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "revoked_at" {
    type = timestamptz
    null = true
  }
  column "ip_address" {
    type = inet
    null = false
  }
  column "user_agent" {
    type = text
    null = false
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_refresh_tokens_parent_id" {
    columns = [column.parent_id]
    ref_columns = [table.refresh_tokens.column.id]
    on_delete = "CASCADE"
  }
  foreign_key "fk_refresh_tokens_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = "CASCADE"
  }
  index "idx_refresh_tokens_user" {
    columns = [column.user_id]
  }
}

table "password_reset_tokens" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = false
  }
  column "token_hash" {
    type = bytea
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "revoked_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
  }
table "email_verification_tokens" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = false
  }
  column "token_hash" {
    type = bytea
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "revoked_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_email_verification_tokens_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "idx_email_verification_tokens_user" {
    columns = [column.user_id]
  }
}
