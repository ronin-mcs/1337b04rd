package domain

import (
	"1337b04rd/models"
)

type PostView struct {
	Post        models.Post
	Op          models.Anon
	Attachments []AttachmentView
	Preview     *AttachmentView
	Comments    []CommentView
}

type AttachmentView struct {
	Attachment models.Attachment
	Link       string
}

func (h *PostService) ConstructCatalogPostViews(posts []models.Post) ([]*PostView, error) {
	postViews := make([]*PostView, 0, len(posts))
	for _, post := range posts {
		op, err := h.anons.GetByID(post.AnonID)
		if err != nil {
			return nil, err
		}

		attachments, err := h.attachments.GetByPostID(post.PostID)
		if err != nil {
			return nil, err
		}

		preview := &AttachmentView{}
		if len(attachments) > 0 {
			preview := &AttachmentView{
				Attachment: attachments[0],
			}
			preview.Link, err = h.fileStorage.GetFileLink(attachments[0].FileKey)
		}

		// for catalog page we can show only first attachment as preview, so we can construct full attachments views only for post page
		var attachmentViews []AttachmentView

		// attachmentViews, err := h.constructAttachmentsViews(attachments)
		// if err != nil {
		// 	return nil, err
		// }

		// we don't need comments for catalog page, so we can pass empty slice
		postView := &PostView{
			Post:        post,
			Op:          *op,
			Attachments: attachmentViews,
			Preview:     preview,
			Comments:    []CommentView{},
		}
		postViews = append(postViews, postView)
	}
	return postViews, nil
}

func (h *PostService) constructAttachmentsViews(attachments []models.Attachment) ([]AttachmentsView, error) {
	attachmentsViews := make([]AttachmentView, 0, len(attachments))
	for _, att := range attachments {
		link, err := h.fileStorage.GetFileLink(att.FileKey)
		if err != nil {
			return nil, err
		}

		attachmentsViews = append(attachmentsViews, AttachmentView{
			Attachment: att,
			Link:       link,
		})
	}
	return attachmentsViews, nil
}

func (h *PostService) GetActualPosts() ([]models.Post, error) {

	posts, err := h.posts.GetAll(true)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (h *PostService) GetArchivedPosts() ([]models.Post, error) {
	posts, err := h.posts.GetAll(false)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
