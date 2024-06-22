package controllers

import (
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

func SendNot(c *fiber.Ctx) error {
	var reqBody struct {
		Title       string `json:"title"`
		Text        string `json:"text"`
		DeviceToken string `json:"deviceToken"`
		PageURL     string `json:"pageURL"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return err
	}

	// Check if required fields are provided
	if reqBody.Title == "" || reqBody.Text == "" || reqBody.PageURL == "" || reqBody.DeviceToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	privateKeyPath := "keys/AuthKey_485K6P55G9.p8"
	keyID := "485K6P55G9"   // Идентификатор ключа (Key ID) из Apple Developer Console
	teamID := "DBJ8D3U6HY"  // Идентификатор команды (Team ID) из Apple Developer Console
	bundleID := "ddrw.myru" // Bundle ID вашего приложения

	authKey, err := token.AuthKeyFromFile(privateKeyPath)
	if err != nil {
		fmt.Println("err download AuthKey:", err)
	}

	tokenSource := &token.Token{
		KeyID:   keyID,
		TeamID:  teamID,
		AuthKey: authKey,
	}

	client := apns2.NewTokenClient(tokenSource).Development()
	// client := apns2.NewTokenClient(tokenSource).Production()

	// Токен устройства, который вы получили после успешной регистрации на уведомления
	deviceToken := reqBody.DeviceToken

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken

	notification.Topic = bundleID

	payload := payload.NewPayload().AlertTitle(reqBody.Title).AlertBody(reqBody.Text).Badge(1).Custom("urlString", reqBody.PageURL)

	notification.Payload = payload

	res, err := client.Push(notification)

	if err != nil {
		fmt.Println("error send:", err)

	}

	fmt.Println("200:", res)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New push added successfully",
		"data":    res,
	})
}

func CreateDevice(c *fiber.Ctx) error {
	var newDevice models.DevicesIOS
	if err := c.BodyParser(&newDevice); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	if newDevice.Device == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Device field cannot be empty",
		})
	}

	var existingDevice models.DevicesIOS
	if err := initializers.DB.Where("device = ?", newDevice.Device).First(&existingDevice).Error; err == nil {
		// Устройство уже существует в базе данных
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"status":  "error",
			"message": "Device already exists",
		})
	}

	newDevice.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&newDevice).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new Device",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New device added successfully",
		"data":    newDevice,
	})
}
