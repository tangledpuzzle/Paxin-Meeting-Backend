package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	uuid "github.com/satori/go.uuid"
)

type TimeEntry struct {
	Hour    int `json:"hour"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type TimeEntryScanner []TimeEntry

func (t TimeEntryScanner) Value() (driver.Value, error) {
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

func (t *TimeEntryScanner) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		var entries []TimeEntry
		if err := json.Unmarshal(v, &entries); err != nil {
			return err
		}
		*t = entries
		return nil
	case string:
		var entries []TimeEntry
		if err := json.Unmarshal([]byte(v), &entries); err != nil {
			return err
		}
		*t = entries
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for TimeEntryScanner: %T", value)
	}
}

type User struct {
	ID                 uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name               string     `gorm:"type:varchar(100);not null"`
	Email              string     `gorm:"type:varchar(100);uniqueIndex:idx_email;not null"`
	Password           string     `gorm:"type:varchar(100);not null"`
	Role               string     `gorm:"type:varchar(50);default:'user';not null"`
	Provider           string     `gorm:"type:varchar(50);default:'local';not null"`
	Photo              string     `gorm:"not null;default:'default.png'"`
	Verified           bool       `gorm:"not null;default:false"`
	Banned             bool       `gorm:"not null;default:false"`
	Plan               string     `gorm:"not null;default:standart"`
	Signed             bool       `gorm:"not null;default:false"`
	ExpiredPlanAt      *time.Time `gorm:"index"`
	Tcid               int64      `gorm:"null;"`
	TotalFollowers     int64      `gorm:"null;default:0"`
	VerificationCode   string
	PasswordResetToken string
	TelegramActivated  bool `gorm:"not null;default:false"`
	TelegramToken      string
	PasswordResetAt    time.Time
	Billing            []Billing        `gorm:"foreignkey:UserID"`
	Profile            []Profile        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Filled             bool             `json:"filled"`
	Session            string           `gorm:"nut null;default:0"`
	Storage            string           `gorm:"nut null"`
	Tid                int64            `gorm:"nut null;default:0"`
	Blogs              []Blog           `gorm:"foreignkey:UserID"`
	CreatedAt          time.Time        `gorm:"not null;default:now()"`
	UpdatedAt          time.Time        `gorm:"not null;default:now()"`
	OnlineHours        TimeEntryScanner `gorm:"type:json;default:'[{\"hour\":0,\"minutes\":0,\"seconds\":0}]'"`
	TotalOnlineHours   TimeEntryScanner `gorm:"type:json;default:'[{\"hour\":0,\"minutes\":0,\"seconds\":0}]'"`
	OfflineHours       int              `gorm:"not null;default:0"`
	TotalRestBlogs     int              `gorm:"not null;default:0"`
	TotalBlogs         int              `gorm:"not null;default:0"`
	Rating             int              `gorm:"not null;default:0"`
	LimitStorage       int              `gorm:"not null;default:20"`
	LastOnline         time.Time        `json:"last_online"`
	Online             bool             `json:"online"`
	Domains            []Domain         `json:"domains"`
	Followings         []*User          `gorm:"many2many:user_relation;joinForeignKey:user_Id;JoinReferences:following_id"`
	Followers          []*User          `gorm:"many2many:user_relation;joinForeignKey:following_id;JoinReferences:user_Id"`
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type SignUpInput struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required,min=6"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,min=6"`
	Photo           string `json:"photo,omitempty"`
}

type SignInInput struct {
	Email    string `json:"email"  validate:"required"`
	Password string `json:"password"  validate:"required"`
}

type DomainResponse struct {
	ID       uint64                 `json:"id"`
	UserID   string                 `json:"user_id"`
	Username string                 `json:"username"`
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
}

