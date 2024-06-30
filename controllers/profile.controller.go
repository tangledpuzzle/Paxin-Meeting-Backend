package controllers

import (
	"encoding/json"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	uuid "github.com/satori/go.uuid"

	"reflect"

	gt "github.com/bas24/googletranslatefree"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// func GetAllProfile(c *fiber.Ctx) error {

// 	var profiles []models.Profile
// 	query := initializers.DB.
// 		Order("created_at DESC").
// 		Preload("Guilds").
// 		Preload("Hashtags").
// 		Preload("City").
// 		Preload("Photos").
// 		Preload("User").
// 		Joins("JOIN users ON profiles.user_id = users.id").
// 		Where("Users.filled = ?", true)

// 	err := utils.Paginate(c, query.Find(&profiles), &profiles)
// 	if err != nil {
// 		return err
// 	}

//     return nil

// }

func GetProfiles(c *fiber.Ctx) error {

	language := c.Query("language")

	var profiles []models.Profile
	query := initializers.DB.
		Preload("Guilds.Translations", "language = ?", language).
		Preload("Hashtags").
		Preload("City.Translations", "language = ?", language).
		Preload("Photos").
		Preload("User").
		Joins("JOIN users ON profiles.user_id = users.id").
		Order("Users.name ASC").
		Where("Users.filled = ?", true)
	// Get the query parameters
	city := c.Query("city")
	hashtags := c.Query("hashtag")
	category := c.Query("category")
	// title := ""
	money := c.Query("money")

	if money != "" && money != "all" {
		if strings.Contains(money, "-") {
			alphabeticRange := strings.Split(money, "-")
			if len(alphabeticRange) != 2 {
				return fmt.Errorf("invalid alphabetic range format")
			}

			lowerAlpha := strings.TrimSpace(alphabeticRange[0])
			upperAlpha := strings.TrimSpace(alphabeticRange[1])
			query = query.Where("(LOWER(SUBSTR(Firstname, 1, 1)) >= ? AND LOWER(SUBSTR(Firstname, 1, 1)) <= ?) AND (LOWER(SUBSTR(Lastname, 1, 1)) >= ? AND LOWER(SUBSTR(Lastname, 1, 1)) <= ?)", lowerAlpha, upperAlpha, lowerAlpha, upperAlpha)

		} else {
			alphabeticRange := strings.Split(money, "-")

			lowerAlpha := strings.TrimSpace(alphabeticRange[0])

			query = query.Where("LOWER(SUBSTR(Firstname, 1, 1)) = ?", lowerAlpha)
		}
	}

	if city != "" && city != "all" {
		// Сначала найдем city_id для указанного города и языка
		var cityTranslation models.CityTranslation
		initializers.DB.Where("name = ? AND language = ?", city, language).First(&cityTranslation)
		fmt.Println(cityTranslation)
		if cityTranslation.ID != 0 {
			subQuery := initializers.DB.Table("profiles_city").
				Select("profile_id").
				Where("city_id = ?", cityTranslation.CityID)

			// Добавим условие, чтобы ваш основной запрос включал только записи с blog_id из подзапроса
			query = query.Where("profiles.id IN (?)", subQuery) // Specify the table alias for "blogs.id"
		}
	}

	// if title != "" && title != "all" {
	// 	query = query.Where("LOWER(Descr) LIKE ?", "%"+title+"%")
	// }

	if category != "" && category != "all" {
		var guildTranslation models.GuildTranslation
		initializers.DB.Where("name = ? AND language = ?", category, language).First(&guildTranslation)
		if guildTranslation.ID != 0 {
			// Создадим подзапрос для поиска всех blog_id, связанных с указанным guild_id
			subQuery := initializers.DB.Table("profiles_guilds").
				Select("profile_id").
				Where("guilds_id = ?", guildTranslation.GuildID)

			// Добавим условие, чтобы ваш основной запрос включил только записи с blog_id из подзапроса
			query = query.Where("profiles.id IN (?)", subQuery)
		}
	}

	if hashtags != "" && hashtags != "all" {
		// Split the hashtags into separate values
		hashtagValues := strings.Split(hashtags, ",")

		// Join the hashtag values with the '#' character
		hashtagValuesWithPrefix := make([]string, len(hashtagValues))
		for i, tag := range hashtagValues {
			hashtagValuesWithPrefix[i] = strings.TrimSpace(tag)
		}

		// Add the hashtags filter to the query
		query = query.Joins("JOIN profiles_hashtags ON profiles.id = profiles_hashtags.profile_id").
			Joins("JOIN hashtags_for_profiles ON profiles_hashtags.hashtags_for_profile_id = hashtags_for_profiles.id").
			Where("hashtags_for_profiles.hashtag IN (?)", hashtagValuesWithPrefix)

	}

	var count int64
	if err := query.Model(&models.Profile{}).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve data",
		})
	}

	limit := c.Query("limit", "10")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid limit parameter",
		})
	}

	err = utils.Paginate(c, query.Limit(limitInt).Find(&profiles), &profiles)
	if err != nil {
		return err
	}

	var userIds []uuid.UUID
	for _, profile := range profiles {
		userIds = append(userIds, profile.UserID)
	}

	// Retrieve streaming data for these user IDs
	var streamings []models.Streaming
	// if err := initializers.DB.Where("user_id IN (?)", userIds).Find(&profiles).Error; err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": "Could not retrieve streaming data",
	// 	})
	// }
	fmt.Println("users------------------")
	fmt.Println(count, city, userIds)

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   streamings,
		"meta": fiber.Map{
			"total": len(profiles),
			"limit": limitInt,
		},
	})

}

