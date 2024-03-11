package controllers

import (
	"hyperpage/initializers"
	"hyperpage/models"

	"github.com/gofiber/fiber/v2"
)

func CreatePresavedfilter(c *fiber.Ctx) error {
	// Получаем аутентифицированного пользователя из контекста (предположим, что он находится в локальных данных)
	user := c.Locals("user").(models.UserResponse)

	// Парсим JSON тело запроса в структуру Presavedfilters
	filter := new(models.Presavedfilters)
	if err := c.BodyParser(filter); err != nil {
		return err
	}

	// Устанавливаем UserID для фильтра
	filter.UserID = user.ID

	// Сохраняем фильтр в базе данных
	result := initializers.DB.Create(filter)
	if result.Error != nil {
		// В случае ошибки сохранения возвращаем ошибку 500
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create presaved filter",
		})
	}

	// Возвращаем успешный ответ
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Presaved filter created successfully",
		"data":    filter,
	})
}
