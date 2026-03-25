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
		claims, err := middleware.AccessClaimsFromRequest(r, jwtAccessKey, issuer)
		if err != nil {
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

			redirectURL, err := httputil.WithQuery(strings.TrimRight(s.frontendURL, "/")+"/authorize", url.Values{
				"request": []string{r.URL.RawQuery},
			})
			if err != nil {
				httputil.HandleError(w, err)
				return
			}

			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}

		userID, err := ulidutil.FromPrefixed("user", claims.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		response, err := s.Authorize(params, userID)
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

		r.Get("/userinfo", func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)

			response, err := s.UserInfo(UserInfoParams{Claims: claims})
			if err != nil {
				httputil.HandleError(w, err)
				return
			}

			httputil.JSONResponse(w, http.StatusOK, response)
		})

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

	r.Post("/token", func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.WithClientInfo(r.Context(), httputil.ClientInfoFromRequest(r))
		if err := r.ParseForm(); err != nil {
			HandleTokenError(w, NewInvalidRequestTokenError("Invalid request"))
			return
		}

		switch r.Form.Get("grant_type") {
		case "authorization_code":
			response, err := s.TokenAuthorizationCode(ctx, TokenAuthorizationCodeParams{
				GrantType:    r.Form.Get("grant_type"),
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

	r.Post("/token/revoke", func(w http.ResponseWriter, r *http.Request) {
		ctx := httputil.WithClientInfo(r.Context(), httputil.ClientInfoFromRequest(r))
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
