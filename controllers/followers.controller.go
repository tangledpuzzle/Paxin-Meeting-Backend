package controllers

import (
	"errors"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"log"

	uuid "github.com/satori/go.uuid"

	"github.com/gofiber/fiber/v2"
)


func Scribe(c *fiber.Ctx) error {


	// Parse request body into a new BlogSearch object
	IDUSER := c.Locals("user")
	if IDUSER == nil {
		// Handle the case when user is nil
		return errors.New("user not found")
	}

	userResp, ok := IDUSER.(models.UserResponse)
	if !ok {
		// Handle the case when user is not of type models.UserResponse
		return errors.New("invalid user type")
	}


	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}
	
	// Parse the request body
	type makelink struct {
		UserID     uuid.UUID
		FollowerID uuid.UUID
	}

	var requestBody makelink
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request body",
		})
	}

	// Check if the userObj.ID matches the requested UserID
	if userObj.ID != requestBody.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Unauthorized to perform this action",
		})
	}


	// Check if the UserID is the same as the logged-in user's ID
	if userObj.ID == requestBody.FollowerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "You cannot scribe yourself",
		})
	}


	var follower, user models.User

	// Fetch the follower and user based on the provided IDs
	if err := initializers.DB.First(&follower, requestBody.FollowerID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Follower not found",
		})
	}

	if err := initializers.DB.First(&user, requestBody.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}


	follower.TotalFollowers += 1

	if err := initializers.DB.Save(&follower).Error; err != nil {
		log.Println("Could not update user follower count:", err)
	}


    // Update the relationship in the database
    initializers.DB.Model(&user).Association("Followers").Append(&follower)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "was add",
	})	
}


func Unscribe(c *fiber.Ctx) error {

	// Parse request body into a new BlogSearch object
	IDUSER := c.Locals("user")
	if IDUSER == nil {
		// Handle the case when user is nil
		return errors.New("user not found")
	}

	userResp, ok := IDUSER.(models.UserResponse)
	if !ok {
		// Handle the case when user is not of type models.UserResponse
		return errors.New("invalid user type")
	}

	userObj := models.User{
		ID:   userResp.ID,
		Role: userResp.Role,
	}
	

	// Parse the request body
	type makeUnlink struct {
		UserID     uuid.UUID
		FollowerID uuid.UUID
	}
	

	var requestBody makeUnlink

	fmt.Println(requestBody)
	
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request body",
		})
	}

	// Check if the userObj.ID matches the requested UserID
	if userObj.ID != requestBody.UserID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Unauthorized to perform this action",
		})
	}

	// Check if the UserID is the same as the logged-in user's ID
	if userObj.ID == requestBody.FollowerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "You cannot unsubscribe yourself",
		})
	}



	var follower, user models.User

	// Fetch the follower and user based on the provided IDs
	if err := initializers.DB.First(&follower, requestBody.FollowerID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Follower not found",
		})
	}

	if err := initializers.DB.First(&user, requestBody.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	initializers.DB.Model(&user).Association("Followers").Delete(&follower)

	follower.TotalFollowers -= 1

	if follower.TotalFollowers < 0 {
		follower.TotalFollowers = 0
	}
	
	if err := initializers.DB.Save(&follower).Error; err != nil {
		log.Println("Could not update user follower count:", err)
	}


	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "was removed",
	})	
	
}

func GetFollowers(c *fiber.Ctx) error {
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

	var usF models.User
    if err := initializers.DB.Preload("Followers").First(&user, "id = ?", userObj.ID).Error; err != nil {
        return nil
    }
	

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   usF.Followers,
	})
}