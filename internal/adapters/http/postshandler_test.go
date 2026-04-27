package httpadapter

import (
	"net/http/httptest"
	"testing"
)

func TestQueryInt(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		key          string
		defaultValue int
		want         int
		wantErr      bool
	}{
		{name: "default", target: "/posts", key: "page", defaultValue: 1, want: 1},
		{name: "valid", target: "/posts?page=3", key: "page", defaultValue: 1, want: 3},
		{name: "negative rejected", target: "/posts?limit=-5", key: "limit", defaultValue: 20, wantErr: true},
		{name: "invalid", target: "/posts?page=abc", key: "page", defaultValue: 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.target, nil)
			got, err := queryInt(req, tt.key, tt.defaultValue)
			if (err != nil) != tt.wantErr {
				t.Fatalf("queryInt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("queryInt() = %d, want %d", got, tt.want)
			}
		})
	}
}
