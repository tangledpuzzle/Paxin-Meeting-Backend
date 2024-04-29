package controllers

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgtype"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"

	gt "github.com/bas24/googletranslatefree"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// test
type TimeEntry struct {
	Hour    int `json:"hour"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type TimeEntryScanner []TimeEntry

func (t TimeEntryScanner) Value() (driver.Value, error) {
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil
}

func (t *TimeEntryScanner) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		var entries []TimeEntry
		if err := json.Unmarshal(v, &entries); err != nil {
			return err
		}
		*t = entries
		return nil
	case string:
		var entries []TimeEntry
		if err := json.Unmarshal([]byte(v), &entries); err != nil {
			return err
		}
		*t = entries
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for TimeEntryScanner: %T", value)
	}
}

type userResponse struct {
	ID                uuid.UUID        `json:"userID"`
	TId               int64            `json:"tid"`
	Online            bool             `json:"online"`
	Photo             string           `json:"photo"`
	Name              string           `json:"name"`
	Role              string           `json:"role"`
	OnlineHours       TimeEntryScanner `json:"online_hours"`
	TotalOnlineHours  TimeEntryScanner `json:"total_online_hours"`
	TotalBlogs        int              `json:"totalblogs"`
	TotalRestBlogs    int              `json:"totalrestblog"`
	TelegramName      string           `json:"telegramname"`
	TelegramActivated bool             `json:"telegramactivated"`
	IsBot             bool             `json:"is_bot"`
}

type CategoryJSON struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CityJSON struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type UserProfileJSON struct {
	MultilangDescr models.MultilangTitle `json:"multilangtitle"`

	// Add other fields from the user profile as needed
}

type blogResponse struct {
	ID               uint64                `json:"id"`
	Title            string                `json:"title"`
	Descr            string                `json:"descr"`
	Slug             string                `json:"slug"`
	Status           string                `json:"status"`
	MultilangTitle   models.MultilangTitle `json:"multilangtitle"`
	MultilangDescr   models.MultilangTitle `json:"multilangdescr"`
	MultilangContent models.MultilangTitle `json:"multilangcontent"`
	Total            float64               `json:"total"`
	Content          string                `json:"content"`
	Lang             string                `json:"lang"`
	Views            int                   `json:"views"`
	UserAvatar       string                `json:"userAvatar"`
	Photos           []models.BlogPhoto    `json:"photos"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
	User             userResponse          `json:"user"`
	City             []CityJSON            `json:"city"`
	Pined            bool                  `json:"pined"`
	Catygory         []CategoryJSON        `json:"catygory"`
	UniqId           string                `json:"uniqId"`
	Sticker          string                `json:"sticker"`
	Hashtags         []string              `json:"hashtags"`
	UserProfile      UserProfileJSON       `json:"userProfile"`
}

func FilterBlogsWithIds(c *fiber.Ctx) error {
	// Define the struct to hold the request data
	type requestData struct {
		Ids       []string `json:"ids"`
		Publisher string   `json:"publisher"`
	}

	// Parse the POST request body into the struct
	var data requestData
	if err := c.BodyParser(&data); err != nil {
		// Handle parsing error
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Query to check if publisher (user) exists
	user := new(models.User)
	if err := initializers.DB.Where("id = ?", data.Publisher).First(user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid Publisher ID",
		})
	}

	// Query blogs with ids and publisher ID
	var blogs []models.Blog
	if err := initializers.DB.Where("user_id = ? AND id IN ? AND status = ?", user.ID, data.Ids, "ACTIVE").Find(&blogs).Error; err != nil {
		// Handle database query error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to query blogs",
			"error":   err.Error(),
		})
	}

	// Check if blogs were found
	if len(blogs) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "No blogs found matching the criteria",
		})
	}

	// Successful response with the found blogs
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"blogs":  blogs,
	})
}

func GetAllBlogs(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)
	isArchive := c.Query("isArchive")
	language := c.Query("language")

	if language == "" {
		language = "en"
	}

	var blogs []models.Blog
	query := initializers.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Preload("Photos").Preload("Hashtags").Preload("City.Translations", "language = ?", language).Preload("Catygory.Translations", "language = ?", language)

	if isArchive == "true" {
		query = query.Where("status = ?", "ARCHIVED")
	} else {
		query = query.Where("status = ?", "ACTIVE")
	}

	err := utils.Paginate(c, query.Find(&blogs), &blogs)
	if err != nil {
		return err
	}

	return nil
}

func SendToArchive(c *fiber.Ctx) error {
	blogID := c.Params("id")

	var blogData interface{}

	// Parse the request body and store the data in the blogData variable
	if err := json.Unmarshal(c.Body(), &blogData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	data := map[string]bool{
		"gooDeal":   true,
		"isArchive": true,
	}

	user := c.Locals("user")
	userResp := user.(models.UserResponse)

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}

	var blog models.Blog
	err := initializers.DB.Where("id = ?", blogID).First(&blog).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not get blog",
		})
	}

	// Check if user is the owner of the blog post or has admin rights
	if userObj.Role != "admin" && blog.UserID != userObj.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	newExpiredAt := blog.ExpiredAt.AddDate(0, 2, 0)

	gooDealValue := data["gooDeal"] // Access the value for the key "gooDeal"
	// Use the retrieved values as needed
	if gooDealValue {
		blog.Status = "ARCHIVED"
		blog.ExpiredAt = &newExpiredAt
		if err := initializers.DB.Save(&blog).Error; err != nil {
			log.Println("Could not update blog", err)
		}
	} else {
		blog.Status = "ARCHIVED"
		blog.ExpiredAt = &newExpiredAt
		if err := initializers.DB.Save(&blog).Error; err != nil {
			log.Println("Could not update blog", err)
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "ok",
	})
}

func generateUniqueID() string {
	// Generate a random byte slice for the unique ID
	idBytes := make([]byte, 8)
	_, err := rand.Read(idBytes)
	if err != nil {
		// Handle the error
		// For simplicity, we'll return an empty string
		return ""
	}

	// Encode the random bytes to base64 URL encoding
	encodedID := base64.RawURLEncoding.EncodeToString(idBytes)

	// Remove any trailing padding characters from the base64 encoding
	encodedID = strings.TrimRight(encodedID, "=")

	return encodedID
}

func Get10RandomBlogHashtags(c *fiber.Ctx) error {
	var blogHashtags []models.Hashtags
	if err := initializers.DB.Order("RANDOM()").Limit(10).Find(&blogHashtags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch random blog hashtags from database",
		})
	}
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   blogHashtags,
	})
}

