package domain

import (
	"1337b04rd/models"
	"errors"
)

func (h *PostService) GetActualPostByID(id int) (*models.Post, error) {
	post, err := h.posts.GetByID(id, true)
	if err != nil {
		return nil, err
	}
	if post == nil {
		servicelogger.Warn("post not found or is archived", "id", id)
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
