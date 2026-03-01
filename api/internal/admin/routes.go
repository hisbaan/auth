package admin

import (
	"auth/internal/middleware"
	"auth/internal/utils/httputil"
	"crypto/ed25519"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"
)

func Router(s *AdminService, jwtAccessKey ed25519.PublicKey, issuer string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(jwtAccessKey, issuer))
	r.Use(middleware.RequireAdmin(issuer))

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		params := ListUsersParams{
			Limit:  20,
			Cursor: r.URL.Query().Get("cursor"),
		}
		if limit := r.URL.Query().Get("limit"); limit != "" {
			if n, err := strconv.Atoi(limit); err == nil {
				params.Limit = n
			}
		}

		response, err := s.ListUsers(params)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := ulid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		response, err := s.GetUser(id)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	r.Post("/users/roles", func(w http.ResponseWriter, r *http.Request) {
		var body UpdateUserRoleParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.AddUserRole(body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	r.Delete("/users/roles", func(w http.ResponseWriter, r *http.Request) {
		var body UpdateUserRoleParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.RemoveUserRole(body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
