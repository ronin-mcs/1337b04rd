package httpadapter

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
)

var httpAdapterLogger = slog.With("adapter", "http")

type CatalogPageData struct {
	Posts []models.Post
	Title string
	Error Error
}

type CommentView struct {
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

func (h *PostHandler) getCreatePost(w http.ResponseWriter, r *http.Request) {

}

func (h *PostHandler) createPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value

	// дальше передаешь sessionID в service
	_ = sessionID
	// ...
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



func (h *PostHandler) createCommentToPost(postID string) {
	
}

func (h *PostHandler) createCommentToComment(postID, commentID string) {

}

func (h *PostHandler) getArchive(page, limit int) {

}
