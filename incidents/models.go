package incidents

import (
	"issue-reporting/users"
	"time"
)

type Incident struct {
	Id             string       `json:"id"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Severity       Severity     `json:"severity"`
	Status         Status       `json:"status"`
	AssignedTo     []users.User `json:"assigned_to"`
	CreatedAt      time.Time    `json:"created_at"`
	TeamId         string       `json:"teamId"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Resolved       bool         `json:"resolved"`
	ResolvedAt     time.Time    `json:"resolved_at"`
	Acknowledged   bool         `json:"acknowledged"`
	AcknowledgedAt time.Time    `json:"acknowledged_at"`
	Actions        []string     `json:"actions"`
	FollowUps      []string     `json:"followUps"`
	Timeline       []Timepoint  `json:"timeline"`
	Metadata       string       `json:"metadata"`
	ReportCreated  bool         `json:"reportCreated"`
}

type Incidents struct {
	Incidents []Incident `json:"incidents"`
}

type Timepoint struct {
	Title     string
	CreatedAt time.Time
	Metadata  string
}

type Severity string

const (
	SeverityLow    Severity = "Low"
	SeverityMedium Severity = "Medium"
	SeverityHigh   Severity = "High"
)

type Status string

const (
	StatusOpen   Status = "Open"
	StatusClosed Status = "Closed"
)

type AssignParams struct {
	User users.User
}
