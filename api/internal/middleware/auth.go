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
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}
			_, claims, err := auth.ValidateToken(publicKey, token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			if err = auth.ValidateClaims(claims, issuer); err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