func GetAllProfile(c *fiber.Ctx) error {

	language := c.Query("language")

	var profiles []models.Profile
	query := initializers.DB.
		Preload("Guilds.Translations", "language = ?", language).
		Preload("Hashtags").
		Preload("City.Translations", "language = ?", language).
		Preload("Photos").
		Preload("User.Blogs").
		Preload("User.Blogs.Photos").
		Preload("User").
		Joins("JOIN users ON profiles.user_id = users.id").
		Order("Users.name ASC").
		Where("Users.filled = ?", true)
	// Get the query parameters
	city := c.Query("city")
	hashtags := c.Query("hashtag")
	category := c.Query("category")
	title := c.Query("title")
	money := c.Query("money")

	if money != "" && money != "all" {
		if strings.Contains(money, "-") {
			alphabeticRange := strings.Split(money, "-")
			if len(alphabeticRange) != 2 {
				return fmt.Errorf("invalid alphabetic range format")
			}

			lowerAlpha := strings.TrimSpace(alphabeticRange[0])
			upperAlpha := strings.TrimSpace(alphabeticRange[1])
			query = query.Where("(LOWER(SUBSTR(Firstname, 1, 1)) >= ? AND LOWER(SUBSTR(Firstname, 1, 1)) <= ?) AND (LOWER(SUBSTR(Lastname, 1, 1)) >= ? AND LOWER(SUBSTR(Lastname, 1, 1)) <= ?)", lowerAlpha, upperAlpha, lowerAlpha, upperAlpha)

		} else {
			alphabeticRange := strings.Split(money, "-")

			lowerAlpha := strings.TrimSpace(alphabeticRange[0])

			query = query.Where("LOWER(SUBSTR(Firstname, 1, 1)) = ?", lowerAlpha)
		}
	}

	if city != "" && city != "all" {
		// Сначала найдем city_id для указанного города и языка
		var cityTranslation models.CityTranslation
		initializers.DB.Where("name = ? AND language = ?", city, language).First(&cityTranslation)

		if cityTranslation.ID != 0 {
			subQuery := initializers.DB.Table("profiles_city").
				Select("profile_id").
				Where("city_id = ?", cityTranslation.CityID)

			// Добавим условие, чтобы ваш основной запрос включал только записи с blog_id из подзапроса
			query = query.Where("profiles.id IN (?)", subQuery) // Specify the table alias for "blogs.id"
		}
	}

	if title != "" && title != "all" {
		query = query.Where("LOWER(Descr) LIKE ?", "%"+title+"%")
	}

	if category != "" && category != "all" {
		var guildTranslation models.GuildTranslation
		initializers.DB.Where("name = ? AND language = ?", category, language).First(&guildTranslation)
		if guildTranslation.ID != 0 {
			// Создадим подзапрос для поиска всех blog_id, связанных с указанным guild_id
			subQuery := initializers.DB.Table("profiles_guilds").
				Select("profile_id").
				Where("guilds_id = ?", guildTranslation.GuildID)

			// Добавим условие, чтобы ваш основной запрос включил только записи с blog_id из подзапроса
			query = query.Where("profiles.id IN (?)", subQuery)
		}
	}

	if hashtags != "" && hashtags != "all" {
		// Split the hashtags into separate values
		hashtagValues := strings.Split(hashtags, ",")

		// Join the hashtag values with the '#' character
		hashtagValuesWithPrefix := make([]string, len(hashtagValues))
		for i, tag := range hashtagValues {
			hashtagValuesWithPrefix[i] = strings.TrimSpace(tag)
		}

		// Add the hashtags filter to the query
		query = query.Joins("JOIN profiles_hashtags ON profiles.id = profiles_hashtags.profile_id").
			Joins("JOIN hashtags_for_profiles ON profiles_hashtags.hashtags_for_profile_id = hashtags_for_profiles.id").
			Where("hashtags_for_profiles.hashtag IN (?)", hashtagValuesWithPrefix)

	}

	var count int64
	if err := query.Model(&models.Profile{}).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve data",
		})
	}

	limit := c.Query("limit", "10")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid limit parameter",
		})
	}

	err = utils.Paginate(c, query.Limit(limitInt).Find(&profiles), &profiles)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   profiles,
		"meta": fiber.Map{
			"total": count,
			"limit": limitInt,
		},
	})

}

func GetProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)
	language := c.Query("language")
	var profile models.Profile
	if err := initializers.DB.Preload("Guilds.Translations", "language = ?", language).Preload("Hashtags").Preload("City.Translations", "language = ?", language).Preload("Photos").Preload("User").First(&profile, "user_id = ?", user.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Profile not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve profile",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   profile,
	})
}

