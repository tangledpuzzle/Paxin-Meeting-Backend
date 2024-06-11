package utils

import (
	"encoding/json"
	"hyperpage/initializers"
	"hyperpage/models"

	"log"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	uuid "github.com/satori/go.uuid"
)

func MoveToArch(bot *tgbotapi.BotAPI) {
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)

	var blogs []models.Blog
	initializers.DB.Where("expired_At < ?", time.Now()).Where("status = ?", "ACTIVE").Find(&blogs)

	for _, blog := range blogs {
		tmid := int(blog.TmId)
		deleteMsg := tgbotapi.NewDeleteMessage(-1001638837209, tmid)
		bot.Send(deleteMsg)

		// initializers.DB.Delete(&blog)
		blog.Status = "ARCHIVED"
		initializers.DB.Save(&blog)
		// Get the user_id from the blog record
		userID := blog.UserID
		// Fetch the corresponding user's data from the users table
		var user models.User
		if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			log.Printf("Failed to fetch user data: %s", err)
			// Handle the error if needed
			continue // Skip to the next blog in case of an error
		}
		msgText := "Здравствуйте, " + user.Name + "! Пост " + blog.Title + " отправлен в архив, вы можете продлить его из личного кабинета в течении 2 месяцев."

		// Declare a queue
		queueName := "post_arch"                          // Replace with your desired queue name
		conn, ch := initializers.ConnectRabbitMQ(&config) // Create a new connection and channel for each request

		_, err := ch.QueueDeclare(
			queueName,
			false, // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			log.Printf("Failed to declare a queue: %s", err)
			// Handle the error if needed
		}

		// Publish a message to the declared queue
		message := "Hello, client" // Replace with your desired message
		err = PublishMessage(ch, queueName, message)
		if err != nil {
			log.Printf("Failed to publish message: %s", err)
			// Handle the error if needed
		}

		// Consume messages from the declared queue
		err = ConsumeMessages(ch, conn, queueName)
		if err != nil {
			log.Printf("Failed to consume messages: %s", err)
			// Handle the error if needed
		}

		// Start consuming messages in a separate goroutine
		go func() {
			// Call the controller to process the message
			privateMsg := tgbotapi.NewMessage(user.Tid, msgText)
			// Send the private message
			_, err = bot.Send(privateMsg)
			if err != nil {
				log.Println("Error sending private message:", err)
			}

		}()
	}
}
func CheckExpiration(bot *tgbotapi.BotAPI) {
	config, _ := initializers.LoadConfig(".")

	var blogs []models.Blog

	var transaction []models.Transaction

	initializers.DB.Where("expired_At < ?", time.Now()).Where("status = ?", "ACTIVE").Find(&blogs)

	var blogIDs []uint
	for _, blog := range blogs {
		blogIDs = append(blogIDs, uint(blog.ID))
	}

	var blogPhotos []models.BlogPhoto
	initializers.DB.Where("blog_id IN (?)", blogIDs).Find(&blogPhotos)

	for _, photo := range blogPhotos {
		var files []struct {
			Path string `json:"path"`
		}

		data, err := photo.Files.Value()
		if err != nil {
			// Handle error if getting JSON data fails
			continue
		}

		err = json.Unmarshal(data.([]byte), &files)
		if err != nil {
			// Handle error if JSON unmarshaling fails
			continue
		}

		for _, file := range files {
			dst := filepath.Join(config.IMGStorePath, file.Path)
			err := os.Remove(dst)
			if err != nil {
				// Handle error if file removal fails
			}
		}
		// Delete the blog_photos record
		initializers.DB.Delete(&photo)

	}
	// Delete the blogs records

	initializers.DB.Model(&transaction).Where("element_id IN (?)", initializers.DB.Table("blogs").Select("id").Where("expired_At < ? AND status = ?", time.Now(), "ACTIVE")).Update("status", "CLOSED_0")

	for _, blog := range blogs {

		tmid := int(blog.TmId)
		deleteMsg := tgbotapi.NewDeleteMessage(-1001638837209, tmid)
		bot.Send(deleteMsg)

		initializers.DB.Delete(&blog)
	}
}

func CheckPlan(bot *tgbotapi.BotAPI) {
	var users []models.User

	// Retrieve all users
	initializers.DB.Find(&users)

	var userIds []uuid.UUID
	for _, user := range users {
		// Check if expired_plan_at is not null and less than the current time
		if user.ExpiredPlanAt != nil && user.ExpiredPlanAt.Before(time.Now()) {
			userIds = append(userIds, user.ID)

			// Update plan, signed, expired_plan_at, and limit_storage columns
			err := initializers.DB.Model(&models.User{}).
				Where("id = ?", user.ID).
				Updates(map[string]interface{}{
					"plan":            "standart",
					"signed":          false,
					"expired_plan_at": nil,
					"limit_storage":   20,
				}).Error
			if err != nil {
				// Handle error if necessary
				// ...
			}
		}
	}

	// Further processing or notifications
	// ...
}

