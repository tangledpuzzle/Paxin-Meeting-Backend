package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetAllVotes(c *fiber.Ctx) error {
	//Get blog's id from request
	blogId := c.Params("id")
	var blog models.Blog

	if err := initializers.DB.Where("blog_id = ?", blogId).First(&blog).Error; err != nil {
		// Handle not found error, probably return a 404
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Blog not found",
			"error":   err.Error(),
		})
	}

	var votes []models.Vote

	if err := initializers.DB.Where("blog_id = ?", blog.ID).Find(&votes).Error; err != nil {
		// Handle not found error, probably return a 404
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Votes not found",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Success retrieving votes",
		"votes":   votes,
	})
}

func AddVote(c *fiber.Ctx) error {
	// Get the IsUP from the request body
	type RequestAddVote struct {
		IsUP bool `json:"isUP"`
	}
	reqBody := new(RequestAddVote)
	if err := c.BodyParser(reqBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	//Get blog's id from request
	blogId := c.Params("id")

	var blog models.Blog

	if err := initializers.DB.Where("blog_id = ?", blogId).First(&blog).Error; err != nil {
		// Handle not found error, probably return a 404
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Blog not found",
			"error":   err.Error(),
		})
	}

	user := c.Locals("user").(models.UserResponse)

	userID := user.ID

	// Lookup any existing vote
	var vote models.Vote
	if err := initializers.DB.Where(&models.Vote{BlogID: blog.ID, UserID: userID}).First(&vote).Error; err != nil {
		// If no vote exists, create a new one.
		if err == gorm.ErrRecordNotFound {
			vote.IsUP = reqBody.IsUP
			vote.UserID = userID
			vote.BlogID = blog.ID
			if err := initializers.DB.Create(&vote).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Could not create vote",
					"error":   err.Error(),
				})
			}
		} else {
			// An error other than ErrRecordNotFound occurred
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Vote lookup failed",
				"error":   err.Error(),
			})
		}
	} else {
		// Vote exists. If IsUp is same, delete vote, else update vote.
		if vote.IsUP == reqBody.IsUP {
			// Delete vote
			if err := initializers.DB.Delete(&vote).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Could not delete vote",
					"error":   err.Error(),
				})
			}
		} else {
			// Update vote
			vote.IsUP = reqBody.IsUP
			if err := initializers.DB.Save(&vote).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Could not update vote",
					"error":   err.Error(),
				})
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Success adding vote",
	})
}
