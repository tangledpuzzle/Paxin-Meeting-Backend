package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TranslationName struct {
	Name string
}

type GuildResponse struct {
	ID        uint
	Hex       string
	UpdatedAt time.Time
	DeletedAt *time.Time
	Name      string // Добавьте поле Name
}

func GetGuildName(c *fiber.Ctx) error {
	name := c.Query("name")
	lang := c.Query("lang")

	// Get query parameters for pagination
	limit := c.Query("limit", "10")
	skip := c.Query("skip", "0")

	limitNumber, err := strconv.Atoi(limit)
	if err != nil || limitNumber < 1 {
		limitNumber = 10
	}

	skipNumber, err := strconv.Atoi(skip)
	if err != nil || skipNumber < 0 {
		skipNumber = 0
	}

	var guilds []models.Guilds
	if err := initializers.DB.
		Joins("JOIN guild_translations ON guilds.id = guild_translations.guild_id").
		Preload("Translations", "language = ?", lang).
		Where("guild_translations.name ILIKE ? AND guild_translations.language = ?", "%"+name+"%", lang).
		Offset(skipNumber).Limit(limitNumber).
		Find(&guilds).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from the database",
		})
	}

	// Get total count of matching cities
	var total int64
	if err := initializers.DB.
		Model(&models.Guilds{}).
		Joins("JOIN guild_translations ON guilds.id = guild_translations.guild_id").
		Where("guild_translations.name ILIKE ? AND guild_translations.language = ?", "%"+name+"%", lang).
		Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch total count from the database",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   guilds,
		"meta": fiber.Map{
			"limit": limitNumber,
			"skip":  skipNumber,
			"total": total,
		},
	})

}

func GetGuilds(c *fiber.Ctx) error {
	// Получите запрошенный язык из параметров запроса
	language := c.Query("language") // Например, "en" для английского

	if language == "" {
		language = "en"
	}

	var guilds []models.Guilds
	err := initializers.DB.Find(&guilds).Error
	if err != nil {
		return err
	}

	// Получите переводы только для указанного языка
	var translations []models.GuildTranslation
	err = initializers.DB.Where("language = ?", language).Find(&translations).Error
	if err != nil {
		return err
	}

	// Создайте карту с переводами на указанном языке
	translationMap := make(map[uint]string)
	for _, translation := range translations {
		translationMap[translation.GuildID] = translation.Name
	}

	// Создайте JSON-ответ с переводами, если они доступны, или используйте основные значения
	var response []map[string]interface{}
	for _, guild := range guilds {
		translatedName, ok := translationMap[guild.ID]
		if ok {
			response = append(response, map[string]interface{}{
				"ID":        guild.ID,
				"Hex":       guild.Hex,
				"UpdatedAt": guild.UpdatedAt,
				"DeletedAt": guild.DeletedAt,
				"Name":      translatedName,
			})
		} else {
			response = append(response, map[string]interface{}{
				"ID":        guild.ID,
				"Hex":       guild.Hex,
				"UpdatedAt": guild.UpdatedAt,
				"DeletedAt": guild.DeletedAt,
			})
		}
	}

	// Верните данные в JSON-ответе
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   response,
	})
}

func GetGuildsAll(c *fiber.Ctx) error {
	// Get query parameters for pagination
	limit := c.Query("limit", "10")
	skip := c.Query("skip", "0") // Use skip directly from the query parameters

	limitNumber, err := strconv.Atoi(limit)
	if err != nil || limitNumber < 1 {
		limitNumber = 10
	}

	skipNumber, err := strconv.Atoi(skip)
	if err != nil || skipNumber < 0 {
		skipNumber = 0
	}

	// get count of all guilds in the database
	var total int64
	if err := initializers.DB.Model(&models.Guilds{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch guilds count from the database",
		})
	}

	// get paginated guild names from the database with translations
	var guilds []models.Guilds
	db := initializers.DB.
		Joins("JOIN guild_translations ON guilds.id = guild_translations.guild_id").
		Preload("Translations").
		Select("DISTINCT guilds.id, guilds.hex, guilds.updated_at, guilds.deleted_at").
		Offset(skipNumber).Limit(limitNumber).
		Find(&guilds)

	if db.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch paginated guilds from the database",
			"error":   db.Error.Error(),
		})
	}

	// return the paginated guild names and metadata as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   guilds,
		"meta": fiber.Map{
			"limit": limitNumber,
			"skip":  skipNumber,
			"total": total,
		},
	})
}

func CreateGuild(c *fiber.Ctx) error {
	var newGuild models.Guilds
	if err := c.BodyParser(&newGuild); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	newGuild.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&newGuild).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new guild",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New guild added successfully",
		"data":    newGuild,
	})
}

func DeleteGuild(c *fiber.Ctx) error {
	guildID := c.Params("id")

	if guildID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Guild ID is required",
		})
	}

	var guild models.Guilds
	if err := initializers.DB.First(&guild, guildID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Guild not found",
		})
	}

	// Find and delete all guild translations associated with the guild
	if err := initializers.DB.Where("guild_id = ?", guild.ID).Delete(&models.GuildTranslation{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete guild translations",
		})
	}

	if err := initializers.DB.Delete(&guild).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete the guild",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Guild and associated translations deleted successfully",
		"data":    nil,
	})
}

func UpdateGuild(c *fiber.Ctx) error {
	guildID := c.Params("id")

	if guildID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Guild ID is required",
		})
	}

	var guild models.Guilds
	if err := initializers.DB.First(&guild, guildID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Guild not found",
		})
	}

	var updatedGuild models.Guilds
	if err := c.BodyParser(&updatedGuild); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	if err := initializers.DB.Model(&guild).Updates(&updatedGuild).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update the guild",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Guild updated successfully",
		"data":    guild,
	})
}

func CreateGuildTranslation(c *fiber.Ctx) error {
	var newTranslation models.GuildTranslation

	if err := c.BodyParser(&newTranslation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	// Check if CityID is provided
	if newTranslation.GuildID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "guild ID is required",
		})
	}

	// You may want to add additional validation or checks here

	// Use newTranslation.CityID to retrieve the city from the database
	var guild models.Guilds
	if err := initializers.DB.First(&guild, newTranslation.GuildID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Guild not found",
		})
	}

	// Assign the retrieved City ID to the translation
	newTranslation.GuildID = guild.ID

	// Add the new translation to the database
	if err := initializers.DB.Create(&newTranslation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new translation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New translation added successfully",
		"data":    newTranslation,
	})
}

func UpdateGuildTranslation(c *fiber.Ctx) error {
	translationID := c.Query("translationID")

	if translationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation ID is required",
		})
	}

	var translation models.GuildTranslation
	if err := initializers.DB.First(&translation, translationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation not found",
		})
	}

	var updatedTranslation models.GuildTranslation
	if err := c.BodyParser(&updatedTranslation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	if err := initializers.DB.Model(&translation).Updates(&updatedTranslation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update the translation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Translation updated successfully",
		"data":    translation,
	})
}

func DeleteGuildTranslation(c *fiber.Ctx) error {
	translationID := c.Query("translationID")

	if translationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation ID is required",
		})
	}

	var translation models.GuildTranslation
	if err := initializers.DB.First(&translation, translationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation not found",
		})
	}

	if err := initializers.DB.Delete(&translation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete the translation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Translation deleted successfully",
		"data":    nil,
	})
}
