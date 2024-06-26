package controllers

import (
	"errors"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type CreateRoomRequest struct {
	AcceptorId     string `json:"acceptorId"`
	InitialMessage string `json:"initialMessage"`
}

type SendMessageRequest struct {
	Content         string `json:"content"`
	ParentMessageID string `json:"parentMessageId,omitempty"` // Use omitempty for an optional field
	MsgType         string `json:"msgType,omitempty"`
	JsonData        string `json:"jsonData,omitempty"` // this is msg field for system, backend only validates this as json
}

type EditMessageRequest struct {
	Content string `json:"content"`
}

type UserLatestMsgRequest struct {
	MessageId string `json:"messageId"`
}

type ChatRoomResponse struct {
	models.ChatRoom
	UnreadMessages string `json:"unreadMessages"` // using a numeric string for the count is more appropriate.
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to create room"})
		}

		roomMembers := []models.ChatRoomMember{
			{RoomID: newRoom.ID, UserID: requestorUser.ID, IsSubscribed: true},
			{RoomID: newRoom.ID, UserID: acceptorUser.ID, IsNew: true},
		}

		if err := initializers.DB.CreateInBatches(roomMembers, len(roomMembers)).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to add members to room"})
		}
		// Create initial message
		initialMessage := models.ChatMessage{
			RoomID:  newRoom.ID,
			UserID:  requestorUser.ID,
			Content: payload.InitialMessage,
		}
		if err := initializers.DB.Create(&initialMessage).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to send initial message"})
		}

		// Update the room's LastMessageId with the ID of the initial message
		if err := initializers.DB.Model(&newRoom).Update("last_message_id", initialMessage.ID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to update room's last message"})
		}

		serializedRoom := utils.SerializeChatRoom(newRoom.ID)
		channels, err := GetRoomMemberChannels(newRoom.ID)
		if err != nil {
			log.Printf("Failed to get room member channels: %s", err)
		} else {
			broadcastPayload := CentrifugoBroadcastPayload{
				Channels: channels,
				Data: struct {
					Type string                 `json:"type"`
					Body map[string]interface{} `json:"body"`
				}{
					Type: "new_room",
					Body: serializedRoom,
				},
				IdempotencyKey: fmt.Sprintf("create_room_%d", newRoom.ID),
			}

			if _, err := CentrifugoBroadcastRoom(fmt.Sprint(newRoom.ID), broadcastPayload); err != nil {
				log.Printf("Failed to broadcast room creation: %s", err)
			}
		}

		roomIDStr := strconv.FormatUint(newRoom.ID, 10)
		pageURL := fmt.Sprintf("https://www.myru.online/ru/chat/%s", roomIDStr)

		sendPushNotificationToOwner(acceptorUser.ID, requestorUser.Name, initialMessage.Content, pageURL)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"room": newRoom,
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
	user := c.Locals("user").(models.UserResponse)

	var room models.ChatRoom
	if err := initializers.DB.
		Model(&models.ChatRoom{}).
		Joins("JOIN chat_room_members ON chat_rooms.id = chat_room_members.room_id").
		Where("chat_rooms.id = ? AND chat_room_members.user_id = ?", roomId, user.ID).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Joins("User").
				Preload("User.Profile", func(db *gorm.DB) *gorm.DB {
					return db.
						Preload("City", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Guilds", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Hashtags").
						Preload("Photos").
						Preload("Documents").
						Preload("Service")
				})
		}).
		Preload("LastMessage").
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Room not found or access denied", "error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to retrieve room details", "error": err.Error()})
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
			return db.Joins("User").
				Preload("User.Profile", func(db *gorm.DB) *gorm.DB {
					return db.
						Preload("City", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Guilds", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Hashtags").
						Preload("Photos").
						Preload("Documents").
						Preload("Service")
				})
		}).
		Preload("LastMessage").
		Find(&rooms)

	var responseRooms []ChatRoomResponse
	// Now, for each room, calculate the unread message count
	for _, room := range rooms {
		var count int64
		err := initializers.DB.
			Model(&models.ChatMessage{}).
			Where(`
            user_id != ? AND room_id = ? AND 
            id > COALESCE(
                (
                    SELECT last_read_message_id
                    FROM chat_room_members 
                    WHERE room_id = ? AND user_id = ?  AND is_subscribed = ?
                ), 
                0
            )
        `, user.ID, room.ID, room.ID, user.ID, true).
			Count(&count).Error

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Could not calculate unread Msg Count, room id is %d", room.ID),
				"error":   err.Error(),
			})
		}

		// Append the enriched room data to the response slice
		responseRooms = append(responseRooms, ChatRoomResponse{
			ChatRoom:       room,
			UnreadMessages: strconv.FormatInt(count, 10),
		})
	}

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch rooms",
			"error":   result.Error.Error(),
		})
	}

	// If no error occurs and rooms were found, return them
	// return c.JSON(fiber.Map{
	// 	"status": "success",
	// 	"data":   rooms,
	// })

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   responseRooms,
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
			return db.Joins("JOIN users ON chat_room_members.user_id = users.id").
				Preload("User.Profile", func(db *gorm.DB) *gorm.DB {
					return db.
						Preload("City", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Guilds", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Hashtags").
						Preload("Photos").
						Preload("Documents").
						Preload("Service")
				})
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

