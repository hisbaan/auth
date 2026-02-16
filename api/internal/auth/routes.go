package auth

import (
	"auth/internal/httputil"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *AuthService) http.Handler {
	r := chi.NewRouter()

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

	r.Post("/reset-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not yet implemented"))
	})

	return r
}
