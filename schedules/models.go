package schedules

import (
	"issue-reporting/users"
	"time"
)

type Schedules struct {
	Items []Schedule
}

type Schedule struct {
	User users.User
	Time TimeRange
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}
