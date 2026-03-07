package middleware

import (
	"auth/internal/utils/httputil"
	"net/http"
)

func ClientInfo() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := httputil.ClientInfoFromRequest(r)
			ctx := httputil.WithClientInfo(r.Context(), info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
