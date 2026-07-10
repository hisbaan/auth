---
outline: deep
---

# Authorization Code Flow

Clients use Authorization Code flow with PKCE. PKCE is required and only `S256` code challenges are accepted.

## 1. Create PKCE Values

Generate a high-entropy `code_verifier`, then derive the `code_challenge`:

```text
code_challenge = BASE64URL-ENCODE(SHA256(code_verifier))
code_challenge_method = S256
```

Store the `code_verifier` until the token exchange.

## 2. Redirect To Authorization

Send the user to `/authorize` with these query parameters:

| Parameter               | Required    | Notes                                          |
| ----------------------- | ----------- | ---------------------------------------------- |
| `response_type`         | Yes         | Must be `code`                                 |
| `client_id`             | Yes         | Registered client ID, such as `client_...`     |
| `redirect_uri`          | Yes         | Must exactly match the registered redirect URI |
| `scope`                 | Yes         | Space-separated scopes; must include `openid`  |
| `state`                 | Recommended | Returned unchanged to your redirect URI        |
| `nonce`                 | Recommended | Included in the ID token after code exchange   |
| `code_challenge`        | Yes         | S256 challenge derived from the verifier       |
| `code_challenge_method` | Yes         | Must be `S256`                                 |
| `prompt`                | Optional    | Supports `none` or `consent`                   |
| `login_hint`            | Optional    | Passed through to the hosted login flow        |

Example:

```http
GET /authorize?
    response_type=code&
    client_id=client_01H...&
    redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback&scope=openid%20profile%20email&
    state=opaque-state&
    nonce=opaque-nonce&
    code_challenge=...&
    code_challenge_method=S256
```

## 3. Handle Login And Consent

If the user is not signed in, the IdP redirects them to the hosted authorization UI.

If the user has not granted the requested scopes, the hosted UI asks for consent. Existing consent is reused when it covers every requested scope.

Use `prompt=consent` to force a fresh consent screen even if consent already exists.

Use `prompt=none` only when your client can handle an immediate redirect error. If the user must sign in, the IdP redirects back with `login_required`. If consent is required, it redirects back with `consent_required`.

## 4. Receive The Redirect

After successful authorization, the user is redirected to your registered `redirect_uri` with:

| Parameter | Notes                                            |
| --------- | ------------------------------------------------ |
| `code`    | Authorization code for the token exchange        |
| `state`   | Present if supplied in the authorization request |
| `iss`     | The issuer identifier (RFC 9207)                 |

Example:

```text
https://app.example.com/callback?code=...&state=opaque-state&iss=https%3A%2F%2Fauth.example.com
```

Verify that `state` matches the value your client generated before exchanging the code. If your OIDC library supports RFC 9207, also verify that `iss` matches the configured issuer; it is included on every authorization response, including errors.

Authorization codes expire after 15 minutes and can be used only once.

## 5. Exchange The Code

Exchange the code at `/token` using the original `code_verifier`. See [Tokens and UserInfo](./tokens-and-userinfo.md).
