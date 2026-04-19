package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
	"fmt"
	"log/slog"
)

var attachmentsRepoLogger = slog.With("adapter", "attachments_repo")

type AttachmentsRepo struct {
	db *sql.DB
}

func NewAttachmentsRepo(db *sql.DB) *AttachmentsRepo {
	return &AttachmentsRepo{db: db}
}

func (h *AttachmentsRepo) Create(attachment *models.Attachment) error {
	query := `
		INSERT INTO attachments (post_id, comment_id, file_key, original_name, content_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING attachment_id
	`

	var commentID sql.NullInt64
	if attachment.CommentID != nil {
		commentID = sql.NullInt64{
			Int64: int64(*attachment.CommentID),
			Valid: true,
		}
	}

	err := h.db.QueryRow(
		query,
		attachment.PostID,
		commentID,
		attachment.FileKey,
		attachment.OriginalName,
		attachment.ContentType,
	).Scan(&attachment.AttachmentID)
	if err != nil {
		attachmentsRepoLogger.Error(
			"create failed",
			"post_id", attachment.PostID,
			"comment_id", attachment.CommentID,
			"file_key", attachment.FileKey,
			"error", err,
		)
		return err
	}

	return nil
}

func (h *AttachmentsRepo) GetByPostID(postID int) ([]models.Attachment, error) {
	query := `
		SELECT attachment_id, post_id, comment_id, file_key, original_name, content_type
		FROM attachments
		WHERE post_id = $1 AND comment_id IS NULL
		ORDER BY attachment_id ASC
	`

	rows, err := h.db.Query(query, postID)
	if err != nil {
		attachmentsRepoLogger.Error("get by post id failed", "post_id", postID, "error", err)
		return nil, err
	}
	defer rows.Close()

	attachments, err := scanAttachments(rows)
	if err != nil {
		attachmentsRepoLogger.Error("scan attachments failed", "post_id", postID, "error", err)
		return nil, err
	}

	return attachments, nil
}

func (h *AttachmentsRepo) GetByCommentID(commentID int) ([]models.Attachment, error) {
	query := `
		SELECT attachment_id, post_id, comment_id, file_key, original_name, content_type
		FROM attachments
		WHERE comment_id = $1
		ORDER BY attachment_id ASC
	`

	rows, err := h.db.Query(query, commentID)
	if err != nil {
		attachmentsRepoLogger.Error("get by comment id failed", "comment_id", commentID, "error", err)
		return nil, err
	}
	defer rows.Close()

	attachments, err := scanAttachments(rows)
	if err != nil {
		attachmentsRepoLogger.Error("scan attachments failed", "comment_id", commentID, "error", err)
		return nil, err
	}

	return attachments, nil
}

func (h *AttachmentsRepo) DeleteByFileKey(fileKey string) error {
	query := `
		DELETE FROM attachments
		WHERE file_key = $1
	`

	result, err := h.db.Exec(query, fileKey)
	if err != nil {
		attachmentsRepoLogger.Error("delete by file key failed", "file_key", fileKey, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		attachmentsRepoLogger.Error("check deleted rows failed", "file_key", fileKey, "error", err)
		return err
	}
	if rowsAffected == 0 {
		err := fmt.Errorf("attachment not found: file_key=%s", fileKey)
		attachmentsRepoLogger.Error("delete by file key failed", "error", err)
		return err
	}

	return nil
}

func scanAttachments(rows *sql.Rows) ([]models.Attachment, error) {
	attachments := []models.Attachment{}

	for rows.Next() {
		var attachment models.Attachment
		var commentID sql.NullInt64

		err := rows.Scan(
			&attachment.AttachmentID,
			&attachment.PostID,
			&commentID,
			&attachment.FileKey,
			&attachment.OriginalName,
			&attachment.ContentType,
		)
		if err != nil {
			return nil, err
		}

		if commentID.Valid {
			id := int(commentID.Int64)
			attachment.CommentID = &id
		}

		attachments = append(attachments, attachment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}
