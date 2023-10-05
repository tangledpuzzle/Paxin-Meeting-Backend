package models

import (
	"github.com/jackc/pgtype"
	uuid "github.com/satori/go.uuid"

	"time"
)


type Domain struct {
    ID        uint64     `gorm:"primaryKey"`
    UserID   uuid.UUID   // Use uuid.UUID type for primary key
    Username string      `gorm:"not null"`
	Status 	 string 	 `gorm:"not null;default:activated"`
    Name     string      `gorm:"not null"`
    CreatedAt time.Time  `gorm:"not null"`
    ExpiredAt   *time.Time `gorm:"index"`
    UpdatedAt time.Time  `gorm:"not null"`
    DeletedAt *time.Time `gorm:"index"`
    Settings  pgtype.JSONB     `gorm:"type:jsonb"` // Use pgtype.JSONB type
}

type DomainSettings struct {
    Logo      string `json:"logo"`
    Title     string `json:"title"`
    MetaDescr string `json:"metadescr"`
    Style     string `json:"style"`
}
