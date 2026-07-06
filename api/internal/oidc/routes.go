package oidc

import (
	"auth/internal/apperror"
	"auth/internal/middleware"
	sessiontokens "auth/internal/session_tokens"
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

func Router(s *OIDCService, jwtAccessKey ed25519.PublicKey, issuer string) http.Handler {
	r := chi.NewRouter()

	r.Get("/authorize", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		params := AuthorizeParams{
			ResponseType:        query.Get("response_type"),
			ClientID:            query.Get("client_id"),
			RedirectURI:         query.Get("redirect_uri"),
			RawQuery:            r.URL.RawQuery,
			CodeChallenge:       query.Get("code_challenge"),
			CodeChallengeMethod: query.Get("code_challenge_method"),
			Scope:               query.Get("scope"),
			State:               query.Get("state"),
			Prompt:              query.Get("prompt"),
			LoginHint:           query.Get("login_hint"),
		}
		nonce := query.Get("nonce")
		if nonce != "" {
			params.Nonce = &nonce
		}

		promptValue, _ := parseAuthorizePrompt(params.Prompt)

		var accessClaims *sessiontokens.AccessClaims
		if claims, claimsErr := middleware.AccessClaimsFromRequest(r, jwtAccessKey, issuer); claimsErr == nil {
			accessClaims = claims
		}
		var refreshToken string
		if cookie, cookieErr := r.Cookie(RefreshTokenCookieName); cookieErr == nil {
			refreshToken = cookie.Value
		}

		session, err := s.ResolveAuthorizeSession(r.Context(), accessClaims, refreshToken)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		if session != nil && session.RotatedTokens != nil {
			setSelfAuthCookies(w, s.cookieDomain, *session.RotatedTokens)
		}

		if session == nil {
			if promptValue == authorizePromptNone {
				_, result, validateErr := s.validateAuthorizeRequest(params)
				if validateErr != nil {
					httputil.HandleError(w, validateErr)
					return
				}
				if result.RedirectURI == "" {
					httputil.HandleError(w, apperror.NewBadRequest("Invalid authorize request"))
					return
				}
				if result.Query.Get("error") == "" {
					result.Query.Set("error", "login_required")
					result.Query.Set("error_description", "The user must sign in to continue")
				}

				redirectURL, err := httputil.WithQuery(result.RedirectURI, result.Query)
				if err != nil {
					httputil.HandleError(w, err)
					return
				}

				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			// No live session (access token invalid and refresh token unusable).
			redirectURL, err := httputil.WithQuery(strings.TrimRight(s.frontendURL, "/")+"/login", url.Values{
				"next": []string{strings.TrimRight(s.issuer, "/") + "/authorize?" + r.URL.RawQuery},
			})
			if err != nil {
				httputil.HandleError(w, err)
				return
			}

			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}

		response, err := s.Authorize(params, session.UserID)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		redirectURL, err := httputil.WithQuery(response.RedirectURI, response.Query)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtAccessKey, issuer))

		userinfoHandler := func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)

			response, err := s.UserInfo(UserInfoParams{Claims: claims})
			if err != nil {
				httputil.HandleError(w, err)
				return
			}

			httputil.JSONResponse(w, http.StatusOK, response)
		}

		r.Get("/userinfo", userinfoHandler)
		r.Post("/userinfo", userinfoHandler)

		r.Get("/authorize/client-info", func(w http.ResponseWriter, r *http.Request) {
			response, err := s.GetAuthorizeClientInfo(r.URL.Query().Get("client_id"))
			if err != nil {
				httputil.HandleError(w, err)
				return
			}

			httputil.JSONResponse(w, http.StatusOK, response)
		})

		r.Post("/authorize/consent", func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
			if claims.TokenSource != sessiontokens.TokenSourceSelf {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			userID, err := ulidutil.FromPrefixed("user", claims.Subject)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			var params AuthorizeConsentParams
			if err := httputil.ParseBody(w, r, &params); err != nil {
				return
			}

			if err := s.GrantConsent(userID, params); err != nil {
				httputil.HandleError(w, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})

	r.With(middleware.RateLimit(30, time.Minute)).Post("/token", func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.WithClientInfo(r.Context(), httputil.ClientInfoFromRequest(r))
		httputil.LimitBody(w, r)
		if err := r.ParseForm(); err != nil {
			HandleTokenError(w, NewInvalidRequestTokenError("Invalid request"))
			return
		}

		switch r.Form.Get("grant_type") {
		case "authorization_code":
			response, err := s.TokenAuthorizationCode(ctx, TokenAuthorizationCodeParams{
				GrantType:    r.Form.Get("grant_type"),
				ClientID:     r.Form.Get("client_id"),
				Code:         r.Form.Get("code"),
				RedirectURI:  r.Form.Get("redirect_uri"),
				CodeVerifier: r.Form.Get("code_verifier"),
			})
			if err != nil {
				HandleTokenError(w, err)
				return
			}

			httputil.JSONResponse(w, http.StatusOK, response)
		case "refresh_token":
			response, err := s.TokenRefreshToken(ctx, TokenRefreshTokenParams{
				GrantType:    r.Form.Get("grant_type"),
				ClientID:     r.Form.Get("client_id"),
				RefreshToken: r.Form.Get("refresh_token"),
			})
			if err != nil {
				HandleTokenError(w, err)
				return
			}

			httputil.JSONResponse(w, http.StatusOK, response)
		default:
			HandleTokenError(w, NewUnsupportedGrantTypeTokenError("Invalid grant type"))
		}
	})

	r.With(middleware.RateLimit(30, time.Minute)).Post("/token/revoke", func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.WithClientInfo(r.Context(), httputil.ClientInfoFromRequest(r))
		httputil.LimitBody(w, r)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := s.RevokeToken(ctx, RevokeTokenParams{
			Token:         r.Form.Get("token"),
			TokenTypeHint: r.Form.Get("token_type_hint"),
		}); err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/authorize/deny", func(w http.ResponseWriter, r *http.Request) {
		request := r.URL.Query().Get("request")
		params, err := authorizeParamsFromRawQuery(request)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		response, err := s.DenyAuthorize(params)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		redirectURL, err := httputil.WithQuery(response.RedirectURI, response.Query)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	return r
}

func setSelfAuthCookies(w http.ResponseWriter, cookieDomain string, tokens sessiontokens.SessionTokens) {
	secure := cookieDomain != "localhost"

	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    tokens.AccessToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Domain:   cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Domain:   cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(168) * time.Hour),
	})
}

func authorizeParamsFromRawQuery(rawQuery string) (AuthorizeParams, error) {
	if rawQuery == "" {
		return AuthorizeParams{}, apperror.NewBadRequest("Missing authorize request")
	}
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return AuthorizeParams{}, apperror.NewBadRequest("Invalid authorize request")
	}

	params := AuthorizeParams{
		ResponseType:        query.Get("response_type"),
		ClientID:            query.Get("client_id"),
		RedirectURI:         query.Get("redirect_uri"),
		RawQuery:            rawQuery,
		CodeChallenge:       query.Get("code_challenge"),
		CodeChallengeMethod: query.Get("code_challenge_method"),
		Scope:               query.Get("scope"),
		State:               query.Get("state"),
		Prompt:              query.Get("prompt"),
		LoginHint:           query.Get("login_hint"),
	}
	if nonce := query.Get("nonce"); nonce != "" {
		params.Nonce = &nonce
	}

	return params, nil
}
