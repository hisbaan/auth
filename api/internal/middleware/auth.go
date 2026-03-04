package middleware

import (
	"auth/internal/auth"
	"context"
	"crypto/ed25519"
	"net/http"
	"strings"
)

const AuthContextKey = "jwtClaims"

func Auth(publicKey ed25519.PublicKey, issuer string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				token = strings.TrimPrefix(authHeader, "Bearer ")
				if token == authHeader {
					http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
					return
				}
			} else {
				cookie, err := r.Cookie("access_token")
				if err != nil || cookie.Value == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				token = cookie.Value
			}

			_, claims, err := auth.ValidateToken[*auth.AccessClaims](publicKey, token, &auth.AccessClaims{})
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if err = auth.ValidateClaims(claims.RegisteredClaims, issuer); err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if claims.TokenType != "access" {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
