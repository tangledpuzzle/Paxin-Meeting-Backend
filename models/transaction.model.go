package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Transaction struct {
	ID        	uint64        `gorm:"primaryKey"`
	UserID   	uuid.UUID     `gorm:"type:uuid;not null"`
    ElementId    uint64     `gorm:"not null"`
	Module		string		`gorm:"not null"`
	Amount    	float64       `gorm:"not null"`
	Description string        `gorm:"not null"` 
	Type        string        `gorm:"not null"`    
	Status       string        `gorm:"null"`    
	Total       string        `gorm:"null"`    

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
	DeletedAt *time.Time    `gorm:"index"`

}
