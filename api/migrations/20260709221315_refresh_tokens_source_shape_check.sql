-- Modify "refresh_tokens" table
ALTER TABLE "refresh_tokens" ADD CONSTRAINT "chk_refresh_tokens_source_shape" CHECK (((token_source = 'client'::text) AND (client_id IS NOT NULL) AND (authorization_id IS NOT NULL)) OR ((token_source = 'self'::text) AND (client_id IS NULL) AND (authorization_id IS NULL)));
