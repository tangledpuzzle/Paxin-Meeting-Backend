package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Notification struct {
	ID        uint      `gorm:"primary_key"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	URL       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}
