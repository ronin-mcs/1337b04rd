package rickandmortyapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"info": {"next": null},
			"results": [
				{"id": 1, "image": "https://example.test/rick.png"},
				{"id": 2, "image": "https://example.test/morty.png"}
			]
		}`))
	}))
	defer server.Close()

	characters, err := fetchCharacters(server.URL)
	if err != nil {
		t.Fatalf("fetchCharacters() error = %v", err)
	}
	if len(characters) != 2 || characters[1] == "" || characters[2] == "" {
		t.Fatalf("characters = %#v, want two characters", characters)
	}
}

func TestRetryAfterDelay(t *testing.T) {
	if got := retryAfterDelay(""); got != 0 {
		t.Fatalf("empty retry delay = %v, want 0", got)
	}
	if got := retryAfterDelay("2"); got != 2*time.Second {
		t.Fatalf("numeric retry delay = %v, want 2s", got)
	}
	if got := retryAfterDelay(time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)); got != 0 {
		t.Fatalf("past retry delay = %v, want 0", got)
	}
}

func TestAvatarFromAPIGetters(t *testing.T) {
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png"))
	}))
	defer imageServer.Close()

	avatar := &AvatarFromAPI{
		Characters:   map[int]string{1: imageServer.URL},
		characterIDs: []int{1},
		client:       imageServer.Client(),
	}

	id, err := avatar.GetRandomCharacterID()
	if err != nil {
		t.Fatalf("GetRandomCharacterID() error = %v", err)
	}
	if id != 1 {
		t.Fatalf("id = %d, want 1", id)
	}

	ids := avatar.GetAllCharacterIDs()
	ids[0] = 99
	if avatar.characterIDs[0] != 1 {
		t.Fatal("GetAllCharacterIDs returned internal slice")
	}

	body, contentType, err := avatar.GetAvatar(1)
	if err != nil {
		t.Fatalf("GetAvatar() error = %v", err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if contentType != "image/png" || string(data) != "png" {
		t.Fatalf("avatar = %q %q, want image/png png", contentType, string(data))
	}
}
