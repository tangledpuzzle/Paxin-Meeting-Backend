package controllers

import (
	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
)

func GetCities(c *fiber.Ctx) error {
	// get all city names from the database
	var cityNames []string
	if err := initializers.DB.Model(&models.City{}).Pluck("name", &cityNames).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch city names from database",
		})
	}

	// return the city names as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   cityNames,
	})
}

func GetName(c *fiber.Ctx) error {
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
	var cities []models.City
	if err := initializers.DB.Where("name ILIKE ?", "%"+name+"%").Find(&cities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from database",
		})
	}
	// var cityNames []string
	// for _, city := range cities {
	// 	cityNames = append(cityNames, city.Name)
	// }


	// Return the matched cities as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   cities,
	})
}
