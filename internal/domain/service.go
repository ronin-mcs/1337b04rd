package domain

import (
	"1337b04rd/models"
	"errors"
)

servicesLogger = slog.With("domain", "services")

type PostService struct {
	AvatarStorage
	FileStorage
	posts    PostRepository
	comments CommentRepository
	anons    AnonRepository
	sessions SessionRepository
}

func NewPostService(avatarStorage AvatarStorage, fileStorage FileStorage, posts PostRepository, comments CommentRepository, anons AnonRepository, sessions SessionRepository) *PostService {
	return &PostService{
		AvatarStorage: avatarStorage,
		FileStorage:   fileStorage,
		posts:         posts,
		comments:      comments,
		anons:         anons,
		sessions:      sessions,
	}
}

func (h *PostService) GetActualPosts() ([]models.Post, error) {

	posts, err := h.posts.GetAll(true)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (h *PostService) GetActualPostByID(id int) (*models.Post, error) {
	post, err := h.posts.GetByID(id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		servicelogger.Warn("post not found or is archived", "id", id)
		return nil, errors.New("post not found or is archived")
	}
	return post, nil
}

func (h *PostService) GetOPInfo(AnonID int) (*models.Anon, error) {
	op, err := h.anons.GetByID(AnonID)
	if err != nil {
		return nil, err
	}
	if op == nil {
		servicelogger.Warn("OP not found", "AnonID", AnonID)
		return nil, errors.New("OP not found")
	}
	return op, nil
}

func (h *PostService) GetCommentsByPostID(postID int) ([]models.Comments, map[int][]models.Comment, map[int]models.Anon, error) {
	comments , commentsByParent, anonsByCommetns, err := h.comments.GetByPostID(postID)
	if err != nil {
		return nil, err
	}

	return comments, commentsByParent, anonsByCommetns, nil
}



// бизнес логика:

// getPosts():
// получаем объекты постов неархивных
// получаем бинарники файлов
// собираем всё в PostService структуру и возвращаем

// createPost()
// из данных отправленных через фронт из post запроса собираем объект пост
// отправляем его в db
// отправляем в s3

// getPostByID()
// получаем из объект поста
// объект комментариев (возможно возвращаем json, всё зависит от того как мы будем парсить их в handler)

// createCommentToPost()
// из запроса собираем объект отправляем в дб

// getArchive()
// получаем объекты постов архивных
// получаем бинарники файлов
// собираем всё в PostService структуру и возвращаем
