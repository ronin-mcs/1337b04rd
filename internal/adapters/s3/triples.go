package s3storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"1337b04rd/models"
)

var s3logger = slog.With("adapter", "s3storage")

type S3Storage struct {
	baseURL   string
	publicURL string
	bucket    string
	client    *http.Client
}

func NewS3Storage(baseURL, publicURL, bucket string) *S3Storage {
	baseURL = strings.TrimRight(baseURL, "/")
	publicURL = strings.TrimRight(publicURL, "/")
	if publicURL == "" {
		publicURL = baseURL
	}

	return &S3Storage{
		baseURL:   baseURL,
		publicURL: publicURL,
		bucket:    bucket,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *S3Storage) SaveFile(fileKey string, fileData io.Reader, contentType string) error {
	if contentType == "" {
		s3logger.Warn("content type not provided, defaulting to application/octet-stream", "fileKey", fileKey)
		contentType = "application/octet-stream"
	}

	if err := s.ensureBucket(context.Background()); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, s.objectURL(s.baseURL, fileKey), fileData)
	if err != nil {
		s3logger.Error("failed to create request to save file", "fileKey", fileKey, "error", err)
		return err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(req)
	if err != nil {
		s3logger.Error("failed to save file to S3", "fileKey", fileKey, "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			return models.ErrServiceUnavailable
		}
		return models.ErrGatewayTimeout
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		s3logger.Error("unexpected status code when saving file to S3", "fileKey", fileKey, "status_code", resp.StatusCode)
		return fmt.Errorf("save file %q: unexpected status %s", fileKey, resp.Status)
	}

	return nil
}

func (s *S3Storage) GetFileLink(fileKey string) (string, error) {
	return s.objectURL(s.publicURL, fileKey), nil
}

func (s *S3Storage) DeleteFile(fileKey string) error {
	req, err := http.NewRequest(http.MethodDelete, s.objectURL(s.baseURL, fileKey), nil)
	if err != nil {
		s3logger.Error("failed to create request to delete file", "fileKey", fileKey, "error", err)
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s3logger.Error("failed to delete file from S3", "fileKey", fileKey, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		s3logger.Error("unexpected status code when deleting file from S3", "fileKey", fileKey, "status_code", resp.StatusCode)
		return fmt.Errorf("delete file %q: unexpected status %s", fileKey, resp.Status)
	}

	return nil
}

func (s *S3Storage) EnsureBucket(ctx context.Context) error {
	return s.ensureBucket(ctx)
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.baseURL+"/"+url.PathEscape(s.bucket), nil)
	if err != nil {
		s3logger.Error("failed to create request to ensure bucket exists", "bucket", s.bucket, "error", err)
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s3logger.Error("failed to ensure bucket exists", "bucket", s.bucket, "error", err)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusConflict:
		s3logger.Info("bucket ensured to exist", "bucket", s.bucket)
		return nil
	default:
		s3logger.Error("unexpected status code when ensuring bucket exists", "bucket", s.bucket, "status_code", resp.StatusCode)
		return fmt.Errorf("ensure bucket %q: unexpected status %s", s.bucket, resp.Status)
	}
}

func (s *S3Storage) objectURL(baseURL, fileKey string) string {
	parts := strings.Split(fileKey, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	url := baseURL + "/" + url.PathEscape(s.bucket) + "/" + strings.Join(parts, "_")
	return url
}
