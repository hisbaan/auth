package oidc

import (
	"errors"
	"net/http"

	"auth/internal/utils/httputil"
)

type TokenError interface {
	error
	StatusCode() int
	ErrorCode() string
	ErrorDescription() string
}

type tokenError struct {
	status      int
	code        string
	description string
}

type tokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (e *tokenError) Error() string {
	return e.description
}

func (e *tokenError) StatusCode() int {
	return e.status
}

func (e *tokenError) ErrorCode() string {
	return e.code
}

func (e *tokenError) ErrorDescription() string {
	return e.description
}

func NewInvalidRequestTokenError(description string) TokenError {
	return &tokenError{status: http.StatusBadRequest, code: "invalid_request", description: description}
}

func NewInvalidGrantTokenError(description string) TokenError {
	return &tokenError{status: http.StatusBadRequest, code: "invalid_grant", description: description}
}

func NewInvalidClientTokenError(description string) TokenError {
	return &tokenError{status: http.StatusUnauthorized, code: "invalid_client", description: description}
}

func NewUnsupportedGrantTypeTokenError(description string) TokenError {
	return &tokenError{status: http.StatusBadRequest, code: "unsupported_grant_type", description: description}
}

func NewServerTokenError() TokenError {
	return &tokenError{status: http.StatusInternalServerError, code: "server_error", description: "Internal server error"}
}

func HandleTokenError(w http.ResponseWriter, err error) {
	var tokenErr TokenError
	if errors.As(err, &tokenErr) {
		httputil.JSONResponse(w, tokenErr.StatusCode(), tokenErrorResponse{
			Error:            tokenErr.ErrorCode(),
			ErrorDescription: tokenErr.ErrorDescription(),
		})
		return
	}

	serverErr := NewServerTokenError()
	httputil.JSONResponse(w, serverErr.StatusCode(), tokenErrorResponse{
		Error:            serverErr.ErrorCode(),
		ErrorDescription: serverErr.ErrorDescription(),
	})
}
