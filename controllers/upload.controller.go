package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"image"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/gofiber/fiber/v2"
)

func UploadPdf(c *fiber.Ctx) error {
	config, _ := initializers.LoadConfig(".")

	file, err := c.FormFile("pdf")
	if err != nil {
		return err
	}

	// Validate file type
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	if fileExt != ".pdf" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Only pdf file type are allowed.",
		})
	}

	// Validate file size
	if file.Size > 10*1024*1024 { // 2 MB
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File size exceeds the limit of 2 MB.",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read the file contents into a byte slice
	fileContents, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	// Create a new file in the uploads directory
	hash := sha256.Sum256(fileContents)
	hashStr := hex.EncodeToString(hash[:])

	userId := c.Locals("user")
	userResp := userId.(models.UserResponse)
	userObj := models.User{
		ID:                userResp.ID,
		Name:              userResp.Name,
		Email:             userResp.Email,
		Role:              userResp.Role,
		TelegramToken:     userResp.TelegramToken,
		TelegramActivated: userResp.TelegramActivated,
		Photo:             userResp.Photo,
		Session:           userResp.Session,
		Storage:           userResp.Storage,
	}

	dst, err := os.Create(fmt.Sprintf(config.IMGStorePath, userObj.Storage+"/%s%s", hashStr, fileExt))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Write the file contents to the new file
	_, err = dst.Write(fileContents)
	if err != nil {
		return err
	}

	// Return JSON response with the uploaded file's name
	return c.JSON(fiber.Map{
		"filename": hashStr + fileExt,
	})
}

func UploadImage(c *fiber.Ctx) error {
	config, _ := initializers.LoadConfig(".")

	file, err := c.FormFile("image")
	if err != nil {
		return err
	}

	// Validate file type
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Only PNG, JPEG, and JPG file types are allowed.",
		})
	}

	// Validate file size
	if file.Size > 10*1024*1024 { // 2 MB
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File size exceeds the limit of 2 MB.",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read the file contents into a byte slice
	fileContents, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	// Create a new file in the uploads directory
	hash := sha256.Sum256(fileContents)
	hashStr := hex.EncodeToString(hash[:])

	userId := c.Locals("user")
	userResp := userId.(models.UserResponse)
	userObj := models.User{
		ID:                userResp.ID,
		Name:              userResp.Name,
		Email:             userResp.Email,
		Role:              userResp.Role,
		TelegramToken:     userResp.TelegramToken,
		TelegramActivated: userResp.TelegramActivated,
		Photo:             userResp.Photo,
		Session:           userResp.Session,
		Storage:           userResp.Storage,
	}

	dst, err := os.Create(fmt.Sprintf(config.IMGStorePath, userObj.Storage+"/%s%s", hashStr, fileExt))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Write the file contents to the new file
	_, err = dst.Write(fileContents)
	if err != nil {
		return err
	}

	// Return JSON response with the uploaded file's name
	return c.JSON(fiber.Map{
		"filename": hashStr + fileExt,
	})
}

type UploadedFile struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

// Function to calculate directory size
func getDirectorySize(dirname string) (int64, error) {
	var size int64

	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

func compressImage(inputPath, outputPath string, maxWidth, maxHeight int) error {
	// Open the input image file
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Resize the image to the desired dimensions
	resized := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	// Save the compressed image
	err = imaging.Save(resized, outputPath)
	if err != nil {
		return err
	}

	return nil
}

func UploadImages(c *fiber.Ctx) error {
	config, _ := initializers.LoadConfig(".")

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	// Loop through files
	var filesData []map[string]string
	files := form.File["image"]
	for _, file := range files {

		// Validate file type
		fileExt := strings.ToLower(filepath.Ext(file.Filename))
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Only PNG, JPEG, and JPG file types are allowed.",
			})
		}

		// Validate file size
		if file.Size > 20*1024*1024 { // 20 MB
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "File size exceeds the limit of 2 MB.",
			})
		}

		// Open the file
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Read the file contents into a byte slice
		fileContents, err := io.ReadAll(src)
		if err != nil {
			return err
		}

		// Create a new file in the uploads directory
		hash := sha256.Sum256(fileContents)
		hashStr := hex.EncodeToString(hash[:])
		name := file.Filename
		// Generate a random number
		randomNum := rand.Intn(1000) // Adjust the maximum range as needed

		// Concatenate the hash string and random number
		filename := hashStr + "_" + strconv.Itoa(randomNum) + fileExt

		userId := c.Locals("user")
		userResp := userId.(models.UserResponse)
		userObj := models.User{
			ID:                userResp.ID,
			Name:              userResp.Name,
			Email:             userResp.Email,
			Role:              userResp.Role,
			TelegramToken:     userResp.TelegramToken,
			TelegramActivated: userResp.TelegramActivated,
			Photo:             userResp.Photo,
			Session:           userResp.Session,
			Storage:           userResp.Storage,
			LimitStorage:      userResp.LimitStorage,
		}

		// Check the size of the directory
		dirname := filepath.Join(config.IMGStorePath, userObj.Storage)
		size, err := getDirectorySize(dirname)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to calculate directory size",
			})
		}

		// Convert the limit and size to megabytes
		limitStorageMB := int64(userObj.LimitStorage) // Convert limit to bytes and cast to int
		sizeMB := size / (1024 * 1024)                // Convert size to megabytes

		// Check if the size exceeds the limit
		if sizeMB > limitStorageMB {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Directory size exceeds the storage limit",
			})
		}

		dst, err := os.Create(fmt.Sprintf(config.IMGStorePath, userObj.Storage+"/%s%s", filename, fileExt))
		if err != nil {
			return err
		}
		defer dst.Close()

		// Write the file contents to the new file
		_, err = dst.Write(fileContents)
		if err != nil {
			return err
		}

		inputPath := filepath.Join(config.IMGStorePath, userObj.Storage, filename+fileExt)
		outputPath := filepath.Join(config.IMGStorePath, userObj.Storage, filename+fileExt)
		maxWidth := 800  // Replace with the desired maximum width of the compressed photo
		maxHeight := 600 // Replace with the desired maximum height of the compressed photo

		err = compressImage(inputPath, outputPath, maxWidth, maxHeight)
		if err != nil {
			// Handle the error
			panic(err)
		}

		// Append the uploaded file's data to the filesData slice
		filesData = append(filesData, map[string]string{
			"name": name,
			"path": fmt.Sprintf(userObj.Storage+"/%s%s", filename, fileExt),
		})
	}

	// Return JSON response with the uploaded file data
	return c.JSON(fiber.Map{
		"files": filesData,
	})
}
