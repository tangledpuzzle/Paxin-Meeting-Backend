package main

import (
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgtype"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("? Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)
}

func main() {
	config, _ := initializers.LoadConfig(".")

	if err := initializers.DB.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Domain{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Billing{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Transaction{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Blog{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.BlogPhoto{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.City{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.CityTranslation{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Payments{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Guilds{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.GuildTranslation{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Profile{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.OnlineStorage{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Codes{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Hashtags{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.HashtagsForProfile{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ProfilePhoto{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ProfileDocuments{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ProfileService{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Langs{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.DevicesIOS{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Vote{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ChatMessage{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ChatRoom{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ChatRoomMember{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ChatCDC{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.ChatOutbox{}); err != nil {
		panic(err)
	}
	if err := initializers.DB.AutoMigrate(&models.Presavedfilters{}); err != nil {
		panic(err)
	}

	if err := initializers.DB.AutoMigrate(&models.Streaming{}); err != nil {
		panic(err)
	}

	// Check if there are any users in the database
	var userCount int64
	initializers.DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		// If there are no users, create a new user with role admin
		password := "1234567890"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("? Could not hash password", err)
		}

		rand.Seed(time.Now().UnixNano())
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		const length = 10
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}
		randomString := string(b)

		// Generate a unique directory name
		dirName := utils.GenerateUniqueDirName()

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Join(config.IMGStorePath, dirName), 0755); err != nil {
			log.Fatal("Could not create directory:", err)
		}

		src := filepath.Join(config.IMGStorePath, "default.jpg")
		dst := filepath.Join(config.IMGStorePath, dirName, "default.jpg")
		if err := os.Symlink(src, dst); err != nil {
			log.Fatal("Could not create symlink:", err)
		}

		admin := models.User{
			Name:          "andysay",
			Email:         "info@ddrw.ru",
			Password:      string(hashedPassword),
			Role:          "admin",
			Verified:      true,
			Storage:       dirName,
			TelegramToken: randomString,
		}
		initializers.DB.Create(&admin)

		settings := pgtype.JSONB{}
		settingsMap := map[string]interface{}{
			"logo":      "logo.png",
			"title":     "My Website",
			"metadescr": "Welcome to my website!",
		}
		if err := settings.Set(settingsMap); err != nil {
			log.Fatal("Error creating settings:", err)
		}

		// Create the Domain record associated with the admin user
		domain := models.Domain{
			UserID:   admin.ID,
			Username: "andysay",
			Name:     "ddrw.ru",
			Settings: settings,
		}
		initializers.DB.Create(&domain)

		// Create a billing record for the first user
		billing := models.Billing{
			UserID: admin.ID,
			Amount: 5000,
		}
		initializers.DB.Create(&billing)

		// Create a profile record for the first user
		profile := models.Profile{
			UserID:    admin.ID,
			Firstname: "Андрей",
			// Lastname:  "Леонов",
			// MiddleN:   "Владимирович",
		}
		initializers.DB.Create(&profile)

		// Create online storage for the admin user
		onlineStorage := models.OnlineStorage{
			UserID: admin.ID,
			Year:   time.Now().Year(),
			Data: []byte(`
				[
					{
						"Month": "` + time.Now().Month().String() + `",
						"Hours": [
							{
								"Hour": 0,
								"Minutes": 0,
								"Seconds": 0
							}
						]
					}
				]
			`),
		}
		initializers.DB.Create(&onlineStorage)

		code := models.Codes{
			Code:      "paxintrade",
			Balance:   "100",
			UserId:    admin.ID,
			Activated: false,
		}
		initializers.DB.Create(&code)

		fmt.Println("✅ Admin user created")
	}
	fmt.Println("✅ Migration complete")
}
