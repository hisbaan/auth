package admin

import (
	"crypto/ed25519"
	"net/http"

	adminevents "auth/internal/admin/events"
	adminrefreshtokens "auth/internal/admin/refresh_tokens"
	adminroles "auth/internal/admin/roles"
	adminusers "auth/internal/admin/users"
	"auth/internal/middleware"
	sessiontokens "auth/internal/session_tokens"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminService, jwtAccessKey ed25519.PublicKey, issuer string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(jwtAccessKey, issuer))
	r.Use(middleware.RequireTokenSource(sessiontokens.TokenSourceSelf))
	r.Use(middleware.RequireAdmin(issuer))

	r.Mount("/users", adminusers.Router(s.Users))
	r.Mount("/roles", adminroles.Router(s.Roles))
	r.Mount("/events", adminevents.Router(s.Events))
	r.Mount("/refresh-tokens", adminrefreshtokens.Router(s.RefreshTokens))

	return r
}