func CreateBlog(c *fiber.Ctx) error {
	configPath := "./app.env"
	config, _ := initializers.LoadConfig(configPath)

	// Parse request body into a new Blog object
	blog := new(models.Blog)
	if err := c.BodyParser(blog); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// replace special characters in blog.Slug
	blog.Slug = replaceSpecialChars(blog.Slug)

	// Fetch languages from the database
	var langs []models.Langs
	err := initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	translations := make(map[string]string)
	translationsDescr := make(map[string]string)
	translationsContent := make(map[string]string)

	for _, lang := range langs {

		result, _ := gt.Translate(blog.Title, blog.Lang, lang.Code)
		translations[lang.Code] = result

		resultDescr, _ := gt.Translate(blog.Descr, blog.Lang, lang.Code)
		translationsDescr[lang.Code] = resultDescr

		resultContent, _ := gt.Translate(blog.Content, blog.Lang, lang.Code)
		translationsContent[lang.Code] = resultContent
	}

	// Set the translated values in the TitleLangs field
	blog.MultilangTitle.En = translations["en"]
	blog.MultilangTitle.Ru = translations["ru"]
	blog.MultilangTitle.Ka = translations["ka"]
	blog.MultilangTitle.Es = translations["es"]

	// Set the translated values in the TitleLangs field
	blog.MultilangDescr.En = translationsDescr["en"]
	blog.MultilangDescr.Ru = translationsDescr["ru"]
	blog.MultilangDescr.Ka = translationsDescr["ka"]
	blog.MultilangDescr.Es = translationsDescr["es"]

	// Set the translated values in the TitleLangs field
	blog.MultilangContent.En = translationsContent["en"]
	blog.MultilangContent.Ru = translationsContent["ru"]
	blog.MultilangContent.Ka = translationsContent["ka"]
	blog.MultilangContent.Es = translationsContent["es"]

	// Retrieve associated Hashtags from the database
	hashtags := []models.Hashtags{}
	for _, tag := range blog.Hashtags {
		// Retrieve the Hashtags record from the database based on tag.Hashtag
		hashtag := models.Hashtags{}
		if err := initializers.DB.Where("hashtag = ?", tag.Hashtag).First(&hashtag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				hashtag.Hashtag = tag.Hashtag
				if createErr := initializers.DB.Create(&hashtag).Error; createErr != nil {
					_ = err
					continue
				}
			} else {
				_ = err
				continue
			}
		}
		hashtags = append(hashtags, hashtag)
	}

	// Assign the retrieved Hashtags to the Blog instance
	blog.Hashtags = hashtags

	// cities := []models.City{}
	// for _, cityName := range blog.City {

	// 	// Retrieve the City record from the database based on cityName
	// 	city := models.City{}
	// 	if err := initializers.DB.Where("name = ?", cityName.Name).First(&city).Error; err != nil {
	// 		// Handle the error if the city is not found
	// 		_ = err
	// 	}
	// 	cities = append(cities, city)
	// }

	// blog.City = cities

	// Validate the required fields in the blog object

	// Assuming you have a field in the blog struct to store these translations, set it here.

	//blog.Content == "" ||

	if blog.Descr == "" || blog.Title == "" || blog.Days == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required fields in the request body",
		})
	}

	// config, _ := initializers.LoadConfig(".")

	// cfg := &initializers.Config{
	// 	TELEGRAM_TOKEN:   config.TELEGRAM_TOKEN,
	// 	TELEGRAM_CHANNEL: config.TELEGRAM_CHANNEL,
	// }

	// bot, err := initializers.ConnectTelegram(cfg)
	// if err != nil {
	// 	log.Panic(err)
	// }

	uniqueID := generateUniqueID()

	// sessionID := c.Query("session")

	// if sessionID == "" {
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"message": "Invalid session ID",
	// 	})
	// }

	userId := c.Locals("user")
	userResp := userId.(models.UserResponse)
	userObj := models.User{
		ID: userResp.ID,
	}

	// Check if user exists
	user := new(models.User)
	if err := initializers.DB.Where("id = ?", userObj.ID).First(user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user ID",
		})
	}

	// Set user ID on blog object
	uid, err := uuid.FromString(user.ID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not create blogs",
		})
	}

	var commission float64

	// Add an amount variable to store the amount for the blog post
	switch blog.Days {
	case 0:
		commission = 0
	case 30:
		commission = 0
	case 60:
		commission = 0
	case 10:
		commission = 0
	case 90:
		commission = 0

	}

	total := blog.Total
	elementId := blog.ID

	// Commission * 0.05
	amount := commission
	module := "blog"
	if err := utils.DeductAmountFromUserBalance(userObj.ID, amount, total, module, elementId); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Insufficient balance",
		})
	}

	if user.TelegramActivated {
		queueName := "blog_activity"                      // Replace with your desired queue name
		conn, ch := initializers.ConnectRabbitMQ(&config) // Create a new connection and channel for each request

		_, err = ch.QueueDeclare(
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
		message := "Blog received!" // Replace with your desired message
		err = utils.PublishMessage(ch, queueName, message)
		if err != nil {
			log.Printf("Failed to publish message: %s", err)
			// Handle the error if needed
		}

		// Consume messages from the declared queue
		err = utils.ConsumeMessages(ch, conn, queueName)
		if err != nil {
			log.Printf("Failed to consume messages: %s", err)
			// Handle the error if needed
		}

		var wg sync.WaitGroup
		wg.Add(1)

		// Start consuming messages in a separate goroutine
		go func() {
			defer wg.Done()

			Data := blog

			type forBot struct {
				Name     string  `json:"title"`
				Url      string  `json:"url"`
				Cat      string  `json:"cat"`
				Username string  `json:"name"`
				City     string  `json:"city"`
				Total    float64 `json:"total"`
			}

			forbot := forBot{
				// City:     blog.City,
				// Cat:      Data.Catygory,
				Name:     Data.Title,
				Total:    Data.Total,
				Url:      "https://paxintrade.com/" + Data.UniqId + "/" + Data.Slug,
				Username: user.Name,
			}

			var addintinal = ""
			utils.UserActivity("newblog", forbot.Username, addintinal)

			// var msgText string
			// if blog.Total == 0 {
			// 	msgText = fmt.Sprintf("Новый пост:\nГород: %s \nРубрика: %s \nЗаголовок: %s \nURL: %s\nАвтор: @%s", forbot.City, forbot.Cat, forbot.Name, forbot.Url, forbot.Username)
			// } else {
			// 	msgText = fmt.Sprintf("Новый пост:\nГород: %s \nРубрика: %s \nЗаголовок: %s \nЦена: %.2f ₽ \nURL: %s\nАвтор: @%s", forbot.City, forbot.Cat, forbot.Name, forbot.Total, forbot.Url, forbot.Username)
			// }

			// // Create a new message for the user's private chat
			// privateMsg := tgbotapi.NewMessage(user.Tid, msgText)

			// // Send the private message
			// _, err = bot.Send(privateMsg)
			// if err != nil {
			// 	log.Println("Error sending private message:", err)
			// }

			// if userResp.Tcid == 0 {

			// 	// Set up an image file to send
			// 	absolutePath := filepath.Join(config.IMGStorePath, "default.jpg")
			// 	imageFile, err := os.Open(absolutePath)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}
			// 	defer imageFile.Close()

			// 	// Read the image file into a byte slice
			// 	imageInfo, _ := imageFile.Stat()
			// 	imageBytes := make([]byte, imageInfo.Size())
			// 	_, err = imageFile.Read(imageBytes)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// 	// Replace with the chat ID you want to send the image to
			// 	// chatID := int64(-userResp.Tcid)

			// 	// Create a photo message configuration
			// 	// Create a photo message configuration
			// 	photoConfig := tgbotapi.NewPhoto(user.Tid, tgbotapi.FileBytes{
			// 		Name:  "image.jpg",
			// 		Bytes: imageBytes,
			// 	})

			// 	photoConfig.Caption = fmt.Sprintf("Новый пост:\nГород: %s \nРубрика: %s \nЗаголовок: %s \nURL: %s\nАвтор: @%s", forbot.City, forbot.Cat, forbot.Name, forbot.Url, forbot.Username)

			// 	// Send the photo
			// 	_, err = bot.Send(photoConfig)
			// 	if err != nil {
			// 		log.Fatal(err)
			// 	}

			// 	privateMsg := tgbotapi.NewMessage(user.Tid, msgText)
			// 	_, err = bot.Send(privateMsg)
			// 	if err != nil {
			// 		log.Println("Error sending private message:", err)
			// 	}
			// }

			// Call the controller to process the message
			// msg := tgbotapi.NewMessage(int64(-userResp.Tcid), msgText)
			// sentMessage, err := bot.Send(msg)
			// if err != nil {
			// 	// Check if the error is due to a not found or kicked chat
			// 	if strings.Contains(err.Error(), "chat not found") || strings.Contains(err.Error(), "bot was kicked") || strings.Contains(err.Error(), "chat_id is empty") {
			// 		fmt.Println("Chat not found or bot was kicked, skipping send message.")
			// 	} else {
			// 		// Handle other errors
			// 		log.Panic(err)
			// 	}
			// }
			// Prepare the private message
			blog.UserID = uid
			blog.ExpiredAt = new(time.Time)
			blog.UniqId = uniqueID

			if blog.Days == 10 {
				expDate := time.Now().AddDate(10, 0, 0) // Add 10 years
				blog.ExpiredAt = &expDate
			} else {
				expDate := time.Now().AddDate(0, 0, blog.Days)
				blog.ExpiredAt = &expDate
			}

			blog.UserAvatar = user.Photo

			user.TotalBlogs += 1

			if blog.Total == 0 {
				blog.NotAds = true
			} else {
				blog.NotAds = false
			}

			if err := initializers.DB.Save(&user).Error; err != nil {
				log.Println("Could not update user's total blogs count:", err)
			}

			// Create blog record in database
			if err := initializers.DB.Create(&blog).Error; err != nil {
				log.Println("Could not create blog:", err)
			}
		}()

		// Wait for the goroutine to complete
		wg.Wait()

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   blog,
		})
	}

	blog.UserID = uid
	blog.ExpiredAt = new(time.Time)
	blog.UniqId = uniqueID

	if blog.Days == 10 {
		expDate := time.Now().AddDate(10, 0, 0) // Add 10 years
		blog.ExpiredAt = &expDate
	} else {
		expDate := time.Now().AddDate(0, 0, blog.Days)
		blog.ExpiredAt = &expDate
	}

	blog.UserAvatar = user.Photo

	user.TotalBlogs += 1

	if blog.Total == 0 {
		blog.NotAds = true
	} else {
		blog.NotAds = false
	}

	if err := initializers.DB.Save(&user).Error; err != nil {
		log.Println("Could not update user's total blogs count:", err)
	}

	// Create blog record in database
	if err := initializers.DB.Create(&blog).Error; err != nil {
		log.Println("Could not create blog:", err)
	}

	fmt.Println("END2")

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   blog,
	})

}

