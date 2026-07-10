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

table "clients" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "redirect_uri" {
    type = text
    null = false
  }
  column "allowed_scopes" {
    type = sql("text[]")
    null = false
  }
  column "revoked_at" {
    type = timestamptz
    null = true
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
  foreign_key "fk_clients_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = "CASCADE"
  }
  index "idx_clients_user" {
    columns = [column.user_id]
  }
}

table "roles" {
  schema = schema.public

  column "name" {
    type = text
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
    columns = [column.name]
  }
}

table "user_roles" {
  schema = schema.public

  column "user_id" {
    type = bytea
    null = false
  }
  column "role" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.user_id, column.role]
  }
  foreign_key "fk_user_roles_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = "CASCADE"
  }
  foreign_key "fk_user_roles_role" {
    columns = [column.role]
    ref_columns = [table.roles.column.name]
    on_delete = "CASCADE"
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
  column "client_id" {
    type = bytea
    null = true
  }
  column "authorization_id" {
    type = bytea
    null = true
  }
  column "token_source" {
    type = text
    null = false
    default = sql("'self'")
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
  foreign_key "fk_refresh_tokens_client_id" {
    columns = [column.client_id]
    ref_columns = [table.clients.column.id]
    on_delete = "CASCADE"
  }
  foreign_key "fk_refresh_tokens_authorization_id" {
    columns = [column.authorization_id]
    ref_columns = [table.user_client_authorizations.column.id]
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
  index "idx_refresh_tokens_client" {
    columns = [column.client_id]
  }
  index "idx_refresh_tokens_authorization" {
    columns = [column.authorization_id]
  }
  check "chk_refresh_tokens_source_shape" {
    expr = "(token_source = 'client' AND client_id IS NOT NULL AND authorization_id IS NOT NULL) OR (token_source = 'self' AND client_id IS NULL AND authorization_id IS NULL)"
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
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_password_reset_tokens_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "idx_password_reset_tokens_user" {
    columns = [column.user_id]
  }
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
  column "email" {
    type = text
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

  column "return_to" {
    type = text
    null = true
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

table "authorization_codes" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "code_hash" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = false
  }
  column "client_id" {
    type = bytea
    null = false
  }
  column "redirect_uri" {
    type = text
    null = false
  }
  column "scopes" {
    type = sql("text[]")
    null = false
  }
  column "code_challenge" {
    type = text
    null = false
  }
  column "code_challenge_method" {
    type = text
    null = false
  }
  column "nonce" {
    type = text
    null = true
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "used_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_authorization_codes_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  foreign_key "fk_authorization_codes_client_id" {
    columns = [column.client_id]
    ref_columns = [table.clients.column.id]
    on_delete = CASCADE
  }
  index "idx_authorization_codes_code_hash" {
    unique  = true
    columns = [column.code_hash]
  }
  index "idx_authorization_codes_user" {
    columns = [column.user_id]
  }
  index "idx_authorization_codes_client" {
    columns = [column.client_id]
  }
}

table "user_client_authorizations" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = false
  }
  column "client_id" {
    type = bytea
    null = false
  }
  column "granted_scopes" {
    type = sql("text[]")
    null = false
  }
  column "last_authorized_at" {
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
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_user_client_authorizations_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  foreign_key "fk_user_client_authorizations_client_id" {
    columns = [column.client_id]
    ref_columns = [table.clients.column.id]
    on_delete = CASCADE
  }
  index "idx_user_client_authorizations_user_client" {
    unique  = true
    columns = [column.user_id, column.client_id]
  }
  index "idx_user_client_authorizations_user" {
    columns = [column.user_id]
  }
  index "idx_user_client_authorizations_client" {
    columns = [column.client_id]
  }
}

table "events" {
  schema = schema.public

  column "id" {
    type = bytea
    null = false
  }
  column "user_id" {
    type = bytea
    null = true
  }
  column "type" {
    type = text
    null = false
  }
  column "data" {
    type = jsonb
    null = false
  }
  column "ip_address" {
    type = inet
    null = false
  }
  column "user_agent" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_events_user_id" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
  }
  index "idx_events_user" {
    columns = [column.user_id]
  }
  index "idx_events_type" {
    columns = [column.type]
  }
}