func GetProfileGuest(c *fiber.Ctx) error {

	type UserWithExtras struct {
		models.User
		HighestIsUpBlog models.Blog `json:"highestIsUpBlog"`
		TotalVotes      int         `json:"totalVotes"`
	}

	language := c.Query("language")
	var access_token string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		access_token = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("access_token") != "" {
		access_token = c.Cookies("access_token")
	}

	config, _ := initializers.LoadConfig(".")

	if access_token != "" && access_token != "undefined" {
		tokenClaims, err := utils.ValidateToken(access_token, config.AccessTokenPublicKey)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}

		name := c.Params("name")
		var profile models.User
		if err := initializers.DB.Preload("Followers").Preload("Followings").Preload("Followings.Followers").Preload("Profile.Guilds.Translations", "language = ?", language).Preload("Profile.Photos").Preload("Profile.Service").Preload("Profile.City.Translations", "language = ?", language).Preload("Profile.Hashtags").Preload("Blogs").Preload("Blogs.Photos").Preload("Blogs.Votes").Preload("Profile.Documents").First(&profile, "name = ?", name).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Profile not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve profile",
			})
		}

		var highestIsUpBlog models.Blog
		maxIsUpVotes := 0

		for _, blog := range profile.Blogs {
			isUpVotes := 0

			for _, vote := range blog.Votes {
				if vote.IsUP {
					isUpVotes++
				}
			}

			if blog.Status == "ACTIVE" && isUpVotes > maxIsUpVotes {
				maxIsUpVotes = isUpVotes
				highestIsUpBlog = blog
			}
		}

		if maxIsUpVotes == 0 && len(profile.Blogs) > 0 {
			highestIsUpBlog = profile.Blogs[len(profile.Blogs)-1]
		}

		userWithExtras := UserWithExtras{
			User:            removeDataFromProfile(profile),
			HighestIsUpBlog: highestIsUpBlog,
			TotalVotes:      maxIsUpVotes,
		}

		response := fiber.Map{
			"status": "success",
			"data":   userWithExtras,
		}

		// Convert UUID to string for comparison
		profileIDString := profile.ID.String()

		response["canFollow"] = true

		// Check if the profile UserID matches tokenClaims.UserID
		if tokenClaims.UserID != "" && tokenClaims.UserID != profileIDString {
			for _, following := range profile.Followings {
				if following.ID.String() == tokenClaims.UserID {
					// If the user is already following, set canFollow to false
					response["canFollow"] = false
					break
				}
			}
		} else {
			response["canFollow"] = false
		}

		return c.JSON(response)

	} else {

		name := c.Params("name")
		var profile models.User
		if err := initializers.DB.Preload("Followings").Preload("Followers").Preload("Profile.Guilds.Translations", "language = ?", language).Preload("Profile.Photos").Preload("Profile.Service").Preload("Profile.City.Translations", "language = ?", language).Preload("Profile.Hashtags").Preload("Blogs").Preload("Blogs.Photos").Preload("Blogs.Votes").Preload("Profile.Documents").First(&profile, "name = ?", name).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Profile not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve profile",
			})
		}

		var highestIsUpBlog models.Blog
		maxIsUpVotes := 0

		for _, blog := range profile.Blogs {
			isUpVotes := 0

			for _, vote := range blog.Votes {
				if vote.IsUP {
					isUpVotes++
				}
			}

			if blog.Status == "ACTIVE" && isUpVotes > maxIsUpVotes {
				maxIsUpVotes = isUpVotes
				highestIsUpBlog = blog
			}
		}

		if maxIsUpVotes == 0 && len(profile.Blogs) > 0 {
			highestIsUpBlog = profile.Blogs[len(profile.Blogs)-1]
		}

		userWithExtras := UserWithExtras{
			User:            removeDataFromProfile(profile),
			HighestIsUpBlog: highestIsUpBlog,
			TotalVotes:      maxIsUpVotes,
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   userWithExtras,
		})

	}

}

func removeDataFromProfile(p models.User) models.User {
	updatedProfile := models.User{}
	valueType := reflect.TypeOf(p)

	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		if field.Name != "Vote" && field.Name != "Blogs" {
			value := reflect.ValueOf(p).Field(i)
			reflect.ValueOf(&updatedProfile).Elem().FieldByName(field.Name).Set(value)
		}
	}
	return updatedProfile
}

