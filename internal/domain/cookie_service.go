package domain

import (
	"1337b04rd/models"
	"time"
)

func (h *PostService) CreateNewSessionID() (int, error) {
	session := &models.Session{
		SessionID: 0,
		Sessions:  map[string]int{},
		ExpiredAt: time.Time{},
	}
	err := h.sessions.Create(session)
	if err != nil {
		return 0, err
	}

	return session.SessionID, nil

}
