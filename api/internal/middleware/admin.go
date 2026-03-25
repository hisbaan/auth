package middleware

import (
	"net/http"
	"slices"

	sessiontokens "auth/internal/session_tokens"
)

func RequireAdmin(issuer string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(AuthContextKey).(*sessiontokens.AccessClaims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if len(claims.Roles) == 0 {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !slices.Contains(claims.Roles, "admin") {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
