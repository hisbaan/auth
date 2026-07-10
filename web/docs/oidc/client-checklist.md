---
outline: deep
---

# Client Checklist

Use this checklist when integrating a client application.

## Configuration

- Read provider metadata from `/.well-known/openid-configuration` when your OIDC library supports discovery.
- Configure your registered `client_id`.
- Configure the exact registered `redirect_uri`.
- Do not configure a client secret.
- Enable Authorization Code flow with PKCE.
- Require `S256` PKCE challenges.
- For browser-based clients, ensure your app origin matches a registered `redirect_uri`.

## Authorization Request

- Generate a unique `state` for every authorization request.
- Generate a unique `nonce` for every authorization request that expects an ID token.
- Request `response_type=code`.
- Include `openid` in every scope request.
- Request only scopes allowed for the registered client.
- Store the PKCE `code_verifier` until the token exchange.

## Callback Handling

- Reject callbacks with a missing or mismatched `state`.
- Verify the `iss` response parameter matches the configured issuer when your library supports RFC 9207.
- Handle authorization errors returned to the redirect URI.
- Exchange each authorization code only once.
- Send the same `redirect_uri` used in the authorization request.
- Send the original PKCE `code_verifier`.

## Token Handling

- Validate ID token signatures using JWKS.
- Validate `iss`, `aud`, `exp`, and `nonce`.
- Use ID token claims for client-side authentication and identity state, not API authorization.
- Store refresh tokens securely.
- Replace stored refresh tokens after every successful refresh.
- Send access tokens as bearer credentials to this issuer's APIs.
- To protect your own API, validate access tokens per [Token Validation](./token-validation.md#access-token): require the `at+jwt` JOSE `typ`, then check `iss`, `aud`, `exp`, and `scope`.
- Never accept an ID token where an access token is expected, or vice versa - the `typ` header check enforces this.

## UserInfo

- Send the access token as an `Authorization: Bearer` token.
- Expect claims to vary by granted scope.
- Handle authorization errors if consent or the client has been revoked.

## Operations

- Use `/token/revoke` when disconnecting a user or rotating stored refresh tokens out of use.
- Expect unknown or already revoked refresh tokens to revoke successfully.
- Monitor token endpoint errors and surface reauthorization prompts when refresh fails.
