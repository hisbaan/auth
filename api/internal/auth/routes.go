package auth

import (
	"auth/internal/apperror"
	securitymiddleware "auth/internal/middleware"
	"auth/internal/utils/httputil"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

func (s *AuthService) setAuthCookies(w http.ResponseWriter, accessToken string, refreshToken string) {
	httpOnly := true
	sameSite := http.SameSiteLaxMode
	secure := s.cookieDomain != "localhost"

	accessCookie := http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    accessToken,
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   s.cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(15) * time.Minute),
	}
	http.SetCookie(w, &accessCookie)

	refreshCookie := http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshToken,
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   s.cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(168) * time.Hour),
	}
	http.SetCookie(w, &refreshCookie)
}

func (s *AuthService) clearAuthCookies(w http.ResponseWriter) {
	accessCookie := http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Domain:   s.cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
	}
	http.SetCookie(w, &accessCookie)

	refreshCookie := http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Domain:   s.cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
	}
	http.SetCookie(w, &refreshCookie)
}

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
	r.With(securitymiddleware.RateLimit(5, time.Minute)).Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var body CreateUserParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.CreateUser(r.Context(), body)
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
	r.With(securitymiddleware.RateLimit(10, time.Minute)).Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var body LoginParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		loginResponse, err := s.Login(r.Context(), body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		s.setAuthCookies(w, loginResponse.AccessToken, loginResponse.RefreshToken)

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
	r.With(securitymiddleware.RateLimit(30, time.Minute)).Post("/refresh", func(w http.ResponseWriter, r *http.Request) {
		var body RefreshParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		if body.RefreshToken == "" {
			cookie, err := r.Cookie(RefreshTokenCookieName)
			if err == nil && cookie != nil {
				body.RefreshToken = cookie.Value
			}
		}

		if body.RefreshToken == "" {
			httputil.HandleError(w, apperror.NewBadRequest("Refresh token is required"))
			return
		}

		refreshResponse, err := s.Refresh(r.Context(), body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		s.setAuthCookies(w, refreshResponse.AccessToken, refreshResponse.RefreshToken)

		httputil.JSONResponse(w, http.StatusOK, refreshResponse)
	})

	//	@Summary		Logout user
	//	@Description	Clears authentication cookies
	//	@Tags			auth
	//	@Success		204
	//	@Router			/auth/logout [post]
	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		token := ""
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				token = ""
			}
		} else {
			cookie, err := r.Cookie(AccessTokenCookieName)
			if err == nil && cookie != nil {
				token = cookie.Value
			}
		}
		s.Logout(r.Context(), token)
		s.clearAuthCookies(w)
		w.WriteHeader(http.StatusNoContent)
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
	r.With(securitymiddleware.RateLimit(5, time.Minute)).Post("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		var body ForgotPasswordParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.ForgotPassword(r.Context(), body)
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
	r.With(securitymiddleware.RateLimit(10, time.Minute)).Post("/password-reset", func(w http.ResponseWriter, r *http.Request) {
		var body PasswordResetParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.PasswordReset(r.Context(), body)
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
	r.With(securitymiddleware.RateLimit(20, time.Minute)).Post("/verify-email", func(w http.ResponseWriter, r *http.Request) {
		var body VerifyEmailParams
		if err := httputil.ParseBody(w, r, &body); err != nil {
			return
		}

		err := s.VerifyEmail(r.Context(), body)
		if err != nil {
			httputil.HandleError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
