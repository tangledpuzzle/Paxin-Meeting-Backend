package controllers

import (
	"errors"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateRoomRequest struct {
	AcceptorId     string `json:"acceptorId"`
	InitialMessage string `json:"initialMessage"`
}

type RoomResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Members     []RoomMemberResponse `json:"members"`
	Version     uint                 `json:"version"`
	CreatedAt   time.Time            `json:"createdAt"`
	BumpedAt    time.Time            `json:"bumpedAt"`
	LastMessage *models.ChatMessage  `json:"lastMessage,omitempty"`
}

type RoomMemberResponse struct {
	ID           uint        `json:"id"`
	RoomID       uint        `json:"roomId"`
	UserID       string      `json:"userId"`
	User         models.User `json:"user"`
	IsSubscribed bool        `json:"isSubscribed"`
	IsNew        bool        `json:"isNew"`
	JoinedAt     time.Time   `json:"joinedAt"`
}

func CreateChatRoom(c *fiber.Ctx) error {
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

func GetRoomDetails(c *fiber.Ctx) error {
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

func GetSubscribedRooms(c *fiber.Ctx) error {
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

func GetNewUnsubscribedRooms(c *fiber.Ctx) error {
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

func SubscribeNewRoom(c *fiber.Ctx) error {
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

func UnsubscribeRoom(c *fiber.Ctx) error {
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
