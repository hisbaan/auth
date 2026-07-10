package middleware

import (
	"auth/internal/apperror"
	"auth/internal/sessions"
	"auth/internal/utils/httputil"
	"auth/internal/utils/jwtutil"
	"context"
	"crypto/ed25519"
	"net/http"
)

const AuthContextKey = "jwtClaims"

func AccessClaimsFromRequest(r *http.Request, publicKey ed25519.PublicKey, issuer string) (*sessions.AccessClaims, error) {
	var token string

	if r.Header.Get("Authorization") != "" {
		bearerToken, err := httputil.BearerToken(r)
		if err != nil {
			return nil, err
		}
		token = bearerToken
	} else {
		cookie, err := r.Cookie(sessions.AccessTokenCookieName)
		if err != nil || cookie.Value == "" {
			return nil, apperror.NewUnauthorized("Unauthorized")
		}
		token = cookie.Value
	}

	claims, err := jwtutil.ValidateToken(publicKey, issuer, jwtutil.SessionTokenJWTType, token, &sessions.AccessClaims{})
	if err != nil {
		return nil, apperror.NewUnauthorized("Unauthorized")
	}

	return claims, nil
}

func Auth(publicKey ed25519.PublicKey, issuer string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := AccessClaimsFromRequest(r, publicKey, issuer)
			if err != nil {
				serr, ok := err.(apperror.HTTPError)
				if !ok {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				http.Error(w, serr.Error(), serr.StatusCode())
				return
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
