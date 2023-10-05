package middleware

import (
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func CheckRole(roles []string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
		}
		if user != nil {
			// convert to User type
			userResp := user.(models.UserResponse)
			userObj := models.User{
				Role: userResp.Role,
			}
			for _, role := range roles {
				if userObj.Role == role {
					return c.Next()
				}
			}
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "You are not authorized to access this resource",
			})

		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "You must be logged in to access this resource",
		})
	}
}
