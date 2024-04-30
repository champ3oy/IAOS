package auth

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(app *fiber.App) {
	authRoutes := app.Group("/auth")
	authRoutes.Post("/register", Register)
	authRoutes.Post("/join", JoinTeam)
	authRoutes.Post("/login", Login)
	authRoutes.Post("/refresh", Refresh)

	teamRoutes := app.Group("/team").Use(middleware.AuthMiddleware())
	teamRoutes.Put("/", UpdateTeam)

}
