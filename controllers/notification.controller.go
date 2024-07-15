package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"

	"github.com/gofiber/fiber/v2"
)

func GetNotifications(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var notifications []models.Notification
	db := initializers.DB.Where("user_id = ?", user.ID)

	if err := utils.Paginate(c, db, &notifications); err != nil {
		return err
	}

	return nil
}
