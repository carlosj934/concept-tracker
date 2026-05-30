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

type UpdateReminderParams struct {
	Message     string     `json:"message"`
	IsRecurring bool       `json:"is_recurring"`
	CronExpr    *string    `json:"cron_expr"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	IsActive    bool       `json:"is_active"`
}
