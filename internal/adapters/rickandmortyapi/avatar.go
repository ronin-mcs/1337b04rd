package rickandmortyapi

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const charactersURL = "https://rickandmortyapi.com/api/character"

const (
	charactersPageDelay = 200 * time.Millisecond
	maxFetchRetries     = 5
)

var imagelogger = slog.With("adapter", "rickandmortyapi")

type AvatarFromAPI struct {
	Characters   map[int]string
	characterIDs []int
	client       *http.Client
}

type charactersResponse struct {
	Info    charactersInfo `json:"info"`
	Results []character    `json:"results"`
}

type charactersInfo struct {
	Next *string `json:"next"`
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
	var allCharacters []character
	client := &http.Client{Timeout: 10 * time.Second}

	for url != "" {
		// for range 1 {
		var resp *http.Response
		var err error

		for attempt := 0; attempt <= maxFetchRetries; attempt++ {
			resp, err = client.Get(url)
			if err != nil {
				imagelogger.Error("failed to fetch characters from API", "url", url, "error", err)
				return nil, fmt.Errorf("get characters: %w", err)
			}

			if resp.StatusCode != http.StatusTooManyRequests {
				break
			}

			resp.Body.Close()
			if attempt == maxFetchRetries {
				imagelogger.Error("too many requests when fetching characters", "url", url, "status_code", resp.StatusCode)
				return nil, fmt.Errorf("get characters: unexpected status %s", resp.Status)
			}

			delay := retryAfterDelay(resp.Header.Get("Retry-After"))
			if delay == 0 {
				delay = time.Duration(attempt+1) * time.Second
			}

			imagelogger.Warn("rate limited when fetching characters, retrying", "url", url, "delay", delay)
			time.Sleep(delay)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			imagelogger.Error("unexpected status code when fetching characters", "url", url, "status_code", resp.StatusCode)
			return nil, fmt.Errorf("get characters: unexpected status %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			imagelogger.Error("failed to read characters response body", "url", url, "error", err)
			return nil, fmt.Errorf("read characters response: %w", err)
		}

		var data charactersResponse
		if err := json.Unmarshal(body, &data); err != nil {
			imagelogger.Error("failed to unmarshal characters response", "url", url, "error", err)
			return nil, fmt.Errorf("unmarshal characters response: %w", err)
		}

		allCharacters = append(allCharacters, data.Results...)
		if data.Info.Next == nil {
			break
		}
		url = *data.Info.Next
		time.Sleep(charactersPageDelay)
	}

	characters := make(map[int]string, len(allCharacters))
	for _, character := range allCharacters {
		characters[character.ID] = character.Image
	}

	return characters, nil
}

func retryAfterDelay(value string) time.Duration {
	if value == "" {
		return 0
	}

	seconds, err := strconv.Atoi(value)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}

	retryTime, err := http.ParseTime(value)
	if err != nil {
		return 0
	}

	delay := time.Until(retryTime)
	if delay < 0 {
		return 0
	}

	return delay
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
		return 0, fmt.Errorf("no characters available for random selection")
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(a.characterIDs))))
	if err != nil {
		imagelogger.Error("failed to generate random character index", "error", err)
		return 0, fmt.Errorf("random character index: %w", err)
	}

	return a.characterIDs[index.Int64()], nil
}

func (a *AvatarFromAPI) GetAllCharacterIDs() []int {
	ids := make([]int, len(a.characterIDs))
	copy(ids, a.characterIDs)
	return ids
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
