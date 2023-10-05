package middleware

import (
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func CheckProfileFilled() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "You must be logged in to access this resource"})
		}

		// Convert to User type
		userResp := user.(models.UserResponse)
		userObj := models.User{
			Role:   userResp.Role,
			Filled: userResp.Filled, // Assuming 'IsProfileFilled' is a boolean field indicating whether the profile is filled or not.
		}

		if !userObj.Filled {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "You must fill your profile"})
		}

		return c.Next()
	}
}
