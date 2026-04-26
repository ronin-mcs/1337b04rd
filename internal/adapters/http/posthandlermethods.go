package httpadapter

import (
	"1337b04rd/internal/domain"
	"1337b04rd/models"
	"bytes"
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
	posts, err := h.postService.GetActualPosts()
	if err != nil {
		h.RenderError(w, models.StatusFromError(err), "failed to get posts")
		return
	}

	if len(posts) == 0 {
		h.RenderError(w, http.StatusNotFound, "No posts found")
		return
	}

	pages := (len(posts) + limit - 1) / limit // round up
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
		h.RenderError(w, models.StatusFromError(err), "failed to construct post views")
		return
	}

	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err, "template_name", "catalog.html")
		h.RenderError(w, models.StatusFromError(err), "internal server error")
		return
	}

	data := CatalogPageData{
		Posts: postViews,
		Title: "Catalog",
		Error: "",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err, "template_name", "catalog.html")
		h.RenderError(w, models.StatusFromError(err), "internal server error")
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
			h.RenderError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		httpAdapterLogger.Error("failed to parse multipart form", "error", err)
		h.RenderError(w, http.StatusBadRequest, "failed to parse form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	title := r.FormValue("subject")
	text := r.FormValue("comment")

	files := make(map[string]io.Reader)

	for _, fileHeader := range r.MultipartForm.File["files"] {
		file, err := fileHeader.Open()
		var pe *os.PathError
		if errors.As(err, &pe) {
			// Системная ошибка (права, диск) — 500
			h.RenderError(w, http.StatusInternalServerError, "failed to open uploaded file")
			return
		} else if err != nil {
			// Вероятно, повреждённый файл от клиента — 400
			h.RenderError(w, http.StatusBadRequest, "invalid uploaded file")
			return
		}
		defer file.Close()

		if _, ok := files[fileHeader.Filename]; ok {
			httpAdapterLogger.Warn("duplicate filename in uploaded files. They will be named with the corresponding dates", "filename", fileHeader.Filename)
			files[fmt.Sprintf("%s_%d", fileHeader.Filename, time.Now().UTC().Unix())] = file
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
		h.RenderError(w, models.StatusFromError(err), "failed to create post")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/posts/%d", postID), http.StatusSeeOther)
}

// done
func (h *PostHandler) getPostPageByID(w http.ResponseWriter, postID string, isActual bool) {
	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		h.RenderError(w, http.StatusInternalServerError, "invalid post id")
		return
	}

	templateName := "archive-post.html"
	if isActual {
		templateName = "post.html"
	}

	var post *models.Post
	if isActual {
		post, err = h.postService.GetActualPostByID(id)
	} else {
		post, err = h.postService.GetArchivedPostByID(id)
	}

	if err == models.ErrPostIsArchived {
		h.RenderError(w, models.StatusFromError(err), "Post is not found")
	}

	if err != nil {
		h.RenderError(w, models.StatusFromError(err), "failed to get post")
		return
	}

	postView, err := h.postService.ConstructPostPagePostView(*post)
	if err != nil {
		httpAdapterLogger.Error("failed to construct post view", "error", err)
		h.RenderError(w, models.StatusFromError(err), "failed to construct post view")
		return
	}

	tmpl, err := template.ParseFiles(fmt.Sprint("templates/", templateName))
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		h.RenderError(w, http.StatusInternalServerError, "failed to parse template file")
		return
	}

	data := PostPageData{
		Post:  *postView,
		Error: "",
	}

	var buf bytes.Buffer // чтоб error.html заменял неправильный шаблон
	err = tmpl.Execute(&buf, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err)
		h.RenderError(w, http.StatusInternalServerError, "failed to execute template")
		return
	}
	buf.WriteTo(w)
}

