package domain

import (
	"1337b04rd/models"
)

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
