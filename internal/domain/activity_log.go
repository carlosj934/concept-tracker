package domain

import (
	"time"
)

type ActivityLog struct {
	ID           string
	ConceptID    string
	UserID       string
	ActivityType string
	DurationMins *int64
	Notes        *string
	LoggedAt     time.Time
	CreatedAt    time.Time
}

type Cursor struct {
	LoggedAt time.Time
	ID       string
}
