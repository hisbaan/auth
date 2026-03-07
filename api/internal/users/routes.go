package users

import (
	"auth/internal/auth"
	"auth/internal/middleware"
	"auth/internal/utils/httputil"
	"auth/internal/utils/ulidutil"
	"crypto/ed25519"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *UsersService) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Auth(s.jwtSigningKey.Public().(ed25519.PublicKey), s.issuer))

	//	@Summary		Get current user
	//	@Description	Returns the authenticated user's profile
	//	@Tags			users
	//	@Security		BearerAuth
	//	@Success		200	{object}	GetUserResponse
	//	@Failure		401
	//	@Router			/users/me [get]
	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().Value(middleware.AuthContextKey).(*auth.AccessClaims)
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
		ctx := r.Context().Value(middleware.AuthContextKey).(*auth.AccessClaims)
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
		ctx := r.Context().Value(middleware.AuthContextKey).(*auth.AccessClaims)
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
		ctx := r.Context().Value(middleware.AuthContextKey).(*auth.AccessClaims)
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

	return r
}
