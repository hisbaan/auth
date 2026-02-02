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
  column "created_at" {
    type = timestamptz
    null = false
  }
  column "updated_at" {
    type = timestamptz
    null = false
  }
  primary_key {
    columns = [column.id]
  }
  index "users_email_key" {
    unique  = true
    columns = [column.email]
  }
  index "users_username_key" {
    unique  = true
    columns = [column.username]
  }
}
