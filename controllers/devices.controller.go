package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateDevice(c *fiber.Ctx) error {
	var newDevice models.DevicesIOS
	if err := c.BodyParser(&newDevice); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	newDevice.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&newDevice).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new Device",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New device added successfully",
		"data":    newDevice,
	})
}
