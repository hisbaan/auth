package users

import (
	"auth/internal/httputil"
	"auth/internal/middleware"
	"crypto/ed25519"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

func Router(s *UsersService) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(s.jwtAccessKey.Public().(ed25519.PublicKey), s.issuer))

	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*jwt.RegisteredClaims)
		userID, err := ulid.Parse(ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		response, err := s.GetUser(userID)
		if err != nil {
			httputil.HandleErrors(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	r.Put("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*jwt.RegisteredClaims)
		userID, err := ulid.Parse(ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		var body UpdateUserParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err = s.UpdateUser(userID, body)
		if err != nil {
			httputil.HandleErrors(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/me/password", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*jwt.RegisteredClaims)
		userID, err := ulid.Parse(ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		var body UpdatePasswordParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err = s.UpdatePassword(userID, body)
		if err != nil {
			httputil.HandleErrors(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Delete("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*jwt.RegisteredClaims)
		userID, err := ulid.Parse(ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		err = s.DeleteUser(userID)
		if err != nil {
			httputil.HandleErrors(w, err)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
