package integration_tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

func TestReminderRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("one-off reminder", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewReminder(testPool)
		scheduledAt := time.Now().Add(24 * time.Hour).UTC()

		r, err := repo.Create(ctx, concept.ID, "user_test123", domain.Reminder{
			Message:     "Review Go notes",
			IsRecurring: false,
			ScheduledAt: &scheduledAt,
			IsActive:    true,
		})
		require.NoError(t, err)
		require.NotEmpty(t, r.ID)
		require.Equal(t, "user_test123", r.UserID)
		require.Equal(t, concept.ID, r.ConceptID)
		require.Equal(t, "Review Go notes", r.Message)
		require.False(t, r.IsRecurring)
		require.Nil(t, r.CronExpr)
		require.True(t, r.IsActive)
	})

	t.Run("recurring reminder", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewReminder(testPool)
		cronExpr := "0 9 * * 1-5"

		r, err := repo.Create(ctx, concept.ID, "user_test123", domain.Reminder{
			Message:     "Daily Go review",
			IsRecurring: true,
			CronExpr:    &cronExpr,
			IsActive:    true,
		})
		require.NoError(t, err)
		require.True(t, r.IsRecurring)
		require.Equal(t, "0 9 * * 1-5", *r.CronExpr)
	})
}

func TestReminderRepository_ListConceptReminders(t *testing.T) {
	ctx := context.Background()

	t.Run("returns reminders for concept", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewReminder(testPool)
		scheduledAt := time.Now().Add(24 * time.Hour).UTC()

		r1, err := repo.Create(ctx, concept.ID, "user_test123", domain.Reminder{
			Message: "Reminder 1", IsActive: true, ScheduledAt: &scheduledAt,
		})
		require.NoError(t, err)

		r2, err := repo.Create(ctx, concept.ID, "user_test123", domain.Reminder{
			Message: "Reminder 2", IsActive: true, ScheduledAt: &scheduledAt,
		})
		require.NoError(t, err)

		reminders, err := repo.ListConceptReminders(ctx, "user_test123", concept.ID)
		require.NoError(t, err)
		require.Len(t, reminders, 2)

		ids := []string{reminders[0].ID, reminders[1].ID}
		require.Contains(t, ids, r1.ID)
		require.Contains(t, ids, r2.ID)
	})

	t.Run("returns empty for concept with no reminders", func(t *testing.T) {
		t.Cleanup(func() { restoreDB(t, ctx) })

		conceptRepo := repository.New(testPool)
		concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

		repo := repository.NewReminder(testPool)
		reminders, err := repo.ListConceptReminders(ctx, "user_test123", concept.ID)
		require.NoError(t, err)
		require.Empty(t, reminders)
	})
}

func TestReminderRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewReminder(testPool)
	scheduledAt := time.Now().Add(24 * time.Hour).UTC()

	created, err := repo.Create(ctx, concept.ID, "user_test123", domain.Reminder{
		Message: "Original message", IsActive: true, ScheduledAt: &scheduledAt,
	})
	require.NoError(t, err)

	newCron := "0 10 * * *"
	updated, err := repo.Update(ctx, "user_test123", created.ID, domain.UpdateReminderParams{
		Message:     "Updated message",
		IsRecurring: true,
		CronExpr:    &newCron,
		IsActive:    false,
	})
	require.NoError(t, err)
	require.Equal(t, "Updated message", updated.Message)
	require.True(t, updated.IsRecurring)
	require.Equal(t, "0 10 * * *", *updated.CronExpr)
	require.False(t, updated.IsActive)
}

func TestReminderRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Cleanup(func() { restoreDB(t, ctx) })

	conceptRepo := repository.New(testPool)
	concept := createTestConcept(t, ctx, conceptRepo, "user_test123", "golang", nil)

	repo := repository.NewReminder(testPool)
	scheduledAt := time.Now().Add(24 * time.Hour).UTC()

	created, err := repo.Create(ctx, concept.ID, "user_test123", domain.Reminder{
		Message: "To be deleted", IsActive: true, ScheduledAt: &scheduledAt,
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, "user_test123", created.ID)
	require.NoError(t, err)

	reminders, err := repo.ListConceptReminders(ctx, "user_test123", concept.ID)
	require.NoError(t, err)
	require.Empty(t, reminders)
}
