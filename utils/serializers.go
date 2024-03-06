package utils

import (
	"encoding/json"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"

	uuid "github.com/satori/go.uuid"
)

func FetchCitiesForProfile(profileID uint64) ([]models.City, error) {
	var cities []models.City
	err := initializers.DB.Joins("JOIN profiles_city ON profiles_city.city_id = cities.id").Where("profiles_city.profile_id = ?", profileID).Find(&cities).Error
	if err != nil {
		return nil, err
	}
	return cities, nil
}

func FetchGuildsForProfile(profileID uint64) ([]models.Guilds, error) {
	var guilds []models.Guilds
	err := initializers.DB.Joins("JOIN profiles_guilds ON profiles_guilds.guild_id = guilds.id").Where("profiles_guilds.profile_id = ?", profileID).Find(&guilds).Error
	if err != nil {
		return nil, err
	}
	return guilds, nil
}

func FetchHashtagsForProfile(profileID uint64) ([]models.Hashtags, error) {
	var hashtags []models.Hashtags
	err := initializers.DB.Joins("JOIN profiles_hashtags ON profiles_hashtags.hashtag_id = hashtags.id").Where("profiles_hashtags.profile_id = ?", profileID).Find(&hashtags).Error
	if err != nil {
		return nil, err
	}
	return hashtags, nil
}

func FetchUserByID(userID uuid.UUID) (models.User, error) {
	var user models.User
	err := initializers.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func FetchPhotosForProfile(profileID uint64) ([]models.ProfilePhoto, error) {
	var photos []models.ProfilePhoto
	err := initializers.DB.Where("profile_id = ?", profileID).Find(&photos).Error
	if err != nil {
		return nil, err
	}
	return photos, nil
}

func SerializeChatRoomWithDetails(room models.ChatRoom) map[string]interface{} {
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

func SerializeChatRoomMember(member models.ChatRoomMember) map[string]interface{} {
	return map[string]interface{}{
		"id":            member.ID,
		"room_id":       member.RoomID,
		"user_id":       member.UserID.String(),
		"is_subscribed": member.IsSubscribed,
		"is_new":        member.IsNew,
		"joined_at":     member.JoinedAt,
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

func SerializeHashtag(hashtag models.Hashtags) map[string]interface{} {
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

func SerializeProfile(profile models.Profile, cities []models.City, guilds []models.Guilds, hashtags []models.Hashtags, photos []models.ProfilePhoto) map[string]interface{} {
	var serializedCities, serializedGuilds, serializedHashtags []map[string]interface{}

	for _, city := range cities {
		serializedCities = append(serializedCities, SerializeCity(city))
	}

	for _, guild := range guilds {
		serializedGuilds = append(serializedGuilds, SerializeGuild(guild))
	}

	for _, hashtag := range hashtags {
		serializedHashtags = append(serializedHashtags, SerializeHashtag(hashtag))
	}

	var serializedPhotos []map[string]interface{}
	for _, photo := range photos {
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
	serializedProfiles := []map[string]interface{}{}
	for _, profile := range user.Profile {
		cities, _ := FetchCitiesForProfile(profile.ID)
		guilds, _ := FetchGuildsForProfile(profile.ID)
		hashtags, _ := FetchHashtagsForProfile(profile.ID)
		photos, _ := FetchPhotosForProfile(profile.ID)
		serializedProfiles = append(serializedProfiles, SerializeProfile(profile, cities, guilds, hashtags, photos))
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
