package middleware

// import (
// 	"fmt"
// 	"hyperpage/models"

// 	"github.com/gofiber/fiber/v2"
// )

// type PermissionMiddleware struct {
// 	Action     string
// 	Resource   string
// 	Permission string
// }

// func CheckPermission(permission PermissionMiddleware) func(c *fiber.Ctx) error {
// 	return func(c *fiber.Ctx) error {
// 		user := c.Locals("user")
// 		if user == nil {
// 			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
// 		}
// 		if user != nil {
//             userResp := user.(models.UserResponse)
//             userObj := models.User{
//                 Role:     userResp.Role,
//             }
// 			// Extract user's role
// 			role := userObj.Role

// 			// Check if the role has the required permission
// 			if hasPermission(role, permission.Action, permission.Resource, permission.Permission) {
// 				return c.Next()
// 			}
// 		}

// 		// If user does not have the required permission, return an error message or redirect to an appropriate page
// 		return c.SendStatus(fiber.StatusForbidden)
// 	}
// }

// func hasPermission(role models.Role, action string, resource string, permission string) bool {
// 	// Check if role has the required permission for the given action and resource
// 	// You can implement your own logic here, such as using a database or a map to store permissions
// 	return true
// }