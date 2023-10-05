package controllers

import (
	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
)

func GetStations(c *fiber.Ctx) error {
	// get all city names from the database
	var stationsNames []string
	if err := initializers.DB.Model(&models.Stations{}).Pluck("name", &stationsNames).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch city names from database",
		})
	}

	// return the city names as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stationsNames,
	})
}

func GetNameStation(c *fiber.Ctx) error {
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
	var stations []models.Stations
	if err := initializers.DB.Where("name ILIKE ?", "%"+name+"%").Find(&stations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from database",
		})
	}

	// Return the matched cities as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stations,
	})
}
