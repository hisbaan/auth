package httputil

import (
	"auth/internal/apperror"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const MaxRequestBodyBytes = 1 << 20

type ClientInfo struct {
	IP        string
	UserAgent string
}

type clientInfoKey struct{}

// extracts the token from the Authorization header.
// errors if the header is absent or is not a non-empty Bearer credential.
func BearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if header == "" || token == header || token == "" {
		return "", apperror.NewUnauthorized("Unauthorized")
	}
	return token, nil
}

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
	var berr apperror.BodyError
	if errors.As(err, &berr) && berr.ErrorBody() != nil {
		JSONResponse(w, berr.StatusCode(), berr.ErrorBody())
		return
	}

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
	return ClientInfo{
		IP:        clientIPFromRequest(r),
		UserAgent: r.UserAgent(),
	}
}

func clientIPFromRequest(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		for entry := range strings.SplitSeq(forwarded, ",") {
			if ip := normalizeIP(entry); ip != "" {
				return ip
			}
		}
	}

	if ip := normalizeIP(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	if ip := normalizeIP(r.RemoteAddr); ip != "" {
		return ip
	}

	return "0.0.0.0"
}

func normalizeIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}

	if host, _, err := net.SplitHostPort(value); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}

	return ""
}

func WithClientInfo(ctx context.Context, info ClientInfo) context.Context {
	return context.WithValue(ctx, clientInfoKey{}, info)
}

func ClientInfoFromContext(ctx context.Context) (ClientInfo, bool) {
	info, ok := ctx.Value(clientInfoKey{}).(ClientInfo)
	return info, ok
}
