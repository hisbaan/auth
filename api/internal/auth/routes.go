package auth

import (
	"auth/internal/httputil"
	"encoding/json"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(s *AuthService) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var body CreateUserParams
		httputil.ParseBody(w, r, &body)

		err := s.CreateUser(body)
		if err != nil {
			httputil.HandleErrors(w, err)
		}
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var body LoginParams
		httputil.ParseBody(w, r, &body)

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
			httputil.HandleErrors(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResponse)
	})

	r.Post("/refresh", func(w http.ResponseWriter, r *http.Request) {
		var body RefreshParams
		httputil.ParseBody(w, r, &body)

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
			httputil.HandleErrors(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(refreshResponse)
	})

	r.Post("/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not yet implemented"))
	})

	r.Post("/reset-password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not yet implemented"))
	})

	return r
}
