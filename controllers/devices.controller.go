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

	if newDevice.Device == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Device field cannot be empty",
		})
	}

	var existingDevice models.DevicesIOS
	if err := initializers.DB.Where("device = ?", newDevice.Device).First(&existingDevice).Error; err == nil {
		// Устройство уже существует в базе данных
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  "error",
			"message": "Device already exists",
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