func GetUnsubscribedNotNewRoomsForDM(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	var rooms []models.ChatRoom
	result := initializers.DB.
		Model(&models.ChatRoom{}).
		Joins("JOIN chat_room_members ON chat_rooms.id = chat_room_members.room_id").
		Where("chat_room_members.user_id = ? AND chat_room_members.is_subscribed = ? AND chat_room_members.is_new = ?", user.ID, false, false).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN users ON chat_room_members.user_id = users.id").
				Preload("User.Profile", func(db *gorm.DB) *gorm.DB {
					return db.
						Preload("City", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Guilds", func(db *gorm.DB) *gorm.DB {
							return db.Preload("Translations")
						}).
						Preload("Hashtags").
						Preload("Photos").
						Preload("Documents").
						Preload("Service")
				})
		}).
		Preload("LastMessage").
		Order("created_at DESC"). // You may wish to order the rooms
		Find(&rooms)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not fetch unsubscribed, not new chat rooms",
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

	serializedRoom := utils.SerializeChatRoom(room.ID)
	channels, err := GetRoomMemberChannels(room.ID)
	if err != nil {
		log.Printf("Failed to get room member channels for broadcasting: %s", err)
	} else {
		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "subscribe_room",
				Body: serializedRoom,
			},
			IdempotencyKey: fmt.Sprintf("subscribe_room_%d", room.ID),
		}

		if _, err := CentrifugoBroadcastRoom(fmt.Sprint(room.ID), broadcastPayload); err != nil {
			log.Printf("Failed to broadcast room subscription: %s", err)
		}
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
	err := initializers.DB.Where("room_id = ? AND user_id = ?", roomId, user.ID).First(&roomMember).Error
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
	roomMember.IsNew = false
	if err := initializers.DB.Save(&roomMember).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update subscription status",
			"error":   err.Error(),
		})
	}

	serializedRoom := utils.SerializeChatRoom(room.ID)
	channels, err := GetRoomMemberChannels(room.ID)
	if err != nil {
		log.Printf("Failed to get room member channels for broadcasting: %s", err)
	} else {
		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "unsubscribe_room",
				Body: serializedRoom,
			},
			IdempotencyKey: fmt.Sprintf("unsubscribe_room_%d", room.ID),
		}

		if _, err := CentrifugoBroadcastRoom(fmt.Sprint(room.ID), broadcastPayload); err != nil {
			log.Printf("Failed to broadcast room unsubscription: %s", err)
		}
	}

	// Successfully updated the subscription to unsubscribed.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Successfully unsubscribed from room",
	})
}

