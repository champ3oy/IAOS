package reports

import (
	"context"
	"fmt"
	"issue-reporting/auth"
	"issue-reporting/database"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetReports(c *fiber.Ctx) error {
	ctx := context.Background()
	email := c.Locals("email").(string)
	var team auth.Teams
	err := database.FindOne("teams", bson.M{"email": email}).Decode(&team)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	filter := bson.M{"incident.teamid": team.TeamId}
	cursor, err := database.Find("reports", filter)
	if err != nil {
		return fmt.Errorf("error finding reports: %v", err)
	}
	defer cursor.Close(ctx)

	var reports []Report
	for cursor.Next(ctx) {
		var report Report
		if err := cursor.Decode(&report); err != nil {
			return fmt.Errorf("error decoding report: %v", err)
		}
		reports = append(reports, report)
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "reports data",
		"reports": &reports,
	})
}
