package httpadapter

import (
	"net/http"
	"strconv"
	"strings"
)

type PostHandler struct {
	postService *domain.PostService
}

func NewPostsHandler(postService *domain.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) Posts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /posts?page=1&limit=20
		page, err := queryInt(r, "page", 1)
		if err != nil {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}

		limit, err := queryInt(r, "limit", 20)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}

		h.getPosts(page, limit)
	default:
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PostHandler) PostByID(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/posts/" {
		http.Redirect(w, r, "/posts", http.StatusMovedPermanently)
		// можно использовать http.StatusPermanentRedirect или http.StatusTemporaryRedirect
		// нужно позже почитать о чём они
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/posts/")
	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(path, "/")
	postID := parts[0]

	// /posts/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			// GET /posts/{id}
			h.getPostByID(postID)

		default:
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /posts/{id}/comments
	if len(parts) == 2 && parts[1] == "comments" {
		switch r.Method {
		case http.MethodPost:
			// POST /posts/{id}/comments
			h.createCommentToPost(postID)
		default:
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) == 4 && parts[1] == "comments" && parts[3] == "replies" {
		commentID := parts[2]
		if commentID == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodPost:
			h.createCommentToComment(postID, commentID)
		default:
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return

	}

	http.NotFound(w, r)
}

func (h *PostHandler) Archive(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /archive?page=1&limit=20
		page, err := queryInt(r, "page", 1)
		if err != nil {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}

		limit, err := queryInt(r, "limit", 20)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}

		h.getArchive(page, limit)
	default:
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /posts/create
		h.getCreatePost(w, r)
	case http.MethodPost:
		// POST /posts/create
		h.createPost(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func queryInt(r *http.Request, name string, defaultValue int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 1 {
		return 0, strconv.ErrSyntax
	}

	return value, nil
}
