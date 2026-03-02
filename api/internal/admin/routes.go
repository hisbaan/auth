package admin

import (
	"auth/internal/middleware"
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminService, jwtAccessKey ed25519.PublicKey, issuer string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(jwtAccessKey, issuer))
	r.Use(middleware.RequireAdmin(issuer))

	//	@Summary		List users
	//	@Description	Returns a paginated list of users (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			cursor	query		string	false	"Pagination cursor"
	//	@Param			limit	query		int		false	"Maximum number of results (max 100)"
	//	@Success		200		{object}	ListUsersResponse
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/users [get]
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

	//	@Summary		Get user by ID
	//	@Description	Returns a user by their ID (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			id	path		string	true	"User ID"
	//	@Success		200	{object}	GetUserResponse
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Failure		404
	//	@Router			/admin/users/{id} [get]
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := ulidutil.FromPrefixed("user", chi.URLParam(r, "id"))
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

	//	@Summary		Add user role
	//	@Description	Assigns a role to a user (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			request	body	UpdateUserRoleParams	true	"User role assignment"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/users/roles [post]
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

	//	@Summary		Remove user role
	//	@Description	Removes a role from a user (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			request	body	UpdateUserRoleParams	true	"User role removal"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/users/roles [delete]
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
