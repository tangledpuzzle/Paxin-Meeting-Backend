package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Meta map[string]string

func (m Meta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Meta) Scan(src interface{}) error {
	if src == nil {
		return errors.New("source is nil")
	}
	var data []byte
	switch src := src.(type) {
	case []byte:
		data = src
	case string:
		data = []byte(src)
	default:
		return errors.New("unsupported Scan, storing driver.Value type into type *models.Meta")
	}
	return json.Unmarshal(data, m)
}

type Presavedfilters struct {
	ID        uint64     `gorm:"primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	Name      string     `gorm:"not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	Meta      Meta       `gorm:"type:jsonb"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
