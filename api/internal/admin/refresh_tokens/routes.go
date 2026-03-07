package refreshtokens

import (
	"net/http"
	"strconv"

	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"

	"github.com/go-chi/chi/v5"
)

func Router(s *AdminRefreshTokensService) http.Handler {
	r := chi.NewRouter()

	//	@Summary		List user refresh tokens
	//	@Description	Returns a paginated list of refresh tokens for a user (admin only)
	//	@Tags			admin
	//	@Security		BearerAuth
	//	@Param			userId	path		string	true	"User ID"
	//	@Param			cursor	query		string	false	"Pagination cursor"
	//	@Param			limit	query		int		false	"Maximum number of results (max 100)"
	//	@Success		200		{object}	ListRefreshTokensResponse
	//	@Failure		400
	//	@Failure		401
	//	@Failure		403
	//	@Router			/admin/refresh-tokens/users/{userId} [get]
	r.Get("/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := ulidutil.FromPrefixed("user", chi.URLParam(r, "userId"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		params := listRefreshTokensParams(r)
		response, err := s.ListUserRefreshTokens(userID, params)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	return r
}

func listRefreshTokensParams(r *http.Request) ListRefreshTokensParams {
	params := ListRefreshTokensParams{
		Limit:  20,
		Cursor: r.URL.Query().Get("cursor"),
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			params.Limit = n
		}
	}

	return params
}
