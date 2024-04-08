package incidents

import (
	"issue-reporting/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers incident-related routes with the Fiber app
func RegisterRoutes(app *fiber.App) {
	incidents := app.Group("/incidents").Use(middleware.AuthMiddleware())
	incidents.Post("/", CreateIncident)
	incidents.Get("/", GetIncidents)
	incidents.Get("/:id", GetIncident)
	incidents.Put("/:id", UpdateIncident)
	incidents.Delete("/:id", DeleteIncident)
	incidents.Get("/assign/:userId/:incidentId", AssignUser)
	incidents.Get("/acknowledges", AcknowledgeAll)
	incidents.Get("/acknowledge/:incidentId", Acknowledge)
}
