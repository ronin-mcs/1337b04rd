package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

	postRepo := dbadpater.NewPostRepo(db)

	// ...

	mux := http.NewServeMux()

	httpadapter.NewRouter(postService)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	slog.Info("Starting server", "port", *port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
