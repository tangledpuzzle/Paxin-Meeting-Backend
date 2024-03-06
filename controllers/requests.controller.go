package controllers

import (
	"hyperpage/models"
	"hyperpage/utils"

	"github.com/gofiber/fiber/v2"
)

func Userq(c *fiber.Ctx) error {

	mode := c.Query("mode")
	if mode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Mode is required",
		})
	}

	var emailData interface{}
	switch mode {
	case "ReqCat":
		var requestBody struct {
			Name  string `json:"name"`
			Descr string `json:"descr"`
		}
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse JSON body",
			})
		}
		emailData = &utils.ReqCat{
			Name:    requestBody.Name,
			Descr:   requestBody.Descr,
			Subject: "New inquiry: Cat",
		}
	case "ReqCity":
		var requestBody struct {
			Name  string `json:"name"`
			Descr string `json:"descr"`
		}
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse JSON body",
			})
		}
		emailData = &utils.ReqCity{
			Subject: "New inquiry: City",
			Name:    requestBody.Name,
			Descr:   requestBody.Descr,
		}
	case "ComplaintUser":
		var requestBody struct {
			Name  string `json:"name"`
			Descr string `json:"descr"`
			Type  string `json:"type"`
		}
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse JSON body",
			})
		}
		emailData = &utils.ComplainUser{
			Subject: "Complaint on the user",
			Name:    requestBody.Name,
			Descr:   requestBody.Descr,
			Type:    requestBody.Type,
		}
	case "ComplaintPost":
		var requestBody struct {
			Name  string `json:"name"`
			Descr string `json:"descr"`
			Type  string `json:"type"`
		}
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse JSON body",
			})
		}
		emailData = &utils.ComplainPost{
			Subject: "Complaint on the post",
			Name:    requestBody.Name,
			Descr:   requestBody.Descr,
			Type:    requestBody.Type,
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid mode specified",
		})
	}

	utils.SendEmail(&models.User{Email: "qa@paxintrade.com"}, emailData, mode, "en")

	return c.SendStatus(fiber.StatusOK)
}
