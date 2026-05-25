---
outline: deep
---

# Token Validation

Clients should validate ID tokens before trusting their identity claims. Use access tokens as bearer credentials for `/userinfo` and API calls; do not use ID tokens to authorize API requests.

## JWKS

Fetch public signing keys from:

```http
GET /.well-known/jwks.json
```

The key set contains Ed25519 public keys:

```json
{
  "keys": [
    {
      "kty": "OKP",
      "use": "sig",
      "kid": "...",
      "alg": "EdDSA",
      "crv": "Ed25519",
      "x": "..."
    }
  ]
}
```

Use the `kid` in the token header to select the key.

## ID Token Claims

ID tokens are signed JWTs with these claims:

| Claim                | Notes                                              |
| -------------------- | -------------------------------------------------- |
| `iss`                | Must equal the issuer from discovery               |
| `sub`                | Stable user ID, such as `user_...`                 |
| `aud`                | Your `client_id`                                   |
| `iat`                | Token issue time                                   |
| `exp`                | Token expiration time                              |
| `nonce`              | Present when supplied in the authorization request |
| `preferred_username` | Present when `profile` was granted                 |
| `email`              | Present when `email` was granted                   |
| `email_verified`     | Present when `email` was granted                   |

## Validation Checklist

For every ID token:

1. Verify the JWT signature with the matching JWKS key.
2. Require `alg` to be `EdDSA`.
3. Require `iss` to match the configured issuer exactly.
4. Require `aud` to contain your `client_id`.
5. Require `exp` to be in the future.
6. If you sent `nonce`, require the returned `nonce` to match.
7. Treat optional claims as absent unless their corresponding scopes were granted.

## Access Tokens

Access tokens are bearer credentials for the auth service. Send them to `/userinfo` or other APIs that accept this issuer's access tokens.

Client applications should not parse access tokens or depend on their internal claims unless that token format is explicitly documented for your integration. Even when an access token looks like a JWT, treat its structure as an implementation detail unless the API contract says otherwise.

ID tokens are the tokens clients validate and read for authentication and identity claims. Access tokens are the tokens clients present to APIs.
