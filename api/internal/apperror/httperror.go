package apperror

import (
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

// HTTPError is the interface for errors that can be converted to HTTP status codes
type HTTPError interface {
	error
	StatusCode() int
}

// Error is the base error type that implements HTTPError
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s\n", e.Message)
}

func (e *Error) StatusCode() int {
	return e.Status
}

// New creates a new HTTPError with the given status and message
func New(status int, msg string) HTTPError {
	return &Error{Status: status, Message: msg}
}

// Bad Request (400)
func NewBadRequest(msg string) HTTPError {
	if msg == "" {
		msg = "Bad Request"
	}
	return &Error{Status: http.StatusBadRequest, Message: msg}
}

// Unauthorized (401)
func NewUnauthorized(msg string) HTTPError {
	if msg == "" {
		msg = "Unauthorized"
	} else {
		msg = fmt.Sprintf("Unauthorized: %s", msg)
	}
	return &Error{Status: http.StatusUnauthorized, Message: msg}
}

// Forbidden (403)
func NewForbidden(msg string) HTTPError {
	if msg == "" {
		msg = "Forbidden"
	} else {
		msg = fmt.Sprintf("Forbidden: %s", msg)
	}
	return &Error{Status: http.StatusForbidden, Message: msg}
}

// Not Found (404)
func NewNotFound(msg string) HTTPError {
	if msg == "" {
		msg = "Not Found"
	} else {
		msg = fmt.Sprintf("Not Found: %s", msg)
	}
	return &Error{Status: http.StatusNotFound, Message: msg}
}

// Conflict (409)
func NewConflict(msg string) HTTPError {
	if msg == "" {
		msg = "Conflict"
	} else {
		msg = fmt.Sprintf("Conflict: %s", msg)
	}
	return &Error{Status: http.StatusConflict, Message: msg}
}

// Unprocessable Entity (422)
func NewUnprocessableEntity(msg string) HTTPError {
	if msg == "" {
		msg = "Unprocessable Entity"
	} else {
		msg = fmt.Sprintf("Unprocessable Entity: %s", msg)
	}
	return &Error{Status: http.StatusUnprocessableEntity, Message: msg}
}

// Too Many Requests (429)
func NewTooManyRequests(msg string) HTTPError {
	if msg == "" {
		msg = "Too Many Requests"
	} else {
		msg = fmt.Sprintf("Too Many Requests: %s", msg)
	}
	return &Error{Status: http.StatusTooManyRequests, Message: msg}
}

// Internal Server Error (500)
func NewInternalServerError(msg string) HTTPError {
	if msg == "" {
		msg = "Internal Server Error"
	} else {
		msg = fmt.Sprintf("Internal Server Error: %s", msg)
	}
	return &Error{Status: http.StatusInternalServerError, Message: msg}
}

// Not Implemented (501)
func NewNotImplemented(msg string) HTTPError {
	if msg == "" {
		msg = "Not Implemented"
	} else {
		msg = fmt.Sprintf("Not Implemented: %s", msg)
	}
	return &Error{Status: http.StatusNotImplemented, Message: msg}
}

// Service Unavailable (503)
func NewServiceUnavailable(msg string) HTTPError {
	if msg == "" {
		msg = "Service Unavailable"
	} else {
		msg = fmt.Sprintf("Service Unavailable: %s", msg)
	}
	return &Error{Status: http.StatusServiceUnavailable, Message: msg}
}

func FromPGError(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23503": // foreign_key_violation
			return NewNotFound("Referenced resource not found")
		case "23505": // unique_violation
			return NewConflict("Resource already exists")
		case "23502": // not_null_violation
			return NewBadRequest("Required field missing")
		case "22001": // string_data_right_truncation
			return NewBadRequest("String too long")
		case "22003": // numeric_value_out_of_range
			return NewBadRequest("Numeric value out of range")
		case "22012": // division_by_zero
			return NewBadRequest("Division by zero")
		case "57014": // query_canceled
			return NewServiceUnavailable("Query cancelled")
		case "53300": // too_many_connections
			return NewServiceUnavailable("Too many connections")
		case "58030": // io_error
			return NewInternalServerError("Database I/O error")
		}
	}
	return NewInternalServerError("Database query error")
}
