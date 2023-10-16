package models

import (
	// "hyperpage/initializers"
	"time"

	"github.com/go-playground/validator/v10"
	uuid "github.com/satori/go.uuid"
)

type Blog struct {
	ID         uint64      `gorm:"primaryKey"`
	Title      string      `gorm:"not null"`
	Descr      string      `gorm:"not null"`
	Slug       string      `gorm:"not null"`
	Content    string      `gorm:"null"`
	Status     string      `gorm:"not null"`
	Sticker    string      `gorm:"not null;default:standart"`
	City       []City      `gorm:"many2many:blog_city;"`
	Catygory   []Guilds    `gorm:"many2many:blog_guilds;"`
	UniqId     string      `gorm:"not null;default:0"`
	Days       int         `gorm:"not null;default:3"`
	Views      int         `gorm:"not null;default:0"`
	Total      float64     `gorm:"null"`
	TmId       float64     `gorm:"not null;default:0"`
	Photos     []BlogPhoto `json:"photos"`
	NotAds     bool        `gorm:"not null;default:true"`
	User       User        `gorm:"foreignKey:UserID"`
	UserAvatar string      `gorm:"not null"`
	Pined      bool        `gorm:"not null;default:false"`
	UserID     uuid.UUID   `gorm:"type:uuid;not null"`
	CreatedAt  time.Time   `gorm:"not null"`
	UpdatedAt  time.Time   `gorm:"not null"`
	DeletedAt  *time.Time  `gorm:"index"`
	ExpiredAt  *time.Time  `gorm:"index"`
	Hashtags   []Hashtags  `gorm:"many2many:blog_hashtags;"`
}

type BlogResponse struct {
	ID        uint64       `json:"id"`
	Title     string       `json:"title"`
	Catygory  []string     `json:"catygory"`
	Days      int          `json:"days"`
	Views     int          `json:"views"`
	Descr     string       `json:"descr"`
	Slug      string       `json:"slug"`
	Content   string       `json:"content"`
	Status    string       `json:"status"`
	UniqId    string       `json:"uniqId"`
	City      []string     `json:"city"`
	Sticker   string       `json:"sticker"`
	Total     float64      `json:"total"`
	Pined     bool         `json:"pined"`
	UserID    uuid.UUID    `json:"userId"`
	TmId      float64      `gorm:"tId"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt *time.Time   `json:"deletedAt"`
	ExpiredAt *time.Time   `json:"expiredAt"`
	Photos    []BlogPhoto  `json:"photos"`
	User      UserResponse `json:"user"`

	Hashtags []string `json:"hashtags"`
}

type CreateBlogInput struct {
	Title    string      `json:"title" validate:"required,min=3,max=80"`
	Content  string      `json:"content"`
	Total    float64     `json:"total" validate:"min=2,max=100"`
	Status   string      `json:"status" validate:"required,min=10"`
	Descr    string      `json:"descr" validate:"required,min=10,max=300"`
	City     []string    `json:"city" validate:"required"`
	Slug     string      `json:"slug" validate:"required"`
	Catygory string      `json:"category" validate:"required"`
	Days     int         `json:"days" validate:"required"`
	Hashtags []string    `json:"hashtags" validate:"required"`
	Photos   []BlogPhoto `json:"photos"`
}

func (b *Blog) Validate() error {
	validate := validator.New()
	return validate.Struct(b)
}