func SendMessageForDM(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)
	roomId := c.Params("roomId")
	u64, err := strconv.ParseUint(roomId, 10, 64)
	if err != nil {
		fmt.Println("Conversion error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to parse roomId"})
	}

	payload := new(SendMessageRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Check if the user is a subscribed member of the room
	var member models.ChatRoomMember
	result := initializers.DB.Model(&models.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ?", u64, user.ID).
		First(&member)
	if result.Error != nil || !member.IsSubscribed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "User is not subscribed to the room"})
	}

	// Check if the user is a member of the room and if the other member is subscribed.
	var count int64
	initializers.DB.Model(&models.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ?", u64, user.ID).
		Count(&count)
	// Check if the user is indeed a member of the room
	if count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "User is not a member of the room"})
	}

	// check if there's another subscribed member in this room
	initializers.DB.Model(&models.ChatRoomMember{}).
		Where("room_id = ? AND user_id != ? AND is_subscribed = ?", u64, user.ID, true).
		Count(&count)
	if count == 0 {
		// This means the other member is not subscribed or does not exist
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "The other member is not subscribed or does not exist"})
	}

	// Initialize the ChatMessage with common fields
	message := models.ChatMessage{
		Content: payload.Content,
		UserID:  user.ID,
		RoomID:  u64,
		MsgType: 0,
	}

	// Check if ParentMessageID is present and valid
	if payload.ParentMessageID != "" {
		parentMessageId, err := strconv.ParseUint(payload.ParentMessageID, 10, 64)
		if err != nil {
			fmt.Println("parentMessageId Conversion error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to parse parentMessageId",
			})
		}
		message.ParentMessageID = &parentMessageId
	} else {
		// When ParentMessageID is not present, it's implicitly understood that message.ParentMessageID is nil
		fmt.Println("Creating new message with content")

	}

	// Check whether there is a msgType
	if payload.MsgType != "" {
		msgType, err := strconv.ParseUint(payload.MsgType, 10, 8)
		if err != nil {
			fmt.Println("msgType Conversion error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to parse msgType",
			})
		}
		message.MsgType = uint8(msgType)

		// additionally store json data
		if payload.JsonData != "" {
			message.JsonData = &payload.JsonData
		}
	} else {
		// When msgType is not present, it's implicitly understood that message.ParentMessageID is nil
		fmt.Println("Creating new message with content, default msgType is 0...")
	}

	if err := initializers.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to send message"})
	}

	// Update the room's LastMessageId after sending a new message
	if err := initializers.DB.Model(&models.ChatRoom{}).Where("id = ?", message.RoomID).Update("last_message_id", message.ID).Error; err != nil {
		fmt.Println("Failed to update room's last message: ", err)
	}

	serializedMessage := utils.SerializeChatMessage(message)
	channels, err := GetRoomMemberChannels(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room member channels for broadcasting: %s", err)
	} else {
		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "new_message",
				Body: serializedMessage,
			},
			IdempotencyKey: fmt.Sprintf("send_message_%d", message.ID),
		}

		if _, err := CentrifugoBroadcastRoom(fmt.Sprint(message.RoomID), broadcastPayload); err != nil {
			log.Printf("Failed to broadcast new message: %s", err)
		}
	}

	roomIDStr := strconv.FormatUint(message.RoomID, 10)
	pageURL := fmt.Sprintf("https://www.myru.online/ru/chat/%s", roomIDStr)

	sendPushNotificationToOwner(member.UserID, user.Name, message.Content, pageURL)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"message": message}})
}

func EditMessageForDM(c *fiber.Ctx) error {
	userID := c.Locals("user").(models.UserResponse).ID
	messageIDParam := c.Params("messageId")
	messageID, err := strconv.ParseUint(messageIDParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid messageId format",
		})
	}

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
			"message": "Message not found or not owned by user",
		})
	}

	message.Content = payload.Content
	message.IsEdited = true
	if err := initializers.DB.Save(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update message",
			"error":   err.Error(),
		})
	}

	serializedMessage := utils.SerializeChatMessage(message)
	channels, err := GetRoomMemberChannels(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room member channels for broadcasting: %s", err)
	} else {
		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "edit_message",
				Body: serializedMessage,
			},
			IdempotencyKey: fmt.Sprintf("edit_message_%d", message.ID),
		}

		if _, err := CentrifugoBroadcastRoom(fmt.Sprint(message.RoomID), broadcastPayload); err != nil {
			log.Printf("Failed to broadcast message update: %s", err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message updated successfully",
		"data": fiber.Map{
			"message": message,
		},
	})
}

func DeleteMessageForDM(c *fiber.Ctx) error {
	userID := c.Locals("user").(models.UserResponse).ID
	messageIDParam := c.Params("messageId")
	messageID, err := strconv.ParseUint(messageIDParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid messageId format",
		})
	}

	var message models.ChatMessage
	result := initializers.DB.First(&message, "id = ? AND user_id = ?", messageID, userID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Message not found or not owned by user",
		})
	}

	// Perform soft delete by updating IsDeleted to true and setting DeletedAt to the current time
	now := time.Now()
	updateResult := initializers.DB.Model(&message).Updates(models.ChatMessage{IsDeleted: true, DeletedAt: &now})
	if updateResult.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to flag message as deleted",
			"error":   updateResult.Error.Error(),
		})
	}

	tempMessage := message
	tempMessage.Content = "This message has been deleted."
	serializedMessage := utils.SerializeChatMessage(tempMessage)
	channels, err := GetRoomMemberChannels(message.RoomID)
	if err != nil {
		log.Printf("Failed to get room member channels for broadcasting: %s", err)
	} else {
		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "delete_message",
				Body: serializedMessage,
			},
			IdempotencyKey: fmt.Sprintf("delete_message_%d", message.ID),
		}

		if _, err := CentrifugoBroadcastRoom(fmt.Sprint(message.RoomID), broadcastPayload); err != nil {
			log.Printf("Failed to broadcast message deletion notice: %s", err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message flagged as deleted successfully",
	})
}

