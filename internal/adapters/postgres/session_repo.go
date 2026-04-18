package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
)

var sessionlogger = slog.With("adapter", "postgres", "repository", "sessions")

type PGSesssionsRepository struct {
	db *sql.DB
}

func NewPGSessionsRepository(db *sql.DB) *PGSesssionsRepository {
	return &PGSesssionsRepository{db: db}
}

func (h *PGSesssionsRepository) Create(session *models.Session) error {
	query := `
		INSERT INTO sessions (session_history) 
		VALUES ($1)
		RETURNING session_id
	`

	session_history, err := json.Marshal(session.Sessions)
	if err != nil {
		sessionlogger.Error("Map marshal failed while creating session in db", "error", err)
		return err
	}

	err = h.db.QueryRow(query, session_history).Scan(&session.session_id)
	sessionlogger.Error("Create session failed", "error", err)
	return err
}

func (h *PGSesssionsRepository) GetByID(id int) (*models.Session, error) {
	query := `
		SELECT (session_id, session_history) FROM sessions 
	`
	var session models.Session
	session_history_json := ""
	err := h.db.QueryRow(query).Scan(&session, session_history_json)
	if err != nil {
		sessionlogger.Error("Get session by ID failed", "error", err)
		return nil, err
	}
	err = json.Unmarshal([]byte(session_history_json), &session.Sessions)
	if err != nil {
		sessionlogger.Error("Map unmarshal failed while getting session from db", "error", err)
		return nil, err
	}
	return &session, nil
}

func (h *PGSesssionsRepository) Delete(id int) error {
	query := `
		DELETE FROM sessions WHERE session_id = $1
	`
	result, err := h.db.Exec(query, id)
	if err != nil {
		sessionlogger.Error("Delete session failed", "error", err)
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
