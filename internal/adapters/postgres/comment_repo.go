package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
	"fmt"
	"log/slog"
)

var commentsLogger = slog.With("adapter", "postgres", "repository", "comments")

type PGCommentsRepository struct {
	db *sql.DB
}

func NewPGCommentsRepository(db *sql.DB) *PGCommentsRepository {
	return &PGCommentsRepository{db: db}
}

func (h *PGCommentsRepository) Create(comment *models.Comment) error {
	query := `
		INSERT INTO comments (post_id, addressed_to, text_content, anon_id)
		VALUES ($1, $2, $3, $4)
		RETURNING comment_id, created_at
	`

	err := h.db.QueryRow(
		query,
		comment.PostID,
		comment.AddressedTo,
		comment.TextContent,
		comment.AnonID,
	).Scan(&comment.CommentID, &comment.CreatedAt)
	if err != nil {
		commentsLogger.Error("create failed", "post_id", comment.PostID, "addressed_to", comment.AddressedTo, "error", err)
		return err
	}

	return nil
}

func (h *PGCommentsRepository) GetByID(id int) (*models.Comment, error) {
	query := `
		SELECT comment_id, post_id, addressed_to, text_content, anon_id, created_at
		FROM comments
		WHERE comment_id = $1
	`

	var comment models.Comment
	err := h.db.QueryRow(query, id).Scan(
		&comment.CommentID,
		&comment.PostID,
		&comment.AddressedTo,
		&comment.TextContent,
		&comment.AnonID,
		&comment.CreatedAt,
	)
	if err != nil {
		commentsLogger.Error("get by id failed", "id", id, "error", err)
		return nil, err
	}

	return &comment, nil
}

func (h *PGCommentsRepository) GetAll() ([]models.Comment, error) {
	query := `
		SELECT comment_id, post_id, addressed_to, text_content, anon_id, created_at
		FROM comments
		ORDER BY comment_id ASC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		commentsLogger.Error("get all failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.CommentID,
			&comment.PostID,
			&comment.AddressedTo,
			&comment.TextContent,
			&comment.AnonID,
			&comment.CreatedAt,
		)
		if err != nil {
			commentsLogger.Error("scan row failed", "error", err)
			return nil, err
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		commentsLogger.Error("iterate rows failed", "error", err)
		return nil, err
	}

	return comments, nil
}

func (h *PGCommentsRepository) Delete(id int) error {
	query := `
		DELETE FROM comments
		WHERE comment_id = $1
	`

	result, err := h.db.Exec(query, id)
	if err != nil {
		commentsLogger.Error("delete failed", "id", id, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		commentsLogger.Error("check deleted rows failed", "id", id, "error", err)
		return err
	}
	if rowsAffected == 0 {
		err := fmt.Errorf("comment not found: id=%d", id)
		commentsLogger.Error("delete failed", "error", err)
		return err
	}

	return nil
}

// GetByPostID retrieves all comments for a given post ID, along with maps of comments by their parent comment and the corresponding anon information for each comment.
func (h *PGCommentsRepository) GetByPostID(postID int) ([]models.Comment, map[int][]models.Comment, map[int]models.Anon, error) {
	query := `
		SELECT
			c.comment_id,
			c.post_id,
			c.addressed_to,
			c.text_content,
			c.anon_id,
			c.created_at,
			a.anon_id,
			a.post_id,
			a.avatar,
			a.name
		FROM comments c
		JOIN anons a ON a.anon_id = c.anon_id
		WHERE c.post_id = $1
		ORDER BY c.addressed_to ASC, c.created_at ASC
	`

	rows, err := h.db.Query(query, postID)
	if err != nil {
		commentsLogger.Error("get by post id failed", "post_id", postID, "error", err)
		return nil, nil, nil, err
	}
	defer rows.Close()

	comments := []models.Comment{}
	commentsByParent := make(map[int][]models.Comment)
	anonsByCommentID := make(map[int]models.Anon)
	for rows.Next() {
		var comment models.Comment
		var anon models.Anon
		err := rows.Scan(
			&comment.CommentID,
			&comment.PostID,
			&comment.AddressedTo,
			&comment.TextContent,
			&comment.AnonID,
			&comment.CreatedAt,
			&anon.AnonID,
			&anon.PostID,
			&anon.Avatar,
			&anon.AnonName,
		)
		if err != nil {
			commentsLogger.Error("scan row failed", "post_id", postID, "error", err)
			return nil, nil, nil, err
		}

		comments = append(comments, comment)
		commentsByParent[comment.AddressedTo] = append(commentsByParent[comment.AddressedTo], comment)
		anonsByCommentID[comment.CommentID] = anon
	}

	if err := rows.Err(); err != nil {
		commentsLogger.Error("iterate rows failed", "post_id", postID, "error", err)
		return nil, nil, nil, err
	}

	return comments, commentsByParent, anonsByCommentID, nil
}
