package controllers

import (
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
)

func SendPushNotification(c *fiber.Ctx) error {
	// Получение user из Locals
	IDUSER := c.Locals("user")
	if IDUSER == nil {
		// Обработка случая, когда user равен nil
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	userResp, ok := IDUSER.(models.UserResponse)
	if !ok {
		// Обработка случая, когда user не является типом models.UserResponse
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user type",
		})
	}

	userObj := models.User{
		ID: userResp.ID,
	}

	followers, err := utils.GetFollowers(userObj.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get followers",
		})
	}

	var reqBody struct {
		Title   string `json:"title"`
		Text    string `json:"text"`
		PageURL string `json:"pageURL"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	for _, follower := range followers {
		err := utils.Push(
			reqBody.Title,
			reqBody.Text,
			follower.DeviceIOS,
			reqBody.PageURL,
		)

		if err != nil {
			fmt.Println("Failed to send push notification: ", err)
		} else {
			err := utils.Notification(reqBody.Title, reqBody.Text, follower.ID.String(), reqBody.PageURL)
			if err != nil {
				fmt.Println("Failed to create notification:", err)
			} else {
				fmt.Println("Notification created successfully.")
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "Push notifications sent successfully",
	})
}

// func sendPushNotificationToFollowers(userID uuid.UUID, title, text, pageURL string) error {
// 	followers, err := utils.GetFollowers(userID)
// 	if err != nil {
// 		fmt.Println("Failed to send push notification: ", err)
// 	}

// 	for _, follower := range followers {

// 		if follower.DeviceIOS == "" {
// 			continue
// 		}

// 		err := utils.Push(
// 			title,
// 			text,
// 			follower.DeviceIOS,
// 			pageURL,
// 		)

// 		if err != nil {
// 			fmt.Println("Failed to send push notification: ", err)
// 		}
// 	}

// 	return nil
// }

func sendPushNotificationToOwner(userID uuid.UUID, title, text, pageURL string) error {
	var user models.User
	if err := initializers.DB.
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		fmt.Println("Failed to fetch user from the database: ", err)
		return err
	}

	if user.DeviceIOS == "" {
		fmt.Println("No device_ios for user: ", userID)
		return nil
	}

	err := utils.Push(
		title,
		text,
		user.DeviceIOS,
		pageURL,
	)

	if err != nil {
		fmt.Println("Failed to send push notification: ", err)
	}

	return nil
}

func sendNotificationToOwner(userID, title, text, pageURL string) error {

	var user models.User
	if err := initializers.DB.
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		fmt.Println("Failed to fetch user from the database: ", err)
		return err
	}

	if user.Session != "" {
		utils.SendPersonalMessageToClient(user.Session, "new_notification")
	} else {
		parsedUserID, err := uuid.FromString(userID)
		if err != nil {
			fmt.Println("Invalid UUID format: ", err)
			return err
		}
		sendPushNotificationToOwner(parsedUserID, title, text, pageURL)
	}

	err := utils.Notification(
		title,
		text,
		userID,
		pageURL,
	)

	if err != nil {
		fmt.Println("Failed to send notification: ", err)
	}

	return nil
}
