package users

import (
	"auth/internal/middleware"
	sessiontokens "auth/internal/session_tokens"
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func Router(s *UsersService) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(s.jwtSigningKey.Public().(ed25519.PublicKey), s.issuer))
	r.Use(middleware.RequireTokenSource(sessiontokens.TokenSourceSelf))

	//	@Summary		Get current user
	//	@Description	Returns the authenticated user's profile
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Success		200	{object}	GetUserResponse
	//	@Failure		401
	//	@Router			/users/me [get]
	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		response, err := s.GetUser(userID)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	//	@Summary		Update current user
	//	@Description	Updates the authenticated user's profile
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			request	body	UpdateUserParams	true	"User update details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		409
	//	@Router			/users/me [put]
	r.Put("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		var body UpdateUserParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err = s.UpdateUser(r.Context(), userID, body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Change password
	//	@Description	Changes the authenticated user's password
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			request	body	UpdatePasswordParams	true	"Password change details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Router			/users/me/password [post]
	r.Post("/me/password", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		var body UpdatePasswordParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err = s.UpdatePassword(r.Context(), userID, body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Delete current user
	//	@Description	Deletes the authenticated user's account
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Success		204
	//	@Failure		401
	//	@Router			/users/me [delete]
	r.Delete("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		err = s.DeleteUser(r.Context(), userID)
		if err != nil {
			httputil.HandleError(w, err)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		List current user's clients
	//	@Description	Returns a paginated list of OAuth clients for the authenticated user
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Param			cursor	query		string	false	"Pagination cursor"
	//	@Param			limit	query		int		false	"Maximum number of results (max 100)"
	//	@Success		200		{object}	ListClientsResponse
	//	@Failure		401
	//	@Router			/users/me/clients [get]
	r.Get("/me/clients", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		params := ListClientsParams{
			Limit:  20,
			Cursor: r.URL.Query().Get("cursor"),
		}
		if limit := r.URL.Query().Get("limit"); limit != "" {
			if n, err := strconv.Atoi(limit); err == nil {
				params.Limit = n
			}
		}

		response, err := s.ListClients(userID, params)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	r.Get("/me/authorizations", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		response, err := s.ListClientAuthorizations(userID)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, response)
	})

	r.Post("/me/authorizations/{clientId}/revoke", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		clientID, err := ulidutil.FromPrefixed("client", chi.URLParam(r, "clientId"))
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		if err := s.RevokeClientAuthorization(r.Context(), userID, clientID); err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Create client
	//	@Description	Creates an OAuth client for the authenticated user
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			request	body		ClientParams	true	"Client details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		409
	//	@Router			/users/me/clients [post]
	r.Post("/me/clients", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		var body ClientParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err = s.CreateClient(r.Context(), userID, body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Update client
	//	@Description	Updates an OAuth client owned by the authenticated user
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Accept			json
	//	@Param			clientId	path		string		true	"Client ID"
	//	@Param			request	body		ClientParams	true	"Updated client details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		404
	//	@Failure		409
	//	@Router			/users/me/clients/{clientId} [put]
	r.Put("/me/clients/{clientId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		clientID, err := ulidutil.FromPrefixed("client", chi.URLParam(r, "clientId"))
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		var body ClientParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err = s.UpdateClient(r.Context(), userID, clientID, body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Revoke client
	//	@Description	Revokes an OAuth client owned by the authenticated user
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Param			clientId	path	string	true	"Client ID"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		404
	//	@Failure		409
	//	@Router			/users/me/clients/{clientId}/revoke [post]
	r.Post("/me/clients/{clientId}/revoke", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		clientID, err := ulidutil.FromPrefixed("client", chi.URLParam(r, "clientId"))
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		err = s.RevokeClient(r.Context(), userID, clientID)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Delete client
	//	@Description	Deletes an OAuth client owned by the authenticated user
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Param			clientId	path	string	true	"Client ID"
	//	@Success		204
	//	@Failure		400
	//	@Failure		401
	//	@Failure		404
	//	@Router			/users/me/clients/{clientId} [delete]
	r.Delete("/me/clients/{clientId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*sessiontokens.AccessClaims)
		userID, err := ulidutil.FromPrefixed("user", ctx.Subject)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		clientID, err := ulidutil.FromPrefixed("client", chi.URLParam(r, "clientId"))
		if err != nil {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		err = s.DeleteClient(r.Context(), userID, clientID)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
