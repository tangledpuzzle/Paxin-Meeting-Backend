package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func Langs(c *fiber.Ctx) error {

	var langs []models.Langs
	err := initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   langs,
	})

}

func AddLang(c *fiber.Ctx) error {
	// Parse request data into a Langs struct
	var newLang models.Langs
	if err := c.BodyParser(&newLang); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	// Add the new language to the database
	if err := initializers.DB.Create(&newLang).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new language",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New language added successfully",
		"data":    newLang,
	})
}

func DeleteLang(c *fiber.Ctx) error {
	// Get language ID from the request parameters
	langID := c.Params("id")

	// Check if the language ID is valid
	var existingLang models.Langs
	if err := initializers.DB.First(&existingLang, langID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Language not found",
		})
	}

	// Delete the language from the database
	if err := initializers.DB.Delete(&existingLang).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete the language",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Language deleted successfully",
		"data":    existingLang,
	})
}

func UpdateLang(c *fiber.Ctx) error {
	// Get language ID from the request parameters
	langID := c.Params("id")

	// Check if the language ID is valid
	var existingLang models.Langs
	if err := initializers.DB.First(&existingLang, langID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Language not found",
		})
	}

	// Parse request data into a Langs struct
	var updatedLang models.Langs
	if err := c.BodyParser(&updatedLang); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	// Update the existing language with the new data
	existingLang.Name = updatedLang.Name

	// Save the changes to the database
	if err := initializers.DB.Save(&existingLang).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update the language",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Language updated successfully",
		"data":    existingLang,
	})
}
