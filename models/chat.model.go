package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type ChatRoomMember struct {
	ID           uint `gorm:"primaryKey"`
	RoomID       uint
	UserID       uuid.UUID
	Room         ChatRoom  `gorm:"foreignKey:RoomID"`
	User         *User     `gorm:"foreignKey:UserID"`
	IsSubscribed bool      `gorm:"not null;default:false"`
	IsNew        bool      `gorm:"not null;default:false"`
	JoinedAt     time.Time `gorm:"not null;default:now()"`
}

type ChatRoom struct {
	ID            uint             `gorm:"primaryKey"`
	Name          string           `gorm:"size:64;unique"`
	Members       []ChatRoomMember `gorm:"foreignKey:RoomID"`
	Version       uint64           `gorm:"default:0"`
	CreatedAt     time.Time        `gorm:"not null;default:now()"`
	BumpedAt      time.Time        `gorm:"not null;default:now()"`
	LastMessageID *uint
	LastMessage   *ChatMessage `gorm:"foreignKey:LastMessageID"`
}

type ChatMessage struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"not null"`
	UserID    uuid.UUID
	RoomID    uint
	User      *User     `gorm:"foreignKey:UserID"`
	Room      ChatRoom  `gorm:"foreignKey:RoomID"`
	IsEdited  bool      `gorm:"not null;default:false"`
	IsDeleted bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not/null;default:now()"`
}
