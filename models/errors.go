package models

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrTooMany            = errors.New("too many requests")
	ErrNotImplemented     = errors.New("not implemented")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrGatewayTimeout     = errors.New("gateway timeout")
	// ================================
	ErrTooManyFiles      = errors.New("too many files")
	ErrPostIsArchived    = errors.New("post not found or is archived")
	ErrPostIsNotArchived = errors.New("post not found or is not archived")
	ErrNoSession         = errors.New("no session cookie found")
)

func StatusFromError(err error) int {
	switch {
	case err == nil:
		return http.StatusOK

	case errors.Is(err, ErrPostIsArchived):
		return http.StatusNotFound

	case errors.Is(err, ErrPostIsNotArchived):
		return http.StatusNotFound

	case errors.Is(err, ErrServiceUnavailable):
		return http.StatusServiceUnavailable

	case errors.Is(err, ErrGatewayTimeout):
		return http.StatusGatewayTimeout

	case errors.Is(err, ErrTooManyFiles):
		return http.StatusRequestEntityTooLarge

	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest

	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized

	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden

	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, ErrConflict):
		return http.StatusConflict

	case errors.Is(err, ErrTooMany):
		return http.StatusTooManyRequests

	case errors.Is(err, ErrNotImplemented):
		return http.StatusNotImplemented

	default:
		return http.StatusInternalServerError
	}
}
