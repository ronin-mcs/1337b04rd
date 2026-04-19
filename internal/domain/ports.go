package domain

import (
	"1337b04rd/models"
	"io"
)

type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id int, IsActual bool) (*models.Post, error)
	GetAll(IsActual bool) ([]models.Post, error)
	UpdateStatus(id int) error
	Delete(id int) error
}

type CommentRepository interface {
	Create(comment *models.Comment) error
	GetByID(id int) (*models.Comment, error)
	GetByPostID(postID int) ([]models.Comment, map[int][]models.Comment, map[int]models.Anon, error)
	GetAll() ([]models.Comment, error)
	Delete(id int) error
}

type AnonRepository interface {
	Create(anon *models.Anon) error
	GetByID(id int) (*models.Anon, error)
	GetAll() ([]models.Anon, error)
	Delete(id int) error
}

type SessionRepository interface {
	Create(session *models.Session) error
	GetByID(id int) (*models.Session, error)
	GetAll() ([]models.Session, error)
	Delete(id int) error
}

type AvatarStorage interface {
	GetRandomCharacterID() (int, error)
	GetAvatar(characterID int) (io.ReadCloser, string, error)
}

type FileStorage interface {
	SaveFile(fileKey string, fileData io.Reader, contentType string) error
	GetFileLink(fileKey string) (string, error)
	DeleteFile(fileKey string) error
}

type AttachmentRepository interface {
	Create(attachment *models.Attachment) error
	GetByPostID(postID int) ([]models.Attachment, error)
	GetByCommentID(commentID int) ([]models.Attachment, error)
	DeleteByFileKey(fileKey string) error
}
