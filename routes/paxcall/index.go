package paxcall

import (
	"encoding/base64"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {

	app.Get("/", func(c *fiber.Ctx) error {
		//Set session in cookie
		id := uuid.NewV4()
		idSession := base64.URLEncoding.EncodeToString(id[:])

		c.Cookie(&fiber.Cookie{
			Name:     "session",
			Value:    idSession,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: false,
			SameSite: "lax",
		})

		return c.Render("index", fiber.Map{
			"Title":       "Powerful paxintrade server",
			"Description": "server developed by paxintrade",
		})
	})

}
