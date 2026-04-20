package domain

import (
	"1337b04rd/models"
	"errors"
)

type CommentView struct {
	PostID     int
	CommentID  int
	UserName   string
	AvatarLink string //actually it's characterID, but we can easily get avatar link from it
	DataTime   string
	Content    string
	Replies    []CommentView
}

func (h *PostService) GetActualPostByID(id int) (*models.Post, error) {
	post, err := h.posts.GetByID(id, true)
	if err != nil {
		return nil, err
	}
	if post == nil {
		servicesLogger.Warn("post not found or is archived", "id", id)
		return nil, errors.New("post not found or is archived")
	}
	return post, nil
}

func (h *PostService) GetCommentsByPostID(postID int) ([]models.Comments, map[int][]models.Comment, map[int]models.Anon, error) {
	comments, commentsByParent, anonsByCommetns, err := h.comments.GetByPostID(postID)
	if err != nil {
		return nil, err
	}

	return comments, commentsByParent, anonsByCommetns, nil
}

func (h *PostService) ConstructPostPagePostView(post models.Post) (*PostView, error) {
	op, err := h.anons.GetByID(post.AnonID)
	if err != nil {
		return nil, err
	}

	attachments, err := h.attachments.GetByPostID(post.PostID)
	if err != nil {
		return nil, err
	}

	attachmentViews, err := h.constructAttachmentsViews(attachments)
	if err != nil {
		return nil, err
	}

	preview := &AttachmentView{}
	if len(attachments) > 0 {
		preview = &AttachmentView{
			Attachment: attachments[0],
		}
		preview.Link, err = h.fileStorage.GetFileLink(attachments[0].FileKey)
		if err != nil {
			return nil, err
		}
	}

	commentsByParent := map[int][]models.Comment{}
	anonsByComments := map[int]models.Anon{}

	_, commentsByParent, anonsByComments, err = h.GetCommentsByPostID(post.PostID)
	if err != nil {
		return nil, err
	}
	commentsViews := h.buildCommentView(0, commentsByParent, anonsByComments)

	postView := &PostView{
		Post:        post,
		Op:          *op,
		Attachments: attachmentViews,
		Preview:     preview,
		Comments:    commentsViews,
	}
	return postView, nil
}

func (h *PostService) buildCommentView(parentID int, byParent map[int][]models.Comment, anonsByComments map[int]models.Anon) []CommentView {
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
