-- Add OIDC continuation support to email verification tokens.
ALTER TABLE "email_verification_tokens" ADD COLUMN "return_to" text NULL;
