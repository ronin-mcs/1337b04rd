package httpadapter

import (
	"1337b04rd/internal/domain"
	"1337b04rd/models"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

var httpAdapterLogger = slog.With("adapter", "http")

const maxUploadSize = 50 << 20  // 50 MB
const maxParseMemory = 10 << 20 // 10 MB

type CatalogPageData struct {
	Posts []*domain.PostView
	Title string
	Error string
}

type PostPageData struct {
	Post  domain.PostView
	Error string
}

// done
func (h *PostHandler) getPosts(w http.ResponseWriter, r *http.Request, page, limit int) {

	file, err := os.Open("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to open template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	posts, err := h.postService.GetActualPosts()
	if err != nil {
		http.Error(w, "failed to get posts", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		data := CatalogPageData{
			Posts: nil,
			Title: "Catalog",
			Error: "No posts found",
		}

		tmpl, err := template.ParseFiles("templates/catalog.html")
		if err != nil {
			httpAdapterLogger.Error("failed to parse template file", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			httpAdapterLogger.Error("failed to execute template", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	pages := len(posts) / limit
	if page > pages {
		httpAdapterLogger.Warn("requested page exceeds total pages. It will redirect to last page automatically.", "requested_page", page, "total_pages", pages)
		http.Redirect(w, r, fmt.Sprintf("/posts?page=%d&limit=%d", pages, limit), http.StatusMovedPermanently)
	}
	if page < 1 {
		httpAdapterLogger.Warn("requested page is less than 1. It will redirect to first page automatically.", "requested_page", page)
		http.Redirect(w, r, fmt.Sprintf("/posts?page=1&limit=%d", limit), http.StatusMovedPermanently)
	}

	posts = posts[(page-1)*limit : min(page*limit, len(posts))]

	postViews, err := h.postService.ConstructCatalogPostViews(posts)
	if err != nil {
		http.Error(w, "failed to construct post views", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := CatalogPageData{
		Posts: postViews,
		Title: "Catalog",
		Error: "",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// done
func (h *PostHandler) getCreatePost(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/create-post.html")
}

// done
func (h *PostHandler) createPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize) // 50 MB

	err := r.ParseMultipartForm(maxParseMemory) // 10 MB
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpAdapterLogger.Warn("request body too large", "error", err)
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		httpAdapterLogger.Error("failed to parse multipart form", "error", err)
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	title := r.FormValue("title")
	text := r.FormValue("text")

	files := make(map[string]io.Reader)

	for _, fileHeader := range r.MultipartForm.File["files"] {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "failed to open uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if _, ok := files[fileHeader.Filename]; ok {
			httpAdapterLogger.Warn("duplicate filename in uploaded files. They will be named with the corresponding dates", "filename", fileHeader.Filename)
			files[fmt.Sprintf("%s_%d", fileHeader.Filename, time.Now().Unix())] = file
			continue
		}
		files[fileHeader.Filename] = file
	}

	post := &models.Post{
		Title:       title,
		TextContent: text,
		AnonID:      0, // will be set in service
	}

	postID, err := h.postService.CreatePost(post, files)
	if err != nil {
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/posts/%d", postID), http.StatusSeeOther)
}

// done
func (h *PostHandler) getPostByID(w http.ResponseWriter, r *http.Request, postID string) {
	file, err := os.Open("templates/post.html")
	if err != nil {
		httpAdapterLogger.Error("failed to open template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	post, err := h.postService.GetActualPostByID(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	postView, err := h.postService.ConstructPostPagePostView(*post)

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := PostPageData{
		Post:  *postView,
		Error: "",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// done
func (h *PostHandler) createCommentToPost(w http.ResponseWriter, r *http.Request, postID string, commentID string) {
	// парсим postID
	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var addressedTo int
	if commentID == "" {
		addressedTo = id
	} else {
		addressedTo, err = strconv.Atoi(commentID)
		if err != nil {
			http.Error(w, "invalid comment id", http.StatusBadRequest)
			return
		}
	}

	// ===============================================================================
	// читатем форму

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize) // 50 MB

	err = r.ParseMultipartForm(maxParseMemory) // 10 MB
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpAdapterLogger.Warn("request body too large", "postID", postID, "AddressedTo", addressedTo, "error", err)
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		httpAdapterLogger.Error("failed to parse multipart form", "postID", postID, "AddressedTo", addressedTo, "error", err)
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	// ===============================================================================

	text := r.FormValue("comment")

	files := make(map[string]io.Reader)

	for _, fileHeader := range r.MultipartForm.File["files"] {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "failed to open uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if _, ok := files[fileHeader.Filename]; ok {
			httpAdapterLogger.Warn("duplicate filename in uploaded files. They will be named with the corresponding dates", "filename", fileHeader.Filename)
			files[fmt.Sprintf("%s_%d", fileHeader.Filename, time.Now().Unix())] = file
			continue
		}
		files[fileHeader.Filename] = file
	}

	// ===============================================================================
	// reading sessionID
	sessionID, err := h.readSessionIDFromCookie(r)
	if err.Error() == "no session cookie found" {
		httpAdapterLogger.Warn("treating user as anonymous due to missing session cookie", "postID", postID, "error", err)
		sessionID = h.setSessionCookie(w) // set a session cookie with session ID picked by repo or 0 for anonymous users
	}

	if err != nil && err.Error() != "no session cookie found" {
		httpAdapterLogger.Warn("failed to read session ID from cookie", "postID", postID, "error", err)
		sessionID = 0 // treat as anonymous
	}

	// ===============================================================================

	comment := &models.Comment{
		PostID:      id,
		AddressedTo: addressedTo,
		TextContent: text,
		AnonID:      0, // will be set in service
	}

	err = h.postService.CreateComment(comment, files, sessionID)
	if err != nil {
		http.Error(w, "failed to create comment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/posts/%d", id), http.StatusSeeOther)
}

// done
func (h *PostHandler) getArchive(w http.ResponseWriter, r *http.Request, page, limit int) {
	file, err := os.Open("templates/archive.html")
	if err != nil {
		httpAdapterLogger.Error("failed to open template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	posts, err := h.postService.GetArchivedPosts()
	if err != nil {
		http.Error(w, "failed to get posts", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		data := CatalogPageData{
			Posts: nil,
			Title: "Catalog",
			Error: "No posts found",
		}

		tmpl, err := template.ParseFiles("templates/catalog.html")
		if err != nil {
			httpAdapterLogger.Error("failed to parse template file", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			httpAdapterLogger.Error("failed to execute template", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	pages := len(posts) / limit
	if page > pages {
		httpAdapterLogger.Warn("requested page exceeds total pages. It will redirect to last page automatically.", "requested_page", page, "total_pages", pages)
		http.Redirect(w, r, fmt.Sprintf("/posts?page=%d&limit=%d", pages, limit), http.StatusMovedPermanently)
	}
	if page < 1 {
		httpAdapterLogger.Warn("requested page is less than 1. It will redirect to first page automatically.", "requested_page", page)
		http.Redirect(w, r, fmt.Sprintf("/posts?page=1&limit=%d", limit), http.StatusMovedPermanently)
	}

	posts = posts[(page-1)*limit : min(page*limit, len(posts))]

	postViews, err := h.postService.ConstructCatalogPostViews(posts)
	if err != nil {
		http.Error(w, "failed to construct post views", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := CatalogPageData{
		Posts: postViews,
		Title: "Catalog",
		Error: "",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) readSessionIDFromCookie(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			httpAdapterLogger.Warn("no session cookie found", "error", err)
			return 0, errors.New("no session cookie found")
		}
		httpAdapterLogger.Error("failed to get session cookie", "error", err)
		return 0, errors.New("failed to get session cookie")
	}

	sessiodIDRaw := cookie.Value
	sessionID, err := strconv.Atoi(sessiodIDRaw)
	if err != nil {
		httpAdapterLogger.Error("invalid session cookie value", "session_id_raw", sessiodIDRaw, "error", err)
		return 0, errors.New("invalid session cookie value")
	}
	return sessionID, nil
}

func (h *PostHandler) setSessionCookie(w http.ResponseWriter) int {
	sessionID, err := h.postService.CreateNewSessionID()
	if err != nil {
		httpAdapterLogger.Error("failed to upload cookie and initialize new Session")
		sessionID = 0
		return sessionID
	}
	value := strconv.Itoa(sessionID)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    value,
		Path:     "/",
		MaxAge:   3600 * 24 * 7,           // 7 weeks
		HttpOnly: true,                    // Без JS доступа
		Secure:   true,                    // Только HTTPS
		SameSite: http.SameSiteStrictMode, // Лучшая защита CSRF
	})
	return sessionID

}
