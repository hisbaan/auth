---
outline: deep
---

# Token Validation

Clients should validate ID tokens before trusting their claims.

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

Access tokens are bearer tokens for the auth service. Client applications normally use them with `/userinfo` and should not rely on their internal format unless explicitly supported by your integration.
