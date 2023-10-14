package main

import (
	"encoding/json"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"
	"io/ioutil"
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

	initializers.DB.AutoMigrate(&models.User{})
	initializers.DB.AutoMigrate(&models.Domain{})
	initializers.DB.AutoMigrate(&models.Billing{})
	initializers.DB.AutoMigrate(&models.Transaction{})
	initializers.DB.AutoMigrate(&models.Blog{})
	initializers.DB.AutoMigrate(&models.BlogPhoto{})
	initializers.DB.AutoMigrate(&models.City{})
	initializers.DB.AutoMigrate(&models.Payments{})
	initializers.DB.AutoMigrate(&models.Guilds{})
	initializers.DB.AutoMigrate(&models.GuildTranslation{})
	initializers.DB.AutoMigrate(&models.Profile{})
	initializers.DB.AutoMigrate(&models.OnlineStorage{})
	initializers.DB.AutoMigrate(&models.Codes{})
	initializers.DB.AutoMigrate(&models.Hashtags{})
	initializers.DB.AutoMigrate(&models.Hashtagsprofile{})
	initializers.DB.AutoMigrate(&models.ProfilePhoto{})
	initializers.DB.AutoMigrate(&models.ProfileDocuments{})
	initializers.DB.AutoMigrate(&models.ProfileService{})

	// Check if there are any users in the database
	var userCount int64
	initializers.DB.Model(&models.User{}).Count(&userCount)

	if userCount == 0 {
		// If there are no users, create a new user with role admin
		password := "techaa123"
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
		if err := os.MkdirAll(filepath.Join("../images", dirName), 0755); err != nil {
			log.Fatal("Could not create directory:", err)
		}

		src := filepath.Join("..", "..", "images", "default.jpg")
		dst := filepath.Join("../images", dirName, "default.jpg")
		if err := os.Symlink(src, dst); err != nil {
			log.Fatal("Could not create symlink:", err)
		}

		type DomainSettings struct {
			Logo      string `json:"logo"`
			Title     string `json:"title"`
			MetaDescr string `json:"metadescr"`
			// ... other settings fields ...
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
			// ... other settings ...
		}
		if err := settings.Set(settingsMap); err != nil {
			// Handle error if needed
		}

		domain := models.Domain{
			UserID:   admin.ID,
			Username: "andysay",
			Name:     "ddrw.ru",
			Settings: settings,
		}

		// Create the Domain record associated with the admin user
		if err := initializers.DB.Create(&domain).Error; err != nil {
			fmt.Println("Error creating domain:", err)
			return
		}

		fmt.Println("✅ Admin user created")

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

		// Get the current month
		currentMonth := time.Now().Month().String()

		// Create online storage for the admin user
		onlineStorage := models.OnlineStorage{
			UserID: admin.ID,
			Year:   time.Now().Year(),
			Data: []byte(`
				[
					{
						"Month": "` + currentMonth + `",
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

		fmt.Println("✅ Billing record and online storage created")

		file, err := ioutil.ReadFile("migrate/city.json")
		if err != nil {
			log.Fatal("Could not read city.json file:", err)
		}

		var cities []models.City
		err = json.Unmarshal(file, &cities)
		if err != nil {
			log.Fatal("Could not parse city.json file:", err)
		}

		for _, city := range cities {
			err := initializers.DB.Create(&city).Error
			if err != nil {
				log.Fatalf("Could not create city %s: %s", city.Name, err)
			}
		}

		fmt.Println("✅ Cities record created")

	}

	// Считать данные из guilds.json
	fileGuilds, err := os.ReadFile("migrate/guilds.json")
	if err != nil {
		log.Fatal("Could not read guilds.json file:", err)
	}

	var guildsData []struct {
		ID  uint   `json:"id"`
		Hex string `json:"hex"`
	}

	err = json.Unmarshal(fileGuilds, &guildsData)
	if err != nil {
		log.Fatal("Could not parse guilds.json file:", err)
	}

	// Создать записи в таблице Guilds
	for _, guildData := range guildsData {
		guild := models.Guilds{
			ID:        guildData.ID,
			Hex:       guildData.Hex,
			UpdatedAt: time.Now(), // Укажите соответствующее время
		}
		err := initializers.DB.Create(&guild).Error
		if err != nil {
			log.Fatalf("Could not create guild with ID %d: %s", guildData.ID, err)
		}
	}

	// Считать данные из guilds_trans.json
	fileTranslations, err := os.ReadFile("migrate/guilds_tr.json")
	if err != nil {
		log.Fatal("Could not read guilds_trans.json file:", err)
	}

	var translationsData []struct {
		ID       uint   `json:"id"`
		GuildID  uint   `json:"GuildID"`
		Language string `json:"Language"`
		Name     string `json:"Name"`
	}

	err = json.Unmarshal(fileTranslations, &translationsData)
	if err != nil {
		log.Fatal("Could not parse guilds_trans.json file:", err)
	}

	// Создать записи в таблице GuildTranslation
	for _, translationData := range translationsData {
		translation := models.GuildTranslation{
			ID:       translationData.ID,
			GuildID:  translationData.GuildID,
			Language: translationData.Language,
			Name:     translationData.Name,
		}
		err := initializers.DB.Create(&translation).Error
		if err != nil {
			log.Fatalf("Could not create translation with ID %d: %s", translationData.ID, err)
		}
	}

	fmt.Println("✅ Guilds and translations saved successfully")

	fmt.Println("✅ Guilds saved successfully")

	fmt.Println("✅ Migration complete")
}
