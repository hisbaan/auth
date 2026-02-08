package auth

import (
	"auth/httperrors"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func parseBody(w http.ResponseWriter, r *http.Request, body any) {
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request"))
	}
}

func handleErrors(w http.ResponseWriter, err error) {
	serr, ok := err.(httperrors.HTTPError)
	if ok {
		w.WriteHeader(serr.StatusCode())
		w.Write([]byte(serr.Error()))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}
}

func Router(s *AuthService) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var body CreateUserParams
		parseBody(w, r, &body)

		err := s.CreateUser(body)
		if err != nil {
			handleErrors(w, err)
		}
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var body LoginParams
		parseBody(w, r, &body)

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		userAgent := r.UserAgent()

		loginResponse, err := s.Login(body, ip, userAgent)
		if err != nil {
			handleErrors(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResponse)
	})

	r.Post("/refresh", func(w http.ResponseWriter, r *http.Request) {
		var body RefreshParams
		parseBody(w, r, &body)

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		userAgent := r.UserAgent()

		refreshResponse, err := s.Refresh(body, ip, userAgent)
		if err != nil {
			handleErrors(w, err)
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
