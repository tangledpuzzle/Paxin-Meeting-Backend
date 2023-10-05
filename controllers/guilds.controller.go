package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func GetGuilds(c *fiber.Ctx) error {
	// // get all city names from the database
	// var Guilds []string
	// if err := initializers.DB.Find(&models.Guilds{}).Error; err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": "Could not fetch city names from database",
	// 	})
	// }
	var guilds []models.Guilds
	err := initializers.DB.Find(&guilds).Error
	if err != nil {
		return err
	}

	// Return the guilds as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   guilds,
	})
}