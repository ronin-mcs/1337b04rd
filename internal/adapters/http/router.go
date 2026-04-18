package httpadapter

import "net/http"

func NewRouter(postService *domain.PostService) *http.ServeMux {
	mux := http.NewServeMux()

	postsHandler := NewPostsHandler(postService)

	mux.HandleFunc("/posts", postsHandler.Posts)
	mux.HandleFunc("/posts/create", postsHandler.Create)
	mux.HandleFunc("/posts/", postsHandler.PostByID)
	mux.HandleFunc("/archive", postsHandler.Archive)

	return mux
}
