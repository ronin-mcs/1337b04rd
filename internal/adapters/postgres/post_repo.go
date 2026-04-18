package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

var postsLogger = slog.With("adapter", "postgres", "repository", "posts")

actualWhereClause := `
	NOT(
		last_updated_at + INTERVAL '15 minutes' <= now() 
		OR (
			created_at + INTERVAL '10 minutes' <= now() 
			AND (
				SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.post_id) >= 0
			)
		) 
`

archivedWhereClause := `
	(
		last_updated_at + INTERVAL '15 minutes' <= now() 
		OR (
			created_at + INTERVAL '10 minutes' <= now() 
			AND (
				SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.post_id) >= 0
			)
		) 
`

type PGPostsRespository struct {
	db *sql.DB
}

func NewPGPostsRepository(db *sql.DB) *PGPostsRespository {
	return &PGPostsRespository{db: db}
}

func (h *PGPostsRespository) Create(post *models.Post) error {
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
		post.AnonID,
		post.CreatedAt,
		post.LastUpdatedAt,
	).Scan(&post.PostID, &post.CreatedAt, &post.LastUpdatedAt)
	if err != nil {
		postsLogger.Error("create failed", "error", err)
		return err
	}

	return nil
}

func (h *PGPostsRespository) GetByID(id int, IsActual bool) (*models.Post, error) {

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
	err := h.db.QueryRow(query, id).Scan(
		&post.PostID,
		&post.Title,
		&post.TextContent,
		&post.AnonID,
		&post.CreatedAt,
		&post.LastUpdatedAt,
	)
	if err != nil {
		postsLogger.Error("get by id failed", "id", id, "error", err)
		return nil, err
	}
	return &post, nil
}

func (h *PGPostsRespository) GetAll(IsActual bool) ([]models.Post, error) {
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
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.PostID,
			&post.Title,
			&post.TextContent,
			&post.AnonID,
			&post.CreatedAt,
			&post.LastUpdatedAt,
		)
		if err != nil {
			postsLogger.Error("scan row failed", "error", err)
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		postsLogger.Error("iterate rows failed", "error", err)
		return nil, err
	}

	return posts, nil
}

func (h *PGPostsRespository) UpdateStatus(id int) error {
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
		err := fmt.Errorf("post not found: id=%d", id)
		postsLogger.Error("update status failed", "error", err)
		return err
	}

	return nil
}

func (h *PGPostsRespository) Delete(id int) error {
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
