package domain

import "1337b04rd/models"

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
	SaveAvatar(avatarName string, avatarData []byte) error
	GetAvatarLink(avatarName string) (string, error)
	DeleteAvatar(avatarName string) error
}

type FileStorage interface {
	SaveFile(fileName string, fileData []byte) error
	GetFileLink(fileName string) (string, error)
	DeleteFile(fileName string) error
}