// done
func (h *PostHandler) createCommentToPost(w http.ResponseWriter, r *http.Request, postID string, commentID string) {
	// парсим postID
	id, err := strconv.Atoi(postID)
	if err != nil {
		h.RenderError(w, http.StatusInternalServerError, "invalid post id")
		return
	}

	var addressedTo int
	if commentID == "" {
		// addressedTo = id
		addressedTo = 0
	} else {
		addressedTo, err = strconv.Atoi(commentID)
		if err != nil {
			h.RenderError(w, http.StatusInternalServerError, "invalid comment id")
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
			h.RenderError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		httpAdapterLogger.Error("failed to parse multipart form", "postID", postID, "AddressedTo", addressedTo, "error", err)
		h.RenderError(w, http.StatusInternalServerError, "failed to parse multipart form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	// ===============================================================================

	text := r.FormValue("comment")

	files := make(map[string]io.Reader)

	if text == "" {
		httpAdapterLogger.Warn("empty comment", "postID", postID, "AddressedTo", addressedTo)
		h.RenderError(w, http.StatusBadRequest, "empty comment")
		return
	}

	for _, fileHeader := range r.MultipartForm.File["files"] {
		file, err := fileHeader.Open()
		var pe *os.PathError
		if errors.As(err, &pe) {
			// Системная ошибка (права, диск) — 500
			h.RenderError(w, http.StatusInternalServerError, "failed to open uploaded file")
			return
		} else if err != nil {
			// Вероятно, повреждённый файл от клиента — 400
			h.RenderError(w, http.StatusBadRequest, "invalid uploaded file")
			return
		}
		defer file.Close()

		if _, ok := files[fileHeader.Filename]; ok {
			httpAdapterLogger.Warn("duplicate filename in uploaded files. They will be named with the corresponding dates", "filename", fileHeader.Filename)
			files[fmt.Sprintf("%s_%d", fileHeader.Filename, time.Now().UTC().Unix())] = file
			continue
		}
		files[fileHeader.Filename] = file
	}

	// ===============================================================================
	// reading sessionID
	sessionID, err := h.readSessionIDFromCookie(r)
	if err != nil && errors.Is(err, models.ErrNoSession) {
		httpAdapterLogger.Warn("treating user as anonymous due to missing session cookie", "postID", postID, "error", err)
		sessionID = h.setSessionCookie(w) // set a session cookie with session ID picked by repo or 0 for anonymous users
	}

	if err != nil && !errors.Is(err, models.ErrNoSession) {
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
		h.RenderError(w, models.StatusFromError(err), "failed to create comment")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/posts/%d", id), http.StatusSeeOther)
}

// done
func (h *PostHandler) getArchive(w http.ResponseWriter, r *http.Request, page, limit int) {
	posts, err := h.postService.GetArchivedPosts()
	if err != nil {
		h.RenderError(w, models.StatusFromError(err), "failed to get posts")
		return
	}

	if len(posts) == 0 {
		h.RenderError(w, http.StatusNotFound, "No posts found")
		return
	}

	pages := (len(posts) + limit - 1) / limit
	if page > pages {
		httpAdapterLogger.Warn("requested page exceeds total pages. It will redirect to last page automatically.", "requested_page", page, "total_pages", pages)
		http.Redirect(w, r, fmt.Sprintf("/archive?page=%d&limit=%d", pages, limit), http.StatusMovedPermanently)
	}
	if page < 1 {
		httpAdapterLogger.Warn("requested page is less than 1. It will redirect to first page automatically.", "requested_page", page)
		http.Redirect(w, r, fmt.Sprintf("/archive?page=1&limit=%d", limit), http.StatusMovedPermanently)
	}

	posts = posts[(page-1)*limit : min(page*limit, len(posts))]

	postViews, err := h.postService.ConstructCatalogPostViews(posts)
	if err != nil {
		h.RenderError(w, models.StatusFromError(err), "failed to construct post views")
		return
	}

	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err, "template_name", "archive.html")
		h.RenderError(w, models.StatusFromError(err), "internal server error")
		return
	}

	data := CatalogPageData{
		Posts: postViews,
		Title: "Catalog",
		Error: "",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err, "template_name", "archive.html")
		h.RenderError(w, models.StatusFromError(err), "internal server error")
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

func (h *PostHandler) RenderError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	templateName := "error.html"
	tmpl, err := template.ParseFiles(fmt.Sprint("templates/", templateName))
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, templateName, map[string]interface{}{
		"Code":    code,
		"Message": message,
	})
}
