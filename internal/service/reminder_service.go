package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

type ReminderService interface {
	ListConceptReminders(ctx context.Context, userID string, conceptID string) ([]domain.Reminder, error)
	Create(ctx context.Context, conceptID string, userID string, reminder domain.Reminder) (domain.Reminder, error)
	Update(ctx context.Context, userID string, id string, update UpdateReminderParams) (domain.Reminder, error)
	Delete(ctx context.Context, userID string, id string) error
}

type reminderService struct {
	repo repository.ReminderRepository
}

func NewReminderService(repo repository.ReminderRepository) ReminderService {
	return &reminderService{
		repo: repo,
	}
}

func (r *reminderService) ListConceptReminders(ctx context.Context, userID string, conceptID string) ([]domain.Reminder, error) {
	l, err := r.repo.ListConceptReminders(ctx, userID, conceptID)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (r *reminderService) Create(ctx context.Context, conceptID string, userID string, reminder domain.Reminder) (domain.Reminder, error) {
	create, err := r.repo.Create(ctx, conceptID, userID, reminder)
	if err != nil {
		return domain.Reminder{}, err
	}

	return create, nil
}

type UpdateReminderParams struct {
	Message     string
	IsRecurring bool
	CronExpr    *string
	ScheduledAt *time.Time
	IsActive    bool
}

func (r *reminderService) Update(ctx context.Context, userID string, id string, update UpdateReminderParams) (domain.Reminder, error) {
	up := repository.UpdateReminderParams{
		Message:     update.Message,
		IsRecurring: update.IsRecurring,
		CronExpr:    update.CronExpr,
		ScheduledAt: update.ScheduledAt,
		IsActive:    update.IsActive,
	}

	u, err := r.repo.Update(ctx, userID, id, up)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Reminder{}, domain.ErrNotFound
		}

		return domain.Reminder{}, err
	}

	return u, nil
}

func (r *reminderService) Delete(ctx context.Context, userID string, id string) error {
	if err := r.repo.Delete(ctx, userID, id); err != nil {
		return err
	}

	return nil
}
