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

func GetPresavedfilters(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var presavedfilters []models.Presavedfilters
	if err := initializers.DB.Where("user_id = ?", user.ID).Find(&presavedfilters).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve presaved filters",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   presavedfilters,
	})
}

func DeletePresavedFilter(c *fiber.Ctx) error {
	// Extract user information from the context
	user := c.Locals("user").(models.UserResponse)

	// Get the ID of the presaved filter from the URL parameters
	filterID := c.Params("id")

	// Check if the user owns the presaved filter
	var filter models.Presavedfilters
	if err := initializers.DB.Where("id = ? AND user_id = ?", filterID, user.ID).First(&filter).Error; err != nil {
		// If the presaved filter is not found or the user doesn't own it, return an error
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Presaved filter not found",
		})
	}

	// Delete the presaved filter from the database
	if err := initializers.DB.Delete(&filter).Error; err != nil {
		// If there is an error while deleting the presaved filter, return an error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete presaved filter",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Presaved filter deleted successfully",
	})
}

func PatchPresavedFilter(c *fiber.Ctx) error {
	// Extract user information from the context
	user := c.Locals("user").(models.UserResponse)

	// Get the ID of the presaved filter from the URL parameters
	filterID := c.Params("id")

	// Check if the user owns the presaved filter
	var filter models.Presavedfilters
	if err := initializers.DB.Where("id = ? AND user_id = ?", filterID, user.ID).First(&filter).Error; err != nil {
		// If the presaved filter is not found or the user doesn't own it, return an error
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Presaved filter not found",
		})
	}

	// Parse request body to get updated filter data
	var reqBody models.Presavedfilters
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to parse request body",
		})
	}

	// Update the presaved filter fields
	filter.Name = reqBody.Name

	// Save the updated filter to the database
	if err := initializers.DB.Save(&filter).Error; err != nil {
		// If there is an error while saving the updated filter, return an error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update presaved filter",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Presaved filter updated successfully",
		"data":    filter,
	})
}
