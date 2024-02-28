package controllers

import (
	"errors"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateRoomRequest struct {
	AcceptorId     string `json:"acceptorId"`
	InitialMessage string `json:"initialMessage"`
}

type SendMessageRequest struct {
	Content string `json:"content"`
	RoomID  uint   `json:"roomId"`
}

type EditMessageRequest struct {
	Content string `json:"content"`
}

func CreateChatRoomForDM(c *fiber.Ctx) error {
	requestor := c.Locals("user").(models.UserResponse)

	// Parse request body
	payload := new(CreateRoomRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var requestorUser, acceptorUser models.User
	getRequestorUserResult := initializers.DB.First(&requestorUser, "id = ?", requestor.ID)
	if getRequestorUserResult.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Failed to find user with ID"})
	}
	getAcceptorUserResult := initializers.DB.First(&acceptorUser, "id = ?", payload.AcceptorId)
	if getAcceptorUserResult.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Failed to find user with ID"})
	}

	// Check existing room with both users
	var room models.ChatRoom
	result := initializers.DB.
		Model(&models.ChatRoom{}).
		Joins("JOIN chat_room_members as rm1 ON rm1.room_id = chat_rooms.id AND rm1.user_id = ?", requestorUser.ID).
		Joins("JOIN chat_room_members as rm2 ON rm2.room_id = chat_rooms.id AND rm2.user_id = ?", acceptorUser.ID).
		Where("chat_rooms.id IN (SELECT room_id FROM chat_room_members GROUP BY room_id HAVING COUNT(DISTINCT user_id) >= 2)").
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Joins("User")
		}).
		Preload("LastMessage").
		First(&room)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Room does not exist, so proceed with creation
		newRoom := models.ChatRoom{Name: requestorUser.Name + " & " + acceptorUser.Name}
		if err := initializers.DB.Create(&newRoom).Error; err != nil {
			// Handle potential error during room creation
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to create room"})
		}

		roomMembers := []models.ChatRoomMember{
			{RoomID: newRoom.ID, UserID: requestorUser.ID, IsSubscribed: true},
			{RoomID: newRoom.ID, UserID: acceptorUser.ID, IsNew: true},
		}

		if err := initializers.DB.CreateInBatches(roomMembers, len(roomMembers)).Error; err != nil {
			// Handle potential error during room members creation
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to add members to room"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"room":    newRoom,
				"members": roomMembers,
			},
		})

	} else if result.Error != nil {
		// Handle errors other than Not Found error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error checking for existing room"})
	}

	// The room already exists
	return c.JSON(fiber.Map{
		"status":  "fail",
		"message": "Room already exists",
		"room":    room,
	})
}

func GetRoomDetailsForDM(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	var room models.ChatRoom
	if err := initializers.DB.
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Joins("User")
		}).
		Preload("LastMessage").
		Where("id = ?", roomId).
		First(&room).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Room not found", "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   room,
	})
}

func GetSubscribedRoomsForDM(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var rooms []models.ChatRoom
	result := initializers.DB.
		Model(&models.ChatRoom{}).
		Joins("JOIN chat_room_members ON chat_rooms.id = chat_room_members.room_id").
		Where("chat_room_members.user_id = ? AND chat_room_members.is_subscribed = ?", user.ID, true).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Joins("User")
		}).
		Preload("LastMessage").
		Find(&rooms)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch rooms",
			"error":   result.Error.Error(),
		})
	}

	// If no error occurs and rooms were found, return them
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   rooms,
	})
}

func GetNewUnsubscribedRoomsForDM(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var rooms []models.ChatRoom
	result := initializers.DB.
		Model(&models.ChatRoom{}).
		Joins("JOIN chat_room_members ON chat_rooms.id = chat_room_members.room_id").
		Where("chat_room_members.user_id = ? AND chat_room_members.is_subscribed = ? AND chat_room_members.is_new = ?", user.ID, false, true).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Joins("User")
		}).
		Preload("LastMessage").
		Find(&rooms)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch new, unsubscribed chat rooms",
			"error":   result.Error.Error(),
		})
	}

	// If no error occurs and rooms are found, return them
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   rooms,
	})
}

