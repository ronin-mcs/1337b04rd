package rickandmortyapi

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"time"
)

const charactersURL = "https://rickandmortyapi.com/api/character"

var imagelogger = slog.With("adapter", "rickandmortyapi")

type AvatarFromAPI struct {
	Characters   map[int]string
	characterIDs []int
	client       *http.Client
}

type charactersResponse struct {
	Results []character `json:"results"`
}

type character struct {
	ID    int    `json:"id"`
	Image string `json:"image"`
}

func NewAvatarFromAPI() (*AvatarFromAPI, error) {
	characters, err := fetchCharacters(charactersURL)
	if err != nil {
		return nil, err
	}

	return &AvatarFromAPI{
		Characters:   characters,
		characterIDs: characterIDs(characters),
		client:       &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func fetchCharacters(url string) (map[int]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		imagelogger.Error("failed to fetch characters from API", "error", err)
		return nil, fmt.Errorf("get characters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		imagelogger.Error("unexpected status code when fetching characters", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("get characters: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		imagelogger.Error("failed to read characters response body", "error", err)
		return nil, fmt.Errorf("read characters response: %w", err)
	}

	var data charactersResponse
	if err := json.Unmarshal(body, &data); err != nil {
		imagelogger.Error("failed to unmarshal characters response", "error", err)
		return nil, fmt.Errorf("unmarshal characters response: %w", err)
	}

	characters := make(map[int]string, len(data.Results))
	for _, character := range data.Results {
		characters[character.ID] = character.Image
	}

	return characters, nil
}

func characterIDs(characters map[int]string) []int {
	ids := make([]int, 0, len(characters))
	for id := range characters {
		ids = append(ids, id)
	}
	return ids
}

func (a *AvatarFromAPI) GetRandomCharacterID() (int, error) {
	if len(a.characterIDs) == 0 {
		imagelogger.Error("no characters available for random selection")
		return 0, fmt.Errorf("no characters available")
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(a.characterIDs))))
	if err != nil {
		imagelogger.Error("failed to generate random character index", "error", err)
		return 0, fmt.Errorf("random character index: %w", err)
	}

	return a.characterIDs[index.Int64()], nil
}

func (a *AvatarFromAPI) GetAvatar(characterID int) (io.ReadCloser, string, error) {
	link, exists := a.Characters[characterID]
	if !exists {
		imagelogger.Warn("character ID not found in avatar storage", "character_id", characterID)
		return nil, "", fmt.Errorf("character ID %d not found", characterID)
	}

	resp, err := a.client.Get(link)
	if err != nil {
		imagelogger.Error("failed to fetch avatar image", "character_id", characterID, "url", link, "error", err)
		return nil, "", fmt.Errorf("get avatar image: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		imagelogger.Error("unexpected status code when fetching avatar image", "character_id", characterID, "status_code", resp.StatusCode)
		return nil, "", fmt.Errorf("get avatar image: unexpected status %s", resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return resp.Body, contentType, nil
}
