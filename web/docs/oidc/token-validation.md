---
outline: deep
---

# Token Validation

Clients receive two validatable JWTs: an **ID token** that authenticates the user to your application, and an **access token** (RFC 9068 JWT) that authorizes API requests. Both are signed with the same issuer keys but carry different JOSE `typ` headers, so one can never be replayed as the other.

- Validate the **ID token** to sign the user in and read identity claims.
- Send the **access token** as a bearer credential to `/userinfo` - and, if your application has its own API, validate it there to authorize requests.

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

## ID Token

ID tokens are signed JWTs with the JOSE header `typ: JWT` and these claims:

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

### Validation checklist

For every ID token:

1. Verify the JWT signature with the matching JWKS key.
2. Require `alg` to be `EdDSA`.
3. Require `iss` to match the configured issuer exactly.
4. Require `aud` to contain your `client_id`.
5. Require `exp` to be in the future.
6. If you sent `nonce`, require the returned `nonce` to match.
7. Treat optional claims as absent unless their corresponding scopes were granted.

Use ID token claims for authentication and identity state. Do not accept ID tokens as API credentials: they have `typ: JWT`, not `typ: at+jwt`.

## Access Token

Access tokens are JWTs that follow the [RFC 9068](https://www.rfc-editor.org/rfc/rfc9068) profile, with the JOSE header `typ: at+jwt` and these claims:

| Claim       | Notes                                          |
| ----------- | ---------------------------------------------- |
| `iss`       | Must equal the issuer from discovery           |
| `sub`       | Stable user ID, such as `user_...`             |
| `aud`       | Your `client_id`                               |
| `client_id` | The client the token was issued to             |
| `scope`     | Space-separated scopes granted to the token    |
| `jti`       | Unique token identifier                        |
| `iat`       | Token issue time                               |
| `exp`       | Token expiration time                          |

These claims are a documented contract: your own backend may validate access tokens to authorize requests to your APIs, using the same JWKS keys.

### Validation checklist

For every access token your API accepts:

1. Verify the JWT signature with the matching JWKS key.
2. Require `alg` to be `EdDSA`.
3. Require the JOSE header `typ` to be `at+jwt`. This is what stops an ID token - or any other JWT from this issuer - from being used as an access token.
4. Require `iss` to match the configured issuer exactly.
5. Require `aud` to contain your `client_id`.
6. Require `exp` to be in the future.
7. Authorize the request based on the `scope` claim.

When calling this issuer's own APIs (such as `/userinfo`), you do not need to validate the access token yourself - send it as a bearer credential and the issuer validates it.
