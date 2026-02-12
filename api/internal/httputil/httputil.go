package httputil

import (
	"auth/internal/apperror"
	"encoding/json"
	"net/http"
)

func ParseBody(w http.ResponseWriter, r *http.Request, body any) {
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request"))
	}
}

func HandleErrors(w http.ResponseWriter, err error) {
	serr, ok := err.(apperror.HTTPError)
	if ok {
		w.WriteHeader(serr.StatusCode())
		w.Write([]byte(serr.Error()))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}
}