func formatPriceWithDots(price int) string {
	formattedPrice := strconv.Itoa(price)
	n := len(formattedPrice)

	var formattedPriceWithDots string
	for i := 0; i < n; i++ {
		formattedPriceWithDots += string(formattedPrice[i])
		if (n-i-1)%3 == 0 && i != n-1 {
			formattedPriceWithDots += "."
		}
	}

	return formattedPriceWithDots
}
func SendBotCallRequest(c *fiber.Ctx) error {
	// Define the struct to hold the message data
	type MessageData struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Tid      int    `json:"blogdata"`
		Uid      string `json:"uid"`
		Slug     string `json:"slug"`
		Category string `json:"category"`
		Price    int    `json:"price"`
	}

	// Parse the POST request body into the struct
	var messageData MessageData
	if err := c.BodyParser(&messageData); err != nil {
		// Handle parsing error
		return err
	}
	config2, _ := initializers.LoadConfig(".")
	cfg := &initializers.Config{
		TELEGRAM_TOKEN: config2.TELEGRAM_TOKEN,
		SERVER_URL:     config2.SERVER_URL,
	}

	// Format the message data
	URL := fmt.Sprintf("%s/%s/%s", config2.SERVER_URL, messageData.Uid, messageData.Slug)
	formattedPrice := formatPriceWithDots(messageData.Price)

	// Start the goroutine to send the message
	go func() {
		// Create the Telegram bot
		bot, err := initializers.ConnectTelegram(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer bot.StopReceivingUpdates()

		chatID := int64(messageData.Tid)

		// Create the message text
		msgText := fmt.Sprintf("Здравствуйте, для вас новый запрос!\nИмя: %s\nТелефон: %s\nСсылка: %s\nКатегория: %s\nЦена: %s ₽",
			messageData.Name, messageData.Phone, URL, messageData.Category, formattedPrice)

		msg := tgbotapi.NewMessage(chatID, msgText)

		// Send the message
		_, err = bot.Send(msg)
		if err != nil {
			// Check if the error message indicates that the bot was blocked by the user
			if err.Error() == "Forbidden: bot was blocked by the user" {
				// Handle the case where the bot is blocked by the user
				fmt.Println("Bot was blocked by the user")
				return
			}
			log.Fatal(err)
		}

		// Call the controller to process the message
		// ...
	}()

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "ok",
	})
}

func AddHashTag(c *fiber.Ctx) error {

	var hashtag models.Hashtags
	if err := c.BodyParser(&hashtag); err != nil {
		return err
	}

	// Save the hashtag to the database
	if err := initializers.DB.Create(&hashtag).Error; err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   hashtag,
	})
}

func SearchHashTag(c *fiber.Ctx) error {

	// Get the name query parameter
	name := c.Query("name")
	fmt.Println("sss", name)
	// Check if the name is provided
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Name parameter is required",
		})
	}

	// Find the cities with names similar to the search query (case-insensitive)
	var hashtags []models.Hashtags
	if err := initializers.DB.Where("hashtag ILIKE ?", "%"+name+"%").Find(&hashtags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from database",
		})
	}

	// Return the matched cities as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   hashtags,
	})
}

func AddBlogTime(c *fiber.Ctx) error {
	var blog models.Blog

	type PostData struct {
		ID    int    `json:"id"`
		Days  string `json:"days"`
		Price string `json:"price"`
	}

	// Parse the POST request body into the struct
	var postData PostData
	if err := c.BodyParser(&postData); err != nil {
		// Handle parsing error
		return err
	}

	// Access the parsed data
	days := postData.Days
	price := postData.Price
	id := postData.ID
	// Convert the string price to a float64
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid price value",
		})
	}

	err = initializers.DB.Where("id = ?", id).First(&blog).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Element not found",
		})
	}
	// Parse request body into a new BlogSearch object
	user := c.Locals("user")
	if user == nil {
		// Handle the case when user is nil
		return errors.New("user not found")
	}

	userResp, ok := user.(models.UserResponse)
	if !ok {
		// Handle the case when user is not of type models.UserResponse
		return errors.New("invalid user type")
	}

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}

	// Check if user is the owner of the blog post or has admin rights
	if userObj.Role != "admin" && blog.UserID != userObj.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	// Convert the string days to an integer
	daysInt, err := strconv.Atoi(days)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid days value",
		})
	}

	// Fetch the current balance
	var billing models.Billing
	err = initializers.DB.Where("user_id = ?", userObj.ID).First(&billing).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch balance",
		})
	}

	// Check if the balance is sufficient
	if billing.Amount < priceFloat {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Insufficient balance",
		})
	}
	// Update the amount field in the balance table
	err = initializers.DB.Model(&models.Billing{}).
		Where("user_id = ?", userObj.ID).
		Updates(map[string]interface{}{
			"amount": gorm.Expr("amount - ?", priceFloat),
		}).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update balance",
		})
	}

	transaction := models.Transaction{
		UserID:      userObj.ID,
		Total:       "0", // Update with the appropriate value
		Amount:      priceFloat,
		Description: "Оплата за продление размещения",
		Module:      "addTimeBlog",
		Type:        "deduction",
		Status:      "CLOSED_1",
	}

	err = initializers.DB.Create(&transaction).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create transaction",
		})
	}

	isArchive := c.Query("isArchive")
	if isArchive == "true" {
		// Set the expired_at date to the current date
		newExpiredAt := time.Now()
		newExpiredAt = newExpiredAt.AddDate(0, 0, daysInt)
		blog.ExpiredAt = &newExpiredAt
		// Update the expired_at column in the database
		err := initializers.DB.Model(&blog).Update("expired_at", newExpiredAt).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to update expired_at value",
			})
		}
	}

	// Add the specified number of days to the existing expired_at value
	newExpiredAt := blog.ExpiredAt.AddDate(0, 0, daysInt)

	// Update the expired_at, days, and status columns in the database
	err = initializers.DB.Model(&blog).Updates(map[string]interface{}{
		"expired_at": newExpiredAt,
		"days":       daysInt,
		"status":     "ACTIVE",
	}).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update expired_at and days values",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "ok",
	})

}

