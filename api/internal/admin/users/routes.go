package users

import (
	"auth/internal/middleware"
	sessiontokens "auth/internal/session_tokens"
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminUsersService) http.Handler {
	r := chi.NewRouter()

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
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	//	@Param			userId	path	string				true	"User ID"
	//	@Param			request	body	UpdateUserRoleBody	true	"User role assignment"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/users/{userId}/roles [post]
	r.Post("/{userId}/roles", func(w http.ResponseWriter, r *http.Request) {
		var body UpdateUserRoleBody
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		err := s.AddUserRole(r.Context(), UpdateUserRoleParams{
			UserID: chi.URLParam(r, "userId"),
			Role:   body.Role,
		}, ctx.Subject)
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
	//	@Param			userId	path	string	true	"User ID"
	//	@Param			role	path	string	true	"Role name"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/users/{userId}/roles/{role} [delete]
	r.Delete("/{userId}/roles/{role}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		err := s.RemoveUserRole(r.Context(), UpdateUserRoleParams{
			UserID: chi.URLParam(r, "userId"),
			Role:   chi.URLParam(r, "role"),
		}, ctx.Subject)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
