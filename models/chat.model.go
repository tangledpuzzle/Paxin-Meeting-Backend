package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type RoomMember struct {
	ID           uint `gorm:"primaryKey"`
	RoomID       uint
	UserID       uuid.UUID
	Room         Room      `gorm:"foreignKey:RoomID"`
	User         *User     `gorm:"foreignKey:UserID"`
	IsSubscribed bool      `gorm:"not null;default:false"`
	IsNew        bool      `gorm:"not null;default:false"`
	JoinedAt     time.Time `gorm:"not null;default:now()"`
}

type Room struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"size:64;unique"`
	Members     []RoomMember `gorm:"foreignKey:RoomID"`
	Version     uint64       `gorm:"default:0"`
	CreatedAt   time.Time    `gorm:"not null;default:now()"`
	BumpedAt    time.Time    `gorm:"not null;default:now()"`
	LastMessage *Message
}

type Message struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"not null"`
	UserID    uuid.UUID
	RoomID    uint
	User      *User     `gorm:"foreignKey:UserID"`
	Room      Room      `gorm:"foreignKey:RoomID"`
	IsEdited  bool      `gorm:"not null;default:false"`
	IsDeleted bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not/null;default:now()"`
}
