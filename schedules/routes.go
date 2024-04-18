package schedules

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	schedule := app.Group("/schedules").Use(middleware.AuthMiddleware())
	schedule.Get("/", GetAllSchedules)
	schedule.Get("/now", GetScheduledNow)
	schedule.Get("/time/:timestamp", GetScheduledAt)
	schedule.Post("/range", ListByTimeRange)
	schedule.Delete("/:id", DeleteSchedule)
	schedule.Put("/:id", UpdateSchedules)
	schedule.Post("/:userCode", CreateSchedules)
}
