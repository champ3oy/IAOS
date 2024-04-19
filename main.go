package main

import (
	"issue-reporting/auth"
	"issue-reporting/cron"
	"issue-reporting/database"
	"issue-reporting/incidents"
	"issue-reporting/schedules"
	"issue-reporting/users"
	"log"
	"os"

	_ "issue-reporting/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	cron.StartNotifyAssignScheduler()
	cron.ReportGeneratorScheduler()
	// cron.StartNotifyAcknowlegedScheduler()

	port := os.Getenv("PORT")
	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New(logger.Config{
		Format:     "${cyan}[${time}] ${red}[${ip}] ${magenta}${bytesSent}bytes ${green}${latency} ${blue}${method} ${white}${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "UTC",
	}))

	auth.RegisterAuthRoutes(app)
	incidents.RegisterRoutes(app)
	users.RegisterRoutes(app)
	schedules.RegisterRoutes(app)

	app.Listen(":" + port)
}
