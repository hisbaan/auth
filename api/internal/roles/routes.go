package roles

import (
	"auth/internal/utils/httputil"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *RolesService) http.Handler {
	r := chi.NewRouter()

	//	@Summary		List roles
	//	@Description	Returns a list of all available roles
	//	@Tags			roles
	//	@Success		200	{object}	ListRolesResponse
	//	@Router			/roles [get]
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response, err := s.ListRoles()
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	return r
}
