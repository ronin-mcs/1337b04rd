package httpadapter

import (
	"1337b04rd/models"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var httpAdapterLogger = slog.With("adapter", "http")

const maxUploadSize = 50 << 20  // 50 MB
const maxParseMemory = 10 << 20 // 10 MB

type CatalogPageData struct {
	Posts []models.Post
	Title string
	Error string
}

type PostView struct {
	PostID      int
	Title       string
	TextContent string
	AnonID      int
	CreatedAt   time.Time
}

type CommentView struct {
	PostID     int
	CommentID  int
	UserName   string
	AvatarLink string
	DataTime   string
	Content    string
	Replies    []CommentView
}

type PostPageData struct {
	Post     models.Post
	OP       models.Anon
	Comments []CommentView
	Error    Error
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

	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := CatalogPageData{
		Posts: posts,
		Title: "Catalog",
		Error: nil,
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
		if errors.As(err, *&http.MaxBytesError{}) {
			httpAdapterLogger.Warn("request body too large", "error", err)
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
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
		http.Error(w, err, http.StatusNotFound)
		return
	}

	OP := post.AnonID
	op, err := h.postService.GetOPInfo(OP)
	if err != nil {
		// httpAdapterLogger.Error("failed to get OP info", "post_id", postID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	comments := []CommentView{}
	commentsByParent := map[int][]models.Comment{}
	anonsByComments := map[int]models.Anon{}

	comments, commentsByParent, anonsByComments, err = h.postService.GetCommentsByPostID(post.PostID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	commentViews := h.buildCommentView(0, commentsByParent, anonsByComments)

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := PostPageData{
		Post:     post,
		OP:       op,
		Comments: commentViews,
		Error:    nil,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) buildCommentView(parentID int, byParent map[int][]models.Comment, anonsByComments map[int]models.Anon) []CommentView {
	comments := byParent[parentID]
	commentViews := make([]CommentView, len(comments))
	for i, comment := range comments {
		anon := anonsByComments[comment.CommentID]
		commentViews[i] = CommentView{
			PostID:     comment.PostID,
			CommentID:  comment.CommentID,
			UserName:   anon.AnonName,
			AvatarLink: anon.Avatar,
			DataTime:   comment.CreatedAt.Format("2006-01-02 15:04:05"),
			Content:    comment.TextContent,
			Replies:    h.buildCommentView(comment.CommentID, byParent, anonsByComments),
		}
	}
	return commentViews
}

func (h *PostHandler) createCommentToPost(w http.ResponseWriter, r *http.Request, postID string) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize) // 50 MB

	err := r.ParseMultipartForm(maxParseMemory) // 10 MB
	if err != nil {
		if errors.As(err, *&http.MaxBytesError{}) {
			httpAdapterLogger.Warn("request body too large", "postID", postID, "error", err)
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		}

		httpAdapterLogger.Error("failed to parse multipart form", "postID", postID, "error", err)
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	text := r.FormValue("comment")

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			// файл не прикрепили, это может быть нормально
			file = nil
			fileHeader = nil
		} else {
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
	}
	if file != nil {
		defer file.Close()
	}

	// text - текст коммента
	// file - содержимое файла
	// fileHeader.Filename - имя файла

}

func (h *PostHandler) createCommentToComment(postID, commentID string) {

}

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

	tmpl, err := template.ParseFiles("templates/catalog.html")
	if err != nil {
		httpAdapterLogger.Error("failed to parse template file", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := CatalogPageData{
		Posts: posts,
		Title: "Catalog",
		Error: nil,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		httpAdapterLogger.Error("failed to execute template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
