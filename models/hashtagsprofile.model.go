package models

import (
	"time"
)

type Hashtagsprofile struct {
	ID        uint       `gorm:"primary_key"`
	Hashtag   string     `gorm:"not null;unique"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:"index"`
}