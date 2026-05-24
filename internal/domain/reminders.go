package domain

import (
	"time"
)

type Reminder struct {
	ID          string
	ConceptID   string
	UserID      string
	Message     string
	IsRecurring bool
	CronExpr    *string
	ScheduledAt *time.Time
	LastSentAt  *time.Time
	IsActive    bool
	CreatedAt   time.Time
}
