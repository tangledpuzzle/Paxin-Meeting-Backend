package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func GetBaseSystemData(c *fiber.Ctx) error {
	var notification models.System

	if err := initializers.DB.First(&notification).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Record not found",
		})
	}

	return c.JSON(notification)
}
