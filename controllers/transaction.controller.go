package controllers

import (
	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
)

func GetTransactions(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	// get the balance of the authenticated user
	var transaction []models.Transaction

	
	err := utils.Paginate(c, initializers.DB.Where("user_id = ?", user.ID).Order("created_at DESC").First(&transaction), &transaction)
	if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Element not found",
			})
		}

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   transaction,
    })
}
