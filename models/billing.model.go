package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Billing struct {
    ID        uint64         `gorm:"primaryKey"`
    UserID    uuid.UUID  	 `gorm:"type:uuid;not null"`
    Amount    float64        `gorm:"not null"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt *time.Time `gorm:"index"`
}


// func GetUserByUsername(db *gorm.DB, Name string) (*UserResponse, error) {
// 	user := &User{}
// 	result := db.Where("name = ?", Name).First(user)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}

// 	userResponse := &UserResponse{
// 		ID:       user.ID,
// 		Name:	  user.Name,

// 	}
// 	result = db.Model(user).Preload("Billings").Find(&userResponse.Billings)
// 	if result.Error != nil {
// 		return nil, result.Error
// 	}

// 	return userResponse, nil
// }
