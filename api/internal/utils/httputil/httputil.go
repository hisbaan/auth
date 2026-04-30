package httputil

import (
	"auth/internal/apperror"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
)

const MaxRequestBodyBytes = 1 << 20

type ClientInfo struct {
	IP        string
	UserAgent string
}

type clientInfoKey struct{}

func ParseBody(w http.ResponseWriter, r *http.Request, body any) error {
	LimitBody(w, r)

	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return err
		}
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return err
	}
	return nil
}

func LimitBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
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
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func WithQuery(rawURL string, values url.Values) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	for key, entries := range values {
		query.Del(key)
		for _, entry := range entries {
			query.Add(key, entry)
		}
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func ClientInfoFromRequest(r *http.Request) ClientInfo {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	return ClientInfo{
		IP:        ip,
		UserAgent: r.UserAgent(),
	}
}

func WithClientInfo(ctx context.Context, info ClientInfo) context.Context {
	return context.WithValue(ctx, clientInfoKey{}, info)
}

func ClientInfoFromContext(ctx context.Context) (ClientInfo, bool) {
	info, ok := ctx.Value(clientInfoKey{}).(ClientInfo)
	return info, ok
}
