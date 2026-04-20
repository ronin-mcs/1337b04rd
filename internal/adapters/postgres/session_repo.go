package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
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
		INSERT INTO sessions (session_history, expires_at) 
		VALUES ($1, $2)
		RETURNING session_id
	`

	sessionHistory, err := json.Marshal(session.Sessions)
	if err != nil {
		sessionlogger.Error("Map marshal failed while creating session in db", "error", err)
		return err
	}

	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	err = h.db.QueryRow(query, sessionHistory, session.ExpiresAt).Scan(&session.SessionId)
	if err != nil {
		sessionlogger.Error("Create session failed", "error", err)
		return err
	}

	return nil
}

func (h *PGSesssionsRepository) GetByID(id int) (*models.Session, error) {
	query := `
		SELECT session_id, session_history, expires_at FROM sessions
		WHERE session_id = $1
	`
	var session models.Session
	var sessionHistoryJSON []byte
	err := h.db.QueryRow(query, id).Scan(&session.SessionId, &sessionHistoryJSON, &session.ExpiresAt)
	if err != nil {
		sessionlogger.Error("Get session by ID failed", "error", err)
		return nil, err
	}
	err = json.Unmarshal(sessionHistoryJSON, &session.Sessions)
	if err != nil {
		sessionlogger.Error("Map unmarshal failed while getting session from db", "error", err)
		return nil, err
	}
	return &session, nil
}

func (h *PGSesssionsRepository) GetAll() ([]models.Session, error) {
	query := `
		SELECT session_id, session_history, expires_at FROM sessions
	`

	rows, err := h.db.Query(query)
	if err != nil {
		sessionlogger.Error("Get all sessions failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		var sessionHistoryJSON []byte

		err := rows.Scan(&session.SessionId, &sessionHistoryJSON, &session.ExpiresAt)
		if err != nil {
			sessionlogger.Error("Scan session failed", "error", err)
			return nil, err
		}

		err = json.Unmarshal(sessionHistoryJSON, &session.Sessions)
		if err != nil {
			sessionlogger.Error("Map unmarshal failed while getting session from db", "error", err)
			return nil, err
		}

		sessions = append(sessions, session)
	}
	if err = rows.Err(); err != nil {
		sessionlogger.Error("Iterate sessions failed", "error", err)
		return nil, err
	}

	return sessions, nil
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
		sessionlogger.Error("check deleted rows failed", "id", id, "error", err)
		return err
	}
	if rowsAffected == 0 {
		err := fmt.Errorf("session not found: id=%d", id)
		sessionlogger.Error("delete failed", "error", err)
		return err
	}

	return nil
}

func (h *PGSesssionsRepository) DeleteExpired() error {
	query := `
		DELETE FROM sessions
		WHERE expires_at < NOW()
	`

	_, err := h.db.Exec(query)
	if err != nil {
		sessionlogger.Error("Delete expired sessions failed", "error", err)
		return err
	}

	return nil
}
