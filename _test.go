package routes

import (
	"github.com/gofiber/fiber/v2"
)

func MainView(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":       "Powerful Paxintrade API Server",
			"Description": "Server is developed by paxintrade",
		})
	})
}


import (
	"github.com/gofiber/fiber/v2"
)

func MainView(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":       "Powerful Paxintrade API Server",
			"Description": "Server is developed by paxintrade",
		})
	})
}
