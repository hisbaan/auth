package oidc

import (
	"auth/internal/middleware"
	"crypto/ed25519"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Router(s *OIDCService) http.Handler {
	r := chi.NewRouter()
	jwtAccessKey := s.jwtSigningKey.Public().(ed25519.PublicKey)

	//	@Summary		OIDC authorization endpoint
	//	@Description	Starts the Authorization Code + PKCE flow. Redirects to the client redirect URI with a code (or error), or to the hosted login/consent UI
	//	@Tags			oidc
	//	@Param			response_type			query	string	true	"Must be 'code'"
	//	@Param			client_id				query	string	true	"Registered client ID"
	//	@Param			redirect_uri			query	string	true	"Must exactly match the registered redirect URI"
	//	@Param			scope					query	string	true	"Space-separated scopes; must include 'openid'"
	//	@Param			state					query	string	false	"Opaque value returned unchanged on the redirect"
	//	@Param			nonce					query	string	false	"Bound into the ID token issued for this request"
	//	@Param			code_challenge			query	string	true	"PKCE S256 code challenge"
	//	@Param			code_challenge_method	query	string	true	"Must be 'S256'"
	//	@Param			prompt					query	string	false	"'none' or 'consent'"
	//	@Param			login_hint				query	string	false	"Hint forwarded to the hosted login flow"
	//	@Success		302
	//	@Failure		400
	//	@Router			/authorize [get]
	r.Get("/authorize", s.handleAuthorize)

	//	@Summary		Deny an authorization request
	//	@Description	Used by the hosted consent UI. Redirects to the client redirect URI with error=access_denied
	//	@Tags			oidc
	//	@Param			request	query	string	true	"URL-encoded original authorize query string"
	//	@Success		302
	//	@Failure		400
	//	@Router			/authorize/deny [get]
	r.Get("/authorize/deny", s.handleAuthorizeDeny)

	//	@Summary		OIDC userinfo endpoint
	//	@Description	Returns claims for the user the access token was issued for, filtered by granted scopes
	//	@Tags			oidc
	//	@Security		BearerAuth
	//	@Success		200	{object}	UserInfoResponse
	//	@Failure		401
	//	@Router			/userinfo [get]
	r.Get("/userinfo", s.handleUserInfo)

	//	@Summary		OIDC userinfo endpoint
	//	@Description	Returns claims for the user the access token was issued for, filtered by granted scopes
	//	@Tags			oidc
	//	@Security		BearerAuth
	//	@Success		200	{object}	UserInfoResponse
	//	@Failure		401
	//	@Router			/userinfo [post]
	r.Post("/userinfo", s.handleUserInfo)

	//	@Summary		OIDC token endpoint
	//	@Description	Exchanges an authorization code (grant_type=authorization_code) or refresh token (grant_type=refresh_token) for tokens
	//	@Tags			oidc
	//	@Accept			x-www-form-urlencoded
	//	@Param			grant_type		formData	string	true	"'authorization_code' or 'refresh_token'"
	//	@Param			client_id		formData	string	true	"Registered client ID"
	//	@Param			code			formData	string	false	"Authorization code (authorization_code grant)"
	//	@Param			redirect_uri	formData	string	false	"Redirect URI used in the authorization request (authorization_code grant)"
	//	@Param			code_verifier	formData	string	false	"PKCE code verifier (authorization_code grant)"
	//	@Param			refresh_token	formData	string	false	"Refresh token (refresh_token grant)"
	//	@Success		200	{object}	TokenResponse
	//	@Failure		400
	//	@Failure		401
	//	@Router			/token [post]
	r.With(middleware.RateLimit(30, time.Minute)).Post("/token", s.handleToken)

	//	@Summary		Revoke a refresh token
	//	@Description	Revokes a client refresh token (RFC 7009). Unknown, expired, or non-client tokens are treated as successful no-ops
	//	@Tags			oidc
	//	@Accept			x-www-form-urlencoded
	//	@Param			token			formData	string	true	"Refresh token to revoke"
	//	@Param			token_type_hint	formData	string	false	"Optional hint; only 'refresh_token' is meaningful"
	//	@Success		204
	//	@Failure		400
	//	@Router			/token/revoke [post]
	r.With(middleware.RateLimit(30, time.Minute)).Post("/token/revoke", s.handleRevokeToken)

	// First-party endpoints for the hosted consent UI. session-authenticated
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtAccessKey, s.issuer))

		//	@Summary		Get client info for the consent screen
		//	@Description	Returns the display name, redirect URI, and allowed scopes of a client
		//	@Tags			oidc
		//	@Security		BearerAuth
		//	@Param			client_id	query	string	true	"Client ID"
		//	@Success		200	{object}	AuthorizeClientInfoResponse
		//	@Failure		400
		//	@Failure		401
		//	@Router			/authorize/client-info [get]
		r.Get("/authorize/client-info", s.handleClientInfo)

		//	@Summary		Grant consent to a client
		//	@Description	Records the signed-in user's consent for the requested scopes
		//	@Tags			oidc
		//	@Accept			json
		//	@Security		BearerAuth
		//	@Param			request	body	AuthorizeConsentParams	true	"Client and scopes to grant"
		//	@Success		204
		//	@Failure		400
		//	@Failure		401
		//	@Router			/authorize/consent [post]
		r.Post("/authorize/consent", s.handleConsent)
	})

	return r
}
