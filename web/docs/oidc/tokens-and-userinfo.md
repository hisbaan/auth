---
outline: deep
---

# Tokens And UserInfo

The token endpoint accepts `application/x-www-form-urlencoded` requests. Clients are public clients and do not authenticate with a secret.

## Authorization Code Exchange

```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&client_id=client_01H...&code=...&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback&code_verifier=...
```

Required fields:

| Field           | Notes                                               |
| --------------- | --------------------------------------------------- |
| `grant_type`    | Must be `authorization_code`                        |
| `client_id`     | Registered public client ID                         |
| `code`          | Authorization code from the callback                |
| `redirect_uri`  | Same redirect URI used in the authorization request |
| `code_verifier` | Original PKCE verifier                              |

Successful response:

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "...",
  "id_token": "..."
}
```

The exact `expires_in` value is controlled by the server session-token configuration.

## Refresh Token Grant

Use the refresh token grant to obtain a new access token after the access token expires.

```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&client_id=client_01H...&refresh_token=...
```

Successful response:

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "...",
  "id_token": "..."
}
```

Refresh tokens rotate. Store the new `refresh_token` from each successful refresh and discard the old one.

The refresh-token response includes a new `id_token`. Validate it before trusting its identity claims. Refresh-issued ID tokens do not include a `nonce` because no new authorization request was made.

## UserInfo

Call `/userinfo` with a bearer access token issued to your client:

```http
GET /userinfo
Authorization: Bearer eyJ...
```

`POST /userinfo` is also supported.

Successful response with all standard scopes granted:

```json
{
  "sub": "user_01H...",
  "preferred_username": "alice",
  "email": "alice@example.com",
  "email_verified": true
}
```

Claims are scope-dependent:

| Scope     | UserInfo fields           |
| --------- | ------------------------- |
| `openid`  | `sub`                     |
| `profile` | `preferred_username`      |
| `email`   | `email`, `email_verified` |

The access token must belong to an active client authorization. If the user revokes consent or the client is revoked, UserInfo returns an authorization error.

## Revocation

Revoke a refresh token with `/token/revoke`:

```http
POST /token/revoke
Content-Type: application/x-www-form-urlencoded

token=...&token_type_hint=refresh_token
```

`token_type_hint` is optional. Unknown, expired, or already revoked tokens are treated as successful no-ops.
