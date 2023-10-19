package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"

	"gorm.io/gorm"
)

func DeserializeUser(c *fiber.Ctx) error {
	language, ok := c.Locals("language").(string)
	if !ok {
		// Handle the case when the language is not set.
		language = "en" // You can set a default language here.
	}

	var access_token string
	authorization := c.Get("Authorization")

	if strings.HasPrefix(authorization, "Bearer ") {
		access_token = strings.TrimPrefix(authorization, "Bearer ")
	} else if c.Cookies("access_token") != "" {
		access_token = c.Cookies("access_token")
	}

	if access_token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "You are not logged in"})
	}

	config, _ := initializers.LoadConfig(".")

	tokenClaims, err := utils.ValidateToken(access_token, config.AccessTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	var user models.User
	err = initializers.DB.Preload("Followings").
		Preload("Followers").
		Preload("Profile.City.Translations", "language = ?", language).
		Preload("Profile.Guilds.Translations", "language = ?", language).
		Preload("Profile.Hashtags").
		Preload("Profile.Photos").
		First(&user, "id = ?", tokenClaims.UserID).Error

	if err == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no longer exists"})
	}

	c.Locals("user", models.FilterUserRecord(&user, language))
	c.Locals("access_token_uuid", tokenClaims.TokenUuid)

	return c.Next()
}
