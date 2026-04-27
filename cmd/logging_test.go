package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSeparatorHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := &separatorHandler{
		handler: slog.NewTextHandler(&buf, nil),
		writer:  &buf,
		mu:      &sync.Mutex{},
	}

	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("handler should be enabled for info logs")
	}
	if err := handler.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "====") || !strings.Contains(got, "hello") {
		t.Fatalf("log output = %q, want separator and message", got)
	}
}

func TestSetupLogging(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "log-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	path := file.Name()
	_ = file.Close()

	closeLog, err := SetupLogging(path)
	if err != nil {
		t.Fatalf("SetupLogging() error = %v", err)
	}
	slog.Info("hello")
	if err := closeLog(); err != nil {
		t.Fatalf("close log error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "hello") {
		t.Fatalf("log file = %q, want message", string(data))
	}
}
