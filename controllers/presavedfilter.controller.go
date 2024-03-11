package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func CreatePresavedfilter(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	filter := new(models.Presavedfilters)
	if err := c.BodyParser(filter); err != nil {
		return err
	}

	filter.UserID = user.ID

	result := initializers.DB.Create(filter)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create presaved filter",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Presaved filter created successfully",
		"data":    filter,
	})
}
