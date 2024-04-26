package users

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	users := app.Group("/users").Use(middleware.AuthMiddleware())
	users.Get("/user", GetCurrentUser)
	users.Get("/team", GetTeam)
	users.Get("/:userCode", GetUser)
	users.Get("/", GetUsers)
	users.Put("/", UpdateUser)
	users.Delete("/:userCode", DeleteUser)
	users.Put("/role/:userCode", AssignRole)
}