func UpdateProfileAdditional(c *fiber.Ctx) error {
	type RequestBody struct {
		Additional string `json:"additional"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	user := c.Locals("user").(models.UserResponse)

	profile := models.Profile{}
	err := initializers.DB.Where("user_id = ?", user.ID).First(&profile).Error
	if err != nil {
		_ = err
		// Handle the error appropriately (e.g., return an error response)
	}

	// Fetch languages from the database
	var langs []models.Langs
	err = initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	translations := make(map[string]string)

	for _, lang := range langs {

		result, _ := gt.Translate(requestBody.Additional, profile.Lang, lang.Code)
		translations[lang.Code] = result

	}

	// Set the translated values in the TitleLangs field
	profile.MultilangAdditional.En = translations["en"]
	profile.MultilangAdditional.Ru = translations["ru"]
	profile.MultilangAdditional.Ka = translations["ka"]
	profile.MultilangAdditional.Es = translations["es"]

	// Update the "Additional" field in the profile
	profile.Additional = requestBody.Additional

	// Save the updated profile back to the database
	err = initializers.DB.Save(&profile).Error
	if err != nil {
		_ = err
		// Handle the error appropriately (e.g., return an error response)
	}

	// Return a success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

func UpdateBotProfileAdditional(c *fiber.Ctx) error {
	type RequestBody struct {
		UserId     string `json:"userid"`
		Additional string `json:"additional"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	profile := models.Profile{}
	err := initializers.DB.Where("user_id = ?", requestBody.UserId).First(&profile).Error
	if err != nil {
		_ = err
		// Handle the error appropriately (e.g., return an error response)
	}

	// Fetch languages from the database
	var langs []models.Langs
	err = initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	translations := make(map[string]string)

	for _, lang := range langs {

		result, _ := gt.Translate(requestBody.Additional, profile.Lang, lang.Code)
		translations[lang.Code] = result

	}

	// Set the translated values in the TitleLangs field
	profile.MultilangAdditional.En = translations["en"]
	profile.MultilangAdditional.Ru = translations["ru"]
	profile.MultilangAdditional.Ka = translations["ka"]
	profile.MultilangAdditional.Es = translations["es"]

	// Update the "Additional" field in the profile
	profile.Additional = requestBody.Additional

	// Save the updated profile back to the database
	err = initializers.DB.Save(&profile).Error
	if err != nil {
		_ = err
		// Handle the error appropriately (e.g., return an error response)
	}

	// Return a success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	type RequestBody struct {
		Firstname  string `json:"firstname"`
		Descr      string `json:"descr"`
		Tcid       int64  `json:"tcid"`
		Additional string `json:"additional"`
		City       []struct {
			ID uint64 `json:"id"`
		} `json:"city"`
		Guilds []struct {
			ID uint64 `json:"id"`
		} `json:"guilds"`
		Hashtags []struct {
			ID uint64 `json:"id"`
		} `json:"hashtags"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}
	user := c.Locals("user").(models.UserResponse)

	var profile models.Profile
	if err := initializers.DB.Preload("Guilds").Preload("Hashtags").Preload("City").Preload("Photos").First(&profile, "user_id = ?", user.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Profile not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve profile",
		})
	}

	// Fetch languages from the database
	var langs []models.Langs
	err := initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	translations := make(map[string]string)

	for _, lang := range langs {

		result, _ := gt.Translate(requestBody.Descr, profile.Lang, lang.Code)
		translations[lang.Code] = result

	}

	// Set the translated values in the TitleLangs field
	profile.MultilangDescr.En = translations["en"]
	profile.MultilangDescr.Ru = translations["ru"]
	profile.MultilangDescr.Ka = translations["ka"]
	profile.MultilangDescr.Es = translations["es"]

	// Create a new slice to store the updated list of cities
	updatedCities := []models.City{}

	// Iterate over the requestBody.City and create City objects from the IDs
	for _, cityID := range requestBody.City {
		city := models.City{
			ID: uint(cityID.ID),
		}
		updatedCities = append(updatedCities, city)
	}

	// Remove existing cities from profile.City that are not present in updatedCities
	existingCities := profile.City

	for i := len(existingCities) - 1; i >= 0; i-- {
		found := false
		for _, updatedCity := range updatedCities {
			if existingCities[i].ID == updatedCity.ID {
				found = true
				break
			}
		}
		if !found {
			profile.City = append(profile.City[:i], profile.City[i+1:]...)
		}
	}

	// Create a new slice to store the updated list of guilds
	updatedGuilds := []models.Guilds{}

	// Iterate over the requestBody.Guilds and create Guild objects from the IDs
	for _, guildID := range requestBody.Guilds {
		guild := models.Guilds{
			ID: uint(guildID.ID),
		}
		updatedGuilds = append(updatedGuilds, guild)
	}

	// Remove existing guilds from profile.Guilds that are not present in updatedGuilds
	existingGuilds := profile.Guilds

	for i := len(existingGuilds) - 1; i >= 0; i-- {
		found := false
		for _, updatedGuild := range updatedGuilds {
			if existingGuilds[i].ID == updatedGuild.ID {
				found = true
				break
			}
		}
		if !found {
			profile.Guilds = append(profile.Guilds[:i], profile.Guilds[i+1:]...)
		}
	}

	// Create a new slice to store the updated list of hashtags
	updatedHashtags := []models.HashtagsForProfile{}

	// Iterate over the requestBody.Hashtags and create HashtagsForProfile objects from the IDs
	for _, hashtagID := range requestBody.Hashtags {
		hashtag := models.HashtagsForProfile{
			ID: uint(hashtagID.ID),
		}
		updatedHashtags = append(updatedHashtags, hashtag)
	}

	// Remove existing hashtags from profile.Hashtags that are not present in updatedHashtags
	existingHashtags := profile.Hashtags

	for i := len(existingHashtags) - 1; i >= 0; i-- {
		found := false
		for _, updatedHashtag := range updatedHashtags {
			if existingHashtags[i].ID == updatedHashtag.ID {
				found = true
				break
			}
		}
		if !found {
			profile.Hashtags = append(profile.Hashtags[:i], profile.Hashtags[i+1:]...)
		}
	}

	// Assign firsstname, lastname, and middlen values from the requestBody to profile object
	profile.Firstname = requestBody.Firstname
	// profile.Lastname = requestBody.Lastname
	// profile.MiddleN = requestBody.MiddleN
	profile.Descr = requestBody.Descr

	// Save the updated profile to the database
	if err := initializers.DB.Save(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update profile",
		})
	}

	// Update the city associations in the database
	if err := initializers.DB.Model(&profile).Association("City").Replace(updatedCities); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update city associations",
		})
	}

	// Update the guild associations in the database
	if err := initializers.DB.Model(&profile).Association("Guilds").Replace(updatedGuilds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update guild associations",
		})
	}

	// Update the hashtags associations in the database
	if err := initializers.DB.Model(&profile).Association("Hashtags").Replace(updatedHashtags); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update hashtags associations",
		})
	}

	var newUser models.User

	// Find the corresponding user record based on user.ID
	if err := initializers.DB.First(&newUser, "id = ?", user.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve user record",
		})
	}

	// Update the filled field of the user record to true
	newUser.Filled = true

	// Save the changes to the user record in the database
	if err := initializers.DB.Save(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update user record",
		})
	}

	// The user record has been successfully updated with filled = true
	fmt.Println("User record has been updated successfully")

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   profile,
	})
}

func UpdateBotProfile(c *fiber.Ctx) error {
	type RequestBody struct {
		UserId     string `json:"userid"`
		Firstname  string `json:"firstname"`
		Descr      string `json:"descr"`
		Tcid       int64  `json:"tcid"`
		Additional string `json:"additional"`
		City       []struct {
			ID uint64 `json:"id"`
		} `json:"city"`
		Guilds []struct {
			ID uint64 `json:"id"`
		} `json:"guilds"`
		Hashtags []struct {
			ID uint64 `json:"id"`
		} `json:"hashtags"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	var profile models.Profile
	if err := initializers.DB.Preload("Guilds").Preload("Hashtags").Preload("City").Preload("Photos").First(&profile, "user_id = ?", requestBody.UserId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Profile not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve profile",
		})
	}

	// Fetch languages from the database
	var langs []models.Langs
	err := initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	translations := make(map[string]string)

	for _, lang := range langs {

		result, _ := gt.Translate(requestBody.Descr, profile.Lang, lang.Code)
		translations[lang.Code] = result

	}

	// Set the translated values in the TitleLangs field
	profile.MultilangDescr.En = translations["en"]
	profile.MultilangDescr.Ru = translations["ru"]
	profile.MultilangDescr.Ka = translations["ka"]
	profile.MultilangDescr.Es = translations["es"]

	// Create a new slice to store the updated list of cities
	updatedCities := []models.City{}

	// Iterate over the requestBody.City and create City objects from the IDs
	for _, cityID := range requestBody.City {
		city := models.City{
			ID: uint(cityID.ID),
		}
		updatedCities = append(updatedCities, city)
	}

	// Remove existing cities from profile.City that are not present in updatedCities
	existingCities := profile.City

	for i := len(existingCities) - 1; i >= 0; i-- {
		found := false
		for _, updatedCity := range updatedCities {
			if existingCities[i].ID == updatedCity.ID {
				found = true
				break
			}
		}
		if !found {
			profile.City = append(profile.City[:i], profile.City[i+1:]...)
		}
	}

	// Create a new slice to store the updated list of guilds
	updatedGuilds := []models.Guilds{}

	// Iterate over the requestBody.Guilds and create Guild objects from the IDs
	for _, guildID := range requestBody.Guilds {
		guild := models.Guilds{
			ID: uint(guildID.ID),
		}
		updatedGuilds = append(updatedGuilds, guild)
	}

	// Remove existing guilds from profile.Guilds that are not present in updatedGuilds
	existingGuilds := profile.Guilds

	for i := len(existingGuilds) - 1; i >= 0; i-- {
		found := false
		for _, updatedGuild := range updatedGuilds {
			if existingGuilds[i].ID == updatedGuild.ID {
				found = true
				break
			}
		}
		if !found {
			profile.Guilds = append(profile.Guilds[:i], profile.Guilds[i+1:]...)
		}
	}

	// Create a new slice to store the updated list of hashtags
	updatedHashtags := []models.HashtagsForProfile{}

	// Iterate over the requestBody.Hashtags and create HashtagsForProfile objects from the IDs
	for _, hashtagID := range requestBody.Hashtags {
		hashtag := models.HashtagsForProfile{
			ID: uint(hashtagID.ID),
		}
		updatedHashtags = append(updatedHashtags, hashtag)
	}

	// Remove existing hashtags from profile.Hashtags that are not present in updatedHashtags
	existingHashtags := profile.Hashtags

	for i := len(existingHashtags) - 1; i >= 0; i-- {
		found := false
		for _, updatedHashtag := range updatedHashtags {
			if existingHashtags[i].ID == updatedHashtag.ID {
				found = true
				break
			}
		}
		if !found {
			profile.Hashtags = append(profile.Hashtags[:i], profile.Hashtags[i+1:]...)
		}
	}

	// Assign firsstname, lastname, and middlen values from the requestBody to profile object
	profile.Firstname = requestBody.Firstname
	// profile.Lastname = requestBody.Lastname
	// profile.MiddleN = requestBody.MiddleN
	profile.Descr = requestBody.Descr

	// Save the updated profile to the database
	if err := initializers.DB.Save(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update profile",
		})
	}

	// Update the city associations in the database
	if err := initializers.DB.Model(&profile).Association("City").Replace(updatedCities); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update city associations",
		})
	}

	// Update the guild associations in the database
	if err := initializers.DB.Model(&profile).Association("Guilds").Replace(updatedGuilds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update guild associations",
		})
	}

	// Update the hashtags associations in the database
	if err := initializers.DB.Model(&profile).Association("Hashtags").Replace(updatedHashtags); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update hashtags associations",
		})
	}

	var newUser models.User

	// Find the corresponding user record based on requestBody.UserId
	if err := initializers.DB.First(&newUser, "id = ?", requestBody.UserId).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve user record",
		})
	}

	// Update the filled field of the user record to true
	newUser.Filled = true

	// Save the changes to the user record in the database
	if err := initializers.DB.Save(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update user record",
		})
	}

	// The user record has been successfully updated with filled = true
	fmt.Println("User record has been updated successfully")

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   profile,
	})
}