func SubscribeNewRoomForDM(c *fiber.Ctx) error {
	// Extract the authenticated user from the context.
	user := c.Locals("user").(models.UserResponse)

	// Extract roomId from the request's parameters.
	roomIdParam := c.Params("roomId")
	var roomId uint64

	// Convert roomIdParam to uint64.
	if _, err := fmt.Sscanf(roomIdParam, "%d", &roomId); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid room ID parameter",
		})
	}

	// Verify that the room exists.
	var room models.ChatRoom
	if err := initializers.DB.First(&room, roomId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Room not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error",
		})
	}

	// Check if the user is already a member of the room but not subscribed.
	var roomMember models.ChatRoomMember
	err := initializers.DB.Where("room_id = ? AND user_id = ? AND is_subscribed = ?", roomId, user.ID, false).First(&roomMember).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"status":  "error",
				"message": "User is either not a member or already subscribed",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error",
		})
	}

	// User is a member but not subscribed, so proceed to update the subscription and IsNew to false.
	roomMember.IsNew = false
	roomMember.IsSubscribed = true
	if err := initializers.DB.Save(&roomMember).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update subscription",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully subscribed to room",
	})
}

func UnsubscribeRoomForDM(c *fiber.Ctx) error {
	// Extract the authenticated user from the context.
	user := c.Locals("user").(models.UserResponse)

	// Extract roomId from the request's parameters.
	roomIdParam := c.Params("roomId")
	var roomId uint64

	// Convert roomIdParam to uint64.
	if _, err := fmt.Sscanf(roomIdParam, "%d", &roomId); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid room ID parameter",
		})
	}

	// Verify that the room exists.
	var room models.ChatRoom
	if err := initializers.DB.First(&room, roomId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Room not found",
			})
		}
		// Handle other possible errors.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error",
		})
	}

	// Check if the user is a subscribed member of the room.
	var roomMember models.ChatRoomMember
	err := initializers.DB.Where("room_id = ? AND user_id = ? AND is_subscribed = ?", roomId, user.ID, true).First(&roomMember).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"status":  "fail",
				"message": "User is not subscribed to this room",
			})
		}
		// Handle other possible errors.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Database error",
		})
	}

	// User is subscribed, proceed to update the subscription to false.
	roomMember.IsSubscribed = false
	if err := initializers.DB.Save(&roomMember).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update subscription status",
			"error":   err.Error(),
		})
	}

	// Successfully updated the subscription to unsubscribed.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully unsubscribed from room",
	})
}

func SendMessageForDM(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	payload := new(SendMessageRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	message := models.ChatMessage{
		Content: payload.Content,
		UserID:  user.ID,
		RoomID:  payload.RoomID,
	}

	if err := initializers.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to send message"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Message sent"})
}

func EditMessageForDM(c *fiber.Ctx) error {
	userID := c.Locals("user").(models.UserResponse).ID
	messageID := c.Params("messageId")

	payload := new(EditMessageRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	var message models.ChatMessage
	result := initializers.DB.First(&message, "id = ? AND user_id = ?", messageID, userID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Message not found",
		})
	}

	message.Content = payload.Content
	message.IsEdited = true
	initializers.DB.Save(&message)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message updated successfully",
	})
}

func DeleteMessageForDM(c *fiber.Ctx) error {
	userID := c.Locals("user").(models.UserResponse).ID
	messageID := c.Params("messageId")

	// Assuming messages are associated with a UserID for authorization
	var message models.ChatMessage
	result := initializers.DB.First(&message, "id = ? AND user_id = ?", messageID, userID)

	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Message not found",
		})
	}

	if err := initializers.DB.Delete(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete message",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message deleted successfully",
	})
}

func GetChatMessagesForDM(c *fiber.Ctx) error {
	roomID := c.Params("roomId")

	var messages []models.ChatMessage
	// Fetch all messages, including those marked as deleted
	result := initializers.DB.Unscoped().Where("room_id = ?", roomID).Find(&messages)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch messages",
		})
	}

	// Iterate through messages to hide content of deleted messages
	for i, msg := range messages {
		if msg.DeletedAt != nil {
			messages[i].Content = "This message has been deleted."
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   messages,
	})
}
