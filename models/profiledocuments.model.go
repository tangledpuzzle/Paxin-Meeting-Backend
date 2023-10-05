package models

import (
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgtype"

	"database/sql/driver"
	"time"
)

type ProfileDocuments struct {
	ID           uint64       `gorm:"primaryKey"`
	ProfileID    uint64       `gorm:"not null"`
	CreatedAt    time.Time    `gorm:"not null"`
	UpdatedAt    time.Time    `gorm:"not null"`
	DeletedAt    *time.Time   `gorm:"index"`
	Name         string       `gorm:"not null"`
	Organization string       `gorm:"null"`
	Descr        string       `gorm:"null"`
	Files        pgtype.JSONB `json:"files" gorm:"type:jsonb"`
}

// Scan implements the sql.Scanner interface
func (bp *ProfileDocuments) Scan(src interface{}) error {
	return bp.Files.Scan(src)
}

// Value implements the driver.Valuer interface
func (bp ProfileDocuments) Value() (driver.Value, error) {
	return bp.Files.Value()
}

func (b *ProfileDocuments) Validate() error {
	validate := validator.New()
	return validate.Struct(b)
}
