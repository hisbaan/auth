---
outline: deep
---

# Errors And Edge Cases

OIDC errors are returned either as redirects to the client's `redirect_uri` or as JSON responses from the token endpoint.

## Authorization Errors

When the request identifies a valid client and redirect URI, authorization errors are returned to the redirect URI with `error`, `error_description`, and `state` when supplied.

| Error | Common cause |
| --- | --- |
| `invalid_request` | Invalid PKCE parameters, unsupported `prompt`, or malformed request |
| `unsupported_response_type` | `response_type` is not `code` |
| `invalid_scope` | Missing `openid` or requesting a scope not allowed for the client |
| `access_denied` | User denied the request or email verification is required |
| `login_required` | `prompt=none` was used but the user is not signed in |
| `consent_required` | `prompt=none` was used but consent is required |

If the client ID is invalid or the redirect URI does not match the registered URI, the IdP returns an HTTP error instead of redirecting to an untrusted URI.

## Prompt Behavior

`prompt=none` requires the IdP to complete the request without UI. It cannot be combined with other prompt values.

`prompt=consent` forces the consent screen even if the user has already granted the requested scopes.

Only one prompt value is supported. Unsupported prompt values return `invalid_request`.

## Token Endpoint Errors

Token endpoint errors are JSON responses:

```json
{
  "error": "invalid_grant",
  "error_description": "Invalid authorization code"
}
```

| Error | HTTP status | Common cause |
| --- | --- | --- |
| `invalid_request` | `400` | Missing or malformed required parameter |
| `invalid_grant` | `400` | Expired, reused, invalid code, invalid verifier, or invalid refresh token |
| `invalid_client` | `401` | Client ID does not match the token or client is invalid |
| `unsupported_grant_type` | `400` | Unsupported `grant_type` |
| `server_error` | `500` | Unexpected server failure |

## Consent And Revocation

User consent is stored per user and client. Consent is sufficient only when the stored granted scopes include every requested scope.

If consent is revoked, existing access tokens can no longer use `/userinfo`, and refresh tokens tied to that authorization fail on refresh.

## Email Verification

Users must have a verified email address to authorize clients and exchange authorization codes. If a user's email becomes unverified, refresh and code exchange fail.