func GetDocuments(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)
	// Retrieve all documents for the specified user profile ID
	var documents []models.ProfileDocuments
	err := initializers.DB.Where("profile_id = ?", user.Profile[0].ID).Find(&documents).Error
	if err != nil {
		// Handle the error if any
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve documents",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   documents,
	})

}

func UpdateProfileDocuments(c *fiber.Ctx) error {

	type File struct {
		Filename string `json:"filename"`
	}

	type RequestBody struct {
		ID           int    `json:"ID"`
		Name         string `json:"name"`
		Organization string `json:"organization"`
		Specified    string `json:"specified"`
		Year         int    `json:"year"`
		Descr        string `json:"descr"`
		Files        []File `json:"files"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	// Serialize the Files field as JSON
	filesJSON, err := json.Marshal(requestBody.Files)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to serialize profile document files",
		})
	}

	// Find the existing document in the database by its ID
	documentID := requestBody.ID
	var existingDocument models.ProfileDocuments
	if err := initializers.DB.First(&existingDocument, documentID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to find the profile document",
		})
	}

	// Update the document fields with the new values
	existingDocument.Name = requestBody.Name
	existingDocument.Organization = requestBody.Organization
	existingDocument.Descr = requestBody.Descr
	// existingDocument.Additional = requestBody.Additional

	existingDocument.Files = pgtype.JSONB{Bytes: filesJSON, Status: pgtype.Present}

	// Save the updated document to the database
	if err := initializers.DB.Save(&existingDocument).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update profile document",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success updated",
		"data":   "ok",
	})
}

func DeleteProfileDocuments(c *fiber.Ctx) error {
	// Retrieve the document ID from the URL route parameter
	documentID := c.Params("id")

	// Find the existing document in the database by its ID
	var existingDocument models.ProfileDocuments
	if err := initializers.DB.First(&existingDocument, documentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Document not found, return appropriate error response
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Profile document not found",
			})
		}

		// Error occurred while finding the document, return internal server error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to find the profile document",
		})
	}

	// Convert the Files field to a JSONB value
	filesJSON, err := json.Marshal(existingDocument.Files)
	if err != nil {
		// Handle the error if the conversion fails
		// For example, you can return an error response
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error converting files to JSON",
		})
	}

	var filesJSONB pgtype.JSONB
	if err := filesJSONB.Set(filesJSON); err != nil {
		// Handle the error if the conversion fails
		// For example, you can return an error response
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Error converting files to JSONB",
		})
	}

	// Delete the document from the database
	if err := initializers.DB.Delete(&existingDocument).Error; err != nil {
		// Error occurred while deleting the document, return internal server error response
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete profile document",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "ok",
	})
}

func NewProfileDocuments(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	type File struct {
		Filename string `json:"filename"`
	}

	type RequestBody struct {
		Name         string `json:"name"`
		Organization string `json:"organization"`
		Specified    string `json:"specified"`
		Year         int    `json:"year"`
		Descr        string `json:"descr"`
		Files        []File `json:"files"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	// Serialize the Files field as JSON
	filesJSON, err := json.Marshal(requestBody.Files)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to serialize profile document files",
		})
	}

	document := models.ProfileDocuments{
		ProfileID:    user.Profile[0].ID,
		Name:         requestBody.Name,
		Organization: requestBody.Organization,
		Descr:        requestBody.Descr,
		Files:        pgtype.JSONB{Bytes: filesJSON, Status: pgtype.Present},
	}

	// Save the document to the database using your preferred database ORM or query builder
	// For example, using GORM:
	if err := initializers.DB.Create(&document).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to save profile document",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   document,
	})

}

