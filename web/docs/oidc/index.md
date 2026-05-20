---
outline: deep
---

# OpenID Connect

`auth` acts as an OpenID Connect identity provider for applications that need to sign users in with their `auth` account.

These docs are for maintainers of client applications. They describe the public OIDC endpoints, supported flows, token formats, and integration requirements.

## Discovery

Use the OpenID Provider metadata document to configure your client when your OIDC library supports discovery:

```http
GET /.well-known/openid-configuration
```

The discovery document exposes the authorization, token, revocation, userinfo, and JWKS endpoints for the current issuer.

## Client Registration

Each client has a registered `client_id`, display name, redirect URI, and allowed scopes.

The redirect URI must match exactly. The authorization request and token request both fail if the `redirect_uri` differs from the registered value or from the URI used to create the authorization code.

Clients are public clients. Do not send a client secret to the token endpoint.

## Scopes

Every authorization request must include `openid`.

| Scope | Claims enabled |
| --- | --- |
| `openid` | `sub` |
| `profile` | `preferred_username` |
| `email` | `email`, `email_verified` |

Users must have a verified email address before they can authorize a client or exchange an authorization code.

## Endpoint Summary

| Endpoint | Method | Purpose |
| --- | --- | --- |
| `/.well-known/openid-configuration` | `GET` | Provider metadata |
| `/.well-known/jwks.json` | `GET` | Public signing keys |
| `/authorize` | `GET` | Start the authorization request |
| `/token` | `POST` | Exchange an authorization code or refresh token |
| `/userinfo` | `GET`, `POST` | Fetch claims for an access token |
| `/token/revoke` | `POST` | Revoke a refresh token |

## Integration Pages

- [Authorization Code Flow](./authorization-code-flow.md)
- [Tokens and UserInfo](./tokens-and-userinfo.md)
- [Token Validation](./token-validation.md)
- [Errors and Edge Cases](./errors.md)
- [Client Checklist](./client-checklist.md)