func GetChatMessagesForDM(c *fiber.Ctx) error {
	// Extract the user ID from context and room ID from URL params.
	userID := c.Locals("user").(models.UserResponse).ID
	roomID := c.Params("roomId")
	roomIDParsed, err := strconv.ParseUint(roomID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid room ID format, must be a positive number",
		})
	}

	// Check if the user is a member of the room.
	var member models.ChatRoomMember
	err = initializers.DB.Where("user_id = ? AND room_id = ?", userID, roomIDParsed).First(&member).Error
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User is not a member of the room or room does not exist",
		})
	}

	// Retrieve pagination parameters with defaults if not provided.

	skip := c.QueryInt("skip", 0)    // Starting from the most recent message.
	limit := c.QueryInt("limit", 10) // Number of messages to fetch.

	// Check if end_msg_id is provided in the query string and handle it.
	endMsgIDParam := c.Query("end_msg_id")
	var endMsgID uint64
	endMsgIDProvided := false

	if endMsgIDParam != "" {
		var parseErr error
		endMsgID, parseErr = strconv.ParseUint(endMsgIDParam, 10, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid end_msg_id format, must be a positive number",
			})
		}
		endMsgIDProvided = true
	}

	// Fetch the messages, updated to respect skip and limit with correct ordering.
	var messages []models.ChatMessage

	// Prepared base query with dynamic conditions
	query := initializers.DB.Unscoped().Model(&models.ChatMessage{}).
		Where("room_id = ?", roomIDParsed).
		Order("created_at DESC")

	// Adjust query based on end_msg_id presence
	if endMsgIDProvided {
		query = query.Offset(skip).Where("id >= ?", endMsgID) // Assuming you want messages before and including endMsgID
	} else {
		// Apply pagination if end_msg_id is not provided
		query = query.Offset(skip).Limit(limit)
	}

	// Execute the query with complex preloading
	err = query.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Profile", func(db *gorm.DB) *gorm.DB {
					return db.
						Preload("City.Translations").
						Preload("Guilds.Translations").
						Preload("Hashtags").
						Preload("Photos").
						Preload("Documents").
						Preload("Service")
				})
		}).
		Preload("ParentMessage").
		Find(&messages).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch messages",
			"error":   err.Error(),
		})
	}

	// Total number of messages in the room for pagination info.
	var totalCount int64
	initializers.DB.Model(&models.ChatMessage{}).
		Where("room_id = ?", roomIDParsed).
		Count(&totalCount)

	// Iterate through messages to hide content of deleted messages.
	for i, msg := range messages {
		if msg.IsDeleted {
			messages[i].Content = "This message has been deleted."
		}
	}

	// Return the paginated chat messages.
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"messages": messages,
		},
		"skip":       skip,
		"limit":      limit,
		"end_msg_id": endMsgIDParam,
		"totalCount": totalCount,
	})
}

