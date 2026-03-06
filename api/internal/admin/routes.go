package admin

import (
	adminevents "auth/internal/admin/events"
	adminroles "auth/internal/admin/roles"
	adminusers "auth/internal/admin/users"
	"auth/internal/middleware"
	"crypto/ed25519"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminService, jwtAccessKey ed25519.PublicKey, issuer string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(jwtAccessKey, issuer))
	r.Use(middleware.RequireAdmin(issuer))

	r.Mount("/users", adminusers.Router(s.Users))
	r.Mount("/roles", adminroles.Router(s.Roles))
	r.Mount("/events", adminevents.Router(s.Events))

	return r
}
