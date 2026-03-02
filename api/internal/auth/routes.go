package auth

import (
	"auth/internal/utils/httputil"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *AuthService) http.Handler {
	r := chi.NewRouter()

	//	@Summary		Register a new user
	//	@Description	Creates a new user account and sends verification email
	//	@Tags			auth
	//	@Accept			json
	//	@Param			request	body	CreateUserParams	true	"User registration details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		409
	//	@Failure		500
	//	@Router			/auth/register [post]
	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var body CreateUserParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.CreateUser(body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Login user
	//	@Description	Authenticates user and returns access and refresh tokens
	//	@Tags			auth
	//	@Accept			json
	//	@Param			request	body		LoginParams	true	"Login credentials"
	//	@Success		200		{object}	LoginResponse
	//	@Failure		400
	//	@Failure		401
	//	@Failure		500
	//	@Router			/auth/login [post]
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var body LoginParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		userAgent := r.UserAgent()

		loginResponse, err := s.Login(body, ip, userAgent)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, loginResponse)
	})

	//	@Summary		Refresh access token
	//	@Description	Issues a new access token using a valid refresh token
	//	@Tags			auth
	//	@Accept			json
	//	@Param			request	body		RefreshParams	true	"Refresh token"
	//	@Success		200		{object}	RefreshResponse
	//	@Failure		400
	//	@Failure		401
	//	@Router			/auth/refresh [post]
	r.Post("/refresh", func(w http.ResponseWriter, r *http.Request) {
		var body RefreshParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		userAgent := r.UserAgent()

		refreshResponse, err := s.Refresh(body, ip, userAgent)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		httputil.JSONResponse(w, http.StatusOK, refreshResponse)
	})

	//	@Summary		Request password reset
	//	@Description	Sends a password reset email to the user
	//	@Tags			auth
	//	@Accept			json
	//	@Param			request	body	ForgotPasswordParams	true	"Email address"
	//	@Success		200
	//	@Failure		400
	//	@Failure		500
	//	@Router			/auth/forgot-password [post]
	r.Post("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		var body ForgotPasswordParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.ForgotPassword(body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	//	@Summary		Reset password
	//	@Description	Resets user password using a valid reset token
	//	@Tags			auth
	//	@Accept			json
	//	@Param			request	body	PasswordResetParams	true	"Password reset details"
	//	@Success		204
	//	@Failure		400
	//	@Failure		500
	//	@Router			/auth/password-reset [post]
	r.Post("/password-reset", func(w http.ResponseWriter, r *http.Request) {
		println("here")

		var body PasswordResetParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.PasswordReset(body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	//	@Summary		Verify email
	//	@Description	Verifies user email using a valid verification token
	//	@Tags			auth
	//	@Accept			json
	//	@Param			request	body	VerifyEmailParams	true	"Email verification token"
	//	@Success		204
	//	@Failure		400
	//	@Failure		500
	//	@Router			/auth/verify-email [post]
	r.Post("/verify-email", func(w http.ResponseWriter, r *http.Request) {
		var body VerifyEmailParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.VerifyEmail(body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
