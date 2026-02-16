package httputil

import (
	"auth/internal/apperror"
	"encoding/json"
	"net/http"
)

func ParseBody(w http.ResponseWriter, r *http.Request, body any) error {
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return err
	}
	return nil
}

func HandleError(w http.ResponseWriter, err error) {
	serr, ok := err.(apperror.HTTPError)
	if ok {
		http.Error(w, serr.Error(), serr.StatusCode())
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func JSONResponse(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
