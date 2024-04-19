package reports

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	reports := app.Group("/reports").Use(middleware.AuthMiddleware())
	reports.Get("/", GetReports)
}
