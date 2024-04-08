package users

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers incident-related routes with the Fiber app
func RegisterRoutes(app *fiber.App) {
	users := app.Group("/users").Use(middleware.AuthMiddleware())
	users.Post("/", CreateUser)
	users.Get("/:userCode", GetUser)
	users.Get("/", GetUsers)
	users.Put("/:userCode", UpdateUser)
	users.Delete("/:userCode", DeleteUser)
	users.Put("/role/:userCode", AssignRole)
}
