package controllers

import (
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func GetDomain(c *fiber.Ctx) error {
	// Get the domain name from the query parameter
	domainName := c.Query("domain")

	var domain models.Domain
	if err := initializers.DB.Where("name = ?", domainName).First(&domain).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch domain from database",
		})
	}

	settingsMap := make(map[string]interface{})
	if err := domain.Settings.AssignTo(&settingsMap); err != nil {
		// Handle error if needed
		_ = err
	}

	fmt.Println(settingsMap)

	domainResponse := models.DomainResponse{
		ID:       domain.ID,
		UserID:   domain.UserID.String(),
		Username: domain.Username,
		Name:     domain.Name,
		Settings: settingsMap,
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   domainResponse,
	})
}

func UpdateSite(c *fiber.Ctx) error {

	var newSettings map[string]interface{}
	if err := c.BodyParser(&newSettings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	user := c.Locals("user").(models.UserResponse)
	result := initializers.DB.Model(&models.Domain{}).
		Where("user_id = ?", user.ID).
		Update("settings", newSettings)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update domain settings",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Domain settings updated successfully",
	})
}

func GetSite(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var domain models.Domain
	err := initializers.DB.Where("user_id = ?", user.ID).First(&domain).Error
	if err != nil {
		return c.JSON(fiber.Map{
			"status": "success",
			"data":   "domain not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   domain,
	})
}
