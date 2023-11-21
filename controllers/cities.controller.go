package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
)

func GetCities(c *fiber.Ctx) error {
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
	if err := initializers.DB.Model(&models.City{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities count from the database",
		})
	}

	// get paginated city names from the database with translations
	var cities []models.City
	db := initializers.DB.
		Joins("JOIN city_translations ON cities.id = city_translations.city_id").
		Preload("Translations").
		Offset(skipNumber).Limit(limitNumber).
		Find(&cities)

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
		"data":   cities,
		"meta": fiber.Map{
			"limit": limitNumber,
			"skip":  skipNumber,
			"total": total,
		},
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

func CreateCity(c *fiber.Ctx) error {
	// Parse request data into a City struct
	var newCity models.City
	if err := c.BodyParser(&newCity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	newCity.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&newCity).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new city",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New city added successfully",
		"data":    newCity,
	})
}

func DeleteCity(c *fiber.Ctx) error {
	cityID := c.Params("id")

	if cityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "City ID is required",
		})
	}

	var city models.City
	if err := initializers.DB.First(&city, cityID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "City not found",
		})
	}

	if err := initializers.DB.Delete(&city).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete the city",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "City deleted successfully",
		"data":    nil,
	})
}

func UpdateCity(c *fiber.Ctx) error {

	cityID := c.Params("id")

	if cityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "City ID is required",
		})
	}

	var city models.City
	if err := initializers.DB.First(&city, cityID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "City not found",
		})
	}

	var updatedCity models.City
	if err := c.BodyParser(&updatedCity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	if err := initializers.DB.Model(&city).Updates(&updatedCity).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update the city",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "City updated successfully",
		"data":    city,
	})
}
