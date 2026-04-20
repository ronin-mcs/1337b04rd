package domain

import "log/slog"

var servicesLogger = slog.With("domain", "services")

type PostService struct {
	avatarStorage AvatarStorage
	fileStorage   FileStorage
	posts         PostRepository
	comments      CommentRepository
	anons         AnonRepository
	sessions      SessionRepository
	attachments   AttachmentRepository
}

func NewPostService(avatarStorage AvatarStorage, fileStorage FileStorage, posts PostRepository, comments CommentRepository, anons AnonRepository, sessions SessionRepository, attachments AttachmentRepository) *PostService {
	return &PostService{
		avatarStorage: avatarStorage,
		fileStorage:   fileStorage,
		posts:         posts,
		comments:      comments,
		anons:         anons,
		sessions:      sessions,
		attachments:   attachments,
	}
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
// получаем postID и возвращаем его на фронт
// генерируем anonID и генерирем рандом имя и аватарку для анона, сохраняем в бд Anons
// как мы его генерируем?

// createCommentToPost()
// из запроса собираем объект отправляем в дб
// создаём нового анона добавляем в дб
// как распределить аватарки?
//вытащить список всех аватаров которые есть в Anon вместе с их количеством указав нужный postID,
// чекать есть ли подобная аватрка, задать max + 1
// если все аватарки уже заняты задать тот у которого меньше всего повторений

// заполняем
// 	CommentID   int // полуаем от дб
// 	PostID      int // получаем из фронта (наверное)
// 	AddressedTo int // получаем так же с фронта
// 	TextContent string // front
// 	CreatedAt   time.Time //генерируется сразу

// 	AnonID      int //

// как?
// создаём anon в дб возвращаем его anonID

// getArchive()
// получаем объекты постов архивных
// получаем бинарники файлов
// собираем всё в PostService структуру и возвращаем
