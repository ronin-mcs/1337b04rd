package s3storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestS3StorageFileOperations(t *testing.T) {
	var requests []string
	var putContentType string
	var putBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.EscapedPath())
		switch r.Method {
		case http.MethodPut:
			if r.URL.EscapedPath() == "/bucket%20name" {
				w.WriteHeader(http.StatusConflict)
				return
			}

			putContentType = r.Header.Get("Content-Type")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			putBody = string(body)
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	storage := NewS3Storage(server.URL+"/", "https://public.test/files", "bucket name")
	if err := storage.SaveFile("dir/file name.txt", strings.NewReader("hello"), ""); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}
	if err := storage.DeleteFile("dir/file name.txt"); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	wantRequests := []string{
		"PUT /bucket%20name",
		"PUT /bucket%20name/dir_file%20name.txt",
		"DELETE /bucket%20name/dir_file%20name.txt",
	}
	if strings.Join(requests, "\n") != strings.Join(wantRequests, "\n") {
		t.Fatalf("requests = %v, want %v", requests, wantRequests)
	}
	if putContentType != "application/octet-stream" {
		t.Fatalf("content type = %q, want default", putContentType)
	}
	if putBody != "hello" {
		t.Fatalf("put body = %q, want hello", putBody)
	}

	link, err := storage.GetFileLink("dir/file name.txt")
	if err != nil {
		t.Fatalf("GetFileLink() error = %v", err)
	}
	if link != "https://public.test/files/bucket%20name/dir_file%20name.txt" {
		t.Fatalf("link = %q, want escaped bucket and key", link)
	}
}

func TestS3StorageEnsureBucket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusConflict)
	}))
	defer server.Close()

	storage := NewS3Storage(server.URL, "", "existing")
	if err := storage.EnsureBucket(context.Background()); err != nil {
		t.Fatalf("EnsureBucket() error = %v", err)
	}
}
