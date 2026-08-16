package model

import "casefile/internal/database"

// Timestamp is a general type alias for a database Timestamp.
type Timestamp = database.Timestamp

// Severity represents the severity level of an Issue.
type Severity string

const (
	SeverityLow      Severity = "Low"
	SeverityMedium   Severity = "Medium"
	SeverityHigh     Severity = "High"
	SeverityCritical Severity = "Critical"
)

// Status represents the status of an Issue.
type Status string

const (
	StatusOpen   Status = "Open"
	StatusClosed Status = "Closed"
)

// Issue represents a detected issue in scanned code.
type Issue struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Severity    Severity  `db:"severity"`
	File        string    `db:"file"`
	Line        int       `db:"line"`
	Status      Status    `db:"status"`
	CreatedAt   Timestamp `db:"created_at"`
	Fingerprint string    `db:"fingerprint"`
}
