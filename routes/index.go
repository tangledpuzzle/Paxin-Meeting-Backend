package routes

import (
	"github.com/gofiber/fiber/v2"
)

func MainView(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":       "Powerful paxintrade server",
			"Description": "server developed by paxintrade",
		})
	})

}
