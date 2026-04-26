package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
	"errors"
	"log/slog"
)

var anonlogger = slog.With("adapter", "postgres", "repository", "anons")

type PGAnonsRepository struct {
	db *sql.DB
}

func NewPGAnonsRepository(db *sql.DB) *PGAnonsRepository {
	return &PGAnonsRepository{db: db}
}

func (h *PGAnonsRepository) Create(anon *models.Anon) error {
	query := `
		INSERT INTO Anons (name, post_id, avatar) 
		VALUES ($1, $2, $3)
		RETURNING anon_id
	`

	err := h.db.QueryRow(query, anon.AnonName, anon.PostID, anon.Avatar).Scan(&anon.AnonID)
	if err != nil {
		anonlogger.Error("Create anon failed", "error", err)
		return err
	}

	return nil
}

func (h *PGAnonsRepository) GetByID(id int) (*models.Anon, error) {
	query := `
		SELECT anon_id, post_id, avatar, name FROM Anons 
		WHERE anon_id = $1
	`
	var anon models.Anon
	err := h.db.QueryRow(query, id).Scan(&anon.AnonID, &anon.PostID, &anon.Avatar, &anon.AnonName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			anonlogger.Info("anon not found", "id", id)
			return nil, models.ErrNotFound
		}
		anonlogger.Error("Get anon by ID failed", "error", err)
		return nil, err
	}
	return &anon, nil
}

func (h *PGAnonsRepository) GetAll() ([]models.Anon, error) {
	query := `
		SELECT (anon_id, post_id, avatar, name) FROM Anons	
	`
	rows, err := h.db.Query(query)
	if err != nil {
		anonlogger.Error("Get all anons failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var anons []models.Anon
	for rows.Next() {
		var anon models.Anon
		err := rows.Scan(&anon.AnonID, &anon.PostID, &anon.Avatar, &anon.AnonName)
		if err != nil {
			anonlogger.Error("Scan anon failed", "error", err)
			return nil, err
		}
		anons = append(anons, anon)
	}
	if err = rows.Err(); err != nil {
		anonlogger.Error("Iterate anons failed", "error", err)
		return nil, err
	}
	return anons, nil
}

// вроде пока не используется нигде
func (h *PGAnonsRepository) GetAllByPostID(id int) ([]models.Anon, error) {
	query := `
		SELECT (anon_id, post_id, avatar, name) FROM Anons	
		WHERE post_id = $1
	`
	rows, err := h.db.Query(query, id)
	if err != nil {
		anonlogger.Error("Get all anons by post ID failed", "error", err, "postID", id)
		return nil, err
	}
	defer rows.Close()

	var anons []models.Anon
	for rows.Next() {
		var anon models.Anon
		err := rows.Scan(&anon.AnonID, &anon.PostID, &anon.Avatar, &anon.AnonName)
		if err != nil {
			anonlogger.Error("Scan anon failed", "error", err)
			return nil, err
		}
		anons = append(anons, anon)
	}
	if err = rows.Err(); err != nil {
		anonlogger.Error("Iterate anons failed", "error", err)
		return nil, err
	}
	return anons, nil
}

func (h *PGAnonsRepository) GetAvatarCountByPostID(id int) (map[string]int, error) {
	query := `
		SELECT avatar, COUNT(*)
		FROM Anons
		WHERE post_id = $1
		GROUP BY avatar
	`
	avatar_count := make(map[string]int)
	rows, err := h.db.Query(query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		anonlogger.Error("Get avatar count by post ID failed", "error", err, "postID", id)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var avatar string
		var count int
		err := rows.Scan(&avatar, &count)
		if err != nil {
			anonlogger.Error("Scan avatar count failed", "error", err)
			return nil, err
		}
		avatar_count[avatar] = count
	}
	if err = rows.Err(); err != nil {
		anonlogger.Error("Iterate avatar counts failed", "error", err)
		return nil, err
	}
	return avatar_count, nil
}

func (h *PGAnonsRepository) Delete(id int) error {
	query := `
		DELETE FROM Anons
		WHERE anon_id = $1
	`
	result, err := h.db.Exec(query, id)
	if err != nil {
		anonlogger.Error("Delete anon failed", "error", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		anonlogger.Error("check deleted rows failed", "id", id, "error", err)
		return err
	}
	if rowsAffected == 0 {
		anonlogger.Error("anon not found", "id", id)
		return errors.New("anon not found")
	}
	return nil
}

func (h *PGAnonsRepository) DeleteByPostID(postID int) error {
	query := `
		DELETE FROM Anons
		WHERE post_id = $1
	`
	result, err := h.db.Exec(query, postID)
	if err != nil {
		anonlogger.Error("Delete anon failed", "error", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		anonlogger.Error("check deleted rows failed", "postID", postID, "error", err)
		return err
	}
	if rowsAffected == 0 {
		anonlogger.Error("anon not found", "postID", postID)
		return errors.New("anon not found")
	}
	return nil
}
