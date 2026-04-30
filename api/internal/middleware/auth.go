package middleware

import (
	"auth/internal/apperror"
	sessiontokens "auth/internal/session_tokens"
	"auth/internal/utils/jwtutil"
	"context"
	"crypto/ed25519"
	"net/http"
	"strings"
)

const AuthContextKey = "jwtClaims"

func AccessClaimsFromRequest(r *http.Request, publicKey ed25519.PublicKey, issuer string) (*sessiontokens.AccessClaims, error) {
	var token string

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token = strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return nil, apperror.NewUnauthorized("Unauthorized")
		}
	} else {
		cookie, err := r.Cookie("access_token")
		if err != nil || cookie.Value == "" {
			return nil, apperror.NewUnauthorized("Unauthorized")
		}
		token = cookie.Value
	}

	_, claims, err := jwtutil.ValidateToken(publicKey, token, &sessiontokens.AccessClaims{})
	if err != nil {
		return nil, apperror.NewUnauthorized("Unauthorized")
	}
	if err = jwtutil.ValidateClaims(claims.RegisteredClaims, issuer); err != nil {
		return nil, apperror.NewUnauthorized("Unauthorized")
	}
	if claims.TokenType != "access" {
		return nil, apperror.NewUnauthorized("Unauthorized")
	}

	return claims, nil
}

func Auth(publicKey ed25519.PublicKey, issuer string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authHeader := r.Header.Get("Authorization"); authHeader != "" && !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

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

func RequireTokenSource(source string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(AuthContextKey).(*sessiontokens.AccessClaims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.TokenSource != source {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