func UpdateProfilePhotos(c *fiber.Ctx) error {

	isUpdate := c.Query("update")
	user := c.Locals("user").(models.UserResponse)

	if isUpdate == "true" {

		var updatePhotos models.ProfilePhoto

		if err := c.BodyParser(&updatePhotos); err != nil {
			// Handle parsing error
			return err
		}

		var existingPhoto models.ProfilePhoto
		err := initializers.DB.Where("profile_id = ?", updatePhotos.ProfileID).First(&existingPhoto).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Handle the case when the profile photo does not exist
				return fmt.Errorf("profile photo not found for ProfileID: %d", updatePhotos.ProfileID)
			}
			// Handle other database errors
			return err
		}

		existingFiles := existingPhoto.Files

		existingPhoto.Files = updatePhotos.Files
		err = initializers.DB.Save(&existingPhoto).Error
		if err != nil {
			// Handle the error when updating the record
			return err
		}

		deleteRemovedFiles(existingFiles, updatePhotos.Files)

		return c.JSON(existingPhoto)

	}

	var files []struct {
		Path string `json:"path"`
	}

	if err := c.BodyParser(&files); err != nil {
		// Handle parsing error
		return err
	}

	type File struct {
		Path string `json:"path"`
	}

	// Iterate over the files
	for _, file := range files {
		var existingPhoto models.ProfilePhoto
		err := initializers.DB.Where("profile_id = ?", user.Profile[0].ID).First(&existingPhoto).Error
		if err == gorm.ErrRecordNotFound {
			// Photo does not exist, add a new record
			newPhoto := models.ProfilePhoto{
				ProfileID: user.Profile[0].ID,
			}

			// Append the file path to the existing Files field
			var existingFiles []File
			if err := existingPhoto.Files.AssignTo(&existingFiles); err != nil {
				// Handle the error when assigning existing files
				fmt.Println("Error assigning existing files:", err)
			}
			existingFiles = append(existingFiles, File{Path: file.Path})
			newFiles, err := json.Marshal(existingFiles)
			if err != nil {
				// Handle the error when marshaling new files
				fmt.Println("Error marshaling new files:", err)
			}
			newPhoto.Files.Set(newFiles)

			err = initializers.DB.Create(&newPhoto).Error
			if err != nil {
				// Handle the error when creating a new record
				fmt.Println("Error creating new profile photo:", err)
			} else {
				fmt.Println("Added new profile photo:", newPhoto.ID)
			}
		} else if err != nil {
			// Handle other database errors
			fmt.Println("Error querying profile photos:", err)
		} else {
			// Photo already exists, update the record if needed
			// Append the file path to the existing Files field
			var existingFiles []File
			if err := existingPhoto.Files.AssignTo(&existingFiles); err != nil {
				// Handle the error when assigning existing files
				fmt.Println("Error assigning existing files:", err)
			}
			existingFiles = append(existingFiles, File{Path: file.Path})
			newFiles, err := json.Marshal(existingFiles)
			if err != nil {
				// Handle the error when marshaling new files
				fmt.Println("Error marshaling new files:", err)
			}
			existingPhoto.Files.Set(newFiles)

			err = initializers.DB.Save(&existingPhoto).Error
			if err != nil {
				// Handle the error when updating the record
				fmt.Println("Error updating profile photo:", err)
			} else {
				fmt.Println("Updated existing profile photo:", existingPhoto.ID)
			}
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "photo gallery",
	})
}

