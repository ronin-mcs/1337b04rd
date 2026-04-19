package dbadapter

import (
	"1337b04rd/models"
	"database/sql"
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
		INSERT INTO Anons (name) 
		VALUES ($1)
		RETURNING anon_id
	`

	err := h.db.QueryRow(query, anon.AnonName).Scan(&anon.AnonID)
	if err != nil {
		anonlogger.Error("Create anon failed", "error", err)
		return err
	}

	return nil
}

func (h *PGAnonsRepository) GetByID(id int) (*models.Anon, error) {
	query := `
		SELECT (anon_id, post_id, avatar, name) FROM Anons 
		WHERE anon_id = $1
	`
	var anon models.Anon
	err := h.db.QueryRow(query, id).Scan(&anon.AnonID, &anon.PostID, &anon.Avatar, &anon.Anon8uiName)
	if err != nil {
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
		err := rows.Scan(&anon.AnonID, &anon.PostID, &anon.Avatar, &anon.Anon8uiName)
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
