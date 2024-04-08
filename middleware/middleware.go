package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware is a middleware function to authenticate users
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Missing authentication token"})
		}
		claims, err := VerifyToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid authentication token"})
		}
		c.Locals("email", claims.Email)
		return c.Next()
	}
}