func AddHashTagProfile(c *fiber.Ctx) error {

	var hashtag models.HashtagsForProfile
	if err := c.BodyParser(&hashtag); err != nil {
		return err
	}

	// Save the hashtag to the database
	if err := initializers.DB.Create(&hashtag).Error; err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   hashtag,
	})
}

func SearchHashTagProfile(c *fiber.Ctx) error {

	// Get the name query parameter
	name := c.Query("name")
	// Check if the name is provided
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Name parameter is required",
		})
	}

	// Find the cities with names similar to the search query (case-insensitive)
	var hashtags []models.HashtagsForProfile
	if err := initializers.DB.Where("hashtag ILIKE ?", "%"+name+"%").Find(&hashtags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from database",
		})
	}

	// Return the matched cities as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   hashtags,
	})
}

func Get10RandomTags(c *fiber.Ctx) error {
	var hashtags []models.HashtagsForProfile
	if err := initializers.DB.Order("RANDOM()").Limit(10).Find(&hashtags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch random tags from database",
		})
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   hashtags,
	})
}

func SendDonat(c *fiber.Ctx) error {
	// Проверка авторизации пользователя
	IDUSER := c.Locals("user")
	if IDUSER == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	userResp, ok := IDUSER.(models.UserResponse)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user type",
		})
	}

	// Чтение тела запроса
	type DonatRequest struct {
		Author string `json:"author"`
		Amount string `json:"amount"`
		Sms    string `json:"sms"`
	}

	var donatReq DonatRequest
	if err := c.BodyParser(&donatReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to parse request body",
		})
	}

	// Конвертация суммы в float
	priceFloat, err := strconv.ParseFloat(donatReq.Amount, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid amount",
		})
	}

	// Получение текущего баланса пользователя
	var billing models.Billing
	err = initializers.DB.Where("user_id = ?", userResp.ID).First(&billing).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch balance",
		})
	}

	// Проверка достаточности баланса
	if billing.Amount < priceFloat {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Insufficient balance",
		})
	}

	// Обновление баланса пользователя
	err = initializers.DB.Model(&models.Billing{}).
		Where("user_id = ?", userResp.ID).
		Updates(map[string]interface{}{
			"amount": gorm.Expr("amount - ?", priceFloat),
		}).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update balance",
		})
	}

	// Создание транзакции
	transaction := models.Transaction{
		UserID:      userResp.ID,
		Total:       strconv.FormatFloat(priceFloat, 'f', 2, 64),
		Amount:      priceFloat,
		Description: "Донат пользователю " + donatReq.Author,
		Module:      "donat",
		Type:        "deduction",
		Status:      "CLOSED_1",
	}

	err = initializers.DB.Create(&transaction).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create transaction",
		})
	}

	// Получение ID автора стрима
	var author models.User
	err = initializers.DB.Where("name = ?", donatReq.Author).First(&author).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch author ID",
		})
	}

	// Увеличение баланса автора стрима
	var authorBilling models.Billing
	err = initializers.DB.Where("name = ?", donatReq.Author).First(&authorBilling).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch author balance",
		})
	}

	// Увеличение баланса автора стрима
	err = initializers.DB.Model(&models.Billing{}).
		Where("user_id = ?", author.ID).
		Updates(map[string]interface{}{
			"amount": gorm.Expr("amount + ?", priceFloat),
		}).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update author balance",
		})
	}

	// Создание транзакции для автора стрима
	transactionReceiver := models.Transaction{
		UserID:      author.ID,
		Total:       strconv.FormatFloat(priceFloat, 'f', 2, 64),
		Amount:      priceFloat,
		Description: "Получение доната от пользователя " + userResp.Name,
		Module:      "donat",
		Type:        "addition",
		Status:      "CLOSED_1",
	}

	err = initializers.DB.Create(&transactionReceiver).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create transaction for receiver",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "donat",
	})
}

