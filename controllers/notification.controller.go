package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetNotifications(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var notifications []models.Notification
	db := initializers.DB.Where("user_id = ?", user.ID)

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid limit parameter",
		})
	}

	skip, err := strconv.ParseInt(c.Query("skip", "0"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid skip parameter",
		})
	}

	var totalCount int64
	if err := db.Model(&models.Notification{}).Count(&totalCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve total count",
		})
	}

	db = db.Limit(limit).Offset(int(skip))

	if err := utils.Paginate(c, db, &notifications); err != nil {
		return err
	}

	var unreadCount int64
	if err := initializers.DB.Model(&models.Notification{}).Where("user_id = ? AND read = ?", user.ID, false).Count(&unreadCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve unread count",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   notifications,
		"unread": unreadCount,
		"meta": fiber.Map{
			"limit": limit,
			"skip":  skip,
			"total": totalCount,
		},
	})
}
