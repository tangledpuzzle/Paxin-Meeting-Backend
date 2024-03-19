package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/datatypes"
)

type ChatRoomMember struct {
	ID                uint64 `gorm:"primaryKey"`
	RoomID            uint64
	UserID            uuid.UUID
	Room              ChatRoom  `gorm:"foreignKey:RoomID"`
	User              User      `gorm:"foreignKey:UserID"`
	IsSubscribed      bool      `gorm:"not null;default:false"`
	IsNew             bool      `gorm:"not null;default:false"`
	JoinedAt          time.Time `gorm:"not null;default:now()"`
	LastReadMessageID *uint64
}

type ChatRoom struct {
	ID            uint64           `gorm:"primaryKey"`
	Name          string           `gorm:"size:64;unique"`
	Members       []ChatRoomMember `gorm:"foreignKey:RoomID"`
	Version       uint64           `gorm:"default:0"`
	CreatedAt     time.Time        `gorm:"not null;default:now()"`
	BumpedAt      time.Time        `gorm:"not null;default:now()"`
	LastMessageID *uint64
	LastMessage   *ChatMessage `gorm:"foreignKey:LastMessageID"`
}

type ChatMessage struct {
	ID      uint64 `gorm:"primaryKey"`
	Content string `gorm:"not null"`
	UserID  uuid.UUID
	RoomID  uint64
	User    User `gorm:"foreignKey:UserID"`
	// Room      ChatRoom   `gorm:"foreignKey:RoomID"`
	IsEdited  bool       `gorm:"not null;default:false"`
	IsDeleted bool       `gorm:"not null;default:false"`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
	DeletedAt *time.Time `gorm:"index"`
	MsgType   uint8      `gorm:"not null;default:0"` // 0: common, 1: conference, 2: attached post link
	JsonData  *string    `gorm:"type:jsonb"`
	// IsRead    bool       `gorm:"not null;default:false"`
	ParentMessageID *uint64
	ParentMessage   *ChatMessage `gorm:"foreignKey:ParentMessageID"`
}

type ChatOutbox struct {
	ID        uint64 `gorm:"primaryKey"`
	Method    string `gorm:"type:text;default:publish"`
	Payload   datatypes.JSON
	Partition int64 `gorm:"default:0"`
	CreatedAt time.Time
}

type ChatCDC struct {
	ID        uint64 `gorm:"primaryKey"`
	Method    string `gorm:"type:text;default:publish"`
	Payload   datatypes.JSON
	Partition int64 `gorm:"default:0"`
	CreatedAt time.Time
}
