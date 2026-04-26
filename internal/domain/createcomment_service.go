package domain

import (
	"1337b04rd/models"
	"errors"
	"fmt"
	"io"
	"math/rand"
)

func (h *PostService) CreateComment(comment *models.Comment, filename_filedata map[string]io.Reader, sessiondID int) error {
	anonID, err := h.retrieveAnonIDForComment(comment.PostID, sessiondID)
	if err != nil {
		return err
	}

	// anon, err := h.anons.GetByID(anonID)
	// if err != nil {
	// 	return err
	// }
	comment.AnonID = anonID
	err = h.comments.Create(comment)
	if err != nil {
		return err
	}

	// =============================================================================

	postID := comment.PostID
	commentID := comment.CommentID

	filekey_filename, err := h.uploadCommentFiles(filename_filedata, postID, commentID)
	if err != nil {
		servicesLogger.Error("failed to upload comment files", "postID", postID, "commentID", commentID, "error", err)
		return err
	}

	attachements := make([]models.Attachment, 0, len(filekey_filename))
	for fileKey, filename := range filekey_filename {
		contentType, _, err := realContentType(filename_filedata[filename])
		if err != nil {
			return err
		}

		attachements = append(attachements, models.Attachment{
			AttachmentID: 0, // will be set in repo
			PostID:       postID,
			CommentID:    &commentID, // this is for post attachments, so commentID is nil
			FileKey:      fileKey,
			OriginalName: filename,
			ContentType:  contentType,
		})
	}

	for _, att := range attachements {
		err = h.attachments.Create(&att)
		if err != nil {
			return err
		}
	}

	// =============================================================================
	// update last_updated_at status for post
	return h.posts.UpdateStatus(postID)
}

func (h *PostService) retrieveAnonIDForComment(postID int, sessionID int) (int, error) {
	// anonymous
	if sessionID == 0 {
		anon, err := h.constructNewAnon(postID)
		if err != nil {
			return 0, err
		}

		return anon.AnonID, nil
	}

	session, err := h.sessions.GetByID(sessionID)

	if err != nil {
		return 0, err
	}

	anon_id, ok := session.Sessions[postID]
	if !ok {
		servicesLogger.Warn("session does not contain anon_id for this post", "postID", postID, "sessionID", sessionID)

		anon, err := h.constructNewAnon(postID)
		if err != nil {
			return 0, err
		}

		err = h.uploadSessionID(postID, sessionID, anon.AnonID)
		if err != nil {
			return 0, err
		}

		return anon.AnonID, nil
	}
	return anon_id, nil
}

func (h *PostService) constructNewAnon(postID int) (*models.Anon, error) {
	avatar, err := h.getAvatarForNewAnon(postID)
	if err != nil {
		return nil, err
	}
	// генерируем anonName через characterID и рандомные символы
	anonname := fmt.Sprintf("Anon%s_%d", avatar, rand.Intn(1000))

	anon := &models.Anon{
		AnonID:   0,
		PostID:   postID,
		Avatar:   avatar, // will be set in repo
		AnonName: anonname,
	}

	// create new anon in db
	err = h.anons.Create(anon)
	if err != nil {
		return nil, err
	}

	return anon, nil
}

func (h *PostService) getAvatarForNewAnon(postID int) (string, error) {
	// дадим аватарку
	// как?

	// вытащим список всех anon'Ов
	// или вытащить из sql через group by количество anon для каждой аватарки в виде map[string]int
	// получаем список всех characters из rickandmorty,
	//
	// если длина map такая же как у этого списка
	// выбираем рандомный
	// иначе
	// проходимся по списку и ищем того которого нет в map
	avatar_count, err := h.anons.GetAvatarCountByPostID(postID) // получаем количество анонов для каждой аватарки в этом посте в виде map[string]int
	if err != nil {
		return "", err
	}

	AllCharacter := h.avatarStorage.GetAllCharacterIDs() // получаем список всех персонажей из рик и морти
	if len(avatar_count) == len(AllCharacter) {
		// если количество анонов равно количеству персонажей, то выбираем рандомного из всех персонажей
		avatar, err := h.avatarStorage.GetRandomCharacterID()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", avatar), nil
	}

	for _, character := range AllCharacter {
		if _, exists := avatar_count[fmt.Sprintf("%d", character)]; !exists {
			// если персонажа нет в map, то выбираем его
			return fmt.Sprintf("%d", character), nil
		}
	}

	return "", errors.New("no available avatars")
}

func (h *PostService) uploadSessionID(postID int, sessionID int, anonID int) error {
	session, err := h.sessions.GetByID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		session = &models.Session{SessionID: sessionID}
		err = h.sessions.Create(session)
		if err != nil {
			return err
		}
	}
	session.Sessions[postID] = anonID               // сохраняем в сессии id анона для этого поста
	return h.sessions.UpdateSessionHistory(session) // обновляем сессию в репозитории
}
