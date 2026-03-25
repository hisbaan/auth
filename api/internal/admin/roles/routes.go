package roles

import (
	"auth/internal/middleware"
	sessiontokens "auth/internal/session_tokens"
	"auth/internal/utils/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminRolesService) http.Handler {
	r := chi.NewRouter()

	//	@Summary		Create role
	//	@Description	Creates a new role (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			request	body	CreateRoleParams	true	"Role details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/roles [post]
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var body CreateRoleParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		err := s.CreateRole(r.Context(), body, ctx.Subject)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Delete role
	//	@Description	Deletes a role (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			name	path	string	true	"Role name"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/roles/{name} [delete]
	r.Delete("/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		if name == "" {
			http.Error(w, "Role name required", http.StatusBadRequest)
			return
		}

		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		err := s.DeleteRole(r.Context(), DeleteRoleParams{Name: name}, ctx.Subject)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
