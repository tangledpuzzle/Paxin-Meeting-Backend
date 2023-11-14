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
	mode := c.Query("mode")
	if mode == "translate" {
		var guildTranslation models.GuildTranslation
		if err := initializers.DB.
			Where("name ILIKE ?", "%"+name+"%").
			First(&guildTranslation).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to fetch guild translation from the database",
			})
		}

		var translatedGuild models.GuildTranslation
		if err := initializers.DB.
			Where("guild_id = ? AND language = ?", guildTranslation.GuildID, lang). // Replace "EN" with your target language code
			First(&translatedGuild).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to fetch translated guild name from the database",
			})
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   translatedGuild.Name,
		})

	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "",
	})
}

func GetGuildsAdmin(c *fiber.Ctx) error {
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

	// get count of all cities in the database
	var total int64
	if err := initializers.DB.Model(&models.Guilds{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch guilds count from the database",
		})
	}

	// get paginated city names from the database with translations
	var guilds []models.Guilds
	db := initializers.DB.
		Joins("JOIN guild_translations ON guilds.id = guild_translations.guild_id").
		Preload("Translations").
		Offset(skipNumber).Limit(limitNumber).
		Find(&guilds)

	if db.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch paginated cities from the database",
			"error":   db.Error.Error(),
		})
	}

	// fmt.Println(db.Statement.SQL.String()) // Log the generated SQL

	// return the paginated city names and metadata as a JSON response
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
