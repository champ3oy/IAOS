package schedules

import (
	"issue-reporting/auth"
	"time"
)

type Schedules struct {
	Items []Schedule
}

type Schedule struct {
	User auth.User
	Time TimeRange
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}
