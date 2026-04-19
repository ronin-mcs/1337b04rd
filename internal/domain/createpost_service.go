package domain

import (
	"1337b04rd/models"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

func (h *PostService) CreatePost(post *models.Post, filename_filedata map[string]io.Reader) (int, error) {
	filekey_filename, err := h.uploadPostFiles(filename_filedata, post.ID)
	if err != nil {
		servicesLogger.Error("failed to upload post files", "postID", post.ID, "error", err)
		return 0, err
	}

	attachements := make([]models.Attachment, 0, len(filekey_filename))
	for fileKey, filename := range filekey_filename {
		contentType, _, err := realContentType(filename_filedata[filename])
		if err != nil {
			return 0, err
		}

		attachements = append(attachements, models.Attachment{
			AtttachmentID: 0,
			PostID:        post.ID,
			CommnentID:    nil,
			FileKey:       fileKey,
			OriginalName:  filename,
			ContentType:   contentType,
		})
	}

	for _, att := range attachements {
		err = h.attachments.Create(&att)
		if err != nil {
			return 0, err
		}
	}

	characterID, err := h.avatarStorage.GetRandomCharacterID()
	if err != nil {
		return 0, err
	}
	anonName := fmt.Sprintf("Anon%d", characterID)

	anon := &models.Anon{
		AnonID:   0,
		PostID:   post.ID,
		Avatar:   characterID, // will be set in repo
		AnonName: anonName,    // will be set in repo
	}
	err = h.anons.Create(anon)
	if err != nil {
		return 0, err
	}

	post.AnonID = anon.AnonID

	err = h.posts.Create(post)
	if err != nil {
		return 0, err
	}

	return post.PostID, nil
}

func (h *PostService) uploadPostFiles(files map[string]io.Reader, postID int) (map[string]string, error) {
	prefix := fmt.Sprintf("posts/%d", postID)
	return h.uploadFiles(files, prefix)
}

func (h *PostService) uploadCommentFiles(files map[string]io.Reader, postID, commentID int) (map[string]string, error) {
	prefix := fmt.Sprintf("comments/%d/%d", postID, commentID)
	return h.uploadFiles(files, prefix)
}

func (h *PostService) uploadFiles(files map[string]io.Reader, prefix string) (map[string]string, error) {
	if len(files) == 0 {
		servicesLogger.Warn(fmt.Sprintf("no files to upload for %s", prefix))
		return nil, nil
	}
	if len(files) > 10 {
		servicesLogger.Warn(fmt.Sprint("too many files to upload%s", prefix), "fileCount", len(files))
		return nil, errors.New(fmt.Sprintf("too many files to upload for %s", prefix))
	}

	filekey_filename := make(map[string]string, len(files))
	for filename, filedata := range files {
		ext := filepath.Ext(filename)
		fileKey, err := h.generateFileKey(prefix, ext)
		if err != nil {
			servicesLogger.Error("failed to generate file key", "prefix", prefix, "filename", filename, "error", err)
			return nil, err
		}
		if _, ok := filekey_filename[fileKey]; ok {
			servicesLogger.Warn("generated duplicate file key, retrying", "prefix", prefix, "filename", filename, "fileKey", fileKey)
			return nil, errors.New("generated duplicate file key, please try again")
		}

		contentType, reader, err := realContentType(filedata)
		if err != nil {
			servicesLogger.Error("failed to determine real content type of file", "prefix", prefix, "filename", filename, "error", err)
			return nil, fmt.Errorf("failed to determine content type: %w", err)
		}

		filekey_filename[fileKey] = filename
		err = h.FileStorage.SaveFile(fileKey, reader, contentType)
		if err != nil {
			return nil, err
		}
	}
	return filekey_filename, nil

}

func (h *PostService) generateFileKey(prefix, ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	name := hex.EncodeToString(b)
	return prefix + "/" + name + ext, nil
}

func realContentType(filedata io.Reader) (string, io.Reader, error) {
	buf := make([]byte, 512)
	n, err := filedata.Read(buf)
	if err != nil && err != io.EOF {
		servicesLogger.Error("failed to read filedata for content type detection", "error", err)
		return "", filedata, errors.New("failed to read filedata header")
	}
	contentType := http.DetectContentType(buf[:n])
	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), filedata)
	return contentType, fullReader, nil
}