type UserResponse struct {
	ID                uuid.UUID         `json:"id,omitempty"`
	Name              string            `json:"name,omitempty"`
	Email             string            `json:"email,omitempty"`
	Blogs             []BlogResponse    `json:"blogs"`
	Role              string            `json:"role,omitempty"`
	TelegramToken     string            `json:"telegram_token,omitempty"`
	TelegramActivated bool              `bool:"telegram_activated"`
	Photo             string            `json:"photo,omitempty"`
	Session           string            `json:"session"`
	Storage           string            `json:"storage"`
	TId               int64             `json:"tid"`
	Tcid              int64             `json:"tcid"`
	LimitStorage      int               `json:"limitstorage"`
	Banned            bool              `bool:"banned"`
	Plan              string            `bool:"plan"`
	Profile           []ProfileResponse `json:"profile"`
	ExpiredPlanAt     *time.Time        `json:"expirePlandAt"`
	OnlineHours       TimeEntryScanner  `json:"online_hours"`
	TotalOnlineHours  TimeEntryScanner  `json:"total_online_hours"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Signed            bool              `json:"signed"`
	Filled            bool              `bool:"filled"`
	Online            bool              `bool:"online"`
	TotalBlogs        int               `json:"totalblogs"`
	TotalRestBlogs    int               `json:"totalrestblog"`
	Domains           []Domain          `json:"domains"`
	Followings        []*User           `json:"followings"`
	Followers         []*User           `json:"followers"`
	TotalFollowers    int64             `json:"totalfollowers"`
}

func FilterUserRecord(user *User, language string) UserResponse {
	var profileResponses []ProfileResponse
	for _, profile := range user.Profile {
		profileResponse := ProfileResponse{
			ID: profile.ID,
			// Lastname: profile.Lastname,
			// MiddleN:  profile.MiddleN,
			Descr: profile.Descr,
		}

		guilds := make([]string, 0, len(profile.Guilds))
		for _, guild := range profile.Guilds {
			for _, translation := range guild.Translations {

				if translation.Language == language {
					pair := fmt.Sprintf(`{"id": "%d", "name": "%s"}`, translation.GuildID, translation.Name)
					guilds = append(guilds, pair)
				}
			}
		}

		profileResponse.Guilds = guilds

		hashtags := make([]string, 0, len(profile.Hashtags))
		for _, hashtag := range profile.Hashtags {
			hashtags = append(hashtags, hashtag.Hashtag)
		}
		profileResponse.Hashtags = hashtags

		cities := make([]string, 0, len(profile.City))
		for _, city := range profile.City {
			pair := fmt.Sprintf(`{"id": "%d", "name": "%s"}`, city.ID, city.Name)
			cities = append(cities, pair)
		}
		profileResponse.City = cities

		profileResponse.Photos = profile.Photos

		profileResponses = append(profileResponses, profileResponse)
	}

	// Check if there are no profiles
	if len(profileResponses) == 0 {
		profileResponses = nil // Set it to nil instead of an empty slice
	}

	return UserResponse{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		Role:              string(user.Role),
		Photo:             user.Photo,
		Session:           user.Session,
		Storage:           user.Storage,
		TelegramToken:     user.TelegramToken,
		TelegramActivated: user.TelegramActivated,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
		OnlineHours:       user.OnlineHours,
		TotalOnlineHours:  user.TotalOnlineHours,
		TId:               user.Tid,
		Tcid:              user.Tcid,
		Profile:           profileResponses,
		LimitStorage:      user.LimitStorage,
		Banned:            user.Banned,
		Plan:              user.Plan,
		Filled:            user.Filled,
		ExpiredPlanAt:     user.ExpiredPlanAt,
		Signed:            user.Signed,
		TotalRestBlogs:    user.TotalRestBlogs,
		Followings:        user.Followings,
		Followers:         user.Followers,
		TotalFollowers:    user.TotalFollowers,
	}
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordInput struct {
	Password        string `json:"password" binding:"required"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
}

var validate = validator.New()

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

func ValidateStruct[T any](payload T) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
