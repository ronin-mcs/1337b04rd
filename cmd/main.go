package main

import (
	httpadapter "1337b04rd/internal/adapters/http"
	dbadapter "1337b04rd/internal/adapters/postgres"
	"1337b04rd/internal/adapters/rickandmortyapi"
	s3storage "1337b04rd/internal/adapters/s3"
	"1337b04rd/internal/domain"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func printUsage() {
	fmt.Println(`hacker board

Usage:
  1337b04rd [--port <N>]  
  1337b04rd --help

Options:
  --help       Show this screen.
  --port N     Port number.
`)
}

func main() {
	closeLogs, err := setupLogging("lastrunlogs.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup logging: %v\n", err)
		os.Exit(1)
	}
	defer closeLogs()

	db, err := sql.Open("postgres", "host=db port=5432 user=latte password=latte dbname=frappuccino sslmode=disable")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	help := flag.Bool("help", false, "Show help")
	port := flag.Int("port", 8080, "Port number")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	postRepo := dbadapter.NewPGPostsRepository(db)
	commentRepo := dbadapter.NewPGCommentsRepository(db)
	sessions := dbadapter.NewPGSessionsRepository(db)
	anonRepo := dbadapter.NewPGAnonsRepository(db)
	attachmentsRepo := dbadapter.NewAttachmentsRepo(db)

	avatarStorage, err := rickandmortyapi.NewAvatarFromAPI()
	if err != nil {
		slog.Error("failed to initialize avatar storage", "error", err)
		os.Exit(1)
	}

	fileStorage := s3storage.NewS3Storage("http://localhost:9000", "1337b04rd")

	postService := domain.NewPostService(avatarStorage, fileStorage, postRepo, commentRepo, anonRepo, sessions, attachmentsRepo)

	if err := sessions.DeleteExpired(); err != nil {
		slog.Error("delete expired sessions on startup failed", "error", err)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if err := sessions.DeleteExpired(); err != nil {
				slog.Error("delete expired sessions failed", "error", err)
			}
		}
	}()

	mux := httpadapter.NewRouter(postService)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	slog.Info("Starting server", "port", *port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
