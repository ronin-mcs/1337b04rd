package models

import (
	"errors"
	"net/http"
	"testing"
)

func TestStatusFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: http.StatusOK},
		{name: "post archived", err: ErrPostIsArchived, want: http.StatusNotFound},
		{name: "post not archived", err: ErrPostIsNotArchived, want: http.StatusNotFound},
		{name: "service unavailable", err: ErrServiceUnavailable, want: http.StatusServiceUnavailable},
		{name: "gateway timeout", err: ErrGatewayTimeout, want: http.StatusGatewayTimeout},
		{name: "bad request", err: ErrBadRequest, want: http.StatusBadRequest},
		{name: "unauthorized", err: ErrUnauthorized, want: http.StatusUnauthorized},
		{name: "forbidden", err: ErrForbidden, want: http.StatusForbidden},
		{name: "conflict", err: ErrConflict, want: http.StatusConflict},
		{name: "too many", err: ErrTooMany, want: http.StatusTooManyRequests},
		{name: "not implemented", err: ErrNotImplemented, want: http.StatusNotImplemented},
		{name: "wrapped not found", err: errors.Join(errors.New("lookup failed"), ErrNotFound), want: http.StatusNotFound},
		{name: "too many files", err: ErrTooManyFiles, want: http.StatusRequestEntityTooLarge},
		{name: "unknown", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusFromError(tt.err); got != tt.want {
				t.Fatalf("StatusFromError(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
