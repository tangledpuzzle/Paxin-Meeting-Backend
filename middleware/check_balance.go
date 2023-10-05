package middleware

// import (
// 	"hyperpage/models"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// func CheckBalance(amount int) func(c *fiber.Ctx) error {
//     return func(c *fiber.Ctx) error {
//         user := c.Locals("user")
//         if user == nil {
//             return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
//         }
//         // convert to User type
//         userResp := user.(models.UserResponse)
//         userID := userResp.ID
//         var balance int64
//         err := models.DB.Model(&models.Billing{}).Where("user_id = ?", userID).Select("balance").Scan(&balance).Error
//         if err != nil {
//             if err == gorm.ErrRecordNotFound {
//                 return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
//                     "message": "Billing record not found",
//                 })
//             }
//             return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
//                 "message": "Internal Server Error",
//             })
//         }
//         if balance >= int64(amount) {
//             return c.Next()
//         }
//         return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
//             "message": "Insufficient balance",
//         })
//     }
// }