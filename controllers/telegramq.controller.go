package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
	"io"
	"net/http"
	"strconv"
	"strings"

	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func generateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)[:length]
	return randomString, nil
}

func ProfileActivity(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	//time.Sleep(1 * time.Second)
	var user models.User
	result := initializers.DB.Where("tid = ?", msg.From.ID).First(&user).Preload("billing")
	fmt.Println(result)

	if result.Error != nil {
		log.Printf("Failed to fetch user data: %s", result.Error)
		// Handle the error if needed
		return
	}

	response := fmt.Sprintf("Активность в сети на платформе в текущем месяце составляет: %s", user.Name)

	// Create a new message configuration
	message := tgbotapi.NewMessage(msg.Chat.ID, response)

	// Send the message
	_, err := bot.Send(message)
	if err != nil {
		log.Printf("Failed to send message: %s", err)
		// Handle the error if needed
	}
}

func BalanceProfile(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	fmt.Println()
	timeStr := "5.000 ₽"

	response := "Баланс профиля составляет: " + timeStr

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

func TryActivated(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, afterSpace string) {
	appConfig, _ := initializers.LoadConfig(".")

	var user models.User

	err := initializers.DB.Where("telegram_token = ?", afterSpace).First(&user).Error
	if err != nil {
		return
	}

	fileURL := "https://images.myru.online/default.jpg" // Set a default profile photo URL
	config := tgbotapi.UserProfilePhotosConfig{
		UserID: msg.From.ID,
		Limit:  1,
	}

	userProfilePhotos, err := bot.GetUserProfilePhotos(config)
	if err != nil {
		// Handle error
		_ = err
	}

	if len(userProfilePhotos.Photos) > 0 {
		// userProfilePhotos is not empty
		photo := userProfilePhotos.Photos[0][0]

		// Call the GetFileDirectURL method to get the direct URL of the file on Telegram's servers
		fileURL, err = bot.GetFileDirectURL(photo.FileID)
		if err != nil {
			log.Panic(err)
		}

		// Create a file with a unique name in the specified directory
		fileName := filepath.Base(fileURL)
		filePath := filepath.Join(appConfig.IMGStorePath, user.Storage, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()

		// Open the file URL for reading
		resp, err := http.Get(fileURL)
		if err != nil {
			return
		}
		// Copy the contents of the file URL to the created file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return
		}

		// Create a file with a unique name in the specified directory
		fileName = filepath.Base(fileURL)
		if msg.From != nil && msg.From.UserName != "" {
			// If the field exists, assign the value
			user.TelegramName = &msg.From.UserName
			user.Tid = msg.From.ID
			user.TelegramActivated = true
			// user.Photo = fileURL
			user.Photo = user.Storage + `/` + fileName

			err = utils.SendPersonalMessageToClient(user.Session, "Activated")
			if err != nil {
				// handle error
				_ = err
			}

			if err := initializers.DB.Save(&user).Error; err != nil {
				return
			}

			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Спасибо @"+msg.From.UserName+" аккаунт активирован!"))

		} else {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка! Присвойте псевдоним в телеграм и повторите попытку"))
			// If the field doesn't exist, return or handle the situation accordingly
			return
		}

		// user.Name = msg.From.UserName
		// user.Tid = msg.From.ID
		// user.TelegramActivated = true
		// user.Photo = user.Storage + `/` + fileName

		// err = utils.SendPersonalMessageToClient(user.Session, "Activated")
		// if err != nil {
		// 	// handle error
		// }

		// if err := initializers.DB.Save(&user).Error; err != nil {
		// 	return
		// }

		// bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Спасибо @"+msg.From.UserName+" аккаунт активирован!"))

		defer resp.Body.Close()

	} else {
		src := filepath.Join(appConfig.IMGStorePath, "default.jpg")
		dst := filepath.Join(appConfig.IMGStorePath, user.Storage, "default.jpg")
		if err := os.Symlink(src, dst); err != nil {
			// Handle error
			_ = err
		}
		// Create a file with a unique name in the specified directory
		fileName := filepath.Base(fileURL)
		if msg.From != nil && msg.From.UserName != "" {
			// If the field exists, assign the value
			user.Name = msg.From.UserName
			user.Tid = msg.From.ID
			user.TelegramActivated = true
			// user.Photo = fileURL
			user.Photo = user.Storage + `/` + fileName

			err = utils.SendPersonalMessageToClient(user.Session, "Activated")
			if err != nil {
				// handle error
				_ = err
			}

			if err := initializers.DB.Save(&user).Error; err != nil {
				return
			}

			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Спасибо @"+msg.From.UserName+" аккаунт активирован!"))

		} else {
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Ошибка! Присвойте псевдоним в телеграм и повторите попытку"))
			// If the field doesn't exist, return or handle the situation accordingly
			return
		}

	}

}

func MakeCodes(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, afterSpace string) {
	var user models.User

	err := initializers.DB.Where("tid = ?", msg.From.ID).First(&user).Error
	if err != nil {
		return
	}

	fmt.Println(user.Name)

	// Split the afterSpace string into individual values
	values := strings.Split(afterSpace, ",")

	// Ensure that there are exactly two values
	if len(values) != 2 {
		fmt.Println("Invalid input: expected two values separated by a comma")
		return
	}

	// Retrieve the number of codes to generate from the first value
	numOfCodesStr := values[0]
	numOfCodes, err := strconv.Atoi(numOfCodesStr)
	if err != nil {
		fmt.Println("Invalid input:", err)
		return
	}

	// Parse the amount value from the second value
	amountPerCodeStr := values[1]
	amountPerCode, err := strconv.Atoi(amountPerCodeStr)
	if err != nil {
		fmt.Println("Invalid input:", err)
		return
	}

	for i := 0; i < numOfCodes; i++ {
		// Generate a random unique string for the code
		randomString, err := generateRandomString(10)
		if err != nil {
			fmt.Println("Error generating random string:", err)
			continue
		}

		// Create a new code with the specified balance
		code := models.Codes{
			Code:      randomString,
			Balance:   strconv.Itoa(amountPerCode), // Convert amountPerCode to a string
			Activated: false,
			UserId:    user.ID,
			Used:      0,
		}

		// Save the code to the database
		err = initializers.DB.Create(&code).Error
		if err != nil {
			fmt.Println("Error creating code:", err)
			continue
		}

		fmt.Println("Code created:", code.Code)

		// Send the code creation status back to the user
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, code.Code))
	}

}
