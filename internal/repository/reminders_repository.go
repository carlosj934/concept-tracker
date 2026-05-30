package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"concept-tracker/internal/domain"
)

type ReminderRepository interface {
	// user facing
	ListConceptReminders(ctx context.Context, userID string, conceptID string) ([]domain.Reminder, error)
	Create(ctx context.Context, conceptID string, userID string, reminder domain.Reminder) (domain.Reminder, error)
	Update(ctx context.Context, userID string, id string, update UpdateReminderParams) (domain.Reminder, error)
	Delete(ctx context.Context, userID string, id string) error

	// worker facing
	GetActiveReminders(ctx context.Context) ([]domain.Reminder, error)
	AdvanceSchedule(ctx context.Context, id string, scheduledAt *time.Time, lastSentAt *time.Time, isActive bool) error
}

type UpdateReminderParams struct {
	Message     string
	IsRecurring bool
	CronExpr    *string
	ScheduledAt *time.Time
	IsActive    bool
}

type postgresReminderRepository struct {
	pool *pgxpool.Pool
}

func NewReminder(p *pgxpool.Pool) ReminderRepository {
	return &postgresReminderRepository{pool: p}
}

func (r *postgresReminderRepository) ListConceptReminders(ctx context.Context, userID string, conceptID string) ([]domain.Reminder, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT id, concept_id, user_id, message, is_recurring, cron_expr, scheduled_at, last_sent_at, is_active, created_at
	FROM reminders
	WHERE user_id = $1 AND concept_id = $2
	`, userID, conceptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []domain.Reminder
	for rows.Next() {
		var re domain.Reminder
		var id, conceptID, userID, message string
		var cronExpr *string
		var isActive, isRecurring bool
		var scheduledAt, lastSentAt *time.Time
		var createdAt time.Time

		err := rows.Scan(&id, &conceptID, &userID, &message, &isRecurring, &cronExpr, &scheduledAt, &lastSentAt, &isActive, &createdAt)
		if err != nil {
			return nil, err
		}

		re.ID = id
		re.ConceptID = conceptID
		re.UserID = userID
		re.Message = message
		re.IsRecurring = isRecurring
		re.CronExpr = cronExpr
		re.ScheduledAt = scheduledAt
		re.LastSentAt = lastSentAt
		re.IsActive = isActive
		re.CreatedAt = createdAt
		reminders = append(reminders, re)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *postgresReminderRepository) Create(ctx context.Context, conceptID string, userID string, reminder domain.Reminder) (domain.Reminder, error) {
	var id string
	var createdAt time.Time

	err := r.pool.QueryRow(ctx, `
	INSERT INTO reminders (concept_id, user_id, message, is_recurring, cron_expr, scheduled_at, last_sent_at, is_active)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, created_at
	`, conceptID, userID, reminder.Message, reminder.IsRecurring, reminder.CronExpr, reminder.ScheduledAt, reminder.LastSentAt, reminder.IsActive).Scan(&id, &createdAt)
	if err != nil {
		return domain.Reminder{}, err
	}

	return domain.Reminder{
		ID:          id,
		ConceptID:   conceptID,
		UserID:      userID,
		Message:     reminder.Message,
		IsRecurring: reminder.IsRecurring,
		CronExpr:    reminder.CronExpr,
		ScheduledAt: reminder.ScheduledAt,
		LastSentAt:  reminder.LastSentAt,
		IsActive:    reminder.IsActive,
		CreatedAt:   createdAt,
	}, nil
}

func (r *postgresReminderRepository) Update(ctx context.Context, userID string, id string, update UpdateReminderParams) (domain.Reminder, error) {
	var i, u, conceptID, message string
	var isRecurring, isActive bool
	var cronExpr *string
	var scheduledAt *time.Time
	var createdAt time.Time
	var lastSentAt *time.Time

	err := r.pool.QueryRow(ctx, `
	UPDATE reminders
	SET message = $1, is_recurring = $2, cron_expr = COALESCE($3, cron_expr), scheduled_at = COALESCE($4, scheduled_at), is_active = $5
	WHERE id = $6 AND user_id = $7
	RETURNING id, user_id, concept_id, message, is_recurring, cron_expr, scheduled_at, last_sent_at, is_active, created_at
	`, update.Message, update.IsRecurring, update.CronExpr, update.ScheduledAt, update.IsActive, id, userID).Scan(&i, &u, &conceptID, &message, &isRecurring, &cronExpr, &scheduledAt, &lastSentAt, &isActive, &createdAt)
	if err != nil {
		return domain.Reminder{}, err
	}

	return domain.Reminder{
		ID:          i,
		ConceptID:   conceptID,
		UserID:      u,
		Message:     message,
		IsRecurring: isRecurring,
		CronExpr:    cronExpr,
		ScheduledAt: scheduledAt,
		LastSentAt:  lastSentAt,
		IsActive:    isActive,
		CreatedAt:   createdAt,
	}, nil
}

func (r *postgresReminderRepository) Delete(ctx context.Context, userID string, id string) error {
	_, err := r.pool.Exec(ctx, `
	DELETE FROM reminders
	WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresReminderRepository) GetActiveReminders(ctx context.Context) ([]domain.Reminder, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT id, user_id, is_recurring, cron_expr
	FROM reminders
	WHERE is_active = true AND scheduled_at <= now()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []domain.Reminder
	for rows.Next() {
		var re domain.Reminder
		var id, userID string
		var isRecurring bool
		var cronExpr *string

		err := rows.Scan(&id, &userID, &isRecurring, &cronExpr)
		if err != nil {
			return nil, err
		}
		re.ID = id
		re.UserID = userID
		re.IsRecurring = isRecurring
		re.CronExpr = cronExpr
		reminders = append(reminders, re)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *postgresReminderRepository) AdvanceSchedule(ctx context.Context, id string, scheduledAt *time.Time, lastSentAt *time.Time, isActive bool) error {
	_, err := r.pool.Exec(ctx, `
	UPDATE reminders
	SET scheduled_at = $1, last_sent_at = $2, is_active = $3
	WHERE id = $4
	`, scheduledAt, lastSentAt, isActive, id)
	if err != nil {
		return err
	}

	return nil
}
