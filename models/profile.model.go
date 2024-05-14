package models

import (
	"time"

	"github.com/jackc/pgtype"
	uuid "github.com/satori/go.uuid"
)

type JSONB pgtype.JSONB

func (j *JSONB) Scan(src interface{}) error {
	return (*pgtype.JSONB)(j).Scan(src)
}

func (j JSONB) Value() (interface{}, error) {
	return (pgtype.JSONB)(j).Value()
}

type Streaming struct {
	RoomID    string    `gorn:"primaryKey"`
	Title     string    `gorm:"not null"`
	UserID    string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type Profile struct {
	ID             uint64         `gorm:"primaryKey"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null"`
	Firstname      string         `gorm:"not null"`
	Tcid           int64          `gorm:"null;"`
	Descr          string         `gorm:"not null"`
	MultilangDescr MultilangTitle `gorm:"embedded;embeddedPrefix:multilang_Descr_"`

	City      []City               `gorm:"many2many:profiles_city;"`
	Guilds    []Guilds             `gorm:"many2many:profiles_guilds;"`
	Hashtags  []HashtagsForProfile `gorm:"many2many:profiles_hashtags;"`
	Photos    []ProfilePhoto       `json:"photos"`
	Documents []ProfileDocuments   `json:"documents"`
	Service   []ProfileService     `json:"service"`

	Additional          string         `json:"additional"`
	MultilangAdditional MultilangTitle `gorm:"embedded;embeddedPrefix:multilang_Additional_"`
	Lang                string         `gorm:"not null;default:en"`

	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:"index"`
	User      User       `gorm:"foreignKey:UserID"`

	Streaming []Streaming `json:"Streaming"`
}

type ProfileResponse struct {
	ID             uint64         `gorm:"not null"`
	Firstname      string         `gorm:"not null"`
	Descr          string         `gorm:"not null"`
	MultilangDescr MultilangTitle `json:"multilangtitle"`

	Tcid                int64              `gorm:"null"`
	City                []string           `gorm:"city"`
	Guilds              []string           `json:"guilds"`
	Hashtags            []string           `json:"hashtags"`
	Photos              []ProfilePhoto     `json:"photos"`
	Documents           []ProfileDocuments `json:"documents"`
	Service             []ProfileService   `json:"service"`
	Additional          string             `json:"additional"`
	MultilangAdditional MultilangTitle     `json:"multilangadditional"`
}
