package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	httpadapter "1337b04rd/internal/adapters/http"
	dbadapter "1337b04rd/internal/adapters/postgres"
	"1337b04rd/internal/adapters/rickandmortyapi"
	s3storage "1337b04rd/internal/adapters/s3"
	"1337b04rd/internal/domain"

	_ "github.com/lib/pq"
)

func printUsage() {
	fmt.Println(`hacker board
Usage:
  1337b04rd [--port <N>]  
  1337b04rd --help

Options:
  --help       Show this screen.
  --port N     Port number.`)
}

func main() {
	closeLogs, err := SetupLogging("lastrunlogs.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup logging: %v\n", err)
		os.Exit(1)
	}
	defer closeLogs()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// print(dsn)

	db, err := sql.Open("postgres", dsn)
	// db, err := sql.Open("postgres", "host=localhost port=5432 user=user password=password dbname=app sslmode=disable")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	for i := 0; i < 100; i++ {
		err = db.Ping()
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}

	if err != nil {
		slog.Error("failed to connect to db", "error", err)
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

	S3_ENDPOINT := os.Getenv("S3_ENDPOINT")
	S3_PUBLIC_ENDPOINT := os.Getenv("S3_PUBLIC_ENDPOINT")
	S3_BUCKET := os.Getenv("S3_BUCKET")
	fileStorage := s3storage.NewS3Storage(S3_ENDPOINT, S3_PUBLIC_ENDPOINT, S3_BUCKET)
	// fileStorage := s3storage.NewS3Storage("http://localhost:9000", "http://localhost:9000", "1337b04rd")

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
	fmt.Println("Server started on port 8080")

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	slog.Info("Starting server", "port", *port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
