package utils

import (
	"errors"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

func GetFollowers(userID uuid.UUID) ([]models.User, error) {
	var followers []models.User

	subQuery := initializers.DB.Table("user_relation").Select("user_id").Where("following_id = ?", userID)

	if err := initializers.DB.Table("users").Where("id IN (?)", subQuery).Find(&followers).Error; err != nil {
		return nil, err
	}

	return followers, nil
}

func Notification(title, message, userIDStr, URL string) error {
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return errors.New("invalid UUID format for userID")
	}

	notification := &models.Notification{
		Title:     title,
		Message:   message,
		UserID:    userID,
		CreatedAt: time.Now(),
		URL:       URL,
		Read:      false,
	}

	if err := initializers.DB.Create(notification).Error; err != nil {
		return err
	}

	return nil
}

func Push(title, text, deviceToken, pageURL string) error {
	privateKeyPath := "keys/AuthKey_485K6P55G9.p8"
	keyID := "485K6P55G9"   // Идентификатор ключа (Key ID) из Apple Developer Console
	teamID := "DBJ8D3U6HY"  // Идентификатор команды (Team ID) из Apple Developer Console
	bundleID := "ddrw.myru" // Bundle ID вашего приложения

	authKey, err := token.AuthKeyFromFile(privateKeyPath)
	if err != nil {
		fmt.Println("err download AuthKey:", err)
		return err
	}

	tokenSource := &token.Token{
		KeyID:   keyID,
		TeamID:  teamID,
		AuthKey: authKey,
	}

	client := apns2.NewTokenClient(tokenSource).Production()

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       bundleID,
	}

	notification.Payload = payload.NewPayload().
		AlertTitle(title).
		AlertBody(text).
		Badge(1).
		Custom("urlString", pageURL).
		Sound("default")

	res, err := client.Push(notification)
	if err != nil {
		fmt.Println("error send:", err)
		return err
	}

	if res.StatusCode != 200 {
		fmt.Printf("Failed to send push notification: %v\n", res)
		return fmt.Errorf("failed to send push notification: %v", res.Reason)
	}

	fmt.Println("200:", res)

	return nil
}
