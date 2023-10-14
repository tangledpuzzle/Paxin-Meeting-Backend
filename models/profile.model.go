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

type Profile struct {
	ID        uint64    `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Firstname string    `gorm:"not null"`
	// Lastname   string        `gorm:"not null"`
	// MiddleN    string        `gorm:"not null"`
	Tcid      int64              `gorm:"null;"`
	Descr     string             `gorm:"not null"`
	City      []City             `gorm:"many2many:profiles_city;"`
	Guilds    []Guilds           `gorm:"many2many:profiles_guilds;"`
	Hashtags  []Hashtagsprofile  `gorm:"many2many:profiles_hashtags;"`
	Photos    []ProfilePhoto     `json:"photos"`
	Documents []ProfileDocuments `json:"documents"`
	Service   []ProfileService   `json:"service"`

	Additional string     `json:"additional"`
	CreatedAt  time.Time  `gorm:"not null"`
	UpdatedAt  time.Time  `gorm:"not null"`
	DeletedAt  *time.Time `gorm:"index"`
	User       User       `gorm:"foreignKey:UserID"`
}

type ProfileResponse struct {
	ID        uint64 `gorm:"not null"`
	Firstname string `gorm:"not null"`
	// Lastname  string         `gorm:"not null"`
	// MiddleN   string         `gorm:"not null"`
	Descr      string             `gorm:"not null"`
	Tcid       int64              `gorm:"null"`
	City       []string           `gorm:"city"`
	Guilds     []string           `json:"guilds"`
	Hashtags   []string           `json:"hashtags"`
	Photos     []ProfilePhoto     `json:"photos"`
	Documents  []ProfileDocuments `json:"documents"`
	Service    []ProfileService   `json:"service"`
	Additional string             `json:"additional"`
}