func UpdateProfileStreaming(c *fiber.Ctx) error {

	var streaming models.Streaming
	var profile models.Profile

	if err := c.BodyParser(&streaming); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	if err := initializers.DB.First(&profile, "user_id = ?", streaming.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to find the profile document",
		})
	}

	if len(profile.Streaming) == 1 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Can't create streaming more than one",
		})
	} else {
		profile.Streaming = append(profile.Streaming, streaming)
	}
	// Save the updated document to the database

	// if err := initializers.DB.Clauses(clause.OnConflict{
	// 	Columns: []clause.Column{{Name: "user_id"}, {Name: "room_id"}}, // specify the columns that make up the unique constraint
	// 	// if the above yields issues, use Constraint directly by naming it like - (commenting the above line)
	// 	// Constraint: "unique_user_room",
	// 	DoUpdates: clause.AssignmentColumns([]string{"title", "created_at"}), // specify the columns to be updated on conflict
	// }).Create(&streaming).Error; err != nil {
	// 	log.Printf("Error saving user data: %v", err)
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": fmt.Sprintf("Failed to save user data: %v", err),
	// 	})
	// }
	if err := initializers.DB.Save(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Failed to save user data: %v", err),
		})
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "ok",
	})
}

func DeleteProfileStreaming(c *fiber.Ctx) error {
	roomID := c.Params("id")
	type RequestData struct {
		UserID    string    `json:"userID"`
		DeletedAt time.Time `json:"time"`
	}
	requestData := new(RequestData)
	if err := c.BodyParser(requestData); err != nil {
		// Handle parsing error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
	}
	fmt.Println("---------------------------------------------")
	fmt.Println(requestData)
	var profile models.Profile
	if err := initializers.DB.First(&profile, "user_id = ?", requestData.UserID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
	}
	fmt.Println("--------------------Profile---------------")
	fmt.Println(profile.Streaming, requestData.UserID, roomID)

	var onlineStreamingHours time.Duration = 0
	var updatedStreamings []models.Streaming
	for _, stream := range profile.Streaming {
		if stream.RoomID != roomID {
			updatedStreamings = append(updatedStreamings, stream)
		} else {
			onlineStreamingHours = requestData.DeletedAt.Sub(stream.CreatedAt)
		}
	}
	var user models.User
	if err := initializers.DB.First(&user, "ID = ?", requestData.UserID).Error; err != nil {
		return err
	}

	var totalStreamingHours models.TimeEntryScanner
	totalStreamingHours = user.TotalOnlineStreamingHours

	additionalHours := int(onlineStreamingHours.Hours())
	additionalMinutes := int(onlineStreamingHours.Minutes()) % 60
	additionalSeconds := int(onlineStreamingHours.Seconds()) % 60

	if len(totalStreamingHours) > 0 {
		totalStreamingHours[0].Seconds += additionalSeconds
		totalStreamingHours[0].Minutes += additionalMinutes + totalStreamingHours[0].Seconds/60
		totalStreamingHours[0].Seconds = totalStreamingHours[0].Seconds % 60
		totalStreamingHours[0].Hour += additionalHours + totalStreamingHours[0].Minutes/60
		totalStreamingHours[0].Minutes = totalStreamingHours[0].Minutes % 60
	} else {
		// In case there is no entry, create the first one
		totalStreamingHours = append(totalStreamingHours, models.TimeEntry{
			Hour:    additionalHours,
			Minutes: additionalMinutes,
			Seconds: additionalSeconds,
		})
	}

	user.TotalOnlineStreamingHours = totalStreamingHours

	if err := initializers.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Failed to save user data: %v", err),
		})
	}

	profile.Streaming = updatedStreamings
	if err := initializers.DB.Save(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
	}
	// if err := initializers.DB.Delete(&models.Streaming{}, "room_id = ? AND user_id = ?", roomID, requestData.UserID).Error; err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": fmt.Sprintf("Failed to delete streaming entry: %v", err),
	// 	})
	// }

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   profile.Streaming,
	})
}
