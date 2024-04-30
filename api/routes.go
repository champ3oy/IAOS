package api

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1").Use(middleware.VerifyAPI())
	api.Post("/incident", CreateIncident)
	api.Post("/log", CreateLog)
}
