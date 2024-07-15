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
	db := initializers.DB.Where("user_id = ?", user.ID).Order("created_at DESC")

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

func MarkNotificationAsRead(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	id, err := strconv.Atoi(notificationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	var notification models.Notification
	if err := initializers.DB.First(&notification, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Notification not found",
		})
	}

	notification.Read = true
	if err := initializers.DB.Save(&notification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update notification status",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   notification,
	})
}

func DeleteNotification(c *fiber.Ctx) error {
	notificationID := c.Params("id")
	id, err := strconv.Atoi(notificationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid notification ID",
		})
	}

	if err := initializers.DB.Delete(&models.Notification{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete notification",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Notification deleted successfully",
	})
}