func SearchBlogByTitle(c *fiber.Ctx) error {
	// Parse request body into a new BlogSearch object
	user := c.Locals("user")
	if user == nil {
		// Handle the case when user is nil
		return errors.New("user not found")
	}

	userResp, ok := user.(models.UserResponse)
	if !ok {
		// Handle the case when user is not of type models.UserResponse
		return errors.New("invalid user type")
	}

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}

	title := strings.ToLower(c.FormValue("title"))

	// Search blogs by title and user ID
	blogs := make([]*models.Blog, 0)
	err := utils.Paginate(c, initializers.DB.
		Where("LOWER(title) LIKE ? AND user_id = ?", "%"+title+"%", userObj.ID).
		Order("created_at DESC").
		Preload("User").
		Preload("Photos"), &blogs)

	if err != nil {
		// Handle the error appropriately
		return err
	}

	// Handle the retrieved blogs
	// ...

	return nil
}

func CreateBlogPhoto(c *fiber.Ctx) error {
	// Get the blog ID from the request query parameter
	blogID, err := strconv.ParseUint(c.Query("blogID"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid blog ID",
		})
	}

	type File struct {
		Path string `json:"path" validate:"required"`
	}

	// Parse the request body
	var reqBody struct {
		Files []File `json:"files" validate:"required"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request body",
		})
	}

	// Convert []File to pgtype.JSONB
	filesJSON := pgtype.JSONB{}
	if err := filesJSON.Set(reqBody.Files); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error converting files to JSON",
		})
	}

	// Create a new BlogPhoto object
	blogPhoto := models.BlogPhoto{
		BlogID: blogID,
		Files:  filesJSON,
	}

	// Validate the BlogPhoto object
	if err := blogPhoto.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation error",
			"errors":  err.Error(),
		})
	}

	// Save the BlogPhoto object to the database
	if err := initializers.DB.Create(&blogPhoto).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error saving to database",
		})
	}

	var path string

	if len(blogPhoto.Files.Bytes) > 0 {
		var jsonData []map[string]interface{}
		if err := json.Unmarshal(blogPhoto.Files.Bytes, &jsonData); err != nil {
			// Handle error
			_ = err
		}

		if len(jsonData) > 0 {
			if p, ok := jsonData[0]["path"].(string); ok {
				path = p
			}
		}
	}

	userId := c.Locals("user")
	userResp := userId.(models.UserResponse)
	userObj := models.User{
		ID: userResp.ID,
	}

	// Check if user exists
	user := new(models.User)
	if err := initializers.DB.Where("id = ?", userObj.ID).First(user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user ID",
		})
	}

	var blog models.Blog
	if err := initializers.DB.Where("id = ?", blogID).First(&blog).Preload("Hashtags").Error; err != nil {
		// Handle the error
		log.Println("Error:", err)
	}

	hashtags := make([]string, len(blog.Hashtags))
	for i, tag := range blog.Hashtags {
		hashtags[i] = tag.Hashtag
	}

	var wg sync.WaitGroup

	if user.TelegramActivated {
		wg.Add(1)

		// Start consuming messages in a separate goroutine
		go func() {

			defer wg.Done()

			type forBot struct {
				Name     string   `json:"title"`
				Url      string   `json:"url"`
				Cat      string   `json:"cat"`
				Username string   `json:"name"`
				City     string   `json:"city"`
				Total    float64  `json:"total"`
				Hashtags []string `json:"hashtags"`
			}

			forbot := forBot{
				// City:     blog.City,
				// Cat:      blog.Catygory,
				Name:     blog.Title,
				Total:    blog.Total,
				Hashtags: hashtags,
				Url:      "https://" + user.Name + ".paxintrade.com/" + blog.UniqId + "/" + blog.Slug,
				// Username: user.Name,
			}

			var addintinal = ""
			utils.UserActivity("newblog", forbot.Username, addintinal)

			var msgText string
			if blog.Total == 0 {
				msgText = fmt.Sprintf("\nГород: %s \nРубрика: %s \nЗаголовок: %s \nURL: %s\nАвтор: @%s", forbot.City, forbot.Cat, forbot.Name, forbot.Url, forbot.Username)
			} else {
				msgText = fmt.Sprintf("\nГород: %s \nРубрика: %s \nЗаголовок: %s \nЦена: %.2f ₽ \nURL: %s\nАвтор: @%s", forbot.City, forbot.Cat, forbot.Name, forbot.Total, forbot.Url, forbot.Username)
			}

			if len(forbot.Hashtags) > 0 {
				hashtagsText := "\nХештеги: " + strings.Join(forbot.Hashtags, ", ")
				msgText += hashtagsText
			}

			// // Create a new message for the user's private chat
			// privateMsg := tgbotapi.NewMessage(user.Tid, msgText)

			// // Send the private message
			// _, err = bot.Send(privateMsg)
			// if err != nil {
			// 	log.Println("Error sending private message:", err)
			// }

			config, _ := initializers.LoadConfig(".")

			cfg := &initializers.Config{
				TELEGRAM_TOKEN:   config.TELEGRAM_TOKEN,
				TELEGRAM_CHANNEL: config.TELEGRAM_CHANNEL,
			}

			bot, err := initializers.ConnectTelegram(cfg)
			if err != nil {
				log.Panic(err)
			}

			if userResp.Tcid == 0 {

				// Set up an image file to send
				absolutePath := filepath.Join(config.IMGStorePath, path)
				imageFile, err := os.Open(absolutePath)
				if err != nil {
					log.Fatal(err)
				}
				defer imageFile.Close()

				// Read the image file into a byte slice
				imageInfo, _ := imageFile.Stat()
				imageBytes := make([]byte, imageInfo.Size())
				_, err = imageFile.Read(imageBytes)
				if err != nil {
					log.Fatal(err)
				}

				// Replace with the chat ID you want to send the image to
				// chatID := int64(-userResp.Tcid)

				// Create a photo message configuration
				// Create a photo message configuration
				photoConfig := tgbotapi.NewPhoto(user.Tid, tgbotapi.FileBytes{
					Name:  "image.jpg",
					Bytes: imageBytes,
				})

				photoConfig.Caption = fmt.Sprintf("\nГород: %s \nРубрика: %s \nЗаголовок: %s \nURL: %s", forbot.City, forbot.Cat, forbot.Name, forbot.Url)

				if len(forbot.Hashtags) > 0 {
					hashtagsText := "\nХештеги: " + strings.Join(forbot.Hashtags, ", ")
					photoConfig.Caption += hashtagsText
				}

				// Send the photo
				_, err = bot.Send(photoConfig)
				if err != nil {
					log.Fatal(err)
				}

				// privateMsg := tgbotapi.NewMessage(user.Tid, msgText)
				// _, err = bot.Send(privateMsg)
				// if err != nil {
				// 	log.Println("Error sending private message:", err)
				// }
			}

			// Call the controller to process the message

			// Set up an image file to send
			absolutePath := filepath.Join(config.IMGStorePath, path)
			imageFile, err := os.Open(absolutePath)
			if err != nil {
				log.Fatal(err)
			}
			defer imageFile.Close()

			// Read the image file into a byte slice
			imageInfo, _ := imageFile.Stat()
			imageBytes := make([]byte, imageInfo.Size())
			_, err = imageFile.Read(imageBytes)
			if err != nil {
				log.Fatal(err)
			}

			// Replace with the chat ID you want to send the image to
			// chatID := int64(-userResp.Tcid)

			// Create a photo message configuration
			// Create a photo message configuration
			photoConfig := tgbotapi.NewPhoto(-userResp.Tcid, tgbotapi.FileBytes{
				Name:  "image.jpg",
				Bytes: imageBytes,
			})

			photoConfig.Caption = fmt.Sprintf("\nГород: %s \nРубрика: %s \nЗаголовок: %s \nURL: %s", forbot.City, forbot.Cat, forbot.Name, forbot.Url)

			if len(forbot.Hashtags) > 0 {
				hashtagsText := "\nХештеги: " + strings.Join(forbot.Hashtags, ", ")
				photoConfig.Caption += hashtagsText
			}

			if user.TelegramActivated {
				// Send the photo
				sentMessage, err := bot.Send(photoConfig)
				if err != nil {
					// Check if the error is due to a not found or kicked chat
					if strings.Contains(err.Error(), "chat not found") || strings.Contains(err.Error(), "bot was kicked") || strings.Contains(err.Error(), "chat_id is empty") {
						fmt.Println("Chat not found or bot was kicked, skipping send message.")
					} else {
						// Handle other errors
						log.Panic(err)
					}
				}

				// msg := tgbotapi.NewMessage(int64(-userResp.Tcid), msgText)
				// sentMessage, err := bot.Send(msg)
				// if err != nil {
				// 	// Check if the error is due to a not found or kicked chat
				// 	if strings.Contains(err.Error(), "chat not found") || strings.Contains(err.Error(), "bot was kicked") || strings.Contains(err.Error(), "chat_id is empty") {
				// 		fmt.Println("Chat not found or bot was kicked, skipping send message.")
				// 	} else {
				// 		// Handle other errors
				// 		log.Panic(err)
				// 	}
				// }
				// Prepare the private message

				messageID := sentMessage.MessageID
				blog.TmId = float64(messageID)
			}

			// blog.UserAvatar = user.Photo

			// user.TotalBlogs += 1

			// if blog.Total == 0 {
			// 	blog.NotAds = true
			// } else {
			// 	blog.NotAds = false
			// }

			// if err := initializers.DB.Save(&user).Error; err != nil {
			// 	log.Println("Could not update user's total blogs count:", err)
			// }

			// Create blog record in database
			if err := initializers.DB.Save(&blog).Error; err != nil {
				log.Println("Could not create blog:", err)
			}
		}()

		wg.Wait()

		return c.JSON(fiber.Map{
			"message": "Blog photo created successfully",
		})
	}

	// Wait for the goroutine to complete

	fmt.Println("hello 2")

	return c.JSON(fiber.Map{
		"message": "Blog photo created successfully",
	})

}

func GetBlogById(c *fiber.Ctx) error {
	blogID := c.Params("id")
	uniqId := c.Get("name")
	language := c.Query("language")

	if language == "" {
		language = "en"
	}
	var blog []models.Blog

	err := utils.Paginate(c, initializers.DB.Where("slug = ? AND uniq_id = ?", blogID, uniqId).First(&blog).Preload("Catygory.Translations", "language = ?", language).Preload("City.Translations", "language = ?", language).Preload("Hashtags").Preload("Photos").Preload("User"), &blog)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Element not found",
		})
	}

	var res []*blogResponse
	for _, b := range blog {

		userID := b.User.ID
		var userProfile models.Profile
		err := initializers.DB.Where("user_id = ?", userID).First(&userProfile).Error

		if err != nil {
			// Handle or log the error
			fmt.Println("Error fetching user profile:", err)
		}

		b.Views++
		// Save the changes
		if err := initializers.DB.Save(&b).Error; err != nil {
			return err
		}
		hashtags := make([]string, len(b.Hashtags))
		for i, tag := range b.Hashtags {
			hashtags[i] = tag.Hashtag
		}

		cities := make([]CityJSON, len(b.City))
		for i, city := range b.City {
			cityJSON := CityJSON{
				ID:   city.ID,
				Name: city.Translations[i].Name,
			}
			cities[i] = cityJSON
		}

		categories := make([]CategoryJSON, len(b.Catygory))
		for i, category := range b.Catygory {
			categoryJSON := CategoryJSON{
				ID:   category.ID,
				Name: category.Translations[i].Name,
			}
			categories[i] = categoryJSON
		}

		userOnlineHours := make(TimeEntryScanner, len(b.User.OnlineHours))
		for i, entry := range b.User.OnlineHours {
			userOnlineHours[i] = TimeEntry{
				Hour:    entry.Hour,
				Minutes: entry.Minutes,
				Seconds: entry.Seconds,
			}
		}

		userTotalOnlineHours := make(TimeEntryScanner, len(b.User.TotalOnlineHours))
		for i, entry := range b.User.TotalOnlineHours {
			userTotalOnlineHours[i] = TimeEntry{
				Hour:    entry.Hour,
				Minutes: entry.Minutes,
				Seconds: entry.Seconds,
			}
		}
		var telegramNameVal string
		if b.User.TelegramName != nil {
			telegramNameVal = *b.User.TelegramName
		} else {
			telegramNameVal = ""
		}
		blogRes := &blogResponse{
			ID:               b.ID,
			Title:            b.Title,
			MultilangTitle:   b.MultilangTitle,
			MultilangDescr:   b.MultilangDescr,
			MultilangContent: b.MultilangContent,

			Descr:      b.Descr,
			Lang:       b.Lang,
			Slug:       b.Slug,
			Status:     b.Status,
			Total:      b.Total,
			Content:    b.Content,
			City:       cities,
			Views:      b.Views,
			UserAvatar: b.UserAvatar,
			Photos:     b.Photos,
			CreatedAt:  b.CreatedAt,
			UpdatedAt:  b.UpdatedAt,
			Catygory:   categories,
			Sticker:    b.Sticker,
			UserProfile: UserProfileJSON{
				MultilangDescr: userProfile.MultilangDescr,
			},
			User: userResponse{
				ID:                b.User.ID,
				TId:               b.User.Tid,
				Online:            b.User.Online,
				Photo:             b.User.Photo,
				Name:              b.User.Name,
				OnlineHours:       userOnlineHours,
				TotalOnlineHours:  userTotalOnlineHours,
				TotalRestBlogs:    b.User.TotalRestBlogs,
				TelegramName:      telegramNameVal,
				TelegramActivated: b.User.TelegramActivated,
				IsBot:             b.User.IsBot,
			},
			Hashtags: hashtags,
		}
		res = append(res, blogRes)
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   res,
	})

}

func DeleteBlog(c *fiber.Ctx) error {
	blogID := c.Params("id")

	user := c.Locals("user")
	userResp := user.(models.UserResponse)

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}

	var blog models.Blog
	err := initializers.DB.Where("id = ?", blogID).First(&blog).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	// Check if user is the owner of the blog post or has admin rights
	if userObj.Role != "admin" && blog.UserID != userObj.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	// Delete all hashtags associated with the blog post using raw SQL query
	query := "DELETE FROM blog_hashtags WHERE blog_id = ?"
	if err := initializers.DB.Exec(query, blogID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	// Delete all votes associated with the blog post using
	query_votes := "DELETE FROM votes WHERE blog_id = ?"
	if err := initializers.DB.Exec(query_votes, blogID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	// Delete all blog guilds associated with the blog post using raw SQL query
	query_guilds := "DELETE FROM blog_guilds WHERE blog_id = ?"
	if err := initializers.DB.Exec(query_guilds, blogID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	// Delete all blog city associated with the blog post using raw SQL query
	query_city := "DELETE FROM blog_city WHERE blog_id = ?"
	if err := initializers.DB.Exec(query_city, blogID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	// Delete all photos associated with the blog post
	var blogPhotos []models.BlogPhoto
	if err := initializers.DB.Where("blog_id = ?", blogID).Find(&blogPhotos).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	for _, blogPhoto := range blogPhotos {
		// Extract the "files" field from the blogPhoto into a pgtype.JSONB object
		var files pgtype.JSONB
		if err := initializers.DB.Raw(`SELECT files FROM blog_photos WHERE id = ?`, blogPhoto.ID).Scan(&files).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not retrieve files for deletion",
			})
		}

		// Decode the JSONB data (files) into a slice of strings
		var filePaths []struct {
			Path string `json:"path"`
		}
		if err := files.AssignTo(&filePaths); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not decode JSONB data",
			})
		}

		// Delete the associated files from the server
		for _, filePath := range filePaths {
			deleteFileFromServer(filePath.Path)
		}

		// Delete the blog photo entry from the "blog_photos" table
		if err := initializers.DB.Delete(&blogPhoto).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not delete element",
			})
		}
	}

	// Proceed with deleting the blog entry
	err = initializers.DB.Delete(&blog).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not delete element",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Element with ID %s has been deleted", blogID),
	})
}

// CLASIC GET WITOUT SHORTING

// func GetAll(c *fiber.Ctx) error {
//     var blogs []models.Blog
// err := utils.Paginate(c, initializers.DB.Order("created_at DESC").Preload("Photos").Preload("User"), &blogs)
// if err != nil {
//     return err
// }

// return nil

// }
func GetAll(c *fiber.Ctx) error {
	var blogs []models.Blog
	language := c.Query("language")

	if language == "" {
		language = "en"
	}

	query := initializers.DB.Order("created_at DESC").
		Preload("Catygory.Translations", "language = ?", language).
		Preload("City.Translations", "language = ?", language).
		Preload("Hashtags").
		Preload("Photos").
		Preload("User").
		Where("status = ?", "ACTIVE")

	// Get the query parameters
	city := c.Query("city")
	skip := c.Query("skip")
	money := c.Query("money")
	title := c.Query("title")
	hashtags := c.Query("hashtag")
	category := c.Query("category")

	if hashtags != "" && hashtags != "all" {
		// Split the hashtags into separate values
		hashtagValues := strings.Split(hashtags, ",")

		// Join the hashtag values with the '#' character
		hashtagValuesWithPrefix := make([]string, len(hashtagValues))
		for i, tag := range hashtagValues {
			hashtagValuesWithPrefix[i] = strings.TrimSpace(tag)
		}

		// Add the hashtags filter to the query
		query = query.Joins("JOIN blog_hashtags bh ON blogs.id = bh.blog_id").
			Joins("JOIN hashtags h ON bh.hashtags_id = h.id").
			Where("h.hashtag IN (?)", hashtagValuesWithPrefix)
	}

	if city != "" && city != "all" {
		// Сначала найдем city_id для указанного города и языка
		var cityTranslation models.CityTranslation
		initializers.DB.Where("name = ? AND language = ?", city, language).First(&cityTranslation)

		if cityTranslation.ID != 0 {
			// Создадим подзапрос для поиска всех blog_id, связанных с указанным city_id
			subQuery := initializers.DB.Table("blog_city").
				Select("blog_id").
				Where("city_id = ?", cityTranslation.CityID)

			// Добавим условие, чтобы ваш основной запрос включал только записи с blog_id из подзапроса
			query = query.Where("blogs.id IN (?)", subQuery) // Specify the table alias for "blogs.id"
		}
	}

	if category != "" && category != "all" {
		var guildTranslation models.GuildTranslation
		initializers.DB.Where("name = ? AND language = ?", category, language).First(&guildTranslation)
		if guildTranslation.ID != 0 {
			// Создадим подзапрос для поиска всех blog_id, связанных с указанным guild_id
			subQuery := initializers.DB.Table("blog_guilds").
				Select("blog_id").
				Where("guilds_id = ?", guildTranslation.GuildID)

			// Добавим условие, чтобы ваш основной запрос включил только записи с blog_id из подзапроса
			query = query.Where("blogs.id IN (?)", subQuery)
		}
	}

	if title != "" && title != "all" {
		query = query.Where("LOWER(title) LIKE ?", "%"+title+"%")
	}
	if money != "" && money != "all" {
		if strings.Contains(money, "-") {
			totalRange := strings.Split(money, "-")
			if len(totalRange) != 2 {
				return fmt.Errorf("invalid total range format: %w", fmt.Errorf("length of totalRange is not 2"))
			}

			lowerTotal, err := strconv.Atoi(strings.TrimSpace(totalRange[0]))
			if err != nil {
				return err
			}

			upperTotal, err := strconv.Atoi(strings.TrimSpace(totalRange[1]))
			if err != nil {
				return err
			}

			query = query.Where("total >= ? AND total <= ?", lowerTotal, upperTotal)
		} else {
			totalInt, err := strconv.Atoi(money)
			if err != nil {
				return err
			}
			query = query.Where("total >= ?", totalInt)
		}
	}

	var count int64
	if err := query.Model(&models.Blog{}).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve data",
		})
	}

	if skip != "" {
		skipInt, err := strconv.Atoi(skip)
		if err != nil {
			return err
		}

		if skipInt >= int(count) {
			skipInt = int(count) - 1
		}

		query = query.Offset(skipInt)
	}

	limit := c.Query("limit", "10")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid limit parameter",
		})
	}

	query = query.Limit(limitInt)

	err = query.Find(&blogs).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not retrieve data",
		})
	}

	var res []*blogResponse
	for _, b := range blogs {
		// b.Views++
		// // Save the changes
		// if err := initializers.DB.Save(&b).Error; err != nil {
		// 	return err
		// }
		hashtags := make([]string, len(b.Hashtags))
		for i, tag := range b.Hashtags {
			hashtags[i] = tag.Hashtag
		}

		userOnlineHours := make(TimeEntryScanner, len(b.User.OnlineHours))
		for i, entry := range b.User.OnlineHours {
			userOnlineHours[i] = TimeEntry{
				Hour:    entry.Hour,
				Minutes: entry.Minutes,
				Seconds: entry.Seconds,
			}
		}

		userTotalOnlineHours := make(TimeEntryScanner, len(b.User.TotalOnlineHours))
		for i, entry := range b.User.TotalOnlineHours {
			userTotalOnlineHours[i] = TimeEntry{
				Hour:    entry.Hour,
				Minutes: entry.Minutes,
				Seconds: entry.Seconds,
			}
		}

		cities := make([]CityJSON, len(b.City))
		for i, city := range b.City {
			cityJSON := CityJSON{
				ID:   city.ID,
				Name: city.Translations[i].Name,
			}
			cities[i] = cityJSON
		}

		categories := make([]CategoryJSON, len(b.Catygory))
		for i, category := range b.Catygory {
			categoryJSON := CategoryJSON{
				ID:   category.ID,
				Name: category.Translations[i].Name,
			}
			categories[i] = categoryJSON
		}

		var telegramNameVal string
		if b.User.TelegramName != nil {
			telegramNameVal = *b.User.TelegramName
		} else {
			telegramNameVal = ""
		}

		blogRes := &blogResponse{
			ID:             b.ID,
			Title:          b.Title,
			MultilangTitle: b.MultilangTitle,
			MultilangDescr: b.MultilangDescr,
			Lang:           b.Lang,
			Descr:          b.Descr,
			Slug:           b.Slug,
			Status:         b.Status,
			Total:          b.Total,
			Content:        b.Content,
			City:           cities,
			UserAvatar:     b.UserAvatar,
			Views:          b.Views,
			Photos:         b.Photos,
			CreatedAt:      b.CreatedAt,
			UpdatedAt:      b.UpdatedAt,
			Pined:          b.Pined,
			Catygory:       categories,
			UniqId:         b.UniqId,
			Sticker:        b.Sticker,
			User: userResponse{
				TId:               b.User.Tid,
				Online:            b.User.Online,
				Photo:             b.User.Photo,
				Name:              b.User.Name,
				TotalBlogs:        b.User.TotalBlogs,
				Role:              b.User.Role,
				OnlineHours:       userOnlineHours,
				TotalOnlineHours:  userTotalOnlineHours,
				TotalRestBlogs:    b.User.TotalRestBlogs,
				TelegramName:      telegramNameVal,
				TelegramActivated: b.User.TelegramActivated,
				IsBot:             b.User.IsBot,
			},
			Hashtags: hashtags,
		}
		res = append(res, blogRes)
	}

	if len(blogs) == 0 {
		return c.JSON(fiber.Map{
			"status": "success",
			"data":   []models.Blog{},
			"meta": fiber.Map{
				"total": count,
				"limit": limitInt,
				"skip":  skip,
			},
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   res,
		"meta": fiber.Map{
			"total": count,
			"limit": limitInt,
			"skip":  skip,
		},
	})
}

func GetRandom(c *fiber.Ctx) error {

	var blogs []models.Blog
	err := initializers.DB.Raw("SELECT * FROM blogs ORDER BY RANDOM() LIMIT 5").Scan(&blogs).Error
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": blogs,
	})
}

func EditBlogGetId(c *fiber.Ctx) error {

	blogID := c.Params("id")
	language := c.Query("language")

	if language == "" {
		language = "en"
	}

	var blog []models.Blog

	err := utils.Paginate(c, initializers.DB.Where("id = ?", blogID).First(&blog).Preload("Hashtags").Preload("Photos").Preload("City.Translations", "language = ?", language).Preload("Catygory.Translations", "language = ?", language).Preload("User"), &blog)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Element not found",
		})
	}

	user := c.Locals("user")
	userResp := user.(models.UserResponse)

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}

	// Assuming there is only one blog in the slice
	if len(blog) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Element not found",
		})
	}

	// Access the first blog in the slice
	blogPost := blog[0]

	// Check if user is the owner of the blog post or has admin rights
	if userObj.Role != "admin" && blogPost.UserID != userObj.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	var res []*blogResponse
	for _, b := range blog {

		hashtags := make([]string, len(b.Hashtags))
		for i, tag := range b.Hashtags {
			hashtags[i] = tag.Hashtag
		}

		userOnlineHours := make(TimeEntryScanner, len(b.User.OnlineHours))
		for i, entry := range b.User.OnlineHours {
			userOnlineHours[i] = TimeEntry{
				Hour:    entry.Hour,
				Minutes: entry.Minutes,
				Seconds: entry.Seconds,
			}
		}

		userTotalOnlineHours := make(TimeEntryScanner, len(b.User.TotalOnlineHours))
		for i, entry := range b.User.TotalOnlineHours {
			userTotalOnlineHours[i] = TimeEntry{
				Hour:    entry.Hour,
				Minutes: entry.Minutes,
				Seconds: entry.Seconds,
			}
		}

		cities := make([]CityJSON, len(b.City))
		for i, city := range b.City {
			cityJSON := CityJSON{
				ID:   city.ID,
				Name: city.Translations[i].Name,
			}
			cities[i] = cityJSON
		}

		categories := make([]CategoryJSON, len(b.Catygory))
		for i, category := range b.Catygory {
			categoryJSON := CategoryJSON{
				ID:   category.ID,
				Name: category.Translations[i].Name,
			}
			categories[i] = categoryJSON
		}

		blogRes := &blogResponse{
			ID:         b.ID,
			Title:      b.Title,
			Descr:      b.Descr,
			Slug:       b.Slug,
			Status:     b.Status,
			Total:      b.Total,
			Content:    b.Content,
			City:       cities,
			UserAvatar: b.UserAvatar,
			Photos:     b.Photos,
			CreatedAt:  b.CreatedAt,
			UpdatedAt:  b.UpdatedAt,
			Catygory:   categories,
			Sticker:    b.Sticker,
			User: userResponse{
				TId:              b.User.Tid,
				Online:           b.User.Online,
				Photo:            b.User.Photo,
				Name:             b.User.Name,
				OnlineHours:      userOnlineHours,
				TotalOnlineHours: userTotalOnlineHours,
				TotalRestBlogs:   b.User.TotalRestBlogs,
				IsBot:            b.User.IsBot,
			},
			Hashtags: hashtags,
		}
		res = append(res, blogRes)
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   res,
	})

}

func GetAllByUser(c *fiber.Ctx) error {

	userId := c.Params("id")
	var blogs []models.Blog
	query := initializers.DB.Where("user_id = ?", userId).Order("pined DESC, created_at DESC").Preload("Photos").Preload("Hashtags")
	query = query.Where("status = ?", "ACTIVE")

	err := utils.Paginate(c, query.Find(&blogs), &blogs)
	if err != nil {
		return err
	}

	return nil

}

func UpdateBlog(c *fiber.Ctx) error {
	blogID := c.Params("id")

	// Retrieve authenticated user from context
	user := c.Locals("user")
	userResp := user.(models.UserResponse)

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}

	var blog models.Blog

	err := initializers.DB.Where("id = ?", blogID).First(&blog).Preload("City").Preload("Catygory").Preload("Hashtags").Preload("Photos").Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Element not found",
		})
	}

	// Check if the user is the owner of the blog post or has admin rights
	if userObj.Role != "admin" && blog.UserID != userObj.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	// Parse the request body
	type RequestBody struct {
		Title string `json:"title"`
		Descr string `json:"descr"`
		City  []struct {
			ID uint64 `json:"id"`
		} `json:"city"`
		Total    float64  `json:"total"`
		Content  string   `json:"content"`
		Pined    bool     `json:"Pined"`
		Hashtags []string `json:"hashtags"`
		Catygory []struct {
			ID uint64 `json:"id"`
		} `json:"Catygory"`
		Photos []struct {
			ID        int64  `json:"ID"`
			BlogID    int64  `json:"BlogID"`
			CreatedAt string `json:"CreatedAt"`
			UpdatedAt string `json:"UpdatedAt"`
			DeletedAt string `json:"DeletedAt"`
			Files     []struct {
				Path string `json:"path"`
			} `json:"files"`
		} `json:"photos"`
	}

	var requestBody RequestBody
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not parse request body",
		})
	}

	if requestBody.Pined {
		// Check if the blog is already pinned by the user
		var pinnedBlog models.Blog
		err = initializers.DB.Where("user_id = ? AND pined = ?", userObj.ID, true).First(&pinnedBlog).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Database error",
				})
			}
		}

		// If a pinned blog already exists, unpin it
		if pinnedBlog.ID != 0 {
			pinnedBlog.Pined = false
			err := initializers.DB.Save(&pinnedBlog).Error
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Database error",
				})
			}

		}

		blog.Pined = requestBody.Pined

		if err := initializers.DB.Save(&blog).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not update blog post",
			})
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   "was pined",
		})

	}

	// Retrieve or create new Hashtags based on the request body
	updatedHashtags := []models.Hashtags{}
	for _, tag := range requestBody.Hashtags {
		hashtag := models.Hashtags{}
		if err := initializers.DB.Where("hashtag = ?", tag).FirstOrCreate(&hashtag, models.Hashtags{Hashtag: tag}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Could not retrieve or create hashtags",
			})
		}
		updatedHashtags = append(updatedHashtags, hashtag)
	}

	// Create a new slice to store the updated list of cities
	updatedCities := []models.City{}

	// Iterate over the requestBody.City and create City objects from the IDs
	for _, cityID := range requestBody.City {
		city := models.City{
			ID: uint(cityID.ID),
		}
		updatedCities = append(updatedCities, city)
	}

	// Create a new slice to store the updated list of catygory
	updatedCatygory := []models.Guilds{}

	// Iterate over the requestBody.catygory and create Catygory objects from the IDs
	for _, catygoryData := range requestBody.Catygory {
		catygory := models.Guilds{
			ID: uint(catygoryData.ID),
		}
		updatedCatygory = append(updatedCatygory, catygory)
	}

	// // Create a new slice to store the updated list of categories
	// updatedCategories := []models.Guilds{}

	// // Iterate over the requestBody.Category and create Category objects from the IDs
	// for _, categoryID := range requestBody.Catygory {
	// 	category := models.Guilds{
	// 		ID: uint(categoryID.ID),
	// 	}
	// 	updatedCategories = append(updatedCategories, category)
	// }

	// Remove the existing hashtags from the blog's association
	if err := initializers.DB.Model(&blog).Association("Hashtags").Clear(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not clear existing hashtags",
		})
	}

	// Remove the existing city from the blog's association
	if err := initializers.DB.Model(&blog).Association("City").Clear(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not clear existing city",
		})
	}

	// Remove the existing city from the blog's association
	if err := initializers.DB.Model(&blog).Association("Catygory").Clear(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not clear existing guilds",
		})
	}

	// // Remove the existing guilds from the blog's association
	// if err := initializers.DB.Model(&blog).Association("Guilds").Clear(); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": "Could not clear existing Guilds",
	// 	})
	// }

	// Assign the updated Hashtags to the Blog instance
	blog.Hashtags = updatedHashtags

	blog.Catygory = updatedCatygory
	blog.Title = requestBody.Title
	blog.Descr = requestBody.Descr
	blog.City = updatedCities
	blog.Total = requestBody.Total
	blog.Pined = requestBody.Pined
	blog.Content = requestBody.Content

	// Fetch languages from the database
	var langs []models.Langs
	err = initializers.DB.Raw("SELECT * FROM langs").Scan(&langs).Error
	if err != nil {
		return err
	}

	translations := make(map[string]string)
	translationsDescr := make(map[string]string)
	translationsContent := make(map[string]string)

	for _, lang := range langs {

		result, _ := gt.Translate(blog.Title, blog.Lang, lang.Code)
		translations[lang.Code] = result

		resultDescr, _ := gt.Translate(blog.Descr, blog.Lang, lang.Code)
		translationsDescr[lang.Code] = resultDescr

		resultContent, _ := gt.Translate(blog.Content, blog.Lang, lang.Code)
		translationsContent[lang.Code] = resultContent
	}

	// Set the translated values in the TitleLangs field
	blog.MultilangTitle.En = translations["en"]
	blog.MultilangTitle.Ru = translations["ru"]
	blog.MultilangTitle.Ka = translations["ka"]
	blog.MultilangTitle.Es = translations["es"]

	// Set the translated values in the TitleLangs field
	blog.MultilangDescr.En = translationsDescr["en"]
	blog.MultilangDescr.Ru = translationsDescr["ru"]
	blog.MultilangDescr.Ka = translationsDescr["ka"]
	blog.MultilangDescr.Es = translationsDescr["es"]

	// Set the translated values in the TitleLangs field
	blog.MultilangContent.En = translationsContent["en"]
	blog.MultilangContent.Ru = translationsContent["ru"]
	blog.MultilangContent.Ka = translationsContent["ka"]
	blog.MultilangContent.Es = translationsContent["es"]

	if err := initializers.DB.Save(&blog).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Could not update blog post",
		})
	}

	// Iterate over the photos in the request body
	for _, photo := range requestBody.Photos {
		// Find the corresponding blog_photos entry by ID
		var blogPhoto models.BlogPhoto
		err := initializers.DB.Where("id = ?", photo.ID).First(&blogPhoto).Error
		if err != nil {
			// Handle the error if the blog_photos entry is not found
			// For example, you can return an error response
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Photo not found",
			})
		}

		// Convert the Files field to a JSONB value
		filesJSON, err := json.Marshal(photo.Files)
		if err != nil {
			// Handle the error if the conversion fails
			// For example, you can return an error response
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Error converting files to JSON",
			})
		}

		var filesJSONB pgtype.JSONB
		if err := filesJSONB.Set(filesJSON); err != nil {
			// Handle the error if the conversion fails
			// For example, you can return an error response
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Error converting files to JSONB",
			})
		}

		// Delete the files that were removed
		deleteRemovedFiles(blogPhoto.Files, filesJSONB)

		// Update the Files field with the JSONB value
		blogPhoto.Files = filesJSONB

		// Save the updated blog_photos entry
		err = initializers.DB.Save(&blogPhoto).Error
		if err != nil {
			// Handle the error if the save operation fails
			// For example, you can return an error response
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to update photo",
			})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Element with ID %s has been updated", blogID),
		"data":    blog,
	})
}

type FileData struct {
	Path string `json:"path"`
}

// Function to delete removed files from the server
func deleteRemovedFiles(existingFiles pgtype.JSONB, newFiles pgtype.JSONB) {
	var existingPaths []string
	var newPaths []string

	// Decode the existing and new files into slices of paths
	var existingData []FileData
	_ = json.Unmarshal(existingFiles.Bytes, &existingData)
	for _, file := range existingData {
		existingPaths = append(existingPaths, file.Path)
	}

	var newData []FileData
	_ = json.Unmarshal(newFiles.Bytes, &newData)
	for _, file := range newData {
		newPaths = append(newPaths, file.Path)
	}

	// Create a map for faster lookup of new paths
	newPathsMap := make(map[string]bool)
	for _, path := range newPaths {
		newPathsMap[path] = true
	}

	// Check if each existing path is present in the new paths
	// If not, delete the file from the server
	for _, path := range existingPaths {
		if !newPathsMap[path] {
			deleteFileFromServer(path)
		}
	}
}

// Function to delete a file from the server
func deleteFileFromServer(path string) {
	config, _ := initializers.LoadConfig(".")

	// Implement the logic to delete the file from the server
	// For example, you can use the os.Remove() function
	absolutePath := filepath.Join(config.IMGStorePath, path)

	_ = os.Remove(absolutePath)
}

// replaceSpecialChars replaces each special character in the input string
func replaceSpecialChars(title string) string {
	// Convert to lowercase
	lowerCaseTitle := strings.ToLower(title)
	// Regex to match non-alphanumeric characters and replace them with a hyphen
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	withHyphens := reg.ReplaceAllString(lowerCaseTitle, "-")
	// Regex to replace multiple hyphens with a single hyphen
	multipleHyphens, _ := regexp.Compile("-+")
	singleHyphen := multipleHyphens.ReplaceAllString(withHyphens, "-")
	// Remove leading and trailing hyphens if present
	trimmedHyphen := strings.Trim(singleHyphen, "-")
	return trimmedHyphen
}
