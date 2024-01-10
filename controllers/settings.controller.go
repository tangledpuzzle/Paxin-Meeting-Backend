package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

// LangsResponse represents the response structure for the Langs endpoint.
type LangsResponse struct {
	Status string         `json:"status"`
	Data   []models.Langs `json:"data"`
}

// AddLangResponse represents the response structure for the AddLang endpoint.
type AddLangResponse struct {
	Status string       `json:"status"`
	Data   models.Langs `json:"data"`
}

// DeleteLangResponse represents the response structure for the DeleteLang endpoint.
type DeleteLangResponse struct {
	Status string       `json:"status"`
	Data   models.Langs `json:"data"`
}

// UpdateLangResponse represents the response structure for the UpdateLang endpoint.
type UpdateLangResponse struct {
	Status string       `json:"status"`
	Data   models.Langs `json:"data"`
}

type LangExample struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Langs returns a list of languages.
// @Summary Get a list of languages.
// @Description Retrieve a list of available languages.
// @Tags Languages
// @Produce json
// @Success 200 {object} LangsResponse "success"
// @Router /settings/langs [get]
func Langs(c *fiber.Ctx) error {

	var langs []models.Langs
	err := initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch languages from the database",
		})
	}

	// Check if no languages were found
	if len(langs) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": "success",
			"data":   "Languages not found, create a new one", // or "langs not found" as you mentioned
		})
	}

	return c.JSON(LangsResponse{
		Status: "success",
		Data:   langs,
	})

}

// AddLang creates a new language.
// @Summary Create a new language.
// @Description Create a new language with the provided data.
// @Tags Languages
// @Produce json
// @Param lang body LangExample true "Language data to be created"
// @Success 200 {object} AddLangResponse "success"
// @Router /settings/addlang [post]
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
	return c.JSON(AddLangResponse{
		Status: "success",
		Data:   newLang,
	})
}

// DeleteLang deletes a language.
// @Summary Delete a language.
// @Description Delete a language with the provided ID.
// @Tags Languages
// @Produce json
// @Param id path int true "Language ID to be deleted"
// @Success 200 {object} DeleteLangResponse "success"
// @Router /settings/deletelang/{id} [delete]
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
	return c.JSON(DeleteLangResponse{
		Status: "success",
		Data:   existingLang,
	})
}

// UpdateLang updates a language.
// @Summary Update a language.
// @Description Update a language with the provided ID and data.
// @Tags Languages
// @Produce json
// @Param id path int true "Language ID to be updated"
// @Param lang body LangExample true "Language data to be updated"
// @Success 200 {object} UpdateLangResponse "success"
// @Router /settings/updatelang/{id} [patch]
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
	return c.JSON(UpdateLangResponse{
		Status: "success",
		Data:   existingLang,
	})
}