func CheckSite(bot *tgbotapi.BotAPI) {
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)

	var blogs []models.Domain
	initializers.DB.Where("expired_At < ?", time.Now()).Where("status = ?", "activated").Find(&blogs)

	for _, blog := range blogs {
		// initializers.DB.Delete(&blog)
		blog.Status = "disabled"
		initializers.DB.Save(&blog)
		// Get the user_id from the blog record
		userID := blog.UserID
		// Fetch the corresponding user's data from the users table
		var user models.User
		if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			log.Printf("Failed to fetch user data: %s", err)
			// Handle the error if needed
			continue // Skip to the next blog in case of an error
		}
		msgText := "Здравствуйте, " + user.Name + "! Ваш веб-сайт " + "https://" + blog.Username + ".myru.online" + " приостановлен, вы можете продлить работу в личном кабинете."

		// Declare a queue
		queueName := "site_disabled"                      // Replace with your desired queue name
		conn, ch := initializers.ConnectRabbitMQ(&config) // Create a new connection and channel for each request

		_, err := ch.QueueDeclare(
			queueName,
			false, // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			log.Printf("Failed to declare a queue: %s", err)
			// Handle the error if needed
		}

		// Publish a message to the declared queue
		message := "Hello, client" // Replace with your desired message
		err = PublishMessage(ch, queueName, message)
		if err != nil {
			log.Printf("Failed to publish message: %s", err)
			// Handle the error if needed
		}

		// Consume messages from the declared queue
		err = ConsumeMessages(ch, conn, queueName)
		if err != nil {
			log.Printf("Failed to consume messages: %s", err)
			// Handle the error if needed
		}

		// Start consuming messages in a separate goroutine
		go func() {
			// Call the controller to process the message
			privateMsg := tgbotapi.NewMessage(user.Tid, msgText)
			// Send the private message
			_, err = bot.Send(privateMsg)
			if err != nil {
				log.Println("Error sending private message:", err)
			}

		}()
	}
}

func CheckSiteTime(bot *tgbotapi.BotAPI) {
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)

	fiveDaysFromNow := time.Now().AddDate(0, 0, 5)

	var blogs []models.Domain
	initializers.DB.Where("expired_At < ?", fiveDaysFromNow).Where("expired_At > ?", time.Now()).Where("status = ?", "activated").Find(&blogs)

	for _, blog := range blogs {
		// initializers.DB.Delete(&blog)
		// blog.Status = "disabled"
		// initializers.DB.Save(&blog)
		// Get the user_id from the blog record
		userID := blog.UserID
		// Fetch the corresponding user's data from the users table
		var user models.User
		if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			log.Printf("Failed to fetch user data: %s", err)
			// Handle the error if needed
			continue // Skip to the next blog in case of an error
		}
		msgText := "Здравствуйте, " + user.Name + "! Ваш веб-сайт " + "https://" + blog.Username + ".myru.online" + " будет приостановлен в работе менее чем через 5 дней, вы можете продлить работу веб-сайта в личном кабинете."

		// Declare a queue
		queueName := "site_disabled"                      // Replace with your desired queue name
		conn, ch := initializers.ConnectRabbitMQ(&config) // Create a new connection and channel for each request

		_, err := ch.QueueDeclare(
			queueName,
			false, // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			log.Printf("Failed to declare a queue: %s", err)
			// Handle the error if needed
		}

		// Publish a message to the declared queue
		message := "Hello, client" // Replace with your desired message
		err = PublishMessage(ch, queueName, message)
		if err != nil {
			log.Printf("Failed to publish message: %s", err)
			// Handle the error if needed
		}

		// Consume messages from the declared queue
		err = ConsumeMessages(ch, conn, queueName)
		if err != nil {
			log.Printf("Failed to consume messages: %s", err)
			// Handle the error if needed
		}

		// Start consuming messages in a separate goroutine
		go func() {
			// Call the controller to process the message
			privateMsg := tgbotapi.NewMessage(user.Tid, msgText)
			// Send the private message
			_, err = bot.Send(privateMsg)
			if err != nil {
				log.Println("Error sending private message:", err)
			}

		}()
	}
}
