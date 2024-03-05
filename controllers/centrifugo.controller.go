package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func GetCentrifugoConnectionToken(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)

	tokenClaims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Minute * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedToken, err := token.SignedString([]byte(config.CentrifugoTokenSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to sign token"})
	}

	return c.JSON(fiber.Map{"token": signedToken})
}

func GetCentrifugoSubscriptionToken(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)
	channel := c.Query("channel")

	if channel != "personal:"+user.ID.String() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"detail": "permission denied"})
	}

	tokenClaims := jwt.MapClaims{
		"sub":     user.ID,
		"exp":     time.Now().Add(time.Minute * 5).Unix(),
		"channel": channel,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedToken, err := token.SignedString([]byte(config.CentrifugoTokenSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to sign token"})
	}

	return c.JSON(fiber.Map{"token": signedToken})
}

func GetRoomMemberChannels(roomID uint) ([]string, error) {
	var members []models.ChatRoomMember
	if err := initializers.DB.Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		return nil, err
	}

	var channels []string
	for _, member := range members {
		channels = append(channels, fmt.Sprintf("personal:%s", member.UserID))
	}

	return channels, nil
}

type CentrifugoBroadcastPayload struct {
	Channels []string `json:"channels"`
	Data     struct {
		Type string                 `json:"type"`
		Body map[string]interface{} `json:"body"`
	} `json:"data"`
	IdempotencyKey string `json:"idempotency_key"`
}

func CentrifugoBroadcastViaAPI(apiEndpoint string, apiKey string, payload CentrifugoBroadcastPayload) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %s", err)
		return "", err // Return empty string on error
	}

	apiURL := fmt.Sprintf("%s/api", apiEndpoint)
	request, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request: %s", err)
		return "", err // Return empty string on error
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("apikey %s", apiKey))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request to Centrifugo: %s", err)
		return "", err // Return empty string on error
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("received non-OK response from Centrifugo: %d", response.StatusCode)
		log.Printf(errorMsg)
		return "", fmt.Errorf(errorMsg) // Return empty string on error
	}

	// Return a success message when no errors occur.
	return "Broadcast sent successfully to Centrifugo", nil
}

func CentrifugoBroadcastRoom(roomID string, broadcastPayload CentrifugoBroadcastPayload) (string, error) {
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)

	switch config.CentrifugoBroadcastMode {
	case "api":
		return CentrifugoBroadcastViaAPI(config.CentrifugoHttpApiEndpoint, config.CentrifugoHttpApiKey, broadcastPayload)
	default:
		log.Printf("Broadcast mode '%s' is not implemented", config.CentrifugoBroadcastMode)
		return "", fmt.Errorf("broadcast mode '%s' is not implemented", config.CentrifugoBroadcastMode)
	}
}
