package utils

import (
	"encoding/json"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func FetchUserByID(userID uuid.UUID) (models.User, error) {
	var user models.User
	err := initializers.DB.Preload("Profile", func(db *gorm.DB) *gorm.DB {
		return db.Preload("City", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Translations")
		}).Preload("Guilds", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Translations")
		}).Preload("Hashtags").Preload("Photos")
	}).Where("id = ?", userID).First(&user).Error

	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func SerializeChatRoomMember(member models.ChatRoomMember) map[string]interface{} {
	user, _ := FetchUserByID(member.User.ID)
	userSerialized := SerializeUser(user)
	return map[string]interface{}{
		"id":            member.ID,
		"room_id":       member.RoomID,
		"user":          userSerialized,
		"is_subscribed": member.IsSubscribed,
		"is_new":        member.IsNew,
		"joined_at":     member.JoinedAt,
	}
}

func SerializeChatRoom(roomID uint) map[string]interface{} {
	var room models.ChatRoom
	err := initializers.DB.Preload("Members.User").Preload("LastMessage").First(&room, roomID).Error
	if err != nil {
		return nil
	}

	roomMap := map[string]interface{}{
		"id":           room.ID,
		"name":         room.Name,
		"version":      room.Version,
		"created_at":   room.CreatedAt,
		"bumped_at":    room.BumpedAt,
		"member_count": len(room.Members),
	}

	if room.LastMessage != nil {
		roomMap["last_message"] = SerializeChatMessage(*room.LastMessage)
	}

	membersSerialized := make([]map[string]interface{}, 0, len(room.Members))
	for _, member := range room.Members {
		membersSerialized = append(membersSerialized, SerializeChatRoomMember(member))
	}
	roomMap["members"] = membersSerialized

	return roomMap
}

func SerializeChatMessage(message models.ChatMessage) map[string]interface{} {
	user, err := FetchUserByID(message.UserID)
	if err != nil {
		return nil
	}
	serializedUser := SerializeUser(user)

	return map[string]interface{}{
		"id":         message.ID,
		"content":    message.Content,
		"user_id":    message.UserID.String(),
		"user":       serializedUser,
		"room_id":    message.RoomID,
		"is_edited":  message.IsEdited,
		"created_at": message.CreatedAt,
		"is_deleted": message.IsDeleted,
	}
}

func SerializeCity(city models.City) map[string]interface{} {
	var translations []map[string]interface{}
	for _, t := range city.Translations {
		translations = append(translations, map[string]interface{}{
			"id":       t.ID,
			"language": t.Language,
			"name":     t.Name,
		})
	}

	return map[string]interface{}{
		"id":           city.ID,
		"country_code": city.CountryCode,
		"hex":          city.Hex,
		"updated_at":   city.UpdatedAt,
		"translations": translations,
	}
}

func SerializeGuild(guild models.Guilds) map[string]interface{} {
	var translations []map[string]interface{}
	for _, t := range guild.Translations {
		translations = append(translations, map[string]interface{}{
			"id":       t.ID,
			"language": t.Language,
			"name":     t.Name,
		})
	}

	return map[string]interface{}{
		"id":           guild.ID,
		"hex":          guild.Hex,
		"updated_at":   guild.UpdatedAt,
		"translations": translations,
	}
}

func SerializeHashtag(hashtag models.HashtagsForProfile) map[string]interface{} {
	return map[string]interface{}{
		"id":         hashtag.ID,
		"hashtag":    hashtag.Hashtag,
		"updated_at": hashtag.UpdatedAt,
	}
}

func SerializeProfilePhoto(photo models.ProfilePhoto) map[string]interface{} {
	var filesData interface{}
	if err := json.Unmarshal(photo.Files.Bytes, &filesData); err != nil {
		fmt.Println("Error unmarshalling ProfilePhoto Files JSONB:", err)
	}

	return map[string]interface{}{
		"id":         photo.ID,
		"profile_id": photo.ProfileID,
		"created_at": photo.CreatedAt,
		"updated_at": photo.UpdatedAt,
		"files":      filesData,
	}
}

func SerializeProfile(profile models.Profile) map[string]interface{} {

	serializedCities := make([]map[string]interface{}, 0)
	for _, city := range profile.City {
		serializedCities = append(serializedCities, SerializeCity(city))
	}

	serializedGuilds := make([]map[string]interface{}, 0)
	for _, guild := range profile.Guilds {
		serializedGuilds = append(serializedGuilds, SerializeGuild(guild))
	}

	serializedHashtags := make([]map[string]interface{}, 0)
	for _, hashtag := range profile.Hashtags {
		serializedHashtags = append(serializedHashtags, SerializeHashtag(hashtag))
	}

	serializedPhotos := make([]map[string]interface{}, 0)
	for _, photo := range profile.Photos {
		serializedPhotos = append(serializedPhotos, SerializeProfilePhoto(photo))
	}

	return map[string]interface{}{
		"id":                   profile.ID,
		"user_id":              profile.UserID,
		"firstname":            profile.Firstname,
		"description":          profile.Descr,
		"multilang_descr":      profile.MultilangDescr,
		"additional":           profile.Additional,
		"multilang_additional": profile.MultilangAdditional,
		"cities":               serializedCities,
		"guilds":               serializedGuilds,
		"hashtags":             serializedHashtags,
		"photos":               serializedPhotos,
		"created_at":           profile.CreatedAt,
		"updated_at":           profile.UpdatedAt,
	}
}

func SerializeUser(user models.User) map[string]interface{} {
	serializedProfiles := make([]map[string]interface{}, 0)
	for _, profile := range user.Profile {
		serializedProfile := SerializeProfile(profile)
		serializedProfiles = append(serializedProfiles, serializedProfile)
	}

	return map[string]interface{}{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"role":           user.Role,
		"telegramname":   user.TelegramName,
		"photo":          user.Photo,
		"limitstorage":   user.LimitStorage,
		"banned":         user.Banned,
		"plan":           user.Plan,
		"profile":        serializedProfiles,
		"expirePlanAt":   user.ExpiredPlanAt,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"online":         user.Online,
		"totalblogs":     user.TotalBlogs,
		"totalrestblog":  user.TotalRestBlogs,
		"totalfollowers": user.TotalFollowers,
	}
}
