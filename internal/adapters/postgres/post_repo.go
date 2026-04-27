package dbadapter

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"1337b04rd/models"
)

var postsLogger = slog.With("adapter", "postgres", "repository", "posts")

var actualWhereClause = `
		(last_updated_at + INTERVAL '15 minutes' > (now() AT TIME ZONE 'UTC')
		AND (
			created_at + INTERVAL '10 minutes' > (now() AT TIME ZONE 'UTC')
			OR EXISTS (
				SELECT 1 FROM comments WHERE comments.post_id = posts.post_id
			)
		))
	`

var archivedWhereClause = `
		(last_updated_at + INTERVAL '15 minutes' <= (now() AT TIME ZONE 'UTC')
		OR (
			created_at + INTERVAL '10 minutes' <= (now() AT TIME ZONE 'UTC')
			AND NOT EXISTS (
				SELECT 1 FROM comments WHERE comments.post_id = posts.post_id
			)
		))
	`

type PGPostsRepository struct {
	db *sql.DB
}

func NewPGPostsRepository(db *sql.DB) *PGPostsRepository {
	return &PGPostsRepository{db: db}
}

func (h *PGPostsRepository) Create(post *models.Post) error {
	now := time.Now().UTC()
	post.CreatedAt = now
	post.LastUpdatedAt = now

	query := `
		INSERT INTO posts (title, text_content, OP_id, created_at, last_updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING post_id, created_at, last_updated_at
	`

	err := h.db.QueryRow(
		query,
		post.Title,
		post.TextContent,
		sql.NullInt64{Int64: int64(post.AnonID), Valid: post.AnonID > 0},
		post.CreatedAt,
		post.LastUpdatedAt,
	).Scan(&post.PostID, &post.CreatedAt, &post.LastUpdatedAt)
	if err != nil {
		postsLogger.Error("create failed", "error", err)
		return err
	}

	return nil
}

func (h *PGPostsRepository) CreateWithOP(post *models.Post, anon *models.Anon) error {
	tx, err := h.db.Begin()
	if err != nil {
		postsLogger.Error("begin create with OP transaction failed", "error", err)
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	post.CreatedAt = now
	post.LastUpdatedAt = now

	createPostQuery := `
		INSERT INTO posts (title, text_content, OP_id, created_at, last_updated_at)
		VALUES ($1, $2, NULL, $3, $4)
		RETURNING post_id, created_at, last_updated_at
	`
	err = tx.QueryRow(
		createPostQuery,
		post.Title,
		post.TextContent,
		post.CreatedAt,
		post.LastUpdatedAt,
	).Scan(&post.PostID, &post.CreatedAt, &post.LastUpdatedAt)
	if err != nil {
		postsLogger.Error("create post in transaction failed", "error", err)
		return err
	}

	anon.PostID = post.PostID
	createAnonQuery := `
		INSERT INTO Anons (name, post_id, avatar)
		VALUES ($1, $2, $3)
		RETURNING anon_id
	`
	err = tx.QueryRow(createAnonQuery, anon.AnonName, anon.PostID, anon.Avatar).Scan(&anon.AnonID)
	if err != nil {
		postsLogger.Error("create OP anon in transaction failed", "post_id", post.PostID, "error", err)
		return err
	}

	assignOPQuery := `
		UPDATE posts
		SET OP_id = $1
		WHERE post_id = $2
	`
	result, err := tx.Exec(assignOPQuery, anon.AnonID, post.PostID)
	if err != nil {
		postsLogger.Error("assign OP in transaction failed", "post_id", post.PostID, "anon_id", anon.AnonID, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		postsLogger.Error("check assigned OP rows failed", "post_id", post.PostID, "error", err)
		return err
	}
	if rowsAffected == 0 {
		err := models.ErrNotFound
		postsLogger.Error("assign OP in transaction failed", "error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		postsLogger.Error("commit create with OP transaction failed", "post_id", post.PostID, "anon_id", anon.AnonID, "error", err)
		return err
	}

	post.AnonID = anon.AnonID
	return nil
}

func (h *PGPostsRepository) GetByID(id int, IsActual bool) (*models.Post, error) {
	query := `
		SELECT post_id, title, text_content, OP_id, created_at, last_updated_at
		FROM posts
		WHERE post_id = $1 AND 
	`

	if IsActual {
		query += actualWhereClause
	} else {
		query += archivedWhereClause
	}

	var post models.Post
	var opID sql.NullInt64
	err := h.db.QueryRow(query, id).Scan(
		&post.PostID,
		&post.Title,
		&post.TextContent,
		&opID,
		&post.CreatedAt,
		&post.LastUpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			postsLogger.Warn("no posts found", "id", id)
			return nil, nil
		}
		postsLogger.Error("get by id failed", "id", id, "error", err)
		return nil, err
	}

	if opID.Valid {
		post.AnonID = int(opID.Int64)
	}

	return &post, nil
}

func (h *PGPostsRepository) GetAll(IsActual bool) ([]models.Post, error) {
	query := `
		SELECT post_id, title, text_content, OP_id, created_at, last_updated_at
		FROM posts
		WHERE 
	`

	if IsActual {
		query += actualWhereClause
	} else {
		query += archivedWhereClause
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query)
	if err != nil {
		postsLogger.Error("get all failed", "error", err)
		return nil, err // 500
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var post models.Post
		var opID sql.NullInt64
		err := rows.Scan(
			&post.PostID,
			&post.Title,
			&post.TextContent,
			&opID,
			&post.CreatedAt,
			&post.LastUpdatedAt,
		)
		if err != nil {
			postsLogger.Error("scan row failed", "error", err)
			return nil, err
		}

		if opID.Valid {
			post.AnonID = int(opID.Int64)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		postsLogger.Error("iterate rows failed", "error", err)
		return nil, err
	}

	if len(posts) == 0 {
		postsLogger.Warn("no posts found")
	}

	return posts, nil
}

func (h *PGPostsRepository) UpdateStatus(id int) error {
	query := `
		UPDATE posts
		SET last_updated_at = $1
		WHERE post_id = $2
	`

	result, err := h.db.Exec(query, time.Now().UTC(), id)
	if err != nil {
		postsLogger.Error("update status failed", "id", id, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		postsLogger.Error("check updated rows failed", "id", id, "error", err)
		return err
	}
	if rowsAffected == 0 {
		err := models.ErrNotFound
		postsLogger.Error("update status failed", "error", err)
		return err
	}

	return nil
}

func (h *PGPostsRepository) Delete(id int) error {
	query := `
		DELETE FROM posts
		WHERE post_id = $1
	`

	result, err := h.db.Exec(query, id)
	if err != nil {
		postsLogger.Error("delete failed", "id", id, "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		postsLogger.Error("check deleted rows failed", "id", id, "error", err)
		return err
	}
	if rowsAffected == 0 {
		err := fmt.Errorf("post not found: id=%d", id)
		postsLogger.Error("delete failed", "error", err)
		return err
	}

	return nil
}

func (h *PGPostsRepository) AssignOP(post_id, anon_id int) error {
	query := `
		UPDATE Posts
		SET OP_id = $1
		WHERE post_id = $2
	`
	_, err := h.db.Exec(query, anon_id, post_id)
	if err != nil {
		anonlogger.Error("assign OP failed", "error", err)
		return err
	}
	return nil
}
