package httpadapter

import (
	"1337b04rd/internal/domain"
	"fmt"
	"net/http"
	"strings"
)

func NewRouter(postService *domain.PostService) http.Handler {
	postsHandler := NewPostsHandler(postService)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" {
			path = strings.TrimRight(path, "/")
			r.URL.Path = path
		}

		switch {
		case path == "/posts":
			fmt.Println("/posts route")
			postsHandler.Posts(w, r)

		case path == "/posts/create":
			fmt.Println("/posts/create route")
			postsHandler.Create(w, r)

		case strings.HasPrefix(path, "/posts/"):
			fmt.Println("/posts/{id} route")
			postsHandler.PostByID(w, r)

		case path == "/archive":
			fmt.Println("/archive route")
			postsHandler.Archive(w, r)

		case strings.HasPrefix(path, "/archive/"):
			fmt.Println("/archive/{id} route")
			postsHandler.Archive(w, r)

		case path == "/":
			fmt.Println("/ route")

		default:
			fmt.Println("not found route:", path)
			http.NotFound(w, r)
		}
	})
}
