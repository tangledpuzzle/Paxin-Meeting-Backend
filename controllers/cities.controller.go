package controllers

import (
	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
)

func GetCities(c *fiber.Ctx) error {
	// get all city names from the database with translations
	var cities []models.City
	if err := initializers.DB.
		Joins("JOIN city_translations ON cities.id = city_translations.city_id").
		Preload("Translations"). // Preload all translations
		Find(&cities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from the database",
		})
	}

	// return the city names as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   cities,
	})
}

func GetName(c *fiber.Ctx) error {
	// Get the name query parameter
	// Get the name and lang query parameters
	name := c.Query("name")
	lang := c.Query("lang")
	mode := c.Query("mode")

	if mode == "translate" {
		var cityTranslation models.CityTranslation
		if err := initializers.DB.
			Where("name ILIKE ?", "%"+name+"%").
			First(&cityTranslation).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to fetch city translation from the database",
			})
		}

		var translatedCity models.CityTranslation
		if err := initializers.DB.
			Where("city_id = ? AND language = ?", cityTranslation.CityID, lang). // Replace "EN" with your target language code
			First(&translatedCity).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to fetch translated city name from the database",
			})
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   translatedCity.Name,
		})

	}
	// Check if the name and lang parameters are provided
	if name == "" || lang == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Both 'name' and 'lang' parameters are required",
		})
	}

	var cities []models.City
	if err := initializers.DB.
		Joins("JOIN city_translations ON cities.id = city_translations.city_id").
		Preload("Translations", "language = ?", lang).
		Where("city_translations.name ILIKE ? AND city_translations.language = ?", "%"+name+"%", lang).
		Find(&cities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from the database",
		})
	}

	// Return the matched cities as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   cities,
	})
}