func MarkMessageAsReadForDM(c *fiber.Ctx) error {
	userID := c.Locals("user").(models.UserResponse).ID
	roomID := c.Params("roomId")
	roomIDParsed, err := strconv.ParseUint(roomID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid room ID format, must be a positive number",
		})
	}

	// Check if the user is a member of the room
	var member models.ChatRoomMember
	err = initializers.DB.Where("user_id = ? AND room_id = ?", userID, roomIDParsed).First(&member).Error
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User is not a member of the room or room does not exist",
		})
	}

	// Parse request body
	payload := new(UserLatestMsgRequest)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	messageIDParsed, err := strconv.ParseUint(payload.MessageId, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid message ID format, must be a positive number",
		})
	}

	var message models.ChatMessage
	result := initializers.DB.First(&message, "id = ?", messageIDParsed)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Message not found or not owned by user",
		})
	}

	if message.RoomID != roomIDParsed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Message not found in this room",
		})
	}

	member.LastReadMessageID = maxUint64Ptr(member.LastReadMessageID, messageIDParsed)

	if err := initializers.DB.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update latest read message",
			"error":   err.Error(),
		})
	}

	channels, err := GetRoomMemberChannels(roomIDParsed)
	if err != nil {
		log.Printf("Failed to get room member channels: %s", err)
	} else {

		// Prepare the 'Body' map
		bodyMap := map[string]interface{}{}

		bodyMap["lastReadMessageId"] = uint64PtrToString(member.LastReadMessageID)
		bodyMap["ownerId"] = message.UserID.String()
		bodyMap["readerId"] = userID.String()
		bodyMap["roomId"] = strconv.FormatUint(message.RoomID, 10)

		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "updated_last_read_msg_id",
				Body: bodyMap,
			},
			IdempotencyKey: fmt.Sprintf("updated_last_read_msg_%s_%s", bodyMap["readerId"], bodyMap["lastReadMessageId"]),
		}

		if _, err := CentrifugoBroadcastRoom(fmt.Sprint(roomIDParsed), broadcastPayload); err != nil {
			log.Printf("Failed to broadcast update latest read msg ID: %s", err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message is marked as read",
		"data": fiber.Map{
			"updated_latest_read": member.LastReadMessageID,
		},
	})
}

func MarkMessageAsUnReadForDM(c *fiber.Ctx) error {
	userID := c.Locals("user").(models.UserResponse).ID
	roomID := c.Params("roomId")
	status := c.Params("status")
	roomIDParsed, err := strconv.ParseUint(roomID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid room ID format, must be a positive number",
		})
	}
	// Parsing a bool value
	statusParsed, err := strconv.ParseBool(status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid status format, must be a bool",
		})
	}

	// Check if the user is a member of the room
	var member models.ChatRoomMember
	err = initializers.DB.Where("user_id = ? AND room_id = ?", userID, roomIDParsed).First(&member).Error
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User is not a member of the room or room does not exist",
		})
	}

	member.IsUnread = statusParsed

	if err := initializers.DB.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update unread status",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Room is marked as unread",
		"data": fiber.Map{
			"marked_as_unread": member.RoomID,
		},
	})
}

func SendUserTypingToCentrifugo(userID uuid.UUID, roomID string) error {
	roomIDParsed, err := strconv.ParseUint(roomID, 10, 64)
	if err != nil {
		return errors.New("invalid room ID format, must be a positive number")
	}

	// Check if the user is a member of the room
	var member models.ChatRoomMember
	err = initializers.DB.Where("user_id = ? AND room_id = ?", userID, roomIDParsed).First(&member).Error
	if err != nil {
		return fmt.Errorf("user is not a member of the room or room does not exist: %s", err)
	}

	channels, err := GetRoomMemberChannels(roomIDParsed)
	if err != nil {
		return fmt.Errorf("failed to get room member channels for broadcasting: %s", err)
	} else {
		bodyMap := map[string]interface{}{}
		bodyMap["userID"] = userID.String()
		bodyMap["roomID"] = roomIDParsed
		broadcastPayload := CentrifugoBroadcastPayload{
			Channels: channels,
			Data: struct {
				Type string                 `json:"type"`
				Body map[string]interface{} `json:"body"`
			}{
				Type: "user_is_typing",
				Body: bodyMap,
			},
			IdempotencyKey: fmt.Sprintf("user_is_typing_%s_%d_%d", userID.String(), roomIDParsed, time.Now().UTC().UnixMilli()),
		}

		if _, err := CentrifugoBroadcastRoom(roomID, broadcastPayload); err != nil {
			return fmt.Errorf("failed to broadcast message deletion notice: %s", err)
		}
	}
	return nil
}

func maxUint64Ptr(a *uint64, b uint64) *uint64 {
	if a == nil {
		return &b // If a is nil, return b
	}
	if *a < b {
		return &b // If b is greater, return b
	}
	return a // Otherwise, return a as it's already the maximum
}

func uint64PtrToString(val *uint64) string {
	if val == nil {
		// If the input is nil, you might want to return a default value
		// or indicate somehow that conversion wasn't possible.
		return ""
	}
	// Dereference the pointer, convert the uint64 value to a string.
	return strconv.FormatUint(*val, 10)
}
