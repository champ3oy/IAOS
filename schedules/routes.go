package schedules

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	users := app.Group("/schedules").Use(middleware.AuthMiddleware())
	users.Get("/now", GetScheduledNow)
	users.Get("/time/:timestamp", GetScheduledAt)
	users.Post("/range", ListByTimeRange)
	users.Delete("/:id", DeleteSchedule)
	users.Put("/:id", UpdateSchedules)
	users.Post("/:userCode", CreateSchedules)
}
